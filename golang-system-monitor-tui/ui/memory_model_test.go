package ui

import (
	"strings"
	"testing"
	"time"

	"golang-system-monitor-tui/models"
)

func TestNewMemoryModel(t *testing.T) {
	model := NewMemoryModel()

	if model.total != 0 {
		t.Errorf("Expected total to be 0, got %d", model.total)
	}
	if model.used != 0 {
		t.Errorf("Expected used to be 0, got %d", model.used)
	}
	if model.available != 0 {
		t.Errorf("Expected available to be 0, got %d", model.available)
	}
	if model.width != 40 {
		t.Errorf("Expected width to be 40, got %d", model.width)
	}
	if model.height != 8 {
		t.Errorf("Expected height to be 8, got %d", model.height)
	}
}

func TestMemoryModel_Init(t *testing.T) {
	model := NewMemoryModel()
	cmd := model.Init()

	if cmd != nil {
		t.Error("Expected Init to return nil command")
	}
}

func TestMemoryModel_Update(t *testing.T) {
	model := NewMemoryModel()
	timestamp := time.Now()

	// Test memory update message
	memoryInfo := models.MemoryInfo{
		Total:     16 * 1024 * 1024 * 1024, // 16GB
		Used:      8 * 1024 * 1024 * 1024,  // 8GB
		Available: 8 * 1024 * 1024 * 1024,  // 8GB
		Swap: models.SwapInfo{
			Total: 4 * 1024 * 1024 * 1024, // 4GB
			Used:  1 * 1024 * 1024 * 1024, // 1GB
			Free:  3 * 1024 * 1024 * 1024, // 3GB
		},
		Timestamp: timestamp,
	}

	updatedModel, cmd := model.Update(MemoryUpdateMsg(memoryInfo))

	if cmd != nil {
		t.Error("Expected Update to return nil command")
	}

	if updatedModel.total != memoryInfo.Total {
		t.Errorf("Expected total to be %d, got %d", memoryInfo.Total, updatedModel.total)
	}
	if updatedModel.used != memoryInfo.Used {
		t.Errorf("Expected used to be %d, got %d", memoryInfo.Used, updatedModel.used)
	}
	if updatedModel.available != memoryInfo.Available {
		t.Errorf("Expected available to be %d, got %d", memoryInfo.Available, updatedModel.available)
	}
	if updatedModel.swap.Total != memoryInfo.Swap.Total {
		t.Errorf("Expected swap total to be %d, got %d", memoryInfo.Swap.Total, updatedModel.swap.Total)
	}
	if updatedModel.lastUpdate != timestamp {
		t.Errorf("Expected lastUpdate to be %v, got %v", timestamp, updatedModel.lastUpdate)
	}
}

func TestMemoryModel_Update_UnknownMessage(t *testing.T) {
	model := NewMemoryModel()
	originalModel := model

	// Test with unknown message type
	updatedModel, cmd := model.Update("unknown message")

	if cmd != nil {
		t.Error("Expected Update to return nil command for unknown message")
	}

	// Model should remain unchanged
	if updatedModel.total != originalModel.total {
		t.Error("Model should not change for unknown message")
	}
}

func TestMemoryModel_View_Placeholder(t *testing.T) {
	model := NewMemoryModel()
	view := model.View()

	if !strings.Contains(view, "Memory Usage") {
		t.Error("Expected view to contain 'Memory Usage' header")
	}
	if !strings.Contains(view, "Loading memory data...") {
		t.Error("Expected view to contain loading placeholder")
	}
}

func TestMemoryModel_View_WithData(t *testing.T) {
	model := NewMemoryModel()
	
	// Update with test data
	memoryInfo := models.MemoryInfo{
		Total:     16 * 1024 * 1024 * 1024, // 16GB
		Used:      8 * 1024 * 1024 * 1024,  // 8GB
		Available: 8 * 1024 * 1024 * 1024,  // 8GB
		Swap: models.SwapInfo{
			Total: 4 * 1024 * 1024 * 1024, // 4GB
			Used:  1 * 1024 * 1024 * 1024, // 1GB
			Free:  3 * 1024 * 1024 * 1024, // 3GB
		},
		Timestamp: time.Now(),
	}
	
	model, _ = model.Update(MemoryUpdateMsg(memoryInfo))
	view := model.View()

	if !strings.Contains(view, "Memory Usage") {
		t.Error("Expected view to contain 'Memory Usage' header")
	}
	if !strings.Contains(view, "RAM:") {
		t.Error("Expected view to contain 'RAM:' section")
	}
	if !strings.Contains(view, "50.0%") {
		t.Error("Expected view to contain '50.0%' usage (8GB/16GB)")
	}
	if !strings.Contains(view, "16.0GB") {
		t.Error("Expected view to contain '16.0GB' total memory")
	}
	if !strings.Contains(view, "8.0GB") {
		t.Error("Expected view to contain '8.0GB' used memory")
	}
	if !strings.Contains(view, "Swap:") {
		t.Error("Expected view to contain 'Swap:' section")
	}
	if !strings.Contains(view, "25.0%") {
		t.Error("Expected view to contain '25.0%' swap usage (1GB/4GB)")
	}
}

func TestMemoryModel_View_NoSwap(t *testing.T) {
	model := NewMemoryModel()
	
	// Update with test data without swap
	memoryInfo := models.MemoryInfo{
		Total:     8 * 1024 * 1024 * 1024, // 8GB
		Used:      4 * 1024 * 1024 * 1024, // 4GB
		Available: 4 * 1024 * 1024 * 1024, // 4GB
		Swap: models.SwapInfo{
			Total: 0, // No swap
			Used:  0,
			Free:  0,
		},
		Timestamp: time.Now(),
	}
	
	model, _ = model.Update(MemoryUpdateMsg(memoryInfo))
	view := model.View()

	if !strings.Contains(view, "Swap: Not configured") {
		t.Error("Expected view to contain 'Swap: Not configured' when no swap is available")
	}
}

func TestMemoryModel_FormatBytes(t *testing.T) {
	model := NewMemoryModel()

	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1024 * 1024, "1.0MB"},
		{1536 * 1024, "1.5MB"},
		{1024 * 1024 * 1024, "1.0GB"},
		{1536 * 1024 * 1024, "1.5GB"},
		{1024 * 1024 * 1024 * 1024, "1.0TB"},
		{1536 * 1024 * 1024 * 1024, "1.5TB"},
	}

	for _, test := range tests {
		result := model.formatBytes(test.bytes)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", test.bytes, result, test.expected)
		}
	}
}

func TestMemoryModel_SetSize(t *testing.T) {
	model := NewMemoryModel()
	newModel := model.SetSize(80, 20)

	if newModel.width != 80 {
		t.Errorf("Expected width to be 80, got %d", newModel.width)
	}
	if newModel.height != 20 {
		t.Errorf("Expected height to be 20, got %d", newModel.height)
	}
}

func TestMemoryModel_Getters(t *testing.T) {
	model := NewMemoryModel()
	
	// Update with test data
	memoryInfo := models.MemoryInfo{
		Total:     16 * 1024 * 1024 * 1024, // 16GB
		Used:      8 * 1024 * 1024 * 1024,  // 8GB
		Available: 8 * 1024 * 1024 * 1024,  // 8GB
		Swap: models.SwapInfo{
			Total: 4 * 1024 * 1024 * 1024, // 4GB
			Used:  1 * 1024 * 1024 * 1024, // 1GB
			Free:  3 * 1024 * 1024 * 1024, // 3GB
		},
		Timestamp: time.Now(),
	}
	
	model, _ = model.Update(MemoryUpdateMsg(memoryInfo))

	if model.GetTotal() != memoryInfo.Total {
		t.Errorf("GetTotal() = %d, expected %d", model.GetTotal(), memoryInfo.Total)
	}
	if model.GetUsed() != memoryInfo.Used {
		t.Errorf("GetUsed() = %d, expected %d", model.GetUsed(), memoryInfo.Used)
	}
	if model.GetAvailable() != memoryInfo.Available {
		t.Errorf("GetAvailable() = %d, expected %d", model.GetAvailable(), memoryInfo.Available)
	}
	if model.GetSwap().Total != memoryInfo.Swap.Total {
		t.Errorf("GetSwap().Total = %d, expected %d", model.GetSwap().Total, memoryInfo.Swap.Total)
	}
}

func TestMemoryModel_GetUsagePercent(t *testing.T) {
	model := NewMemoryModel()

	// Test with no data
	if model.GetUsagePercent() != 0 {
		t.Errorf("Expected usage percent to be 0 with no data, got %f", model.GetUsagePercent())
	}

	// Test with data
	memoryInfo := models.MemoryInfo{
		Total:     1000,
		Used:      750,
		Available: 250,
		Timestamp: time.Now(),
	}
	
	model, _ = model.Update(MemoryUpdateMsg(memoryInfo))
	expectedPercent := 75.0
	actualPercent := model.GetUsagePercent()

	if actualPercent != expectedPercent {
		t.Errorf("GetUsagePercent() = %f, expected %f", actualPercent, expectedPercent)
	}
}

func TestMemoryModel_GetSwapUsagePercent(t *testing.T) {
	model := NewMemoryModel()

	// Test with no swap
	if model.GetSwapUsagePercent() != 0 {
		t.Errorf("Expected swap usage percent to be 0 with no swap, got %f", model.GetSwapUsagePercent())
	}

	// Test with swap data
	memoryInfo := models.MemoryInfo{
		Total:     1000,
		Used:      500,
		Available: 500,
		Swap: models.SwapInfo{
			Total: 400,
			Used:  100,
			Free:  300,
		},
		Timestamp: time.Now(),
	}
	
	model, _ = model.Update(MemoryUpdateMsg(memoryInfo))
	expectedPercent := 25.0
	actualPercent := model.GetSwapUsagePercent()

	if actualPercent != expectedPercent {
		t.Errorf("GetSwapUsagePercent() = %f, expected %f", actualPercent, expectedPercent)
	}
}

func TestMemoryModel_StyleManagerIntegration(t *testing.T) {
	model := NewMemoryModel()

	// Test that style manager is initialized
	if model.styleManager == nil {
		t.Error("Expected style manager to be initialized")
	}

	// Test progress bar rendering through style manager
	bar := model.styleManager.RenderProgressBar(50.0, 10, false)
	if !strings.Contains(bar, "█") {
		t.Error("Expected progress bar to contain filled characters")
	}
	if !strings.Contains(bar, "░") {
		t.Error("Expected progress bar to contain empty characters")
	}

	// Test with zero width
	bar = model.styleManager.RenderProgressBar(50.0, 0, false)
	if len(bar) == 0 {
		t.Error("Expected progress bar to have default width when width is 0")
	}

	// Test with 100% usage
	bar = model.styleManager.RenderProgressBar(100.0, 10, false)
	if strings.Contains(bar, "░") {
		t.Error("Expected no empty characters for 100% usage")
	}

	// Test with over 100% usage
	bar = model.styleManager.RenderProgressBar(150.0, 10, false)
	if strings.Contains(bar, "░") {
		t.Error("Expected no empty characters for over 100% usage")
	}
}

// Benchmark tests for performance validation
func BenchmarkMemoryModel_Update(b *testing.B) {
	model := NewMemoryModel()
	memoryInfo := models.MemoryInfo{
		Total:     16 * 1024 * 1024 * 1024,
		Used:      8 * 1024 * 1024 * 1024,
		Available: 8 * 1024 * 1024 * 1024,
		Swap: models.SwapInfo{
			Total: 4 * 1024 * 1024 * 1024,
			Used:  1 * 1024 * 1024 * 1024,
			Free:  3 * 1024 * 1024 * 1024,
		},
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.Update(MemoryUpdateMsg(memoryInfo))
	}
}

func BenchmarkMemoryModel_View(b *testing.B) {
	model := NewMemoryModel()
	memoryInfo := models.MemoryInfo{
		Total:     16 * 1024 * 1024 * 1024,
		Used:      8 * 1024 * 1024 * 1024,
		Available: 8 * 1024 * 1024 * 1024,
		Swap: models.SwapInfo{
			Total: 4 * 1024 * 1024 * 1024,
			Used:  1 * 1024 * 1024 * 1024,
			Free:  3 * 1024 * 1024 * 1024,
		},
		Timestamp: time.Now(),
	}
	model, _ = model.Update(MemoryUpdateMsg(memoryInfo))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.View()
	}
}

func BenchmarkMemoryModel_FormatBytes(b *testing.B) {
	model := NewMemoryModel()
	testBytes := uint64(16 * 1024 * 1024 * 1024) // 16GB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.formatBytes(testBytes)
	}
}