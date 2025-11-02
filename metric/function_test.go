package metric_test

import (
	"testing"

	"github.com/jonwinton/ddqb/metric"
)

func TestFunctionBuilder(t *testing.T) {
	tests := []struct {
		name     string
		build    func() (string, error)
		expected string
		wantErr  bool
	}{
		{
			name: "function with no args",
			build: func() (string, error) {
				return metric.NewFunctionBuilder("fill").Build()
			},
			expected: ".fill()",
			wantErr:  false,
		},
		{
			name: "function with single arg",
			build: func() (string, error) {
				return metric.NewFunctionBuilder("fill").WithArg("0").Build()
			},
			expected: ".fill(0)",
			wantErr:  false,
		},
		{
			name: "function with multiple args",
			build: func() (string, error) {
				return metric.NewFunctionBuilder("rollup").WithArgs("60", "sum").Build()
			},
			expected: ".rollup(60, sum)",
			wantErr:  false,
		},
		{
			name: "function with args added separately",
			build: func() (string, error) {
				return metric.NewFunctionBuilder("rollup").
					WithArg("60").
					WithArg("sum").
					Build()
			},
			expected: ".rollup(60, sum)",
			wantErr:  false,
		},
		{
			name: "error - empty function name",
			build: func() (string, error) {
				return metric.NewFunctionBuilder("").WithArg("0").Build()
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
