package interfaces_test

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/anomalousventures/tracks/internal/validation"
)

// TestInterfacesPackageOnlyContainsInterfaces verifies that newly added interface
// files only contain interface type definitions, not concrete implementations.
// Note: renderer.go and buildinfo.go pre-date this test and contain helper types.
func TestInterfacesPackageOnlyContainsInterfaces(t *testing.T) {
	// Files added as part of Epic 0.5 Phase 1 (issues #159, #160)
	filesToCheck := []string{"validator.go", "generator.go"}

	for _, filename := range filesToCheck {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, filename, nil, 0)
		if err != nil {
			t.Fatalf("Failed to parse %s: %v", filename, err)
		}

		ast.Inspect(file, func(n ast.Node) bool {
			if ts, ok := n.(*ast.TypeSpec); ok {
				if _, isStruct := ts.Type.(*ast.StructType); isStruct {
					t.Errorf("Found struct type %s in %s - new interface files should only contain interfaces per ADR-002",
						ts.Name.Name, filename)
				}
			}
			return true
		})
	}
}

// TestValidatorInterfaceSatisfaction verifies that the validation package's validator
// implementation satisfies the Validator interface at compile time.
func TestValidatorInterfaceSatisfaction(t *testing.T) {
	var _ interfaces.Validator = validation.NewValidator()
	t.Log("✓ Validator implementation satisfies interfaces.Validator")
}

// TestProjectGeneratorInterfaceExists verifies the ProjectGenerator interface
// is defined and can be implemented.
func TestProjectGeneratorInterfaceExists(t *testing.T) {
	var _ interfaces.ProjectGenerator = (*mockGenerator)(nil)
	t.Log("✓ ProjectGenerator interface exists and can be implemented")
}

// mockGenerator is a test double for ProjectGenerator
type mockGenerator struct{}

func (m *mockGenerator) Generate(ctx context.Context, cfg any) error {
	return nil
}

func (m *mockGenerator) Validate(cfg any) error {
	return nil
}
