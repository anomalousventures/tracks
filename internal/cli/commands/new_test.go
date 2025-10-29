package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/anomalousventures/tracks/tests/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

// Use this for tests that don't need to inspect mock calls.
func setupTestCommand(t *testing.T) *cobra.Command {
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
	cmd := NewNewCommand(factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))
	return cobraCmd
}

// Use this when you need to verify renderer method calls.
func setupTestCommandWithMock(t *testing.T) (*cobra.Command, *mocks.MockRenderer) {
	mockRenderer := mocks.NewMockRenderer(t)

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}
	cmd := NewNewCommand(factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))
	return cobraCmd, mockRenderer
}

func TestNewNewCommand(t *testing.T) {
	mockRenderer := mocks.NewMockRenderer(t)
	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	cmd := NewNewCommand(rendererFactory, flusher)

	if cmd == nil {
		t.Fatal("NewNewCommand returned nil")
	}

	if cmd.newRenderer == nil {
		t.Error("newRenderer field not set")
	}

	if cmd.flushRenderer == nil {
		t.Error("flushRenderer field not set")
	}
}

func TestNewCommand_Command(t *testing.T) {
	mockRenderer := mocks.NewMockRenderer(t)
	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	newCmd := NewNewCommand(rendererFactory, flusher)
	cobraCmd := newCmd.Command()

	if cobraCmd == nil {
		t.Fatal("Command() returned nil")
	}

	if cobraCmd.Use != "new [project-name]" {
		t.Errorf("expected Use 'new [project-name]', got %q", cobraCmd.Use)
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

func TestNewCommand_CommandUsage(t *testing.T) {
	cobraCmd := setupTestCommand(t)

	cobraCmd.SetArgs([]string{})
	if err := cobraCmd.Execute(); err == nil {
		t.Error("expected error when no arguments provided, got nil")
	}

	cobraCmd.SetArgs([]string{"project1", "project2"})
	if err := cobraCmd.Execute(); err == nil {
		t.Error("expected error when too many arguments provided, got nil")
	}
}

func TestNewCommand_Run(t *testing.T) {
	mockRenderer := mocks.NewMockRenderer(t)
	mockRenderer.On("Title", "Creating new Tracks application: myapp").Once()
	mockRenderer.On("Section", interfaces.Section{Body: "(Full implementation coming soon)"}).Once()

	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}

	flusherCalled := false
	flusher := func(cmd *cobra.Command, r interfaces.Renderer) {
		flusherCalled = true
		if r != mockRenderer {
			t.Error("flusher called with different renderer")
		}
	}

	newCmd := NewNewCommand(rendererFactory, flusher)
	cobraCmd := newCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	// Execute with valid argument
	cobraCmd.SetArgs([]string{"myapp"})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !flusherCalled {
		t.Error("flusher was not called")
	}
}

func TestNewCommand_RunWithDifferentProjectNames(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		wantTitle   string
	}{
		{
			name:        "simple name",
			projectName: "myapp",
			wantTitle:   "Creating new Tracks application: myapp",
		},
		{
			name:        "name with hyphens",
			projectName: "my-awesome-app",
			wantTitle:   "Creating new Tracks application: my-awesome-app",
		},
		{
			name:        "single char name",
			projectName: "x",
			wantTitle:   "Creating new Tracks application: x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cobraCmd, mockRenderer := setupTestCommandWithMock(t)
			mockRenderer.On("Title", tt.wantTitle).Once()
			mockRenderer.On("Section", mock.Anything).Once()

			cobraCmd.SetArgs([]string{tt.projectName})
			if err := cobraCmd.Execute(); err != nil {
				t.Fatalf("execution failed: %v", err)
			}
		})
	}
}

func TestNewCommand_RendererFactoryCalledWithCommand(t *testing.T) {
	mockRenderer := mocks.NewMockRenderer(t)
	mockRenderer.On("Title", mock.Anything).Once()
	mockRenderer.On("Section", mock.Anything).Once()

	var capturedCmd *cobra.Command
	rendererFactory := func(cmd *cobra.Command) interfaces.Renderer {
		capturedCmd = cmd
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	newCmd := NewNewCommand(rendererFactory, flusher)
	cobraCmd := newCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"testapp"})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if capturedCmd != cobraCmd {
		t.Error("renderer factory not called with correct command")
	}
}

func TestNewCommand_FlusherCalledWithCommandAndRenderer(t *testing.T) {
	mockRenderer := mocks.NewMockRenderer(t)
	mockRenderer.On("Title", mock.Anything).Once()
	mockRenderer.On("Section", mock.Anything).Once()

	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}

	var capturedCmd *cobra.Command
	var capturedRenderer interfaces.Renderer
	flusher := func(cmd *cobra.Command, r interfaces.Renderer) {
		capturedCmd = cmd
		capturedRenderer = r
	}

	newCmd := NewNewCommand(rendererFactory, flusher)
	cobraCmd := newCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"testapp"})
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

func TestNewCommand_CommandDescriptions(t *testing.T) {
	cobraCmd := setupTestCommand(t)

	keyTechnologies := []string{"templ", "SQLC", "production-ready", "Go"}
	for _, tech := range keyTechnologies {
		if !strings.Contains(cobraCmd.Long, tech) {
			t.Errorf("Long description missing mention of %q", tech)
		}
	}

	if !strings.Contains(cobraCmd.Example, "tracks new myapp") {
		t.Error("Example missing basic usage pattern")
	}
}
