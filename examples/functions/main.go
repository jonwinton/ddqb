package main

import (
	"fmt"
	"log"

	"github.com/jonwinton/ddqb"
)

func main() {
	fmt.Println("DataDog Query Builder - Function Examples")
	fmt.Println("========================================")

	// Example 1: Fill function with zero
	fmt.Println("Example 1: Fill function with zero")
	query, err := ddqb.Metric().
		Metric("system.cpu.idle").
		ApplyFunction(ddqb.Function("fill").WithArg("0")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 2: Rollup function
	fmt.Println("Example 2: Rollup function")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		ApplyFunction(ddqb.Function("rollup").WithArgs("60", "avg")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 3: Moving average function
	fmt.Println("Example 3: Moving average function")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		ApplyFunction(ddqb.Function("moving_average").WithArg("5")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 4: Timeshift function
	fmt.Println("Example 4: Timeshift function")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		ApplyFunction(ddqb.Function("timeshift").WithArg("1d")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 5: Multiple functions chained
	fmt.Println("Example 5: Multiple functions chained")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		ApplyFunction(ddqb.Function("fill").WithArg("0")).
		ApplyFunction(ddqb.Function("rollup").WithArgs("60", "avg")).
		ApplyFunction(ddqb.Function("moving_average").WithArg("5")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 6: Complex query with functions
	fmt.Println("Example 6: Complex query with functions")
	query, err = ddqb.Metric().
		Aggregator("avg").
		TimeWindow("5m").
		Metric("system.cpu.idle").
		Filter(ddqb.Filter("host").Regex("web-.*")).
		Filter(ddqb.Filter("env").Equal("prod")).
		GroupBy("host").
		ApplyFunction(ddqb.Function("fill").WithArg("0")).
		ApplyFunction(ddqb.Function("rollup").WithArgs("60", "avg")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)
}