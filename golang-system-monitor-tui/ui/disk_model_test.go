package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"golang-system-monitor-tui/models"
)

func TestNewDiskModel(t *testing.T) {
	model := NewDiskModel()
	
	if len(model.filesystems) != 0 {
		t.Errorf("Expected empty filesystems, got %d", len(model.filesystems))
	}
	
	if model.width != 50 {
		t.Errorf("Expected width 50, got %d", model.width)
	}
	
	if model.height != 10 {
		t.Errorf("Expected height 10, got %d", model.height)
	}
}

func TestDiskModel_Init(t *testing.T) {
	model := NewDiskModel()
	cmd := model.Init()
	
	if cmd != nil {
		t.Error("Expected Init to return nil command")
	}
}

func TestDiskModel_Update(t *testing.T) {
	model := NewDiskModel()
	
	// Test with DiskUpdateMsg
	diskInfo := []models.DiskInfo{
		{
			Device:      "/dev/sda1",
			Mountpoint:  "/",
			Filesystem:  "ext4",
			Total:       1000000000, // 1GB
			Used:        500000000,  // 500MB
			Available:   500000000,  // 500MB
			UsedPercent: 50.0,
		},
		{
			Device:      "/dev/sda2",
			Mountpoint:  "/home",
			Filesystem:  "ext4",
			Total:       2000000000, // 2GB
			Used:        1800000000, // 1.8GB
			Available:   200000000,  // 200MB
			UsedPercent: 90.0,
		},
	}
	
	updateMsg := DiskUpdateMsg(diskInfo)
	updatedModel, cmd := model.Update(updateMsg)
	
	if cmd != nil {
		t.Error("Expected Update to return nil command")
	}
	
	if len(updatedModel.filesystems) != 2 {
		t.Errorf("Expected 2 filesystems, got %d", len(updatedModel.filesystems))
	}
	
	if updatedModel.filesystems[0].Device != "/dev/sda1" {
		t.Errorf("Expected device /dev/sda1, got %s", updatedModel.filesystems[0].Device)
	}
	
	if updatedModel.filesystems[1].UsedPercent != 90.0 {
		t.Errorf("Expected usage 90.0%%, got %.1f%%", updatedModel.filesystems[1].UsedPercent)
	}
	
	// Test with unrelated message
	unrelatedModel, cmd := model.Update(tea.KeyMsg{})
	if cmd != nil {
		t.Error("Expected Update to return nil command for unrelated message")
	}
	
	if len(unrelatedModel.filesystems) != 0 {
		t.Error("Expected filesystems to remain unchanged for unrelated message")
	}
}

func TestDiskModel_View_EmptyData(t *testing.T) {
	model := NewDiskModel()
	view := model.View()
	
	if !strings.Contains(view, "Disk Usage") {
		t.Error("Expected view to contain 'Disk Usage' header")
	}
	
	if !strings.Contains(view, "Loading disk data...") {
		t.Error("Expected view to contain loading placeholder")
	}
}

func TestDiskModel_View_WithData(t *testing.T) {
	model := NewDiskModel()
	
	// Add test data
	diskInfo := []models.DiskInfo{
		{
			Device:      "/dev/sda1",
			Mountpoint:  "/",
			Filesystem:  "ext4",
			Total:       1000000000, // 1GB
			Used:        500000000,  // 500MB
			Available:   500000000,  // 500MB
			UsedPercent: 50.0,
		},
	}
	
	updateMsg := DiskUpdateMsg(diskInfo)
	model, _ = model.Update(updateMsg)
	
	view := model.View()
	
	if !strings.Contains(view, "Disk Usage") {
		t.Error("Expected view to contain 'Disk Usage' header")
	}
	
	if !strings.Contains(view, "/") {
		t.Error("Expected view to contain mountpoint '/'")
	}
	
	if !strings.Contains(view, "50.0%") {
		t.Error("Expected view to contain usage percentage '50.0%'")
	}
	
	if !strings.Contains(view, "476.8MB") {
		t.Error("Expected view to contain formatted used space")
	}
	
	if !strings.Contains(view, "953.7MB") {
		t.Error("Expected view to contain formatted total space")
	}
}

func TestDiskModel_StyleManagerIntegration(t *testing.T) {
	model := NewDiskModel()
	
	// Test that style manager is initialized
	if model.styleManager == nil {
		t.Error("Expected style manager to be initialized")
	}
	
	tests := []struct {
		name       string
		percentage float64
		expectRed  bool
	}{
		{"Normal usage", 50.0, false},
		{"Warning usage", 75.0, false},
		{"Critical usage", 90.0, true},
		{"Over critical", 95.0, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := model.styleManager.RenderProgressBar(tt.percentage, 20, false)
			
			// Check that the bar contains progress characters
			if !strings.Contains(bar, "█") && !strings.Contains(bar, "░") {
				t.Error("Expected progress bar to contain progress characters")
			}
			
			// Test color selection through style manager
			color := model.styleManager.GetUsageColor(tt.percentage)
			criticalColor := model.styleManager.GetUsageColor(95.0) // Known critical percentage
			if tt.expectRed && color != criticalColor {
				t.Errorf("Expected critical color for %.1f%% usage", tt.percentage)
			}
		})
	}
}

func TestDiskModel_FormatBytes(t *testing.T) {
	model := NewDiskModel()
	
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0B"},
		{512, "512B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1048576, "1.0MB"},
		{1073741824, "1.0GB"},
		{1099511627776, "1.0TB"},
		{1536000000000, "1.4TB"},
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := model.formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDiskModel_SetSize(t *testing.T) {
	model := NewDiskModel()
	model = model.SetSize(80, 20)
	
	if model.width != 80 {
		t.Errorf("Expected width 80, got %d", model.width)
	}
	
	if model.height != 20 {
		t.Errorf("Expected height 20, got %d", model.height)
	}
}

func TestDiskModel_GetFilesystems(t *testing.T) {
	model := NewDiskModel()
	
	diskInfo := []models.DiskInfo{
		{
			Device:      "/dev/sda1",
			Mountpoint:  "/",
			UsedPercent: 50.0,
		},
	}
	
	updateMsg := DiskUpdateMsg(diskInfo)
	model, _ = model.Update(updateMsg)
	
	filesystems := model.GetFilesystems()
	
	if len(filesystems) != 1 {
		t.Errorf("Expected 1 filesystem, got %d", len(filesystems))
	}
	
	if filesystems[0].Device != "/dev/sda1" {
		t.Errorf("Expected device /dev/sda1, got %s", filesystems[0].Device)
	}
}

func TestDiskModel_GetHighUsageFilesystems(t *testing.T) {
	model := NewDiskModel()
	
	diskInfo := []models.DiskInfo{
		{
			Device:      "/dev/sda1",
			Mountpoint:  "/",
			UsedPercent: 50.0,
		},
		{
			Device:      "/dev/sda2",
			Mountpoint:  "/home",
			UsedPercent: 85.0,
		},
		{
			Device:      "/dev/sda3",
			Mountpoint:  "/var",
			UsedPercent: 95.0,
		},
	}
	
	updateMsg := DiskUpdateMsg(diskInfo)
	model, _ = model.Update(updateMsg)
	
	// Test threshold of 80%
	highUsage := model.GetHighUsageFilesystems(80.0)
	
	if len(highUsage) != 2 {
		t.Errorf("Expected 2 high usage filesystems, got %d", len(highUsage))
	}
	
	// Verify the correct filesystems are returned
	found85 := false
	found95 := false
	for _, fs := range highUsage {
		if fs.UsedPercent == 85.0 {
			found85 = true
		}
		if fs.UsedPercent == 95.0 {
			found95 = true
		}
	}
	
	if !found85 || !found95 {
		t.Error("Expected to find filesystems with 85% and 95% usage")
	}
}

func TestDiskModel_GetCriticalFilesystems(t *testing.T) {
	model := NewDiskModel()
	
	diskInfo := []models.DiskInfo{
		{
			Device:      "/dev/sda1",
			Mountpoint:  "/",
			UsedPercent: 50.0,
		},
		{
			Device:      "/dev/sda2",
			Mountpoint:  "/home",
			UsedPercent: 85.0,
		},
		{
			Device:      "/dev/sda3",
			Mountpoint:  "/var",
			UsedPercent: 95.0,
		},
	}
	
	updateMsg := DiskUpdateMsg(diskInfo)
	model, _ = model.Update(updateMsg)
	
	critical := model.GetCriticalFilesystems()
	
	if len(critical) != 1 {
		t.Errorf("Expected 1 critical filesystem, got %d", len(critical))
	}
	
	if critical[0].UsedPercent != 95.0 {
		t.Errorf("Expected critical filesystem with 95%% usage, got %.1f%%", critical[0].UsedPercent)
	}
}

func TestDiskModel_HasCriticalUsage(t *testing.T) {
	model := NewDiskModel()
	
	// Test without critical usage
	diskInfo := []models.DiskInfo{
		{
			Device:      "/dev/sda1",
			Mountpoint:  "/",
			UsedPercent: 50.0,
		},
		{
			Device:      "/dev/sda2",
			Mountpoint:  "/home",
			UsedPercent: 85.0,
		},
	}
	
	updateMsg := DiskUpdateMsg(diskInfo)
	model, _ = model.Update(updateMsg)
	
	if model.HasCriticalUsage() {
		t.Error("Expected no critical usage for filesystems under 90%")
	}
	
	// Test with critical usage
	diskInfo[1].UsedPercent = 95.0
	updateMsg = DiskUpdateMsg(diskInfo)
	model, _ = model.Update(updateMsg)
	
	if !model.HasCriticalUsage() {
		t.Error("Expected critical usage for filesystem with 95% usage")
	}
}

func TestDiskModel_TotalSpaceCalculations(t *testing.T) {
	model := NewDiskModel()
	
	diskInfo := []models.DiskInfo{
		{
			Device:      "/dev/sda1",
			Mountpoint:  "/",
			Total:       1000000000, // 1GB
			Used:        500000000,  // 500MB
			UsedPercent: 50.0,
		},
		{
			Device:      "/dev/sda2",
			Mountpoint:  "/home",
			Total:       2000000000, // 2GB
			Used:        1000000000, // 1GB
			UsedPercent: 50.0,
		},
	}
	
	updateMsg := DiskUpdateMsg(diskInfo)
	model, _ = model.Update(updateMsg)
	
	totalSpace := model.GetTotalDiskSpace()
	expectedTotal := uint64(3000000000) // 3GB
	if totalSpace != expectedTotal {
		t.Errorf("Expected total space %d, got %d", expectedTotal, totalSpace)
	}
	
	totalUsed := model.GetTotalUsedSpace()
	expectedUsed := uint64(1500000000) // 1.5GB
	if totalUsed != expectedUsed {
		t.Errorf("Expected total used %d, got %d", expectedUsed, totalUsed)
	}
	
	overallPercent := model.GetOverallUsagePercent()
	expectedPercent := 50.0 // 1.5GB / 3GB = 50%
	if overallPercent != expectedPercent {
		t.Errorf("Expected overall usage %.1f%%, got %.1f%%", expectedPercent, overallPercent)
	}
}

func TestDiskModel_TotalSpaceCalculations_EmptyData(t *testing.T) {
	model := NewDiskModel()
	
	totalSpace := model.GetTotalDiskSpace()
	if totalSpace != 0 {
		t.Errorf("Expected total space 0 for empty data, got %d", totalSpace)
	}
	
	totalUsed := model.GetTotalUsedSpace()
	if totalUsed != 0 {
		t.Errorf("Expected total used 0 for empty data, got %d", totalUsed)
	}
	
	overallPercent := model.GetOverallUsagePercent()
	if overallPercent != 0 {
		t.Errorf("Expected overall usage 0%% for empty data, got %.1f%%", overallPercent)
	}
}

func TestDiskModel_LongMountpointTruncation(t *testing.T) {
	model := NewDiskModel()
	
	diskInfo := []models.DiskInfo{
		{
			Device:      "/dev/sda1",
			Mountpoint:  "/very/long/mountpoint/path/that/should/be/truncated",
			Filesystem:  "ext4",
			Total:       1000000000,
			Used:        500000000,
			UsedPercent: 50.0,
		},
	}
	
	updateMsg := DiskUpdateMsg(diskInfo)
	model, _ = model.Update(updateMsg)
	
	view := model.View()
	
	// Check that the long mountpoint is truncated with "..."
	if !strings.Contains(view, "...") {
		t.Error("Expected long mountpoint to be truncated with '...'")
	}
	
	// Check that the original long path is not displayed in full
	if strings.Contains(view, "/very/long/mountpoint/path/that/should/be/truncated") {
		t.Error("Expected long mountpoint to be truncated, but full path was displayed")
	}
}

// Benchmark tests for performance validation
func BenchmarkDiskModel_Update(b *testing.B) {
	model := NewDiskModel()
	
	diskInfo := []models.DiskInfo{
		{
			Device:      "/dev/sda1",
			Mountpoint:  "/",
			Total:       1000000000,
			Used:        500000000,
			UsedPercent: 50.0,
		},
	}
	
	updateMsg := DiskUpdateMsg(diskInfo)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.Update(updateMsg)
	}
}

func BenchmarkDiskModel_View(b *testing.B) {
	model := NewDiskModel()
	
	diskInfo := []models.DiskInfo{
		{
			Device:      "/dev/sda1",
			Mountpoint:  "/",
			Total:       1000000000,
			Used:        500000000,
			UsedPercent: 50.0,
		},
		{
			Device:      "/dev/sda2",
			Mountpoint:  "/home",
			Total:       2000000000,
			Used:        1800000000,
			UsedPercent: 90.0,
		},
	}
	
	updateMsg := DiskUpdateMsg(diskInfo)
	model, _ = model.Update(updateMsg)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.View()
	}
}