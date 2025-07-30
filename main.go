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

	// Create load balancer
	lb, err := balancer.NewLoadBalancer(cfg)
	if err != nil {
		log.Fatalf("Failed to create load balancer: %v", err)
	}

	// Create HTTP server for the load balancer
	loadBalancerServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: lb,
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
