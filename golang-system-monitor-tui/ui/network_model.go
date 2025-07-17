package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"golang-system-monitor-tui/models"
)

// NetworkUpdateMsg represents a network update message
type NetworkUpdateMsg []models.NetworkInfo

// NetworkModel represents the network monitoring component
type NetworkModel struct {
	interfaces    []models.NetworkInfo         // Current network interface information
	previousData  []models.NetworkInfo         // Previous measurement for rate calculation
	rates         map[string]models.NetworkStats // Calculated transfer rates
	lastUpdate    time.Time                    // Last update timestamp
	width         int                          // Component width for rendering
	height        int                          // Component height for rendering
	styleManager  *StyleManager               // Style manager for consistent styling
	hasError bool         // Whether the component has an error
	errorMessage string   // Current error message
	lastError time.Time   // Timestamp of last error
}

// NewNetworkModel creates a new network model instance
func NewNetworkModel() NetworkModel {
	return NetworkModel{
		interfaces:   []models.NetworkInfo{},
		previousData: []models.NetworkInfo{},
		rates:        make(map[string]models.NetworkStats),
		lastUpdate:   time.Now(),
		width:        50,
		height:       10,
		styleManager: NewStyleManager(),
	}
}

// Init initializes the network model
func (m NetworkModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the network model state
func (m NetworkModel) Update(msg tea.Msg) (NetworkModel, tea.Cmd) {
	switch msg := msg.(type) {
	case NetworkUpdateMsg:
		// Clear any previous errors on successful update
		m.hasError = false
		m.errorMessage = ""
		
		// Store previous data for rate calculation
		m.previousData = m.interfaces
		
		// Update current interface data
		m.interfaces = []models.NetworkInfo(msg)
		m.lastUpdate = time.Now()
		
		// Calculate transfer rates if we have previous data
		if len(m.previousData) > 0 {
			m.rates = m.calculateRates(m.previousData, m.interfaces)
		}
		
	case models.ErrorMsg:
		// Handle error messages for Network component
		if msg.Component == "Network" {
			m.hasError = true
			m.errorMessage = msg.Message
			m.lastError = msg.Timestamp
		}
	}
	return m, nil
}

// View renders the network model
func (m NetworkModel) View() string {
	var sections []string
	
	// Header
	header := m.styleManager.RenderHeader("Network Activity")
	sections = append(sections, header)

	// Handle error state
	if m.hasError {
		sections = append(sections, m.styleManager.RenderErrorText("Error: "+m.errorMessage))
		sections = append(sections, m.styleManager.RenderMutedText("Network data unavailable"))
		
		// Show fallback display with N/A values
		sections = append(sections, "Interfaces: N/A")
		sections = append(sections, "Activity: N/A")
		
		// Add spacing
		for len(sections) < m.height {
			sections = append(sections, "")
		}
		return strings.Join(sections, "\n")
	}

	// Handle loading state
	if len(m.interfaces) == 0 {
		return m.styleManager.RenderPlaceholder("Network Activity", "Loading network data...")
	}

	// Normal display
	// Render each network interface
	for _, iface := range m.interfaces {
		// Get transfer rates for this interface
		stats, hasRates := m.rates[iface.Interface]
		
		// Interface name (truncate if too long)
		interfaceName := iface.Interface
		if len(interfaceName) > 12 {
			interfaceName = interfaceName[:9] + "..."
		}
		
		// Create interface line with transfer rates
		var rateLine string
		if hasRates {
			rateLine = fmt.Sprintf("%-12s ↑ %8s ↓ %8s", 
				interfaceName,
				m.formatRate(stats.SendRate),
				m.formatRate(stats.RecvRate))
		} else {
			rateLine = fmt.Sprintf("%-12s ↑ %8s ↓ %8s", 
				interfaceName, "N/A", "N/A")
		}
		
		// Apply color based on activity level using style manager
		styledLine := m.styleByActivityWithManager(rateLine, stats)
		sections = append(sections, styledLine)
		
		// Add total bytes transferred (optional detail line)
		totalLine := fmt.Sprintf("%-12s   %8s   %8s", 
			"",
			m.formatBytes(iface.BytesSent),
			m.formatBytes(iface.BytesRecv))
		
		sections = append(sections, m.styleManager.RenderMutedText(totalLine))
	}

	// Add spacing if we have fewer lines than available height
	for len(sections) < m.height {
		sections = append(sections, "")
	}

	return strings.Join(sections, "\n")
}

// calculateRates calculates transfer rates between two network measurements
func (m NetworkModel) calculateRates(previous, current []models.NetworkInfo) map[string]models.NetworkStats {
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
				
				// Handle counter rollover by checking if current < previous
				if curr.BytesSent >= prev.BytesSent {
					sendRate = float64(curr.BytesSent-prev.BytesSent) / timeDiff
				} else {
					// Counter rollover detected, set rate to 0
					sendRate = 0
				}
				
				if curr.BytesRecv >= prev.BytesRecv {
					recvRate = float64(curr.BytesRecv-prev.BytesRecv) / timeDiff
				} else {
					// Counter rollover detected, set rate to 0
					recvRate = 0
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

// styleByActivityWithManager applies color styling based on network activity level using style manager
func (m NetworkModel) styleByActivityWithManager(text string, stats models.NetworkStats) string {
	totalRate := stats.SendRate + stats.RecvRate
	
	switch {
	case totalRate >= 10*1024*1024: // >= 10 MB/s - High activity
		return m.styleManager.RenderCriticalText(text)
	case totalRate >= 1*1024*1024: // >= 1 MB/s - Medium activity
		return m.styleManager.RenderWarningText(text)
	case totalRate > 0: // Any activity
		return lipgloss.NewStyle().Foreground(m.styleManager.GetUsageColor(50)).Render(text)
	default: // No activity
		return m.styleManager.RenderMutedText(text)
	}
}

// formatRate converts bytes per second to human-readable format
func (m NetworkModel) formatRate(bytesPerSec float64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytesPerSec >= GB:
		return fmt.Sprintf("%.1fGB/s", bytesPerSec/GB)
	case bytesPerSec >= MB:
		return fmt.Sprintf("%.1fMB/s", bytesPerSec/MB)
	case bytesPerSec >= KB:
		return fmt.Sprintf("%.1fKB/s", bytesPerSec/KB)
	case bytesPerSec > 0:
		return fmt.Sprintf("%.0fB/s", bytesPerSec)
	default:
		return "0B/s"
	}
}

// formatBytes converts bytes to human-readable format
func (m NetworkModel) formatBytes(bytes uint64) string {
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
func (m NetworkModel) SetSize(width, height int) NetworkModel {
	m.width = width
	m.height = height
	return m
}

// GetInterfaces returns the current network interface information
func (m NetworkModel) GetInterfaces() []models.NetworkInfo {
	return m.interfaces
}

// GetRates returns the current transfer rates
func (m NetworkModel) GetRates() map[string]models.NetworkStats {
	return m.rates
}

// GetTotalSendRate returns the total send rate across all interfaces
func (m NetworkModel) GetTotalSendRate() float64 {
	var total float64
	for _, stats := range m.rates {
		total += stats.SendRate
	}
	return total
}

// GetTotalRecvRate returns the total receive rate across all interfaces
func (m NetworkModel) GetTotalRecvRate() float64 {
	var total float64
	for _, stats := range m.rates {
		total += stats.RecvRate
	}
	return total
}

// GetHighActivityInterfaces returns interfaces with high network activity (>= 1MB/s)
func (m NetworkModel) GetHighActivityInterfaces() []string {
	var highActivity []string
	for iface, stats := range m.rates {
		if stats.SendRate+stats.RecvRate >= 1*1024*1024 { // >= 1 MB/s
			highActivity = append(highActivity, iface)
		}
	}
	return highActivity
}

// HasHighActivity returns true if any interface has high network activity
func (m NetworkModel) HasHighActivity() bool {
	return len(m.GetHighActivityInterfaces()) > 0
}

// GetInterfaceByName returns network info for a specific interface
func (m NetworkModel) GetInterfaceByName(name string) (models.NetworkInfo, bool) {
	for _, iface := range m.interfaces {
		if iface.Interface == name {
			return iface, true
		}
	}
	return models.NetworkInfo{}, false
}

// GetRateByInterface returns transfer rates for a specific interface
func (m NetworkModel) GetRateByInterface(name string) (models.NetworkStats, bool) {
	stats, exists := m.rates[name]
	return stats, exists
}

// HasError returns whether the component has an error
func (m NetworkModel) HasError() bool {
	return m.hasError
}

// GetErrorMessage returns the current error message
func (m NetworkModel) GetErrorMessage() string {
	return m.errorMessage
}

// ClearError clears the current error state
func (m NetworkModel) ClearError() NetworkModel {
	m.hasError = false
	m.errorMessage = ""
	return m
}

// SetError sets an error state for the component
func (m NetworkModel) SetError(message string) NetworkModel {
	m.hasError = true
	m.errorMessage = message
	m.lastError = time.Now()
	return m
}