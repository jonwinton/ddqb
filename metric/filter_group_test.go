package metric

import (
	"testing"
)

func TestFilterGroupBuilder_AND(t *testing.T) {
	tests := []struct {
		name     string
		build    func() (string, error)
		expected string
		wantErr  bool
	}{
		{
			name: "simple AND group",
			build: func() (string, error) {
				group := NewFilterGroupBuilder()
				group.And(NewFilterBuilder("env").Equal("prod"))
				group.And(NewFilterBuilder("host").Equal("web-1"))
				return group.Build()
			},
			expected: "(env:prod AND host:web-1)",
			wantErr:  false,
		},
		{
			name: "AND group with three filters",
			build: func() (string, error) {
				group := NewFilterGroupBuilder()
				group.And(NewFilterBuilder("env").Equal("prod"))
				group.And(NewFilterBuilder("host").Equal("web-1"))
				group.And(NewFilterBuilder("region").Equal("us-east-1"))
				return group.Build()
			},
			expected: "(env:prod AND host:web-1 AND region:us-east-1)",
			wantErr:  false,
		},
		{
			name: "single filter in group",
			build: func() (string, error) {
				group := NewFilterGroupBuilder()
				group.And(NewFilterBuilder("env").Equal("prod"))
				return group.Build()
			},
			expected: "env:prod",
			wantErr:  false,
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

func TestFilterGroupBuilder_OR(t *testing.T) {
	tests := []struct {
		name     string
		build    func() (string, error)
		expected string
		wantErr  bool
	}{
		{
			name: "simple OR group",
			build: func() (string, error) {
				group := NewFilterGroupBuilder()
				group.Or(NewFilterBuilder("env").Equal("prod"))
				group.Or(NewFilterBuilder("env").Equal("staging"))
				return group.Build()
			},
			expected: "(env:prod OR env:staging)",
			wantErr:  false,
		},
		{
			name: "OR group with three filters",
			build: func() (string, error) {
				group := NewFilterGroupBuilder()
				group.Or(NewFilterBuilder("host").Equal("web-1"))
				group.Or(NewFilterBuilder("host").Equal("web-2"))
				group.Or(NewFilterBuilder("host").Equal("web-3"))
				return group.Build()
			},
			expected: "(host:web-1 OR host:web-2 OR host:web-3)",
			wantErr:  false,
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

func TestFilterGroupBuilder_Not(t *testing.T) {
	tests := []struct {
		name     string
		build    func() (string, error)
		expected string
		wantErr  bool
	}{
		{
			name: "negated AND group",
			build: func() (string, error) {
				group := NewFilterGroupBuilder()
				group.And(NewFilterBuilder("env").Equal("prod"))
				group.And(NewFilterBuilder("host").Equal("web-1"))
				group.Not()
				return group.Build()
			},
			expected: "NOT (env:prod AND host:web-1)",
			wantErr:  false,
		},
		{
			name: "negated OR group",
			build: func() (string, error) {
				group := NewFilterGroupBuilder()
				group.Or(NewFilterBuilder("env").Equal("prod"))
				group.Or(NewFilterBuilder("env").Equal("staging"))
				group.Not()
				return group.Build()
			},
			expected: "NOT (env:prod OR env:staging)",
			wantErr:  false,
		},
		{
			name: "negated single filter",
			build: func() (string, error) {
				group := NewFilterGroupBuilder()
				group.And(NewFilterBuilder("env").Equal("prod"))
				group.Not()
				return group.Build()
			},
			expected: "NOT env:prod",
			wantErr:  false,
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

func TestFilterGroupBuilder_Nested(t *testing.T) {
	tests := []struct {
		name     string
		build    func() (string, error)
		expected string
		wantErr  bool
	}{
		{
			name: "nested OR in AND",
			build: func() (string, error) {
				outerGroup := NewFilterGroupBuilder()
				outerGroup.And(NewFilterBuilder("env").Equal("prod"))

				innerGroup := NewFilterGroupBuilder()
				innerGroup.Or(NewFilterBuilder("host").Equal("web-1"))
				innerGroup.Or(NewFilterBuilder("host").Equal("web-2"))

				outerGroup.And(innerGroup)
				return outerGroup.Build()
			},
			expected: "(env:prod AND (host:web-1 OR host:web-2))",
			wantErr:  false,
		},
		{
			name: "nested AND in OR",
			build: func() (string, error) {
				outerGroup := NewFilterGroupBuilder()
				outerGroup.Or(NewFilterBuilder("env").Equal("prod"))

				innerGroup := NewFilterGroupBuilder()
				innerGroup.And(NewFilterBuilder("host").Equal("web-1"))
				innerGroup.And(NewFilterBuilder("region").Equal("us-east-1"))

				outerGroup.Or(innerGroup)
				return outerGroup.Build()
			},
			expected: "(env:prod OR (host:web-1 AND region:us-east-1))",
			wantErr:  false,
		},
		{
			name: "complex nested groups",
			build: func() (string, error) {
				outerGroup := NewFilterGroupBuilder()

				envGroup := NewFilterGroupBuilder()
				envGroup.Or(NewFilterBuilder("env").Equal("prod"))
				envGroup.Or(NewFilterBuilder("env").Equal("staging"))

				hostGroup := NewFilterGroupBuilder()
				hostGroup.Or(NewFilterBuilder("host").Equal("web-1"))
				hostGroup.Or(NewFilterBuilder("host").Equal("api-1"))

				outerGroup.And(envGroup)
				outerGroup.And(hostGroup)
				return outerGroup.Build()
			},
			expected: "((env:prod OR env:staging) AND (host:web-1 OR host:api-1))",
			wantErr:  false,
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

func TestFilterGroupBuilder_EmptyGroup(t *testing.T) {
	group := NewFilterGroupBuilder()
	_, err := group.Build()
	if err == nil {
		t.Error("Build() should return error for empty group")
	}
}

func TestFilterGroupBuilder_WithQueryBuilder(t *testing.T) {
	tests := []struct {
		name     string
		build    func() (string, error)
		expected string
		wantErr  bool
	}{
		{
			name: "metric query with OR group",
			build: func() (string, error) {
				group := NewFilterGroupBuilder()
				group.Or(NewFilterBuilder("env").Equal("prod"))
				group.Or(NewFilterBuilder("env").Equal("staging"))

				builder := NewMetricQueryBuilder()
				builder.Metric("system.cpu.idle")
				builder.Filter(group)
				return builder.Build()
			},
			expected: "system.cpu.idle{(env:prod OR env:staging)}",
			wantErr:  false,
		},
		{
			name: "metric query with multiple groups",
			build: func() (string, error) {
				envGroup := NewFilterGroupBuilder()
				envGroup.Or(NewFilterBuilder("env").Equal("prod"))
				envGroup.Or(NewFilterBuilder("env").Equal("staging"))

				hostFilter := NewFilterBuilder("host").Equal("web-1")

				builder := NewMetricQueryBuilder()
				builder.Metric("system.cpu.idle")
				builder.Filter(envGroup)
				builder.Filter(hostFilter)
				return builder.Build()
			},
			expected: "system.cpu.idle{((env:prod OR env:staging) AND host:web-1)}",
			wantErr:  false,
		},
		{
			name: "metric query with negated group",
			build: func() (string, error) {
				group := NewFilterGroupBuilder()
				group.And(NewFilterBuilder("env").Equal("prod"))
				group.And(NewFilterBuilder("host").Equal("web-1"))
				group.Not()

				builder := NewMetricQueryBuilder()
				builder.Metric("system.cpu.idle")
				builder.Filter(group)
				return builder.Build()
			},
			expected: "system.cpu.idle{NOT (env:prod AND host:web-1)}",
			wantErr:  false,
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
