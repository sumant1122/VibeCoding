package models

import (
	"time"
)

// CPUInfo represents CPU usage information
type CPUInfo struct {
	Cores     int       `json:"cores"`
	Usage     []float64 `json:"usage"`     // Per-core usage percentages
	Total     float64   `json:"total"`     // Overall usage percentage
	Timestamp time.Time `json:"timestamp"`
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Total     uint64    `json:"total"`
	Used      uint64    `json:"used"`
	Available uint64    `json:"available"`
	Swap      SwapInfo  `json:"swap"`
	Timestamp time.Time `json:"timestamp"`
}

// SwapInfo represents swap memory information
type SwapInfo struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
}

// DiskInfo represents disk usage information
type DiskInfo struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Filesystem  string  `json:"filesystem"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"used_percent"`
}

// NetworkInfo represents network interface information
type NetworkInfo struct {
	Interface   string    `json:"interface"`
	BytesSent   uint64    `json:"bytes_sent"`
	BytesRecv   uint64    `json:"bytes_recv"`
	PacketsSent uint64    `json:"packets_sent"`
	PacketsRecv uint64    `json:"packets_recv"`
	Timestamp   time.Time `json:"timestamp"`
}

// NetworkStats represents calculated network statistics
type NetworkStats struct {
	SendRate float64 `json:"send_rate"` // Bytes per second
	RecvRate float64 `json:"recv_rate"` // Bytes per second
}