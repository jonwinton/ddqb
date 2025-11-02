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

	// Filter adds a filter condition or filter group to the query.
	Filter(filter FilterExpression) MetricQueryBuilder

	// GetFilters returns all filter expressions in the query.
	// This allows direct access to modify FilterGroupBuilder instances.
	GetFilters() []FilterExpression

	// FindGroup finds the first FilterGroupBuilder that matches the predicate function.
	// Returns nil if no matching group is found.
	FindGroup(predicate func(FilterGroupBuilder) bool) FilterGroupBuilder

	// AddToGroup adds a filter to the specified FilterGroupBuilder.
	// The filter is added using the group's existing operator (AND or OR).
	AddToGroup(group FilterGroupBuilder, filter FilterExpression) MetricQueryBuilder

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
	filters    []FilterExpression
	groupBy    []string
	functions  []FunctionBuilder
}

// NewMetricQueryBuilder creates a new metric query builder.
func NewMetricQueryBuilder() MetricQueryBuilder {
	return &metricQueryBuilder{
		filters:   make([]FilterExpression, 0),
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

// Filter adds a filter condition or filter group to the query.
func (b *metricQueryBuilder) Filter(filter FilterExpression) MetricQueryBuilder {
	b.filters = append(b.filters, filter)
	return b
}

// GetFilters returns all filter expressions in the query.
// Note: The returned slice shares the same underlying array as the builder's filters.
// Modifying FilterGroupBuilder instances in this slice will modify the query.
func (b *metricQueryBuilder) GetFilters() []FilterExpression {
	return b.filters
}

// FindGroup finds the first FilterGroupBuilder that matches the predicate function.
// It searches recursively through all filters and nested groups.
func (b *metricQueryBuilder) FindGroup(predicate func(FilterGroupBuilder) bool) FilterGroupBuilder {
	for _, filter := range b.filters {
		if group := findGroupRecursive(filter, predicate); group != nil {
			return group
		}
	}
	return nil
}

// AddToGroup adds a filter to the specified FilterGroupBuilder.
func (b *metricQueryBuilder) AddToGroup(group FilterGroupBuilder, filter FilterExpression) MetricQueryBuilder {
	if group == nil {
		// If group is nil, just add as a new filter
		b.filters = append(b.filters, filter)
		return b
	}

	// Cast to concrete type to modify
	if groupImpl, ok := group.(*filterGroupBuilder); ok {
		if groupImpl.operator == AndOperator {
			groupImpl.AND(filter)
		} else {
			groupImpl.OR(filter)
		}
	}
	return b
}

// findGroupRecursive recursively searches for a group matching the predicate.
func findGroupRecursive(expr FilterExpression, predicate func(FilterGroupBuilder) bool) FilterGroupBuilder {
	group, ok := expr.(FilterGroupBuilder)
	if ok && predicate(group) {
		return group
	}

	// If it's a concrete group, search nested expressions
	if groupImpl, ok := expr.(*filterGroupBuilder); ok {
		for _, nestedExpr := range groupImpl.expressions {
			if found := findGroupRecursive(nestedExpr, predicate); found != nil {
				return found
			}
		}
	}

	return nil
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
		// Check if any filter uses explicit operators (FilterGroupBuilder)
		// If so, we must wrap everything in a group with explicit AND operators
		// to avoid mixing comma notation with explicit AND/OR (invalid syntax)
		hasExplicitOperators := false
		for _, filter := range b.filters {
			if _, ok := filter.(FilterGroupBuilder); ok {
				hasExplicitOperators = true
				break
			}
		}

		if hasExplicitOperators {
			// Wrap all filters in a group with explicit AND operators
			group := NewFilterGroupBuilder()
			for _, filter := range b.filters {
				group.AND(filter)
			}
			groupStr, err := group.Build()
			if err != nil {
				return "", fmt.Errorf("error building filter group: %w", err)
			}
			parts = append(parts, fmt.Sprintf("{%s}", groupStr))
		} else {
			// All filters are simple - use comma notation (implicit AND)
			var filterStrs []string
			for _, filter := range b.filters {
				filterStr, err := filter.Build()
				if err != nil {
					return "", fmt.Errorf("error building filter: %w", err)
				}
				filterStrs = append(filterStrs, filterStr)
			}
			parts = append(parts, fmt.Sprintf("{%s}", strings.Join(filterStrs, ", ")))
		}
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

