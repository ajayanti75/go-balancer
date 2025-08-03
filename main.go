package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go-balancer/internal/balancer"
	"go-balancer/internal/config"
	"go-balancer/internal/errors"
)

func main() {
	// Parse command line flags
	var (
		port           = flag.Int("port", 8000, "Port to listen on")
		backends       = flag.String("backends", "http://localhost:8080,http://localhost:8081,http://localhost:8082", "Comma-separated list of backend servers")
		healthPath     = flag.String("health-path", "/", "Path to use for health checking")
		healthInterval = flag.Int("health-interval", 10, "Health check interval in seconds")
		healthTimeout  = flag.Int("health-timeout", 2, "Health check timeout in seconds")
		backendTimeout = flag.Int("backend-timeout", 30, "Timeout for backend requests in seconds")
	)
	flag.Parse()

	// Parse backends string into slice
	backendList := strings.Split(*backends, ",")
	for i, backend := range backendList {
		backendList[i] = strings.TrimSpace(backend)
	}

	// Create config
	cfg := &config.Config{
		Port:                *port,
		Backends:            backendList,
		HealthCheckPath:     *healthPath,
		HealthCheckInterval: time.Duration(*healthInterval) * time.Second,
		HealthCheckTimeout:  time.Duration(*healthTimeout) * time.Second,
		BackendTimeout:      time.Duration(*backendTimeout) * time.Second,
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		// Handle structured validation errors
		if validationErr, ok := err.(*config.ValidationError); ok {
			log.Printf("Configuration validation failed with %d errors:", len(validationErr.Errors))
			for i, vErr := range validationErr.Errors {
				log.Printf("  %d. %v", i+1, vErr)
			}
		} else if lbErr, ok := err.(*errors.LoadBalancerError); ok {
			log.Printf("Configuration validation failed: %v", lbErr)
		} else {
			log.Printf("Configuration validation failed: %v", err)
		}
		return
	}

	// Create load balancer
	lb, err := balancer.NewLoadBalancer(cfg)
	if err != nil {
		// Handle structured load balancer creation errors
		if lbErr, ok := err.(*errors.LoadBalancerError); ok {
			log.Printf("Failed to create load balancer: %v", lbErr)
			if errors.IsConfigurationError(lbErr) {
				log.Printf("This is a configuration error. Please check your backend URLs and settings.")
			} else if errors.IsBackendError(lbErr) {
				log.Printf("This is a backend error. Please check that your backend servers are configured correctly.")
			}
		} else {
			log.Printf("Failed to create load balancer: %v", err)
		}
		return
	}

	// Create HTTP server with both load balancer and metrics
	mux := http.NewServeMux()

	// Handle metrics endpoint
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		lb.GetMetricsProvider().ServeHTTP(w, r)
	})

	// Handle all other requests with the load balancer
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		lb.ServeHTTP(w, r)
	})

	loadBalancerServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	log.Printf("Load balancer starting on port %d", cfg.Port)
	log.Printf("Forwarding requests to backends: %v", cfg.Backends)
	log.Printf("Health checks: every %s, timeout %s, path %s",
		cfg.HealthCheckInterval, cfg.HealthCheckTimeout, cfg.HealthCheckPath)
	log.Printf("Backend request timeout: %s", cfg.BackendTimeout)

	// Start the load balancer server
	if err := loadBalancerServer.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}
}
