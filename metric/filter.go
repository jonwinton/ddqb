package metric

import (
	"fmt"
	"strings"
)

// FilterOperation represents the type of filter operation.
type FilterOperation int

const (
	// Equal represents an equality filter (key:value).
	Equal FilterOperation = iota
	// NotEqual represents a negated equality filter (key!:value).
	NotEqual
	// GreaterThan represents a greater than filter (key>value).
	GreaterThan
	// LessThan represents a less than filter (key<value).
	LessThan
	// Regex represents a regex filter (key:~value).
	Regex
	// In represents an IN filter.
	In
	// NotIn represents a NOT IN filter.
	NotIn
)

// FilterBuilder provides a fluent interface for building filter conditions.
type FilterBuilder interface {
	// Equal creates an equality filter (key:value).
	Equal(value string) FilterBuilder

	// NotEqual creates a negated equality filter (key!:value).
	NotEqual(value string) FilterBuilder

	// GreaterThan creates a greater than filter (key>value).
	GreaterThan(value string) FilterBuilder

	// LessThan creates a less than filter (key<value).
	LessThan(value string) FilterBuilder

	// Regex creates a regex filter (key:~value).
	Regex(pattern string) FilterBuilder

	// In creates an IN filter.
	In(values ...string) FilterBuilder

	// NotIn creates a NOT IN filter.
	NotIn(values ...string) FilterBuilder

	// Build returns the built filter as a string.
	Build() (string, error)
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

// NotEqual creates a negated equality filter (key!:value).
func (b *filterBuilder) NotEqual(value string) FilterBuilder {
	b.operation = NotEqual
	b.values = []string{value}
	return b
}

// GreaterThan creates a greater than filter (key>value).
func (b *filterBuilder) GreaterThan(value string) FilterBuilder {
	b.operation = GreaterThan
	b.values = []string{value}
	return b
}

// LessThan creates a less than filter (key<value).
func (b *filterBuilder) LessThan(value string) FilterBuilder {
	b.operation = LessThan
	b.values = []string{value}
	return b
}

// Regex creates a regex filter (key:~value).
func (b *filterBuilder) Regex(pattern string) FilterBuilder {
	b.operation = Regex
	b.values = []string{pattern}
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
		return fmt.Sprintf("%s!:%s", b.key, b.values[0]), nil
	case GreaterThan:
		if len(b.values) != 1 {
			return "", fmt.Errorf("greater than filter requires exactly one value")
		}
		return fmt.Sprintf("%s>%s", b.key, b.values[0]), nil
	case LessThan:
		if len(b.values) != 1 {
			return "", fmt.Errorf("less than filter requires exactly one value")
		}
		return fmt.Sprintf("%s<%s", b.key, b.values[0]), nil
	case Regex:
		if len(b.values) != 1 {
			return "", fmt.Errorf("regex filter requires exactly one pattern")
		}
		return fmt.Sprintf("%s:~%s", b.key, b.values[0]), nil
	case In:
		if len(b.values) == 0 {
			return "", fmt.Errorf("in filter requires at least one value")
		}
		valueList := strings.Join(quoteValues(b.values), ", ")
		return fmt.Sprintf("%s IN [%s]", b.key, valueList), nil
	case NotIn:
		if len(b.values) == 0 {
			return "", fmt.Errorf("not in filter requires at least one value")
		}
		valueList := strings.Join(quoteValues(b.values), ", ")
		return fmt.Sprintf("%s NOT IN [%s]", b.key, valueList), nil
	default:
		return "", fmt.Errorf("unknown filter operation")
	}
}

// quoteValues adds quotes around string values for IN/NOT IN filters
func quoteValues(values []string) []string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("\"%s\"", v)
	}
	return quoted
}