package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

// BackendServer represents our test backend server
type BackendServer struct {
	port int
}

// ServeHTTP implements the http.Handler interface for our backend
func (bs *BackendServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request
	log.Printf("Received request from %s: %s %s",
		r.RemoteAddr, r.Method, r.URL.Path)
	log.Printf("Host: %s", r.Host)
	log.Printf("User-Agent: %s", r.Header.Get("User-Agent"))

	// Send response
	response := fmt.Sprintf("Hello From Backend Server on port %d", bs.port)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))

	log.Printf("Replied with a hello message")
}

// Simple backend server for testing the load balancer
func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	// Create backend server instance
	backend := &BackendServer{port: *port}

	// Create HTTP server (consistent with our load balancer approach)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: backend,
	}

	log.Printf("Backend server starting on port %d", *port)

	if err := server.ListenAndServe(); err != nil {
		if err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}
}
