package metric_test

import (
	"testing"

	"github.com/jonwinton/ddqb/metric"
)

// TestFilterBuilderErrors tests error conditions for the filter builder.
func TestFilterBuilderErrors(t *testing.T) {
	tests := []struct {
		name        string
		filterBuild func() (string, error)
		expectError bool
	}{
		{
			name: "empty key",
			filterBuild: func() (string, error) {
				// Create builder with empty key
				builder := metric.NewFilterBuilder("")
				builder.Equal("value")
				return builder.Build()
			},
			expectError: true,
		},
		{
			name: "in filter with empty values array",
			filterBuild: func() (string, error) {
				builder := metric.NewFilterBuilder("host")
				builder.In() // No values
				return builder.Build()
			},
			expectError: true,
		},
		{
			name: "not in filter with empty values array",
			filterBuild: func() (string, error) {
				builder := metric.NewFilterBuilder("host")
				builder.NotIn() // No values
				return builder.Build()
			},
			expectError: true,
		},
		{
			name: "default filter operation",
			filterBuild: func() (string, error) {
				// Create a builder but don't set any operation
				builder := metric.NewFilterBuilder("host")
				// Force a direct Build call without setting an operation
				return builder.Build()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.filterBuild()

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s but got nil", tt.name)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.name, err)
			}
		})
	}
}
