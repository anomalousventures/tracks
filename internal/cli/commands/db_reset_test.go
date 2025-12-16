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

func setupDBResetTestCommand(t *testing.T) (*cobra.Command, *mocks.MockProjectDetector, *mocks.MockRenderer) {
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

	cmd := NewDBResetCommand(mockDetector, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	return cobraCmd, mockDetector, mockRenderer
}

func TestNewDBResetCommand(t *testing.T) {
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	cmd := NewDBResetCommand(mockDetector, factory, flusher)

	if cmd == nil {
		t.Fatal("NewDBResetCommand returned nil")
	}

	cobraCmd := cmd.Command()
	if cobraCmd == nil {
		t.Fatal("Command() returned nil - DI may have failed")
	}
}

func TestDBResetCommand_Command(t *testing.T) {
	cobraCmd, _, _ := setupDBResetTestCommand(t)

	if cobraCmd.Use != "reset" {
		t.Errorf("expected Use 'reset', got %q", cobraCmd.Use)
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

	forceFlag := cobraCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("--force flag is missing")
	}
}

func TestDBResetCommand_NotInProject(t *testing.T) {
	cobraCmd, mockDetector, _ := setupDBResetTestCommand(t)
	cobraCmd.SetArgs([]string{"--force"})

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

func TestDBResetCommand_UnsupportedDriver(t *testing.T) {
	cobraCmd, mockDetector, _ := setupDBResetTestCommand(t)
	cobraCmd.SetArgs([]string{"--force"})

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
	if !strings.Contains(err.Error(), "make migrate-reset") {
		t.Errorf("expected helpful suggestion for make migrate-reset, got: %v", err)
	}
}

func TestDBResetCommand_UnsupportedDriver_LibSQL(t *testing.T) {
	cobraCmd, mockDetector, _ := setupDBResetTestCommand(t)
	cobraCmd.SetArgs([]string{"--force"})

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

func TestDBResetCommand_CommandDescriptions(t *testing.T) {
	cobraCmd, _, _ := setupDBResetTestCommand(t)

	keyPhrases := []string{"reset", "drop", "warning"}
	for _, phrase := range keyPhrases {
		if !strings.Contains(strings.ToLower(cobraCmd.Long), phrase) {
			t.Errorf("Long description missing mention of %q", phrase)
		}
	}

	if !strings.Contains(cobraCmd.Example, "--force") {
		t.Error("Example missing --force usage")
	}
}

func TestDBResetCommand_ConfirmationDeclined(t *testing.T) {
	mockDetector := mocks.NewMockProjectDetector(t)
	mockDetector.On("Detect", mock.Anything, ".").
		Return(&interfaces.TracksProject{
			Name:       "testproject",
			ModulePath: "example.com/testproject",
			DBDriver:   "postgres",
		}, "/tmp/testproject", nil)

	mockDBManager := mocks.NewMockDatabaseManager(t)
	mockDBManager.On("LoadEnv", mock.Anything, "/tmp/testproject").Return(nil)
	mockDBManager.On("GetDatabaseURL").Return("postgres://localhost/test")

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

	cmd := NewDBResetCommandWithFactory(mockDetector, factory, flusher, dbFactory)
	cobraCmd := cmd.Command()

	outBuf := new(bytes.Buffer)
	cobraCmd.SetOut(outBuf)
	cobraCmd.SetErr(new(bytes.Buffer))
	cobraCmd.SetIn(strings.NewReader("n\n"))

	err := cobraCmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "Reset cancelled") {
		t.Errorf("expected 'Reset cancelled' message, got: %s", output)
	}
}

func TestDBResetCommand_ForceSkipsConfirmation(t *testing.T) {
	cobraCmd, mockDetector, mockDBManager, _ := setupDBResetWithMockedDB(t)
	cobraCmd.SetArgs([]string{"--force"})

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

func setupDBResetWithMockedDB(t *testing.T) (*cobra.Command, *mocks.MockProjectDetector, *mocks.MockDatabaseManager, *mocks.MockRenderer) {
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

	cmd := NewDBResetCommandWithFactory(mockDetector, factory, flusher, dbFactory)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	return cobraCmd, mockDetector, mockDBManager, mockRenderer
}

func TestDBResetCommand_LoadEnvError(t *testing.T) {
	cobraCmd, mockDetector, mockDBManager, _ := setupDBResetWithMockedDB(t)
	cobraCmd.SetArgs([]string{"--force"})

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

func TestDBResetCommand_EmptyDatabaseURL(t *testing.T) {
	cobraCmd, mockDetector, mockDBManager, _ := setupDBResetWithMockedDB(t)
	cobraCmd.SetArgs([]string{"--force"})

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

func TestDBResetCommand_ConnectError(t *testing.T) {
	cobraCmd, mockDetector, mockDBManager, _ := setupDBResetWithMockedDB(t)
	cobraCmd.SetArgs([]string{"--force"})

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
