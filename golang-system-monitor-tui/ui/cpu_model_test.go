package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"golang-system-monitor-tui/models"
)

func TestNewCPUModel(t *testing.T) {
	model := NewCPUModel()

	if len(model.usage) != 0 {
		t.Errorf("Expected empty usage slice, got %v", model.usage)
	}
	if len(model.history) != 0 {
		t.Errorf("Expected empty history slice, got %v", model.history)
	}
	if model.total != 0.0 {
		t.Errorf("Expected total usage to be 0.0, got %f", model.total)
	}
	if model.cores != 0 {
		t.Errorf("Expected cores to be 0, got %d", model.cores)
	}
	if model.maxHistory != 60 {
		t.Errorf("Expected maxHistory to be 60, got %d", model.maxHistory)
	}
	if model.width != 40 {
		t.Errorf("Expected width to be 40, got %d", model.width)
	}
	if model.height != 10 {
		t.Errorf("Expected height to be 10, got %d", model.height)
	}
}

func TestCPUModel_Init(t *testing.T) {
	model := NewCPUModel()
	cmd := model.Init()

	if cmd != nil {
		t.Errorf("Expected Init() to return nil, got %v", cmd)
	}
}

func TestCPUModel_Update_CPUUpdateMsg(t *testing.T) {
	model := NewCPUModel()
	timestamp := time.Now()

	cpuInfo := models.CPUInfo{
		Cores:     4,
		Usage:     []float64{25.5, 50.0, 75.2, 90.1},
		Total:     60.2,
		Timestamp: timestamp,
	}

	updateMsg := CPUUpdateMsg(cpuInfo)
	updatedModel, cmd := model.Update(updateMsg)

	if cmd != nil {
		t.Errorf("Expected Update() to return nil cmd, got %v", cmd)
	}

	if len(updatedModel.usage) != 4 {
		t.Errorf("Expected 4 core usage values, got %d", len(updatedModel.usage))
	}

	expectedUsage := []float64{25.5, 50.0, 75.2, 90.1}
	for i, expected := range expectedUsage {
		if updatedModel.usage[i] != expected {
			t.Errorf("Expected usage[%d] to be %f, got %f", i, expected, updatedModel.usage[i])
		}
	}

	if updatedModel.total != 60.2 {
		t.Errorf("Expected total to be 60.2, got %f", updatedModel.total)
	}

	if updatedModel.cores != 4 {
		t.Errorf("Expected cores to be 4, got %d", updatedModel.cores)
	}

	if !updatedModel.lastUpdate.Equal(timestamp) {
		t.Errorf("Expected lastUpdate to be %v, got %v", timestamp, updatedModel.lastUpdate)
	}
}

func TestCPUModel_Update_HistoryTracking(t *testing.T) {
	model := NewCPUModel()

	// First update
	cpuInfo1 := models.CPUInfo{
		Cores:     2,
		Usage:     []float64{30.0, 40.0},
		Total:     35.0,
		Timestamp: time.Now(),
	}
	model, _ = model.Update(CPUUpdateMsg(cpuInfo1))

	// Check history initialization
	if len(model.history) != 2 {
		t.Errorf("Expected history to have 2 cores, got %d", len(model.history))
	}

	if len(model.history[0]) != 1 || model.history[0][0] != 30.0 {
		t.Errorf("Expected first core history to contain [30.0], got %v", model.history[0])
	}

	if len(model.history[1]) != 1 || model.history[1][0] != 40.0 {
		t.Errorf("Expected second core history to contain [40.0], got %v", model.history[1])
	}

	// Second update
	cpuInfo2 := models.CPUInfo{
		Cores:     2,
		Usage:     []float64{35.0, 45.0},
		Total:     40.0,
		Timestamp: time.Now(),
	}
	model, _ = model.Update(CPUUpdateMsg(cpuInfo2))

	// Check history accumulation
	if len(model.history[0]) != 2 {
		t.Errorf("Expected first core history to have 2 entries, got %d", len(model.history[0]))
	}

	expectedHistory0 := []float64{30.0, 35.0}
	for i, expected := range expectedHistory0 {
		if model.history[0][i] != expected {
			t.Errorf("Expected history[0][%d] to be %f, got %f", i, expected, model.history[0][i])
		}
	}
}

func TestCPUModel_Update_HistoryLimit(t *testing.T) {
	model := NewCPUModel()
	model.maxHistory = 3 // Set small limit for testing

	// Add more updates than the limit
	for i := 0; i < 5; i++ {
		cpuInfo := models.CPUInfo{
			Cores:     1,
			Usage:     []float64{float64(i * 10)},
			Total:     float64(i * 10),
			Timestamp: time.Now(),
		}
		model, _ = model.Update(CPUUpdateMsg(cpuInfo))
	}

	// Check that history is limited
	if len(model.history[0]) != 3 {
		t.Errorf("Expected history to be limited to 3 entries, got %d", len(model.history[0]))
	}

	// Check that we kept the most recent entries
	expectedHistory := []float64{20.0, 30.0, 40.0}
	for i, expected := range expectedHistory {
		if model.history[0][i] != expected {
			t.Errorf("Expected history[0][%d] to be %f, got %f", i, expected, model.history[0][i])
		}
	}
}

func TestCPUModel_Update_OtherMessages(t *testing.T) {
	model := NewCPUModel()
	originalModel := model

	// Test with a different message type
	otherMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	updatedModel, cmd := model.Update(otherMsg)

	if cmd != nil {
		t.Errorf("Expected Update() to return nil cmd for non-CPU messages, got %v", cmd)
	}

	// Model should remain unchanged
	if len(updatedModel.usage) != len(originalModel.usage) {
		t.Errorf("Expected model to remain unchanged for non-CPU messages")
	}
}

func TestCPUModel_View_NoData(t *testing.T) {
	model := NewCPUModel()
	view := model.View()

	if !strings.Contains(view, "CPU Usage") {
		t.Errorf("Expected view to contain 'CPU Usage' header")
	}

	if !strings.Contains(view, "Loading CPU data...") {
		t.Errorf("Expected view to contain loading message when no data available")
	}
}

func TestCPUModel_View_WithData(t *testing.T) {
	model := NewCPUModel()

	// Add CPU data
	cpuInfo := models.CPUInfo{
		Cores:     2,
		Usage:     []float64{45.5, 78.2},
		Total:     61.85,
		Timestamp: time.Now(),
	}
	model, _ = model.Update(CPUUpdateMsg(cpuInfo))

	view := model.View()

	if !strings.Contains(view, "CPU Usage") {
		t.Errorf("Expected view to contain 'CPU Usage' header")
	}

	if !strings.Contains(view, "Total:") {
		t.Errorf("Expected view to contain total CPU usage")
	}

	if !strings.Contains(view, "61.9%") {
		t.Errorf("Expected view to contain total percentage, got: %s", view)
	}

	if !strings.Contains(view, "Core 1:") {
		t.Errorf("Expected view to contain Core 1 information")
	}

	if !strings.Contains(view, "Core 2:") {
		t.Errorf("Expected view to contain Core 2 information")
	}

	if !strings.Contains(view, "45.5%") {
		t.Errorf("Expected view to contain first core percentage")
	}

	if !strings.Contains(view, "78.2%") {
		t.Errorf("Expected view to contain second core percentage")
	}
}

func TestCPUModel_StyleManagerIntegration(t *testing.T) {
	model := NewCPUModel()

	// Test that style manager is initialized
	if model.styleManager == nil {
		t.Error("Expected style manager to be initialized")
	}

	// Test progress bar rendering through style manager
	bar := model.styleManager.RenderProgressBar(50.0, 20, false)
	if len(bar) == 0 {
		t.Error("Expected non-empty progress bar from style manager")
	}

	// Test that progress bars contain expected characters
	if !strings.Contains(bar, "█") && !strings.Contains(bar, "░") {
		t.Error("Expected progress bar to contain progress characters")
	}
}

func TestCPUModel_SetSize(t *testing.T) {
	model := NewCPUModel()
	newModel := model.SetSize(80, 20)

	if newModel.width != 80 {
		t.Errorf("Expected width to be 80, got %d", newModel.width)
	}

	if newModel.height != 20 {
		t.Errorf("Expected height to be 20, got %d", newModel.height)
	}
}

func TestCPUModel_Getters(t *testing.T) {
	model := NewCPUModel()

	// Add some data
	cpuInfo := models.CPUInfo{
		Cores:     3,
		Usage:     []float64{10.0, 20.0, 30.0},
		Total:     20.0,
		Timestamp: time.Now(),
	}
	model, _ = model.Update(CPUUpdateMsg(cpuInfo))

	// Test GetUsage
	usage := model.GetUsage()
	expectedUsage := []float64{10.0, 20.0, 30.0}
	if len(usage) != len(expectedUsage) {
		t.Errorf("Expected usage length %d, got %d", len(expectedUsage), len(usage))
	}
	for i, expected := range expectedUsage {
		if usage[i] != expected {
			t.Errorf("Expected usage[%d] to be %f, got %f", i, expected, usage[i])
		}
	}

	// Test GetTotal
	if model.GetTotal() != 20.0 {
		t.Errorf("Expected total to be 20.0, got %f", model.GetTotal())
	}

	// Test GetCores
	if model.GetCores() != 3 {
		t.Errorf("Expected cores to be 3, got %d", model.GetCores())
	}

	// Test GetHistory
	history := model.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected history to have 3 cores, got %d", len(history))
	}
	if len(history[0]) != 1 || history[0][0] != 10.0 {
		t.Errorf("Expected first core history to contain [10.0], got %v", history[0])
	}
}

func TestCPUModel_ProgressBarColors(t *testing.T) {
	model := NewCPUModel()

	// Test different usage levels to ensure color coding works through style manager
	testCases := []struct {
		percentage float64
		name       string
	}{
		{30.0, "normal"},
		{75.0, "warning"},
		{95.0, "critical"},
	}

	for _, tc := range testCases {
		bar := model.styleManager.RenderProgressBar(tc.percentage, 20, false)
		if len(bar) == 0 {
			t.Errorf("Expected non-empty progress bar for %s usage", tc.name)
		}
		// We can't easily test colors in unit tests, but we ensure the function doesn't crash
	}
}

// Error handling tests

func TestCPUModel_ErrorHandling_InitialState(t *testing.T) {
	model := NewCPUModel()

	if model.HasError() {
		t.Error("Expected HasError() to return false initially")
	}

	if model.GetErrorMessage() != "" {
		t.Errorf("Expected empty error message initially, got %s", model.GetErrorMessage())
	}
}

func TestCPUModel_ErrorHandling_SetError(t *testing.T) {
	model := NewCPUModel()
	errorMessage := "Test error message"

	model = model.SetError(errorMessage)

	if !model.HasError() {
		t.Error("Expected HasError() to return true after SetError()")
	}

	if model.GetErrorMessage() != errorMessage {
		t.Errorf("Expected error message '%s', got '%s'", errorMessage, model.GetErrorMessage())
	}

	if model.lastError.IsZero() {
		t.Error("Expected lastError timestamp to be set")
	}
}

func TestCPUModel_ErrorHandling_ClearError(t *testing.T) {
	model := NewCPUModel()
	model = model.SetError("Test error")

	// Verify error is set
	if !model.HasError() {
		t.Error("Expected error to be set before clearing")
	}

	model = model.ClearError()

	if model.HasError() {
		t.Error("Expected HasError() to return false after ClearError()")
	}

	if model.GetErrorMessage() != "" {
		t.Errorf("Expected empty error message after ClearError(), got %s", model.GetErrorMessage())
	}
}

func TestCPUModel_ErrorHandling_UpdateWithErrorMsg(t *testing.T) {
	model := NewCPUModel()
	
	// Create an error message for CPU component
	errorMsg := models.ErrorMsg{
		Type:      models.SystemAccessError,
		Message:   "Failed to access CPU data",
		Component: "CPU",
		Timestamp: time.Now(),
		Original:  nil,
	}

	updatedModel, cmd := model.Update(errorMsg)

	if cmd != nil {
		t.Errorf("Expected Update() to return nil cmd for error messages, got %v", cmd)
	}

	if !updatedModel.HasError() {
		t.Error("Expected model to have error after receiving ErrorMsg")
	}

	if updatedModel.GetErrorMessage() != "Failed to access CPU data" {
		t.Errorf("Expected error message 'Failed to access CPU data', got '%s'", updatedModel.GetErrorMessage())
	}
}

func TestCPUModel_ErrorHandling_UpdateWithNonCPUErrorMsg(t *testing.T) {
	model := NewCPUModel()
	
	// Create an error message for different component
	errorMsg := models.ErrorMsg{
		Type:      models.SystemAccessError,
		Message:   "Failed to access Memory data",
		Component: "Memory",
		Timestamp: time.Now(),
		Original:  nil,
	}

	updatedModel, cmd := model.Update(errorMsg)

	if cmd != nil {
		t.Errorf("Expected Update() to return nil cmd for error messages, got %v", cmd)
	}

	// Should not set error for non-CPU error messages
	if updatedModel.HasError() {
		t.Error("Expected model to not have error for non-CPU error messages")
	}
}

func TestCPUModel_ErrorHandling_ClearErrorOnSuccessfulUpdate(t *testing.T) {
	model := NewCPUModel()
	
	// Set an error first
	model = model.SetError("Previous error")
	if !model.HasError() {
		t.Error("Expected error to be set initially")
	}

	// Send successful CPU update
	cpuInfo := models.CPUInfo{
		Cores:     2,
		Usage:     []float64{25.0, 35.0},
		Total:     30.0,
		Timestamp: time.Now(),
	}

	updatedModel, _ := model.Update(CPUUpdateMsg(cpuInfo))

	// Error should be cleared on successful update
	if updatedModel.HasError() {
		t.Error("Expected error to be cleared on successful CPU update")
	}

	if updatedModel.GetErrorMessage() != "" {
		t.Errorf("Expected empty error message after successful update, got '%s'", updatedModel.GetErrorMessage())
	}
}

func TestCPUModel_ErrorHandling_ViewWithError(t *testing.T) {
	model := NewCPUModel()
	model = model.SetError("CPU access denied")

	view := model.View()

	if !strings.Contains(view, "CPU Usage") {
		t.Error("Expected view to contain 'CPU Usage' header even with error")
	}

	if !strings.Contains(view, "Error: CPU access denied") {
		t.Error("Expected view to contain error message")
	}

	if !strings.Contains(view, "CPU data unavailable") {
		t.Error("Expected view to contain unavailable message")
	}

	if !strings.Contains(view, "Total: N/A") {
		t.Error("Expected view to contain N/A fallback for total")
	}

	if !strings.Contains(view, "Cores: N/A") {
		t.Error("Expected view to contain N/A fallback for cores")
	}

	// Should not contain actual CPU data
	if strings.Contains(view, "Core 1:") {
		t.Error("Expected view to not contain actual core data when in error state")
	}
}

func TestCPUModel_ErrorHandling_ViewWithErrorAndData(t *testing.T) {
	model := NewCPUModel()
	
	// Add CPU data first
	cpuInfo := models.CPUInfo{
		Cores:     2,
		Usage:     []float64{45.5, 78.2},
		Total:     61.85,
		Timestamp: time.Now(),
	}
	model, _ = model.Update(CPUUpdateMsg(cpuInfo))

	// Then set an error
	model = model.SetError("Subsequent error")

	view := model.View()

	// Should show error state, not the data
	if !strings.Contains(view, "Error: Subsequent error") {
		t.Error("Expected view to show error message")
	}

	if !strings.Contains(view, "Total: N/A") {
		t.Error("Expected view to show N/A fallback instead of actual data")
	}

	// Should not show actual data when in error state
	if strings.Contains(view, "61.9%") {
		t.Error("Expected view to not show actual CPU percentage when in error state")
	}
}