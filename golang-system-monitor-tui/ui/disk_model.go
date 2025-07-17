package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"golang-system-monitor-tui/models"
)

// DiskUpdateMsg represents a disk update message
type DiskUpdateMsg []models.DiskInfo

// DiskModel represents the disk monitoring component
type DiskModel struct {
	filesystems []models.DiskInfo // Current filesystem information
	lastUpdate  time.Time         // Last update timestamp
	width       int               // Component width for rendering
	height      int               // Component height for rendering
	styleManager *StyleManager    // Style manager for consistent styling
	hasError bool         // Whether the component has an error
	errorMessage string   // Current error message
	lastError time.Time   // Timestamp of last error
}

// NewDiskModel creates a new disk model instance
func NewDiskModel() DiskModel {
	return DiskModel{
		filesystems:  []models.DiskInfo{},
		lastUpdate:   time.Now(),
		width:        50,
		height:       10,
		styleManager: NewStyleManager(),
	}
}

// Init initializes the disk model
func (m DiskModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the disk model state
func (m DiskModel) Update(msg tea.Msg) (DiskModel, tea.Cmd) {
	switch msg := msg.(type) {
	case DiskUpdateMsg:
		// Clear any previous errors on successful update
		m.hasError = false
		m.errorMessage = ""
		
		// Update filesystem data
		m.filesystems = []models.DiskInfo(msg)
		m.lastUpdate = time.Now()
		
	case models.ErrorMsg:
		// Handle error messages for Disk component
		if msg.Component == "Disk" {
			m.hasError = true
			m.errorMessage = msg.Message
			m.lastError = msg.Timestamp
		}
	}
	return m, nil
}

// View renders the disk model
func (m DiskModel) View() string {
	var sections []string
	
	// Header
	header := m.styleManager.RenderHeader("Disk Usage")
	sections = append(sections, header)

	// Handle error state
	if m.hasError {
		sections = append(sections, m.styleManager.RenderErrorText("Error: "+m.errorMessage))
		sections = append(sections, m.styleManager.RenderMutedText("Disk data unavailable"))
		
		// Show fallback display with N/A values
		sections = append(sections, "Filesystems: N/A")
		sections = append(sections, "Usage: N/A")
		
		// Add spacing
		for len(sections) < m.height {
			sections = append(sections, "")
		}
		return strings.Join(sections, "\n")
	}

	// Handle loading state
	if len(m.filesystems) == 0 {
		return m.styleManager.RenderPlaceholder("Disk Usage", "Loading disk data...")
	}

	// Normal display
	// Render each filesystem
	for _, fs := range m.filesystems {
		// Truncate long mountpoints for better display
		mountpoint := fs.Mountpoint
		if len(mountpoint) > 15 {
			mountpoint = mountpoint[:12] + "..."
		}
		
		// Create filesystem line with progress bar
		barWidth := m.styleManager.GetProgressBarWidth(m.width, 18) // 15 chars for mountpoint + 3 for spacing
		fsBar := m.styleManager.RenderProgressBar(fs.UsedPercent, barWidth, false)
		
		fsLine := fmt.Sprintf("%-15s %s %.1f%%", 
			mountpoint, fsBar, fs.UsedPercent)
		
		// Apply warning/critical styling if needed
		if fs.UsedPercent >= 90 {
			sections = append(sections, m.styleManager.RenderCriticalText(fsLine))
		} else if fs.UsedPercent >= 70 {
			sections = append(sections, m.styleManager.RenderWarningText(fsLine))
		} else {
			sections = append(sections, fsLine)
		}

		// Add size details in human-readable format
		sizeDetails := fmt.Sprintf("%-15s %s / %s", 
			"", 
			m.formatBytes(fs.Used), 
			m.formatBytes(fs.Total))
		sections = append(sections, m.styleManager.RenderMutedText(sizeDetails))
	}

	// Add spacing if we have fewer lines than available height
	for len(sections) < m.height {
		sections = append(sections, "")
	}

	return strings.Join(sections, "\n")
}



// formatBytes converts bytes to human-readable format (GB/MB/KB)
func (m DiskModel) formatBytes(bytes uint64) string {
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
func (m DiskModel) SetSize(width, height int) DiskModel {
	m.width = width
	m.height = height
	return m
}

// GetFilesystems returns the current filesystem information
func (m DiskModel) GetFilesystems() []models.DiskInfo {
	return m.filesystems
}

// GetHighUsageFilesystems returns filesystems with usage above the specified threshold
func (m DiskModel) GetHighUsageFilesystems(threshold float64) []models.DiskInfo {
	var highUsage []models.DiskInfo
	for _, fs := range m.filesystems {
		if fs.UsedPercent >= threshold {
			highUsage = append(highUsage, fs)
		}
	}
	return highUsage
}

// GetCriticalFilesystems returns filesystems with usage >= 90%
func (m DiskModel) GetCriticalFilesystems() []models.DiskInfo {
	return m.GetHighUsageFilesystems(90.0)
}

// HasCriticalUsage returns true if any filesystem has usage >= 90%
func (m DiskModel) HasCriticalUsage() bool {
	return len(m.GetCriticalFilesystems()) > 0
}

// GetTotalDiskSpace returns the total disk space across all filesystems
func (m DiskModel) GetTotalDiskSpace() uint64 {
	var total uint64
	for _, fs := range m.filesystems {
		total += fs.Total
	}
	return total
}

// GetTotalUsedSpace returns the total used space across all filesystems
func (m DiskModel) GetTotalUsedSpace() uint64 {
	var used uint64
	for _, fs := range m.filesystems {
		used += fs.Used
	}
	return used
}

// GetOverallUsagePercent returns the overall usage percentage across all filesystems
func (m DiskModel) GetOverallUsagePercent() float64 {
	totalSpace := m.GetTotalDiskSpace()
	if totalSpace == 0 {
		return 0
	}
	return float64(m.GetTotalUsedSpace()) / float64(totalSpace) * 100
}

// HasError returns whether the component has an error
func (m DiskModel) HasError() bool {
	return m.hasError
}

// GetErrorMessage returns the current error message
func (m DiskModel) GetErrorMessage() string {
	return m.errorMessage
}

// ClearError clears the current error state
func (m DiskModel) ClearError() DiskModel {
	m.hasError = false
	m.errorMessage = ""
	return m
}

// SetError sets an error state for the component
func (m DiskModel) SetError(message string) DiskModel {
	m.hasError = true
	m.errorMessage = message
	m.lastError = time.Now()
	return m
}