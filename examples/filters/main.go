package main

import (
	"fmt"
	"log"

	"github.com/jonwinton/ddqb"
)

func main() {
	fmt.Println("DataDog Query Builder - Filter Examples")
	fmt.Println("=======================================")

	// Example 1: Equal filter
	fmt.Println("Example 1: Equal filter")
	query, err := ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(ddqb.Filter("host").Equal("web-1")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 2: Not Equal filter
	fmt.Println("Example 2: Not Equal filter")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(ddqb.Filter("host").NotEqual("web-1")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 3: Regex filter
	fmt.Println("Example 3: Regex filter")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(ddqb.Filter("host").Regex("web-.*")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 4: IN filter
	fmt.Println("Example 4: IN filter")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(ddqb.Filter("host").In("web-1", "web-2", "web-3")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 5: NOT IN filter
	fmt.Println("Example 5: NOT IN filter")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(ddqb.Filter("host").NotIn("db-1", "db-2")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 6: Multiple filters (combined with AND)
	fmt.Println("Example 6: Multiple filters (combined with AND)")
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(ddqb.Filter("host").Regex("web-.*")).
		Filter(ddqb.Filter("env").Equal("prod")).
		Build()

	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)
}