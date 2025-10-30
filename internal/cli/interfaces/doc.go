// Package interfaces contains interface definitions owned by the CLI commands.
//
// Following ADR-002, interfaces are defined by their consumers, not providers.
// The CLI commands are the consumers, so they define the interfaces they need.
// Provider packages (generator, validation) implement these interfaces.
//
// This pattern:
//   - Prevents import cycles (providers import interface package, not vice versa)
//   - Enables proper dependency inversion (high-level doesn't depend on low-level)
//   - Improves testability (easy to mock from consumer's perspective)
//   - Follows Go idiom: "accept interfaces, return structs"
//
// Example:
//
//	// Consumer defines what it needs
//	type Validator interface {
//	    ValidateProjectName(string) error
//	}
//
//	// Provider implements the interface
//	// internal/validation/validator.go
//	type Validator struct { ... }
//	func (v *Validator) ValidateProjectName(name string) error { ... }
//
//	// Command uses the interface
//	// internal/cli/commands/new.go
//	type NewCommand struct {
//	    validator interfaces.Validator  // Depends on interface, not implementation
//	}
//
// See ADR-002 for the full decision rationale.
package interfaces
