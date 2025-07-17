package services

import (
	"errors"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"golang-system-monitor-tui/models"
)

func TestNewGopsutilCollector(t *testing.T) {
	collector := NewGopsutilCollector()
	if collector == nil {
		t.Fatal("NewGopsutilCollector should return a non-nil collector")
	}
	
	if collector.errorHandler == nil {
		t.Error("Expected error handler to be initialized")
	}
}

func TestNewGopsutilCollectorWithErrorHandler(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	errorHandler := models.NewErrorHandler(logger)
	
	collector := NewGopsutilCollectorWithErrorHandler(errorHandler)
	if collector == nil {
		t.Fatal("NewGopsutilCollectorWithErrorHandler should return a non-nil collector")
	}
	
	if collector.errorHandler != errorHandler {
		t.Error("Expected custom error handler to be set")
	}
}

func TestGopsutilCollector_CollectCPU(t *testing.T) {
	collector := NewGopsutilCollector()
	
	cpuInfo, err := collector.CollectCPU()
	if err != nil {
		t.Fatalf("CollectCPU failed: %v", err)
	}

	// Validate CPU info structure
	if cpuInfo.Cores <= 0 {
		t.Errorf("Expected positive number of cores, got %d", cpuInfo.Cores)
	}

	if len(cpuInfo.Usage) != cpuInfo.Cores {
		t.Errorf("Expected %d core usage values, got %d", cpuInfo.Cores, len(cpuInfo.Usage))
	}

	// Validate usage percentages are within valid range
	for i, usage := range cpuInfo.Usage {
		if usage < 0 || usage > 100 {
			t.Errorf("Core %d usage %f is not within valid range [0, 100]", i, usage)
		}
	}

	if cpuInfo.Total < 0 || cpuInfo.Total > 100 {
		t.Errorf("Total CPU usage %f is not within valid range [0, 100]", cpuInfo.Total)
	}

	// Validate timestamp is recent
	if time.Since(cpuInfo.Timestamp) > time.Minute {
		t.Errorf("CPU info timestamp is too old: %v", cpuInfo.Timestamp)
	}
}

func TestGopsutilCollector_CollectMemory(t *testing.T) {
	collector := NewGopsutilCollector()
	
	memInfo, err := collector.CollectMemory()
	if err != nil {
		t.Fatalf("CollectMemory failed: %v", err)
	}

	// Validate memory info structure
	if memInfo.Total == 0 {
		t.Error("Expected non-zero total memory")
	}

	if memInfo.Used > memInfo.Total {
		t.Errorf("Used memory (%d) cannot exceed total memory (%d)", memInfo.Used, memInfo.Total)
	}

	if memInfo.Available > memInfo.Total {
		t.Errorf("Available memory (%d) cannot exceed total memory (%d)", memInfo.Available, memInfo.Total)
	}

	// Validate swap info
	if memInfo.Swap.Used > memInfo.Swap.Total {
		t.Errorf("Used swap (%d) cannot exceed total swap (%d)", memInfo.Swap.Used, memInfo.Swap.Total)
	}

	// Validate timestamp is recent
	if time.Since(memInfo.Timestamp) > time.Minute {
		t.Errorf("Memory info timestamp is too old: %v", memInfo.Timestamp)
	}
}

func TestGopsutilCollector_CollectDisk(t *testing.T) {
	collector := NewGopsutilCollector()
	
	diskInfos, err := collector.CollectDisk()
	if err != nil {
		t.Fatalf("CollectDisk failed: %v", err)
	}

	// We should have at least one disk partition
	if len(diskInfos) == 0 {
		t.Skip("No disk partitions found, skipping disk collection test")
	}

	for i, diskInfo := range diskInfos {
		// Validate disk info structure
		if diskInfo.Device == "" {
			t.Errorf("Disk %d has empty device name", i)
		}

		if diskInfo.Mountpoint == "" {
			t.Errorf("Disk %d has empty mountpoint", i)
		}

		if diskInfo.Total == 0 {
			t.Errorf("Disk %d has zero total space", i)
		}

		if diskInfo.Used > diskInfo.Total {
			t.Errorf("Disk %d used space (%d) cannot exceed total space (%d)", i, diskInfo.Used, diskInfo.Total)
		}

		if diskInfo.Available > diskInfo.Total {
			t.Errorf("Disk %d available space (%d) cannot exceed total space (%d)", i, diskInfo.Available, diskInfo.Total)
		}

		if diskInfo.UsedPercent < 0 || diskInfo.UsedPercent > 100 {
			t.Errorf("Disk %d used percentage %f is not within valid range [0, 100]", i, diskInfo.UsedPercent)
		}
	}
}

func TestGopsutilCollector_CollectNetwork(t *testing.T) {
	collector := NewGopsutilCollector()
	
	networkInfos, err := collector.CollectNetwork()
	if err != nil {
		t.Fatalf("CollectNetwork failed: %v", err)
	}

	// We should have at least one network interface
	if len(networkInfos) == 0 {
		t.Skip("No network interfaces found, skipping network collection test")
	}

	for i, netInfo := range networkInfos {
		// Validate network info structure
		if netInfo.Interface == "" {
			t.Errorf("Network interface %d has empty name", i)
		}

		// Validate timestamp is recent
		if time.Since(netInfo.Timestamp) > time.Minute {
			t.Errorf("Network info %d timestamp is too old: %v", i, netInfo.Timestamp)
		}

		// Note: Bytes and packets can be zero for inactive interfaces, so we don't validate their values
	}
}

// TestGopsutilCollector_ImplementsInterface verifies that GopsutilCollector implements SystemCollector
func TestGopsutilCollector_ImplementsInterface(t *testing.T) {
	var _ models.SystemCollector = (*GopsutilCollector)(nil)
}

// TestGopsutilCollector_CollectDisk_ErrorHandling tests disk collection with enhanced error handling
func TestGopsutilCollector_CollectDisk_ErrorHandling(t *testing.T) {
	collector := NewGopsutilCollector()
	
	diskInfos, err := collector.CollectDisk()
	if err != nil {
		t.Fatalf("CollectDisk failed: %v", err)
	}

	// Test that we filter out special filesystems
	for _, diskInfo := range diskInfos {
		// Ensure we don't include special filesystems
		specialFS := []string{"proc", "sysfs", "devtmpfs", "tmpfs", "devpts", "cgroup", "cgroup2", "pstore", "bpf", "tracefs"}
		for _, fs := range specialFS {
			if diskInfo.Filesystem == fs {
				t.Errorf("Special filesystem %s should be filtered out", fs)
			}
		}

		// Validate that all returned disk info has valid data
		if diskInfo.Device == "" {
			t.Error("Device name should not be empty")
		}
		if diskInfo.Mountpoint == "" {
			t.Error("Mountpoint should not be empty")
		}
		if diskInfo.Total == 0 {
			t.Error("Total space should not be zero for real filesystems")
		}
	}
}

// TestGopsutilCollector_CollectNetwork_FilteredInterfaces tests network collection filtering
func TestGopsutilCollector_CollectNetwork_FilteredInterfaces(t *testing.T) {
	collector := NewGopsutilCollector()
	
	networkInfos, err := collector.CollectNetwork()
	if err != nil {
		t.Fatalf("CollectNetwork failed: %v", err)
	}

	// Test that loopback interfaces are filtered out
	for _, netInfo := range networkInfos {
		if netInfo.Interface == "lo" || netInfo.Interface == "Loopback" {
			t.Errorf("Loopback interface %s should be filtered out", netInfo.Interface)
		}
	}
}

// TestGopsutilCollector_CalculateNetworkRates tests network rate calculations
func TestGopsutilCollector_CalculateNetworkRates(t *testing.T) {
	collector := NewGopsutilCollector()
	
	// Create mock network data for testing rate calculations
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
			BytesRecv: 1500,
			Timestamp: baseTime,
		},
	}
	
	current := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 2000, // +1000 bytes
			BytesRecv: 4000, // +2000 bytes
			Timestamp: baseTime.Add(time.Second), // 1 second later
		},
		{
			Interface: "wlan0",
			BytesSent: 1000, // +500 bytes
			BytesRecv: 2500, // +1000 bytes
			Timestamp: baseTime.Add(time.Second), // 1 second later
		},
		{
			Interface: "eth1", // New interface not in previous
			BytesSent: 100,
			BytesRecv: 200,
			Timestamp: baseTime.Add(time.Second),
		},
	}
	
	rates := collector.CalculateNetworkRates(previous, current)
	
	// Test eth0 rates
	if eth0Rate, exists := rates["eth0"]; exists {
		expectedSendRate := 1000.0 // 1000 bytes per second
		expectedRecvRate := 2000.0 // 2000 bytes per second
		
		if eth0Rate.SendRate != expectedSendRate {
			t.Errorf("Expected eth0 send rate %f, got %f", expectedSendRate, eth0Rate.SendRate)
		}
		if eth0Rate.RecvRate != expectedRecvRate {
			t.Errorf("Expected eth0 recv rate %f, got %f", expectedRecvRate, eth0Rate.RecvRate)
		}
	} else {
		t.Error("Expected eth0 rate calculation")
	}
	
	// Test wlan0 rates
	if wlan0Rate, exists := rates["wlan0"]; exists {
		expectedSendRate := 500.0 // 500 bytes per second
		expectedRecvRate := 1000.0 // 1000 bytes per second
		
		if wlan0Rate.SendRate != expectedSendRate {
			t.Errorf("Expected wlan0 send rate %f, got %f", expectedSendRate, wlan0Rate.SendRate)
		}
		if wlan0Rate.RecvRate != expectedRecvRate {
			t.Errorf("Expected wlan0 recv rate %f, got %f", expectedRecvRate, wlan0Rate.RecvRate)
		}
	} else {
		t.Error("Expected wlan0 rate calculation")
	}
	
	// Test that eth1 (new interface) has no rate calculation
	if _, exists := rates["eth1"]; exists {
		t.Error("eth1 should not have rate calculation as it wasn't in previous measurement")
	}
}

// TestGopsutilCollector_CalculateNetworkRates_CounterRollover tests handling of counter rollover
func TestGopsutilCollector_CalculateNetworkRates_CounterRollover(t *testing.T) {
	collector := NewGopsutilCollector()
	
	baseTime := time.Now()
	
	// Simulate counter rollover (current < previous)
	previous := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 1000,
			BytesRecv: 2000,
			Timestamp: baseTime,
		},
	}
	
	current := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 500,  // Less than previous (rollover)
			BytesRecv: 1000, // Less than previous (rollover)
			Timestamp: baseTime.Add(time.Second),
		},
	}
	
	rates := collector.CalculateNetworkRates(previous, current)
	
	if eth0Rate, exists := rates["eth0"]; exists {
		// Rates should be 0 when rollover is detected
		if eth0Rate.SendRate != 0 {
			t.Errorf("Expected send rate 0 for counter rollover, got %f", eth0Rate.SendRate)
		}
		if eth0Rate.RecvRate != 0 {
			t.Errorf("Expected recv rate 0 for counter rollover, got %f", eth0Rate.RecvRate)
		}
	} else {
		t.Error("Expected eth0 rate calculation even with rollover")
	}
}

// TestGopsutilCollector_IntegrationTest_FullCycle tests complete data collection cycle
func TestGopsutilCollector_IntegrationTest_FullCycle(t *testing.T) {
	collector := NewGopsutilCollector()
	
	// Test complete data collection cycle
	cpuInfo, err := collector.CollectCPU()
	if err != nil {
		t.Errorf("CPU collection failed: %v", err)
	} else {
		t.Logf("Collected CPU info: %d cores, %.2f%% total usage", cpuInfo.Cores, cpuInfo.Total)
	}
	
	memInfo, err := collector.CollectMemory()
	if err != nil {
		t.Errorf("Memory collection failed: %v", err)
	} else {
		t.Logf("Collected memory info: %d total, %d used, %d available", memInfo.Total, memInfo.Used, memInfo.Available)
	}
	
	diskInfos, err := collector.CollectDisk()
	if err != nil {
		t.Errorf("Disk collection failed: %v", err)
	} else {
		t.Logf("Collected %d disk partitions", len(diskInfos))
		for _, disk := range diskInfos {
			t.Logf("  %s: %.1f%% used (%s)", disk.Mountpoint, disk.UsedPercent, disk.Filesystem)
		}
	}
	
	networkInfos, err := collector.CollectNetwork()
	if err != nil {
		t.Errorf("Network collection failed: %v", err)
	} else {
		t.Logf("Collected %d network interfaces", len(networkInfos))
		for _, net := range networkInfos {
			t.Logf("  %s: %d bytes sent, %d bytes received", net.Interface, net.BytesSent, net.BytesRecv)
		}
	}
	
	// Test network rate calculation with real data
	if len(networkInfos) > 0 {
		// Wait a bit and collect again for rate calculation
		time.Sleep(100 * time.Millisecond)
		
		networkInfos2, err := collector.CollectNetwork()
		if err != nil {
			t.Errorf("Second network collection failed: %v", err)
		} else {
			rates := collector.CalculateNetworkRates(networkInfos, networkInfos2)
			t.Logf("Calculated rates for %d interfaces", len(rates))
			for iface, rate := range rates {
				t.Logf("  %s: %.2f bytes/s sent, %.2f bytes/s received", iface, rate.SendRate, rate.RecvRate)
			}
		}
	}
}

// TestGopsutilCollector_ErrorRecovery tests error recovery scenarios
func TestGopsutilCollector_ErrorRecovery(t *testing.T) {
	collector := NewGopsutilCollector()
	
	// Test that partial failures don't prevent other data collection
	// This is more of a behavioral test to ensure the collector is robust
	
	// Collect all data types multiple times to test consistency
	for i := 0; i < 3; i++ {
		_, err := collector.CollectCPU()
		if err != nil {
			t.Logf("CPU collection attempt %d failed (may be expected): %v", i+1, err)
		}
		
		_, err = collector.CollectMemory()
		if err != nil {
			t.Logf("Memory collection attempt %d failed (may be expected): %v", i+1, err)
		}
		
		_, err = collector.CollectDisk()
		if err != nil {
			t.Logf("Disk collection attempt %d failed (may be expected): %v", i+1, err)
		}
		
		_, err = collector.CollectNetwork()
		if err != nil {
			t.Logf("Network collection attempt %d failed (may be expected): %v", i+1, err)
		}
		
		// Small delay between attempts
		time.Sleep(10 * time.Millisecond)
	}
}

// Error categorization tests

func TestGopsutilCollector_isPermissionError(t *testing.T) {
	collector := NewGopsutilCollector()
	
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"permission denied", errors.New("permission denied"), true},
		{"access denied", errors.New("access denied"), true},
		{"operation not permitted", errors.New("operation not permitted"), true},
		{"Permission Denied (case insensitive)", errors.New("Permission Denied"), true},
		{"other error", errors.New("some other error"), false},
		{"timeout error", errors.New("timeout occurred"), false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := collector.isPermissionError(tc.err)
			if result != tc.expected {
				t.Errorf("isPermissionError(%v) = %v, want %v", tc.err, result, tc.expected)
			}
		})
	}
}

func TestGopsutilCollector_isTemporaryError(t *testing.T) {
	collector := NewGopsutilCollector()
	
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"timeout", errors.New("timeout occurred"), true},
		{"temporary", errors.New("temporary failure"), true},
		{"try again", errors.New("try again later"), true},
		{"resource temporarily unavailable", errors.New("resource temporarily unavailable"), true},
		{"Timeout (case insensitive)", errors.New("Timeout"), true},
		{"permission error", errors.New("permission denied"), false},
		{"other error", errors.New("some other error"), false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := collector.isTemporaryError(tc.err)
			if result != tc.expected {
				t.Errorf("isTemporaryError(%v) = %v, want %v", tc.err, result, tc.expected)
			}
		})
	}
}

// Error handling integration tests

func TestGopsutilCollector_ErrorHandling_SystemErrorTypes(t *testing.T) {
	collector := NewGopsutilCollector()
	
	// Test that collector methods return SystemError types when they fail
	// Note: These tests may pass in normal environments, but verify error handling structure
	
	// Test CPU collection error handling
	_, err := collector.CollectCPU()
	if err != nil {
		if systemErr, ok := err.(models.SystemError); ok {
			if systemErr.Component != "CPU" {
				t.Errorf("Expected CPU component in error, got %s", systemErr.Component)
			}
			t.Logf("CPU error properly categorized: %v (Type: %v)", systemErr, systemErr.Type)
		} else {
			t.Errorf("Expected SystemError type for CPU collection error, got %T", err)
		}
	}
	
	// Test Memory collection error handling
	_, err = collector.CollectMemory()
	if err != nil {
		if systemErr, ok := err.(models.SystemError); ok {
			if systemErr.Component != "Memory" {
				t.Errorf("Expected Memory component in error, got %s", systemErr.Component)
			}
			t.Logf("Memory error properly categorized: %v (Type: %v)", systemErr, systemErr.Type)
		} else {
			t.Errorf("Expected SystemError type for Memory collection error, got %T", err)
		}
	}
	
	// Test Disk collection error handling
	_, err = collector.CollectDisk()
	if err != nil {
		if systemErr, ok := err.(models.SystemError); ok {
			if systemErr.Component != "Disk" {
				t.Errorf("Expected Disk component in error, got %s", systemErr.Component)
			}
			t.Logf("Disk error properly categorized: %v (Type: %v)", systemErr, systemErr.Type)
		} else {
			t.Errorf("Expected SystemError type for Disk collection error, got %T", err)
		}
	}
	
	// Test Network collection error handling
	_, err = collector.CollectNetwork()
	if err != nil {
		if systemErr, ok := err.(models.SystemError); ok {
			if systemErr.Component != "Network" {
				t.Errorf("Expected Network component in error, got %s", systemErr.Component)
			}
			t.Logf("Network error properly categorized: %v (Type: %v)", systemErr, systemErr.Type)
		} else {
			t.Errorf("Expected SystemError type for Network collection error, got %T", err)
		}
	}
}

func TestGopsutilCollector_ErrorHandling_PartialFailures(t *testing.T) {
	collector := NewGopsutilCollector()
	
	// Test that partial failures are handled gracefully
	// This tests the enhanced error handling in disk collection
	// where some partitions might fail but others succeed
	
	diskInfos, err := collector.CollectDisk()
	if err != nil {
		// If we get an error, it should be a SystemError
		if systemErr, ok := err.(models.SystemError); ok {
			t.Logf("Disk collection failed with categorized error: %v (Type: %v)", systemErr, systemErr.Type)
		} else {
			t.Errorf("Expected SystemError type for disk collection error, got %T", err)
		}
	} else {
		// If successful, we should have valid disk info
		if len(diskInfos) == 0 {
			t.Error("Expected at least some disk information or an error")
		}
		t.Logf("Successfully collected %d disk partitions", len(diskInfos))
	}
}

func TestGopsutilCollector_ErrorHandling_GracefulDegradation(t *testing.T) {
	collector := NewGopsutilCollector()
	
	// Test that the collector can handle scenarios where some data is available
	// but other data fails (graceful degradation)
	
	// Test memory collection with potential swap failure
	memInfo, err := collector.CollectMemory()
	if err != nil {
		if systemErr, ok := err.(models.SystemError); ok {
			t.Logf("Memory collection failed: %v (Type: %v)", systemErr, systemErr.Type)
		}
	} else {
		// Even if swap fails, we should get basic memory info
		if memInfo.Total == 0 {
			t.Error("Expected non-zero total memory even with partial failures")
		}
		
		// Swap might be 0 if not configured or if collection failed
		if memInfo.Swap.Total == 0 {
			t.Log("Swap information not available (may be expected)")
		}
		
		t.Logf("Memory collection succeeded: %d total, swap: %d total", memInfo.Total, memInfo.Swap.Total)
	}
}

func TestGopsutilCollector_ErrorHandling_LoggingIntegration(t *testing.T) {
	// Test that errors are properly logged through the error handler
	var logOutput strings.Builder
	logger := log.New(&logOutput, "", 0)
	errorHandler := models.NewErrorHandler(logger)
	collector := NewGopsutilCollectorWithErrorHandler(errorHandler)
	
	// Attempt to collect data - any errors should be logged
	collector.CollectCPU()
	collector.CollectMemory()
	collector.CollectDisk()
	collector.CollectNetwork()
	
	// Check if any errors were logged
	logContent := logOutput.String()
	if logContent != "" {
		t.Logf("Errors were logged (expected in some environments): %s", logContent)
		
		// Verify log format contains component information
		if strings.Contains(logContent, "CPU") || 
		   strings.Contains(logContent, "Memory") || 
		   strings.Contains(logContent, "Disk") || 
		   strings.Contains(logContent, "Network") {
			t.Log("Error logging includes component information as expected")
		} else {
			t.Error("Expected error logs to include component information")
		}
	} else {
		t.Log("No errors logged (system is functioning normally)")
	}
}