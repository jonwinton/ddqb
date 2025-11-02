package ddqb_test

import (
	"fmt"
	"log"

	"github.com/jonwinton/ddqb"
)

func Example() {
	// Create a simple metric query
	query, err := ddqb.Metric().
		Metric("system.cpu.idle").
		Build()
	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Println(query)

	// Create a metric query with aggregation
	query, err = ddqb.Metric().
		Aggregator("avg").
		TimeWindow("5m").
		Metric("system.cpu.idle").
		Build()
	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Println(query)

	// Create a metric query with filtering
	query, err = ddqb.Metric().
		Metric("system.cpu.idle").
		Filter(ddqb.Filter("host").Equal("web-1")).
		Build()
	if err != nil {
		log.Fatalf("Failed to build query: %v", err)
	}
	fmt.Println(query)

	// Create a complex metric query
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
	fmt.Println(query)

	// Output:
	// system.cpu.idle{*}
	// avg(5m):system.cpu.idle{*}
	// system.cpu.idle{host:web-1}
	// avg(5m):system.cpu.idle{host:web-1, env:prod} by {host}.fill(0).rollup(60, sum)
}
