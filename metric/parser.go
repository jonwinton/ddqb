package metric

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jonwinton/ddqp"
)

// ParseQuery parses a DataDog query string and returns a MetricQueryBuilder
// that can be modified using the fluent API.
func ParseQuery(queryString string) (MetricQueryBuilder, error) {
	// Extract time window if present (DDQP doesn't parse avg(5m): format)
	timeWindow, cleanedQuery := extractAndRemoveTimeWindow(queryString)

	parser := ddqp.NewMetricQueryParser()
	parsed, err := parser.Parse(cleanedQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Check if this is a wrapped aggregator function (like moving_rollup(...))
	// We only support direct metric queries for now
	if parsed.AggregatorFuction != nil {
		return nil, fmt.Errorf("aggregator functions wrapping queries are not yet supported")
	}

	if parsed.Query == nil {
		return nil, fmt.Errorf("query is missing required Query component")
	}

	builder := NewMetricQueryBuilder()

	// Set aggregator if present
	if parsed.Query.Aggregator != nil {
		builder = builder.Aggregator(parsed.Query.Aggregator.Name)
		// Set time window if we extracted one
		if timeWindow != "" {
			builder = builder.TimeWindow(timeWindow)
		}
	}

	// Set metric name
	builder = builder.Metric(parsed.Query.MetricName)

	// Convert filters
	if parsed.Query.Filters != nil {
		filters, err := convertFilters(parsed.Query.Filters)
		if err != nil {
			return nil, fmt.Errorf("failed to convert filters: %w", err)
		}
		for _, filter := range filters {
			builder = builder.Filter(filter)
		}
	}

	// Set grouping
	if len(parsed.Query.Grouping) > 0 {
		builder = builder.GroupBy(parsed.Query.Grouping...)
	}

	// Convert functions
	for _, fn := range parsed.Query.Function {
		functionBuilder := NewFunctionBuilder(fn.Name)
		for _, arg := range fn.Args {
			functionBuilder = functionBuilder.WithArg(arg.String())
		}
		builder = builder.ApplyFunction(functionBuilder)
	}

	return builder, nil
}

// convertFilters converts DDQP filter structures to DDQB FilterBuilder instances
func convertFilters(mf *ddqp.MetricFilter) ([]FilterBuilder, error) {
	var filters []FilterBuilder

	// Handle the left parameter
	if mf.Left != nil {
		filter, err := convertParam(mf.Left)
		if err != nil {
			return nil, err
		}
		if filter != nil {
			filters = append(filters, filter)
		}
	}

	// Handle additional parameters
	for _, param := range mf.Parameters {
		filter, err := convertParam(param)
		if err != nil {
			return nil, err
		}
		if filter != nil {
			filters = append(filters, filter)
		}
	}

	return filters, nil
}

// convertParam converts a DDQP Param to a DDQB FilterBuilder
func convertParam(param *ddqp.Param) (FilterBuilder, error) {
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

	// Handle grouped filters - for now, we'll extract simple filters from them
	// Complex AND/OR logic in grouped filters is not fully supported in DDQB yet
	if param.GroupedFilter != nil {
		// Extract filters from grouped filter
		var filters []FilterBuilder
		for _, p := range param.GroupedFilter.Parameters {
			if p.SimpleFilter != nil {
				filter, err := convertSimpleFilter(p.SimpleFilter)
				if err != nil {
					return nil, err
				}
				if filter != nil {
					filters = append(filters, filter)
				}
			}
		}
		// For now, we'll return the first filter or handle it differently
		// TODO: Properly handle grouped filters with AND/OR logic
		if len(filters) > 0 {
			return filters[0], nil
		}
	}

	// Handle separator (comma, AND, OR, etc.) - these are just separators, not filters
	if param.Separator != nil {
		return nil, nil
	}

	return nil, nil
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

