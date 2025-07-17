package services

import (
	"log"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"golang-system-monitor-tui/models"
)

// GopsutilCollector implements SystemCollector using gopsutil library
type GopsutilCollector struct{
	errorHandler *models.ErrorHandler
}

// NewGopsutilCollector creates a new instance of GopsutilCollector
func NewGopsutilCollector() *GopsutilCollector {
	return &GopsutilCollector{
		errorHandler: models.NewErrorHandler(log.Default()),
	}
}

// NewGopsutilCollectorWithErrorHandler creates a new instance with custom error handler
func NewGopsutilCollectorWithErrorHandler(errorHandler *models.ErrorHandler) *GopsutilCollector {
	return &GopsutilCollector{
		errorHandler: errorHandler,
	}
}

// CollectCPU gathers CPU usage information including per-core and total usage
func (g *GopsutilCollector) CollectCPU() (models.CPUInfo, error) {
	// Get per-core CPU usage percentages
	perCoreUsage, err := cpu.Percent(time.Second, true)
	if err != nil {
		// Categorize the error based on its content
		if g.isPermissionError(err) {
			return models.CPUInfo{}, models.CreateSystemError(models.PermissionError, "CPU", "Permission denied accessing CPU information", err)
		} else if g.isTemporaryError(err) {
			return models.CPUInfo{}, models.CreateSystemError(models.TemporaryError, "CPU", "Temporary error collecting CPU data", err)
		}
		return models.CPUInfo{}, models.CreateSystemError(models.SystemAccessError, "CPU", "Failed to collect per-core CPU usage", err)
	}

	// Get total CPU usage percentage
	totalUsage, err := cpu.Percent(time.Second, false)
	if err != nil {
		// If we have per-core data but total fails, calculate total from per-core
		if len(perCoreUsage) > 0 {
			var sum float64
			for _, usage := range perCoreUsage {
				sum += usage
			}
			total := sum / float64(len(perCoreUsage))
			
			return models.CPUInfo{
				Cores:     len(perCoreUsage),
				Usage:     perCoreUsage,
				Total:     total,
				Timestamp: time.Now(),
			}, nil
		}
		
		// Categorize the error
		if g.isPermissionError(err) {
			return models.CPUInfo{}, models.CreateSystemError(models.PermissionError, "CPU", "Permission denied accessing CPU information", err)
		}
		return models.CPUInfo{}, models.CreateSystemError(models.SystemAccessError, "CPU", "Failed to collect total CPU usage", err)
	}

	var total float64
	if len(totalUsage) > 0 {
		total = totalUsage[0]
	}

	return models.CPUInfo{
		Cores:     len(perCoreUsage),
		Usage:     perCoreUsage,
		Total:     total,
		Timestamp: time.Now(),
	}, nil
}

// CollectMemory gathers memory usage information including RAM and swap
func (g *GopsutilCollector) CollectMemory() (models.MemoryInfo, error) {
	// Get virtual memory statistics
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		// Categorize the error
		if g.isPermissionError(err) {
			return models.MemoryInfo{}, models.CreateSystemError(models.PermissionError, "Memory", "Permission denied accessing memory information", err)
		} else if g.isTemporaryError(err) {
			return models.MemoryInfo{}, models.CreateSystemError(models.TemporaryError, "Memory", "Temporary error collecting memory data", err)
		}
		return models.MemoryInfo{}, models.CreateSystemError(models.SystemAccessError, "Memory", "Failed to collect virtual memory statistics", err)
	}

	// Get swap memory statistics
	swapStat, err := mem.SwapMemory()
	if err != nil {
		// If we have VM stats but swap fails, return VM stats with empty swap
		if vmStat != nil {
			return models.MemoryInfo{
				Total:     vmStat.Total,
				Used:      vmStat.Used,
				Available: vmStat.Available,
				Swap: models.SwapInfo{
					Total: 0,
					Used:  0,
					Free:  0,
				},
				Timestamp: time.Now(),
			}, nil
		}
		
		// Categorize the error
		if g.isPermissionError(err) {
			return models.MemoryInfo{}, models.CreateSystemError(models.PermissionError, "Memory", "Permission denied accessing swap information", err)
		}
		return models.MemoryInfo{}, models.CreateSystemError(models.SystemAccessError, "Memory", "Failed to collect swap memory statistics", err)
	}

	return models.MemoryInfo{
		Total:     vmStat.Total,
		Used:      vmStat.Used,
		Available: vmStat.Available,
		Swap: models.SwapInfo{
			Total: swapStat.Total,
			Used:  swapStat.Used,
			Free:  swapStat.Free,
		},
		Timestamp: time.Now(),
	}, nil
}

// CollectDisk gathers disk usage information for all mounted filesystems
func (g *GopsutilCollector) CollectDisk() ([]models.DiskInfo, error) {
	// Get disk partitions
	partitions, err := disk.Partitions(false)
	if err != nil {
		// Categorize the error
		if g.isPermissionError(err) {
			return nil, models.CreateSystemError(models.PermissionError, "Disk", "Permission denied accessing disk partitions", err)
		} else if g.isTemporaryError(err) {
			return nil, models.CreateSystemError(models.TemporaryError, "Disk", "Temporary error collecting disk partitions", err)
		}
		return nil, models.CreateSystemError(models.SystemAccessError, "Disk", "Failed to collect disk partitions", err)
	}

	var diskInfos []models.DiskInfo
	var lastError error
	var errorCount int

	for _, partition := range partitions {
		// Skip special filesystems that are not real storage devices
		if partition.Fstype == "proc" || partition.Fstype == "sysfs" || 
		   partition.Fstype == "devtmpfs" || partition.Fstype == "tmpfs" ||
		   partition.Fstype == "devpts" || partition.Fstype == "cgroup" ||
		   partition.Fstype == "cgroup2" || partition.Fstype == "pstore" ||
		   partition.Fstype == "bpf" || partition.Fstype == "tracefs" {
			continue
		}

		// Get usage statistics for each partition
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			// Store the last error but continue processing other partitions
			lastError = err
			errorCount++
			continue
		}

		diskInfo := models.DiskInfo{
			Device:      partition.Device,
			Mountpoint:  partition.Mountpoint,
			Filesystem:  partition.Fstype,
			Total:       usage.Total,
			Used:        usage.Used,
			Available:   usage.Free,
			UsedPercent: usage.UsedPercent,
		}
		diskInfos = append(diskInfos, diskInfo)
	}

	// If we have some disk info but encountered errors, return partial results
	if len(diskInfos) > 0 {
		return diskInfos, nil
	}

	// If we have no disk info and encountered errors, return categorized error
	if lastError != nil {
		if g.isPermissionError(lastError) {
			return nil, models.CreateSystemError(models.PermissionError, "Disk", "Permission denied accessing disk usage information", lastError)
		} else if g.isTemporaryError(lastError) {
			return nil, models.CreateSystemError(models.TemporaryError, "Disk", "Temporary error collecting disk usage", lastError)
		}
		return nil, models.CreateSystemError(models.DataCollectionError, "Disk", "Failed to collect disk usage for any filesystem", lastError)
	}

	// No partitions found (shouldn't happen on normal systems)
	if len(diskInfos) == 0 {
		return nil, models.CreateSystemError(models.SystemAccessError, "Disk", "No accessible disk partitions found", nil)
	}

	return diskInfos, nil
}

// CollectNetwork gathers network interface statistics
func (g *GopsutilCollector) CollectNetwork() ([]models.NetworkInfo, error) {
	// Get network interface statistics
	netStats, err := net.IOCounters(true)
	if err != nil {
		// Categorize the error
		if g.isPermissionError(err) {
			return nil, models.CreateSystemError(models.PermissionError, "Network", "Permission denied accessing network interface statistics", err)
		} else if g.isTemporaryError(err) {
			return nil, models.CreateSystemError(models.TemporaryError, "Network", "Temporary error collecting network data", err)
		}
		return nil, models.CreateSystemError(models.SystemAccessError, "Network", "Failed to collect network interface statistics", err)
	}

	var networkInfos []models.NetworkInfo
	timestamp := time.Now()

	for _, stat := range netStats {
		// Skip loopback interfaces for cleaner output (different names on different platforms)
		if stat.Name == "lo" || stat.Name == "Loopback" || 
		   stat.Name == "Loopback Pseudo-Interface 1" {
			continue
		}

		networkInfo := models.NetworkInfo{
			Interface:   stat.Name,
			BytesSent:   stat.BytesSent,
			BytesRecv:   stat.BytesRecv,
			PacketsSent: stat.PacketsSent,
			PacketsRecv: stat.PacketsRecv,
			Timestamp:   timestamp,
		}
		networkInfos = append(networkInfos, networkInfo)
	}

	// Check if we have any network interfaces
	if len(networkInfos) == 0 {
		return nil, models.CreateSystemError(models.SystemAccessError, "Network", "No accessible network interfaces found", nil)
	}

	return networkInfos, nil
}

// CalculateNetworkRates calculates transfer rates between two network measurements
func (g *GopsutilCollector) CalculateNetworkRates(previous, current []models.NetworkInfo) map[string]models.NetworkStats {
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

// isPermissionError checks if an error is related to permissions
func (g *GopsutilCollector) isPermissionError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "permission denied") ||
		   strings.Contains(errStr, "access denied") ||
		   strings.Contains(errStr, "operation not permitted")
}

// isTemporaryError checks if an error is temporary and might resolve itself
func (g *GopsutilCollector) isTemporaryError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") ||
		   strings.Contains(errStr, "temporary") ||
		   strings.Contains(errStr, "try again") ||
		   strings.Contains(errStr, "resource temporarily unavailable")
}