package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"golang-system-monitor-tui/models"
)

// CPUUpdateMsg represents a CPU update message
type CPUUpdateMsg models.CPUInfo

// CPUModel represents the CPU monitoring component
type CPUModel struct {
	usage    []float64    // Current per-core usage
	history  [][]float64  // Historical data for graphs (last 60 seconds)
	total    float64      // Overall CPU usage
	cores    int          // Number of CPU cores
	maxHistory int        // Maximum history entries to keep
	lastUpdate time.Time  // Last update timestamp
	width    int          // Component width for rendering
	height   int          // Component height for rendering
	styleManager *StyleManager // Style manager for consistent styling
	hasError bool         // Whether the component has an error
	errorMessage string   // Current error message
	lastError time.Time   // Timestamp of last error
}

// NewCPUModel creates a new CPU model instance
func NewCPUModel() CPUModel {
	return CPUModel{
		usage:        []float64{},
		history:      [][]float64{},
		total:        0.0,
		cores:        0,
		maxHistory:   60, // Keep 60 seconds of history
		lastUpdate:   time.Now(),
		width:        40,
		height:       10,
		styleManager: NewStyleManager(),
	}
}

// Init initializes the CPU model
func (m CPUModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the CPU model state
func (m CPUModel) Update(msg tea.Msg) (CPUModel, tea.Cmd) {
	switch msg := msg.(type) {
	case CPUUpdateMsg:
		// Clear any previous errors on successful update
		m.hasError = false
		m.errorMessage = ""
		
		// Update current usage data
		m.usage = msg.Usage
		m.total = msg.Total
		m.cores = msg.Cores
		m.lastUpdate = msg.Timestamp

		// Add current usage to history
		if len(m.usage) > 0 {
			// Initialize history if needed
			if len(m.history) == 0 {
				m.history = make([][]float64, len(m.usage))
				for i := range m.history {
					m.history[i] = make([]float64, 0, m.maxHistory)
				}
			}

			// Add current usage to each core's history
			for i, usage := range m.usage {
				if i < len(m.history) {
					m.history[i] = append(m.history[i], usage)
					// Keep only the last maxHistory entries
					if len(m.history[i]) > m.maxHistory {
						m.history[i] = m.history[i][1:]
					}
				}
			}
		}
		
	case models.ErrorMsg:
		// Handle error messages for CPU component
		if msg.Component == "CPU" {
			m.hasError = true
			m.errorMessage = msg.Message
			m.lastError = msg.Timestamp
		}
	}
	return m, nil
}

// View renders the CPU model
func (m CPUModel) View() string {
	var sections []string
	
	// Header
	header := m.styleManager.RenderHeader("CPU Usage")
	sections = append(sections, header)

	// Handle error state
	if m.hasError {
		sections = append(sections, m.styleManager.RenderErrorText("Error: "+m.errorMessage))
		sections = append(sections, m.styleManager.RenderMutedText("CPU data unavailable"))
		
		// Show fallback display with N/A values
		sections = append(sections, "Total: N/A")
		sections = append(sections, "Cores: N/A")
		
		// Add spacing
		for len(sections) < m.height {
			sections = append(sections, "")
		}
		return strings.Join(sections, "\n")
	}

	// Handle loading state
	if m.cores == 0 {
		return m.styleManager.RenderPlaceholder("CPU Usage", "Loading CPU data...")
	}

	// Normal display
	// Total CPU usage
	barWidth := m.styleManager.GetProgressBarWidth(m.width, 8) // "Total: " = 7 chars + space
	totalBar := m.styleManager.RenderProgressBar(m.total, barWidth, false)
	totalLine := fmt.Sprintf("Total: %s %.1f%%", totalBar, m.total)
	sections = append(sections, totalLine)

	// Per-core usage
	for i, usage := range m.usage {
		barWidth := m.styleManager.GetProgressBarWidth(m.width, 10) // "Core X: " = ~9 chars + space
		coreBar := m.styleManager.RenderProgressBar(usage, barWidth, false)
		coreLine := fmt.Sprintf("Core %d: %s %.1f%%", i+1, coreBar, usage)
		sections = append(sections, coreLine)
	}

	// Add spacing if we have fewer cores than available height
	for len(sections) < m.height {
		sections = append(sections, "")
	}

	return strings.Join(sections, "\n")
}



// SetSize sets the component dimensions
func (m CPUModel) SetSize(width, height int) CPUModel {
	m.width = width
	m.height = height
	return m
}

// GetUsage returns the current CPU usage data
func (m CPUModel) GetUsage() []float64 {
	return m.usage
}

// GetTotal returns the total CPU usage
func (m CPUModel) GetTotal() float64 {
	return m.total
}

// GetHistory returns the historical usage data
func (m CPUModel) GetHistory() [][]float64 {
	return m.history
}

// GetCores returns the number of CPU cores
func (m CPUModel) GetCores() int {
	return m.cores
}

// HasError returns whether the component has an error
func (m CPUModel) HasError() bool {
	return m.hasError
}

// GetErrorMessage returns the current error message
func (m CPUModel) GetErrorMessage() string {
	return m.errorMessage
}

// ClearError clears the current error state
func (m CPUModel) ClearError() CPUModel {
	m.hasError = false
	m.errorMessage = ""
	return m
}

// SetError sets an error state for the component
func (m CPUModel) SetError(message string) CPUModel {
	m.hasError = true
	m.errorMessage = message
	m.lastError = time.Now()
	return m
}