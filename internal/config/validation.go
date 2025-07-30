package config

import (
	"fmt"
	"net/url"
)

// ValidateConfig validates a configuration struct
func ValidateConfig(c *Config) error {
	var errors []string

	// Validate port
	if c.Port <= 0 || c.Port > 65535 {
		errors = append(errors, fmt.Sprintf("port %d must be between 1 and 65535", c.Port))
	}

	// Validate backends
	if len(c.Backends) == 0 {
		errors = append(errors, "at least one backend is required")
	}

	// Validate each backend URL
	for i, backend := range c.Backends {
		if _, err := url.Parse(backend); err != nil {
			errors = append(errors, fmt.Sprintf("backend[%d] '%s' is not a valid URL: %v", i, backend, err))
		}
	}

	// Validate health check path
	if c.HealthCheckPath == "" {
		errors = append(errors, "health check path cannot be empty")
	}

	// Validate health check interval
	if c.HealthCheckInterval <= 0 {
		errors = append(errors, "health check interval must be positive")
	}

	// Validate health check timeout
	if c.HealthCheckTimeout <= 0 {
		errors = append(errors, "health check timeout must be positive")
	}

	// Validate timeout relationship
	if c.HealthCheckTimeout >= c.HealthCheckInterval {
		errors = append(errors, fmt.Sprintf("health check timeout (%s) must be less than interval (%s)", 
			c.HealthCheckTimeout, c.HealthCheckInterval))
	}

	// Validate backend timeout
	if c.BackendTimeout <= 0 {
		errors = append(errors, "backend timeout must be positive")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n  - %s", 
			fmt.Sprintf("%s", errors[0]))
	}

	return nil
}

// Validate method on Config struct
func (c *Config) Validate() error {
	return ValidateConfig(c)
}
