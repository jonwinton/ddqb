# DDQB - DataDog Query Builder - Context

## Project Overview

DDQB (DataDog Query Builder) is a Go library that provides a fluent, chainable API for building DataDog metric queries programmatically. It serves as a companion library to [DDQP](https://github.com/jonwinton/ddqp) (DataDog Query Parser).

## Purpose

Building complex DataDog queries through string manipulation is error-prone and difficult to maintain. DDQB solves this by providing:

- A fluent, chainable API for query construction
- Type safety and validation
- Programmatic query building without string manipulation
- Integration with DDQP's parsing structures

## Architecture

### Package Structure

```
ddqb/
├── ddqb.go              # Main package entry point with convenience functions
├── builder.go           # Base interfaces (Builder, Renderer)
├── metric/              # Core metric query building logic
│   ├── metric.go        # MetricQueryBuilder implementation
│   ├── filter.go        # FilterBuilder implementation
│   ├── function.go      # FunctionBuilder implementation
│   └── parser.go        # Query parsing using DDQP
└── examples/            # Usage examples
```

### Core Components

#### 1. MetricQueryBuilder (`metric/metric.go`)

The main builder interface for constructing DataDog metric queries. Supports:

- Setting metric names
- Setting aggregators (avg, sum, etc.)
- Setting time windows (e.g., "5m", "1h")
- Adding filters
- Grouping by dimensions
- Applying functions

**Key Methods:**

- `Metric(name string)` - Sets the metric name (required)
- `Aggregator(agg string)` - Sets aggregation method
- `TimeWindow(window string)` - Sets time window
- `Filter(filter FilterBuilder)` - Adds a filter condition
- `GroupBy(groups ...string)` - Sets grouping parameters
- `ApplyFunction(fn FunctionBuilder)` - Applies a function
- `Build() (string, error)` - Builds and returns the query string

#### 2. FilterBuilder (`metric/filter.go`)

Provides a fluent interface for building filter conditions. Supports:

**Operations:**

- `Equal(value string)` - Equality filter (key:value)
- `NotEqual(value string)` - Negated equality (key!:value)
- `Regex(pattern string)` - Regex filter (key:~pattern)
- `In(values ...string)` - IN filter (key IN (val1, val2))
- `NotIn(values ...string)` - NOT IN filter (key NOT IN (val1, val2))

#### 3. FunctionBuilder (`metric/function.go`)

Provides a fluent interface for building functions to apply to queries.

**Methods:**

- `WithArg(arg string)` - Adds a single argument
- `WithArgs(args ...string)` - Adds multiple arguments
- `Build() (string, error)` - Builds function string (format: `.function_name(arg1, arg2)`)

#### 4. Query Parser (`metric/parser.go`)

Parses existing DataDog query strings using DDQP and converts them into MetricQueryBuilder instances that can be modified.

**Key Function:**

- `ParseQuery(queryString string) (MetricQueryBuilder, error)` - Parses a query string and returns a builder

### Main Package API (`ddqb.go`)

Convenience functions that wrap the metric package:

- `Metric()` - Creates a new MetricQueryBuilder
- `Filter(key string)` - Creates a new FilterBuilder with the given key
- `Function(name string)` - Creates a new FunctionBuilder with the given name
- `FromQuery(queryString string)` - Parses an existing query and returns a builder

## Query Format

DDQB builds queries in DataDog's metric query format:

```
[aggregator([time_window]):]metric_name{filter1, filter2} by {group1, group2}.function1(args).function2(args)
```

**Components:**

- Aggregator: Optional (avg, sum, min, max, etc.)
- Time Window: Optional (e.g., "5m", "1h", "1d")
- Metric Name: Required
- Filters: Required (use `{*}` if no filters)
- Group By: Optional (`by {field1, field2}`)
- Functions: Optional (chained with `.function_name(args)`)

## Example Usage

```go
import "github.com/jonwinton/ddqb"

// Simple query
query, err := ddqb.Metric().
    Metric("system.cpu.idle").
    Build()
// Result: "system.cpu.idle{*}"

// Complex query
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
// Result: "avg(5m):system.cpu.idle{host:web-1, env:prod} by {host}.fill(0).rollup(60, sum)"

// Parse and modify existing query
builder, err := ddqb.FromQuery("avg(5m):system.cpu.idle{host:web-1}")
modifiedQuery, err := builder.TimeWindow("10m").Filter(ddqb.Filter("env").Equal("prod")).Build()
```

## Integration with DDQP

DDQB integrates with DDQP (DataDog Query Parser) to:

- Parse existing query strings into builder structures
- Convert DDQP's parsed structures into DDQB builders
- Enable round-trip conversion (parse → modify → build)

## Design Patterns

1. **Fluent/Builder Pattern**: All builders return themselves to enable method chaining
2. **Interface Segregation**: Separate interfaces for MetricQueryBuilder, FilterBuilder, FunctionBuilder
3. **Validation**: Build methods validate required fields and return errors
4. **Immutability**: Builders create new instances rather than mutating state (though current implementation may mutate - check code)

## Testing

The project includes comprehensive tests:

- Unit tests for each component (`*_test.go` files)
- Example tests (`example_test.go`)
- Format validation tests (`datadog_format_test.go`)
- Filter validation tests (`filter_validation_test.go`)

## Examples

See the `examples/` directory for:

- Basic usage (`examples/basic/`)
- Filter operations (`examples/filters/`)
- Function applications (`examples/functions/`)
- Advanced patterns (`examples/advanced/`)
- Query parsing (`examples/parse/`)

## Project Status

Initial development phase. The library focuses on metric queries and supports:

- ✅ Metric query building
- ✅ Filter operations (equal, not equal, regex, in, not in)
- ✅ Function application
- ✅ Query parsing from strings
- ⚠️ Aggregator functions wrapping queries (not yet supported per parser.go)

## License

Apache License 2.0
