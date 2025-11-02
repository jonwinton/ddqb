// Package ddqb provides a fluent API for building Datadog queries.
package ddqb

// Builder is the base interface for all query builders.
// It defines the methods that all builders must implement.
type Builder interface {
	// Build returns the built query as a string.
	Build() (string, error)
}

// Renderer defines an interface for objects that can render themselves as Datadog query strings.
type Renderer interface {
	// String returns the object as a Datadog query string.
	String() string
}
