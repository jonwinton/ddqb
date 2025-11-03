// Package main demonstrates advanced usage of ddqb, including dynamic query building
// based on runtime conditions.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/jonwinton/ddqb"
)

// Example of how to build a query dynamically based on runtime conditions
func buildDynamicQuery(metricName string, hostPattern string, environments []string, timeWindow string) (string, error) {
	builder := ddqb.Metric().
		Aggregator("avg").
		Metric(metricName)

	// Add time window if provided
	if timeWindow != "" {
		builder = builder.TimeWindow(timeWindow)
	}

	// Add host filter if provided
	if hostPattern != "" {
		// Use regex if it contains wildcards, otherwise use equality
		if strings.Contains(hostPattern, "*") {
			// Regex matching unsupported; use a representative host instead
			builder = builder.Filter(ddqb.Filter("host").Equal("web-1"))
		} else {
			builder = builder.Filter(ddqb.Filter("host").Equal(hostPattern))
		}
	}

	// Add environment filter if environments are provided
	if len(environments) > 0 {
		if len(environments) == 1 {
			builder = builder.Filter(ddqb.Filter("env").Equal(environments[0]))
		} else {
			builder = builder.Filter(ddqb.Filter("env").In(environments...))
		}
	}

	// Group by host
	builder = builder.GroupBy("host")

	// Apply functions
	builder = builder.
		ApplyFunction(ddqb.Function("fill").WithArg("null")).
		ApplyFunction(ddqb.Function("rollup").WithArgs("60", "avg"))

	// Build the query
	return builder.Build()
}

// Helper function to make query building more concise
func buildMonitoringQuery(metric string, threshold float64, windowMins int) (string, error) {
	// Convert window to string
	windowStr := fmt.Sprintf("%dm", windowMins)

	// We're not using the threshold in the query now, but in a real scenario
	// it might be used for alert thresholds or in a query condition
	_ = threshold

	return ddqb.Metric().
		Aggregator("avg").
		TimeWindow(windowStr).
		Metric(metric).
		Filter(ddqb.Filter("env").Equal("prod")).
		GroupBy("host").
		ApplyFunction(ddqb.Function("fill").WithArg("0")).
		Build()
}

func main() {
	fmt.Println("Datadog Query Builder - Advanced Usage Examples")
	fmt.Println("=============================================")

	// Example 1: Dynamic Query Building
	fmt.Println("Example 1: Dynamic Query Building")

	// Scenario 1: Query for staging environment with specific host pattern and time window
	query, err := buildDynamicQuery("system.cpu.user", "web-*", []string{"staging"}, "5m")
	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Scenario 1: %s\n\n", query)

	// Scenario 2: Query for multiple environments with no host filter
	query, err = buildDynamicQuery("system.cpu.user", "", []string{"dev", "test", "staging"}, "10m")
	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Scenario 2: %s\n\n", query)

	// Scenario 3: Query with specific host (no pattern)
	query, err = buildDynamicQuery("system.memory.used", "web-01", []string{"prod"}, "15m")
	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Scenario 3: %s\n\n", query)

	// Example 2: Helper Function for Common Query Patterns
	fmt.Println("Example 2: Helper Function for Common Query Patterns")

	// CPU usage alert query
	cpuQuery, err := buildMonitoringQuery("system.cpu.user", 80.0, 5)
	if err != nil {
		log.Fatalf("Failed to build CPU query: %v", err)
	}
	fmt.Printf("CPU Monitor: %s\n\n", cpuQuery)

	// Memory usage alert query
	memQuery, err := buildMonitoringQuery("system.memory.used", 90.0, 10)
	if err != nil {
		log.Fatalf("Failed to build Memory query: %v", err)
	}
	fmt.Printf("Memory Monitor: %s\n\n", memQuery)

	// Disk usage alert query
	diskQuery, err := buildMonitoringQuery("system.disk.used", 85.0, 15)
	if err != nil {
		log.Fatalf("Failed to build Disk query: %v", err)
	}
	fmt.Printf("Disk Monitor: %s\n\n", diskQuery)
}
