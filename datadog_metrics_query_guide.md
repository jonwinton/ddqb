# Datadog Metrics Query Syntax Guide for AI Agents

## Overview
This document provides a comprehensive guide to Datadog metrics query syntax, specifically designed for AI agents to generate accurate queries. All examples are based on official Datadog documentation.

## Basic Query Structure

### Standard Format
```
<AGGREGATOR>:<METRIC_NAME>{<TAG_FILTERS>} [by {<GROUP_BY_TAGS>}] [.<FUNCTION>()]
```

### Components Breakdown
- **AGGREGATOR**: How to combine data points (avg, sum, min, max, count)
- **METRIC_NAME**: The name of the metric (e.g., `system.cpu.user`, `aws.ec2.cpuutilization`)
- **TAG_FILTERS**: Conditions to filter the data (enclosed in curly braces)
- **GROUP_BY_TAGS**: Tags to group results by (optional, after `by` keyword)
- **FUNCTION**: Additional processing functions (optional, with dot notation)

## Aggregators

### Available Aggregators
- `avg` - Average of all values
- `sum` - Sum of all values  
- `min` - Minimum value
- `max` - Maximum value
- `count` - Count of data points

### Examples
```
avg:system.cpu.user{*}
sum:page.views{domain:example.com}
max:system.load.1{host:web-server}
```

## Tag Filtering

### Basic Tag Filtering
```
# Single tag filter
avg:system.cpu.user{env:production}

# Multiple tags (implicit AND)
avg:system.cpu.user{env:production,region:us-east-1}

# Wildcard - all tags
avg:system.cpu.user{*}
```

### Boolean Operators

#### AND Operator
```
# Explicit AND
avg:system.cpu.user{env:staging AND region:us-east-1}

# Implicit AND (comma-separated)
avg:system.cpu.user{env:staging,region:us-east-1}
```

#### OR Operator
```
# OR with parentheses for grouping
avg:system.cpu.user{env:staging AND (availability-zone:us-east-1a OR availability-zone:us-east-1c)} by {availability-zone}
```

#### IN Operator
```
# IN operator for multiple values
avg:system.cpu.user{env:shop.ist AND availability-zone IN (us-east-1a, us-east-1b, us-east4-b)} by {availability-zone}
```

#### NOT IN Operator
```
# NOT IN operator for exclusion
avg:system.cpu.user{env:prod AND location NOT IN (atlanta,seattle,las-vegas)}
```

#### NOT Operator (Exclusion)
```
# Exclude specific tag values
avg:system.disk.in_use{!device:/dev/loop*} by {device}
```

### Wildcard Filtering

#### Prefix Wildcard
```
# Services ending with "-canary"
sum:kubernetes.pods.running{service:*-canary} by {service}
```

#### Suffix Wildcard
```
# Devices starting with "/dev/loop" (excluded)
avg:system.disk.in_use{!device:/dev/loop*} by {device}
```

#### Infix Wildcard
```
# Regions containing "east" anywhere in name
avg:system.disk.utilized{region:*east*} by {region}
```

## Grouping (by clause)

### Single Tag Grouping
```
avg:system.cpu.user{env:production} by {host}
```

### Multiple Tag Grouping
```
avg:aws.ec2.cpuutilization{*} by {env,host}
```

### No Grouping
```
# Aggregate across all matching series
avg:system.cpu.user{env:production}
```

## Functions and Modifiers

### Rollup Function
```
# Custom time aggregation
avg:system.disk.free{*}.rollup(avg, 60)
avg:system.cpu.user{*}.rollup(sum, 300)
avg:system.load.1{*}.rollup(max, 120)
```

#### Rollup Parameters
- **Aggregator**: `avg`, `sum`, `min`, `max`, `count`
- **Interval**: Time in seconds (e.g., 60 for 1 minute)

### Type Modifiers

#### as_count()
```
# For COUNT and RATE metrics
sum:trace.rack.request.hits{service:web-store} by {resource_name}.as_count()
```

#### as_rate()
```
# Convert to rate per second
sum:my.counter.metric{*}.as_rate()
```

### Arithmetic Operations
```
# Division between metrics
jvm.heap_memory / jvm.heap_memory_max

# Arithmetic with constants
system.cpu.user * 100
```

### Moving Functions
```
# Moving rollup over time window
moving_rollup(system.cpu.user, 300, avg)
```

## Distribution Metrics

### Percentile Queries
```
# 99th percentile
p99:request_latency_distribution{app:A OR app:B} by {app}

# 95th percentile
p95:response_time{service:api}

# 50th percentile (median)
p50:request_duration{*}
```

### Available Percentiles
- `p50`, `p75`, `p90`, `p95`, `p99`, `p999`

## Advanced Examples

### Complex Boolean Logic
```
avg:system.cpu.user{env:staging AND (availability-zone:us-east-1a OR availability-zone:us-east-1c)} by {availability-zone}
```

### Multiple Conditions with IN
```
avg:system.cpu.user{env:shop.ist AND availability-zone IN (us-east-1a, us-east-1b, us-east4-b)} by {availability-zone}
```

### Exclusion with Wildcards
```
avg:system.disk.in_use{!device:/dev/loop*} by {device}
```

### Service Pattern Matching
```
sum:kubernetes.pods.running{service:*-canary} by {service}
```

### Geographic Filtering
```
avg:system.disk.utilized{region:*east*} by {region}
```

## Common Patterns

### Application Performance Monitoring (APM)
```
# Request count
sum:trace.rack.request.hits{service:web-store} by {resource_name}.as_count()

# Error count  
sum:trace.rack.request.errors{service:web-store} by {resource_name}.as_count()

# Response time percentiles
p99:trace.rack.request.duration{service:web-store} by {resource_name}
```

### Infrastructure Monitoring
```
# CPU utilization by host
avg:system.cpu.user{env:production} by {host}

# Memory usage
avg:system.mem.used{env:production} by {host}

# Disk usage by device
avg:system.disk.used{*} by {device,host}
```

### Kubernetes Metrics
```
# Pod count by service
sum:kubernetes.pods.running{*} by {service}

# Container CPU
avg:kubernetes.cpu.usage.total{*} by {pod_name}

# Container memory
avg:kubernetes.memory.usage{*} by {pod_name}
```

## Best Practices for AI Agents

### 1. Always Use Appropriate Aggregators
- Use `avg` for gauge metrics (CPU, memory usage)
- Use `sum` for count metrics (requests, errors)
- Use `max`/`min` for threshold monitoring

### 2. Tag Filtering Specificity
- Be as specific as possible with tag filters
- Use wildcards judiciously to avoid overly broad queries
- Prefer explicit tag names over wildcards when possible

### 3. Grouping Considerations
- Group by relevant dimensions for analysis
- Avoid grouping by high-cardinality tags unless necessary
- Consider the visualization context when choosing grouping

### 4. Function Usage
- Use `.as_count()` for COUNT and RATE type metrics
- Apply `.rollup()` for custom time aggregation
- Use percentiles (`p95`, `p99`) for latency metrics

### 5. Boolean Logic
- Use parentheses to clarify complex boolean expressions
- Prefer `IN` operator for multiple value matching
- Use `NOT IN` for exclusions rather than multiple `NOT` conditions

## Common Errors to Avoid

### 1. Incorrect Aggregator Usage
```
# WRONG: Using sum on gauge metrics
sum:system.cpu.user{*}

# CORRECT: Using avg on gauge metrics  
avg:system.cpu.user{*}
```

### 2. Missing Curly Braces
```
# WRONG: Missing tag filter braces
avg:system.cpu.user env:production

# CORRECT: Proper tag filter syntax
avg:system.cpu.user{env:production}
```

### 3. Improper Boolean Grouping
```
# WRONG: Ambiguous boolean logic
avg:system.cpu.user{env:staging AND zone:us-east-1a OR zone:us-east-1c}

# CORRECT: Clear grouping with parentheses
avg:system.cpu.user{env:staging AND (zone:us-east-1a OR zone:us-east-1c)}
```

### 4. Wildcard Overuse
```
# WRONG: Overly broad wildcard
avg:system.cpu.user{*} by {*}

# CORRECT: Specific filtering and grouping
avg:system.cpu.user{env:production} by {host}
```

## Validation Checklist

Before generating a Datadog metrics query, verify:

1. ✅ Aggregator matches metric type (avg for gauges, sum for counts)
2. ✅ Metric name is valid and exists
3. ✅ Tag filters use proper syntax with curly braces
4. ✅ Boolean operators are properly grouped with parentheses
5. ✅ Grouping tags are relevant and not too high-cardinality
6. ✅ Functions are appropriate for the metric type
7. ✅ Wildcard usage is intentional and not overly broad

## Example Query Patterns by Use Case

### Error Rate Monitoring
```
sum:trace.flask.request.errors{service:api} by {resource_name}.as_count() / sum:trace.flask.request.hits{service:api} by {resource_name}.as_count()
```

### Resource Utilization
```
avg:system.cpu.user{env:production} by {host}
avg:system.mem.pct_usable{env:production} by {host}
```

### Throughput Monitoring  
```
sum:nginx.net.request_per_s{*} by {server}
```

### Latency Monitoring
```
p95:trace.http.request.duration{service:web-app} by {resource_name}
```

This guide provides the foundation for generating accurate Datadog metrics queries. Always refer to the specific metric documentation for additional context and available tags.
