package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultColorScheme(t *testing.T) {
	scheme := DefaultColorScheme()
	
	// Test that all colors are defined
	if scheme.Normal == "" {
		t.Error("Normal color should be defined")
	}
	if scheme.Warning == "" {
		t.Error("Warning color should be defined")
	}
	if scheme.Critical == "" {
		t.Error("Critical color should be defined")
	}
	if scheme.Header == "" {
		t.Error("Header color should be defined")
	}
	if scheme.Focused == "" {
		t.Error("Focused color should be defined")
	}
	if scheme.Unfocused == "" {
		t.Error("Unfocused color should be defined")
	}
}

func TestNewStyleManager(t *testing.T) {
	sm := NewStyleManager()
	
	if sm == nil {
		t.Fatal("StyleManager should not be nil")
	}
	
	if sm.width != 80 {
		t.Errorf("Expected default width 80, got %d", sm.width)
	}
	
	if sm.height != 24 {
		t.Errorf("Expected default height 24, got %d", sm.height)
	}
}

func TestSetDimensions(t *testing.T) {
	sm := NewStyleManager()
	
	sm.SetDimensions(120, 40)
	
	if sm.width != 120 {
		t.Errorf("Expected width 120, got %d", sm.width)
	}
	
	if sm.height != 40 {
		t.Errorf("Expected height 40, got %d", sm.height)
	}
}

func TestGetUsageColor(t *testing.T) {
	sm := NewStyleManager()
	
	tests := []struct {
		percentage float64
		expected   lipgloss.Color
	}{
		{0, sm.colors.Normal},
		{50, sm.colors.Normal},
		{69.9, sm.colors.Normal},
		{70, sm.colors.Warning},
		{85, sm.colors.Warning},
		{89.9, sm.colors.Warning},
		{90, sm.colors.Critical},
		{95, sm.colors.Critical},
		{100, sm.colors.Critical},
	}
	
	for _, test := range tests {
		result := sm.GetUsageColor(test.percentage)
		if result != test.expected {
			t.Errorf("For percentage %.1f, expected color %s, got %s", 
				test.percentage, test.expected, result)
		}
	}
}

func TestRenderProgressBar(t *testing.T) {
	sm := NewStyleManager()
	
	// Test basic progress bar
	bar := sm.RenderProgressBar(50, 20, false)
	if bar == "" {
		t.Error("Progress bar should not be empty")
	}
	
	// Test progress bar with percentage
	barWithPercent := sm.RenderProgressBar(75, 20, true)
	if !strings.Contains(barWithPercent, "75.0%") {
		t.Error("Progress bar with percentage should contain percentage text")
	}
	
	// Test zero width handling
	zeroWidthBar := sm.RenderProgressBar(50, 0, false)
	if zeroWidthBar == "" {
		t.Error("Zero width bar should default to minimum width")
	}
	
	// Test over 100% handling
	overBar := sm.RenderProgressBar(150, 20, false)
	if overBar == "" {
		t.Error("Over 100% bar should be handled gracefully")
	}
}

func TestRenderHeader(t *testing.T) {
	sm := NewStyleManager()
	
	header := sm.RenderHeader("Test Header")
	if header == "" {
		t.Error("Header should not be empty")
	}
	
	// The actual styled text will contain ANSI codes, so we can't do exact string matching
	// but we can verify it's not empty and contains our text
	if !strings.Contains(header, "Test Header") {
		t.Error("Header should contain the original text")
	}
}

func TestRenderComponentBorder(t *testing.T) {
	sm := NewStyleManager()
	
	content := "Test Content"
	
	// Test focused border
	focusedBorder := sm.RenderComponentBorder(content, true, 40, 10)
	if focusedBorder == "" {
		t.Error("Focused border should not be empty")
	}
	
	// Test unfocused border
	unfocusedBorder := sm.RenderComponentBorder(content, false, 40, 10)
	if unfocusedBorder == "" {
		t.Error("Unfocused border should not be empty")
	}
	
	// Both should contain the original content
	if !strings.Contains(focusedBorder, content) {
		t.Error("Focused border should contain the original content")
	}
	
	if !strings.Contains(unfocusedBorder, content) {
		t.Error("Unfocused border should contain the original content")
	}
	
	// Test that the function works with different focus states (we can't easily test visual differences)
	// but we can ensure both calls succeed and produce output
	if len(focusedBorder) == 0 || len(unfocusedBorder) == 0 {
		t.Error("Both focused and unfocused borders should produce non-empty output")
	}
}

func TestRenderPlaceholder(t *testing.T) {
	sm := NewStyleManager()
	
	placeholder := sm.RenderPlaceholder("Test Title", "Test Message")
	if placeholder == "" {
		t.Error("Placeholder should not be empty")
	}
	
	if !strings.Contains(placeholder, "Test Title") {
		t.Error("Placeholder should contain title")
	}
	
	if !strings.Contains(placeholder, "Test Message") {
		t.Error("Placeholder should contain message")
	}
}

func TestCalculateComponentDimensions(t *testing.T) {
	sm := NewStyleManager()
	
	// Test normal terminal size
	sm.SetDimensions(80, 24)
	width, height := sm.CalculateComponentDimensions()
	
	if width < 30 {
		t.Errorf("Component width should be at least 30, got %d", width)
	}
	
	if height < 8 {
		t.Errorf("Component height should be at least 8, got %d", height)
	}
	
	// Test small terminal size
	sm.SetDimensions(60, 20)
	smallWidth, smallHeight := sm.CalculateComponentDimensions()
	
	if smallWidth < 30 {
		t.Errorf("Small terminal component width should be at least 30, got %d", smallWidth)
	}
	
	if smallHeight < 8 {
		t.Errorf("Small terminal component height should be at least 8, got %d", smallHeight)
	}
}

func TestIsSmallTerminal(t *testing.T) {
	sm := NewStyleManager()
	
	// Test normal size
	sm.SetDimensions(80, 24)
	if sm.IsSmallTerminal() {
		t.Error("80x24 should not be considered small")
	}
	
	// Test small width
	sm.SetDimensions(70, 24)
	if !sm.IsSmallTerminal() {
		t.Error("70x24 should be considered small")
	}
	
	// Test small height
	sm.SetDimensions(80, 20)
	if !sm.IsSmallTerminal() {
		t.Error("80x20 should be considered small")
	}
	
	// Test both small
	sm.SetDimensions(60, 20)
	if !sm.IsSmallTerminal() {
		t.Error("60x20 should be considered small")
	}
}

func TestGetMinimumDimensions(t *testing.T) {
	sm := NewStyleManager()
	
	width, height := sm.GetMinimumDimensions()
	
	if width != 80 {
		t.Errorf("Expected minimum width 80, got %d", width)
	}
	
	if height != 24 {
		t.Errorf("Expected minimum height 24, got %d", height)
	}
}

func TestRenderResponsiveLayout(t *testing.T) {
	sm := NewStyleManager()
	
	components := []string{"Component1", "Component2", "Component3", "Component4"}
	
	// Test normal terminal (should use 2x2 layout)
	sm.SetDimensions(80, 24)
	normalLayout := sm.RenderResponsiveLayout(components)
	if normalLayout == "" {
		t.Error("Normal layout should not be empty")
	}
	
	// Test small terminal (should use vertical layout)
	sm.SetDimensions(60, 20)
	smallLayout := sm.RenderResponsiveLayout(components)
	if smallLayout == "" {
		t.Error("Small layout should not be empty")
	}
	
	// Layouts should be different for different terminal sizes
	if normalLayout == smallLayout {
		t.Error("Normal and small layouts should be different")
	}
}

func TestRender2x2Layout(t *testing.T) {
	sm := NewStyleManager()
	
	// Test with exactly 4 components
	components := []string{"A", "B", "C", "D"}
	layout := sm.render2x2Layout(components)
	if layout == "" {
		t.Error("2x2 layout should not be empty")
	}
	
	// Test with fewer than 4 components (should pad)
	shortComponents := []string{"A", "B"}
	shortLayout := sm.render2x2Layout(shortComponents)
	if shortLayout == "" {
		t.Error("2x2 layout with padding should not be empty")
	}
}

func TestRenderVerticalLayout(t *testing.T) {
	sm := NewStyleManager()
	
	components := []string{"Component1", "Component2", "", "Component4"}
	layout := sm.renderVerticalLayout(components)
	
	if layout == "" {
		t.Error("Vertical layout should not be empty")
	}
	
	// Should contain non-empty components
	if !strings.Contains(layout, "Component1") {
		t.Error("Vertical layout should contain Component1")
	}
	
	if !strings.Contains(layout, "Component2") {
		t.Error("Vertical layout should contain Component2")
	}
	
	if !strings.Contains(layout, "Component4") {
		t.Error("Vertical layout should contain Component4")
	}
}

func TestGetProgressBarWidth(t *testing.T) {
	sm := NewStyleManager()
	
	// Test normal case
	width := sm.GetProgressBarWidth(50, 10)
	expected := 50 - 10 - 10 // componentWidth - labelWidth - padding
	if width != expected {
		t.Errorf("Expected progress bar width %d, got %d", expected, width)
	}
	
	// Test minimum width enforcement
	minWidth := sm.GetProgressBarWidth(15, 10)
	if minWidth < 10 {
		t.Errorf("Progress bar width should be at least 10, got %d", minWidth)
	}
}

func TestRenderApplicationHeader(t *testing.T) {
	sm := NewStyleManager()
	sm.SetDimensions(80, 24)
	
	header := sm.RenderApplicationHeader("Test App")
	if header == "" {
		t.Error("Application header should not be empty")
	}
	
	if !strings.Contains(header, "Test App") {
		t.Error("Application header should contain the title")
	}
}

func TestRenderApplicationFooter(t *testing.T) {
	sm := NewStyleManager()
	sm.SetDimensions(80, 24)
	
	shortcuts := []string{"q: quit", "h: help"}
	footer := sm.RenderApplicationFooter(shortcuts)
	
	if footer == "" {
		t.Error("Application footer should not be empty")
	}
	
	if !strings.Contains(footer, "q: quit") {
		t.Error("Application footer should contain shortcuts")
	}
	
	if !strings.Contains(footer, "•") {
		t.Error("Application footer should contain separator")
	}
}

func TestRenderHelpScreen(t *testing.T) {
	sm := NewStyleManager()
	sm.SetDimensions(80, 24)
	
	content := "Help content here"
	help := sm.RenderHelpScreen(content)
	
	if help == "" {
		t.Error("Help screen should not be empty")
	}
	
	if !strings.Contains(help, content) {
		t.Error("Help screen should contain the provided content")
	}
}

// Test styling consistency across different usage levels
func TestStylingConsistency(t *testing.T) {
	sm := NewStyleManager()
	
	// Test that different usage levels produce different colors
	normalColor := sm.GetUsageColor(50)
	warningColor := sm.GetUsageColor(80)
	criticalColor := sm.GetUsageColor(95)
	
	if normalColor == warningColor {
		t.Error("Normal and warning colors should be different")
	}
	
	if warningColor == criticalColor {
		t.Error("Warning and critical colors should be different")
	}
	
	if normalColor == criticalColor {
		t.Error("Normal and critical colors should be different")
	}
}

// Test responsive behavior with various terminal sizes
func TestResponsiveBehavior(t *testing.T) {
	sm := NewStyleManager()
	
	testSizes := []struct {
		width, height int
		shouldBeSmall bool
	}{
		{80, 24, false},
		{100, 30, false},
		{79, 24, true},
		{80, 23, true},
		{60, 20, true},
		{40, 15, true},
	}
	
	for _, test := range testSizes {
		sm.SetDimensions(test.width, test.height)
		isSmall := sm.IsSmallTerminal()
		
		if isSmall != test.shouldBeSmall {
			t.Errorf("Terminal %dx%d: expected small=%v, got small=%v", 
				test.width, test.height, test.shouldBeSmall, isSmall)
		}
	}
}