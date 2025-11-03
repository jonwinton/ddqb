package metric_test

import (
	"strings"
	"testing"

	"github.com/jonwinton/ddqb"
	"github.com/jonwinton/ddqb/metric"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name        string
		queryString string
		build       func(metric.QueryBuilder) metric.QueryBuilder
		expected    string
		wantErr     bool
	}{
		{
			name:        "simple metric query",
			queryString: "system.cpu.idle{*}",
			build:       func(b metric.QueryBuilder) metric.QueryBuilder { return b },
			expected:    "system.cpu.idle{*}",
			wantErr:     false,
		},
		{
			name:        "metric query with aggregator",
			queryString: "avg:system.cpu.idle{*}",
			build:       func(b metric.QueryBuilder) metric.QueryBuilder { return b },
			expected:    "avg:system.cpu.idle{*}",
			wantErr:     false,
		},
		{
			name:        "metric query with aggregator and time window",
			queryString: "avg(5m):system.cpu.idle{*}",
			build:       func(b metric.QueryBuilder) metric.QueryBuilder { return b },
			expected:    "avg(5m):system.cpu.idle{*}",
			wantErr:     false,
		},
		{
			name:        "metric query with filter",
			queryString: "system.cpu.idle{host:web-1}",
			build:       func(b metric.QueryBuilder) metric.QueryBuilder { return b },
			expected:    "system.cpu.idle{host:web-1}",
			wantErr:     false,
		},
		{
			name:        "metric query with multiple filters",
			queryString: "system.cpu.idle{host:web-1,env:prod}",
			build:       func(b metric.QueryBuilder) metric.QueryBuilder { return b },
			expected:    "system.cpu.idle{host:web-1, env:prod}",
			wantErr:     false,
		},
		{
			name:        "metric query with group by",
			queryString: "system.cpu.idle{*} by {host}",
			build:       func(b metric.QueryBuilder) metric.QueryBuilder { return b },
			expected:    "system.cpu.idle{*} by {host}",
			wantErr:     false,
		},
		{
			name:        "metric query with function",
			queryString: "system.cpu.idle{*}.fill(0)",
			build:       func(b metric.QueryBuilder) metric.QueryBuilder { return b },
			expected:    "system.cpu.idle{*}.fill(0)",
			wantErr:     false,
		},
		{
			name:        "complex metric query",
			queryString: "avg(5m):system.cpu.idle{host:web-1,env:prod} by {host}.fill(0).rollup(60,avg)",
			build:       func(b metric.QueryBuilder) metric.QueryBuilder { return b },
			expected:    "avg(5m):system.cpu.idle{host:web-1, env:prod} by {host}.fill(0).rollup(60, avg)",
			wantErr:     false,
		},
		{
			name:        "query with IN filter",
			queryString: "system.cpu.idle{host IN (web-1,web-2,web-3)}",
			build:       func(b metric.QueryBuilder) metric.QueryBuilder { return b },
			expected:    "system.cpu.idle{host IN (web-1,web-2,web-3)}",
			wantErr:     false,
		},
		{
			name:        "query with NOT IN filter",
			queryString: "system.cpu.idle{host NOT IN (db-1,db-2)}",
			build:       func(b metric.QueryBuilder) metric.QueryBuilder { return b },
			expected:    "system.cpu.idle{host NOT IN (db-1,db-2)}",
			wantErr:     false,
		},
		{
			name:        "query with not equal filter",
			queryString: "system.cpu.idle{!host:web-1}",
			build:       func(b metric.QueryBuilder) metric.QueryBuilder { return b },
			expected:    "system.cpu.idle{!host:web-1}",
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
		modify      func(metric.QueryBuilder) metric.QueryBuilder
		expected    string
	}{
		{
			name:        "parse and change time window",
			queryString: "avg(5m):system.cpu.idle{*}",
			modify: func(b metric.QueryBuilder) metric.QueryBuilder {
				return b.TimeWindow("10m")
			},
			expected: "avg(10m):system.cpu.idle{*}",
		},
		{
			name:        "parse and add filter",
			queryString: "system.cpu.idle{host:web-1}",
			modify: func(b metric.QueryBuilder) metric.QueryBuilder {
				return b.Filter(ddqb.Filter("env").Equal("prod"))
			},
			expected: "system.cpu.idle{host:web-1, env:prod}",
		},
		{
			name:        "parse and change aggregator",
			queryString: "avg:system.cpu.idle{*}",
			modify: func(b metric.QueryBuilder) metric.QueryBuilder {
				return b.Aggregator("sum")
			},
			expected: "sum:system.cpu.idle{*}",
		},
		{
			name:        "parse and add function",
			queryString: "system.cpu.idle{*}",
			modify: func(b metric.QueryBuilder) metric.QueryBuilder {
				return b.ApplyFunction(ddqb.Function("fill").WithArg("0"))
			},
			expected: "system.cpu.idle{*}.fill(0)",
		},
		{
			name:        "parse and modify multiple components",
			queryString: "avg(5m):system.cpu.idle{host:web-1}",
			modify: func(b metric.QueryBuilder) metric.QueryBuilder {
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
			name:        "aggregator function wrapper passthrough",
			queryString: "moving_rollup(sum:metric{*}, 60)",
			wantErr:     false,
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

func TestParseComplexNestedFilters(t *testing.T) {
	// Test parsing a complex nested filter query with AND, OR, AND NOT, and OR NOT
	// Starting query: env:prod AND (host:web-1 OR host:web-2) AND NOT (region:us-west-1)
	queryString := "system.cpu.idle{env:prod AND (host:web-1 OR host:web-2) AND NOT (region:us-west-1)}"
	expectedAfterParse := "system.cpu.idle{(env:prod AND (host:web-1 AND host:web-2) AND region:us-west-1)}"
	expectedAfterAddingFilter := "system.cpu.idle{((env:prod AND (host:web-1 AND host:web-2) AND region:us-west-1) AND service:api)}"

	builder, err := metric.ParseQuery(queryString)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	// Verify the parsed query can be rebuilt (round-trip test)
	result, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if result != expectedAfterParse {
		t.Errorf("Build() after parse = %q, want %q", result, expectedAfterParse)
	}

	// Now add a new filter and verify it's added correctly
	builder = builder.Filter(ddqb.Filter("service").Equal("api"))

	result, err = builder.Build()
	if err != nil {
		t.Fatalf("Build() after adding filter error = %v", err)
	}

	if result != expectedAfterAddingFilter {
		t.Errorf("Build() after adding filter = %q, want %q", result, expectedAfterAddingFilter)
	}
}

func TestParseComplexNestedFiltersWithORNOT(t *testing.T) {
	// Test parsing a complex query with OR NOT as well
	// Starting query: env:prod OR NOT (host:web-1) AND (region:us-east-1 OR region:us-west-2)
	queryString := "avg(5m):system.cpu.idle{env:prod OR NOT (host:web-1) AND (region:us-east-1 OR region:us-west-2)}"
	expectedAfterParse := "avg(5m):system.cpu.idle{(env:prod AND (host:web-1 AND (region:us-east-1 AND region:us-west-2)))}"
	expectedAfterAddingFilter := "avg(5m):system.cpu.idle{(env:prod AND (host:web-1 AND (region:us-east-1 AND region:us-west-2)) AND team:backend)}"

	builder, err := metric.ParseQuery(queryString)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	// Verify the parsed query can be rebuilt
	result, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if result != expectedAfterParse {
		t.Errorf("Build() after parse = %q, want %q", result, expectedAfterParse)
	}

	// Add a new filter
	builder = builder.Filter(ddqb.Filter("team").Equal("backend"))

	result, err = builder.Build()
	if err != nil {
		t.Fatalf("Build() after adding filter error = %v", err)
	}

	if result != expectedAfterAddingFilter {
		t.Errorf("Build() after adding filter = %q, want %q", result, expectedAfterAddingFilter)
	}
}

func TestGetFiltersAndModifyGroups(t *testing.T) {
	queryString := "avg(5m):system.cpu.idle{(env:prod AND (host:web-1 AND (region:us-east-1 AND region:us-west-2)))}"
	builder, _ := metric.ParseQuery(queryString)

	// Access filters directly
	filters := builder.GetFilters()
	if len(filters) == 0 {
		t.Fatal("Expected at least one filter")
	}

	// Modify the first group directly
	if group, ok := filters[0].(metric.FilterGroupBuilder); ok {
		group.And(ddqb.Filter("service").Equal("api"))
	}

	result, _ := builder.Build()
	expected := "avg(5m):system.cpu.idle{(env:prod AND (host:web-1 AND (region:us-east-1 AND region:us-west-2)) AND service:api)}"
	if result != expected {
		t.Errorf("Build() after direct modification = %q, want %q", result, expected)
	}
}

func TestAddFilterToDeepestNestedGroup(t *testing.T) {
	queryString := "avg(5m):system.cpu.idle{(env:prod AND (host:web-1 AND (region:us-east-1 AND region:us-west-2)))}"
	expected := "avg(5m):system.cpu.idle{(env:prod AND (host:web-1 AND (region:us-east-1 AND region:us-west-2 AND region:ap-southeast-2)))}"

	builder, err := metric.ParseQuery(queryString)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	// Find the deepest group by searching for a group that contains both region filters
	// and whose Build() output matches the pattern of the deepest nested group
	// The deepest group should build to exactly "(region:us-east-1 AND region:us-west-2)"
	deepestGroup := builder.FindGroup(func(g metric.FilterGroupBuilder) bool {
		built, _ := g.Build()
		// The deepest group should contain both region filters and not have nested parentheses
		// We check if the built string is exactly "(region:us-east-1 AND region:us-west-2)"
		// or starts with that pattern
		hasBothRegions := contains(built, "region:us-east-1") && contains(built, "region:us-west-2")
		if !hasBothRegions {
			return false
		}
		// Check if this is a simple group (no nested parentheses after the opening paren)
		// by looking for the pattern: starts with "(" and contains both regions without nested "("
		opensWithParen := len(built) > 0 && built[0] == '('
		if !opensWithParen {
			return false
		}
		// Check for nested parentheses - if we find a "(" after the first one (and not part of "NOT"),
		// this is not the deepest group
		foundNestedParen := false
		for i := 1; i < len(built); i++ {
			if built[i] == '(' && built[i-1] != 'N' {
				foundNestedParen = true
				break
			}
		}
		// The deepest group should not have nested parentheses
		return !foundNestedParen
	})

	if deepestGroup == nil {
		t.Fatal("Expected to find a deepest nested group")
	}

	// Add filter to the deepest group
	builder = builder.AddToGroup(deepestGroup, ddqb.Filter("region").Equal("ap-southeast-2"))

	result, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if result != expected {
		t.Errorf("Build() after adding to deepest group = %q, want %q", result, expected)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					findInMiddle(s, substr))))
}

func findInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFindGroupAndAddToGroup(t *testing.T) {
	queryString := "avg(5m):system.cpu.idle{(env:prod AND (host:web-1 AND (region:us-east-1 AND region:us-west-2)))}"
	builder, _ := metric.ParseQuery(queryString)

	// Find any group (first one found)
	group := builder.FindGroup(func(_ metric.FilterGroupBuilder) bool {
		return true // Find first group
	})

	if group == nil {
		t.Fatal("Expected to find a group")
	}

	// Add filter to the found group
	builder = builder.AddToGroup(group, ddqb.Filter("region").Equal("us-west-1"))

	result, _ := builder.Build()
	expected := "avg(5m):system.cpu.idle{(env:prod AND (host:web-1 AND (region:us-east-1 AND region:us-west-2)) AND region:us-west-1)}"
	if result != expected {
		t.Errorf("Build() after FindGroup + AddToGroup = %q, want %q", result, expected)
	}
}

func TestExpressionNormalization_MixedAndComma(t *testing.T) {
	// Start with an expression containing comma-style filters and a negation
	query := "top(system.cpu.idle{host:web-1, env:staging, !region:us-west-2}, 1, 'max', 'desc')"
	builder, err := metric.ParseQuery(query)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	// Add a filter group using explicit OR, which mixes styles
	fg := metric.NewFilterGroupBuilder()
	fg.Or(ddqb.Filter("service").Equal("api"))
	fg.Or(ddqb.Filter("team").Equal("backend"))
	builder = builder.Filter(fg)

	out, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Expect explicit boolean style inside braces (no commas), and NOT for negation
	// Extract the first filter block {...}
	start := strings.Index(out, "{")
	end := strings.Index(out[start+1:], "}")
	var filterBlock string
	if start != -1 && end != -1 {
		filterBlock = out[start : start+end+2]
	} else {
		filterBlock = out
	}
	if contains(filterBlock, ",") {
		t.Errorf("expected no commas after normalization, got: %s", out)
	}
	if !contains(filterBlock, " AND ") {
		t.Errorf("expected AND separators after normalization, got: %s", out)
	}
	if !contains(filterBlock, " NOT ") {
		t.Errorf("expected NOT for negations after normalization, got: %s", out)
	}
}

func TestExpressionNormalization_DefaultCommaWhenNoExplicit(t *testing.T) {
	// Expression with comma filters; append a simple filter (no explicit group)
	query := "top(system.cpu.idle{host:web-1, env:staging}, 1, 'max', 'desc')"
	builder, err := metric.ParseQuery(query)
	if err != nil {
		t.Fatalf("ParseQuery() error = %v", err)
	}

	builder = builder.Filter(ddqb.Filter("region").Equal("us-west-2"))
	out, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Should remain comma style (no AND), as there was no explicit boolean usage
	start := strings.Index(out, "{")
	end := strings.Index(out[start+1:], "}")
	var filterBlock string
	if start != -1 && end != -1 {
		filterBlock = out[start : start+end+2]
	} else {
		filterBlock = out
	}
	if !contains(filterBlock, ",") {
		t.Errorf("expected commas to remain when no explicit boolean operators, got: %s", out)
	}
	if contains(filterBlock, " AND ") {
		t.Errorf("did not expect AND when no explicit boolean operators, got: %s", out)
	}
}
