package models

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ErrorType represents different categories of errors
type ErrorType int

const (
	SystemAccessError ErrorType = iota
	DataCollectionError
	RenderError
	PermissionError
	TemporaryError
)

// SystemError represents an error with context and type information
type SystemError struct {
	Type      ErrorType
	Message   string
	Component string
	Timestamp time.Time
	Original  error
}

// Error implements the error interface
func (e SystemError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Component, e.typeString(), e.Message)
}

// typeString returns a human-readable string for the error type
func (e SystemError) typeString() string {
	switch e.Type {
	case SystemAccessError:
		return "System Access Error"
	case DataCollectionError:
		return "Data Collection Error"
	case RenderError:
		return "Render Error"
	case PermissionError:
		return "Permission Error"
	case TemporaryError:
		return "Temporary Error"
	default:
		return "Unknown Error"
	}
}

// IsRecoverable returns true if the error is recoverable
func (e SystemError) IsRecoverable() bool {
	return e.Type == TemporaryError || e.Type == DataCollectionError
}

// ErrorMsg represents an error message for the Bubble Tea framework
type ErrorMsg SystemError

// ErrorHandler manages error handling and recovery
type ErrorHandler struct {
	logger *log.Logger
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *log.Logger) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
	}
}

// HandleSystemError handles system access errors
func (h *ErrorHandler) HandleSystemError(component string, err error) tea.Cmd {
	systemErr := SystemError{
		Type:      SystemAccessError,
		Message:   err.Error(),
		Component: component,
		Timestamp: time.Now(),
		Original:  err,
	}
	
	if h.logger != nil {
		h.logger.Printf("System error in %s: %v", component, err)
	}
	
	return func() tea.Msg {
		return ErrorMsg(systemErr)
	}
}

// HandleDataError handles data collection errors
func (h *ErrorHandler) HandleDataError(component string, err error) tea.Cmd {
	systemErr := SystemError{
		Type:      DataCollectionError,
		Message:   err.Error(),
		Component: component,
		Timestamp: time.Now(),
		Original:  err,
	}
	
	if h.logger != nil {
		h.logger.Printf("Data collection error in %s: %v", component, err)
	}
	
	return func() tea.Msg {
		return ErrorMsg(systemErr)
	}
}

// HandlePermissionError handles permission-related errors
func (h *ErrorHandler) HandlePermissionError(component string, err error) tea.Cmd {
	systemErr := SystemError{
		Type:      PermissionError,
		Message:   err.Error(),
		Component: component,
		Timestamp: time.Now(),
		Original:  err,
	}
	
	if h.logger != nil {
		h.logger.Printf("Permission error in %s: %v", component, err)
	}
	
	return func() tea.Msg {
		return ErrorMsg(systemErr)
	}
}

// HandleTemporaryError handles temporary errors that may resolve themselves
func (h *ErrorHandler) HandleTemporaryError(component string, err error) tea.Cmd {
	systemErr := SystemError{
		Type:      TemporaryError,
		Message:   err.Error(),
		Component: component,
		Timestamp: time.Now(),
		Original:  err,
	}
	
	if h.logger != nil {
		h.logger.Printf("Temporary error in %s: %v", component, err)
	}
	
	return func() tea.Msg {
		return ErrorMsg(systemErr)
	}
}

// HandleRenderError handles rendering-related errors
func (h *ErrorHandler) HandleRenderError(component string, err error) tea.Cmd {
	systemErr := SystemError{
		Type:      RenderError,
		Message:   err.Error(),
		Component: component,
		Timestamp: time.Now(),
		Original:  err,
	}
	
	if h.logger != nil {
		h.logger.Printf("Render error in %s: %v", component, err)
	}
	
	return func() tea.Msg {
		return ErrorMsg(systemErr)
	}
}

// CreateSystemError creates a new system error
func CreateSystemError(errorType ErrorType, component, message string, original error) SystemError {
	return SystemError{
		Type:      errorType,
		Message:   message,
		Component: component,
		Timestamp: time.Now(),
		Original:  original,
	}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, component string, errorType ErrorType) SystemError {
	return SystemError{
		Type:      errorType,
		Message:   err.Error(),
		Component: component,
		Timestamp: time.Now(),
		Original:  err,
	}
}