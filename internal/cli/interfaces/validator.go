package interfaces

// Validator validates project configuration values.
//
// This interface is owned by the CLI commands package, following ADR-002:
// interfaces are defined by consumers, not providers. The validation
// implementation lives in internal/generator/validator_impl.go.
//
// This pattern prevents import cycles and enables proper dependency inversion:
// CLI (high-level) defines interface, validation (low-level) implements it.
//
// Example usage:
//
//	logger := zerolog.New(os.Stderr).Level(zerolog.InfoLevel)
//	validator := generator.NewValidator(logger)
//	if err := validator.ValidateProjectName("my-app"); err != nil {
//	    return err
//	}
type Validator interface {
	ValidateProjectName(name string) error
	ValidateModulePath(path string) error
	ValidateDirectory(path string) error
	ValidateDatabaseDriver(driver string) error
}
