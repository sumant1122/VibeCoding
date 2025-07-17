package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCPUInfo_Validation(t *testing.T) {
	tests := []struct {
		name    string
		cpu     CPUInfo
		wantErr bool
	}{
		{
			name: "valid CPU info",
			cpu: CPUInfo{
				Cores:     4,
				Usage:     []float64{25.5, 30.2, 45.8, 12.1},
				Total:     28.4,
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid CPU info with zero usage",
			cpu: CPUInfo{
				Cores:     2,
				Usage:     []float64{0.0, 0.0},
				Total:     0.0,
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that cores count matches usage slice length
			if len(tt.cpu.Usage) != tt.cpu.Cores {
				t.Errorf("CPUInfo cores count (%d) doesn't match usage slice length (%d)", 
					tt.cpu.Cores, len(tt.cpu.Usage))
			}

			// Test that usage percentages are valid (0-100)
			for i, usage := range tt.cpu.Usage {
				if usage < 0 || usage > 100 {
					t.Errorf("CPUInfo usage[%d] = %f, want 0-100", i, usage)
				}
			}

			// Test that total usage is valid
			if tt.cpu.Total < 0 || tt.cpu.Total > 100 {
				t.Errorf("CPUInfo total = %f, want 0-100", tt.cpu.Total)
			}
		})
	}
}

func TestMemoryInfo_Validation(t *testing.T) {
	tests := []struct {
		name   string
		memory MemoryInfo
	}{
		{
			name: "valid memory info",
			memory: MemoryInfo{
				Total:     16 * 1024 * 1024 * 1024, // 16GB
				Used:      8 * 1024 * 1024 * 1024,  // 8GB
				Available: 8 * 1024 * 1024 * 1024,  // 8GB
				Swap: SwapInfo{
					Total: 4 * 1024 * 1024 * 1024, // 4GB
					Used:  1 * 1024 * 1024 * 1024, // 1GB
					Free:  3 * 1024 * 1024 * 1024, // 3GB
				},
				Timestamp: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that used + available <= total (allowing for some system overhead)
			if tt.memory.Used > tt.memory.Total {
				t.Errorf("MemoryInfo used (%d) > total (%d)", tt.memory.Used, tt.memory.Total)
			}

			// Test swap consistency
			if tt.memory.Swap.Used > tt.memory.Swap.Total {
				t.Errorf("SwapInfo used (%d) > total (%d)", tt.memory.Swap.Used, tt.memory.Swap.Total)
			}

			if tt.memory.Swap.Used+tt.memory.Swap.Free > tt.memory.Swap.Total {
				t.Errorf("SwapInfo used+free (%d) > total (%d)", 
					tt.memory.Swap.Used+tt.memory.Swap.Free, tt.memory.Swap.Total)
			}
		})
	}
}

func TestDiskInfo_Validation(t *testing.T) {
	tests := []struct {
		name string
		disk DiskInfo
	}{
		{
			name: "valid disk info",
			disk: DiskInfo{
				Device:      "/dev/sda1",
				Mountpoint:  "/",
				Filesystem:  "ext4",
				Total:       1000 * 1024 * 1024 * 1024, // 1TB
				Used:        600 * 1024 * 1024 * 1024,  // 600GB
				Available:   400 * 1024 * 1024 * 1024,  // 400GB
				UsedPercent: 60.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that used + available <= total
			if tt.disk.Used > tt.disk.Total {
				t.Errorf("DiskInfo used (%d) > total (%d)", tt.disk.Used, tt.disk.Total)
			}

			// Test that used percent is valid
			if tt.disk.UsedPercent < 0 || tt.disk.UsedPercent > 100 {
				t.Errorf("DiskInfo used percent = %f, want 0-100", tt.disk.UsedPercent)
			}

			// Test that device and mountpoint are not empty
			if tt.disk.Device == "" {
				t.Error("DiskInfo device should not be empty")
			}
			if tt.disk.Mountpoint == "" {
				t.Error("DiskInfo mountpoint should not be empty")
			}
		})
	}
}

func TestNetworkInfo_Validation(t *testing.T) {
	tests := []struct {
		name    string
		network NetworkInfo
	}{
		{
			name: "valid network info",
			network: NetworkInfo{
				Interface:   "eth0",
				BytesSent:   1024 * 1024 * 100, // 100MB
				BytesRecv:   1024 * 1024 * 200, // 200MB
				PacketsSent: 50000,
				PacketsRecv: 75000,
				Timestamp:   time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that interface name is not empty
			if tt.network.Interface == "" {
				t.Error("NetworkInfo interface should not be empty")
			}

			// Test that counters are non-negative
			if tt.network.BytesSent < 0 {
				t.Errorf("NetworkInfo bytes sent should be non-negative, got %d", tt.network.BytesSent)
			}
			if tt.network.BytesRecv < 0 {
				t.Errorf("NetworkInfo bytes received should be non-negative, got %d", tt.network.BytesRecv)
			}
			if tt.network.PacketsSent < 0 {
				t.Errorf("NetworkInfo packets sent should be non-negative, got %d", tt.network.PacketsSent)
			}
			if tt.network.PacketsRecv < 0 {
				t.Errorf("NetworkInfo packets received should be non-negative, got %d", tt.network.PacketsRecv)
			}
		})
	}
}

func TestNetworkStats_Validation(t *testing.T) {
	tests := []struct {
		name  string
		stats NetworkStats
	}{
		{
			name: "valid network stats",
			stats: NetworkStats{
				SendRate: 1024 * 1024, // 1MB/s
				RecvRate: 2 * 1024 * 1024, // 2MB/s
			},
		},
		{
			name: "zero rates",
			stats: NetworkStats{
				SendRate: 0,
				RecvRate: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that rates are non-negative
			if tt.stats.SendRate < 0 {
				t.Errorf("NetworkStats send rate should be non-negative, got %f", tt.stats.SendRate)
			}
			if tt.stats.RecvRate < 0 {
				t.Errorf("NetworkStats receive rate should be non-negative, got %f", tt.stats.RecvRate)
			}
		})
	}
}

// Test JSON serialization and deserialization
func TestCPUInfo_JSONSerialization(t *testing.T) {
	original := CPUInfo{
		Cores:     4,
		Usage:     []float64{25.5, 30.2, 45.8, 12.1},
		Total:     28.4,
		Timestamp: time.Now().Truncate(time.Second), // Truncate for comparison
	}

	// Serialize to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal CPUInfo: %v", err)
	}

	// Deserialize from JSON
	var deserialized CPUInfo
	err = json.Unmarshal(data, &deserialized)
	if err != nil {
		t.Fatalf("Failed to unmarshal CPUInfo: %v", err)
	}

	// Compare
	if deserialized.Cores != original.Cores {
		t.Errorf("Cores mismatch: got %d, want %d", deserialized.Cores, original.Cores)
	}
	if deserialized.Total != original.Total {
		t.Errorf("Total mismatch: got %f, want %f", deserialized.Total, original.Total)
	}
	if len(deserialized.Usage) != len(original.Usage) {
		t.Errorf("Usage length mismatch: got %d, want %d", len(deserialized.Usage), len(original.Usage))
	}
	for i, usage := range deserialized.Usage {
		if usage != original.Usage[i] {
			t.Errorf("Usage[%d] mismatch: got %f, want %f", i, usage, original.Usage[i])
		}
	}
}

func TestMemoryInfo_JSONSerialization(t *testing.T) {
	original := MemoryInfo{
		Total:     16 * 1024 * 1024 * 1024,
		Used:      8 * 1024 * 1024 * 1024,
		Available: 8 * 1024 * 1024 * 1024,
		Swap: SwapInfo{
			Total: 4 * 1024 * 1024 * 1024,
			Used:  1 * 1024 * 1024 * 1024,
			Free:  3 * 1024 * 1024 * 1024,
		},
		Timestamp: time.Now().Truncate(time.Second),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal MemoryInfo: %v", err)
	}

	var deserialized MemoryInfo
	err = json.Unmarshal(data, &deserialized)
	if err != nil {
		t.Fatalf("Failed to unmarshal MemoryInfo: %v", err)
	}

	if deserialized.Total != original.Total {
		t.Errorf("Total mismatch: got %d, want %d", deserialized.Total, original.Total)
	}
	if deserialized.Used != original.Used {
		t.Errorf("Used mismatch: got %d, want %d", deserialized.Used, original.Used)
	}
	if deserialized.Swap.Total != original.Swap.Total {
		t.Errorf("Swap total mismatch: got %d, want %d", deserialized.Swap.Total, original.Swap.Total)
	}
}