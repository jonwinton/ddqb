package metric_test

import (
	"testing"

	"github.com/jonwinton/ddqb/metric"
)

func TestFilterBuilder(t *testing.T) {
	tests := []struct {
		name     string
		build    func() (string, error)
		expected string
		wantErr  bool
	}{
		{
			name: "equal filter",
			build: func() (string, error) {
				return metric.NewFilterBuilder("host").Equal("web-1").Build()
			},
			expected: "host:web-1",
			wantErr:  false,
		},
		{
			name: "not equal filter",
			build: func() (string, error) {
				return metric.NewFilterBuilder("host").NotEqual("web-1").Build()
			},
			expected: "host!:web-1",
			wantErr:  false,
		},
		{
			name: "greater than filter",
			build: func() (string, error) {
				return metric.NewFilterBuilder("cpu").GreaterThan("80").Build()
			},
			expected: "cpu>80",
			wantErr:  false,
		},
		{
			name: "less than filter",
			build: func() (string, error) {
				return metric.NewFilterBuilder("cpu").LessThan("80").Build()
			},
			expected: "cpu<80",
			wantErr:  false,
		},
		{
			name: "regex filter",
			build: func() (string, error) {
				return metric.NewFilterBuilder("host").Regex("web-.*").Build()
			},
			expected: "host:~web-.*",
			wantErr:  false,
		},
		{
			name: "in filter",
			build: func() (string, error) {
				return metric.NewFilterBuilder("host").In("web-1", "web-2", "web-3").Build()
			},
			expected: "host IN [\"web-1\", \"web-2\", \"web-3\"]",
			wantErr:  false,
		},
		{
			name: "not in filter",
			build: func() (string, error) {
				return metric.NewFilterBuilder("host").NotIn("db-1", "db-2").Build()
			},
			expected: "host NOT IN [\"db-1\", \"db-2\"]",
			wantErr:  false,
		},
		{
			name: "error - empty key",
			build: func() (string, error) {
				return metric.NewFilterBuilder("").Equal("value").Build()
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