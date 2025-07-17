package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ColorScheme defines the application color palette
type ColorScheme struct {
	// Usage level colors
	Normal   lipgloss.Color // Green for normal usage (0-70%)
	Warning  lipgloss.Color // Yellow for warning usage (70-90%)
	Critical lipgloss.Color // Red for critical usage (90%+)
	
	// UI element colors
	Header   lipgloss.Color // Cyan for headers and titles
	Focused  lipgloss.Color // Cyan for focused components
	Unfocused lipgloss.Color // Gray for unfocused components
	Text     lipgloss.Color // Default text color
	Muted    lipgloss.Color // Gray for secondary text
	Background lipgloss.Color // Background color
}

// DefaultColorScheme returns the default color scheme
func DefaultColorScheme() ColorScheme {
	return ColorScheme{
		Normal:     lipgloss.Color("2"),  // Green
		Warning:    lipgloss.Color("3"),  // Yellow
		Critical:   lipgloss.Color("1"),  // Red
		Header:     lipgloss.Color("6"),  // Cyan
		Focused:    lipgloss.Color("6"),  // Cyan
		Unfocused:  lipgloss.Color("8"),  // Gray
		Text:       lipgloss.Color("15"), // White
		Muted:      lipgloss.Color("8"),  // Gray
		Background: lipgloss.Color("0"),  // Black
	}
}

// StyleManager handles all styling operations
type StyleManager struct {
	colors ColorScheme
	width  int
	height int
}

// NewStyleManager creates a new style manager
func NewStyleManager() *StyleManager {
	return &StyleManager{
		colors: DefaultColorScheme(),
		width:  80,
		height: 24,
	}
}

// SetDimensions updates the terminal dimensions
func (s *StyleManager) SetDimensions(width, height int) {
	s.width = width
	s.height = height
}

// GetUsageColor returns the appropriate color for a usage percentage
func (s *StyleManager) GetUsageColor(percentage float64) lipgloss.Color {
	switch {
	case percentage >= 90:
		return s.colors.Critical
	case percentage >= 70:
		return s.colors.Warning
	default:
		return s.colors.Normal
	}
}

// RenderProgressBar creates a styled progress bar
func (s *StyleManager) RenderProgressBar(percentage float64, width int, showPercentage bool) string {
	if width <= 0 {
		width = 20
	}

	// Calculate filled portion
	filled := int((percentage / 100.0) * float64(width))
	if filled > width {
		filled = width
	}

	// Create the bar
	filledChar := "█"
	emptyChar := "░"
	bar := strings.Repeat(filledChar, filled) + strings.Repeat(emptyChar, width-filled)

	// Apply color based on usage level
	color := s.GetUsageColor(percentage)
	styledBar := lipgloss.NewStyle().Foreground(color).Render(bar)

	// Add percentage if requested
	if showPercentage {
		percentText := lipgloss.NewStyle().
			Foreground(s.colors.Text).
			Render(lipgloss.PlaceHorizontal(6, lipgloss.Right, fmt.Sprintf("%.1f%%", percentage)))
		return styledBar + " " + percentText
	}

	return styledBar
}

// RenderHeader creates a styled header
func (s *StyleManager) RenderHeader(title string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(s.colors.Header).
		Render(title)
}

// RenderComponentBorder creates a styled border for components
func (s *StyleManager) RenderComponentBorder(content string, focused bool, width, height int) string {
	var borderColor lipgloss.Color
	if focused {
		borderColor = s.colors.Focused
	} else {
		borderColor = s.colors.Unfocused
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Width(width).
		Height(height).
		Padding(0, 1)

	return style.Render(content)
}

// RenderPlaceholder creates a styled placeholder text
func (s *StyleManager) RenderPlaceholder(title, message string) string {
	header := s.RenderHeader(title)
	placeholder := lipgloss.NewStyle().
		Foreground(s.colors.Muted).
		Render(message)
	
	return header + "\n" + placeholder
}

// RenderMutedText creates styled muted text
func (s *StyleManager) RenderMutedText(text string) string {
	return lipgloss.NewStyle().
		Foreground(s.colors.Muted).
		Render(text)
}

// RenderHighlightText creates styled highlighted text
func (s *StyleManager) RenderHighlightText(text string) string {
	return lipgloss.NewStyle().
		Foreground(s.colors.Header).
		Bold(true).
		Render(text)
}

// RenderWarningText creates styled warning text
func (s *StyleManager) RenderWarningText(text string) string {
	return lipgloss.NewStyle().
		Foreground(s.colors.Warning).
		Bold(true).
		Render(text)
}

// RenderCriticalText creates styled critical text
func (s *StyleManager) RenderCriticalText(text string) string {
	return lipgloss.NewStyle().
		Foreground(s.colors.Critical).
		Bold(true).
		Render(text)
}

// RenderErrorText creates styled error text
func (s *StyleManager) RenderErrorText(text string) string {
	return lipgloss.NewStyle().
		Foreground(s.colors.Critical).
		Bold(true).
		Render(text)
}

// CalculateComponentDimensions calculates optimal component dimensions
func (s *StyleManager) CalculateComponentDimensions() (width, height int) {
	// Reserve space for borders, padding, header, and footer
	availableWidth := s.width - 6  // Account for borders and spacing
	availableHeight := s.height - 6 // Account for header, footer, and spacing

	// Split into 2x2 grid
	componentWidth := availableWidth / 2
	componentHeight := availableHeight / 2

	// Ensure minimum dimensions
	if componentWidth < 30 {
		componentWidth = 30
	}
	if componentHeight < 8 {
		componentHeight = 8
	}

	return componentWidth, componentHeight
}

// IsSmallTerminal checks if the terminal is too small for optimal display
func (s *StyleManager) IsSmallTerminal() bool {
	return s.width < 80 || s.height < 24
}

// GetMinimumDimensions returns the minimum required terminal dimensions
func (s *StyleManager) GetMinimumDimensions() (width, height int) {
	return 80, 24
}

// RenderResponsiveLayout creates a layout that adapts to terminal size
func (s *StyleManager) RenderResponsiveLayout(components []string) string {
	if s.IsSmallTerminal() {
		// For small terminals, stack components vertically
		return s.renderVerticalLayout(components)
	}
	
	// For normal terminals, use 2x2 grid
	return s.render2x2Layout(components)
}

// render2x2Layout creates a 2x2 grid layout
func (s *StyleManager) render2x2Layout(components []string) string {
	if len(components) < 4 {
		// Pad with empty components if needed
		for len(components) < 4 {
			components = append(components, "")
		}
	}

	// Create top and bottom rows
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, components[0], " ", components[1])
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, components[2], " ", components[3])
	
	return lipgloss.JoinVertical(lipgloss.Left, topRow, "", bottomRow)
}

// renderVerticalLayout creates a vertical stack layout for small terminals
func (s *StyleManager) renderVerticalLayout(components []string) string {
	var nonEmptyComponents []string
	for _, component := range components {
		if strings.TrimSpace(component) != "" {
			nonEmptyComponents = append(nonEmptyComponents, component)
		}
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, nonEmptyComponents...)
}

// RenderApplicationHeader creates the main application header
func (s *StyleManager) RenderApplicationHeader(title string) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(s.colors.Header).
		Align(lipgloss.Center).
		Width(s.width).
		Render(title)
}

// RenderApplicationFooter creates the main application footer
func (s *StyleManager) RenderApplicationFooter(shortcuts []string) string {
	footerText := strings.Join(shortcuts, " • ")
	return lipgloss.NewStyle().
		Foreground(s.colors.Muted).
		Align(lipgloss.Center).
		Width(s.width).
		Render(footerText)
}

// RenderHelpScreen creates a styled help screen
func (s *StyleManager) RenderHelpScreen(content string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.colors.Header).
		Padding(2).
		Margin(2).
		Width(s.width - 8).
		Height(s.height - 8).
		Render(content)
}

// GetProgressBarWidth calculates optimal progress bar width for a component
func (s *StyleManager) GetProgressBarWidth(componentWidth int, labelWidth int) int {
	// Reserve space for label, percentage, and padding
	barWidth := componentWidth - labelWidth - 10
	if barWidth < 10 {
		barWidth = 10
	}
	return barWidth
}