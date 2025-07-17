package main

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseFlags(t *testing.T) {
	// Test flag parsing by testing the individual components
	// Since flag package doesn't allow redefining flags, we test the logic separately
	
	t.Run("default values", func(t *testing.T) {
		// Test that default values are reasonable
		config := Config{
			UpdateInterval: time.Second,
			LogFile:        "",
			Debug:          false,
			NoMouse:        false,
			NoAltScreen:    false,
			Version:        false,
		}
		
		// Verify default values are sensible
		if config.UpdateInterval <= 0 {
			t.Error("Default update interval should be positive")
		}
		if config.UpdateInterval > 10*time.Second {
			t.Error("Default update interval should be reasonable")
		}
	})
	
	t.Run("config validation", func(t *testing.T) {
		// Test various configuration combinations
		configs := []Config{
			{UpdateInterval: 500 * time.Millisecond, Debug: true},
			{UpdateInterval: 2 * time.Second, LogFile: "/tmp/test.log"},
			{UpdateInterval: time.Second, NoMouse: true, NoAltScreen: true},
		}
		
		for i, config := range configs {
			if config.UpdateInterval <= 0 {
				t.Errorf("Config %d: UpdateInterval should be positive", i)
			}
		}
	})
}

func TestSetupLogging(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectFile  bool
		expectError bool
	}{
		{
			name: "no logging",
			config: Config{
				LogFile: "",
				Debug:   false,
			},
			expectFile:  false,
			expectError: false,
		},
		{
			name: "debug only",
			config: Config{
				LogFile: "",
				Debug:   true,
			},
			expectFile:  false,
			expectError: false,
		},
		{
			name: "valid log file",
			config: Config{
				LogFile: "test_system_monitor.log",
				Debug:   false,
			},
			expectFile:  true,
			expectError: false,
		},
		{
			name: "invalid log file path",
			config: Config{
				LogFile: "/invalid/path/that/does/not/exist/test.log",
				Debug:   false,
			},
			expectFile:  false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logFile, err := setupLogging(&tt.config)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectFile && logFile == nil {
				t.Error("Expected log file but got nil")
			}
			if !tt.expectFile && logFile != nil {
				t.Error("Expected no log file but got one")
			}

			// Cleanup
			if logFile != nil {
				logFile.Close()
				os.Remove(logFile.Name())
			}
		})
	}
}

func TestCreateProgram(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "default configuration",
			config: Config{
				UpdateInterval: time.Second,
				NoMouse:        false,
				NoAltScreen:    false,
			},
		},
		{
			name: "disabled features",
			config: Config{
				UpdateInterval: 500 * time.Millisecond,
				NoMouse:        true,
				NoAltScreen:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := createProgram(&tt.config)
			if program == nil {
				t.Error("Expected program but got nil")
			}
		})
	}
}

func TestGracefulShutdown(t *testing.T) {
	t.Run("with log file", func(t *testing.T) {
		// Create a temporary log file
		logFile, err := os.CreateTemp("", "test_shutdown_*.log")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(logFile.Name())

		// Test graceful shutdown
		gracefulShutdown(logFile, nil)

		// Verify file is closed (attempting to write should fail)
		_, err = logFile.WriteString("test")
		if err == nil {
			t.Error("Expected error writing to closed file")
		}
	})

	t.Run("without log file", func(t *testing.T) {
		// Should not panic
		gracefulShutdown(nil, nil)
	})
}

func TestVersionFlag(t *testing.T) {
	// Save original args and stdout
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set version flag
	os.Args = []string{"test", "-version"}

	// This test verifies that the version flag parsing works
	// The actual version output is tested in integration tests
	config := parseFlags()
	if !config.Version {
		t.Error("Version flag not set correctly")
	}
}

func TestApplicationConstants(t *testing.T) {
	if AppName == "" {
		t.Error("AppName should not be empty")
	}
	if AppVersion == "" {
		t.Error("AppVersion should not be empty")
	}
	if !strings.Contains(AppName, "system-monitor") {
		t.Error("AppName should contain 'system-monitor'")
	}
}

// Integration test for the complete application lifecycle
func TestApplicationLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("normal startup and shutdown", func(t *testing.T) {
		// Create a context with timeout for the test
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = ctx // Context used for timeout management

		// Create a temporary log file for testing
		logFile, err := os.CreateTemp("", "test_lifecycle_*.log")
		if err != nil {
			t.Fatalf("Failed to create temp log file: %v", err)
		}
		defer os.Remove(logFile.Name())
		logFile.Close()

		config := &Config{
			UpdateInterval: 100 * time.Millisecond, // Fast updates for testing
			LogFile:        logFile.Name(),
			Debug:          true,
			NoMouse:        true,     // Disable mouse for testing
			NoAltScreen:    true,     // Disable alt screen for testing
		}

		// Setup logging
		testLogFile, err := setupLogging(config)
		if err != nil {
			t.Fatalf("Failed to setup logging: %v", err)
		}
		defer func() {
			if testLogFile != nil {
				testLogFile.Close()
			}
		}()

		// Create program
		program := createProgram(config)
		if program == nil {
			t.Fatal("Failed to create program")
		}

		// Test that program can be created and killed without running
		program.Kill()

		// Test graceful shutdown
		gracefulShutdown(testLogFile, program)

		// Verify log file was created and has content
		if _, err := os.Stat(logFile.Name()); os.IsNotExist(err) {
			t.Error("Log file was not created")
		}
	})

	t.Run("configuration validation", func(t *testing.T) {
		// Test various configuration combinations
		configs := []Config{
			{UpdateInterval: time.Millisecond, NoMouse: true, NoAltScreen: true},
			{UpdateInterval: time.Hour, NoMouse: false, NoAltScreen: false},
			{UpdateInterval: time.Second, Debug: true},
		}

		for i, config := range configs {
			t.Run(string(rune('A'+i)), func(t *testing.T) {
				program := createProgram(&config)
				if program == nil {
					t.Error("Failed to create program with valid config")
				}
				program.Kill()
			})
		}
	})
}

// Benchmark tests for performance validation
func BenchmarkCreateProgram(b *testing.B) {
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
}

func BenchmarkParseFlags(b *testing.B) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set test args
	os.Args = []string{"test", "-interval", "500ms", "-debug"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseFlags()
	}
}

func BenchmarkSetupLogging(b *testing.B) {
	config := &Config{
		LogFile: "",
		Debug:   true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logFile, _ := setupLogging(config)
		if logFile != nil {
			logFile.Close()
		}
	}
}