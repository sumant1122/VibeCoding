package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewMainModel(t *testing.T) {
	model := NewMainModel()

	// Test initial state
	if model.focused != FocusCPU {
		t.Errorf("Expected initial focus to be FocusCPU, got %v", model.focused)
	}

	if model.showHelp {
		t.Error("Expected showHelp to be false initially")
	}

	if model.width != 80 || model.height != 24 {
		t.Errorf("Expected default dimensions 80x24, got %dx%d", model.width, model.height)
	}

	// Test that all component models are initialized
	if model.cpu.GetCores() < 0 {
		t.Error("CPU model not properly initialized")
	}
}

func TestMainModelInit(t *testing.T) {
	model := NewMainModel()
	cmd := model.Init()

	// Init should return a batch command (which may be nil if no sub-commands)
	// The important thing is that it doesn't panic and returns successfully
	_ = cmd // We don't need to check if it's nil since batch commands can be nil
}

func TestMainModelKeyboardNavigation(t *testing.T) {
	model := NewMainModel()

	tests := []struct {
		name        string
		key         string
		initialFocus FocusedComponent
		expectedFocus FocusedComponent
	}{
		{"Tab from CPU", "tab", FocusCPU, FocusMemory},
		{"Tab from Memory", "tab", FocusMemory, FocusDisk},
		{"Tab from Disk", "tab", FocusDisk, FocusNetwork},
		{"Tab from Network", "tab", FocusNetwork, FocusCPU},
		
		{"Shift+Tab from CPU", "shift+tab", FocusCPU, FocusNetwork},
		{"Shift+Tab from Memory", "shift+tab", FocusMemory, FocusCPU},
		{"Shift+Tab from Disk", "shift+tab", FocusDisk, FocusMemory},
		{"Shift+Tab from Network", "shift+tab", FocusNetwork, FocusDisk},
		
		{"Right arrow from CPU", "right", FocusCPU, FocusMemory},
		{"Right arrow from Network", "right", FocusNetwork, FocusCPU},
		
		{"Left arrow from Memory", "left", FocusMemory, FocusCPU},
		{"Left arrow from CPU", "left", FocusCPU, FocusNetwork},
		
		{"Down arrow from CPU", "down", FocusCPU, FocusDisk},
		{"Down arrow from Memory", "down", FocusMemory, FocusNetwork},
		{"Down arrow from Disk", "down", FocusDisk, FocusDisk}, // Should stay
		{"Down arrow from Network", "down", FocusNetwork, FocusNetwork}, // Should stay
		
		{"Up arrow from Disk", "up", FocusDisk, FocusCPU},
		{"Up arrow from Network", "up", FocusNetwork, FocusMemory},
		{"Up arrow from CPU", "up", FocusCPU, FocusCPU}, // Should stay
		{"Up arrow from Memory", "up", FocusMemory, FocusMemory}, // Should stay
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.focused = tt.initialFocus
			
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "tab" {
				keyMsg = tea.KeyMsg{Type: tea.KeyTab}
			} else if tt.key == "shift+tab" {
				keyMsg = tea.KeyMsg{Type: tea.KeyShiftTab}
			} else if tt.key == "right" {
				keyMsg = tea.KeyMsg{Type: tea.KeyRight}
			} else if tt.key == "left" {
				keyMsg = tea.KeyMsg{Type: tea.KeyLeft}
			} else if tt.key == "down" {
				keyMsg = tea.KeyMsg{Type: tea.KeyDown}
			} else if tt.key == "up" {
				keyMsg = tea.KeyMsg{Type: tea.KeyUp}
			}

			updatedModel, _ := model.Update(keyMsg)
			mainModel := updatedModel.(MainModel)

			if mainModel.focused != tt.expectedFocus {
				t.Errorf("Expected focus %v, got %v", tt.expectedFocus, mainModel.focused)
			}
		})
	}
}

func TestMainModelQuitKeys(t *testing.T) {
	model := NewMainModel()

	quitKeys := []string{"q", "ctrl+c"}
	
	for _, key := range quitKeys {
		t.Run("Quit with "+key, func(t *testing.T) {
			var keyMsg tea.KeyMsg
			if key == "q" {
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
			} else if key == "ctrl+c" {
				keyMsg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}

			_, cmd := model.Update(keyMsg)
			
			// Should return tea.Quit command
			if cmd == nil {
				t.Error("Expected quit command, got nil")
			}
		})
	}
}

func TestMainModelHelpToggle(t *testing.T) {
	model := NewMainModel()

	// Test help toggle with '?'
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	updatedModel, _ := model.Update(keyMsg)
	mainModel := updatedModel.(MainModel)

	if !mainModel.showHelp {
		t.Error("Expected help to be shown after pressing '?'")
	}

	// Test help toggle again to hide
	updatedModel, _ = mainModel.Update(keyMsg)
	mainModel = updatedModel.(MainModel)

	if mainModel.showHelp {
		t.Error("Expected help to be hidden after pressing '?' again")
	}

	// Test help toggle with 'h'
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")}
	updatedModel, _ = model.Update(keyMsg)
	mainModel = updatedModel.(MainModel)

	if !mainModel.showHelp {
		t.Error("Expected help to be shown after pressing 'h'")
	}
}

func TestMainModelWindowResize(t *testing.T) {
	model := NewMainModel()

	// Test window resize
	resizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(resizeMsg)
	mainModel := updatedModel.(MainModel)

	if mainModel.width != 120 || mainModel.height != 40 {
		t.Errorf("Expected dimensions 120x40, got %dx%d", mainModel.width, mainModel.height)
	}
}

func TestMainModelComponentUpdates(t *testing.T) {
	model := NewMainModel()

	// Test CPU update
	cpuMsg := CPUUpdateMsg{
		Cores: 4,
		Usage: []float64{25.0, 50.0, 75.0, 90.0},
		Total: 60.0,
	}

	updatedModel, _ := model.Update(cpuMsg)
	mainModel := updatedModel.(MainModel)

	if mainModel.cpu.GetCores() != 4 {
		t.Errorf("Expected 4 CPU cores, got %d", mainModel.cpu.GetCores())
	}

	if mainModel.cpu.GetTotal() != 60.0 {
		t.Errorf("Expected total CPU usage 60.0, got %f", mainModel.cpu.GetTotal())
	}
}

func TestMainModelView(t *testing.T) {
	model := NewMainModel()

	// Test normal view
	view := model.View()
	
	if !strings.Contains(view, "System Monitor") {
		t.Error("Expected view to contain 'System Monitor' title")
	}

	if !strings.Contains(view, "CPU Usage") {
		t.Error("Expected view to contain 'CPU Usage'")
	}

	if !strings.Contains(view, "Memory Usage") {
		t.Error("Expected view to contain 'Memory Usage'")
	}

	if !strings.Contains(view, "Disk Usage") {
		t.Error("Expected view to contain 'Disk Usage'")
	}

	if !strings.Contains(view, "Network Activity") {
		t.Error("Expected view to contain 'Network Activity'")
	}

	// Test help view
	model.showHelp = true
	helpView := model.View()

	if !strings.Contains(helpView, "Keyboard Shortcuts") {
		t.Error("Expected help view to contain 'Keyboard Shortcuts'")
	}

	if !strings.Contains(helpView, "Navigation:") {
		t.Error("Expected help view to contain navigation section")
	}
}

func TestFocusNavigation(t *testing.T) {
	model := NewMainModel()

	// Test nextFocus
	tests := []struct {
		current  FocusedComponent
		expected FocusedComponent
	}{
		{FocusCPU, FocusMemory},
		{FocusMemory, FocusDisk},
		{FocusDisk, FocusNetwork},
		{FocusNetwork, FocusCPU},
	}

	for _, tt := range tests {
		model.focused = tt.current
		next := model.nextFocus()
		if next != tt.expected {
			t.Errorf("nextFocus from %v: expected %v, got %v", tt.current, tt.expected, next)
		}
	}

	// Test prevFocus
	prevTests := []struct {
		current  FocusedComponent
		expected FocusedComponent
	}{
		{FocusCPU, FocusNetwork},
		{FocusMemory, FocusCPU},
		{FocusDisk, FocusMemory},
		{FocusNetwork, FocusDisk},
	}

	for _, tt := range prevTests {
		model.focused = tt.current
		prev := model.prevFocus()
		if prev != tt.expected {
			t.Errorf("prevFocus from %v: expected %v, got %v", tt.current, tt.expected, prev)
		}
	}
}

func TestVerticalNavigation(t *testing.T) {
	model := NewMainModel()

	// Test downFocus
	downTests := []struct {
		current  FocusedComponent
		expected FocusedComponent
	}{
		{FocusCPU, FocusDisk},
		{FocusMemory, FocusNetwork},
		{FocusDisk, FocusDisk},     // Should stay
		{FocusNetwork, FocusNetwork}, // Should stay
	}

	for _, tt := range downTests {
		model.focused = tt.current
		down := model.downFocus()
		if down != tt.expected {
			t.Errorf("downFocus from %v: expected %v, got %v", tt.current, tt.expected, down)
		}
	}

	// Test upFocus
	upTests := []struct {
		current  FocusedComponent
		expected FocusedComponent
	}{
		{FocusCPU, FocusCPU},     // Should stay
		{FocusMemory, FocusMemory}, // Should stay
		{FocusDisk, FocusCPU},
		{FocusNetwork, FocusMemory},
	}

	for _, tt := range upTests {
		model.focused = tt.current
		up := model.upFocus()
		if up != tt.expected {
			t.Errorf("upFocus from %v: expected %v, got %v", tt.current, tt.expected, up)
		}
	}
}

func TestContainsKey(t *testing.T) {
	model := NewMainModel()

	keys := []string{"q", "quit", "exit"}

	if !model.containsKey(keys, "q") {
		t.Error("Expected containsKey to return true for 'q'")
	}

	if !model.containsKey(keys, "quit") {
		t.Error("Expected containsKey to return true for 'quit'")
	}

	if model.containsKey(keys, "invalid") {
		t.Error("Expected containsKey to return false for 'invalid'")
	}
}

func TestGettersAndSetters(t *testing.T) {
	model := NewMainModel()

	// Test GetFocusedComponent
	if model.GetFocusedComponent() != FocusCPU {
		t.Error("Expected initial focus to be FocusCPU")
	}

	// Test SetFocusedComponent
	model = model.SetFocusedComponent(FocusMemory)
	if model.GetFocusedComponent() != FocusMemory {
		t.Error("Expected focus to be set to FocusMemory")
	}

	// Test component getters
	cpu := model.GetCPUModel()
	memory := model.GetMemoryModel()
	disk := model.GetDiskModel()
	network := model.GetNetworkModel()

	if cpu.GetCores() < 0 {
		t.Error("CPU model getter failed")
	}

	if memory.GetTotal() < 0 {
		t.Error("Memory model getter failed")
	}

	if len(disk.GetFilesystems()) < 0 {
		t.Error("Disk model getter failed")
	}

	if len(network.GetInterfaces()) < 0 {
		t.Error("Network model getter failed")
	}

	// Test IsShowingHelp and SetShowHelp
	if model.IsShowingHelp() {
		t.Error("Expected help to be hidden initially")
	}

	model = model.SetShowHelp(true)
	if !model.IsShowingHelp() {
		t.Error("Expected help to be shown after SetShowHelp(true)")
	}
}

func TestDefaultKeyMap(t *testing.T) {
	keyMap := DefaultKeyMap()

	// Test that all key mappings are defined
	if len(keyMap.Up) == 0 {
		t.Error("Expected Up keys to be defined")
	}

	if len(keyMap.Down) == 0 {
		t.Error("Expected Down keys to be defined")
	}

	if len(keyMap.Left) == 0 {
		t.Error("Expected Left keys to be defined")
	}

	if len(keyMap.Right) == 0 {
		t.Error("Expected Right keys to be defined")
	}

	if len(keyMap.Tab) == 0 {
		t.Error("Expected Tab keys to be defined")
	}

	if len(keyMap.Quit) == 0 {
		t.Error("Expected Quit keys to be defined")
	}

	if len(keyMap.Refresh) == 0 {
		t.Error("Expected Refresh keys to be defined")
	}

	if len(keyMap.Help) == 0 {
		t.Error("Expected Help keys to be defined")
	}

	// Test specific key mappings
	expectedQuitKeys := []string{"q", "ctrl+c"}
	for i, key := range expectedQuitKeys {
		if i < len(keyMap.Quit) && keyMap.Quit[i] != key {
			t.Errorf("Expected quit key %s, got %s", key, keyMap.Quit[i])
		}
	}
}

func TestFocusedComponentConstants(t *testing.T) {
	// Test that the constants are properly defined
	if FocusCPU != 0 {
		t.Errorf("Expected FocusCPU to be 0, got %d", FocusCPU)
	}

	if FocusMemory != 1 {
		t.Errorf("Expected FocusMemory to be 1, got %d", FocusMemory)
	}

	if FocusDisk != 2 {
		t.Errorf("Expected FocusDisk to be 2, got %d", FocusDisk)
	}

	if FocusNetwork != 3 {
		t.Errorf("Expected FocusNetwork to be 3, got %d", FocusNetwork)
	}
}

func TestMainModelRefreshKey(t *testing.T) {
	model := NewMainModel()

	// Test manual refresh with 'r' key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	updatedModel, cmd := model.Update(keyMsg)
	mainModel := updatedModel.(MainModel)

	// The model should remain unchanged (refresh doesn't change state directly)
	if mainModel.focused != model.focused {
		t.Error("Expected focus to remain unchanged after refresh")
	}

	// Command can be nil for now since refresh implementation is placeholder
	_ = cmd
}

func TestMainModelHelpDisplay(t *testing.T) {
	model := NewMainModel()

	// Test help display content
	model.showHelp = true
	helpView := model.View()

	expectedContent := []string{
		"System Monitor - Keyboard Shortcuts",
		"Navigation:",
		"↑/↓/←/→, hjkl",
		"Tab, Shift+Tab",
		"Actions:",
		"q, Ctrl+C",
		"r",
		"?, h",
		"Components:",
		"CPU",
		"Memory",
		"Disk",
		"Network",
		"Press any key to return",
	}

	for _, content := range expectedContent {
		if !strings.Contains(helpView, content) {
			t.Errorf("Expected help view to contain '%s'", content)
		}
	}
}

func TestMainModelKeyboardShortcutMapping(t *testing.T) {
	model := NewMainModel()
	keyMap := model.keys

	// Test that all required keyboard shortcuts are mapped correctly
	tests := []struct {
		name     string
		keys     []string
		expected []string
	}{
		{"Quit keys", keyMap.Quit, []string{"q", "ctrl+c"}},
		{"Help keys", keyMap.Help, []string{"?", "h"}},
		{"Refresh keys", keyMap.Refresh, []string{"r"}},
		{"Tab keys", keyMap.Tab, []string{"tab"}},
		{"Navigation keys", keyMap.Up, []string{"up", "k"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, expectedKey := range tt.expected {
				if i >= len(tt.keys) {
					t.Errorf("Missing expected key '%s' in %s", expectedKey, tt.name)
					continue
				}
				if tt.keys[i] != expectedKey {
					t.Errorf("Expected key '%s' at position %d in %s, got '%s'", 
						expectedKey, i, tt.name, tt.keys[i])
				}
			}
		})
	}
}

func TestMainModelHelpToggleFromAnyState(t *testing.T) {
	model := NewMainModel()

	// Test help toggle from different focus states
	focusStates := []FocusedComponent{FocusCPU, FocusMemory, FocusDisk, FocusNetwork}

	for _, focus := range focusStates {
		t.Run("Help toggle from focus "+string(rune(focus+'0')), func(t *testing.T) {
			model.focused = focus
			model.showHelp = false

			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
			updatedModel, _ := model.Update(keyMsg)
			mainModel := updatedModel.(MainModel)

			if !mainModel.showHelp {
				t.Error("Expected help to be shown")
			}

			// Focus should remain unchanged when toggling help
			if mainModel.focused != focus {
				t.Errorf("Expected focus to remain %v, got %v", focus, mainModel.focused)
			}
		})
	}
}