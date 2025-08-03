package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorCode represents different types of errors that can occur
type ErrorCode int

const (
	// Configuration errors
	ErrInvalidConfig ErrorCode = iota + 1000
	ErrInvalidPort
	ErrInvalidBackend
	ErrInvalidHealthCheck
	ErrInvalidTimeout

	// Backend errors
	ErrBackendUnavailable
	ErrBackendTimeout
	ErrBackendConnection
	ErrBackendResponse
	ErrNoHealthyBackends

	// Load balancer errors
	ErrStrategyFailure
	ErrPoolEmpty
	ErrMetricsFailure

	// Health check errors
	ErrHealthCheckFailed
	ErrHealthCheckTimeout

	// Request errors
	ErrRequestTimeout
	ErrRequestFailed
	ErrResponseCopy
)

// LoadBalancerError represents a structured error with context
type LoadBalancerError struct {
	Code      ErrorCode
	Message   string
	Cause     error
	Context   map[string]interface{}
	Timestamp time.Time
}

// Error implements the error interface
func (e *LoadBalancerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for error wrapping
func (e *LoadBalancerError) Unwrap() error {
	return e.Cause
}

// Is implements error comparison for errors.Is
func (e *LoadBalancerError) Is(target error) bool {
	if target == nil {
		return false
	}

	if t, ok := target.(*LoadBalancerError); ok {
		return e.Code == t.Code
	}

	return false
}

// HTTPStatusCode returns the appropriate HTTP status code for this error
func (e *LoadBalancerError) HTTPStatusCode() int {
	switch e.Code {
	case ErrInvalidConfig, ErrInvalidPort, ErrInvalidBackend, ErrInvalidHealthCheck, ErrInvalidTimeout:
		return http.StatusBadRequest
	case ErrBackendUnavailable, ErrNoHealthyBackends:
		return http.StatusServiceUnavailable
	case ErrBackendTimeout, ErrRequestTimeout:
		return http.StatusGatewayTimeout
	case ErrBackendConnection, ErrBackendResponse:
		return http.StatusBadGateway
	case ErrStrategyFailure, ErrPoolEmpty, ErrMetricsFailure:
		return http.StatusInternalServerError
	case ErrHealthCheckFailed, ErrHealthCheckTimeout:
		return http.StatusServiceUnavailable
	case ErrRequestFailed, ErrResponseCopy:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// NewError creates a new LoadBalancerError
func NewError(code ErrorCode, message string, cause error) *LoadBalancerError {
	return &LoadBalancerError{
		Code:      code,
		Message:   message,
		Cause:     cause,
		Context:   make(map[string]interface{}),
		Timestamp: time.Now(),
	}
}

// WithContext adds context information to the error
func (e *LoadBalancerError) WithContext(key string, value interface{}) *LoadBalancerError {
	e.Context[key] = value
	return e
}

// GetContext retrieves context information from the error
func (e *LoadBalancerError) GetContext(key string) (interface{}, bool) {
	value, exists := e.Context[key]
	return value, exists
}

// Configuration Error Constructors
func NewInvalidConfigError(message string, cause error) *LoadBalancerError {
	return NewError(ErrInvalidConfig, message, cause)
}

func NewInvalidPortError(port int) *LoadBalancerError {
	return NewError(ErrInvalidPort, fmt.Sprintf("invalid port: %d", port), nil).
		WithContext("port", port)
}

func NewInvalidBackendError(backend string, cause error) *LoadBalancerError {
	return NewError(ErrInvalidBackend, fmt.Sprintf("invalid backend: %s", backend), cause).
		WithContext("backend", backend)
}

func NewInvalidHealthCheckError(message string) *LoadBalancerError {
	return NewError(ErrInvalidHealthCheck, message, nil)
}

func NewInvalidTimeoutError(timeout time.Duration, field string) *LoadBalancerError {
	return NewError(ErrInvalidTimeout, fmt.Sprintf("invalid %s timeout: %s", field, timeout), nil).
		WithContext("timeout", timeout).
		WithContext("field", field)
}

// Backend Error Constructors
func NewBackendUnavailableError(backend string) *LoadBalancerError {
	return NewError(ErrBackendUnavailable, fmt.Sprintf("backend unavailable: %s", backend), nil).
		WithContext("backend", backend)
}

func NewBackendTimeoutError(backend string, cause error) *LoadBalancerError {
	return NewError(ErrBackendTimeout, fmt.Sprintf("backend timeout: %s", backend), cause).
		WithContext("backend", backend)
}

func NewBackendConnectionError(backend string, cause error) *LoadBalancerError {
	return NewError(ErrBackendConnection, fmt.Sprintf("backend connection failed: %s", backend), cause).
		WithContext("backend", backend)
}

func NewBackendResponseError(backend string, statusCode int) *LoadBalancerError {
	return NewError(ErrBackendResponse, fmt.Sprintf("backend response error: %s (status: %d)", backend, statusCode), nil).
		WithContext("backend", backend).
		WithContext("status_code", statusCode)
}

func NewNoHealthyBackendsError() *LoadBalancerError {
	return NewError(ErrNoHealthyBackends, "no healthy backends available", nil)
}

// Load Balancer Error Constructors
func NewStrategyFailureError(strategy string, cause error) *LoadBalancerError {
	return NewError(ErrStrategyFailure, fmt.Sprintf("load balancing strategy failed: %s", strategy), cause).
		WithContext("strategy", strategy)
}

func NewPoolEmptyError() *LoadBalancerError {
	return NewError(ErrPoolEmpty, "server pool is empty", nil)
}

func NewMetricsFailureError(cause error) *LoadBalancerError {
	return NewError(ErrMetricsFailure, "metrics collection failed", cause)
}

// Health Check Error Constructors
func NewHealthCheckFailedError(backend string, cause error) *LoadBalancerError {
	return NewError(ErrHealthCheckFailed, fmt.Sprintf("health check failed: %s", backend), cause).
		WithContext("backend", backend)
}

func NewHealthCheckTimeoutError(backend string) *LoadBalancerError {
	return NewError(ErrHealthCheckTimeout, fmt.Sprintf("health check timeout: %s", backend), nil).
		WithContext("backend", backend)
}

// Request Error Constructors
func NewRequestTimeoutError(cause error) *LoadBalancerError {
	return NewError(ErrRequestTimeout, "request timeout", cause)
}

func NewRequestFailedError(cause error) *LoadBalancerError {
	return NewError(ErrRequestFailed, "request failed", cause)
}

func NewResponseCopyError(cause error) *LoadBalancerError {
	return NewError(ErrResponseCopy, "failed to copy response", cause)
}

// IsConfigurationError checks if the error is a configuration-related error
func IsConfigurationError(err error) bool {
	if lbErr, ok := err.(*LoadBalancerError); ok {
		return lbErr.Code >= ErrInvalidConfig && lbErr.Code <= ErrInvalidTimeout
	}
	return false
}

// IsBackendError checks if the error is a backend-related error
func IsBackendError(err error) bool {
	if lbErr, ok := err.(*LoadBalancerError); ok {
		return lbErr.Code >= ErrBackendUnavailable && lbErr.Code <= ErrNoHealthyBackends
	}
	return false
}

// IsHealthCheckError checks if the error is a health check-related error
func IsHealthCheckError(err error) bool {
	if lbErr, ok := err.(*LoadBalancerError); ok {
		return lbErr.Code >= ErrHealthCheckFailed && lbErr.Code <= ErrHealthCheckTimeout
	}
	return false
}

// IsRequestError checks if the error is a request-related error
func IsRequestError(err error) bool {
	if lbErr, ok := err.(*LoadBalancerError); ok {
		return lbErr.Code >= ErrRequestTimeout && lbErr.Code <= ErrResponseCopy
	}
	return false
}
