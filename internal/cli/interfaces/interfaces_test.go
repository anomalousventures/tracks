package interfaces_test

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/rs/zerolog"
)

// TestInterfacesPackageOnlyContainsInterfaces verifies that the interfaces
// package only contains interface type definitions, not concrete structs.
func TestInterfacesPackageOnlyContainsInterfaces(t *testing.T) {
	interfacesPath := "."
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, interfacesPath, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go") && fi.Name() != "doc.go"
	}, 0)
	if err != nil {
		t.Fatalf("Failed to parse interfaces package: %v", err)
	}

	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				if ts, ok := n.(*ast.TypeSpec); ok {
					if _, isStruct := ts.Type.(*ast.StructType); isStruct {
						t.Errorf("Found struct type %s in %s - interfaces package should only contain interfaces per ADR-002",
							ts.Name.Name, filename)
					}
				}
				return true
			})
		}
	}
}

// TestValidatorInterfaceSatisfaction verifies that the generator's validator
// implementation satisfies the Validator interface at compile time.
func TestValidatorInterfaceSatisfaction(t *testing.T) {
	logger := zerolog.New(os.Stderr).Level(zerolog.Disabled)
	var _ interfaces.Validator = generator.NewValidator(logger)
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
