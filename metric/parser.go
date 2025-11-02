package metric

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jonwinton/ddqp"
)

// ParseQuery parses a Datadog query string and returns a QueryBuilder
// that can be modified using the fluent API.
func ParseQuery(queryString string) (QueryBuilder, error) {
	// Extract time window if present (DDQP doesn't parse avg(5m): format)
	timeWindow, cleanedQuery := extractAndRemoveTimeWindow(queryString)

	// Use the GenericParser so we can accept metric expressions and queries
	parser := ddqp.NewGenericParser()
	parsed, err := parser.Parse(cleanedQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// If we got a plain MetricQuery without wrapper aggregator, use the structured builder
	if parsed.MetricQuery != nil && parsed.MetricQuery.AggregatorFuction == nil {
		mq := parsed.MetricQuery
		if mq.Query == nil {
			return nil, fmt.Errorf("query is missing required Query component")
		}

		builder := NewMetricQueryBuilder()

		// Set aggregator if present
		if mq.Query.Aggregator != nil {
			builder = builder.Aggregator(mq.Query.Aggregator.Name)
			// Set time window if we extracted one
			if timeWindow != "" {
				builder = builder.TimeWindow(timeWindow)
			}
		}

		// Set metric name
		builder = builder.Metric(mq.Query.MetricName)

		// Convert filters
		if mq.Query.Filters != nil {
			filters, err := convertFilters(mq.Query.Filters)
			if err != nil {
				return nil, fmt.Errorf("failed to convert filters: %w", err)
			}
			for _, filter := range filters {
				builder = builder.Filter(filter)
			}
		}

		// Set grouping
		if len(mq.Query.Grouping) > 0 {
			builder = builder.GroupBy(mq.Query.Grouping...)
		}

		// Convert functions
		for _, fn := range mq.Query.Function {
			functionBuilder := NewFunctionBuilder(fn.Name)
			for _, arg := range fn.Args {
				functionBuilder = functionBuilder.WithArg(arg.String())
			}
			builder = builder.ApplyFunction(functionBuilder)
		}

		return builder, nil
	}

	// Otherwise, it's a MetricExpression or a wrapped MetricQuery. Return a passthrough builder
	// that preserves the original query string (including any time window prefix we detected).
	return newExpressionPassthroughBuilder(queryString), nil
}

// convertFilters converts DDQP filter structures to DDQB FilterExpression instances
func convertFilters(mf *ddqp.MetricFilter) ([]FilterExpression, error) {
	var expressions []FilterExpression
	var currentGroup *filterGroupBuilder
	var groupOperator GroupOperator

	// Process left parameter if present
	if mf.Left != nil {
		expr, err := convertParam(mf.Left)
		if err != nil {
			return nil, err
		}
		if expr != nil {
			expressions = append(expressions, expr)
		}
	}

	// Process additional parameters, tracking separators to build groups
	for _, param := range mf.Parameters {
		// Check if this is a separator
		if param.Separator != nil {
			// Only create groups for explicit AND/OR operators, not for commas
			// Commas represent implicit AND and should remain as separate expressions
			if param.Separator.And {
				// Start or continue an AND group
				if currentGroup == nil {
					// Start a new group
					currentGroup = &filterGroupBuilder{
						expressions: make([]FilterExpression, 0),
						operator:    AndOperator,
						negated:     false,
					}
					// Move the last expression into the group if there is one
					if len(expressions) > 0 {
						currentGroup.expressions = append(currentGroup.expressions, expressions[len(expressions)-1])
						expressions = expressions[:len(expressions)-1]
					}
				}
				groupOperator = AndOperator
				currentGroup.operator = AndOperator
			} else if param.Separator.Or {
				// Start or continue an OR group
				if currentGroup == nil {
					// Start a new group
					currentGroup = &filterGroupBuilder{
						expressions: make([]FilterExpression, 0),
						operator:    OrOperator,
						negated:     false,
					}
					// Move the last expression into the group if there is one
					if len(expressions) > 0 {
						currentGroup.expressions = append(currentGroup.expressions, expressions[len(expressions)-1])
						expressions = expressions[:len(expressions)-1]
					}
				}
				groupOperator = OrOperator
				currentGroup.operator = OrOperator
			}
			// For commas, we don't create groups - they remain as separate expressions
			// which will be joined with commas (implicit AND) in the Build() method
			continue
		}

		// Convert the parameter to an expression
		expr, err := convertParam(param)
		if err != nil {
			return nil, err
		}
		if expr == nil {
			continue
		}

		// Add to current group or as standalone expression
		if currentGroup != nil {
			// Add to current group with the appropriate operator
			if groupOperator == AndOperator {
				currentGroup.AND(expr)
			} else {
				currentGroup.OR(expr)
			}
		} else {
			// Standalone expression (will be joined with commas for implicit AND)
			expressions = append(expressions, expr)
		}
	}

	// Close any open group
	if currentGroup != nil {
		expressions = append(expressions, currentGroup)
	}

	return expressions, nil
}

// convertParam converts a DDQP Param to a DDQB FilterExpression
func convertParam(param *ddqp.Param) (FilterExpression, error) {
	if param == nil {
		return nil, nil
	}

	// Handle asterisk (wildcard filter - we skip it as DDQB uses {*} by default)
	if param.Asterisk {
		return nil, nil
	}

	// Handle simple filters
	if param.SimpleFilter != nil {
		return convertSimpleFilter(param.SimpleFilter)
	}

	// Handle grouped filters - recursively convert with proper AND/OR logic
	if param.GroupedFilter != nil {
		return convertGroupedFilter(param.GroupedFilter)
	}

	// Handle separator (comma, AND, OR, etc.) - these are handled in convertFilters
	if param.Separator != nil {
		return nil, nil
	}

	return nil, nil
}

// convertGroupedFilter converts a DDQP GroupedFilter to a DDQB FilterGroupBuilder
func convertGroupedFilter(gf *ddqp.GroupedFilter) (FilterExpression, error) {
	if gf == nil {
		return nil, nil
	}

	group := NewFilterGroupBuilder()
	currentOperator := AndOperator // Default to AND

	// Process parameters in the grouped filter
	for _, param := range gf.Parameters {
		// Check for separator to determine operator
		if param.Separator != nil {
			if param.Separator.And || param.Separator.Comma {
				currentOperator = AndOperator
			} else if param.Separator.Or {
				currentOperator = OrOperator
			}
			continue
		}

		// Convert parameter to expression
		expr, err := convertParam(param)
		if err != nil {
			return nil, err
		}
		if expr == nil {
			continue
		}

		// Add to group with appropriate operator
		if currentOperator == AndOperator {
			group.AND(expr)
		} else {
			group.OR(expr)
		}
	}

	// Check if group has any expressions
	groupImpl := group.(*filterGroupBuilder)
	if len(groupImpl.expressions) == 0 {
		return nil, nil
	}

	return group, nil
}

// convertSimpleFilter converts a DDQP SimpleFilter to a DDQB FilterBuilder
func convertSimpleFilter(sf *ddqp.SimpleFilter) (FilterBuilder, error) {
	if sf == nil {
		return nil, nil
	}

	key := sf.FilterKey
	if key == "" {
		return nil, fmt.Errorf("filter key is empty")
	}

	builder := NewFilterBuilder(key)

	if sf.FilterSeparator == nil {
		return nil, fmt.Errorf("filter separator is missing")
	}

	// Handle filter value
	value, err := extractFilterValue(sf.FilterValue)
	if err != nil {
		return nil, fmt.Errorf("failed to extract filter value: %w", err)
	}

	// Convert based on separator type
	fs := sf.FilterSeparator
	switch {
	case fs.Colon:
		if sf.Negative {
			return builder.NotEqual(value), nil
		}
		return builder.Equal(value), nil
	case fs.Regex:
		return builder.Regex(value), nil
	case fs.In:
		// For IN filters, extract list values
		values, err := extractFilterValues(sf.FilterValue)
		if err != nil {
			return nil, err
		}
		return builder.In(values...), nil
	case fs.NotIn:
		// For NOT IN filters, extract list values
		values, err := extractFilterValues(sf.FilterValue)
		if err != nil {
			return nil, err
		}
		return builder.NotIn(values...), nil
	default:
		// Default to equal if separator is not recognized
		if sf.Negative {
			return builder.NotEqual(value), nil
		}
		return builder.Equal(value), nil
	}
}

// extractFilterValue extracts a single string value from a FilterValue
func extractFilterValue(fv *ddqp.FilterValue) (string, error) {
	if fv == nil {
		return "", fmt.Errorf("filter value is nil")
	}

	if fv.SimpleValue != nil {
		return extractValueString(fv.SimpleValue), nil
	}

	if len(fv.ListValue) > 0 {
		// For list values, return the first non-separator value
		for _, v := range fv.ListValue {
			valStr := extractValueString(v)
			if valStr != "" {
				return valStr, nil
			}
		}
		return "", fmt.Errorf("filter value list has no valid values")
	}

	return "", fmt.Errorf("filter value has no content")
}

// extractFilterValues extracts multiple string values from a FilterValue (for IN/NOT IN)
func extractFilterValues(fv *ddqp.FilterValue) ([]string, error) {
	if fv == nil {
		return nil, fmt.Errorf("filter value is nil")
	}

	var values []string

	// For IN/NOT IN filters, we expect ListValue
	if len(fv.ListValue) > 0 {
		for _, v := range fv.ListValue {
			// Skip separator values (commas, AND, OR, etc.)
			if v.Separator != nil {
				continue
			}
			// Extract the actual value string, removing quotes if present
			valStr := extractValueString(v)
			if valStr != "" {
				values = append(values, valStr)
			}
		}
	} else if fv.SimpleValue != nil {
		// Fallback: if we have a simple value, use it as a single-item list
		valStr := extractValueString(fv.SimpleValue)
		if valStr != "" {
			values = append(values, valStr)
		}
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("no values found in filter")
	}

	return values, nil
}

// extractValueString extracts a clean string value from a Value, skipping separators
func extractValueString(v *ddqp.Value) string {
	if v == nil {
		return ""
	}

	// Skip separator values
	if v.Separator != nil {
		return ""
	}

	// Extract based on value type
	if v.Str != nil {
		return strings.Trim(*v.Str, "\"'")
	}
	if v.Identifier != nil {
		return *v.Identifier
	}
	if v.Number != nil {
		return fmt.Sprintf("%g", *v.Number)
	}
	if v.Boolean != nil {
		return fmt.Sprintf("%v", *v.Boolean)
	}
	if v.Wildcard != nil {
		return *v.Wildcard
	}

	return ""
}

// extractAndRemoveTimeWindow extracts time window from query and returns both the time window
// and the cleaned query string without the time window (for DDQP parsing)
// DDQP doesn't support avg(5m): format, so we need to pre-process
func extractAndRemoveTimeWindow(queryString string) (timeWindow string, cleanedQuery string) {
	// Pattern to match aggregator with time window: avg(5m), sum(10m), etc.
	// Matches any aggregator name followed by (time_window) where time_window is like 5m, 10s, 1h, last_5m, etc.
	pattern := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\(([0-9]+[smhd]|last_[0-9]+[smhd])\):(.*)$`)
	matches := pattern.FindStringSubmatch(queryString)
	if len(matches) == 4 {
		// Found time window: matches[1] is aggregator, matches[2] is time window, matches[3] is rest of query
		aggregator := matches[1]
		timeWindow = matches[2]
		cleanedQuery = aggregator + ":" + matches[3]
		return timeWindow, cleanedQuery
	}
	// No time window found, return original query
	return "", queryString
}
