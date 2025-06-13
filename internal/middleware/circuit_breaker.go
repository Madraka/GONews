package middleware

import (
	"errors"
	"sync"
	"time"

	"news/internal/metrics"
)

// Circuit breaker states
const (
	StateClosed   = "closed"    // Normal operation - requests allowed
	StateOpen     = "open"      // Circuit breaker triggered - requests rejected
	StateHalfOpen = "half_open" // Testing if service has recovered
)

var (
	// ErrCircuitBreakerOpen is returned when the circuit breaker is open
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name                 string
	state                string
	failureThreshold     int
	consecutiveFailures  int
	successThreshold     int
	consecutiveSuccesses int
	timeout              time.Duration
	lastFailureTime      time.Time
	mutex                *sync.RWMutex
	onStateChange        func(name, from, to string)
}

// CircuitBreakerOption is a function that configures a CircuitBreaker
type CircuitBreakerOption func(*CircuitBreaker)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, options ...CircuitBreakerOption) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:             name,
		state:            StateClosed,
		failureThreshold: 5,               // Default: 5 failures
		successThreshold: 2,               // Default: 2 successes
		timeout:          5 * time.Second, // Default: 5 second timeout
		mutex:            &sync.RWMutex{},
		onStateChange:    func(name, from, to string) {}, // Default empty handler
	}

	// Apply options
	for _, option := range options {
		option(cb)
	}

	return cb
}

// WithFailureThreshold sets the failure threshold
func WithFailureThreshold(threshold int) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.failureThreshold = threshold
	}
}

// WithSuccessThreshold sets the success threshold
func WithSuccessThreshold(threshold int) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.successThreshold = threshold
	}
}

// WithTimeout sets the timeout duration
func WithTimeout(timeout time.Duration) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.timeout = timeout
	}
}

// WithOnStateChange sets the state change handler
func WithOnStateChange(handler func(name, from, to string)) CircuitBreakerOption {
	return func(cb *CircuitBreaker) {
		cb.onStateChange = handler
	}
}

// Execute executes the given function if the circuit breaker allows it
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Track metrics
	defer metrics.TrackDatabaseOperation("circuit_breaker_" + cb.name)()

	if !cb.AllowRequest() {
		return ErrCircuitBreakerOpen
	}

	err := fn()

	cb.HandleResult(err)
	return err
}

// AllowRequest checks if the request should be allowed
func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	// If closed, allow request
	if cb.state == StateClosed {
		return true
	}

	// If open, check timeout
	if cb.state == StateOpen {
		// Check if timeout has elapsed
		if time.Since(cb.lastFailureTime) > cb.timeout {
			// Transition to half-open
			cb.mutex.RUnlock()
			cb.transitionState(StateHalfOpen)
			cb.mutex.RLock()
			return true
		}
		return false
	}

	// If half-open, allow limited traffic
	return true
}

// HandleResult processes the result of a request
func (cb *CircuitBreaker) HandleResult(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		// Failed request
		cb.consecutiveSuccesses = 0
		cb.consecutiveFailures++
		cb.lastFailureTime = time.Now()

		// Check if threshold reached
		if (cb.state == StateClosed && cb.consecutiveFailures >= cb.failureThreshold) ||
			cb.state == StateHalfOpen {
			cb.transitionState(StateOpen)
		}
	} else {
		// Successful request
		cb.consecutiveFailures = 0

		// In half-open state, count successes
		if cb.state == StateHalfOpen {
			cb.consecutiveSuccesses++

			// Check if threshold reached
			if cb.consecutiveSuccesses >= cb.successThreshold {
				cb.transitionState(StateClosed)
			}
		} else {
			// Reset success count in normal operation
			cb.consecutiveSuccesses = 0
		}
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	oldState := cb.state
	cb.state = StateClosed
	cb.consecutiveFailures = 0
	cb.consecutiveSuccesses = 0

	// Notify of state change
	cb.onStateChange(cb.name, oldState, StateClosed)
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() string {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetConsecutiveFailures returns the number of consecutive failures
func (cb *CircuitBreaker) GetConsecutiveFailures() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.consecutiveFailures
}

// GetConsecutiveSuccesses returns the number of consecutive successes
func (cb *CircuitBreaker) GetConsecutiveSuccesses() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.consecutiveSuccesses
}

// GetLastFailureTime returns the time of the last failure
func (cb *CircuitBreaker) GetLastFailureTime() time.Time {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.lastFailureTime
}

// private helper to transition state with notification
func (cb *CircuitBreaker) transitionState(newState string) {
	oldState := cb.state
	cb.state = newState

	// Reset counters on state change
	if oldState != newState {
		if newState == StateClosed {
			cb.consecutiveFailures = 0
		} else if newState == StateHalfOpen {
			cb.consecutiveSuccesses = 0
		}

		// Call state change handler
		cb.onStateChange(cb.name, oldState, newState)
	}
}
