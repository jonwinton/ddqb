package metric

import (
	"fmt"
	"strings"
)

// FilterExpression is a common interface for both individual filters and filter groups.
// This allows QueryBuilder to accept either FilterBuilder or FilterGroupBuilder instances.
type FilterExpression interface {
	// Build returns the built filter expression as a string.
	Build() (string, error)
}

// FilterOperation represents the type of filter operation.
type FilterOperation int

const (
	// Equal represents an equality filter (key:value).
	Equal FilterOperation = iota
	// NotEqual represents a negated equality filter (!key:value).
	NotEqual
	// In represents an IN filter.
	In
	// NotIn represents a NOT IN filter.
	NotIn
)

// FilterBuilder provides a fluent interface for building filter conditions.
// FilterBuilder implements FilterExpression.
type FilterBuilder interface {
	FilterExpression

	// Equal creates an equality filter (key:value).
	Equal(value string) FilterBuilder

	// NotEqual creates a negated equality filter (!key:value).
	NotEqual(value string) FilterBuilder

	// In creates an IN filter.
	In(values ...string) FilterBuilder

	// NotIn creates a NOT IN filter.
	NotIn(values ...string) FilterBuilder
}

// filterBuilder is the concrete implementation of the FilterBuilder interface.
type filterBuilder struct {
	key       string
	operation FilterOperation // Defaults to an invalid value
	values    []string
}

// NewFilterBuilder creates a new filter builder with the given key.
func NewFilterBuilder(key string) FilterBuilder {
	return &filterBuilder{
		key:    key,
		values: make([]string, 0),
	}
}

// Equal creates an equality filter (key:value).
func (b *filterBuilder) Equal(value string) FilterBuilder {
	b.operation = Equal
	b.values = []string{value}
	return b
}

// NotEqual creates a negated equality filter (!key:value).
func (b *filterBuilder) NotEqual(value string) FilterBuilder {
	b.operation = NotEqual
	b.values = []string{value}
	return b
}

// In creates an IN filter.
func (b *filterBuilder) In(values ...string) FilterBuilder {
	b.operation = In
	b.values = values
	return b
}

// NotIn creates a NOT IN filter.
func (b *filterBuilder) NotIn(values ...string) FilterBuilder {
	b.operation = NotIn
	b.values = values
	return b
}

// Build returns the built filter as a string.
func (b *filterBuilder) Build() (string, error) {
	if b.key == "" {
		return "", fmt.Errorf("filter key is required")
	}

	switch b.operation {
	case Equal:
		if len(b.values) != 1 {
			return "", fmt.Errorf("equal filter requires exactly one value")
		}
		return fmt.Sprintf("%s:%s", b.key, b.values[0]), nil
	case NotEqual:
		if len(b.values) != 1 {
			return "", fmt.Errorf("not equal filter requires exactly one value")
		}
		return fmt.Sprintf("!%s:%s", b.key, b.values[0]), nil
	case In:
		if len(b.values) == 0 {
			return "", fmt.Errorf("in filter requires at least one value")
		}
		valueList := strings.Join(b.values, ",")
		return fmt.Sprintf("%s IN (%s)", b.key, valueList), nil
	case NotIn:
		if len(b.values) == 0 {
			return "", fmt.Errorf("not in filter requires at least one value")
		}
		valueList := strings.Join(b.values, ",")
		return fmt.Sprintf("%s NOT IN (%s)", b.key, valueList), nil
	default:
		return "", fmt.Errorf("unknown filter operation")
	}
}
