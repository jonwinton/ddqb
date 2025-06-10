package metric_test

import (
	"testing"

	"github.com/jonwinton/ddqb/metric"
)

// TestMetricBuilderErrors tests error cases for the metric builder
func TestMetricBuilderErrors(t *testing.T) {
	tests := []struct {
		name     string
		build    func() (string, error)
		wantErr  bool
	}{
		{
			name: "error - missing metric",
			build: func() (string, error) {
				return metric.NewMetricQueryBuilder().
					Aggregator("avg").
					Build()
			},
			wantErr: true,
		},
		{
			name: "error - filter build failure",
			build: func() (string, error) {
				// Create a filter that will fail on build (empty key)
				emptyKeyFilter := metric.NewFilterBuilder("")
				
				return metric.NewMetricQueryBuilder().
					Metric("system.cpu.idle").
					Filter(emptyKeyFilter).
					Build()
			},
			wantErr: true,
		},
		{
			name: "error - function build failure",
			build: func() (string, error) {
				// Create a function that will fail on build (empty name)
				emptyFunc := metric.NewFunctionBuilder("")
				
				return metric.NewMetricQueryBuilder().
					Metric("system.cpu.idle").
					ApplyFunction(emptyFunc).
					Build()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.build()
			
			if tt.wantErr && err == nil {
				t.Errorf("Expected error for %s but got nil", tt.name)
			}
			
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.name, err)
			}
		})
	}
}