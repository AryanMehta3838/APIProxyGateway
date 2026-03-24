package telemetry

import (
	"strings"
	"testing"
)

func TestCollectorRenderIncludesRequestRateLimitAndDurationMetrics(t *testing.T) {
	t.Parallel()

	collector := New()
	collector.Observe("echo", "GET", 200, 0.012)
	collector.Observe("echo", "GET", 429, 0.003)

	metrics := collector.Render()

	for _, want := range []string{
		`apiproxy_http_requests_total{route="echo",method="GET",status="200"} 1`,
		`apiproxy_http_requests_total{route="echo",method="GET",status="429"} 1`,
		`apiproxy_rate_limit_denied_total{route="echo"} 1`,
		`apiproxy_http_request_duration_seconds_count{route="echo",method="GET"} 2`,
		`apiproxy_http_request_duration_seconds_sum{route="echo",method="GET"} 0.015`,
	} {
		if !strings.Contains(metrics, want) {
			t.Fatalf("expected metrics output to contain %q, got:\n%s", want, metrics)
		}
	}
}
