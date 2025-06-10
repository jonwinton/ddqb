package main

import (
	"fmt"
	"log"

	"github.com/jonwinton/ddqb"
)

func main() {
	// Example 1: Simple metric query
	fmt.Println("Example 1: Simple metric query")
	query, err := ddqb.Metric().
		Metric("system.cpu.idle").
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)
	// Result will include {*} as DataDog requires: system.cpu.idle {*}

	// Example 2: Metric query with aggregation
	fmt.Println("Example 2: Metric query with aggregation")
	query, err = ddqb.Metric().
		Aggregator("avg").
		TimeWindow("5m").
		Metric("system.cpu.idle").
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 3: Metric query with single filter
	fmt.Println("Example 3: Metric query with single filter")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(ddqb.Filter("host").Equal("web-1")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 4: Metric query with multiple filters
	fmt.Println("Example 4: Metric query with multiple filters")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(ddqb.Filter("host").Equal("web-1")).
		Filter(ddqb.Filter("env").Equal("prod")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 5: Metric query with grouping
	fmt.Println("Example 5: Metric query with grouping")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		GroupBy("host", "env").
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 6: Metric query with function
	fmt.Println("Example 6: Metric query with function")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		ApplyFunction(ddqb.Function("fill").WithArg("0")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 7: Complex metric query
	fmt.Println("Example 7: Complex metric query")
	query, err = ddqb.Metric().
		Aggregator("avg").
		TimeWindow("5m").
		Metric("system.cpu.idle").
		Filter(ddqb.Filter("host").Equal("web-1")).
		Filter(ddqb.Filter("env").Equal("prod")).
		GroupBy("host").
		ApplyFunction(ddqb.Function("fill").WithArg("0")).
		ApplyFunction(ddqb.Function("rollup").WithArgs("60", "sum")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)
}