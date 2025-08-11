package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/google/uuid"
)

// ErrorCategory represents the category of error
type ErrorCategory string

const (
	// Infrastructure Errors (1xxx)
	ErrConnectionFailed    = "MCP-1001" // Network connectivity issues
	ErrTimeout            = "MCP-1002" // Operation timeout
	ErrRateLimit          = "MCP-1003" // Rate limit exceeded
	ErrQuotaExceeded      = "MCP-1004" // Resource quota exceeded
	ErrServiceUnavailable = "MCP-1005" // Dependency service down

	// Authentication Errors (2xxx)
	ErrAuthMissing    = "MCP-2001" // Missing credentials
	ErrAuthInvalid    = "MCP-2002" // Invalid credentials
	ErrAuthExpired    = "MCP-2003" // Expired token/key
	ErrAuthPermission = "MCP-2004" // Insufficient permissions

	// Validation Errors (3xxx)
	ErrInvalidInput        = "MCP-3001" // Invalid parameters
	ErrMissingRequired     = "MCP-3002" // Missing required field
	ErrFormatInvalid       = "MCP-3003" // Invalid format/schema
	ErrConstraintViolation = "MCP-3004" // Business rule violation

	// Operation Errors (4xxx)
	ErrOperationUnknown = "MCP-4001" // Unknown operation
	ErrSessionInvalid   = "MCP-4002" // Invalid session ID
	ErrStateConflict    = "MCP-4003" // Operation state conflict
	ErrResourceNotFound = "MCP-4004" // Resource not found

	// System Errors (5xxx)
	ErrInternalError   = "MCP-5001" // Unexpected internal error
	ErrMemoryExhausted = "MCP-5002" // Out of memory
	ErrDiskFull        = "MCP-5003" // Storage exhausted
	ErrPanic           = "MCP-5004" // Panic recovery
)

// ErrorSeverity represents the severity level
type ErrorSeverity int

const (
	SeverityCritical ErrorSeverity = iota // System failure, immediate action
	SeverityHigh                          // Service degraded, urgent
	SeverityMedium                        // Feature impacted, important
	SeverityLow                          // Minor issue, informational
)

// MCPError represents a comprehensive error with context
type MCPError struct {
	Code          string                 `json:"code"`
	Category      ErrorCategory          `json:"category"`
	Message       string                 `json:"message"`
	Severity      ErrorSeverity          `json:"severity"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Retryable     bool                   `json:"retryable"`
	RetryAfter    *time.Duration         `json:"retry_after,omitempty"`
	UserMessage   string                 `json:"user_message,omitempty"`
	DebugInfo     string                 `json:"debug_info,omitempty"`
	CorrelationID string                 `json:"correlation_id"`
	StackTrace    []string               `json:"stack_trace,omitempty"`
}

// Error implements the error interface
func (e *MCPError) Error() string {
	if e.UserMessage != "" {
		return e.UserMessage
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ShouldRetry determines if the error is retryable
func (e *MCPError) ShouldRetry() bool {
	return e.Retryable && e.Severity > SeverityCritical
}

// WithContext adds context to the error
func (e *MCPError) WithContext(key string, value interface{}) *MCPError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// ToJSON serializes the error to JSON
func (e *MCPError) ToJSON() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// New creates a new MCPError
func New(code string, message string) *MCPError {
	return &MCPError{
		Code:          code,
		Message:       message,
		Timestamp:     time.Now(),
		CorrelationID: uuid.New().String(),
		StackTrace:    captureStackTrace(),
	}
}

// NewWithSeverity creates an error with specific severity
func NewWithSeverity(code string, message string, severity ErrorSeverity) *MCPError {
	err := New(code, message)
	err.Severity = severity
	err.Category = getCategoryFromCode(code)
	err.Retryable = isRetryableCode(code)
	return err
}

// Wrap wraps an existing error
func Wrap(err error, code string) *MCPError {
	if err == nil {
		return nil
	}

	// If already an MCPError, enhance it
	if mcpErr, ok := err.(*MCPError); ok {
		mcpErr.Code = code
		return mcpErr
	}

	// Create new MCPError from standard error
	return &MCPError{
		Code:          code,
		Message:       err.Error(),
		Timestamp:     time.Now(),
		CorrelationID: uuid.New().String(),
		Category:      getCategoryFromCode(code),
		Severity:      getSeverityFromCode(code),
		Retryable:     isRetryableCode(code),
		StackTrace:    captureStackTrace(),
	}
}

// getCategoryFromCode determines category from error code
func getCategoryFromCode(code string) ErrorCategory {
	if len(code) < 6 {
		return ErrorCategory("unknown")
	}

	prefix := code[4:5] // Get the first digit after "MCP-"
	switch prefix {
	case "1":
		return ErrorCategory("infrastructure")
	case "2":
		return ErrorCategory("authentication")
	case "3":
		return ErrorCategory("validation")
	case "4":
		return ErrorCategory("operation")
	case "5":
		return ErrorCategory("system")
	default:
		return ErrorCategory("unknown")
	}
}

// getSeverityFromCode determines severity from error code
func getSeverityFromCode(code string) ErrorSeverity {
	switch code {
	case ErrPanic, ErrMemoryExhausted, ErrDiskFull:
		return SeverityCritical
	case ErrServiceUnavailable, ErrAuthExpired:
		return SeverityHigh
	case ErrRateLimit, ErrSessionInvalid:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// isRetryableCode determines if an error code is retryable
func isRetryableCode(code string) bool {
	switch code {
	case ErrConnectionFailed, ErrTimeout, ErrRateLimit, ErrServiceUnavailable:
		return true
	case ErrAuthInvalid, ErrInvalidInput, ErrOperationUnknown:
		return false
	default:
		return false
	}
}

// captureStackTrace captures the current stack trace
func captureStackTrace() []string {
	const maxDepth = 10
	pc := make([]uintptr, maxDepth)
	n := runtime.Callers(3, pc) // Skip runtime.Callers, captureStackTrace, and caller
	
	stack := make([]string, 0, n)
	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pc[i])
		if fn != nil {
			file, line := fn.FileLine(pc[i])
			stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
		}
	}
	return stack
}

// ErrorRecovery provides recovery strategies
type ErrorRecovery struct {
	MaxRetries     int
	RetryDelay     time.Duration
	BackoffFactor  float64
	CircuitBreaker *CircuitBreaker
}

// CircuitBreaker implements circuit breaker pattern
type CircuitBreaker struct {
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
	
	failures  int
	successes int
	state     string // "closed", "open", "half-open"
	lastError time.Time
}

// RecoverWithStrategy attempts recovery based on error type
func RecoverWithStrategy(err *MCPError, recovery *ErrorRecovery) error {
	if !err.ShouldRetry() {
		return err
	}

	// Check circuit breaker
	if recovery.CircuitBreaker != nil {
		if recovery.CircuitBreaker.IsOpen() {
			return fmt.Errorf("circuit breaker open: %w", err)
		}
	}

	// Apply retry delay if specified
	if err.RetryAfter != nil {
		time.Sleep(*err.RetryAfter)
	} else if recovery.RetryDelay > 0 {
		time.Sleep(recovery.RetryDelay)
	}

	return nil // Ready to retry
}

// IsOpen checks if circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	if cb.state == "open" {
		// Check if timeout has passed
		if time.Since(cb.lastError) > cb.Timeout {
			cb.state = "half-open"
			return false
		}
		return true
	}
	return false
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.successes++
	if cb.state == "half-open" && cb.successes >= cb.SuccessThreshold {
		cb.state = "closed"
		cb.failures = 0
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.failures++
	cb.lastError = time.Now()
	
	if cb.failures >= cb.FailureThreshold {
		cb.state = "open"
		cb.successes = 0
	}
}