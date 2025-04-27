package circuitbreaker

import (
	"context"
	"errors"
	"time"

	"github.com/sony/gobreaker"
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

type CircuitBreaker struct {
	cb *gobreaker.CircuitBreaker
}

// Settings holds the configuration for the circuit breaker
type Settings struct {
	Name          string
	MaxRequests   uint32
	Interval      time.Duration
	Timeout       time.Duration
	ReadyToTrip   func(counts gobreaker.Counts) bool
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
}

// New creates a new circuit breaker with the given name and settings
func New(name string, settings ...Setting) *CircuitBreaker {
	// Default settings
	config := Settings{
		Name:        name,
		MaxRequests: 5,
		Interval:    10 * time.Second,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// Log state changes
			switch to {
			case gobreaker.StateOpen:
				// Circuit breaker is open, all requests will fail fast
			case gobreaker.StateHalfOpen:
				// Circuit breaker is half-open, allowing a limited number of requests
			case gobreaker.StateClosed:
				// Circuit breaker is closed, all requests are allowed
			}
		},
	}

	// Apply custom settings
	for _, setting := range settings {
		setting(&config)
	}

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:          config.Name,
		MaxRequests:   config.MaxRequests,
		Interval:      config.Interval,
		Timeout:       config.Timeout,
		ReadyToTrip:   config.ReadyToTrip,
		OnStateChange: config.OnStateChange,
	})

	return &CircuitBreaker{cb: cb}
}

// Setting is a function that configures the circuit breaker
type Setting func(*Settings)

// WithMaxRequests sets the maximum number of requests allowed to pass through when half-open
func WithMaxRequests(maxRequests uint32) Setting {
	return func(s *Settings) {
		s.MaxRequests = maxRequests
	}
}

// WithInterval sets the reset interval for the circuit breaker
func WithInterval(interval time.Duration) Setting {
	return func(s *Settings) {
		s.Interval = interval
	}
}

// WithTimeout sets the timeout for the circuit breaker to stay open
func WithTimeout(timeout time.Duration) Setting {
	return func(s *Settings) {
		s.Timeout = timeout
	}
}

// WithReadyToTrip sets the function that determines when to trip the circuit breaker
func WithReadyToTrip(readyToTrip func(counts gobreaker.Counts) bool) Setting {
	return func(s *Settings) {
		s.ReadyToTrip = readyToTrip
	}
}

// WithOnStateChange sets the function that is called when the circuit breaker state changes
func WithOnStateChange(onStateChange func(name string, from gobreaker.State, to gobreaker.State)) Setting {
	return func(s *Settings) {
		s.OnStateChange = onStateChange
	}
}

// Execute runs the given function with circuit breaker protection
func (c *CircuitBreaker) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	// Create a channel to receive the result
	resultChan := make(chan struct {
		val interface{}
		err error
	}, 1)

	// Run the function in a goroutine
	go func() {
		val, err := fn()
		resultChan <- struct {
			val interface{}
			err error
		}{val, err}
	}()

	// Wait for the result or context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultChan:
		return result.val, result.err
	}
}

// ExecuteWithBreaker runs the given function with circuit breaker protection
func (c *CircuitBreaker) ExecuteWithBreaker(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	// Execute the function with circuit breaker protection
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.Execute(ctx, fn)
	})

	if err != nil {
		if err == gobreaker.ErrOpenState {
			return nil, ErrCircuitOpen
		}
		return nil, err
	}

	return result, nil
}
