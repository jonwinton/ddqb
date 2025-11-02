// Package main demonstrates parsing existing DataDog query strings and modifying them using ddqb.
package main

import (
	"fmt"
	"log"

	"github.com/jonwinton/ddqb"
)

func main() {
	fmt.Println("DataDog Query Builder - Parse and Modify Examples")
	fmt.Println("=================================================")

	// Example 1: Parse a simple query and modify it
	fmt.Println("\nExample 1: Parse a simple query and modify it")
	query1 := "system.cpu.idle{host:web-1}"
	builder1, err := ddqb.FromQuery(query1)
	if err != nil {
		log.Fatalf("Failed to parse query: %v", err)
	}

	// Modify by adding a filter
	modified1, err := builder1.
		Filter(ddqb.Filter("env").Equal("prod")).
		Build()
	if err != nil {
		log.Fatalf("Failed to build modified query: %v", err)
	}
	fmt.Printf("Original:  %s\n", query1)
	fmt.Printf("Modified:  %s\n", modified1)

	// Example 2: Parse a complex query and change time window
	fmt.Println("\nExample 2: Parse a complex query and change time window")
	query2 := "avg(5m):system.cpu.idle{host:web-1,env:prod} by {host}.fill(0)"
	builder2, err := ddqb.FromQuery(query2)
	if err != nil {
		log.Fatalf("Failed to parse query: %v", err)
	}

	// Change time window and add a function
	modified2, err := builder2.
		TimeWindow("10m").
		ApplyFunction(ddqb.Function("rollup").WithArgs("60", "avg")).
		Build()
	if err != nil {
		log.Fatalf("Failed to build modified query: %v", err)
	}
	fmt.Printf("Original:  %s\n", query2)
	fmt.Printf("Modified:  %s\n", modified2)

	// Example 3: Parse query with filters and modify aggregator
	fmt.Println("\nExample 3: Parse query with filters and modify aggregator")
	query3 := "avg:system.cpu.idle{host:~web-.*,env:prod} by {host}"
	builder3, err := ddqb.FromQuery(query3)
	if err != nil {
		log.Fatalf("Failed to parse query: %v", err)
	}

	// Change aggregator and add a filter
	modified3, err := builder3.
		Aggregator("sum").
		Filter(ddqb.Filter("region").Equal("us-east-1")).
		Build()
	if err != nil {
		log.Fatalf("Failed to build modified query: %v", err)
	}
	fmt.Printf("Original:  %s\n", query3)
	fmt.Printf("Modified:  %s\n", modified3)

	// Example 4: Parse query with IN filter and modify
	fmt.Println("\nExample 4: Parse query with IN filter and modify")
	query4 := "sum:system.cpu.idle{host IN (web-1,web-2,web-3)}"
	builder4, err := ddqb.FromQuery(query4)
	if err != nil {
		log.Fatalf("Failed to parse query: %v", err)
	}

	// Add environment filter
	modified4, err := builder4.
		Filter(ddqb.Filter("env").Equal("prod")).
		GroupBy("host").
		Build()
	if err != nil {
		log.Fatalf("Failed to build modified query: %v", err)
	}
	fmt.Printf("Original:  %s\n", query4)
	fmt.Printf("Modified:  %s\n", modified4)
	// Note: The IN filter format in the output will be: host IN (web-1,web-2,web-3)

	// Example 5: Round-trip - parse, modify, and verify
	fmt.Println("\nExample 5: Round-trip - parse, modify, and verify")
	originalQuery := "avg(5m):system.cpu.idle{host:web-1} by {host}.fill(0)"
	builder5, err := ddqb.FromQuery(originalQuery)
	if err != nil {
		log.Fatalf("Failed to parse query: %v", err)
	}

	// Make several modifications
	roundTripQuery, err := builder5.
		TimeWindow("10m").
		Filter(ddqb.Filter("env").Equal("prod")).
		ApplyFunction(ddqb.Function("rollup").WithArgs("60", "avg")).
		Build()
	if err != nil {
		log.Fatalf("Failed to build modified query: %v", err)
	}
	fmt.Printf("Original:  %s\n", originalQuery)
	fmt.Printf("Modified:  %s\n", roundTripQuery)

	// Parse the modified query again to verify it works
	builder6, err := ddqb.FromQuery(roundTripQuery)
	if err != nil {
		log.Fatalf("Failed to parse modified query: %v", err)
	}
	rebuilt, err := builder6.Build()
	if err != nil {
		log.Fatalf("Failed to rebuild query: %v", err)
	}
	fmt.Printf("Rebuilt:    %s\n", rebuilt)
}
