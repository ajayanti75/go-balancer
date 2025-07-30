package metrics

import (
	"fmt"
	"net/http"
)

type PrometheusMetricsProvider struct {
	metrics *Metrics
}

func NewPrometheusMetricsProvider(metrics *Metrics) *PrometheusMetricsProvider {
	return &PrometheusMetricsProvider{metrics: metrics}
}

func (p *PrometheusMetricsProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	snapshot := p.metrics.GetSnapshot()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	fmt.Fprintf(w, "# HELP go_balancer_requests_total Total number of requests processed\n")
	fmt.Fprintf(w, "# TYPE go_balancer_requests_total counter\n")
	fmt.Fprintf(w, "go_balancer_requests_total %d\n", snapshot.TotalRequests)

	fmt.Fprintf(w, "# HELP go_balancer_requests_success_total Total number of successful requests\n")
	fmt.Fprintf(w, "# TYPE go_balancer_requests_success_total counter\n")
	fmt.Fprintf(w, "go_balancer_requests_success_total %d\n", snapshot.SuccessfulRequests)

	fmt.Fprintf(w, "# HELP go_balancer_requests_failed_total Total number of failed requests\n")
	fmt.Fprintf(w, "# TYPE go_balancer_requests_failed_total counter\n")
	fmt.Fprintf(w, "go_balancer_requests_failed_total %d\n", snapshot.FailedRequests)

	fmt.Fprintf(w, "# HELP go_balancer_backend_healthy Current health status (1=healthy, 0=unhealthy)\n")
	fmt.Fprintf(w, "# TYPE go_balancer_backend_healthy gauge\n")
	fmt.Fprintf(w, "go_balancer_backend_healthy{state=\"healthy\"} %d\n", snapshot.HealthyBackends)
	fmt.Fprintf(w, "go_balancer_backend_healthy{state=\"total\"} %d\n", snapshot.TotalBackends)

	// For backend-specific metrics, we need to access the maps directly (with lock)
	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	fmt.Fprintf(w, "# HELP go_balancer_backend_requests_total Total requests sent to backend\n")
	fmt.Fprintf(w, "# TYPE go_balancer_backend_requests_total counter\n")
	for backend, count := range p.metrics.backendRequests {
		fmt.Fprintf(w, "go_balancer_backend_requests_total{backend=\"%s\"} %d\n", backend, count)
	}

	fmt.Fprintf(w, "# HELP go_balancer_backend_failures_total Total failures from backend\n")
	fmt.Fprintf(w, "# TYPE go_balancer_backend_failures_total counter\n")
	for backend, count := range p.metrics.backendFailures {
		fmt.Fprintf(w, "go_balancer_backend_failures_total{backend=\"%s\"} %d\n", backend, count)
	}
}
