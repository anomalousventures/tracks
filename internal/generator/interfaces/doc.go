// Package interfaces defines consumer-owned interfaces for generator components.
//
// Per ADR-002, interfaces should be defined in the consumer package, not the
// provider package. This enables dependency inversion and loose coupling.
//
// The internal/generator package consumes interfaces defined here, while
// provider packages (such as internal/generator/template) implement these
// interfaces. This pattern prevents import cycles and follows Go best practices.
package interfaces
