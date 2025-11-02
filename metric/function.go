package metric

import (
	"fmt"
	"strings"
)

// FunctionBuilder provides a fluent interface for building functions to apply to queries.
type FunctionBuilder interface {
	// WithArg adds an argument to the function.
	WithArg(arg string) FunctionBuilder

	// WithArgs adds multiple arguments to the function.
	WithArgs(args ...string) FunctionBuilder

	// Build returns the built function as a string.
	Build() (string, error)
}

// functionBuilder is the concrete implementation of the FunctionBuilder interface.
type functionBuilder struct {
	name string
	args []string
}

// NewFunctionBuilder creates a new function builder with the given name.
func NewFunctionBuilder(name string) FunctionBuilder {
	return &functionBuilder{
		name: name,
		args: make([]string, 0),
	}
}

// WithArg adds an argument to the function.
func (b *functionBuilder) WithArg(arg string) FunctionBuilder {
	b.args = append(b.args, arg)
	return b
}

// WithArgs adds multiple arguments to the function.
func (b *functionBuilder) WithArgs(args ...string) FunctionBuilder {
	b.args = append(b.args, args...)
	return b
}

// Build returns the built function as a string.
func (b *functionBuilder) Build() (string, error) {
	if b.name == "" {
		return "", fmt.Errorf("function name is required")
	}

	// Format: .function_name(arg1, arg2, ...)
	if len(b.args) > 0 {
		return fmt.Sprintf(".%s(%s)", b.name, strings.Join(b.args, ", ")), nil
	}

	// Format: .function_name()
	return fmt.Sprintf(".%s()", b.name), nil
}
