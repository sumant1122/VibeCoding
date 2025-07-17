package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	tea "github.com/charmbracelet/bubbletea"
	
	"golang-system-monitor-tui/ui"
)

// Config holds application configuration options
type Config struct {
	UpdateInterval time.Duration
	LogFile        string
	Debug          bool
	NoMouse        bool
	NoAltScreen    bool
	Version        bool
}

// Version information
const (
	AppName    = "golang-system-monitor-tui"
	AppVersion = "1.0.0"
)

// parseFlags parses command-line arguments and returns configuration
func parseFlags() *Config {
	config := &Config{}
	
	flag.DurationVar(&config.UpdateInterval, "interval", time.Second, "Update interval for system metrics (e.g., 500ms, 2s)")
	flag.StringVar(&config.LogFile, "log", "", "Log file path (default: no logging)")
	flag.BoolVar(&config.Debug, "debug", false, "Enable debug logging")
	flag.BoolVar(&config.NoMouse, "no-mouse", false, "Disable mouse support")
	flag.BoolVar(&config.NoAltScreen, "no-alt-screen", false, "Disable alternate screen buffer")
	flag.BoolVar(&config.Version, "version", false, "Show version information")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", AppName)
		fmt.Fprintf(os.Stderr, "%s - A terminal-based system resource monitor\n\n", AppName)
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nKeyboard shortcuts:\n")
		fmt.Fprintf(os.Stderr, "  q, Ctrl+C    Quit application\n")
		fmt.Fprintf(os.Stderr, "  arrows, tab  Navigate between components\n")
		fmt.Fprintf(os.Stderr, "  r            Manual refresh\n")
		fmt.Fprintf(os.Stderr, "  ?, h         Toggle help\n")
	}
	
	flag.Parse()
	return config
}

// setupLogging configures logging based on configuration
func setupLogging(config *Config) (*os.File, error) {
	if config.LogFile == "" && !config.Debug {
		// Disable logging by default
		log.SetOutput(os.Stderr)
		return nil, nil
	}
	
	var logFile *os.File
	var err error
	
	if config.LogFile != "" {
		logFile, err = os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		log.SetOutput(logFile)
	}
	
	if config.Debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Debug logging enabled")
	}
	
	return logFile, nil
}

// createProgram creates and configures the Bubble Tea program
func createProgram(config *Config) *tea.Program {
	// Create the main model with configuration
	model := ui.NewMainModelWithConfig(config.UpdateInterval)
	
	// Configure program options based on config
	var options []tea.ProgramOption
	
	if !config.NoAltScreen {
		options = append(options, tea.WithAltScreen())
	}
	
	if !config.NoMouse {
		options = append(options, tea.WithMouseCellMotion())
	}
	
	// Add input handling for better responsiveness
	options = append(options, tea.WithInput(os.Stdin))
	
	return tea.NewProgram(model, options...)
}

// gracefulShutdown handles cleanup operations
func gracefulShutdown(logFile *os.File, program *tea.Program) {
	if logFile != nil {
		log.Println("Application shutting down gracefully")
		logFile.Close()
	}
	
	// Kill the program if it's still running
	if program != nil {
		program.Kill()
	}
}

func main() {
	// Parse command-line arguments
	config := parseFlags()
	
	// Handle version flag
	if config.Version {
		fmt.Printf("%s version %s\n", AppName, AppVersion)
		os.Exit(0)
	}
	
	// Setup logging
	logFile, err := setupLogging(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up logging: %v\n", err)
		os.Exit(1)
	}
	
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = ctx // Context used for potential future enhancements
	
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Create the Bubble Tea program
	program := createProgram(config)
	
	// Channel to receive program result
	resultChan := make(chan error, 1)
	
	// Run the program in a goroutine
	go func() {
		if config.Debug {
			log.Printf("Starting %s with update interval: %v", AppName, config.UpdateInterval)
		}
		
		_, err := program.Run()
		resultChan <- err
	}()
	
	// Wait for either program completion or shutdown signal
	select {
	case err := <-resultChan:
		// Program completed normally
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
			gracefulShutdown(logFile, program)
			os.Exit(1)
		}
		
	case sig := <-sigChan:
		// Received shutdown signal
		if config.Debug {
			log.Printf("Received signal: %v, shutting down gracefully", sig)
		}
		
		// Cancel context and initiate shutdown
		cancel()
		
		// Create a timeout for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		
		// Wait for graceful shutdown or timeout
		done := make(chan struct{})
		go func() {
			gracefulShutdown(logFile, program)
			close(done)
		}()
		
		select {
		case <-done:
			if config.Debug {
				log.Println("Graceful shutdown completed")
			}
		case <-shutdownCtx.Done():
			fmt.Fprintf(os.Stderr, "Shutdown timeout exceeded, forcing exit\n")
		}
	}
	
	// Final cleanup
	gracefulShutdown(logFile, program)
	
	if config.Debug {
		log.Println("Application terminated successfully")
	}
}