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

func setupUIUpgradeTestCommand(t *testing.T) (*cobra.Command, *mocks.MockUIExecutor, *mocks.MockProjectDetector, *mocks.MockRenderer) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {
		mockRenderer.Flush()
	}

	cmd := NewUIUpgradeCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	return cobraCmd, mockExecutor, mockDetector, mockRenderer
}

func TestUIUpgradeCommand_Command(t *testing.T) {
	cobraCmd, _, _, _ := setupUIUpgradeTestCommand(t)

	if cobraCmd == nil {
		t.Fatal("Command() returned nil")
	}

	if !strings.HasPrefix(cobraCmd.Use, "upgrade") {
		t.Errorf("expected Use to start with 'upgrade', got %q", cobraCmd.Use)
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
}

func TestUIUpgradeCommand_Run_NotInProject(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	mockDetector.On("Detect", mock.Anything, ".").
		Return(nil, "", nil).Once()

	mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
		return strings.Contains(s.Body, "not in a Tracks project")
	})).Once()
	mockRenderer.On("Flush").Return(nil).Once()

	factory := func(*cobra.Command) interfaces.Renderer { return mockRenderer }
	flusher := func(*cobra.Command, interfaces.Renderer) { mockRenderer.Flush() }

	cmd := NewUIUpgradeCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	mockDetector.AssertExpectations(t)
	mockRenderer.AssertExpectations(t)
}

func TestUIUpgradeCommand_Run_Success(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	project := &interfaces.TracksProject{Name: "testapp"}
	mockDetector.On("Detect", mock.Anything, ".").
		Return(project, "/tmp/testapp", nil).Once()

	mockExecutor.On("Upgrade", mock.Anything, "/tmp/testapp", "").
		Return(nil).Once()
	mockExecutor.On("Version", mock.Anything, "/tmp/testapp").
		Return("v0.2.0", nil).Once()

	mockRenderer.On("Title", "Upgrading templUI").Once()
	mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
		return strings.Contains(s.Body, "Successfully upgraded") &&
			strings.Contains(s.Body, "latest") &&
			strings.Contains(s.Body, "v0.2.0")
	})).Once()
	mockRenderer.On("Flush").Return(nil).Once()

	factory := func(*cobra.Command) interfaces.Renderer { return mockRenderer }
	flusher := func(*cobra.Command, interfaces.Renderer) { mockRenderer.Flush() }

	cmd := NewUIUpgradeCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	mockExecutor.AssertExpectations(t)
	mockDetector.AssertExpectations(t)
	mockRenderer.AssertExpectations(t)
}

func TestUIUpgradeCommand_Run_ExecutorError(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	project := &interfaces.TracksProject{Name: "testapp"}
	mockDetector.On("Detect", mock.Anything, ".").
		Return(project, "/tmp/testapp", nil).Once()

	mockExecutor.On("Upgrade", mock.Anything, "/tmp/testapp", "").
		Return(errors.New("upgrade failed")).Once()

	mockRenderer.On("Title", mock.Anything).Once()
	mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
		return strings.Contains(s.Body, "Error:") && strings.Contains(s.Body, "upgrade failed")
	})).Once()
	mockRenderer.On("Flush").Return(nil).Once()

	factory := func(*cobra.Command) interfaces.Renderer { return mockRenderer }
	flusher := func(*cobra.Command, interfaces.Renderer) { mockRenderer.Flush() }

	cmd := NewUIUpgradeCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	mockExecutor.AssertExpectations(t)
	mockRenderer.AssertExpectations(t)
}

func TestUIUpgradeCommand_Run_VersionError(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	project := &interfaces.TracksProject{Name: "testapp"}
	mockDetector.On("Detect", mock.Anything, ".").
		Return(project, "/tmp/testapp", nil).Once()

	mockExecutor.On("Upgrade", mock.Anything, "/tmp/testapp", "").
		Return(nil).Once()
	mockExecutor.On("Version", mock.Anything, "/tmp/testapp").
		Return("", errors.New("version check failed")).Once()

	mockRenderer.On("Title", mock.Anything).Once()
	mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
		return strings.Contains(s.Body, "Upgrade completed successfully")
	})).Once()
	mockRenderer.On("Flush").Return(nil).Once()

	factory := func(*cobra.Command) interfaces.Renderer { return mockRenderer }
	flusher := func(*cobra.Command, interfaces.Renderer) { mockRenderer.Flush() }

	cmd := NewUIUpgradeCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	mockExecutor.AssertExpectations(t)
}

func TestUIUpgradeCommand_Run_DetectorError(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	mockDetector.On("Detect", mock.Anything, ".").
		Return(nil, "", errors.New("failed to detect project")).Once()

	mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
		return strings.Contains(s.Body, "Error:") && strings.Contains(s.Body, "failed to detect project")
	})).Once()
	mockRenderer.On("Flush").Return(nil).Once()

	factory := func(*cobra.Command) interfaces.Renderer { return mockRenderer }
	flusher := func(*cobra.Command, interfaces.Renderer) { mockRenderer.Flush() }

	cmd := NewUIUpgradeCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	mockDetector.AssertExpectations(t)
	mockRenderer.AssertExpectations(t)
}
