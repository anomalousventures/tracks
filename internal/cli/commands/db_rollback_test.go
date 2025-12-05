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

func setupDBRollbackTestCommand(t *testing.T) (*cobra.Command, *mocks.MockProjectDetector, *mocks.MockRenderer) {
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

	cmd := NewDBRollbackCommand(mockDetector, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	return cobraCmd, mockDetector, mockRenderer
}

func TestNewDBRollbackCommand(t *testing.T) {
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	cmd := NewDBRollbackCommand(mockDetector, factory, flusher)

	if cmd == nil {
		t.Fatal("NewDBRollbackCommand returned nil")
	}

	cobraCmd := cmd.Command()
	if cobraCmd == nil {
		t.Fatal("Command() returned nil - DI may have failed")
	}
}

func TestDBRollbackCommand_Command(t *testing.T) {
	cobraCmd, _, _ := setupDBRollbackTestCommand(t)

	if cobraCmd.Use != "rollback" {
		t.Errorf("expected Use 'rollback', got %q", cobraCmd.Use)
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

func TestDBRollbackCommand_Flags(t *testing.T) {
	cobraCmd, _, _ := setupDBRollbackTestCommand(t)

	stepsFlag := cobraCmd.Flag("steps")
	if stepsFlag == nil {
		t.Fatal("steps flag not found")
	}
	if stepsFlag.Shorthand != "n" {
		t.Errorf("steps shorthand expected 'n', got %q", stepsFlag.Shorthand)
	}
	if stepsFlag.DefValue != "1" {
		t.Errorf("steps default expected '1', got %q", stepsFlag.DefValue)
	}
}

func TestDBRollbackCommand_NotInProject(t *testing.T) {
	cobraCmd, mockDetector, _ := setupDBRollbackTestCommand(t)

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

func TestDBRollbackCommand_UnsupportedDriver(t *testing.T) {
	cobraCmd, mockDetector, _ := setupDBRollbackTestCommand(t)

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
	if !strings.Contains(err.Error(), "make migrate-down") {
		t.Errorf("expected helpful suggestion for make migrate-down, got: %v", err)
	}
}

func TestDBRollbackCommand_UnsupportedDriver_LibSQL(t *testing.T) {
	cobraCmd, mockDetector, _ := setupDBRollbackTestCommand(t)

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

func TestDBRollbackCommand_CommandDescriptions(t *testing.T) {
	cobraCmd, _, _ := setupDBRollbackTestCommand(t)

	keyPhrases := []string{"roll back", "migration"}
	for _, phrase := range keyPhrases {
		if !strings.Contains(strings.ToLower(cobraCmd.Long), phrase) {
			t.Errorf("Long description missing mention of %q", phrase)
		}
	}

	if !strings.Contains(cobraCmd.Example, "--steps") {
		t.Error("Example missing --steps usage")
	}
}
