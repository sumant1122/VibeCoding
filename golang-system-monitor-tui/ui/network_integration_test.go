package ui

import (
	"testing"
	"time"

	"golang-system-monitor-tui/models"
	"golang-system-monitor-tui/services"
)

func TestNetworkModel_IntegrationWithCollector(t *testing.T) {
	// Create a real collector
	collector := services.NewGopsutilCollector()
	
	// Create network model
	model := NewNetworkModel()
	
	// Collect initial network data
	networkInfo1, err := collector.CollectNetwork()
	if err != nil {
		t.Skipf("Skipping integration test due to network collection error: %v", err)
	}
	
	if len(networkInfo1) == 0 {
		t.Skip("Skipping integration test: no network interfaces found")
	}
	
	// Update model with first measurement
	model, _ = model.Update(NetworkUpdateMsg(networkInfo1))
	
	// Verify model was updated
	interfaces := model.GetInterfaces()
	if len(interfaces) == 0 {
		t.Errorf("Expected network interfaces to be populated")
	}
	
	// Wait a bit and collect again for rate calculation
	time.Sleep(100 * time.Millisecond)
	
	networkInfo2, err := collector.CollectNetwork()
	if err != nil {
		t.Errorf("Failed to collect network data for rate calculation: %v", err)
		return
	}
	
	// Update model with second measurement
	model, _ = model.Update(NetworkUpdateMsg(networkInfo2))
	
	// Verify rates were calculated (even if they're zero)
	rates := model.GetRates()
	if len(rates) == 0 {
		t.Errorf("Expected network rates to be calculated")
	}
	
	// Verify view renders without errors
	view := model.View()
	if len(view) == 0 {
		t.Errorf("Expected non-empty view")
	}
	
	// Test that we can find interfaces by name
	for _, iface := range interfaces {
		found, exists := model.GetInterfaceByName(iface.Interface)
		if !exists {
			t.Errorf("Expected to find interface %s", iface.Interface)
		}
		if found.Interface != iface.Interface {
			t.Errorf("Expected interface name %s, got %s", iface.Interface, found.Interface)
		}
	}
}

func TestNetworkModel_RateCalculationAccuracy(t *testing.T) {
	model := NewNetworkModel()
	baseTime := time.Now()
	
	// Create test data with known transfer amounts
	networkInfo1 := []models.NetworkInfo{
		{
			Interface:   "test0",
			BytesSent:   1000000,  // 1MB
			BytesRecv:   2000000,  // 2MB
			PacketsSent: 1000,
			PacketsRecv: 2000,
			Timestamp:   baseTime,
		},
	}
	
	networkInfo2 := []models.NetworkInfo{
		{
			Interface:   "test0",
			BytesSent:   1512000,  // 1.5MB (+512KB in 1 second = 512KB/s)
			BytesRecv:   3048000,  // ~3MB (+1MB in 1 second = 1MB/s)
			PacketsSent: 1500,
			PacketsRecv: 3000,
			Timestamp:   baseTime.Add(time.Second),
		},
	}
	
	// First update
	model, _ = model.Update(NetworkUpdateMsg(networkInfo1))
	
	// Second update
	model, _ = model.Update(NetworkUpdateMsg(networkInfo2))
	
	// Verify rate calculation accuracy
	stats, exists := model.GetRateByInterface("test0")
	if !exists {
		t.Errorf("Expected to find rates for test0 interface")
		return
	}
	
	expectedSendRate := 512000.0 // 512KB/s
	expectedRecvRate := 1048000.0 // ~1MB/s
	
	if stats.SendRate != expectedSendRate {
		t.Errorf("Expected send rate %.0f B/s, got %.0f B/s", expectedSendRate, stats.SendRate)
	}
	
	if stats.RecvRate != expectedRecvRate {
		t.Errorf("Expected recv rate %.0f B/s, got %.0f B/s", expectedRecvRate, stats.RecvRate)
	}
	
	// Test formatting of these rates
	sendRateStr := model.formatRate(stats.SendRate)
	recvRateStr := model.formatRate(stats.RecvRate)
	
	if sendRateStr != "500.0KB/s" {
		t.Errorf("Expected send rate format '500.0KB/s', got '%s'", sendRateStr)
	}
	
	if recvRateStr != "1023.4KB/s" {
		t.Errorf("Expected recv rate format '1023.4KB/s', got '%s'", recvRateStr)
	}
}

func TestNetworkModel_ViewFormatting(t *testing.T) {
	model := NewNetworkModel()
	baseTime := time.Now()
	
	// Create test data with various interface names and rates
	networkInfo1 := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 0,
			BytesRecv: 0,
			Timestamp: baseTime,
		},
		{
			Interface: "very-long-interface-name",
			BytesSent: 0,
			BytesRecv: 0,
			Timestamp: baseTime,
		},
	}
	
	networkInfo2 := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 1024000,    // 1MB/s
			BytesRecv: 2048000,    // 2MB/s
			Timestamp: baseTime.Add(time.Second),
		},
		{
			Interface: "very-long-interface-name",
			BytesSent: 512000,     // 500KB/s
			BytesRecv: 1024000,    // 1MB/s
			Timestamp: baseTime.Add(time.Second),
		},
	}
	
	// Update model
	model, _ = model.Update(NetworkUpdateMsg(networkInfo1))
	model, _ = model.Update(NetworkUpdateMsg(networkInfo2))
	
	view := model.View()
	
	// Check that view contains expected elements
	if !contains(view, "Network Activity") {
		t.Errorf("Expected view to contain header 'Network Activity'")
	}
	
	if !contains(view, "eth0") {
		t.Errorf("Expected view to contain interface 'eth0'")
	}
	
	if !contains(view, "very-long...") {
		t.Errorf("Expected view to contain truncated long interface name")
	}
	
	if !contains(view, "↑") && !contains(view, "↓") {
		t.Errorf("Expected view to contain upload/download arrows")
	}
	
	if !contains(view, "1000.0KB/s") {
		t.Errorf("Expected view to contain rate '1000.0KB/s'")
	}
	
	if !contains(view, "2.0MB/s") {
		t.Errorf("Expected view to contain rate '2.0MB/s'")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}