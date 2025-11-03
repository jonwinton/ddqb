package metric_test

import (
	"testing"

	"github.com/jonwinton/ddqb/metric"
)

// TestFilterInputValidation adds tests specifically to cover validation
// error cases in the filter builder
func TestFilterInputValidation(t *testing.T) {
	tests := []struct {
		name        string
		filterBuild func() metric.FilterBuilder
		operation   string
	}{
		{
			name: "Equal validation",
			filterBuild: func() metric.FilterBuilder {
				return metric.NewFilterBuilder("host").Equal("")
			},
			operation: "Equal",
		},
		{
			name: "NotEqual validation",
			filterBuild: func() metric.FilterBuilder {
				return metric.NewFilterBuilder("host").NotEqual("")
			},
			operation: "NotEqual",
		},
	}

	// This test just makes sure that the filter builder methods return
	// a usable builder even with empty values
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.filterBuild()
			if builder == nil {
				t.Errorf("%s operation returned nil builder", tt.operation)
			}
		})
	}
}
