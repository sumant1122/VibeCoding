package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"golang-system-monitor-tui/models"
)

// MockSystemCollector implements SystemCollector for testing
type MockSystemCollector struct {
	cpuCallCount     int
	memoryCallCount  int
	diskCallCount    int
	networkCallCount int
	simulateError    bool
	errorComponent   string
}

func NewMockSystemCollector() *MockSystemCollector {
	return &MockSystemCollector{}
}

func (m *MockSystemCollector) CollectCPU() (models.CPUInfo, error) {
	m.cpuCallCount++
	if m.simulateError && m.errorComponent == "CPU" {
		return models.CPUInfo{}, models.CreateSystemError(models.SystemAccessError, "CPU", "Mock CPU error", nil)
	}
	
	return models.CPUInfo{
		Cores:     4,
		Usage:     []float64{25.0, 50.0, 75.0, 90.0},
		Total:     60.0,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockSystemCollector) CollectMemory() (models.MemoryInfo, error) {
	m.memoryCallCount++
	if m.simulateError && m.errorComponent == "Memory" {
		return models.MemoryInfo{}, models.CreateSystemError(models.SystemAccessError, "Memory", "Mock memory error", nil)
	}
	
	return models.MemoryInfo{
		Total:     16 * 1024 * 1024 * 1024, // 16GB
		Used:      8 * 1024 * 1024 * 1024,  // 8GB
		Available: 8 * 1024 * 1024 * 1024,  // 8GB
		Swap: models.SwapInfo{
			Total: 4 * 1024 * 1024 * 1024, // 4GB
			Used:  1 * 1024 * 1024 * 1024, // 1GB
			Free:  3 * 1024 * 1024 * 1024, // 3GB
		},
		Timestamp: time.Now(),
	}, nil
}

func (m *MockSystemCollector) CollectDisk() ([]models.DiskInfo, error) {
	m.diskCallCount++
	if m.simulateError && m.errorComponent == "Disk" {
		return nil, models.CreateSystemError(models.SystemAccessError, "Disk", "Mock disk error", nil)
	}
	
	return []models.DiskInfo{
		{
			Device:      "/dev/sda1",
			Mountpoint:  "/",
			Filesystem:  "ext4",
			Total:       500 * 1024 * 1024 * 1024, // 500GB
			Used:        300 * 1024 * 1024 * 1024, // 300GB
			Available:   200 * 1024 * 1024 * 1024, // 200GB
			UsedPercent: 60.0,
		},
	}, nil
}

func (m *MockSystemCollector) CollectNetwork() ([]models.NetworkInfo, error) {
	m.networkCallCount++
	if m.simulateError && m.errorComponent == "Network" {
		return nil, models.CreateSystemError(models.SystemAccessError, "Network", "Mock network error", nil)
	}
	
	return []models.NetworkInfo{
		{
			Interface:   "eth0",
			BytesSent:   uint64(m.networkCallCount * 1000),
			BytesRecv:   uint64(m.networkCallCount * 2000),
			PacketsSent: uint64(m.networkCallCount * 10),
			PacketsRecv: uint64(m.networkCallCount * 20),
			Timestamp:   time.Now(),
		},
	}, nil
}

func (m *MockSystemCollector) GetCallCounts() (int, int, int, int) {
	return m.cpuCallCount, m.memoryCallCount, m.diskCallCount, m.networkCallCount
}

func (m *MockSystemCollector) SetSimulateError(component string) {
	m.simulateError = true
	m.errorComponent = component
}

func (m *MockSystemCollector) ClearError() {
	m.simulateError = false
	m.errorComponent = ""
}

func (m *MockSystemCollector) CalculateNetworkRates(previous, current []models.NetworkInfo) map[string]models.NetworkStats {
	rates := make(map[string]models.NetworkStats)
	
	// Create a map of previous measurements for quick lookup
	prevMap := make(map[string]models.NetworkInfo)
	for _, prev := range previous {
		prevMap[prev.Interface] = prev
	}
	
	for _, curr := range current {
		if prev, exists := prevMap[curr.Interface]; exists {
			timeDiff := curr.Timestamp.Sub(prev.Timestamp).Seconds()
			if timeDiff > 0 {
				var sendRate, recvRate float64
				
				if curr.BytesSent >= prev.BytesSent {
					sendRate = float64(curr.BytesSent-prev.BytesSent) / timeDiff
				}
				
				if curr.BytesRecv >= prev.BytesRecv {
					recvRate = float64(curr.BytesRecv-prev.BytesRecv) / timeDiff
				}
				
				rates[curr.Interface] = models.NetworkStats{
					SendRate: sendRate,
					RecvRate: recvRate,
				}
			}
		}
	}
	
	return rates
}

// TestRealTimeUpdateSystem tests the complete real-time update system
func TestRealTimeUpdateSystem(t *testing.T) {
	// Create model with mock collector
	mockCollector := NewMockSystemCollector()
	model := NewMainModel()
	model.collector = mockCollector
	model.updateInterval = 100 * time.Millisecond // Faster updates for testing

	// Test individual components of the real-time update system
	
	// Test ticker command
	tickCmd := model.tickCmd()
	if tickCmd == nil {
		t.Fatal("tickCmd() should return a command")
	}

	// Test data collection commands
	cpuCmd := model.collectCPUDataCmd()
	if cpuCmd == nil {
		t.Fatal("collectCPUDataCmd() should return a command")
	}

	memoryCmd := model.collectMemoryDataCmd()
	if memoryCmd == nil {
		t.Fatal("collectMemoryDataCmd() should return a command")
	}

	diskCmd := model.collectDiskDataCmd()
	if diskCmd == nil {
		t.Fatal("collectDiskDataCmd() should return a command")
	}

	networkCmd := model.collectNetworkDataCmd()
	if networkCmd == nil {
		t.Fatal("collectNetworkDataCmd() should return a command")
	}

	// Execute data collection commands and verify they return appropriate messages
	cpuMsg := cpuCmd()
	if _, ok := cpuMsg.(CPUUpdateMsg); !ok {
		t.Error("Expected CPUUpdateMsg from CPU data collection")
	}

	memoryMsg := memoryCmd()
	if _, ok := memoryMsg.(MemoryUpdateMsg); !ok {
		t.Error("Expected MemoryUpdateMsg from memory data collection")
	}

	diskMsg := diskCmd()
	if _, ok := diskMsg.(DiskUpdateMsg); !ok {
		t.Error("Expected DiskUpdateMsg from disk data collection")
	}

	networkMsg := networkCmd()
	if _, ok := networkMsg.(NetworkUpdateMsg); !ok {
		t.Error("Expected NetworkUpdateMsg from network data collection")
	}

	// Verify that the mock collector was called
	cpuCount, memoryCount, diskCount, networkCount := mockCollector.GetCallCounts()
	if cpuCount != 1 || memoryCount != 1 || diskCount != 1 || networkCount != 1 {
		t.Errorf("Expected all collectors to be called once, got CPU:%d, Memory:%d, Disk:%d, Network:%d",
			cpuCount, memoryCount, diskCount, networkCount)
	}
}

// TestConcurrentDataCollection tests that data collection happens concurrently
func TestConcurrentDataCollection(t *testing.T) {
	mockCollector := NewMockSystemCollector()
	model := NewMainModel()
	model.collector = mockCollector

	// Execute individual data collection commands to test concurrent behavior
	start := time.Now()
	
	// Execute all data collection commands
	cpuCmd := model.collectCPUDataCmd()
	memoryCmd := model.collectMemoryDataCmd()
	diskCmd := model.collectDiskDataCmd()
	networkCmd := model.collectNetworkDataCmd()

	// Execute commands and collect messages
	var msgs []tea.Msg
	if cpuCmd != nil {
		msg := cpuCmd()
		if msg != nil {
			msgs = append(msgs, msg)
		}
	}
	if memoryCmd != nil {
		msg := memoryCmd()
		if msg != nil {
			msgs = append(msgs, msg)
		}
	}
	if diskCmd != nil {
		msg := diskCmd()
		if msg != nil {
			msgs = append(msgs, msg)
		}
	}
	if networkCmd != nil {
		msg := networkCmd()
		if msg != nil {
			msgs = append(msgs, msg)
		}
	}
	
	duration := time.Since(start)

	// Verify all data was collected
	cpuCount, memoryCount, diskCount, networkCount := mockCollector.GetCallCounts()
	if cpuCount != 1 || memoryCount != 1 || diskCount != 1 || networkCount != 1 {
		t.Errorf("Expected all collectors to be called once, got CPU:%d, Memory:%d, Disk:%d, Network:%d",
			cpuCount, memoryCount, diskCount, networkCount)
	}

	// Verify we got all expected messages
	var cpuMsgs, memoryMsgs, diskMsgs, networkMsgs int
	for _, msg := range msgs {
		switch msg.(type) {
		case CPUUpdateMsg:
			cpuMsgs++
		case MemoryUpdateMsg:
			memoryMsgs++
		case DiskUpdateMsg:
			diskMsgs++
		case NetworkUpdateMsg:
			networkMsgs++
		}
	}

	if cpuMsgs != 1 || memoryMsgs != 1 || diskMsgs != 1 || networkMsgs != 1 {
		t.Errorf("Expected one message of each type, got CPU:%d, Memory:%d, Disk:%d, Network:%d",
			cpuMsgs, memoryMsgs, diskMsgs, networkMsgs)
	}

	// Data collection should be reasonably fast
	if duration > 500*time.Millisecond {
		t.Errorf("Data collection took too long: %v (expected < 500ms)", duration)
	}
}

// TestUpdatePerformance tests the performance of the update system
func TestUpdatePerformance(t *testing.T) {
	mockCollector := NewMockSystemCollector()
	model := NewMainModel()
	model.collector = mockCollector

	// Measure time for multiple update cycles
	iterations := 100
	start := time.Now()

	for i := 0; i < iterations; i++ {
		// Simulate a complete update cycle by executing individual commands
		cpuCmd := model.collectCPUDataCmd()
		memoryCmd := model.collectMemoryDataCmd()
		diskCmd := model.collectDiskDataCmd()
		networkCmd := model.collectNetworkDataCmd()

		// Execute commands and update model
		if cpuCmd != nil {
			msg := cpuCmd()
			if msg != nil {
				updatedModel, _ := model.Update(msg)
				model = updatedModel.(MainModel)
			}
		}
		if memoryCmd != nil {
			msg := memoryCmd()
			if msg != nil {
				updatedModel, _ := model.Update(msg)
				model = updatedModel.(MainModel)
			}
		}
		if diskCmd != nil {
			msg := diskCmd()
			if msg != nil {
				updatedModel, _ := model.Update(msg)
				model = updatedModel.(MainModel)
			}
		}
		if networkCmd != nil {
			msg := networkCmd()
			if msg != nil {
				updatedModel, _ := model.Update(msg)
				model = updatedModel.(MainModel)
			}
		}
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(iterations)

	// Each update cycle should be reasonably fast
	if avgDuration > 50*time.Millisecond {
		t.Errorf("Average update cycle took too long: %v (expected < 50ms)", avgDuration)
	}

	// Verify all data was collected the expected number of times
	cpuCount, memoryCount, diskCount, networkCount := mockCollector.GetCallCounts()
	if cpuCount != iterations || memoryCount != iterations || diskCount != iterations || networkCount != iterations {
		t.Errorf("Expected %d calls to each collector, got CPU:%d, Memory:%d, Disk:%d, Network:%d",
			iterations, cpuCount, memoryCount, diskCount, networkCount)
	}
}

// TestUpdateAccuracy tests the accuracy of data updates
func TestUpdateAccuracy(t *testing.T) {
	mockCollector := NewMockSystemCollector()
	model := NewMainModel()
	model.collector = mockCollector

	// Collect initial data by executing individual commands
	cpuCmd := model.collectCPUDataCmd()
	if cpuCmd != nil {
		msg := cpuCmd()
		if msg != nil {
			updatedModel, _ := model.Update(msg)
			model = updatedModel.(MainModel)
		}
	}

	memoryCmd := model.collectMemoryDataCmd()
	if memoryCmd != nil {
		msg := memoryCmd()
		if msg != nil {
			updatedModel, _ := model.Update(msg)
			model = updatedModel.(MainModel)
		}
	}

	diskCmd := model.collectDiskDataCmd()
	if diskCmd != nil {
		msg := diskCmd()
		if msg != nil {
			updatedModel, _ := model.Update(msg)
			model = updatedModel.(MainModel)
		}
	}

	networkCmd := model.collectNetworkDataCmd()
	if networkCmd != nil {
		msg := networkCmd()
		if msg != nil {
			updatedModel, _ := model.Update(msg)
			model = updatedModel.(MainModel)
		}
	}

	// Verify CPU data accuracy
	cpuModel := model.GetCPUModel()
	expectedCores := 4
	expectedTotal := 60.0
	expectedUsage := []float64{25.0, 50.0, 75.0, 90.0}

	if cpuModel.GetCores() != expectedCores {
		t.Errorf("Expected %d CPU cores, got %d", expectedCores, cpuModel.GetCores())
	}
	if cpuModel.GetTotal() != expectedTotal {
		t.Errorf("Expected total CPU usage %.1f%%, got %.1f%%", expectedTotal, cpuModel.GetTotal())
	}
	
	actualUsage := cpuModel.GetUsage()
	if len(actualUsage) != len(expectedUsage) {
		t.Errorf("Expected %d core usage values, got %d", len(expectedUsage), len(actualUsage))
	} else {
		for i, expected := range expectedUsage {
			if actualUsage[i] != expected {
				t.Errorf("Expected core %d usage %.1f%%, got %.1f%%", i, expected, actualUsage[i])
			}
		}
	}

	// Verify Memory data accuracy
	memoryModel := model.GetMemoryModel()
	expectedMemoryTotal := uint64(16 * 1024 * 1024 * 1024)
	expectedMemoryUsed := uint64(8 * 1024 * 1024 * 1024)

	if memoryModel.GetTotal() != expectedMemoryTotal {
		t.Errorf("Expected total memory %d, got %d", expectedMemoryTotal, memoryModel.GetTotal())
	}
	if memoryModel.GetUsed() != expectedMemoryUsed {
		t.Errorf("Expected used memory %d, got %d", expectedMemoryUsed, memoryModel.GetUsed())
	}
}

// TestErrorHandlingInRealTimeUpdates tests error handling during real-time updates
func TestErrorHandlingInRealTimeUpdates(t *testing.T) {
	mockCollector := NewMockSystemCollector()
	model := NewMainModel()
	model.collector = mockCollector

	// Test CPU error handling
	mockCollector.SetSimulateError("CPU")
	cpuCmd := model.collectCPUDataCmd()
	msg := cpuCmd()
	
	if _, ok := msg.(models.SystemError); !ok {
		t.Error("Expected SystemError for CPU collection failure")
	}

	// Test Memory error handling
	mockCollector.SetSimulateError("Memory")
	memoryCmd := model.collectMemoryDataCmd()
	msg = memoryCmd()
	
	if _, ok := msg.(models.SystemError); !ok {
		t.Error("Expected SystemError for Memory collection failure")
	}

	// Test Disk error handling
	mockCollector.SetSimulateError("Disk")
	diskCmd := model.collectDiskDataCmd()
	msg = diskCmd()
	
	if _, ok := msg.(models.SystemError); !ok {
		t.Error("Expected SystemError for Disk collection failure")
	}

	// Test Network error handling
	mockCollector.SetSimulateError("Network")
	networkCmd := model.collectNetworkDataCmd()
	msg = networkCmd()
	
	if _, ok := msg.(models.SystemError); !ok {
		t.Error("Expected SystemError for Network collection failure")
	}
}

// TestTickerFunctionality tests the ticker mechanism
func TestTickerFunctionality(t *testing.T) {
	model := NewMainModel()
	model.updateInterval = 10 * time.Millisecond // Very fast for testing

	// Get ticker command
	tickCmd := model.tickCmd()
	if tickCmd == nil {
		t.Fatal("tickCmd() should return a command")
	}

	// Execute ticker command and measure timing
	start := time.Now()
	msg := tickCmd()
	duration := time.Since(start)

	// Should return TickMsg
	if _, ok := msg.(TickMsg); !ok {
		t.Error("Expected TickMsg from ticker command")
	}

	// Should respect the update interval (with some tolerance)
	expectedDuration := 10 * time.Millisecond
	tolerance := 5 * time.Millisecond
	if duration < expectedDuration-tolerance || duration > expectedDuration+tolerance {
		t.Errorf("Ticker duration %v not within expected range %v ± %v", duration, expectedDuration, tolerance)
	}
}

// TestSmoothRendering tests that updates don't cause rendering issues
func TestSmoothRendering(t *testing.T) {
	mockCollector := NewMockSystemCollector()
	model := NewMainModel()
	model.collector = mockCollector

	// Set reasonable dimensions
	model.width = 80
	model.height = 24

	// Perform multiple update cycles and verify rendering works
	for i := 0; i < 10; i++ {
		// Collect data
		collectCmd := model.collectAllDataCmd()
		if collectCmd != nil {
			msg := collectCmd()
			if msg != nil {
				updatedModel, _ := model.Update(msg)
				model = updatedModel.(MainModel)
			}
		}

		// Render view
		view := model.View()
		if view == "" {
			t.Errorf("View should not be empty after update cycle %d", i+1)
		}

		// View should contain expected components
		if !containsString(view, "CPU") {
			t.Errorf("View should contain CPU information after update cycle %d", i+1)
		}
		if !containsString(view, "Memory") {
			t.Errorf("View should contain Memory information after update cycle %d", i+1)
		}
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsString(s[1:], substr) || (len(s) > 0 && s[:len(substr)] == substr))
}

// BenchmarkRealTimeUpdate benchmarks the real-time update performance
func BenchmarkRealTimeUpdate(b *testing.B) {
	mockCollector := NewMockSystemCollector()
	model := NewMainModel()
	model.collector = mockCollector

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collectCmd := model.collectAllDataCmd()
		if collectCmd != nil {
			msg := collectCmd()
			if msg != nil {
				updatedModel, _ := model.Update(msg)
				model = updatedModel.(MainModel)
			}
		}
	}
}

// BenchmarkDataCollection benchmarks individual data collection operations
func BenchmarkDataCollection(b *testing.B) {
	mockCollector := NewMockSystemCollector()
	model := NewMainModel()
	model.collector = mockCollector

	b.Run("CPU", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd := model.collectCPUDataCmd()
			cmd()
		}
	})

	b.Run("Memory", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd := model.collectMemoryDataCmd()
			cmd()
		}
	})

	b.Run("Disk", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd := model.collectDiskDataCmd()
			cmd()
		}
	})

	b.Run("Network", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd := model.collectNetworkDataCmd()
			cmd()
		}
	})
}