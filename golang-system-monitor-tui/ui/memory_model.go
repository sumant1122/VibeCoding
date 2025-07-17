package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"golang-system-monitor-tui/models"
)

// MemoryUpdateMsg represents a memory update message
type MemoryUpdateMsg models.MemoryInfo

// MemoryModel represents the memory monitoring component
type MemoryModel struct {
	total      uint64    // Total RAM in bytes
	used       uint64    // Used RAM in bytes
	available  uint64    // Available RAM in bytes
	swap       models.SwapInfo // Swap memory information
	lastUpdate time.Time // Last update timestamp
	width      int       // Component width for rendering
	height     int       // Component height for rendering
	styleManager *StyleManager // Style manager for consistent styling
	hasError bool         // Whether the component has an error
	errorMessage string   // Current error message
	lastError time.Time   // Timestamp of last error
}

// NewMemoryModel creates a new memory model instance
func NewMemoryModel() MemoryModel {
	return MemoryModel{
		total:        0,
		used:         0,
		available:    0,
		swap:         models.SwapInfo{},
		lastUpdate:   time.Now(),
		width:        40,
		height:       8,
		styleManager: NewStyleManager(),
	}
}

// Init initializes the memory model
func (m MemoryModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the memory model state
func (m MemoryModel) Update(msg tea.Msg) (MemoryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case MemoryUpdateMsg:
		// Clear any previous errors on successful update
		m.hasError = false
		m.errorMessage = ""
		
		// Update memory data
		m.total = msg.Total
		m.used = msg.Used
		m.available = msg.Available
		m.swap = msg.Swap
		m.lastUpdate = msg.Timestamp
		
	case models.ErrorMsg:
		// Handle error messages for Memory component
		if msg.Component == "Memory" {
			m.hasError = true
			m.errorMessage = msg.Message
			m.lastError = msg.Timestamp
		}
	}
	return m, nil
}

// View renders the memory model
func (m MemoryModel) View() string {
	var sections []string
	
	// Header
	header := m.styleManager.RenderHeader("Memory Usage")
	sections = append(sections, header)

	// Handle error state
	if m.hasError {
		sections = append(sections, m.styleManager.RenderErrorText("Error: "+m.errorMessage))
		sections = append(sections, m.styleManager.RenderMutedText("Memory data unavailable"))
		
		// Show fallback display with N/A values
		sections = append(sections, "RAM: N/A")
		sections = append(sections, "Swap: N/A")
		
		// Add spacing
		for len(sections) < m.height {
			sections = append(sections, "")
		}
		return strings.Join(sections, "\n")
	}

	// Handle loading state
	if m.total == 0 {
		return m.styleManager.RenderPlaceholder("Memory Usage", "Loading memory data...")
	}

	// Normal display
	// RAM usage
	ramUsagePercent := float64(m.used) / float64(m.total) * 100
	barWidth := m.styleManager.GetProgressBarWidth(m.width, 6) // "RAM: " = 5 chars + space
	ramBar := m.styleManager.RenderProgressBar(ramUsagePercent, barWidth, false)
	ramLine := fmt.Sprintf("RAM: %s %.1f%%", ramBar, ramUsagePercent)
	sections = append(sections, ramLine)

	// RAM details in human-readable format
	ramDetails := fmt.Sprintf("     %s / %s", 
		m.formatBytes(m.used), 
		m.formatBytes(m.total))
	sections = append(sections, m.styleManager.RenderMutedText(ramDetails))

	// Swap usage (if swap is configured)
	if m.swap.Total > 0 {
		swapUsagePercent := float64(m.swap.Used) / float64(m.swap.Total) * 100
		barWidth := m.styleManager.GetProgressBarWidth(m.width, 7) // "Swap: " = 6 chars + space
		swapBar := m.styleManager.RenderProgressBar(swapUsagePercent, barWidth, false)
		swapLine := fmt.Sprintf("Swap: %s %.1f%%", swapBar, swapUsagePercent)
		sections = append(sections, swapLine)

		// Swap details in human-readable format
		swapDetails := fmt.Sprintf("      %s / %s", 
			m.formatBytes(m.swap.Used), 
			m.formatBytes(m.swap.Total))
		sections = append(sections, m.styleManager.RenderMutedText(swapDetails))
	} else {
		sections = append(sections, m.styleManager.RenderMutedText("Swap: Not configured"))
	}

	// Add spacing if we have fewer lines than available height
	for len(sections) < m.height {
		sections = append(sections, "")
	}

	return strings.Join(sections, "\n")
}



// formatBytes converts bytes to human-readable format (GB/MB/KB)
func (m MemoryModel) formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1fTB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

// SetSize sets the component dimensions
func (m MemoryModel) SetSize(width, height int) MemoryModel {
	m.width = width
	m.height = height
	return m
}

// GetTotal returns the total memory in bytes
func (m MemoryModel) GetTotal() uint64 {
	return m.total
}

// GetUsed returns the used memory in bytes
func (m MemoryModel) GetUsed() uint64 {
	return m.used
}

// GetAvailable returns the available memory in bytes
func (m MemoryModel) GetAvailable() uint64 {
	return m.available
}

// GetSwap returns the swap memory information
func (m MemoryModel) GetSwap() models.SwapInfo {
	return m.swap
}

// GetUsagePercent returns the memory usage percentage
func (m MemoryModel) GetUsagePercent() float64 {
	if m.total == 0 {
		return 0
	}
	return float64(m.used) / float64(m.total) * 100
}

// GetSwapUsagePercent returns the swap usage percentage
func (m MemoryModel) GetSwapUsagePercent() float64 {
	if m.swap.Total == 0 {
		return 0
	}
	return float64(m.swap.Used) / float64(m.swap.Total) * 100
}

// HasError returns whether the component has an error
func (m MemoryModel) HasError() bool {
	return m.hasError
}

// GetErrorMessage returns the current error message
func (m MemoryModel) GetErrorMessage() string {
	return m.errorMessage
}

// ClearError clears the current error state
func (m MemoryModel) ClearError() MemoryModel {
	m.hasError = false
	m.errorMessage = ""
	return m
}

// SetError sets an error state for the component
func (m MemoryModel) SetError(message string) MemoryModel {
	m.hasError = true
	m.errorMessage = message
	m.lastError = time.Now()
	return m
}