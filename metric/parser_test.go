package metric_test

import (
	"testing"

	"github.com/jonwinton/ddqb"
	"github.com/jonwinton/ddqb/metric"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		build       func(metric.MetricQueryBuilder) metric.MetricQueryBuilder
		expected    string
		wantErr     bool
	}{
		{
			name:        "simple metric query",
			queryString: "system.cpu.idle{*}",
			build:       func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder { return b },
			expected:    "system.cpu.idle{*}",
			wantErr:     false,
		},
		{
			name:        "metric query with aggregator",
			queryString: "avg:system.cpu.idle{*}",
			build:       func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder { return b },
			expected:    "avg:system.cpu.idle{*}",
			wantErr:     false,
		},
		{
			name:        "metric query with aggregator and time window",
			queryString: "avg(5m):system.cpu.idle{*}",
			build:       func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder { return b },
			expected:    "avg(5m):system.cpu.idle{*}",
			wantErr:     false,
		},
		{
			name:        "metric query with filter",
			queryString: "system.cpu.idle{host:web-1}",
			build:       func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder { return b },
			expected:    "system.cpu.idle{host:web-1}",
			wantErr:     false,
		},
		{
			name:        "metric query with multiple filters",
			queryString: "system.cpu.idle{host:web-1,env:prod}",
			build:       func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder { return b },
			expected:    "system.cpu.idle{host:web-1, env:prod}",
			wantErr:     false,
		},
		{
			name:        "metric query with group by",
			queryString: "system.cpu.idle{*} by {host}",
			build:       func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder { return b },
			expected:    "system.cpu.idle{*} by {host}",
			wantErr:     false,
		},
		{
			name:        "metric query with function",
			queryString: "system.cpu.idle{*}.fill(0)",
			build:       func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder { return b },
			expected:    "system.cpu.idle{*}.fill(0)",
			wantErr:     false,
		},
		{
			name:        "complex metric query",
			queryString: "avg(5m):system.cpu.idle{host:web-1,env:prod} by {host}.fill(0).rollup(60,avg)",
			build:       func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder { return b },
			expected:    "avg(5m):system.cpu.idle{host:web-1, env:prod} by {host}.fill(0).rollup(60, avg)",
			wantErr:     false,
		},
		{
			name:        "query with regex filter",
			queryString: "system.cpu.idle{host:~web-.*}",
			build:       func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder { return b },
			expected:    "system.cpu.idle{host:~web-.*}",
			wantErr:     false,
		},
		{
			name:        "query with IN filter",
			queryString: "system.cpu.idle{host IN (web-1,web-2,web-3)}",
			build:       func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder { return b },
			expected:    "system.cpu.idle{host IN (web-1,web-2,web-3)}",
			wantErr:     false,
		},
		{
			name:        "query with NOT IN filter",
			queryString: "system.cpu.idle{host NOT IN (db-1,db-2)}",
			build:       func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder { return b },
			expected:    "system.cpu.idle{host NOT IN (db-1,db-2)}",
			wantErr:     false,
		},
		{
			name:        "query with not equal filter",
			queryString: "system.cpu.idle{!host:web-1}",
			build:       func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder { return b },
			expected:    "system.cpu.idle{host!:web-1}",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := metric.ParseQuery(tt.queryString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Apply any modifications
			builder = tt.build(builder)

			// Build the query
			result, err := builder.Build()
			if err != nil {
				t.Errorf("Build() error = %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Build() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseQueryModify(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		modify      func(metric.MetricQueryBuilder) metric.MetricQueryBuilder
		expected    string
	}{
		{
			name:        "parse and change time window",
			queryString: "avg(5m):system.cpu.idle{*}",
			modify: func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder {
				return b.TimeWindow("10m")
			},
			expected: "avg(10m):system.cpu.idle{*}",
		},
		{
			name:        "parse and add filter",
			queryString: "system.cpu.idle{host:web-1}",
			modify: func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder {
				return b.Filter(ddqb.Filter("env").Equal("prod"))
			},
			expected: "system.cpu.idle{host:web-1, env:prod}",
		},
		{
			name:        "parse and change aggregator",
			queryString: "avg:system.cpu.idle{*}",
			modify: func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder {
				return b.Aggregator("sum")
			},
			expected: "sum:system.cpu.idle{*}",
		},
		{
			name:        "parse and add function",
			queryString: "system.cpu.idle{*}",
			modify: func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder {
				return b.ApplyFunction(ddqb.Function("fill").WithArg("0"))
			},
			expected: "system.cpu.idle{*}.fill(0)",
		},
		{
			name:        "parse and modify multiple components",
			queryString: "avg(5m):system.cpu.idle{host:web-1}",
			modify: func(b metric.MetricQueryBuilder) metric.MetricQueryBuilder {
				return b.
					TimeWindow("10m").
					Filter(ddqb.Filter("env").Equal("prod")).
					GroupBy("host").
					ApplyFunction(ddqb.Function("fill").WithArg("0"))
			},
			expected: "avg(10m):system.cpu.idle{host:web-1, env:prod} by {host}.fill(0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := metric.ParseQuery(tt.queryString)
			if err != nil {
				t.Fatalf("ParseQuery() error = %v", err)
			}

			// Modify the builder
			builder = tt.modify(builder)

			// Build the query
			result, err := builder.Build()
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("Build() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseQueryErrors(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		wantErr     bool
	}{
		{
			name:        "invalid query - missing metric",
			queryString: "{host:web-1}",
			wantErr:     true,
		},
		{
			name:        "invalid query - malformed",
			queryString: "avg:system.cpu.idle{host:",
			wantErr:     true,
		},
		{
			name:        "unsupported - aggregator function wrapper",
			queryString: "moving_rollup(sum:metric{*}, 60)",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := metric.ParseQuery(tt.queryString)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFromQueryTopLevel(t *testing.T) {
	// Test the top-level API
	builder, err := ddqb.FromQuery("avg(5m):system.cpu.idle{host:web-1} by {host}")
	if err != nil {
		t.Fatalf("FromQuery() error = %v", err)
	}

	result, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	expected := "avg(5m):system.cpu.idle{host:web-1} by {host}"
	if result != expected {
		t.Errorf("Build() = %q, want %q", result, expected)
	}
}

