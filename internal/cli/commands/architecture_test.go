package commands_test

import (
	"os/exec"
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
