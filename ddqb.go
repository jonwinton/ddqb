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

// FromQuery parses an existing DataDog query string and returns a MetricQueryBuilder
// that can be modified using the fluent API.
//
// Example:
//
//	builder, err := ddqb.FromQuery("avg(5m):system.cpu.idle{host:web-1} by {host}.fill(0)")
//	if err != nil {
//		// handle error
//	}
//	modifiedQuery, err := builder.TimeWindow("10m").Filter(ddqb.Filter("env").Equal("prod")).Build()
func FromQuery(queryString string) (metric.MetricQueryBuilder, error) {
	return metric.ParseQuery(queryString)
}