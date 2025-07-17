package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"golang-system-monitor-tui/models"
	"golang-system-monitor-tui/services"
	"golang-system-monitor-tui/ui"
)

// BenchmarkSystemDataCollection benchmarks the core data collection operations
func BenchmarkSystemDataCollection(b *testing.B) {
	collector := services.NewGopsutilCollector()

	b.Run("CPU Collection", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := collector.CollectCPU()
			if err != nil {
				b.Fatalf("CPU collection failed: %v", err)
			}
		}
	})

	b.Run("Memory Collection", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := collector.CollectMemory()
			if err != nil {
				b.Fatalf("Memory collection failed: %v", err)
			}
		}
	})

	b.Run("Disk Collection", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := collector.CollectDisk()
			if err != nil {
				b.Fatalf("Disk collection failed: %v", err)
			}
		}
	})

	b.Run("Network Collection", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := collector.CollectNetwork()
			if err != nil {
				b.Fatalf("Network collection failed: %v", err)
			}
		}
	})

	b.Run("Full Collection Cycle", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			collector.CollectCPU()
			collector.CollectMemory()
			collector.CollectDisk()
			collector.CollectNetwork()
		}
	})
}

// BenchmarkUIComponents benchmarks UI component operations
func BenchmarkUIComponents(b *testing.B) {
	// Create sample data for benchmarking
	cpuInfo := models.CPUInfo{
		Cores: 8,
		Usage: []float64{25.5, 30.2, 45.8, 60.1, 15.3, 80.7, 35.4, 50.9},
		Total: 42.7,
		Timestamp: time.Now(),
	}

	memInfo := models.MemoryInfo{
		Total:     16 * 1024 * 1024 * 1024, // 16GB
		Used:      8 * 1024 * 1024 * 1024,  // 8GB
		Available: 8 * 1024 * 1024 * 1024,  // 8GB
		Swap: models.SwapInfo{
			Total: 4 * 1024 * 1024 * 1024, // 4GB
			Used:  1 * 1024 * 1024 * 1024, // 1GB
			Free:  3 * 1024 * 1024 * 1024, // 3GB
		},
		Timestamp: time.Now(),
	}

	diskInfos := []models.DiskInfo{
		{
			Device:      "/dev/sda1",
			Mountpoint:  "/",
			Filesystem:  "ext4",
			Total:       500 * 1024 * 1024 * 1024, // 500GB
			Used:        300 * 1024 * 1024 * 1024, // 300GB
			Available:   200 * 1024 * 1024 * 1024, // 200GB
			UsedPercent: 60.0,
		},
		{
			Device:      "/dev/sda2",
			Mountpoint:  "/home",
			Filesystem:  "ext4",
			Total:       1000 * 1024 * 1024 * 1024, // 1TB
			Used:        250 * 1024 * 1024 * 1024,  // 250GB
			Available:   750 * 1024 * 1024 * 1024,  // 750GB
			UsedPercent: 25.0,
		},
	}

	networkInfos := []models.NetworkInfo{
		{
			Interface:   "eth0",
			BytesSent:   1024 * 1024 * 100, // 100MB
			BytesRecv:   1024 * 1024 * 200, // 200MB
			PacketsSent: 50000,
			PacketsRecv: 75000,
			Timestamp:   time.Now(),
		},
		{
			Interface:   "wlan0",
			BytesSent:   1024 * 1024 * 50, // 50MB
			BytesRecv:   1024 * 1024 * 80, // 80MB
			PacketsSent: 25000,
			PacketsRecv: 40000,
			Timestamp:   time.Now(),
		},
	}

	b.Run("CPU Model Update", func(b *testing.B) {
		cpuModel := ui.NewCPUModel()
		updateMsg := ui.CPUUpdateMsg(cpuInfo)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cpuModel.Update(updateMsg)
		}
	})

	b.Run("Memory Model Update", func(b *testing.B) {
		memoryModel := ui.NewMemoryModel()
		updateMsg := ui.MemoryUpdateMsg(memInfo)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			memoryModel.Update(updateMsg)
		}
	})

	b.Run("Disk Model Update", func(b *testing.B) {
		diskModel := ui.NewDiskModel()
		updateMsg := ui.DiskUpdateMsg(diskInfos)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			diskModel.Update(updateMsg)
		}
	})

	b.Run("Network Model Update", func(b *testing.B) {
		networkModel := ui.NewNetworkModel()
		updateMsg := ui.NetworkUpdateMsg(networkInfos)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			networkModel.Update(updateMsg)
		}
	})

	b.Run("CPU Model View Rendering", func(b *testing.B) {
		cpuModel := ui.NewCPUModel()
		cpuModel.Update(ui.CPUUpdateMsg(cpuInfo))
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cpuModel.View()
		}
	})

	b.Run("Memory Model View Rendering", func(b *testing.B) {
		memoryModel := ui.NewMemoryModel()
		memoryModel.Update(ui.MemoryUpdateMsg(memInfo))
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = memoryModel.View()
		}
	})

	b.Run("Disk Model View Rendering", func(b *testing.B) {
		diskModel := ui.NewDiskModel()
		diskModel.Update(ui.DiskUpdateMsg(diskInfos))
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = diskModel.View()
		}
	})

	b.Run("Network Model View Rendering", func(b *testing.B) {
		networkModel := ui.NewNetworkModel()
		networkModel.Update(ui.NetworkUpdateMsg(networkInfos))
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = networkModel.View()
		}
	})
}

// BenchmarkMainModel benchmarks the main application model operations
func BenchmarkMainModel(b *testing.B) {
	mainModel := ui.NewMainModel()

	// Sample data for updates
	cpuInfo := models.CPUInfo{
		Cores: 4,
		Usage: []float64{25.0, 50.0, 75.0, 90.0},
		Total: 60.0,
		Timestamp: time.Now(),
	}

	memInfo := models.MemoryInfo{
		Total:     8 * 1024 * 1024 * 1024, // 8GB
		Used:      4 * 1024 * 1024 * 1024, // 4GB
		Available: 4 * 1024 * 1024 * 1024, // 4GB
		Timestamp: time.Now(),
	}

	b.Run("Main Model Update CPU", func(b *testing.B) {
		updateMsg := ui.CPUUpdateMsg(cpuInfo)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mainModel.Update(updateMsg)
		}
	})

	b.Run("Main Model Update Memory", func(b *testing.B) {
		updateMsg := ui.MemoryUpdateMsg(memInfo)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mainModel.Update(updateMsg)
		}
	})

	b.Run("Main Model View Rendering", func(b *testing.B) {
		// Update with sample data first
		mainModel.Update(ui.CPUUpdateMsg(cpuInfo))
		mainModel.Update(ui.MemoryUpdateMsg(memInfo))
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = mainModel.View()
		}
	})

	b.Run("Main Model Navigation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Cycle through all focus states
			mainModel = mainModel.SetFocusedComponent(ui.FocusCPU)
			mainModel = mainModel.SetFocusedComponent(ui.FocusMemory)
			mainModel = mainModel.SetFocusedComponent(ui.FocusDisk)
			mainModel = mainModel.SetFocusedComponent(ui.FocusNetwork)
		}
	})
}

// BenchmarkNetworkRateCalculation benchmarks network rate calculations
func BenchmarkNetworkRateCalculation(b *testing.B) {
	collector := services.NewGopsutilCollector()
	
	// Create sample network data for rate calculation
	baseTime := time.Now()
	previous := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 1000000,
			BytesRecv: 2000000,
			Timestamp: baseTime,
		},
		{
			Interface: "wlan0",
			BytesSent: 500000,
			BytesRecv: 1500000,
			Timestamp: baseTime,
		},
	}
	
	current := []models.NetworkInfo{
		{
			Interface: "eth0",
			BytesSent: 2000000,
			BytesRecv: 4000000,
			Timestamp: baseTime.Add(time.Second),
		},
		{
			Interface: "wlan0",
			BytesSent: 1000000,
			BytesRecv: 2500000,
			Timestamp: baseTime.Add(time.Second),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = collector.CalculateNetworkRates(previous, current)
	}
}

// BenchmarkErrorHandling benchmarks error handling operations
func BenchmarkErrorHandling(b *testing.B) {
	logger := log.New(os.Stderr, "", 0)
	errorHandler := models.NewErrorHandler(logger)

	b.Run("Error Handler Processing", func(b *testing.B) {
		testErr := fmt.Errorf("test error")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			errorHandler.HandlePermissionError("CPU", testErr)
		}
	})

	b.Run("Error Creation", func(b *testing.B) {
		testErr := fmt.Errorf("test error")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = models.CreateSystemError(models.PermissionError, "CPU", "test error", testErr)
		}
	})
}

// BenchmarkApplicationStartupShutdown benchmarks application lifecycle
func BenchmarkApplicationStartupShutdown(b *testing.B) {
	b.Run("Program Creation", func(b *testing.B) {
		config := &Config{
			UpdateInterval: time.Second,
			NoMouse:        true,
			NoAltScreen:    true,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			program := createProgram(config)
			program.Kill()
		}
	})

	b.Run("Configuration Parsing", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			config := &Config{
				UpdateInterval: 500 * time.Millisecond,
				LogFile:        "",
				Debug:          false,
				NoMouse:        true,
				NoAltScreen:    true,
			}
			_ = config
		}
	})
}

// BenchmarkMemoryUsage benchmarks memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("CPU Model Memory", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cpuModel := ui.NewCPUModel()
			cpuInfo := models.CPUInfo{
				Cores: 8,
				Usage: make([]float64, 8),
				Total: 50.0,
				Timestamp: time.Now(),
			}
			cpuModel.Update(ui.CPUUpdateMsg(cpuInfo))
		}
	})

	b.Run("Network Model Memory", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			networkModel := ui.NewNetworkModel()
			networkInfos := make([]models.NetworkInfo, 5)
			for j := range networkInfos {
				networkInfos[j] = models.NetworkInfo{
					Interface: "eth" + string(rune('0'+j)),
					BytesSent: uint64(j * 1000000),
					BytesRecv: uint64(j * 2000000),
					Timestamp: time.Now(),
				}
			}
			networkModel.Update(ui.NetworkUpdateMsg(networkInfos))
		}
	})

	b.Run("Disk Model Memory", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			diskModel := ui.NewDiskModel()
			diskInfos := make([]models.DiskInfo, 3)
			for j := range diskInfos {
				diskInfos[j] = models.DiskInfo{
					Device:      "/dev/sd" + string(rune('a'+j)) + "1",
					Mountpoint:  "/" + string(rune('a'+j)),
					Filesystem:  "ext4",
					Total:       uint64((j + 1) * 100 * 1024 * 1024 * 1024),
					Used:        uint64((j + 1) * 50 * 1024 * 1024 * 1024),
					Available:   uint64((j + 1) * 50 * 1024 * 1024 * 1024),
					UsedPercent: 50.0,
				}
			}
			diskModel.Update(ui.DiskUpdateMsg(diskInfos))
		}
	})
}

// BenchmarkConcurrentOperations benchmarks concurrent data collection
func BenchmarkConcurrentOperations(b *testing.B) {
	collector := services.NewGopsutilCollector()

	b.Run("Concurrent Data Collection", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Simulate concurrent data collection
				go collector.CollectCPU()
				go collector.CollectMemory()
				go collector.CollectDisk()
				go collector.CollectNetwork()
			}
		})
	})

	b.Run("Concurrent UI Updates", func(b *testing.B) {
		mainModel := ui.NewMainModel()
		cpuInfo := models.CPUInfo{
			Cores: 4,
			Usage: []float64{25.0, 50.0, 75.0, 90.0},
			Total: 60.0,
			Timestamp: time.Now(),
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mainModel.Update(ui.CPUUpdateMsg(cpuInfo))
			}
		})
	})
}

// BenchmarkRealTimeUpdates benchmarks real-time update performance
func BenchmarkRealTimeUpdates(b *testing.B) {
	collector := services.NewGopsutilCollector()
	mainModel := ui.NewMainModel()

	b.Run("Full Update Cycle", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate a complete update cycle
			if cpuInfo, err := collector.CollectCPU(); err == nil {
				mainModel.Update(ui.CPUUpdateMsg(cpuInfo))
			}
			if memInfo, err := collector.CollectMemory(); err == nil {
				mainModel.Update(ui.MemoryUpdateMsg(memInfo))
			}
			if diskInfos, err := collector.CollectDisk(); err == nil {
				mainModel.Update(ui.DiskUpdateMsg(diskInfos))
			}
			if networkInfos, err := collector.CollectNetwork(); err == nil {
				mainModel.Update(ui.NetworkUpdateMsg(networkInfos))
			}
		}
	})

	b.Run("Update and Render Cycle", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Collect data
			if cpuInfo, err := collector.CollectCPU(); err == nil {
				mainModel.Update(ui.CPUUpdateMsg(cpuInfo))
			}
			// Render view
			_ = mainModel.View()
		}
	})
}

// BenchmarkStringFormatting benchmarks string formatting operations used in UI
func BenchmarkStringFormatting(b *testing.B) {
	b.Run("Memory Bytes Formatting", func(b *testing.B) {
		memoryModel := ui.NewMemoryModel()
		values := []uint64{
			1024,                    // 1KB
			1024 * 1024,            // 1MB
			1024 * 1024 * 1024,     // 1GB
			1024 * 1024 * 1024 * 1024, // 1TB
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, val := range values {
				_ = fmt.Sprintf("%.1f GB", float64(val)/(1024*1024*1024))
			}
		}
		_ = memoryModel // Use the model to avoid unused variable
	})

	b.Run("Network Rate Formatting", func(b *testing.B) {
		networkModel := ui.NewNetworkModel()
		rates := []float64{1024.5, 1048576.7, 1073741824.9}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, rate := range rates {
				_ = fmt.Sprintf("%.1f MB/s", rate/(1024*1024))
			}
		}
		_ = networkModel // Use the model to avoid unused variable
	})

	b.Run("Disk Bytes Formatting", func(b *testing.B) {
		diskModel := ui.NewDiskModel()
		values := []uint64{
			1024 * 1024 * 1024,     // 1GB
			1024 * 1024 * 1024 * 1024, // 1TB
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, val := range values {
				_ = fmt.Sprintf("%.1f GB", float64(val)/(1024*1024*1024))
			}
		}
		_ = diskModel // Use the model to avoid unused variable
	})
}

// BenchmarkContextOperations benchmarks context-related operations
func BenchmarkContextOperations(b *testing.B) {
	b.Run("Context Creation and Cancellation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			cancel()
			_ = ctx
		}
	})

	b.Run("Context with Values", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.WithValue(context.Background(), "key", "value")
			_ = ctx.Value("key")
		}
	})
}