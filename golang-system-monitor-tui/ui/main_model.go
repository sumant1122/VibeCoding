package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"golang-system-monitor-tui/models"
	"golang-system-monitor-tui/services"
)

// FocusedComponent represents which component is currently focused
type FocusedComponent int

const (
	FocusCPU FocusedComponent = iota
	FocusMemory
	FocusDisk
	FocusNetwork
)

// KeyMap defines the keyboard shortcuts
type KeyMap struct {
	Up       []string
	Down     []string
	Left     []string
	Right    []string
	Tab      []string
	ShiftTab []string
	Quit     []string
	Refresh  []string
	Help     []string
}

// DefaultKeyMap returns the default key mappings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:       []string{"up", "k"},
		Down:     []string{"down", "j"},
		Left:     []string{"left", "h"},
		Right:    []string{"right", "l"},
		Tab:      []string{"tab"},
		ShiftTab: []string{"shift+tab"},
		Quit:     []string{"q", "ctrl+c"},
		Refresh:  []string{"r"},
		Help:     []string{"?", "h"},
	}
}

// TickMsg represents a ticker message for real-time updates
type TickMsg time.Time

// MainModel represents the main application model integrating all components
type MainModel struct {
	cpu     CPUModel
	memory  MemoryModel
	disk    DiskModel
	network NetworkModel
	focused FocusedComponent
	keys    KeyMap
	width   int
	height  int
	showHelp bool
	styleManager *StyleManager
	collector models.SystemCollector
	ticker   *time.Ticker
	updateInterval time.Duration
}

// NewMainModel creates a new main application model
func NewMainModel() MainModel {
	styleManager := NewStyleManager()
	collector := services.NewGopsutilCollector()
	return MainModel{
		cpu:            NewCPUModel(),
		memory:         NewMemoryModel(),
		disk:           NewDiskModel(),
		network:        NewNetworkModel(),
		focused:        FocusCPU,
		keys:           DefaultKeyMap(),
		width:          80,
		height:         24,
		showHelp:       false,
		styleManager:   styleManager,
		collector:      collector,
		updateInterval: time.Second, // 1-second update interval
	}
}

// NewMainModelWithConfig creates a new main application model with custom configuration
func NewMainModelWithConfig(updateInterval time.Duration) MainModel {
	styleManager := NewStyleManager()
	collector := services.NewGopsutilCollector()
	return MainModel{
		cpu:            NewCPUModel(),
		memory:         NewMemoryModel(),
		disk:           NewDiskModel(),
		network:        NewNetworkModel(),
		focused:        FocusCPU,
		keys:           DefaultKeyMap(),
		width:          80,
		height:         24,
		showHelp:       false,
		styleManager:   styleManager,
		collector:      collector,
		updateInterval: updateInterval,
	}
}

// Init initializes the main model
func (m MainModel) Init() tea.Cmd {
	return tea.Batch(
		m.cpu.Init(),
		m.memory.Init(),
		m.disk.Init(),
		m.network.Init(),
		m.tickCmd(), // Start the ticker for real-time updates
		m.collectAllDataCmd(), // Initial data collection
	)
}

// Update handles messages and updates the main model state
func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle terminal resize
		m.width = msg.Width
		m.height = msg.Height
		m.styleManager.SetDimensions(m.width, m.height)
		m = m.updateComponentSizes()

	case tea.KeyMsg:
		// Handle keyboard input
		switch {
		case m.containsKey(m.keys.Quit, msg.String()):
			return m, tea.Quit

		case m.containsKey(m.keys.Help, msg.String()):
			m.showHelp = !m.showHelp

		case m.containsKey(m.keys.Refresh, msg.String()):
			// Manual refresh - trigger immediate data collection
			cmds = append(cmds, m.collectAllDataCmd())

		case m.containsKey(m.keys.Tab, msg.String()):
			m.focused = m.nextFocus()

		case m.containsKey(m.keys.ShiftTab, msg.String()):
			m.focused = m.prevFocus()

		case m.containsKey(m.keys.Right, msg.String()):
			m.focused = m.nextFocus()

		case m.containsKey(m.keys.Left, msg.String()):
			m.focused = m.prevFocus()

		case m.containsKey(m.keys.Down, msg.String()):
			m.focused = m.downFocus()

		case m.containsKey(m.keys.Up, msg.String()):
			m.focused = m.upFocus()
		}

	case CPUUpdateMsg:
		var cmd tea.Cmd
		m.cpu, cmd = m.cpu.Update(msg)
		cmds = append(cmds, cmd)

	case MemoryUpdateMsg:
		var cmd tea.Cmd
		m.memory, cmd = m.memory.Update(msg)
		cmds = append(cmds, cmd)

	case DiskUpdateMsg:
		var cmd tea.Cmd
		m.disk, cmd = m.disk.Update(msg)
		cmds = append(cmds, cmd)

	case NetworkUpdateMsg:
		var cmd tea.Cmd
		m.network, cmd = m.network.Update(msg)
		cmds = append(cmds, cmd)

	case TickMsg:
		// Handle ticker for real-time updates
		cmds = append(cmds, m.collectAllDataCmd()) // Collect new data
		cmds = append(cmds, m.tickCmd())           // Schedule next tick

	case models.ErrorMsg:
		// Forward error messages to appropriate components
		var cmd tea.Cmd
		switch msg.Component {
		case "CPU":
			m.cpu, cmd = m.cpu.Update(msg)
		case "Memory":
			m.memory, cmd = m.memory.Update(msg)
		case "Disk":
			m.disk, cmd = m.disk.Update(msg)
		case "Network":
			m.network, cmd = m.network.Update(msg)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View renders the main application view
func (m MainModel) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	// Calculate component dimensions using style manager
	componentWidth, componentHeight := m.styleManager.CalculateComponentDimensions()

	// Update component sizes
	m.cpu = m.cpu.SetSize(componentWidth, componentHeight)
	m.memory = m.memory.SetSize(componentWidth, componentHeight)
	m.disk = m.disk.SetSize(componentWidth, componentHeight)
	m.network = m.network.SetSize(componentWidth, componentHeight)

	// Render components with focus styling using style manager
	cpuView := m.styleManager.RenderComponentBorder(m.cpu.View(), m.focused == FocusCPU, componentWidth, componentHeight)
	memoryView := m.styleManager.RenderComponentBorder(m.memory.View(), m.focused == FocusMemory, componentWidth, componentHeight)
	diskView := m.styleManager.RenderComponentBorder(m.disk.View(), m.focused == FocusDisk, componentWidth, componentHeight)
	networkView := m.styleManager.RenderComponentBorder(m.network.View(), m.focused == FocusNetwork, componentWidth, componentHeight)

	// Create responsive layout using style manager
	components := []string{cpuView, memoryView, diskView, networkView}
	content := m.styleManager.RenderResponsiveLayout(components)

	// Add header and footer using style manager
	header := m.styleManager.RenderApplicationHeader("System Monitor")
	shortcuts := []string{"q: quit", "arrows/tab: navigate", "r: refresh", "?: help"}
	footer := m.styleManager.RenderApplicationFooter(shortcuts)

	return lipgloss.JoinVertical(lipgloss.Left, header, "", content, "", footer)
}



// renderHelp renders the help screen
func (m MainModel) renderHelp() string {
	helpContent := []string{
		"System Monitor - Keyboard Shortcuts",
		"",
		"Navigation:",
		"  ↑/↓/←/→, hjkl  Navigate between components",
		"  Tab, Shift+Tab  Cycle through components",
		"",
		"Actions:",
		"  q, Ctrl+C       Quit application",
		"  r               Manual refresh",
		"  ?, h            Toggle this help",
		"",
		"Components:",
		"  CPU             Real-time CPU usage per core",
		"  Memory          RAM and swap usage",
		"  Disk            Filesystem usage and warnings",
		"  Network         Interface activity and rates",
		"",
		"Press any key to return to the main view",
	}

	content := strings.Join(helpContent, "\n")
	return m.styleManager.RenderHelpScreen(content)
}

// updateComponentSizes updates all component sizes based on current terminal size
func (m MainModel) updateComponentSizes() MainModel {
	componentWidth := (m.width - 3) / 2
	componentHeight := (m.height - 4) / 2

	m.cpu = m.cpu.SetSize(componentWidth, componentHeight)
	m.memory = m.memory.SetSize(componentWidth, componentHeight)
	m.disk = m.disk.SetSize(componentWidth, componentHeight)
	m.network = m.network.SetSize(componentWidth, componentHeight)

	return m
}

// nextFocus returns the next focus component in sequence
func (m MainModel) nextFocus() FocusedComponent {
	switch m.focused {
	case FocusCPU:
		return FocusMemory
	case FocusMemory:
		return FocusDisk
	case FocusDisk:
		return FocusNetwork
	case FocusNetwork:
		return FocusCPU
	default:
		return FocusCPU
	}
}

// prevFocus returns the previous focus component in sequence
func (m MainModel) prevFocus() FocusedComponent {
	switch m.focused {
	case FocusCPU:
		return FocusNetwork
	case FocusMemory:
		return FocusCPU
	case FocusDisk:
		return FocusMemory
	case FocusNetwork:
		return FocusDisk
	default:
		return FocusCPU
	}
}

// downFocus handles down arrow navigation (top row to bottom row)
func (m MainModel) downFocus() FocusedComponent {
	switch m.focused {
	case FocusCPU:
		return FocusDisk
	case FocusMemory:
		return FocusNetwork
	case FocusDisk:
		return FocusDisk // Stay on disk if already on bottom row
	case FocusNetwork:
		return FocusNetwork // Stay on network if already on bottom row
	default:
		return FocusCPU
	}
}

// upFocus handles up arrow navigation (bottom row to top row)
func (m MainModel) upFocus() FocusedComponent {
	switch m.focused {
	case FocusCPU:
		return FocusCPU // Stay on CPU if already on top row
	case FocusMemory:
		return FocusMemory // Stay on memory if already on top row
	case FocusDisk:
		return FocusCPU
	case FocusNetwork:
		return FocusMemory
	default:
		return FocusCPU
	}
}

// containsKey checks if a key string is in the provided key list
func (m MainModel) containsKey(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}

// GetFocusedComponent returns the currently focused component
func (m MainModel) GetFocusedComponent() FocusedComponent {
	return m.focused
}

// SetFocusedComponent sets the focused component
func (m MainModel) SetFocusedComponent(focus FocusedComponent) MainModel {
	m.focused = focus
	return m
}

// GetCPUModel returns the CPU model
func (m MainModel) GetCPUModel() CPUModel {
	return m.cpu
}

// GetMemoryModel returns the memory model
func (m MainModel) GetMemoryModel() MemoryModel {
	return m.memory
}

// GetDiskModel returns the disk model
func (m MainModel) GetDiskModel() DiskModel {
	return m.disk
}

// GetNetworkModel returns the network model
func (m MainModel) GetNetworkModel() NetworkModel {
	return m.network
}

// IsShowingHelp returns whether the help screen is currently displayed
func (m MainModel) IsShowingHelp() bool {
	return m.showHelp
}

// SetShowHelp sets the help display state
func (m MainModel) SetShowHelp(show bool) MainModel {
	m.showHelp = show
	return m
}

// tickCmd creates a command that sends a TickMsg after the update interval
func (m MainModel) tickCmd() tea.Cmd {
	return tea.Tick(m.updateInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// collectAllDataCmd creates a batch command to collect all system data concurrently
func (m MainModel) collectAllDataCmd() tea.Cmd {
	return tea.Batch(
		m.collectCPUDataCmd(),
		m.collectMemoryDataCmd(),
		m.collectDiskDataCmd(),
		m.collectNetworkDataCmd(),
	)
}

// collectCPUDataCmd creates a command to collect CPU data in a goroutine
func (m MainModel) collectCPUDataCmd() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		cpuInfo, err := m.collector.CollectCPU()
		if err != nil {
			return err
		}
		return CPUUpdateMsg(cpuInfo)
	})
}

// collectMemoryDataCmd creates a command to collect memory data in a goroutine
func (m MainModel) collectMemoryDataCmd() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		memoryInfo, err := m.collector.CollectMemory()
		if err != nil {
			return err
		}
		return MemoryUpdateMsg(memoryInfo)
	})
}

// collectDiskDataCmd creates a command to collect disk data in a goroutine
func (m MainModel) collectDiskDataCmd() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		diskInfo, err := m.collector.CollectDisk()
		if err != nil {
			return err
		}
		return DiskUpdateMsg(diskInfo)
	})
}

// collectNetworkDataCmd creates a command to collect network data in a goroutine
func (m MainModel) collectNetworkDataCmd() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		networkInfo, err := m.collector.CollectNetwork()
		if err != nil {
			return err
		}
		return NetworkUpdateMsg(networkInfo)
	})
}