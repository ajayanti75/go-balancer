package errors

import (
	"net/http"
	"strings"
	"testing"
)

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name             string
		err              *LoadBalancerError
		expectedCode     ErrorCode
		expectedHTTP     int
		expectedCategory string
	}{
		{
			name:             "Invalid Port Error",
			err:              NewInvalidPortError(99999),
			expectedCode:     ErrInvalidPort,
			expectedHTTP:     http.StatusBadRequest,
			expectedCategory: "configuration",
		},
		{
			name:             "Backend Unavailable Error",
			err:              NewBackendUnavailableError("backend-1"),
			expectedCode:     ErrBackendUnavailable,
			expectedHTTP:     http.StatusServiceUnavailable,
			expectedCategory: "backend",
		},
		{
			name:             "No Healthy Backends Error",
			err:              NewNoHealthyBackendsError(),
			expectedCode:     ErrNoHealthyBackends,
			expectedHTTP:     http.StatusServiceUnavailable,
			expectedCategory: "backend",
		},
		{
			name:             "Health Check Failed Error",
			err:              NewHealthCheckFailedError("backend-1", nil),
			expectedCode:     ErrHealthCheckFailed,
			expectedHTTP:     http.StatusServiceUnavailable,
			expectedCategory: "health_check",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.expectedCode {
				t.Errorf("Expected error code %d, got %d", tt.expectedCode, tt.err.Code)
			}

			if tt.err.HTTPStatusCode() != tt.expectedHTTP {
				t.Errorf("Expected HTTP status %d, got %d", tt.expectedHTTP, tt.err.HTTPStatusCode())
			}

			// Test error categorization
			switch tt.expectedCategory {
			case "configuration":
				if !IsConfigurationError(tt.err) {
					t.Errorf("Expected configuration error, but categorization failed")
				}
			case "backend":
				if !IsBackendError(tt.err) {
					t.Errorf("Expected backend error, but categorization failed")
				}
			case "health_check":
				if !IsHealthCheckError(tt.err) {
					t.Errorf("Expected health check error, but categorization failed")
				}
			}
		})
	}
}

func TestErrorContext(t *testing.T) {
	err := NewInvalidBackendError("test-backend", nil).
		WithContext("port", 8080).
		WithContext("scheme", "http")

	// Test context retrieval
	if backend, exists := err.GetContext("backend"); !exists || backend != "test-backend" {
		t.Errorf("Expected backend context 'test-backend', got %v (exists: %v)", backend, exists)
	}

	if port, exists := err.GetContext("port"); !exists || port != 8080 {
		t.Errorf("Expected port context 8080, got %v (exists: %v)", port, exists)
	}

	if scheme, exists := err.GetContext("scheme"); !exists || scheme != "http" {
		t.Errorf("Expected scheme context 'http', got %v (exists: %v)", scheme, exists)
	}

	// Test non-existent context
	if _, exists := err.GetContext("nonexistent"); exists {
		t.Errorf("Expected nonexistent context to not exist")
	}
}

func TestErrorWrapping(t *testing.T) {
	originalErr := NewBackendConnectionError("backend-1", http.ErrServerClosed)

	// Test error unwrapping
	if originalErr.Unwrap() != http.ErrServerClosed {
		t.Errorf("Expected unwrapped error to be http.ErrServerClosed")
	}

	// Test error message includes cause
	expectedSubstring := "server closed"
	if !contains(strings.ToLower(originalErr.Error()), expectedSubstring) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedSubstring, originalErr.Error())
	}
}

// Helper function since Go 1.21 doesn't have strings.Contains in all contexts
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
