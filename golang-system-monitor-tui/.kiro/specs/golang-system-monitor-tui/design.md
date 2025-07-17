# Design Document

## Overview

The Golang System Monitor TUI application will be built using the Bubble Tea framework for TUI development and Lipgloss for styling. The application follows a Model-View-Update (MVU) architecture pattern, where the application state is managed centrally and updates flow through a predictable cycle. The system will collect resource metrics using Go's standard library and third-party packages like gopsutil for cross-platform system information gathering.

## Architecture

The application follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────┐
│           TUI Layer                 │
│  (Bubble Tea + Lipgloss)           │
├─────────────────────────────────────┤
│         Application Layer           │
│    (Models, Views, Updates)        │
├─────────────────────────────────────┤
│         Service Layer               │
│   (Resource Collectors)            │
├─────────────────────────────────────┤
│         Data Layer                  │
│  (System APIs + gopsutil)          │
└─────────────────────────────────────┘
```

### Key Architectural Decisions

1. **Bubble Tea Framework**: Chosen for its mature TUI capabilities and clean MVU pattern
2. **gopsutil Library**: Cross-platform system information gathering
3. **Concurrent Data Collection**: Background goroutines for non-blocking resource updates
4. **Component-Based UI**: Modular components for each resource type (CPU, Memory, Disk, Network)

## Components and Interfaces

### Core Components

#### 1. Main Application Model
```go
type Model struct {
    cpu     CPUModel
    memory  MemoryModel
    disk    DiskModel
    network NetworkModel
    focused int
    help    help.Model
    keys    KeyMap
}
```

#### 2. Resource Models
Each resource type has its own model implementing a common interface:

```go
type ResourceModel interface {
    Update(tea.Msg) (ResourceModel, tea.Cmd)
    View() string
    Init() tea.Cmd
}

type CPUModel struct {
    usage    []float64  // Per-core usage
    history  [][]float64 // Historical data for graphs
    total    float64    // Overall CPU usage
}

type MemoryModel struct {
    total     uint64
    used      uint64
    available uint64
    swap      SwapInfo
}

type DiskModel struct {
    filesystems []FilesystemInfo
}

type NetworkModel struct {
    interfaces []NetworkInterface
    stats      map[string]NetworkStats
}
```

#### 3. Data Collection Services
```go
type SystemCollector interface {
    CollectCPU() (CPUInfo, error)
    CollectMemory() (MemoryInfo, error)
    CollectDisk() ([]DiskInfo, error)
    CollectNetwork() ([]NetworkInfo, error)
}

type GopsutilCollector struct{}
```

#### 4. UI Components
```go
type ComponentRenderer interface {
    RenderCPU(CPUModel) string
    RenderMemory(MemoryModel) string
    RenderDisk(DiskModel) string
    RenderNetwork(NetworkModel) string
    RenderHelp(KeyMap) string
}
```

### Interfaces

#### System Data Collection
- `SystemCollector`: Abstracts system information gathering
- `ResourceModel`: Common interface for all resource display models
- `ComponentRenderer`: Handles visual rendering of components

#### Message Types
```go
type TickMsg time.Time
type CPUUpdateMsg CPUInfo
type MemoryUpdateMsg MemoryInfo
type DiskUpdateMsg []DiskInfo
type NetworkUpdateMsg []NetworkInfo
type ErrorMsg error
```

## Data Models

### CPU Information
```go
type CPUInfo struct {
    Cores     int
    Usage     []float64 // Per-core usage percentages
    Total     float64   // Overall usage percentage
    Timestamp time.Time
}
```

### Memory Information
```go
type MemoryInfo struct {
    Total     uint64
    Used      uint64
    Available uint64
    Swap      SwapInfo
    Timestamp time.Time
}

type SwapInfo struct {
    Total uint64
    Used  uint64
    Free  uint64
}
```

### Disk Information
```go
type DiskInfo struct {
    Device     string
    Mountpoint string
    Filesystem string
    Total      uint64
    Used       uint64
    Available  uint64
    UsedPercent float64
}
```

### Network Information
```go
type NetworkInfo struct {
    Interface string
    BytesSent uint64
    BytesRecv uint64
    PacketsSent uint64
    PacketsRecv uint64
    Timestamp time.Time
}

type NetworkStats struct {
    SendRate float64 // Bytes per second
    RecvRate float64 // Bytes per second
}
```

## Error Handling

### Error Categories
1. **System Access Errors**: Permission denied, unavailable resources
2. **Data Collection Errors**: Temporary failures in gathering metrics
3. **Rendering Errors**: Terminal size issues, display problems

### Error Handling Strategy
```go
type ErrorHandler struct {
    logger *log.Logger
}

func (e *ErrorHandler) HandleSystemError(err error) tea.Cmd
func (e *ErrorHandler) HandleDataError(err error) tea.Cmd
func (e *ErrorHandler) HandleRenderError(err error) tea.Cmd
```

- Non-critical errors display "N/A" for affected metrics
- Critical errors show error messages but keep the application running
- All errors are logged for debugging purposes
- Graceful degradation when some system information is unavailable

## Testing Strategy

### Unit Testing
- **Resource Models**: Test state transitions and data handling
- **Data Collectors**: Mock system calls and test data parsing
- **UI Components**: Test rendering logic with sample data
- **Error Handlers**: Test error scenarios and recovery

### Integration Testing
- **End-to-End Flow**: Test complete data collection and display cycle
- **System Integration**: Test with actual system resources
- **Terminal Compatibility**: Test across different terminal types

### Test Structure
```go
// Unit tests for each component
func TestCPUModel_Update(t *testing.T)
func TestMemoryCollector_Collect(t *testing.T)
func TestDiskRenderer_Render(t *testing.T)

// Integration tests
func TestSystemMonitor_FullCycle(t *testing.T)
func TestErrorRecovery_SystemUnavailable(t *testing.T)
```

### Mocking Strategy
- Mock `SystemCollector` interface for predictable testing
- Mock terminal dimensions for UI testing
- Mock system calls for error scenario testing

## Visual Design

### Layout Structure
```
┌─────────────────────────────────────────────────────────────┐
│                    System Monitor                           │
├─────────────────────────────────────────────────────────────┤
│ CPU Usage                          │ Memory Usage           │
│ ████████████░░░░░░░░ 65%          │ ██████████░░░░ 70%     │
│ Core 1: ████████░░░░ 45%          │ RAM: 8.2GB / 16GB     │
│ Core 2: ████████████░░ 75%        │ Swap: 1.2GB / 4GB     │
├─────────────────────────────────────────────────────────────┤
│ Disk Usage                         │ Network Activity       │
│ /dev/sda1 ████████░░░░ 60%        │ eth0: ↑ 1.2MB/s       │
│ /dev/sdb1 ████░░░░░░░░ 25%        │       ↓ 850KB/s       │
├─────────────────────────────────────────────────────────────┤
│ Press 'q' to quit, 'h' for help, arrows to navigate        │
└─────────────────────────────────────────────────────────────┘
```

### Color Scheme
- **Normal Usage (0-70%)**: Green
- **Warning Usage (70-90%)**: Yellow
- **Critical Usage (90%+)**: Red
- **Headers**: Cyan
- **Help Text**: Gray

### Responsive Design
- Minimum terminal size: 80x24
- Adaptive layout based on terminal dimensions
- Horizontal scrolling for overflow content
- Graceful handling of very small terminals