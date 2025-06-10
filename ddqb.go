// Package ddqb provides a fluent API for building DataDog queries.
package ddqb

import "github.com/jonwinton/ddqb/metric"

// Metric creates a new metric query builder.
// This is the main entry point for building metric queries.
func Metric() metric.MetricQueryBuilder {
	return metric.NewMetricQueryBuilder()
}

// Filter creates a new filter builder with the given key.
// This is a convenience function for creating filter builders.
func Filter(key string) metric.FilterBuilder {
	return metric.NewFilterBuilder(key)
}

// Function creates a new function builder with the given name.
// This is a convenience function for creating function builders.
func Function(name string) metric.FunctionBuilder {
	return metric.NewFunctionBuilder(name)
}