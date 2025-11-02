package metric

import (
	"fmt"
	"strings"
)

// GroupOperator represents the boolean operator used in a filter group.
type GroupOperator int

const (
	// AndOperator represents an AND operation between filters.
	AndOperator GroupOperator = iota
	// OrOperator represents an OR operation between filters.
	OrOperator
)

// FilterGroupBuilder provides a fluent interface for building filter groups with boolean logic.
// FilterGroupBuilder implements FilterExpression.
type FilterGroupBuilder interface {
	FilterExpression

	// And adds a filter or nested group with AND operator.
	And(expr FilterExpression) FilterGroupBuilder

	// Or adds a filter or nested group with OR operator.
	Or(expr FilterExpression) FilterGroupBuilder

	// Not negates the entire group (wraps in NOT (...)).
	Not() FilterGroupBuilder
}

// filterGroupBuilder is the concrete implementation of the FilterGroupBuilder interface.
type filterGroupBuilder struct {
	expressions []FilterExpression
	operator    GroupOperator // The operator used in this group (AND or OR)
	negated     bool
}

// NewFilterGroupBuilder creates a new filter group builder.
func NewFilterGroupBuilder() FilterGroupBuilder {
	return &filterGroupBuilder{
		expressions: make([]FilterExpression, 0),
		operator:    AndOperator, // Default to AND
		negated:     false,
	}
}

// And adds a filter or nested group with AND operator.
// Sets the group operator to AND if this is the first expression added.
func (b *filterGroupBuilder) And(expr FilterExpression) FilterGroupBuilder {
	if len(b.expressions) == 0 {
		// First expression - set operator to AND
		b.operator = AndOperator
		b.expressions = append(b.expressions, expr)
	} else {
		// Mixing operators requires a nested group
		// For now, we'll allow it but users should use nested groups for clarity
		b.expressions = append(b.expressions, expr)
	}
	return b
}

// Or adds a filter or nested group with OR operator.
// Sets the group operator to OR if this is the first expression added.
func (b *filterGroupBuilder) Or(expr FilterExpression) FilterGroupBuilder {
	if len(b.expressions) == 0 {
		// First expression - set operator to OR
		b.operator = OrOperator
		b.expressions = append(b.expressions, expr)
	} else {
		// Mixing operators requires a nested group
		// For now, we'll allow it but users should use nested groups for clarity
		b.expressions = append(b.expressions, expr)
	}
	return b
}

// Not negates the entire group.
func (b *filterGroupBuilder) Not() FilterGroupBuilder {
	b.negated = true
	return b
}

// Build returns the built filter group as a string with proper parentheses and operators.
func (b *filterGroupBuilder) Build() (string, error) {
	if len(b.expressions) == 0 {
		return "", fmt.Errorf("filter group must contain at least one expression")
	}

	// Build all expressions
	var parts []string
	for _, expr := range b.expressions {
		filterStr, err := expr.Build()
		if err != nil {
			return "", fmt.Errorf("error building filter expression: %w", err)
		}
		parts = append(parts, filterStr)
	}

	// Join parts with the appropriate operator
	var opStr string
	if b.operator == AndOperator {
		opStr = " AND "
	} else {
		opStr = " OR "
	}

	groupStr := strings.Join(parts, opStr)

	// Wrap in parentheses if there are multiple expressions
	if len(b.expressions) > 1 {
		groupStr = fmt.Sprintf("(%s)", groupStr)
	}

	// Apply negation if needed
	if b.negated {
		groupStr = fmt.Sprintf("NOT %s", groupStr)
	}

	return groupStr, nil
}
