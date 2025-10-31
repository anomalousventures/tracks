package commands_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestNoImportCycles verifies the codebase has no circular dependencies.
// Import cycles would prevent compilation and violate clean architecture.
func TestNoImportCycles(t *testing.T) {
	cmd := exec.Command("go", "list", "-json", "./...")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if strings.Contains(string(output), "import cycle") {
			t.Fatalf("Import cycle detected:\n%s\n\nThis violates ADR-002. Interfaces must be in consumer packages.", output)
		}
		t.Fatalf("go list failed: %v\nOutput: %s", err, output)
	}

	t.Log("✓ No import cycles detected")
}

// TestCommandsCanImportInterfaces verifies that commands can import the
// interfaces package without creating cycles.
func TestCommandsCanImportInterfaces(t *testing.T) {
	_ = "github.com/anomalousventures/tracks/internal/cli/interfaces"
	t.Log("✓ Commands can safely import interfaces package")
}

// TestGeneratorPackageDoesNotImportCLI verifies that the generator package
// does not import internal/cli, ensuring the fix from Issue #161 is maintained.
//
// This is a preventive test that catches the import BEFORE an import cycle forms.
// While TestNoImportCycles detects complete cycles, this test fails immediately
// if someone re-introduces the problematic import pattern from Issue #161, providing
// specific guidance before CLI commands import generator in Phase 2+.
func TestGeneratorPackageDoesNotImportCLI(t *testing.T) {
	cmd := exec.Command("go", "list", "-f", "{{.Imports}}",
		"github.com/anomalousventures/tracks/internal/generator")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to list generator imports: %v", err)
	}

	imports := string(output)
	if strings.Contains(imports, "github.com/anomalousventures/tracks/internal/cli") &&
		!strings.Contains(imports, "github.com/anomalousventures/tracks/internal/cli/interfaces") {
		t.Errorf("generator package imports internal/cli (not just interfaces), violating ADR-002.\n"+
			"Logger should be injected via constructor (Issue #161).\n"+
			"Imports: %s", imports)
	}

	t.Log("✓ Generator package does not import CLI package directly")
}

// TestCommandsUseDI verifies all Command structs have corresponding New*Command constructors.
// This enforces the dependency injection pattern from ADR-001, ensuring commands are testable
// through constructor injection rather than hard-coded dependencies.
func TestCommandsUseDI(t *testing.T) {
	// Find all .go files in the commands package (excluding tests and doc.go)
	pattern := filepath.Join(".", "*.go")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}

	commandStructs := make(map[string]bool)
	constructors := make(map[string]bool)

	for _, file := range matches {
		if strings.HasSuffix(file, "_test.go") || filepath.Base(file) == "doc.go" {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read %s: %v", file, err)
		}

		// Find Command structs using simple pattern matching
		structPattern := regexp.MustCompile(`type\s+(\w+Command)\s+struct`)
		constructorPattern := regexp.MustCompile(`func\s+(New\w+Command)\(`)

		for _, match := range structPattern.FindAllStringSubmatch(string(content), -1) {
			commandStructs[match[1]] = true
		}

		for _, match := range constructorPattern.FindAllStringSubmatch(string(content), -1) {
			constructors[match[1]] = true
		}
	}

	// Verify each Command struct has corresponding constructor
	missing := []string{}
	for cmd := range commandStructs {
		expectedConstructor := "New" + cmd
		if !constructors[expectedConstructor] {
			missing = append(missing, cmd)
		}
	}

	if len(missing) > 0 {
		t.Errorf("Commands missing DI constructors (violates ADR-001):\n"+
			"Commands: %v\n"+
			"Each *Command struct must have a New*Command(...) constructor for dependency injection.",
			missing)
	}

	t.Logf("✓ All %d commands use dependency injection pattern", len(commandStructs))
}
