package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"golang-system-monitor-tui/models"
)

func TestNewNetworkModel(t *testing.T) {
	model := NewNetworkModel()

	if len(model.interfaces) != 0 {
		t.Errorf("Expected empty interfaces slice, got %v", model.interfaces)
	}
	if len(model.previousData) != 0 {
		t.Errorf("Expected empty previousData slice, got %v", model.previousData)
	}
	if len(model.rates) != 0 {
		t.Errorf("Expected empty rates map, got %v", model.rates)
	}
	if model.width != 50 {
		t.Errorf("Expected width to be 50, got %d", model.width)
	}
	if model.height != 10 {
		t.Errorf("Expected height to be 10, got %d", model.height)
	}
}

func TestNetworkModel_Init(t *testing.T) {
	model := NewNetworkModel()
	cmd := model.Init()

	if cmd != nil {
		t.Errorf("Expected Init() to return nil, got %v", cmd)
	}
}

func TestNetworkModel_Update_NetworkUpdateMsg(t *testing.T) {
	model := NewNetworkModel()
	timestamp := time.Now()

	networkInfo := []models.NetworkInfo{
		{
			Interface:   "eth0",
			BytesSent:   1024000,
			BytesRecv:   2048000,
			PacketsSent: 1000,
			PacketsRecv: 2000,
			Timestamp:   timestamp,
		},
		{
			Interface:   "wlan0",
			BytesSent:   512000,
			BytesRecv:   1024000,
			PacketsSent: 500,
			PacketsRecv: 1000,
			Timestamp:   timestamp,
		},
	}

	updateMsg := NetworkUpdateMsg(networkInfo)
	updatedModel, cmd := model.Update(updateMsg)

	if cmd != nil {
		t.Errorf("Expected Update() to return nil cmd, got %v", cmd)
	}

	if len(updatedModel.interfaces) != 2 {
		t.Errorf("Expected 2 network interfaces, got %d", len(updatedModel.interfaces))
	}

	// Check first interface
	if updatedModel.interfaces[0].Interface != "eth0" {
		t.Errorf("Expected first interface to be 'eth0', got '%s'", updatedModel.interfaces[0].Interface)
	}
	if updatedModel.interfaces[0].BytesSent != 1024000 {
		t.Errorf("Expected BytesSent to be 1024000, got %d", updatedModel.interfaces[0].BytesSent)
	}
	if updatedModel.interfaces[0].BytesRecv != 2048000 {
		t.Errorf("Expected BytesRecv to be 2048000, got %d", updatedModel.interfaces[0].BytesRecv)
	}

	// Check second interface
	if updatedModel.interfaces[1].Interface != "wlan0" {
		t.Errorf("Expected second interface to be 'wlan0', got '%s'", updatedModel.interfaces[1].Interface)
	}
	if updatedModel.interfaces[1].BytesSent != 512000 {
		t.Errorf("Expected BytesSent to be 512000, got %d", updatedModel.interfaces[1].BytesSent)
	}
	if updatedModel.interfaces[1].BytesRecv != 1024000 {
		t.Errorf("Expected BytesRecv to be 1024000, got %d", updatedModel.interfaces[1].BytesRecv)
	}

	if !updatedModel.lastUpdate.Equal(timestamp) {
		t.Errorf("Expected lastUpdate to be %v, got %v", timestamp, updatedModel.lastUpdate)
	}
}

func TestNetworkModel_Update_RateCalculation(t *testing.T) {
	model := NewNetworkModel()
	baseTime := time.Now()

	// First update - establish baseline
	networkInfo1 := []models.NetworkInfo{
		{
			Interface:   "eth0",
			BytesSent:   1000000,
			BytesRecv:   2000000,
			PacketsSent: 1000,
			PacketsRecv: 2000,
			Timestamp:   baseTime,
		},
	}
	model, _ = model.Update(NetworkUpdateMsg(networkInfo1))

	// Second update - 1 second later with increased bytes
	networkInfo2 := []models.NetworkInfo{
		{
			Interface:   "eth0",
			BytesSent:   1001024, // +1024 bytes sent
			BytesRecv:   2002048, // +2048 bytes received
			PacketsSent: 1010,
			PacketsRecv: 2020,
			Timestamp:   baseTime.Add(time.Second),
		},
	}
	model, _ = model.Update(NetworkUpdateMsg(networkInfo2))

	// Check that rates were calculated
	if len(model.rates) != 1 {
		t.Errorf("Expected 1 rate entry, got %d", len(model.rates))
	}

	stats, exists := model.rates["eth0"]
	if !exists {
		t.Errorf("Expected rate entry for eth0")
	}

	// Should be 1024 bytes/second send rate
	if stats.SendRate != 1024.0 {
		t.Errorf("Expected SendRate to be 1024.0, got %f", stats.SendRate)
	}

	// Should be 2048 bytes/second receive rate
	if stats.RecvRate != 2048.0 {
		t.Errorf("Expected RecvRate to be 2048.0, got %f", stats.RecvRate)
	}
}

func TestNetworkModel_calculateRates(t *testing.T) {
	model := NewNetworkModel()
	baseTime := time.Now()

	previous := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 1000,
			BytesRecv: 2000,
			Timestamp: baseTime,
		},
		{
			Interface: "wlan0",
			BytesSent: 500,
			BytesRecv: 1000,
			Timestamp: baseTime,
		},
	}

	current := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 2000, // +1000 bytes in 2 seconds = 500 B/s
			BytesRecv: 4000, // +2000 bytes in 2 seconds = 1000 B/s
			Timestamp: baseTime.Add(2 * time.Second),
		},
		{
			Interface: "wlan0",
			BytesSent: 1500, // +1000 bytes in 2 seconds = 500 B/s
			BytesRecv: 3000, // +2000 bytes in 2 seconds = 1000 B/s
			Timestamp: baseTime.Add(2 * time.Second),
		},
	}

	rates := model.calculateRates(previous, current)

	if len(rates) != 2 {
		t.Errorf("Expected 2 rate entries, got %d", len(rates))
	}

	// Check eth0 rates
	eth0Stats, exists := rates["eth0"]
	if !exists {
		t.Errorf("Expected rate entry for eth0")
	}
	if eth0Stats.SendRate != 500.0 {
		t.Errorf("Expected eth0 SendRate to be 500.0, got %f", eth0Stats.SendRate)
	}
	if eth0Stats.RecvRate != 1000.0 {
		t.Errorf("Expected eth0 RecvRate to be 1000.0, got %f", eth0Stats.RecvRate)
	}

	// Check wlan0 rates
	wlan0Stats, exists := rates["wlan0"]
	if !exists {
		t.Errorf("Expected rate entry for wlan0")
	}
	if wlan0Stats.SendRate != 500.0 {
		t.Errorf("Expected wlan0 SendRate to be 500.0, got %f", wlan0Stats.SendRate)
	}
	if wlan0Stats.RecvRate != 1000.0 {
		t.Errorf("Expected wlan0 RecvRate to be 1000.0, got %f", wlan0Stats.RecvRate)
	}
}

func TestNetworkModel_calculateRates_CounterRollover(t *testing.T) {
	model := NewNetworkModel()
	baseTime := time.Now()

	previous := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 4294967295, // Near max uint32
			BytesRecv: 4294967295,
			Timestamp: baseTime,
		},
	}

	current := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 1000, // Counter rolled over
			BytesRecv: 2000, // Counter rolled over
			Timestamp: baseTime.Add(time.Second),
		},
	}

	rates := model.calculateRates(previous, current)

	// Should handle rollover by setting rate to 0
	eth0Stats, exists := rates["eth0"]
	if !exists {
		t.Errorf("Expected rate entry for eth0")
	}
	if eth0Stats.SendRate != 0.0 {
		t.Errorf("Expected SendRate to be 0.0 for counter rollover, got %f", eth0Stats.SendRate)
	}
	if eth0Stats.RecvRate != 0.0 {
		t.Errorf("Expected RecvRate to be 0.0 for counter rollover, got %f", eth0Stats.RecvRate)
	}
}

func TestNetworkModel_Update_OtherMessages(t *testing.T) {
	model := NewNetworkModel()
	originalModel := model

	// Test with a different message type
	otherMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updatedModel, cmd := model.Update(otherMsg)

	if cmd != nil {
		t.Errorf("Expected Update() to return nil cmd for non-network messages, got %v", cmd)
	}

	// Model should remain unchanged
	if len(updatedModel.interfaces) != len(originalModel.interfaces) {
		t.Errorf("Expected model to remain unchanged for non-network messages")
	}
}

func TestNetworkModel_View_NoData(t *testing.T) {
	model := NewNetworkModel()
	view := model.View()

	if !strings.Contains(view, "Network Activity") {
		t.Errorf("Expected view to contain 'Network Activity' header")
	}

	if !strings.Contains(view, "Loading network data...") {
		t.Errorf("Expected view to contain loading message when no data available")
	}
}

func TestNetworkModel_View_WithData(t *testing.T) {
	model := NewNetworkModel()
	baseTime := time.Now()

	// Add network data with rates
	networkInfo1 := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 1000000,
			BytesRecv: 2000000,
			Timestamp: baseTime,
		},
	}
	model, _ = model.Update(NetworkUpdateMsg(networkInfo1))

	networkInfo2 := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 1001024, // +1024 B/s
			BytesRecv: 2002048, // +2048 B/s
			Timestamp: baseTime.Add(time.Second),
		},
	}
	model, _ = model.Update(NetworkUpdateMsg(networkInfo2))

	view := model.View()

	if !strings.Contains(view, "Network Activity") {
		t.Errorf("Expected view to contain 'Network Activity' header")
	}

	if !strings.Contains(view, "eth0") {
		t.Errorf("Expected view to contain interface name 'eth0'")
	}

	if !strings.Contains(view, "↑") {
		t.Errorf("Expected view to contain upload arrow '↑'")
	}

	if !strings.Contains(view, "↓") {
		t.Errorf("Expected view to contain download arrow '↓'")
	}

	// Should show rates
	if !strings.Contains(view, "1.0KB/s") {
		t.Errorf("Expected view to contain send rate '1.0KB/s', got: %s", view)
	}

	if !strings.Contains(view, "2.0KB/s") {
		t.Errorf("Expected view to contain receive rate '2.0KB/s', got: %s", view)
	}
}

func TestNetworkModel_formatRate(t *testing.T) {
	model := NewNetworkModel()

	tests := []struct {
		bytesPerSec float64
		expected    string
	}{
		{0, "0B/s"},
		{512, "512B/s"},
		{1024, "1.0KB/s"},
		{1536, "1.5KB/s"},
		{1048576, "1.0MB/s"},
		{1572864, "1.5MB/s"},
		{1073741824, "1.0GB/s"},
		{1610612736, "1.5GB/s"},
	}

	for _, test := range tests {
		result := model.formatRate(test.bytesPerSec)
		if result != test.expected {
			t.Errorf("formatRate(%.0f) = %s, expected %s", test.bytesPerSec, result, test.expected)
		}
	}
}

func TestNetworkModel_formatBytes(t *testing.T) {
	model := NewNetworkModel()

	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1048576, "1.0MB"},
		{1572864, "1.5MB"},
		{1073741824, "1.0GB"},
		{1610612736, "1.5GB"},
		{1099511627776, "1.0TB"},
	}

	for _, test := range tests {
		result := model.formatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", test.bytes, result, test.expected)
		}
	}
}

func TestNetworkModel_SetSize(t *testing.T) {
	model := NewNetworkModel()
	newModel := model.SetSize(80, 20)

	if newModel.width != 80 {
		t.Errorf("Expected width to be 80, got %d", newModel.width)
	}

	if newModel.height != 20 {
		t.Errorf("Expected height to be 20, got %d", newModel.height)
	}
}

func TestNetworkModel_GetInterfaces(t *testing.T) {
	model := NewNetworkModel()
	timestamp := time.Now()

	networkInfo := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 1024,
			BytesRecv: 2048,
			Timestamp: timestamp,
		},
	}

	model, _ = model.Update(NetworkUpdateMsg(networkInfo))
	interfaces := model.GetInterfaces()

	if len(interfaces) != 1 {
		t.Errorf("Expected 1 interface, got %d", len(interfaces))
	}

	if interfaces[0].Interface != "eth0" {
		t.Errorf("Expected interface name 'eth0', got '%s'", interfaces[0].Interface)
	}
}

func TestNetworkModel_GetRates(t *testing.T) {
	model := NewNetworkModel()
	baseTime := time.Now()

	// Set up rate calculation
	networkInfo1 := []models.NetworkInfo{
		{Interface: "eth0", BytesSent: 1000, BytesRecv: 2000, Timestamp: baseTime},
	}
	model, _ = model.Update(NetworkUpdateMsg(networkInfo1))

	networkInfo2 := []models.NetworkInfo{
		{Interface: "eth0", BytesSent: 2000, BytesRecv: 4000, Timestamp: baseTime.Add(time.Second)},
	}
	model, _ = model.Update(NetworkUpdateMsg(networkInfo2))

	rates := model.GetRates()
	if len(rates) != 1 {
		t.Errorf("Expected 1 rate entry, got %d", len(rates))
	}

	stats, exists := rates["eth0"]
	if !exists {
		t.Errorf("Expected rate entry for eth0")
	}
	if stats.SendRate != 1000.0 {
		t.Errorf("Expected SendRate 1000.0, got %f", stats.SendRate)
	}
	if stats.RecvRate != 2000.0 {
		t.Errorf("Expected RecvRate 2000.0, got %f", stats.RecvRate)
	}
}

func TestNetworkModel_GetTotalRates(t *testing.T) {
	model := NewNetworkModel()
	baseTime := time.Now()

	// Set up multiple interfaces with rates
	networkInfo1 := []models.NetworkInfo{
		{Interface: "eth0", BytesSent: 1000, BytesRecv: 2000, Timestamp: baseTime},
		{Interface: "wlan0", BytesSent: 500, BytesRecv: 1000, Timestamp: baseTime},
	}
	model, _ = model.Update(NetworkUpdateMsg(networkInfo1))

	networkInfo2 := []models.NetworkInfo{
		{Interface: "eth0", BytesSent: 2000, BytesRecv: 4000, Timestamp: baseTime.Add(time.Second)},
		{Interface: "wlan0", BytesSent: 1000, BytesRecv: 2000, Timestamp: baseTime.Add(time.Second)},
	}
	model, _ = model.Update(NetworkUpdateMsg(networkInfo2))

	totalSendRate := model.GetTotalSendRate()
	totalRecvRate := model.GetTotalRecvRate()

	// eth0: 1000 B/s send, wlan0: 500 B/s send = 1500 B/s total
	if totalSendRate != 1500.0 {
		t.Errorf("Expected total send rate 1500.0, got %f", totalSendRate)
	}

	// eth0: 2000 B/s recv, wlan0: 1000 B/s recv = 3000 B/s total
	if totalRecvRate != 3000.0 {
		t.Errorf("Expected total recv rate 3000.0, got %f", totalRecvRate)
	}
}

func TestNetworkModel_GetHighActivityInterfaces(t *testing.T) {
	model := NewNetworkModel()
	baseTime := time.Now()

	// Set up interfaces with different activity levels
	networkInfo1 := []models.NetworkInfo{
		{Interface: "eth0", BytesSent: 0, BytesRecv: 0, Timestamp: baseTime},
		{Interface: "wlan0", BytesSent: 0, BytesRecv: 0, Timestamp: baseTime},
	}
	model, _ = model.Update(NetworkUpdateMsg(networkInfo1))

	networkInfo2 := []models.NetworkInfo{
		{Interface: "eth0", BytesSent: 2*1024*1024, BytesRecv: 0, Timestamp: baseTime.Add(time.Second)}, // 2 MB/s
		{Interface: "wlan0", BytesSent: 512*1024, BytesRecv: 0, Timestamp: baseTime.Add(time.Second)},   // 512 KB/s
	}
	model, _ = model.Update(NetworkUpdateMsg(networkInfo2))

	highActivity := model.GetHighActivityInterfaces()
	if len(highActivity) != 1 {
		t.Errorf("Expected 1 high activity interface, got %d", len(highActivity))
	}

	if highActivity[0] != "eth0" {
		t.Errorf("Expected high activity interface to be 'eth0', got '%s'", highActivity[0])
	}

	if !model.HasHighActivity() {
		t.Errorf("Expected HasHighActivity() to return true")
	}
}

func TestNetworkModel_GetInterfaceByName(t *testing.T) {
	model := NewNetworkModel()
	timestamp := time.Now()

	networkInfo := []models.NetworkInfo{
		{Interface: "eth0", BytesSent: 1024, BytesRecv: 2048, Timestamp: timestamp},
		{Interface: "wlan0", BytesSent: 512, BytesRecv: 1024, Timestamp: timestamp},
	}

	model, _ = model.Update(NetworkUpdateMsg(networkInfo))

	// Test existing interface
	iface, exists := model.GetInterfaceByName("eth0")
	if !exists {
		t.Errorf("Expected to find interface 'eth0'")
	}
	if iface.BytesSent != 1024 {
		t.Errorf("Expected BytesSent 1024, got %d", iface.BytesSent)
	}

	// Test non-existing interface
	_, exists = model.GetInterfaceByName("nonexistent")
	if exists {
		t.Errorf("Expected not to find interface 'nonexistent'")
	}
}

func TestNetworkModel_GetRateByInterface(t *testing.T) {
	model := NewNetworkModel()
	baseTime := time.Now()

	// Set up rate calculation
	networkInfo1 := []models.NetworkInfo{
		{Interface: "eth0", BytesSent: 1000, BytesRecv: 2000, Timestamp: baseTime},
	}
	model, _ = model.Update(NetworkUpdateMsg(networkInfo1))

	networkInfo2 := []models.NetworkInfo{
		{Interface: "eth0", BytesSent: 2000, BytesRecv: 4000, Timestamp: baseTime.Add(time.Second)},
	}
	model, _ = model.Update(NetworkUpdateMsg(networkInfo2))

	// Test existing interface
	stats, exists := model.GetRateByInterface("eth0")
	if !exists {
		t.Errorf("Expected to find rates for interface 'eth0'")
	}
	if stats.SendRate != 1000.0 {
		t.Errorf("Expected SendRate 1000.0, got %f", stats.SendRate)
	}

	// Test non-existing interface
	_, exists = model.GetRateByInterface("nonexistent")
	if exists {
		t.Errorf("Expected not to find rates for interface 'nonexistent'")
	}
}

func TestNetworkModel_StyleManagerIntegration(t *testing.T) {
	model := NewNetworkModel()
	
	// Test that style manager is initialized
	if model.styleManager == nil {
		t.Error("Expected style manager to be initialized")
	}
	
	testText := "test interface"

	// Test different activity levels
	testCases := []struct {
		sendRate float64
		recvRate float64
		name     string
	}{
		{0, 0, "no activity"},
		{500*1024, 0, "low activity"},
		{5*1024*1024, 0, "medium activity"},
		{15*1024*1024, 0, "high activity"},
	}

	for _, tc := range testCases {
		stats := models.NetworkStats{
			SendRate: tc.sendRate,
			RecvRate: tc.recvRate,
		}
		styled := model.styleByActivityWithManager(testText, stats)
		if len(styled) == 0 {
			t.Errorf("Expected non-empty styled text for %s", tc.name)
		}
		// We can't easily test colors in unit tests, but we ensure the function doesn't crash
	}
}