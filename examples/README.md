# DDQB Examples

This directory contains example Go programs that demonstrate how to use the DataDog Query Builder (DDQB) library.

## Running the Examples

To run any of the examples, use the `go run` command from within their respective directories:

```bash
# Run the basic usage example
cd basic && go run main.go

# Run the filter examples
cd filters && go run main.go

# Run the function examples
cd functions && go run main.go

# Run the advanced usage examples
cd advanced && go run main.go

# Run the parse and modify examples
cd parse && go run main.go
```

## Examples Overview

### 1. Basic Example

Demonstrates the core functionality of DDQB with several examples showing how to build various types of metric queries. This is a good starting point for understanding the library's capabilities.

### 2. Filter Examples

Focuses on the different filter operations supported by DDQB:

- Equal / Not Equal filters
- Regex filters
- IN / NOT IN filters
- Multiple filters combined

### 3. Function Examples

Showcases how to apply various functions to metric queries:

- fill()
- rollup()
- moving_average()
- timeshift()
- Multiple functions chained together

### 4. Advanced Examples

Demonstrates more complex use cases and patterns:

- Dynamic query building based on runtime conditions
- Utility functions for common query patterns
- Converting between formats (e.g., glob to regex)
- Handling different parameter types

### 5. Parse Examples

Demonstrates how to parse existing DataDog query strings and modify them:

- Parsing simple queries and adding filters
- Parsing complex queries and modifying components (time windows, aggregators, functions)
- Parsing queries with various filter types (regex, IN, NOT IN)
- Round-trip parsing and rebuilding queries

## Example Output

Each example will print the constructed query strings, allowing you to see exactly how DDQB translates builder method calls into DataDog query syntax.
