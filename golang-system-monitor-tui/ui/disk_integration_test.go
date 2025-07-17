package ui

import (
	"testing"

	"golang-system-monitor-tui/services"
)

// TestDiskModel_Integration tests the integration between DiskModel and GopsutilCollector
func TestDiskModel_Integration(t *testing.T) {
	// Create collector and disk model
	collector := services.NewGopsutilCollector()
	model := NewDiskModel()
	
	// Collect real disk data
	diskInfo, err := collector.CollectDisk()
	if err != nil {
		t.Skipf("Skipping integration test due to disk collection error: %v", err)
	}
	
	// Update model with real data
	updateMsg := DiskUpdateMsg(diskInfo)
	updatedModel, cmd := model.Update(updateMsg)
	
	if cmd != nil {
		t.Error("Expected Update to return nil command")
	}
	
	// Verify model was updated
	filesystems := updatedModel.GetFilesystems()
	if len(filesystems) == 0 {
		t.Skip("No filesystems found, skipping integration test")
	}
	
	// Verify view renders without errors
	view := updatedModel.View()
	if len(view) == 0 {
		t.Error("Expected non-empty view output")
	}
	
	// Verify basic content is present
	if !containsAny(view, []string{"Disk Usage", "/"}) {
		t.Error("Expected view to contain disk usage information")
	}
	
	// Test warning threshold logic with real data
	criticalFilesystems := updatedModel.GetCriticalFilesystems()
	hasCritical := updatedModel.HasCriticalUsage()
	
	if len(criticalFilesystems) > 0 && !hasCritical {
		t.Error("HasCriticalUsage should return true when critical filesystems exist")
	}
	
	if len(criticalFilesystems) == 0 && hasCritical {
		t.Error("HasCriticalUsage should return false when no critical filesystems exist")
	}
}

// Helper function to check if string contains any of the provided substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}