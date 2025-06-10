package metric_test

import (
	"testing"

	"github.com/jonwinton/ddqb"
	"github.com/jonwinton/ddqb/metric"
)

func TestMetricQueryBuilder(t *testing.T) {
	tests := []struct {
		name     string
		build    func() (string, error)
		expected string
		wantErr  bool
	}{
		{
			name: "simple metric query",
			build: func() (string, error) {
				return metric.NewMetricQueryBuilder().
					Metric("system.cpu.idle").
					Build()
			},
			expected: "system.cpu.idle{*}",
			wantErr:  false,
		},
		{
			name: "metric query with aggregator",
			build: func() (string, error) {
				return metric.NewMetricQueryBuilder().
					Aggregator("avg").
					Metric("system.cpu.idle").
					Build()
			},
			expected: "avg:system.cpu.idle{*}",
			wantErr:  false,
		},
		{
			name: "metric query with aggregator and time window",
			build: func() (string, error) {
				return metric.NewMetricQueryBuilder().
					Aggregator("avg").
					TimeWindow("5m").
					Metric("system.cpu.idle").
					Build()
			},
			expected: "avg(5m):system.cpu.idle{*}",
			wantErr:  false,
		},
		{
			name: "metric query with filter",
			build: func() (string, error) {
				return metric.NewMetricQueryBuilder().
					Metric("system.cpu.idle").
					Filter(metric.NewFilterBuilder("host").Equal("web-1")).
					Build()
			},
			expected: "system.cpu.idle{host:web-1}",
			wantErr:  false,
		},
		{
			name: "metric query with multiple filters",
			build: func() (string, error) {
				return metric.NewMetricQueryBuilder().
					Metric("system.cpu.idle").
					Filter(metric.NewFilterBuilder("host").Equal("web-1")).
					Filter(metric.NewFilterBuilder("env").Equal("prod")).
					Build()
			},
			expected: "system.cpu.idle{host:web-1, env:prod}",
			wantErr:  false,
		},
		{
			name: "metric query with group by",
			build: func() (string, error) {
				return metric.NewMetricQueryBuilder().
					Metric("system.cpu.idle").
					GroupBy("host", "env").
					Build()
			},
			expected: "system.cpu.idle{*} by {host, env}",
			wantErr:  false,
		},
		{
			name: "metric query with function",
			build: func() (string, error) {
				return metric.NewMetricQueryBuilder().
					Metric("system.cpu.idle").
					ApplyFunction(metric.NewFunctionBuilder("fill").WithArg("0")).
					Build()
			},
			expected: "system.cpu.idle{*}.fill(0)",
			wantErr:  false,
		},
		{
			name: "complex metric query",
			build: func() (string, error) {
				return metric.NewMetricQueryBuilder().
					Aggregator("avg").
					TimeWindow("5m").
					Metric("system.cpu.idle").
					Filter(metric.NewFilterBuilder("host").Equal("web-1")).
					Filter(metric.NewFilterBuilder("env").Equal("prod")).
					GroupBy("host").
					ApplyFunction(metric.NewFunctionBuilder("fill").WithArg("0")).
					ApplyFunction(metric.NewFunctionBuilder("rollup").WithArg("60")).
					Build()
			},
			expected: "avg(5m):system.cpu.idle{host:web-1, env:prod} by {host}.fill(0).rollup(60)",
			wantErr:  false,
		},
		{
			name: "error - missing metric",
			build: func() (string, error) {
				return metric.NewMetricQueryBuilder().
					Aggregator("avg").
					Build()
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.build()

			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.expected {
				t.Errorf("Build() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Test the top-level API for convenience
func TestTopLevelAPI(t *testing.T) {
	query, err := ddqb.Metric().
		Aggregator("avg").
		TimeWindow("5m").
		Metric("system.cpu.idle").
		Filter(ddqb.Filter("host").Equal("web-1")).
		GroupBy("host").
		ApplyFunction(ddqb.Function("fill").WithArg("0")).
		Build()

	if err != nil {
		t.Errorf("Build() returned an error: %v", err)
	}

	expected := "avg(5m):system.cpu.idle{host:web-1} by {host}.fill(0)"
	if query != expected {
		t.Errorf("Build() = %q, want %q", query, expected)
	}
}