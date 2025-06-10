// Package metric provides builders for creating DataDog metric queries.
package metric

import (
	"fmt"
	"strings"
)

// MetricQueryBuilder provides a fluent interface for building metric queries.
type MetricQueryBuilder interface {
	// Metric sets the metric name for the query.
	Metric(name string) MetricQueryBuilder

	// Aggregator sets the aggregation method for the query (e.g., "avg", "sum").
	Aggregator(agg string) MetricQueryBuilder

	// Filter adds a filter condition to the query.
	Filter(filter FilterBuilder) MetricQueryBuilder

	// GroupBy sets grouping parameters for the query.
	GroupBy(groups ...string) MetricQueryBuilder

	// ApplyFunction applies a function to the query.
	ApplyFunction(fn FunctionBuilder) MetricQueryBuilder

	// TimeWindow sets the time window for the query (e.g., "1m", "5m").
	TimeWindow(window string) MetricQueryBuilder

	// Build returns the built query as a string.
	Build() (string, error)
}

// metricQueryBuilder is the concrete implementation of the MetricQueryBuilder interface.
type metricQueryBuilder struct {
	metric     string
	aggregator string
	timeWindow string
	filters    []FilterBuilder
	groupBy    []string
	functions  []FunctionBuilder
}

// NewMetricQueryBuilder creates a new metric query builder.
func NewMetricQueryBuilder() MetricQueryBuilder {
	return &metricQueryBuilder{
		filters:   make([]FilterBuilder, 0),
		groupBy:   make([]string, 0),
		functions: make([]FunctionBuilder, 0),
	}
}

// Metric sets the metric name for the query.
func (b *metricQueryBuilder) Metric(name string) MetricQueryBuilder {
	b.metric = name
	return b
}

// Aggregator sets the aggregation method for the query (e.g., "avg", "sum").
func (b *metricQueryBuilder) Aggregator(agg string) MetricQueryBuilder {
	b.aggregator = agg
	return b
}

// Filter adds a filter condition to the query.
func (b *metricQueryBuilder) Filter(filter FilterBuilder) MetricQueryBuilder {
	b.filters = append(b.filters, filter)
	return b
}

// GroupBy sets grouping parameters for the query.
func (b *metricQueryBuilder) GroupBy(groups ...string) MetricQueryBuilder {
	b.groupBy = append(b.groupBy, groups...)
	return b
}

// ApplyFunction applies a function to the query.
func (b *metricQueryBuilder) ApplyFunction(fn FunctionBuilder) MetricQueryBuilder {
	b.functions = append(b.functions, fn)
	return b
}

// TimeWindow sets the time window for the query (e.g., "1m", "5m").
func (b *metricQueryBuilder) TimeWindow(window string) MetricQueryBuilder {
	b.timeWindow = window
	return b
}

// Build returns the built query as a string.
func (b *metricQueryBuilder) Build() (string, error) {
	if b.metric == "" {
		return "", fmt.Errorf("metric name is required")
	}

	// Start building the query
	var parts []string

	// Add aggregator and time window if provided
	if b.aggregator != "" {
		if b.timeWindow != "" {
			parts = append(parts, fmt.Sprintf("%s(%s):", b.aggregator, b.timeWindow))
		} else {
			parts = append(parts, fmt.Sprintf("%s:", b.aggregator))
		}
	}

	// Add metric name
	parts = append(parts, b.metric)

	// Add filters if provided, or {*} if no filters
	if len(b.filters) > 0 {
		var filterStrs []string
		for _, filter := range b.filters {
			filterStr, err := filter.Build()
			if err != nil {
				return "", fmt.Errorf("error building filter: %w", err)
			}
			filterStrs = append(filterStrs, filterStr)
		}
		parts = append(parts, fmt.Sprintf("{%s}", strings.Join(filterStrs, ", ")))
	} else {
		// DataDog requires {*} for queries without filters
		parts = append(parts, "{*}")
	}

	// Add group by if provided
	if len(b.groupBy) > 0 {
		parts = append(parts, fmt.Sprintf(" by {%s}", strings.Join(b.groupBy, ", ")))
	}

	// Add functions if provided
	for _, fn := range b.functions {
		fnStr, err := fn.Build()
		if err != nil {
			return "", fmt.Errorf("error building function: %w", err)
		}
		parts = append(parts, fnStr)
	}

	return strings.Join(parts, ""), nil
}

