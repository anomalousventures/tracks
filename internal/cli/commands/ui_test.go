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

func setupUITestCommand(t *testing.T) (*cobra.Command, *mocks.MockUIExecutor, *mocks.MockProjectDetector) {
	mockExecutor := mocks.NewMockUIExecutor(t)
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

	cmd := NewUICommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	return cobraCmd, mockExecutor, mockDetector
}

func TestNewUICommand(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	cmd := NewUICommand(mockDetector, mockExecutor, rendererFactory, flusher)

	if cmd == nil {
		t.Fatal("NewUICommand returned nil")
	}

	if cmd.detector == nil {
		t.Error("detector field not set")
	}

	if cmd.executor == nil {
		t.Error("executor field not set")
	}

	if cmd.newRenderer == nil {
		t.Error("newRenderer field not set")
	}

	if cmd.flushRenderer == nil {
		t.Error("flushRenderer field not set")
	}
}

func TestUICommand_Command(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	uiCmd := NewUICommand(mockDetector, mockExecutor, rendererFactory, flusher)
	cobraCmd := uiCmd.Command()

	if cobraCmd == nil {
		t.Fatal("Command() returned nil")
	}

	if cobraCmd.Use != "ui" {
		t.Errorf("expected Use 'ui', got %q", cobraCmd.Use)
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

	if cobraCmd.Run == nil {
		t.Error("Run is nil, expected function")
	}

	versionFlag := cobraCmd.Flags().Lookup("version")
	if versionFlag == nil {
		t.Fatal("--version flag not found")
	}
	if versionFlag.DefValue != "false" {
		t.Errorf("expected --version default 'false', got %q", versionFlag.DefValue)
	}
}

func TestUICommand_Run_ShowVersion(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	mockExecutor.On("Version", mock.Anything, ".").Return("v0.1.0", nil).Once()
	mockRenderer.On("Title", "templUI v0.1.0").Once()

	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusherCalled := false
	flusher := func(cmd *cobra.Command, r interfaces.Renderer) {
		flusherCalled = true
	}

	uiCmd := NewUICommand(mockDetector, mockExecutor, rendererFactory, flusher)
	cobraCmd := uiCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"--version"})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !flusherCalled {
		t.Error("flusher was not called")
	}

	mockExecutor.AssertExpectations(t)
	mockRenderer.AssertExpectations(t)
}

func TestUICommand_Run_VersionError(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	mockExecutor.On("Version", mock.Anything, ".").Return("", errors.New("templui not found")).Once()
	mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
		return strings.Contains(s.Body, "Error:") && strings.Contains(s.Body, "templui not found")
	})).Once()

	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(cmd *cobra.Command, r interfaces.Renderer) {}

	uiCmd := NewUICommand(mockDetector, mockExecutor, rendererFactory, flusher)
	cobraCmd := uiCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"--version"})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	mockExecutor.AssertExpectations(t)
	mockRenderer.AssertExpectations(t)
}

func TestUICommand_Run_NoFlags(t *testing.T) {
	cobraCmd, _, _ := setupUITestCommand(t)

	var outBuf bytes.Buffer
	cobraCmd.SetOut(&outBuf)
	cobraCmd.SetArgs([]string{})

	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "ui") {
		t.Error("expected help output to contain 'ui'")
	}
}

func TestUICommand_CommandDescriptions(t *testing.T) {
	cobraCmd, _, _ := setupUITestCommand(t)

	keyPhrases := []string{"templUI", "component", "Tracks"}
	for _, phrase := range keyPhrases {
		if !strings.Contains(cobraCmd.Long, phrase) {
			t.Errorf("Long description missing mention of %q", phrase)
		}
	}

	if !strings.Contains(cobraCmd.Example, "tracks ui --version") {
		t.Error("Example missing --version usage pattern")
	}
}

func TestUICommand_RendererFactoryCalledWithCommand(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	mockExecutor.On("Version", mock.Anything, ".").Return("v0.1.0", nil).Once()
	mockRenderer.On("Title", mock.Anything).Once()

	var capturedCmd *cobra.Command
	rendererFactory := func(cmd *cobra.Command) interfaces.Renderer {
		capturedCmd = cmd
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	uiCmd := NewUICommand(mockDetector, mockExecutor, rendererFactory, flusher)
	cobraCmd := uiCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"--version"})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if capturedCmd != cobraCmd {
		t.Error("renderer factory not called with correct command")
	}
}

func TestUICommand_FlusherCalledWithCommandAndRenderer(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	mockExecutor.On("Version", mock.Anything, ".").Return("v0.1.0", nil).Once()
	mockRenderer.On("Title", mock.Anything).Once()

	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}

	var capturedCmd *cobra.Command
	var capturedRenderer interfaces.Renderer
	flusher := func(cmd *cobra.Command, r interfaces.Renderer) {
		capturedCmd = cmd
		capturedRenderer = r
	}

	uiCmd := NewUICommand(mockDetector, mockExecutor, rendererFactory, flusher)
	cobraCmd := uiCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"--version"})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if capturedCmd != cobraCmd {
		t.Error("flusher not called with correct command")
	}

	if capturedRenderer != mockRenderer {
		t.Error("flusher not called with correct renderer")
	}
}
