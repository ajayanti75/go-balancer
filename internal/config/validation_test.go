package config

import (
	"testing"
	"time"

	"go-balancer/internal/errors"
)

func TestValidConfiguration(t *testing.T) {
	cfg := &Config{
		Port:                8000,
		Backends:            []string{"http://localhost:8080", "https://backend.example.com:443"},
		HealthCheckPath:     "/health",
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  2 * time.Second,
		BackendTimeout:      30 * time.Second,
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Expected valid configuration to pass validation, got: %v", err)
	}
}

func TestInvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"Port too low", 0},
		{"Port too high", 99999},
		{"Negative port", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Port:                tt.port,
				Backends:            []string{"http://localhost:8080"},
				HealthCheckPath:     "/",
				HealthCheckInterval: 10 * time.Second,
				HealthCheckTimeout:  2 * time.Second,
				BackendTimeout:      30 * time.Second,
			}

			err := cfg.Validate()
			if err == nil {
				t.Errorf("Expected port %d to be invalid", tt.port)
				return
			}

			validationErr, ok := err.(*ValidationError)
			if !ok {
				t.Errorf("Expected ValidationError, got %T", err)
				return
			}

			if len(validationErr.Errors) == 0 {
				t.Errorf("Expected validation errors for invalid port")
				return
			}

			// Check that we got a port error
			found := false
			for _, vErr := range validationErr.Errors {
				if vErr.Code == errors.ErrInvalidPort {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected to find ErrInvalidPort in validation errors")
			}
		})
	}
}

func TestInvalidBackends(t *testing.T) {
	tests := []struct {
		name             string
		backends         []string
		expectErrorCount int
	}{
		{"No backends", []string{}, 1},
		{"Empty backend", []string{""}, 1},
		{"Invalid URL", []string{"not-a-url"}, 2}, // Missing scheme and host
		{"Missing scheme", []string{"localhost:8080"}, 1},
		{"Missing host", []string{"http://"}, 1},
		{"Multiple invalid", []string{"invalid1", "invalid2"}, 4}, // 2 errors each
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Port:                8000,
				Backends:            tt.backends,
				HealthCheckPath:     "/",
				HealthCheckInterval: 10 * time.Second,
				HealthCheckTimeout:  2 * time.Second,
				BackendTimeout:      30 * time.Second,
			}

			err := cfg.Validate()
			if err == nil {
				t.Errorf("Expected backends %v to be invalid", tt.backends)
				return
			}

			validationErr, ok := err.(*ValidationError)
			if !ok {
				t.Errorf("Expected ValidationError, got %T", err)
				return
			}

			if len(validationErr.Errors) < tt.expectErrorCount {
				t.Errorf("Expected at least %d validation errors, got %d", tt.expectErrorCount, len(validationErr.Errors))
			}
		})
	}
}

func TestTimeoutValidation(t *testing.T) {
	tests := []struct {
		name           string
		healthInterval time.Duration
		healthTimeout  time.Duration
		backendTimeout time.Duration
		expectValid    bool
	}{
		{"Valid timeouts", 10 * time.Second, 2 * time.Second, 30 * time.Second, true},
		{"Health timeout >= interval", 10 * time.Second, 10 * time.Second, 30 * time.Second, false},
		{"Health timeout > interval", 10 * time.Second, 15 * time.Second, 30 * time.Second, false},
		{"Zero health interval", 0, 2 * time.Second, 30 * time.Second, false},
		{"Zero health timeout", 10 * time.Second, 0, 30 * time.Second, false},
		{"Zero backend timeout", 10 * time.Second, 2 * time.Second, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Port:                8000,
				Backends:            []string{"http://localhost:8080"},
				HealthCheckPath:     "/",
				HealthCheckInterval: tt.healthInterval,
				HealthCheckTimeout:  tt.healthTimeout,
				BackendTimeout:      tt.backendTimeout,
			}

			err := cfg.Validate()
			if tt.expectValid && err != nil {
				t.Errorf("Expected configuration to be valid, got error: %v", err)
			} else if !tt.expectValid && err == nil {
				t.Errorf("Expected configuration to be invalid, but validation passed")
			}
		})
	}
}

func TestValidationErrorAggregation(t *testing.T) {
	// Create a config with multiple errors
	cfg := &Config{
		Port:                99999,                       // Invalid port
		Backends:            []string{"invalid-url", ""}, // Invalid backends
		HealthCheckPath:     "",                          // Empty path
		HealthCheckInterval: 1 * time.Second,             //
		HealthCheckTimeout:  2 * time.Second,             // Timeout >= interval
		BackendTimeout:      0,                           // Zero timeout
	}

	err := cfg.Validate()
	if err == nil {
		t.Errorf("Expected configuration with multiple errors to fail validation")
		return
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Errorf("Expected ValidationError, got %T", err)
		return
	}

	// Should have multiple errors
	if len(validationErr.Errors) < 5 {
		t.Errorf("Expected at least 5 validation errors, got %d", len(validationErr.Errors))
	}

	// Test error message
	errorMsg := validationErr.Error()
	if errorMsg == "" {
		t.Errorf("Expected non-empty error message")
	}
}
