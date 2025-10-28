// Package commands contains Cobra command implementations with dependency injection.
//
// Each command is implemented as a struct with injected dependencies, following
// the dependency injection pattern for testability:
//
//	type NewCommand struct {
//	    validator interfaces.Validator
//	    generator interfaces.ProjectGenerator
//	}
//
//	func NewNewCommand(v interfaces.Validator, g interfaces.ProjectGenerator) *NewCommand {
//	    return &NewCommand{validator: v, generator: g}
//	}
//
//	func (c *NewCommand) Command() *cobra.Command {
//	    // Returns configured cobra.Command
//	}
//
// This pattern allows:
//   - Easy mocking of dependencies in tests
//   - Clear declaration of what each command needs
//   - Compile-time verification of dependencies
//   - Reusable command instances
package commands
