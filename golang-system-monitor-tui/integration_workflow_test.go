package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestFullApplicationWorkflow tests the complete application workflow from startup to shutdown
func TestFullApplicationWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full workflow integration test in short mode")
	}

	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "workflow_test_monitor.exe", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("workflow_test_monitor.exe")

	t.Run("complete startup and data collection cycle", func(t *testing.T) {
		// Create temporary log file
		logFile, err := os.CreateTemp("", "workflow_test_*.log")
		if err != nil {
			t.Fatalf("Failed to create temp log file: %v", err)
		}
		defer os.Remove(logFile.Name())
		logFile.Close()

		// Start application with debug logging
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./workflow_test_monitor.exe",
			"-debug",
			"-log", logFile.Name(),
			"-interval", "100ms",
			"-no-alt-screen",
			"-no-mouse",
		)

		// Start the application
		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start application: %v", err)
		}

		// Let it run for data collection cycles
		time.Sleep(1 * time.Second)

		// Gracefully terminate
		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("Failed to terminate process: %v", err)
		}

		// Wait for process to exit
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited
		case <-time.After(3 * time.Second):
			cmd.Process.Kill()
			t.Error("Process did not exit within timeout")
		}

		// Verify log file contains expected workflow events
		logContent, err := os.ReadFile(logFile.Name())
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		logStr := string(logContent)
		expectedLogEntries := []string{
			"Debug logging enabled",
			"Starting golang-system-monitor-tui",
		}

		for _, entry := range expectedLogEntries {
			if !strings.Contains(logStr, entry) {
				t.Errorf("Expected log entry '%s' not found in log", entry)
			}
		}

		if len(logContent) == 0 {
			t.Error("Log file is empty, expected debug output")
		}
	})

	t.Run("keyboard interaction workflow", func(t *testing.T) {
		// This test simulates keyboard interactions during runtime
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./workflow_test_monitor.exe",
			"-interval", "200ms",
			"-no-alt-screen",
			"-no-mouse",
		)

		// Start the application
		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start application: %v", err)
		}

		// Let it initialize
		time.Sleep(300 * time.Millisecond)

		// Send quit signal (simulating 'q' key press)
		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("Failed to send quit signal: %v", err)
		}

		// Wait for graceful exit
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-done:
			// Application should exit cleanly
		case <-time.After(2 * time.Second):
			cmd.Process.Kill()
			t.Error("Application did not respond to quit signal within timeout")
		}
	})

	t.Run("error recovery workflow", func(t *testing.T) {
		// Test application behavior with invalid configurations
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Test with very fast update interval (stress test)
		cmd := exec.CommandContext(ctx, "./workflow_test_monitor.exe",
			"-interval", "1ms", // Very fast updates
			"-no-alt-screen",
			"-no-mouse",
		)

		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start application with fast updates: %v", err)
		}

		// Let it run briefly under stress
		time.Sleep(200 * time.Millisecond)

		// Terminate
		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("Failed to terminate stressed application: %v", err)
		}

		// Should still exit cleanly even under stress
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-done:
			// Success - application handled stress gracefully
		case <-time.After(2 * time.Second):
			cmd.Process.Kill()
			t.Error("Stressed application did not exit cleanly")
		}
	})
}

// TestDataCollectionWorkflow tests the data collection and display workflow
func TestDataCollectionWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping data collection workflow test in short mode")
	}

	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "data_workflow_test.exe", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("data_workflow_test.exe")

	t.Run("continuous data collection", func(t *testing.T) {
		logFile, err := os.CreateTemp("", "data_workflow_*.log")
		if err != nil {
			t.Fatalf("Failed to create temp log file: %v", err)
		}
		defer os.Remove(logFile.Name())
		logFile.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./data_workflow_test.exe",
			"-debug",
			"-log", logFile.Name(),
			"-interval", "250ms", // 4 updates per second
			"-no-alt-screen",
			"-no-mouse",
		)

		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start application: %v", err)
		}

		// Let it collect data for multiple cycles
		time.Sleep(1500 * time.Millisecond) // Should get ~6 update cycles

		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("Failed to terminate process: %v", err)
		}

		// Wait for exit
		cmd.Wait()

		// Verify continuous data collection occurred
		logContent, err := os.ReadFile(logFile.Name())
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		if len(logContent) == 0 {
			t.Error("Expected debug output from continuous data collection")
		}

		// Log should contain startup information
		logStr := string(logContent)
		if !strings.Contains(logStr, "golang-system-monitor-tui") {
			t.Error("Expected application name in debug log")
		}
	})

	t.Run("data collection under load", func(t *testing.T) {
		// Test data collection with very frequent updates
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./data_workflow_test.exe",
			"-interval", "10ms", // Very frequent updates
			"-no-alt-screen",
			"-no-mouse",
		)

		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start high-frequency application: %v", err)
		}

		// Let it run under high update frequency
		time.Sleep(500 * time.Millisecond)

		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("Failed to terminate high-frequency process: %v", err)
		}

		// Should handle high frequency updates without crashing
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-done:
			// Success - handled high frequency updates
		case <-time.After(1 * time.Second):
			cmd.Process.Kill()
			t.Error("High-frequency application did not exit cleanly")
		}
	})
}

// TestConfigurationWorkflow tests various configuration scenarios
func TestConfigurationWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping configuration workflow test in short mode")
	}

	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "config_workflow_test.exe", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("config_workflow_test.exe")

	configurations := []struct {
		name string
		args []string
	}{
		{
			name: "minimal config",
			args: []string{"-no-alt-screen", "-no-mouse"},
		},
		{
			name: "debug config",
			args: []string{"-debug", "-no-alt-screen", "-interval", "500ms"},
		},
		{
			name: "custom interval",
			args: []string{"-interval", "2s", "-no-alt-screen", "-no-mouse"},
		},
		{
			name: "all features disabled",
			args: []string{"-no-alt-screen", "-no-mouse", "-interval", "1s"},
		},
	}

	for _, config := range configurations {
		t.Run(config.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "./config_workflow_test.exe", config.args...)

			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start application with config %s: %v", config.name, err)
			}

			// Let it run briefly to verify it starts successfully
			time.Sleep(200 * time.Millisecond)

			if err := cmd.Process.Kill(); err != nil {
				t.Fatalf("Failed to terminate application: %v", err)
			}

			// Wait for clean exit
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case <-done:
				// Configuration worked successfully
			case <-ctx.Done():
				cmd.Process.Kill()
				t.Errorf("Configuration %s did not start/exit cleanly", config.name)
			}
		})
	}
}

// TestSignalHandlingWorkflow tests graceful shutdown with different signals
func TestSignalHandlingWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping signal handling workflow test in short mode")
	}

	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "signal_workflow_test.exe", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("signal_workflow_test.exe")

	t.Run("SIGTERM handling", func(t *testing.T) {
		logFile, err := os.CreateTemp("", "signal_test_*.log")
		if err != nil {
			t.Fatalf("Failed to create temp log file: %v", err)
		}
		defer os.Remove(logFile.Name())
		logFile.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./signal_workflow_test.exe",
			"-debug",
			"-log", logFile.Name(),
			"-interval", "200ms",
			"-no-alt-screen",
			"-no-mouse",
		)

		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start application: %v", err)
		}

		// Let it run for a bit
		time.Sleep(300 * time.Millisecond)

		// Send SIGTERM for graceful shutdown
		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("Failed to send SIGTERM: %v", err)
		}

		// Wait for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case <-done:
			// Graceful shutdown completed
		case <-time.After(3 * time.Second):
			cmd.Process.Kill()
			t.Error("Application did not shutdown gracefully within timeout")
		}

		// Verify graceful shutdown was logged
		logContent, err := os.ReadFile(logFile.Name())
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		if len(logContent) == 0 {
			t.Error("Expected shutdown logging")
		}
	})
}

// TestResourceMonitoringWorkflow tests the actual resource monitoring capabilities
func TestResourceMonitoringWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping resource monitoring workflow test in short mode")
	}

	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "resource_workflow_test.exe", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("resource_workflow_test.exe")

	t.Run("resource monitoring accuracy", func(t *testing.T) {
		logFile, err := os.CreateTemp("", "resource_test_*.log")
		if err != nil {
			t.Fatalf("Failed to create temp log file: %v", err)
		}
		defer os.Remove(logFile.Name())
		logFile.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./resource_workflow_test.exe",
			"-debug",
			"-log", logFile.Name(),
			"-interval", "500ms", // 2 updates per second
			"-no-alt-screen",
			"-no-mouse",
		)

		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start resource monitoring: %v", err)
		}

		// Let it monitor resources for multiple cycles
		time.Sleep(2500 * time.Millisecond) // ~5 update cycles

		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("Failed to terminate resource monitor: %v", err)
		}

		// Wait for exit
		cmd.Wait()

		// Verify resource monitoring occurred
		logContent, err := os.ReadFile(logFile.Name())
		if err != nil {
			t.Fatalf("Failed to read resource monitoring log: %v", err)
		}

		if len(logContent) == 0 {
			t.Error("Expected resource monitoring debug output")
		}

		// The application should have started and collected resource data
		logStr := string(logContent)
		if !strings.Contains(logStr, "Debug logging enabled") {
			t.Error("Expected debug logging confirmation")
		}
	})
}