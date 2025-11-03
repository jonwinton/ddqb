package metric_test

import (
	"testing"

	"github.com/jonwinton/ddqb"
)

// TestDatadogQueryFormat verifies that queries are formatted
// according to standard Datadog query format (without spaces between components)
func TestDatadogQueryFormat(t *testing.T) {
	tests := []struct {
		name     string
		build    func() (string, error)
		expected string
	}{
		{
			name: "Simple metric with wildcard filter",
			build: func() (string, error) {
				return ddqb.Metric().
					Metric("system.cpu.idle").
					Build()
			},
			expected: "system.cpu.idle{*}",
		},
		{
			name: "Aggregated metric with filter",
			build: func() (string, error) {
				return ddqb.Metric().
					Aggregator("avg").
					Metric("system.cpu.idle").
					Filter(ddqb.Filter("host").Equal("web-01")).
					Build()
			},
			expected: "avg:system.cpu.idle{host:web-01}",
		},
		{
			name: "Aggregated metric with time window",
			build: func() (string, error) {
				return ddqb.Metric().
					Aggregator("avg").
					TimeWindow("5m").
					Metric("system.cpu.idle").
					Build()
			},
			expected: "avg(5m):system.cpu.idle{*}",
		},
		{
			name: "Complex query with grouping",
			build: func() (string, error) {
				return ddqb.Metric().
					Aggregator("sum").
					Metric("system.cpu.idle").
					Filter(ddqb.Filter("environment").Equal("production")).
					GroupBy("host").
					Build()
			},
			expected: "sum:system.cpu.idle{environment:production} by {host}",
		},
		{
			name: "Query with function",
			build: func() (string, error) {
				return ddqb.Metric().
					Metric("system.cpu.idle").
					Filter(ddqb.Filter("host").Equal("web-01")).
					ApplyFunction(ddqb.Function("rollup").WithArg("60")).
					Build()
			},
			expected: "system.cpu.idle{host:web-01}.rollup(60)",
		},
		{
			name: "Full complex query",
			build: func() (string, error) {
				return ddqb.Metric().
					Aggregator("avg").
					TimeWindow("5m").
					Metric("system.cpu.idle").
					Filter(ddqb.Filter("environment").Equal("production")).
					Filter(ddqb.Filter("host").Equal("web-01")).
					GroupBy("host").
					ApplyFunction(ddqb.Function("fill").WithArg("null")).
					ApplyFunction(ddqb.Function("rollup").WithArgs("60", "avg")).
					Build()
			},
			expected: "avg(5m):system.cpu.idle{environment:production, host:web-01} by {host}.fill(null).rollup(60, avg)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.build()
			if err != nil {
				t.Fatalf("Failed to build query: %v", err)
			}

			if result != tc.expected {
				t.Errorf("Query format incorrect\nExpected: %s\nActual:   %s", tc.expected, result)
			}
		})
	}
}
