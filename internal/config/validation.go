package config

import (
	"fmt"
	"net/url"

	"go-balancer/internal/errors"
)

// ValidationError aggregates multiple validation errors
type ValidationError struct {
	Errors []*errors.LoadBalancerError
}

func (ve *ValidationError) Error() string {
	if len(ve.Errors) == 0 {
		return "no validation errors"
	}

	if len(ve.Errors) == 1 {
		return fmt.Sprintf("validation failed: %s", ve.Errors[0].Error())
	}

	return fmt.Sprintf("validation failed with %d errors: %s (and %d more)",
		len(ve.Errors), ve.Errors[0].Error(), len(ve.Errors)-1)
}

func (ve *ValidationError) Add(err *errors.LoadBalancerError) {
	ve.Errors = append(ve.Errors, err)
}

func (ve *ValidationError) HasErrors() bool {
	return len(ve.Errors) > 0
}

// ValidateConfig validates a configuration struct
func ValidateConfig(c *Config) error {
	validationErr := &ValidationError{}

	// Validate port
	if c.Port <= 0 || c.Port > 65535 {
		validationErr.Add(errors.NewInvalidPortError(c.Port))
	}

	// Validate backends
	if len(c.Backends) == 0 {
		validationErr.Add(errors.NewInvalidConfigError("at least one backend is required", nil))
	}

	// Validate each backend URL
	for i, backend := range c.Backends {
		if backend == "" {
			validationErr.Add(errors.NewInvalidBackendError(
				fmt.Sprintf("backend[%d]", i),
				fmt.Errorf("backend cannot be empty"),
			).WithContext("index", i))
			continue
		}

		parsedURL, err := url.Parse(backend)
		if err != nil {
			validationErr.Add(errors.NewInvalidBackendError(backend, err).WithContext("index", i))
			continue
		}

		if parsedURL.Scheme == "" {
			validationErr.Add(errors.NewInvalidBackendError(
				backend,
				fmt.Errorf("must include a scheme (http:// or https://)"),
			).WithContext("index", i))
		}

		if parsedURL.Host == "" {
			validationErr.Add(errors.NewInvalidBackendError(
				backend,
				fmt.Errorf("must include a host"),
			).WithContext("index", i))
		}
	}

	// Validate health check path
	if c.HealthCheckPath == "" {
		validationErr.Add(errors.NewInvalidHealthCheckError("health check path cannot be empty"))
	}

	// Validate health check interval
	if c.HealthCheckInterval <= 0 {
		validationErr.Add(errors.NewInvalidTimeoutError(c.HealthCheckInterval, "health check interval"))
	}

	// Validate health check timeout
	if c.HealthCheckTimeout <= 0 {
		validationErr.Add(errors.NewInvalidTimeoutError(c.HealthCheckTimeout, "health check timeout"))
	}

	// Validate timeout relationship
	if c.HealthCheckTimeout >= c.HealthCheckInterval {
		validationErr.Add(errors.NewInvalidConfigError(
			fmt.Sprintf("health check timeout (%s) must be less than interval (%s)",
				c.HealthCheckTimeout, c.HealthCheckInterval),
			nil,
		).WithContext("timeout", c.HealthCheckTimeout).WithContext("interval", c.HealthCheckInterval))
	}

	// Validate backend timeout
	if c.BackendTimeout <= 0 {
		validationErr.Add(errors.NewInvalidTimeoutError(c.BackendTimeout, "backend timeout"))
	}

	if validationErr.HasErrors() {
		return validationErr
	}

	return nil
}

// Validate method on Config struct
func (c *Config) Validate() error {
	return ValidateConfig(c)
}
