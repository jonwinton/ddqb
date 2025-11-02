# DDQB - Datadog Query Builder

DDQB is a Go library that provides a fluent, chainable API for building Datadog queries programmatically. It's designed as a companion to [DDQP](https://github.com/jonwinton/ddqp) (Datadog Query Parser).

## Overview

Building complex Datadog queries through string manipulation can be error-prone and difficult to maintain. DDQB aims to solve this by providing:

- A fluent, chainable API for query construction
- Type safety and validation
- Programmatic query building without string manipulation
- Integration with DDQP's parsing structures

## Usage

```go
import "github.com/jonwinton/ddqb"

// Create a simple metric query
query, err := ddqb.Metric().
    Metric("system.cpu.idle").
    Build()

// Create a metric query with aggregation
query, err := ddqb.Metric().
    Aggregator("avg").
    TimeWindow("5m").
    Metric("system.cpu.idle").
    Build()

// Create a query with filters
query, err := ddqb.Metric().
    Metric("system.cpu.idle").
    Filter(ddqb.Filter("host").Equal("web-1")).
    Filter(ddqb.Filter("env").Equal("prod")).
    Build()

// Create a complex metric query
query, err := ddqb.Metric().
    Aggregator("avg").
    TimeWindow("5m").
    Metric("system.cpu.idle").
    Filter(ddqb.Filter("host").Equal("web-1")).
    Filter(ddqb.Filter("env").Equal("prod")).
    GroupBy("host").
    ApplyFunction(ddqb.Function("fill").WithArg("0")).
    ApplyFunction(ddqb.Function("rollup").WithArgs("60", "sum")).
    Build()
```

## Features

### Metric Queries

- Set metrics with `Metric(name)`
- Use aggregators with `Aggregator(agg)`
- Define time windows with `TimeWindow(window)`
- Add filters with `Filter(filterBuilder)`
- Group by dimensions with `GroupBy(fields...)`
- Apply functions with `ApplyFunction(functionBuilder)`

### Filters

- Equal: `Filter("host").Equal("web-1")`
- Not Equal: `Filter("host").NotEqual("web-1")`
- Regex: `Filter("host").Regex("web-.*")`
- In: `Filter("host").In("web-1", "web-2", "web-3")`
- Not In: `Filter("host").NotIn("db-1", "db-2")`

### Functions

- Apply functions with arguments:
  ```go
  Function("fill").WithArg("0")
  Function("rollup").WithArgs("60", "sum")
  ```

## Project Status

This project is in the initial development phase. Contributions and feedback are welcome!

## License

Apache License 2.0

## Documentation

This repo uses **MkDocs + Material** for docs.

- Local preview:

  ```bash
  pip install -r requirements.txt
  mkdocs serve
  ```

- Using just:
  ```bash
  just docs-serve
  just docs-build
  just docs-deploy
  just docs-api
  # list available tasks
  just help
  ```

* Manual deploy without just (optional):

  ```bash
  mkdocs gh-deploy --force
  ```

* Published site: https://jonwinton.github.io/ddqb
