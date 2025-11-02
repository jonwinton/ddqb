# Examples

Two ways to use DDQB:

1) Building queries from scratch
2) Parsing and updating existing queries

## 1) Building queries from scratch

### A. Simple metric

```go
q, err := ddqb.Metric().
  Metric("system.cpu.idle").
  Build()
```

### B. Aggregation + time window

```go
q, err := ddqb.Metric().
  Aggregator("avg").
  TimeWindow("5m").
  Metric("system.cpu.idle").
  Build()
```

### C. Basic filters and grouping

```go
q, err := ddqb.Metric().
  Metric("system.cpu.user").
  Filter(ddqb.Filter("env").Equal("prod")).
  Filter(ddqb.Filter("service").Equal("api")).
  GroupBy("host", "availability-zone").
  Build()
```

### D. Regex, IN/NOT IN, and NotEqual filters

```go
q, err := ddqb.Metric().
  Metric("system.disk.in_use").
  Filter(ddqb.Filter("device").Regex("/dev/nvme.*")).
  Filter(ddqb.Filter("env").NotEqual("staging")).
  Filter(ddqb.Filter("availability-zone").In("us-east-1a", "us-east-1b")).
  Filter(ddqb.Filter("service").NotIn("batch", "etl")).
  GroupBy("device", "host").
  Build()
```

### E. Functions (fill, rollup, as_rate, as_count)

```go
q, err := ddqb.Metric().
  Metric("nginx.net.request_per_s").
  ApplyFunction(ddqb.Function("fill").WithArg("0")).
  ApplyFunction(ddqb.Function("rollup").WithArgs("60", "sum")).
  ApplyFunction(ddqb.Function("as_rate")).
  Build()
```

### F. Explicit boolean logic with filter groups

```go
group := ddqb.FilterGroup().
  OR(ddqb.Filter("availability-zone").Equal("us-east-1a")).
  OR(ddqb.Filter("availability-zone").Equal("us-east-1c"))

q, err := ddqb.Metric().
  Metric("system.cpu.user").
  Filter(ddqb.Filter("env").Equal("staging")).
  Filter(group).
  GroupBy("availability-zone").
  Build()
```

### G. Advanced combined example

```go
complexGroup := ddqb.FilterGroup().
  AND(ddqb.Filter("team").Equal("core")).
  AND(
    ddqb.FilterGroup().
      OR(ddqb.Filter("service").Equal("api")).
      OR(ddqb.Filter("service").Equal("web")),
  )

q, err := ddqb.Metric().
  Aggregator("avg").
  TimeWindow("10m").
  Metric("system.mem.used").
  Filter(ddqb.Filter("env").Equal("prod")).
  Filter(ddqb.Filter("region").In("us-east-1", "us-west-2")).
  Filter(complexGroup).
  GroupBy("service", "host").
  ApplyFunction(ddqb.Function("fill").WithArg("0")).
  ApplyFunction(ddqb.Function("rollup").WithArgs("300", "avg")).
  Build()
```

## 2) Parsing and updating queries

### A. Simple parse and tweak

```go
b, err := ddqb.FromQuery("avg(5m):system.cpu.idle{host:web-1}")
if err != nil { panic(err) }

q, err := b.
  TimeWindow("10m").
  Filter(ddqb.Filter("env").Equal("prod")).
  Build()
```

### B. Add grouping and functions

```go
b, err := ddqb.FromQuery("sum:nginx.net.request_per_s{service:shop} by {resource_name}")
if err != nil { panic(err) }

q, err := b.
  GroupBy("resource_name", "host").
  ApplyFunction(ddqb.Function("as_rate")).
  ApplyFunction(ddqb.Function("rollup").WithArgs("120", "sum")).
  Build()
```

### C. Modify boolean groups in-place

```go
package main

import (
  "strings"
  "github.com/jonwinton/ddqb"
  "github.com/jonwinton/ddqb/metric"
)

func main() {
  // Start with a query using an OR group across AZs
  start := "avg:system.cpu.user{env:staging AND (availability-zone:us-east-1a OR availability-zone:us-east-1c)} by {availability-zone}"
  b, err := ddqb.FromQuery(start)
  if err != nil { panic(err) }

  // Find the OR group and add another AZ to it
  g := b.FindGroup(func(gr metric.FilterGroupBuilder) bool {
    built, _ := gr.Build()
    return strings.Contains(built, "availability-zone")
  })

  b = b.AddToGroup(g, ddqb.Filter("availability-zone").Equal("us-east-1b"))

  q, err := b.Build()
  if err != nil { panic(err) }
  _ = q
}
```

### D. Complex end-to-end edits

```go
original := "avg(5m):system.mem.used{env:prod, team:core} by {service}.fill(0).rollup(60, sum)"
b, err := ddqb.FromQuery(original)
if err != nil { panic(err) }

// Adjust aggregator window, add filters, and expand grouping
q, err := b.
  TimeWindow("15m").
  Filter(ddqb.Filter("region").In("us-east-1", "us-west-2")).
  GroupBy("service", "host").
  ApplyFunction(ddqb.Function("as_rate")).
  Build()
```

