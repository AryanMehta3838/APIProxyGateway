package telemetry

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
)

var durationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5}

type Collector struct {
	mu              sync.Mutex
	requests        map[requestKey]int
	durations       map[durationKey]durationMetric
	rateLimitDenied map[string]int
}

type requestKey struct {
	Route  string
	Method string
	Status int
}

type durationKey struct {
	Route  string
	Method string
}

type durationMetric struct {
	Count   int
	Sum     float64
	Buckets []int
}

func New() *Collector {
	return &Collector{
		requests:        make(map[requestKey]int),
		durations:       make(map[durationKey]durationMetric),
		rateLimitDenied: make(map[string]int),
	}
}

func (c *Collector) Observe(route string, method string, status int, durationSeconds float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if route == "" {
		route = "unknown"
	}

	reqKey := requestKey{Route: route, Method: method, Status: status}
	c.requests[reqKey]++

	dKey := durationKey{Route: route, Method: method}
	current := c.durations[dKey]
	if current.Buckets == nil {
		current.Buckets = make([]int, len(durationBuckets))
	}
	current.Count++
	current.Sum += durationSeconds
	for i, bucket := range durationBuckets {
		if durationSeconds <= bucket {
			current.Buckets[i]++
		}
	}
	c.durations[dKey] = current

	if status == http.StatusTooManyRequests {
		c.rateLimitDenied[route]++
	}
}

func (c *Collector) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(c.Render()))
	})
}

func (c *Collector) Render() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	var b strings.Builder

	b.WriteString("# HELP apiproxy_http_requests_total Total HTTP requests handled.\n")
	b.WriteString("# TYPE apiproxy_http_requests_total counter\n")
	requestLines := make([]string, 0, len(c.requests))
	for key, value := range c.requests {
		requestLines = append(requestLines, fmt.Sprintf(
			`apiproxy_http_requests_total{route="%s",method="%s",status="%d"} %d`,
			key.Route, key.Method, key.Status, value,
		))
	}
	sort.Strings(requestLines)
	for _, line := range requestLines {
		b.WriteString(line)
		b.WriteByte('\n')
	}

	b.WriteString("# HELP apiproxy_http_request_duration_seconds HTTP request duration in seconds.\n")
	b.WriteString("# TYPE apiproxy_http_request_duration_seconds histogram\n")
	durationLines := make([]string, 0, len(c.durations)*(len(durationBuckets)+2))
	for key, metric := range c.durations {
		for i, bucket := range durationBuckets {
			durationLines = append(durationLines, fmt.Sprintf(
				`apiproxy_http_request_duration_seconds_bucket{route="%s",method="%s",le="%g"} %d`,
				key.Route, key.Method, bucket, metric.Buckets[i],
			))
		}
		durationLines = append(durationLines, fmt.Sprintf(
			`apiproxy_http_request_duration_seconds_bucket{route="%s",method="%s",le="+Inf"} %d`,
			key.Route, key.Method, metric.Count,
		))
		durationLines = append(durationLines, fmt.Sprintf(
			`apiproxy_http_request_duration_seconds_sum{route="%s",method="%s"} %g`,
			key.Route, key.Method, metric.Sum,
		))
		durationLines = append(durationLines, fmt.Sprintf(
			`apiproxy_http_request_duration_seconds_count{route="%s",method="%s"} %d`,
			key.Route, key.Method, metric.Count,
		))
	}
	sort.Strings(durationLines)
	for _, line := range durationLines {
		b.WriteString(line)
		b.WriteByte('\n')
	}

	b.WriteString("# HELP apiproxy_rate_limit_denied_total Rate-limited responses.\n")
	b.WriteString("# TYPE apiproxy_rate_limit_denied_total counter\n")
	rateLimitLines := make([]string, 0, len(c.rateLimitDenied))
	for route, value := range c.rateLimitDenied {
		rateLimitLines = append(rateLimitLines, fmt.Sprintf(
			`apiproxy_rate_limit_denied_total{route="%s"} %d`,
			route, value,
		))
	}
	sort.Strings(rateLimitLines)
	for _, line := range rateLimitLines {
		b.WriteString(line)
		b.WriteByte('\n')
	}

	return b.String()
}
