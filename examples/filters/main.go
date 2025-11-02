// Package main demonstrates various filter operations available in ddqb for Datadog metric queries.
package main

import (
	"fmt"
	"log"

	"github.com/jonwinton/ddqb"
)

func main() {
	fmt.Println("Datadog Query Builder - Filter Examples")
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

	// Example 7: OR query
	fmt.Println("Example 7: OR query")
	orGroup := ddqb.FilterGroup()
	orGroup.Or(ddqb.Filter("env").Equal("prod"))
	orGroup.Or(ddqb.Filter("env").Equal("staging"))
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(orGroup).
		Build()
	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 8: AND NOT query
	fmt.Println("Example 8: AND NOT query")
	andNotGroup := ddqb.FilterGroup()
	andNotGroup.And(ddqb.Filter("env").Equal("prod"))
	notGroup := ddqb.FilterGroup()
	notGroup.And(ddqb.Filter("host").Equal("web-1"))
	notGroup.Not()
	andNotGroup.And(notGroup)
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(andNotGroup).
		Build()
	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 9: OR NOT query
	fmt.Println("Example 9: OR NOT query")
	orNotGroup := ddqb.FilterGroup()
	orNotGroup.Or(ddqb.Filter("env").Equal("prod"))
	notGroup2 := ddqb.FilterGroup()
	notGroup2.And(ddqb.Filter("host").Equal("web-1"))
	notGroup2.Not()
	orNotGroup.Or(notGroup2)
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(orNotGroup).
		Build()
	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 10: Nested groups (AND with nested OR)
	fmt.Println("Example 10: Nested groups (AND with nested OR)")
	outerGroup := ddqb.FilterGroup()
	outerGroup.And(ddqb.Filter("env").Equal("prod"))

	innerGroup := ddqb.FilterGroup()
	innerGroup.Or(ddqb.Filter("host").Equal("web-1"))
	innerGroup.Or(ddqb.Filter("host").Equal("web-2"))

	outerGroup.And(innerGroup)
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(outerGroup).
		Build()
	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 11: Complex nested groups
	fmt.Println("Example 11: Complex nested groups")
	envGroup := ddqb.FilterGroup()
	envGroup.Or(ddqb.Filter("env").Equal("prod"))
	envGroup.Or(ddqb.Filter("env").Equal("staging"))

	hostGroup := ddqb.FilterGroup()
	hostGroup.Or(ddqb.Filter("host").Regex("web-.*"))
	hostGroup.Or(ddqb.Filter("host").Regex("api-.*"))

	complexGroup := ddqb.FilterGroup()
	complexGroup.And(envGroup)
	complexGroup.And(hostGroup)
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(complexGroup).
		Build()
	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)

	// Example 12: Multiple groups combined (implicit AND)
	fmt.Println("Example 12: Multiple groups combined (implicit AND)")
	group1 := ddqb.FilterGroup()
	group1.Or(ddqb.Filter("env").Equal("prod"))
	group1.Or(ddqb.Filter("env").Equal("staging"))

	group2 := ddqb.FilterGroup()
	group2.Or(ddqb.Filter("region").Equal("us-east-1"))
	group2.Or(ddqb.Filter("region").Equal("us-west-2"))

	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(group1).
		Filter(group2).
		Build()
	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Printf("Query: %s\n\n", query)
}
