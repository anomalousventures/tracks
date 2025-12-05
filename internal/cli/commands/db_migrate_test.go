package commands

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/anomalousventures/tracks/tests/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func setupDBMigrateTestCommand(t *testing.T) (*cobra.Command, *mocks.MockProjectDetector, *mocks.MockRenderer) {
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)
	mockRenderer.On("Title", mock.Anything).Return().Maybe()
	mockRenderer.On("Section", mock.Anything).Return().Maybe()
	mockRenderer.On("Flush").Return(nil).Maybe()

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {
		mockRenderer.Flush()
	}

	cmd := NewDBMigrateCommand(mockDetector, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	return cobraCmd, mockDetector, mockRenderer
}

func TestNewDBMigrateCommand(t *testing.T) {
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	cmd := NewDBMigrateCommand(mockDetector, factory, flusher)

	if cmd == nil {
		t.Fatal("NewDBMigrateCommand returned nil")
	}

	cobraCmd := cmd.Command()
	if cobraCmd == nil {
		t.Fatal("Command() returned nil - DI may have failed")
	}
}

func TestDBMigrateCommand_Command(t *testing.T) {
	cobraCmd, _, _ := setupDBMigrateTestCommand(t)

	if cobraCmd.Use != "migrate" {
		t.Errorf("expected Use 'migrate', got %q", cobraCmd.Use)
	}

	if cobraCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cobraCmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cobraCmd.Example == "" {
		t.Error("Example is empty")
	}

	if cobraCmd.RunE == nil {
		t.Error("RunE is nil, expected function")
	}
}

func TestDBMigrateCommand_Flags(t *testing.T) {
	cobraCmd, _, _ := setupDBMigrateTestCommand(t)

	stepsFlag := cobraCmd.Flag("steps")
	if stepsFlag == nil {
		t.Fatal("steps flag not found")
	}
	if stepsFlag.Shorthand != "n" {
		t.Errorf("steps shorthand expected 'n', got %q", stepsFlag.Shorthand)
	}
	if stepsFlag.DefValue != "0" {
		t.Errorf("steps default expected '0', got %q", stepsFlag.DefValue)
	}

	dryRunFlag := cobraCmd.Flag("dry-run")
	if dryRunFlag == nil {
		t.Fatal("dry-run flag not found")
	}
	if dryRunFlag.DefValue != "false" {
		t.Errorf("dry-run default expected 'false', got %q", dryRunFlag.DefValue)
	}
}

func TestDBMigrateCommand_NotInProject(t *testing.T) {
	cobraCmd, mockDetector, _ := setupDBMigrateTestCommand(t)

	mockDetector.On("Detect", mock.Anything, ".").
		Return(nil, "", errors.New("not found"))

	err := cobraCmd.Execute()

	if err == nil {
		t.Fatal("expected error for missing project")
	}
	if !strings.Contains(err.Error(), "not in a Tracks project directory") {
		t.Errorf("expected 'not in a Tracks project directory' error, got: %v", err)
	}
}

func TestDBMigrateCommand_UnsupportedDriver(t *testing.T) {
	cobraCmd, mockDetector, _ := setupDBMigrateTestCommand(t)

	mockDetector.On("Detect", mock.Anything, ".").
		Return(&interfaces.TracksProject{
			Name:       "testproject",
			ModulePath: "example.com/testproject",
			DBDriver:   "sqlite3",
		}, "/tmp/testproject", nil)

	err := cobraCmd.Execute()

	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
	if !strings.Contains(err.Error(), "only supports Postgres") {
		t.Errorf("expected 'only supports Postgres' error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "make migrate-up") {
		t.Errorf("expected helpful suggestion for make migrate-up, got: %v", err)
	}
}

func TestDBMigrateCommand_UnsupportedDriver_LibSQL(t *testing.T) {
	cobraCmd, mockDetector, _ := setupDBMigrateTestCommand(t)

	mockDetector.On("Detect", mock.Anything, ".").
		Return(&interfaces.TracksProject{
			Name:       "testproject",
			ModulePath: "example.com/testproject",
			DBDriver:   "go-libsql",
		}, "/tmp/testproject", nil)

	err := cobraCmd.Execute()

	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
	if !strings.Contains(err.Error(), "only supports Postgres") {
		t.Errorf("expected 'only supports Postgres' error, got: %v", err)
	}
}

func TestDBMigrateCommand_CommandDescriptions(t *testing.T) {
	cobraCmd, _, _ := setupDBMigrateTestCommand(t)

	keyPhrases := []string{"migration", "pending"}
	for _, phrase := range keyPhrases {
		if !strings.Contains(strings.ToLower(cobraCmd.Long), phrase) {
			t.Errorf("Long description missing mention of %q", phrase)
		}
	}

	if !strings.Contains(cobraCmd.Example, "--steps") {
		t.Error("Example missing --steps usage")
	}

	if !strings.Contains(cobraCmd.Example, "--dry-run") {
		t.Error("Example missing --dry-run usage")
	}
}
