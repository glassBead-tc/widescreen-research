package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/errors"
)

// Strategy defines retry strategy interface
type Strategy interface {
	NextDelay(attempt int) time.Duration
	ShouldRetry(attempt int, err error) bool
}

// Config defines retry configuration
type Config struct {
	MaxAttempts int
	Strategy    Strategy
	Jitter      float64
	OnRetry     func(attempt int, err error)
}

// ExponentialBackoff implements exponential backoff strategy
type ExponentialBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// NextDelay calculates next delay for exponential backoff
func (e *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	delay := float64(e.InitialDelay) * math.Pow(e.Multiplier, float64(attempt))
	if delay > float64(e.MaxDelay) {
		return e.MaxDelay
	}
	return time.Duration(delay)
}

// ShouldRetry determines if retry should continue
func (e *ExponentialBackoff) ShouldRetry(attempt int, err error) bool {
	// Check if error is retryable
	if mcpErr, ok := err.(*errors.MCPError); ok {
		return mcpErr.ShouldRetry()
	}
	return true // Default to retry for non-MCP errors
}

// LinearBackoff implements linear backoff strategy
type LinearBackoff struct {
	Delay       time.Duration
	MaxAttempts int
}

// NextDelay returns constant delay for linear backoff
func (l *LinearBackoff) NextDelay(attempt int) time.Duration {
	return l.Delay
}

// ShouldRetry determines if retry should continue
func (l *LinearBackoff) ShouldRetry(attempt int, err error) bool {
	if attempt >= l.MaxAttempts {
		return false
	}
	if mcpErr, ok := err.(*errors.MCPError); ok {
		return mcpErr.ShouldRetry()
	}
	return true
}

// CircuitBreakerStrategy implements circuit breaker pattern
type CircuitBreakerStrategy struct {
	FailureThreshold   int
	SuccessThreshold   int
	Timeout           time.Duration
	HalfOpenAttempts  int
	
	failures         int
	successes        int
	state            string
	lastFailureTime  time.Time
	halfOpenAttempt  int
}

// NextDelay returns delay based on circuit breaker state
func (c *CircuitBreakerStrategy) NextDelay(attempt int) time.Duration {
	switch c.state {
	case "open":
		// Check if we can transition to half-open
		if time.Since(c.lastFailureTime) > c.Timeout {
			c.state = "half-open"
			c.halfOpenAttempt = 0
			return 100 * time.Millisecond
		}
		return c.Timeout
	case "half-open":
		return 500 * time.Millisecond
	default:
		return 0
	}
}

// ShouldRetry determines if retry should continue based on circuit state
func (c *CircuitBreakerStrategy) ShouldRetry(attempt int, err error) bool {
	if err == nil {
		c.RecordSuccess()
		return false
	}

	c.RecordFailure()
	
	switch c.state {
	case "open":
		return false
	case "half-open":
		c.halfOpenAttempt++
		return c.halfOpenAttempt < c.HalfOpenAttempts
	default:
		return true
	}
}

// RecordSuccess records successful attempt
func (c *CircuitBreakerStrategy) RecordSuccess() {
	c.successes++
	if c.state == "half-open" && c.successes >= c.SuccessThreshold {
		c.state = "closed"
		c.failures = 0
	}
}

// RecordFailure records failed attempt
func (c *CircuitBreakerStrategy) RecordFailure() {
	c.failures++
	c.lastFailureTime = time.Now()
	
	if c.failures >= c.FailureThreshold {
		c.state = "open"
		c.successes = 0
	}
}

// ExecuteWithRetry executes operation with retry logic
func ExecuteWithRetry[T any](
	ctx context.Context,
	operation func() (T, error),
	config Config,
) (T, error) {
	var result T
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Execute operation
		result, err := operation()
		if err == nil {
			return result, nil
		}
		
		// Check if we should retry
		if !config.Strategy.ShouldRetry(attempt, err) {
			return result, err
		}
		
		// Calculate delay
		delay := config.Strategy.NextDelay(attempt)
		
		// Apply jitter if configured
		if config.Jitter > 0 {
			delay = applyJitter(delay, config.Jitter)
		}
		
		// Call retry callback if provided
		if config.OnRetry != nil {
			config.OnRetry(attempt, err)
		}
		
		// Wait with context cancellation check
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return result, fmt.Errorf("retry cancelled: %w", ctx.Err())
		}
		
		lastErr = err
	}
	
	return result, fmt.Errorf("max retries (%d) exceeded: %w", config.MaxAttempts, lastErr)
}

// ExecuteWithRetryAsync executes operation with retry logic asynchronously
func ExecuteWithRetryAsync[T any](
	ctx context.Context,
	operation func() (T, error),
	config Config,
) <-chan Result[T] {
	resultChan := make(chan Result[T], 1)
	
	go func() {
		defer close(resultChan)
		result, err := ExecuteWithRetry(ctx, operation, config)
		resultChan <- Result[T]{Value: result, Error: err}
	}()
	
	return resultChan
}

// Result wraps async operation result
type Result[T any] struct {
	Value T
	Error error
}

// applyJitter adds random jitter to delay
func applyJitter(delay time.Duration, jitterFactor float64) time.Duration {
	jitter := float64(delay) * jitterFactor
	randomJitter := (rand.Float64() - 0.5) * 2 * jitter
	finalDelay := float64(delay) + randomJitter
	
	if finalDelay < 0 {
		return 0
	}
	
	return time.Duration(finalDelay)
}

// DefaultConfigs provides pre-configured retry configurations
var DefaultConfigs = struct {
	Fast     Config
	Standard Config
	Robust   Config
}{
	Fast: Config{
		MaxAttempts: 3,
		Strategy: &LinearBackoff{
			Delay:       100 * time.Millisecond,
			MaxAttempts: 3,
		},
		Jitter: 0.1,
	},
	Standard: Config{
		MaxAttempts: 5,
		Strategy: &ExponentialBackoff{
			InitialDelay: 100 * time.Millisecond,
			MaxDelay:     10 * time.Second,
			Multiplier:   2,
		},
		Jitter: 0.2,
	},
	Robust: Config{
		MaxAttempts: 10,
		Strategy: &CircuitBreakerStrategy{
			FailureThreshold:  5,
			SuccessThreshold:  2,
			Timeout:          60 * time.Second,
			HalfOpenAttempts: 3,
		},
		Jitter: 0.3,
	},
}