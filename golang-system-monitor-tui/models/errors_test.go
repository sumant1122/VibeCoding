package models

import (
	"errors"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSystemError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      SystemError
		expected string
	}{
		{
			name: "SystemAccessError",
			err: SystemError{
				Type:      SystemAccessError,
				Message:   "test message",
				Component: "CPU",
				Timestamp: time.Now(),
				Original:  errors.New("original error"),
			},
			expected: "[CPU] System Access Error: test message",
		},
		{
			name: "PermissionError",
			err: SystemError{
				Type:      PermissionError,
				Message:   "access denied",
				Component: "Memory",
				Timestamp: time.Now(),
				Original:  errors.New("permission denied"),
			},
			expected: "[Memory] Permission Error: access denied",
		},
		{
			name: "DataCollectionError",
			err: SystemError{
				Type:      DataCollectionError,
				Message:   "collection failed",
				Component: "Disk",
				Timestamp: time.Now(),
				Original:  errors.New("collection error"),
			},
			expected: "[Disk] Data Collection Error: collection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("SystemError.Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSystemError_IsRecoverable(t *testing.T) {
	tests := []struct {
		name       string
		errorType  ErrorType
		recoverable bool
	}{
		{"TemporaryError", TemporaryError, true},
		{"DataCollectionError", DataCollectionError, true},
		{"SystemAccessError", SystemAccessError, false},
		{"PermissionError", PermissionError, false},
		{"RenderError", RenderError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SystemError{Type: tt.errorType}
			if got := err.IsRecoverable(); got != tt.recoverable {
				t.Errorf("SystemError.IsRecoverable() = %v, want %v", got, tt.recoverable)
			}
		})
	}
}

func TestNewErrorHandler(t *testing.T) {
	logger := log.New(os.Stdout, "test: ", log.LstdFlags)
	handler := NewErrorHandler(logger)

	if handler == nil {
		t.Fatal("NewErrorHandler() returned nil")
	}

	if handler.logger != logger {
		t.Error("NewErrorHandler() did not set logger correctly")
	}
}

func TestErrorHandler_HandleSystemError(t *testing.T) {
	var logOutput strings.Builder
	logger := log.New(&logOutput, "", 0)
	handler := NewErrorHandler(logger)

	originalErr := errors.New("test system error")
	cmd := handler.HandleSystemError("TestComponent", originalErr)

	if cmd == nil {
		t.Fatal("HandleSystemError() returned nil command")
	}

	// Execute the command to get the message
	msg := cmd()
	errorMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("Expected ErrorMsg, got %T", msg)
	}

	if errorMsg.Type != SystemAccessError {
		t.Errorf("Expected SystemAccessError, got %v", errorMsg.Type)
	}

	if errorMsg.Component != "TestComponent" {
		t.Errorf("Expected component 'TestComponent', got %v", errorMsg.Component)
	}

	if errorMsg.Original != originalErr {
		t.Errorf("Expected original error to be preserved")
	}

	// Check that error was logged
	logContent := logOutput.String()
	if !strings.Contains(logContent, "System error in TestComponent") {
		t.Errorf("Expected log message not found in: %s", logContent)
	}
}

func TestErrorHandler_HandleDataError(t *testing.T) {
	var logOutput strings.Builder
	logger := log.New(&logOutput, "", 0)
	handler := NewErrorHandler(logger)

	originalErr := errors.New("test data error")
	cmd := handler.HandleDataError("DataComponent", originalErr)

	if cmd == nil {
		t.Fatal("HandleDataError() returned nil command")
	}

	// Execute the command to get the message
	msg := cmd()
	errorMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("Expected ErrorMsg, got %T", msg)
	}

	if errorMsg.Type != DataCollectionError {
		t.Errorf("Expected DataCollectionError, got %v", errorMsg.Type)
	}

	if errorMsg.Component != "DataComponent" {
		t.Errorf("Expected component 'DataComponent', got %v", errorMsg.Component)
	}

	// Check that error was logged
	logContent := logOutput.String()
	if !strings.Contains(logContent, "Data collection error in DataComponent") {
		t.Errorf("Expected log message not found in: %s", logContent)
	}
}

func TestErrorHandler_HandlePermissionError(t *testing.T) {
	var logOutput strings.Builder
	logger := log.New(&logOutput, "", 0)
	handler := NewErrorHandler(logger)

	originalErr := errors.New("permission denied")
	cmd := handler.HandlePermissionError("PermComponent", originalErr)

	if cmd == nil {
		t.Fatal("HandlePermissionError() returned nil command")
	}

	// Execute the command to get the message
	msg := cmd()
	errorMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("Expected ErrorMsg, got %T", msg)
	}

	if errorMsg.Type != PermissionError {
		t.Errorf("Expected PermissionError, got %v", errorMsg.Type)
	}

	if errorMsg.Component != "PermComponent" {
		t.Errorf("Expected component 'PermComponent', got %v", errorMsg.Component)
	}

	// Check that error was logged
	logContent := logOutput.String()
	if !strings.Contains(logContent, "Permission error in PermComponent") {
		t.Errorf("Expected log message not found in: %s", logContent)
	}
}

func TestErrorHandler_HandleTemporaryError(t *testing.T) {
	var logOutput strings.Builder
	logger := log.New(&logOutput, "", 0)
	handler := NewErrorHandler(logger)

	originalErr := errors.New("temporary failure")
	cmd := handler.HandleTemporaryError("TempComponent", originalErr)

	if cmd == nil {
		t.Fatal("HandleTemporaryError() returned nil command")
	}

	// Execute the command to get the message
	msg := cmd()
	errorMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("Expected ErrorMsg, got %T", msg)
	}

	if errorMsg.Type != TemporaryError {
		t.Errorf("Expected TemporaryError, got %v", errorMsg.Type)
	}

	if errorMsg.Component != "TempComponent" {
		t.Errorf("Expected component 'TempComponent', got %v", errorMsg.Component)
	}

	// Check that error was logged
	logContent := logOutput.String()
	if !strings.Contains(logContent, "Temporary error in TempComponent") {
		t.Errorf("Expected log message not found in: %s", logContent)
	}
}

func TestErrorHandler_HandleRenderError(t *testing.T) {
	var logOutput strings.Builder
	logger := log.New(&logOutput, "", 0)
	handler := NewErrorHandler(logger)

	originalErr := errors.New("render failure")
	cmd := handler.HandleRenderError("RenderComponent", originalErr)

	if cmd == nil {
		t.Fatal("HandleRenderError() returned nil command")
	}

	// Execute the command to get the message
	msg := cmd()
	errorMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("Expected ErrorMsg, got %T", msg)
	}

	if errorMsg.Type != RenderError {
		t.Errorf("Expected RenderError, got %v", errorMsg.Type)
	}

	if errorMsg.Component != "RenderComponent" {
		t.Errorf("Expected component 'RenderComponent', got %v", errorMsg.Component)
	}

	// Check that error was logged
	logContent := logOutput.String()
	if !strings.Contains(logContent, "Render error in RenderComponent") {
		t.Errorf("Expected log message not found in: %s", logContent)
	}
}

func TestCreateSystemError(t *testing.T) {
	originalErr := errors.New("original error")
	systemErr := CreateSystemError(DataCollectionError, "TestComponent", "test message", originalErr)

	if systemErr.Type != DataCollectionError {
		t.Errorf("Expected DataCollectionError, got %v", systemErr.Type)
	}

	if systemErr.Component != "TestComponent" {
		t.Errorf("Expected component 'TestComponent', got %v", systemErr.Component)
	}

	if systemErr.Message != "test message" {
		t.Errorf("Expected message 'test message', got %v", systemErr.Message)
	}

	if systemErr.Original != originalErr {
		t.Errorf("Expected original error to be preserved")
	}

	if systemErr.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	systemErr := WrapError(originalErr, "WrapComponent", PermissionError)

	if systemErr.Type != PermissionError {
		t.Errorf("Expected PermissionError, got %v", systemErr.Type)
	}

	if systemErr.Component != "WrapComponent" {
		t.Errorf("Expected component 'WrapComponent', got %v", systemErr.Component)
	}

	if systemErr.Message != "original error" {
		t.Errorf("Expected message 'original error', got %v", systemErr.Message)
	}

	if systemErr.Original != originalErr {
		t.Errorf("Expected original error to be preserved")
	}

	if systemErr.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}

func TestErrorHandler_WithNilLogger(t *testing.T) {
	handler := NewErrorHandler(nil)
	originalErr := errors.New("test error")

	// Should not panic with nil logger
	cmd := handler.HandleSystemError("TestComponent", originalErr)
	if cmd == nil {
		t.Fatal("HandleSystemError() returned nil command with nil logger")
	}

	// Execute the command to ensure it works
	msg := cmd()
	errorMsg, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("Expected ErrorMsg, got %T", msg)
	}

	if errorMsg.Component != "TestComponent" {
		t.Errorf("Expected component 'TestComponent', got %v", errorMsg.Component)
	}
}

// Test error message conversion to tea.Msg
func TestErrorMsg_AsTeaMsg(t *testing.T) {
	systemErr := SystemError{
		Type:      SystemAccessError,
		Message:   "test message",
		Component: "CPU",
		Timestamp: time.Now(),
		Original:  errors.New("original"),
	}

	errorMsg := ErrorMsg(systemErr)

	// Test that it can be used as a tea.Msg
	var msg tea.Msg = errorMsg

	// Convert back and verify
	convertedErr, ok := msg.(ErrorMsg)
	if !ok {
		t.Fatalf("Expected ErrorMsg, got %T", msg)
	}

	if convertedErr.Type != SystemAccessError {
		t.Errorf("Expected SystemAccessError, got %v", convertedErr.Type)
	}

	if convertedErr.Component != "CPU" {
		t.Errorf("Expected component 'CPU', got %v", convertedErr.Component)
	}
}