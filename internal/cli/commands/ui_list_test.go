package commands

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/anomalousventures/tracks/tests/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func setupUIListTestCommand(t *testing.T) (*cobra.Command, *mocks.MockUIExecutor, *mocks.MockProjectDetector, *mocks.MockRenderer) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {
		mockRenderer.Flush()
	}

	cmd := NewUIListCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	return cobraCmd, mockExecutor, mockDetector, mockRenderer
}

func TestUIListCommand_Command(t *testing.T) {
	cobraCmd, _, _, _ := setupUIListTestCommand(t)

	if cobraCmd == nil {
		t.Fatal("Command() returned nil")
	}

	if !strings.HasPrefix(cobraCmd.Use, "list") {
		t.Errorf("expected Use to start with 'list', got %q", cobraCmd.Use)
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

func TestUIListCommand_Run_NotInProject(t *testing.T) {
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

	cmd := NewUIListCommand(mockDetector, mockExecutor, factory, flusher)
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

func TestUIListCommand_Run_Success(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	project := &interfaces.TracksProject{Name: "testapp"}
	mockDetector.On("Detect", mock.Anything, ".").
		Return(project, "/tmp/testapp", nil).Once()

	components := []interfaces.UIComponent{
		{Name: "button", Category: "Forms", Installed: false},
		{Name: "card", Category: "Layout", Installed: false},
	}
	mockExecutor.On("List", mock.Anything, "/tmp/testapp", "").
		Return(components, nil).Once()

	mockRenderer.On("Title", "Available templUI Components").Once()
	mockRenderer.On("Table", mock.MatchedBy(func(t interfaces.Table) bool {
		return len(t.Headers) == 3 && len(t.Rows) == 2
	})).Once()
	mockRenderer.On("Flush").Return(nil).Once()

	factory := func(*cobra.Command) interfaces.Renderer { return mockRenderer }
	flusher := func(*cobra.Command, interfaces.Renderer) { mockRenderer.Flush() }

	cmd := NewUIListCommand(mockDetector, mockExecutor, factory, flusher)
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

func TestUIListCommand_Run_ExecutorError(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	project := &interfaces.TracksProject{Name: "testapp"}
	mockDetector.On("Detect", mock.Anything, ".").
		Return(project, "/tmp/testapp", nil).Once()

	mockExecutor.On("List", mock.Anything, "/tmp/testapp", "").
		Return(nil, errors.New("failed to list components")).Once()

	mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
		return strings.Contains(s.Body, "Error:") && strings.Contains(s.Body, "failed to list components")
	})).Once()
	mockRenderer.On("Flush").Return(nil).Once()

	factory := func(*cobra.Command) interfaces.Renderer { return mockRenderer }
	flusher := func(*cobra.Command, interfaces.Renderer) { mockRenderer.Flush() }

	cmd := NewUIListCommand(mockDetector, mockExecutor, factory, flusher)
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

func TestUIListCommand_TableContent(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	project := &interfaces.TracksProject{Name: "testapp"}
	mockDetector.On("Detect", mock.Anything, ".").
		Return(project, "/tmp/testapp", nil).Once()

	components := []interfaces.UIComponent{
		{Name: "button", Category: "Forms", Installed: false},
		{Name: "toast", Category: "Feedback", Installed: false},
	}
	mockExecutor.On("List", mock.Anything, "/tmp/testapp", "").
		Return(components, nil).Once()

	var capturedTable interfaces.Table
	mockRenderer.On("Title", mock.Anything).Once()
	mockRenderer.On("Table", mock.Anything).Run(func(args mock.Arguments) {
		capturedTable = args.Get(0).(interfaces.Table)
	}).Once()
	mockRenderer.On("Flush").Return(nil).Once()

	factory := func(*cobra.Command) interfaces.Renderer { return mockRenderer }
	flusher := func(*cobra.Command, interfaces.Renderer) { mockRenderer.Flush() }

	cmd := NewUIListCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if len(capturedTable.Headers) != 3 {
		t.Errorf("expected 3 headers, got %d", len(capturedTable.Headers))
	}
	if capturedTable.Headers[0] != "NAME" {
		t.Errorf("expected first header 'NAME', got %q", capturedTable.Headers[0])
	}
	if len(capturedTable.Rows) != 2 {
		t.Errorf("expected 2 rows, got %d", len(capturedTable.Rows))
	}
	if capturedTable.Rows[0][0] != "button" {
		t.Errorf("expected first row name 'button', got %q", capturedTable.Rows[0][0])
	}
}

func TestGetInstalledComponents(t *testing.T) {
	tmpDir := t.TempDir()
	uiDir := filepath.Join(tmpDir, "internal", "http", "views", "components", "ui")
	if err := os.MkdirAll(uiDir, 0755); err != nil {
		t.Fatalf("failed to create ui dir: %v", err)
	}

	_ = os.WriteFile(filepath.Join(uiDir, "button.templ"), []byte(""), 0644)
	_ = os.WriteFile(filepath.Join(uiDir, "card.templ"), []byte(""), 0644)
	_ = os.WriteFile(filepath.Join(uiDir, "README.md"), []byte(""), 0644)

	installed := getInstalledComponents(tmpDir)

	if !installed["button"] {
		t.Error("expected button to be installed")
	}
	if !installed["card"] {
		t.Error("expected card to be installed")
	}
	if installed["README"] {
		t.Error("README.md should not be counted as installed component")
	}
}

func TestGetInstalledComponents_NoDir(t *testing.T) {
	installed := getInstalledComponents("/nonexistent/path")
	if len(installed) != 0 {
		t.Errorf("expected empty map for nonexistent dir, got %d entries", len(installed))
	}
}
