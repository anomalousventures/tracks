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

func setupDBRollbackWithMockedDB(t *testing.T) (*cobra.Command, *mocks.MockProjectDetector, *mocks.MockDatabaseManager, *mocks.MockRenderer) {
	mockDetector := mocks.NewMockProjectDetector(t)
	mockDBManager := mocks.NewMockDatabaseManager(t)
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
	dbFactory := func(_ string) interfaces.DatabaseManager {
		return mockDBManager
	}

	cmd := NewDBRollbackCommandWithFactory(mockDetector, factory, flusher, dbFactory)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	return cobraCmd, mockDetector, mockDBManager, mockRenderer
}

func TestDBRollbackCommand_LoadEnvError(t *testing.T) {
	cobraCmd, mockDetector, mockDBManager, _ := setupDBRollbackWithMockedDB(t)

	mockDetector.On("Detect", mock.Anything, ".").
		Return(&interfaces.TracksProject{
			Name:       "testproject",
			ModulePath: "example.com/testproject",
			DBDriver:   "postgres",
		}, "/tmp/testproject", nil)

	mockDBManager.On("LoadEnv", mock.Anything, "/tmp/testproject").
		Return(errors.New("env file not found"))

	err := cobraCmd.Execute()

	if err == nil {
		t.Fatal("expected error for LoadEnv failure")
	}
	if !strings.Contains(err.Error(), "failed to load environment") {
		t.Errorf("expected 'failed to load environment' error, got: %v", err)
	}
}

func TestDBRollbackCommand_EmptyDatabaseURL(t *testing.T) {
	cobraCmd, mockDetector, mockDBManager, _ := setupDBRollbackWithMockedDB(t)

	mockDetector.On("Detect", mock.Anything, ".").
		Return(&interfaces.TracksProject{
			Name:       "testproject",
			ModulePath: "example.com/testproject",
			DBDriver:   "postgres",
		}, "/tmp/testproject", nil)

	mockDBManager.On("LoadEnv", mock.Anything, "/tmp/testproject").Return(nil)
	mockDBManager.On("GetDatabaseURL").Return("")

	err := cobraCmd.Execute()

	if err == nil {
		t.Fatal("expected error for empty DATABASE_URL")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL is not set") {
		t.Errorf("expected 'DATABASE_URL is not set' error, got: %v", err)
	}
}

func TestDBRollbackCommand_ConnectError(t *testing.T) {
	cobraCmd, mockDetector, mockDBManager, _ := setupDBRollbackWithMockedDB(t)

	mockDetector.On("Detect", mock.Anything, ".").
		Return(&interfaces.TracksProject{
			Name:       "testproject",
			ModulePath: "example.com/testproject",
			DBDriver:   "postgres",
		}, "/tmp/testproject", nil)

	mockDBManager.On("LoadEnv", mock.Anything, "/tmp/testproject").Return(nil)
	mockDBManager.On("GetDatabaseURL").Return("postgres://localhost/test")
	mockDBManager.On("Connect", mock.Anything).Return(nil, errors.New("connection refused"))

	err := cobraCmd.Execute()

	if err == nil {
		t.Fatal("expected error for Connect failure")
	}
	if !strings.Contains(err.Error(), "failed to connect to database") {
		t.Errorf("expected 'failed to connect to database' error, got: %v", err)
	}
}
