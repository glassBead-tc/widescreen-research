package timeout

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Manager manages timeout configuration
type Manager struct {
	global    time.Duration
	operation map[string]time.Duration
	dynamic   bool
	loadFunc  func() float64
	mu        sync.RWMutex
}

// NewManager creates a new timeout manager
func NewManager(globalTimeout time.Duration) *Manager {
	return &Manager{
		global:    globalTimeout,
		operation: make(map[string]time.Duration),
		dynamic:   false,
		loadFunc:  defaultLoadFunc,
	}
}

// Config represents timeout configuration
type Config struct {
	Global     time.Duration            `json:"global"`
	Operations map[string]time.Duration `json:"operations"`
	Dynamic    bool                     `json:"dynamic"`
}

// LoadConfig loads timeout configuration
func (m *Manager) LoadConfig(config Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.global = config.Global
	m.operation = config.Operations
	m.dynamic = config.Dynamic
}

// SetOperationTimeout sets timeout for specific operation
func (m *Manager) SetOperationTimeout(operation string, timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.operation[operation] = timeout
}

// GetTimeout returns appropriate timeout for operation
func (m *Manager) GetTimeout(ctx context.Context, operation string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Check context deadline first
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < m.global {
			return remaining
		}
	}
	
	// Get operation-specific timeout
	if opTimeout, exists := m.operation[operation]; exists {
		if m.dynamic {
			return m.adjustForLoad(opTimeout)
		}
		return opTimeout
	}
	
	// Default to global timeout
	if m.dynamic {
		return m.adjustForLoad(m.global)
	}
	return m.global
}

// adjustForLoad adjusts timeout based on system load
func (m *Manager) adjustForLoad(base time.Duration) time.Duration {
	load := m.loadFunc()
	
	switch {
	case load > 0.8:
		return base * 2  // Double timeout under high load
	case load > 0.6:
		return time.Duration(float64(base) * 1.5) // 50% increase under moderate load
	case load > 0.4:
		return time.Duration(float64(base) * 1.2) // 20% increase under light load
	default:
		return base
	}
}

// SetLoadFunction sets custom load calculation function
func (m *Manager) SetLoadFunction(f func() float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loadFunc = f
}

// WithTimeout creates context with timeout
func (m *Manager) WithTimeout(ctx context.Context, operation string) (context.Context, context.CancelFunc) {
	timeout := m.GetTimeout(ctx, operation)
	return context.WithTimeout(ctx, timeout)
}

// WithDeadline creates context with deadline
func (m *Manager) WithDeadline(ctx context.Context, operation string, deadline time.Time) (context.Context, context.CancelFunc) {
	timeout := m.GetTimeout(ctx, operation)
	operationDeadline := time.Now().Add(timeout)
	
	// Use earlier deadline
	if deadline.Before(operationDeadline) {
		return context.WithDeadline(ctx, deadline)
	}
	return context.WithDeadline(ctx, operationDeadline)
}

// defaultLoadFunc provides default load calculation
func defaultLoadFunc() float64 {
	// In production, this would check actual system metrics
	// For now, return a moderate load value
	return 0.3
}

// OperationTimeouts defines default timeouts for operations
var OperationTimeouts = map[string]time.Duration{
	"elicitation":           30 * time.Second,
	"orchestrate-research":  5 * time.Minute,
	"sequential-thinking":   60 * time.Second,
	"gcp-provision":        2 * time.Minute,
	"analyze-findings":     90 * time.Second,
	"websets-orchestrate":  3 * time.Minute,
	"websets-call":         45 * time.Second,
	"get_guide":            5 * time.Second,
}

// TimeoutMiddleware provides timeout enforcement for operations
type TimeoutMiddleware struct {
	manager *Manager
}

// NewTimeoutMiddleware creates new timeout middleware
func NewTimeoutMiddleware(manager *Manager) *TimeoutMiddleware {
	return &TimeoutMiddleware{
		manager: manager,
	}
}

// Wrap wraps an operation with timeout enforcement
func (m *TimeoutMiddleware) Wrap(operation string, handler func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		// Create context with timeout
		timeoutCtx, cancel := m.manager.WithTimeout(ctx, operation)
		defer cancel()
		
		// Create error channel
		errChan := make(chan error, 1)
		
		// Execute operation in goroutine
		go func() {
			errChan <- handler(timeoutCtx)
		}()
		
		// Wait for completion or timeout
		select {
		case err := <-errChan:
			return err
		case <-timeoutCtx.Done():
			if timeoutCtx.Err() == context.DeadlineExceeded {
				return &TimeoutError{
					Operation: operation,
					Timeout:   m.manager.GetTimeout(ctx, operation),
				}
			}
			return timeoutCtx.Err()
		}
	}
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	Operation string
	Timeout   time.Duration
}

// Error implements error interface
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("operation %s timed out after %v", e.Operation, e.Timeout)
}

// IsTimeout checks if error is a timeout
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for TimeoutError
	if _, ok := err.(*TimeoutError); ok {
		return true
	}
	
	// Check for context timeout
	if err == context.DeadlineExceeded {
		return true
	}
	
	return false
}

// TimeoutTracker tracks operation execution times
type TimeoutTracker struct {
	executions map[string][]time.Duration
	mu         sync.RWMutex
}

// NewTimeoutTracker creates new timeout tracker
func NewTimeoutTracker() *TimeoutTracker {
	return &TimeoutTracker{
		executions: make(map[string][]time.Duration),
	}
}

// Record records execution time
func (t *TimeoutTracker) Record(operation string, duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if t.executions[operation] == nil {
		t.executions[operation] = make([]time.Duration, 0, 100)
	}
	
	t.executions[operation] = append(t.executions[operation], duration)
	
	// Keep only last 100 executions
	if len(t.executions[operation]) > 100 {
		t.executions[operation] = t.executions[operation][1:]
	}
}

// GetP99 returns 99th percentile execution time
func (t *TimeoutTracker) GetP99(operation string) time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	durations := t.executions[operation]
	if len(durations) == 0 {
		return 0
	}
	
	// Simple P99 calculation (would use proper algorithm in production)
	index := int(float64(len(durations)) * 0.99)
	if index >= len(durations) {
		index = len(durations) - 1
	}
	
	return durations[index]
}

// GetRecommendedTimeout calculates recommended timeout based on history
func (t *TimeoutTracker) GetRecommendedTimeout(operation string) time.Duration {
	p99 := t.GetP99(operation)
	if p99 == 0 {
		// No history, use default
		if timeout, exists := OperationTimeouts[operation]; exists {
			return timeout
		}
		return 30 * time.Second
	}
	
	// Add 20% buffer to P99
	return time.Duration(float64(p99) * 1.2)
}