package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// TestBackendServer represents our test backend server for load balancer testing
type TestBackendServer struct {
	port int
}

// ServeHTTP implements the http.Handler interface for our test backend
func (bs *TestBackendServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request from %s: %s %s", r.RemoteAddr, r.Method, r.URL.Path)
	log.Printf("Host: %s", r.Host)
	log.Printf("User-Agent: %s", r.Header.Get("User-Agent"))

	// Log custom headers to demonstrate header forwarding capability
	if testHeader := r.Header.Get("X-Test-Header"); testHeader != "" {
		log.Printf("Custom header X-Test-Header: %s", testHeader)
	}

	response := fmt.Sprintf("Hello From Backend Server on port %d", bs.port)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))

	log.Printf("Replied with a hello message")
}

func main() {
	numServers := flag.Int("num", 3, "Number of backend servers to run")
	portsStr := flag.String("ports", "8080,8081,8082", "Comma-separated list of ports to listen on")
	flag.Parse()

	portStrings := strings.Split(*portsStr, ",")
	if len(portStrings) != *numServers {
		log.Fatalf("Number of ports (%d) must match number of servers (%d)", len(portStrings), *numServers)
	}

	var ports []int
	for _, ps := range portStrings {
		p, err := strconv.Atoi(strings.TrimSpace(ps))
		if err != nil {
			log.Fatalf("Invalid port: %v", err)
		}
		ports = append(ports, p)
	}

	for i := 0; i < *numServers; i++ {
		port := ports[i]
		backend := &TestBackendServer{port: port}
		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: backend,
		}
		go func(s *http.Server, p int) {
			log.Printf("Test backend server starting on port %d", p)
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Test server on port %d failed: %v", p, err)
			}
		}(server, port)
	}

	select {} // Block forever to keep all servers running
}
