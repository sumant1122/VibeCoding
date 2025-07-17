package models

import (
	tea "github.com/charmbracelet/bubbletea"
)

// SystemCollector interface abstracts system information gathering
type SystemCollector interface {
	CollectCPU() (CPUInfo, error)
	CollectMemory() (MemoryInfo, error)
	CollectDisk() ([]DiskInfo, error)
	CollectNetwork() ([]NetworkInfo, error)
	CalculateNetworkRates(previous, current []NetworkInfo) map[string]NetworkStats
}

// ResourceModel interface for consistent component behavior
type ResourceModel interface {
	Update(tea.Msg) (ResourceModel, tea.Cmd)
	View() string
	Init() tea.Cmd
}