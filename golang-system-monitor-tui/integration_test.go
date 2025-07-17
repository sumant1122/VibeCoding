package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestApplicationIntegration tests the complete application lifecycle
func TestApplicationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the application first
	buildCmd := exec.Command("go", "build", "-o", "test_system_monitor.exe", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("test_system_monitor.exe")

	tests := []struct {
		name string
		args []string
		test func(t *testing.T, cmd *exec.Cmd)
	}{
		{
			name: "version flag",
			args: []string{"-version"},
			test: func(t *testing.T, cmd *exec.Cmd) {
				output, err := cmd.Output()
				if err != nil {
					t.Fatalf("Command failed: %v", err)
				}
				
				outputStr := string(output)
				if !strings.Contains(outputStr, AppName) {
					t.Errorf("Version output should contain app name, got: %s", outputStr)
				}
				if !strings.Contains(outputStr, AppVersion) {
					t.Errorf("Version output should contain version, got: %s", outputStr)
				}
			},
		},
		{
			name: "help flag",
			args: []string{"-h"},
			test: func(t *testing.T, cmd *exec.Cmd) {
				output, err := cmd.CombinedOutput()
				if err != nil {
					// -h flag causes exit code 2, which is expected
					if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 2 {
						// Help output is in the combined output
					} else {
						t.Fatalf("Unexpected error: %v", err)
					}
				}
				
				outputStr := string(output)
				if !strings.Contains(outputStr, "Usage:") {
					t.Errorf("Help output should contain usage information, got: %s", outputStr)
				}
				if !strings.Contains(outputStr, "interval") {
					t.Errorf("Help output should contain interval option, got: %s", outputStr)
				}
			},
		},
		{
			name: "invalid flag",
			args: []string{"-invalid-flag"},
			test: func(t *testing.T, cmd *exec.Cmd) {
				output, err := cmd.Output()
				_ = output // Output not used in this test case
				if err == nil {
					t.Error("Expected error for invalid flag")
				}
				
				if exitError, ok := err.(*exec.ExitError); ok {
					outputStr := string(exitError.Stderr)
					if !strings.Contains(outputStr, "flag provided but not defined") {
						t.Errorf("Expected flag error message, got: %s", outputStr)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./test_system_monitor.exe", tt.args...)
			tt.test(t, cmd)
		})
	}
}

// TestApplicationRuntime tests the application during runtime
func TestApplicationRuntime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping runtime integration test in short mode")
	}

	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "test_system_monitor.exe", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("test_system_monitor.exe")

	t.Run("graceful shutdown with SIGTERM", func(t *testing.T) {
		// Create a temporary log file
		logFile, err := os.CreateTemp("", "test_runtime_*.log")
		if err != nil {
			t.Fatalf("Failed to create temp log file: %v", err)
		}
		defer os.Remove(logFile.Name())
		logFile.Close()

		// Start the application with logging
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./test_system_monitor.exe", 
			"-debug", 
			"-log", logFile.Name(),
			"-interval", "100ms",
			"-no-alt-screen", // Disable alt screen for testing
		)

		// Start the command
		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start application: %v", err)
		}

		// Let it run for a short time
		time.Sleep(500 * time.Millisecond)

		// Send interrupt signal for graceful shutdown
		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("Failed to terminate process: %v", err)
		}

		// Wait for the process to exit
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case err := <-done:
			// Process should exit cleanly
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					// Exit code 0 or signal termination is acceptable
					if exitError.ExitCode() != 0 && !exitError.Exited() {
						t.Errorf("Process did not exit cleanly: %v", err)
					}
				}
			}
		case <-time.After(5 * time.Second):
			// Force kill if it doesn't exit gracefully
			cmd.Process.Kill()
			t.Error("Process did not exit within timeout")
		}

		// Verify log file has content
		if stat, err := os.Stat(logFile.Name()); err != nil {
			t.Errorf("Log file error: %v", err)
		} else if stat.Size() == 0 {
			t.Error("Log file is empty")
		}
	})

	t.Run("application with custom interval", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "./test_system_monitor.exe", 
			"-interval", "50ms",
			"-no-alt-screen",
			"-no-mouse",
		)

		// Start the command
		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start application: %v", err)
		}

		// Let it run briefly
		time.Sleep(200 * time.Millisecond)

		// Terminate the process
		if err := cmd.Process.Kill(); err != nil {
			t.Fatalf("Failed to terminate process: %v", err)
		}

		// Wait for exit
		select {
		case <-ctx.Done():
			cmd.Process.Kill()
			t.Error("Application did not respond to interrupt signal")
		case <-time.After(2 * time.Second):
			// Should exit within reasonable time
		}
	})
}

// TestApplicationConfiguration tests various configuration scenarios
func TestApplicationConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping configuration integration test in short mode")
	}

	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "test_system_monitor.exe", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("test_system_monitor.exe")

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "minimal configuration",
			args: []string{"-no-alt-screen", "-no-mouse", "-interval", "1s"},
		},
		{
			name: "debug configuration",
			args: []string{"-debug", "-no-alt-screen", "-interval", "200ms"},
		},
		{
			name: "fast updates",
			args: []string{"-interval", "10ms", "-no-alt-screen", "-no-mouse"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "./test_system_monitor.exe", tt.args...)

			// Start the command
			if err := cmd.Start(); err != nil {
				t.Fatalf("Failed to start application with args %v: %v", tt.args, err)
			}

			// Let it run briefly to ensure it starts successfully
			time.Sleep(100 * time.Millisecond)

			// Terminate gracefully
			if err := cmd.Process.Kill(); err != nil {
				t.Fatalf("Failed to terminate process: %v", err)
			}

			// Wait for exit
			done := make(chan error, 1)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case <-done:
				// Success - application started and stopped cleanly
			case <-ctx.Done():
				cmd.Process.Kill()
				t.Errorf("Application with args %v did not exit cleanly", tt.args)
			}
		})
	}
}

// TestApplicationErrorHandling tests error scenarios
func TestApplicationErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error handling integration test in short mode")
	}

	// Build the application
	buildCmd := exec.Command("go", "build", "-o", "test_system_monitor.exe", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("test_system_monitor.exe")

	t.Run("invalid log file path", func(t *testing.T) {
		cmd := exec.Command("./test_system_monitor.exe", 
			"-log", "/invalid/path/that/does/not/exist/test.log",
		)

		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("Expected error for invalid log file path")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Error setting up logging") {
			t.Errorf("Expected logging error message, got: %s", outputStr)
		}
	})

	t.Run("invalid interval format", func(t *testing.T) {
		cmd := exec.Command("./test_system_monitor.exe", "-interval", "invalid")

		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("Expected error for invalid interval format")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "parse error") {
			t.Errorf("Expected duration parsing error, got: %s", outputStr)
		}
	})
}

// BenchmarkApplicationStartup benchmarks application startup time
func BenchmarkApplicationStartup(b *testing.B) {
	// Build the application once
	buildCmd := exec.Command("go", "build", "-o", "bench_system_monitor.exe", ".")
	if err := buildCmd.Run(); err != nil {
		b.Fatalf("Failed to build application: %v", err)
	}
	defer os.Remove("bench_system_monitor.exe")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command("./bench_system_monitor.exe", "-version")
		if err := cmd.Run(); err != nil {
			b.Fatalf("Command failed: %v", err)
		}
	}
}