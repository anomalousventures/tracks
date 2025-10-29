package commands

import (
	"bytes"
	"testing"

	"github.com/anomalousventures/tracks/internal/cli/renderer"
	"github.com/spf13/cobra"
)

// Mock renderer that records calls
type mockRenderer struct {
	titleCalls   []string
	sectionCalls []renderer.Section
	flushed      bool
}

func (m *mockRenderer) Title(text string) {
	m.titleCalls = append(m.titleCalls, text)
}

func (m *mockRenderer) Section(s renderer.Section) {
	m.sectionCalls = append(m.sectionCalls, s)
}

func (m *mockRenderer) Table(t renderer.Table) {}

func (m *mockRenderer) Progress(spec renderer.ProgressSpec) renderer.Progress {
	return nil
}

func (m *mockRenderer) Flush() error {
	m.flushed = true
	return nil
}

func TestNewNewCommand(t *testing.T) {
	rendererFactory := func(*cobra.Command) renderer.Renderer {
		return &mockRenderer{}
	}
	flusher := func(*cobra.Command, renderer.Renderer) {}

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
	rendererFactory := func(*cobra.Command) renderer.Renderer {
		return &mockRenderer{}
	}
	flusher := func(*cobra.Command, renderer.Renderer) {}

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
	rendererFactory := func(*cobra.Command) renderer.Renderer {
		return &mockRenderer{}
	}
	flusher := func(*cobra.Command, renderer.Renderer) {}

	newCmd := NewNewCommand(rendererFactory, flusher)
	cobraCmd := newCmd.Command()

	// Test that it requires exactly 1 argument
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
	mock := &mockRenderer{}
	rendererFactory := func(*cobra.Command) renderer.Renderer {
		return mock
	}

	flusherCalled := false
	flusher := func(cmd *cobra.Command, r renderer.Renderer) {
		flusherCalled = true
		if r != mock {
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

	// Verify renderer was called
	if len(mock.titleCalls) != 1 {
		t.Errorf("expected 1 Title call, got %d", len(mock.titleCalls))
	} else {
		expectedTitle := "Creating new Tracks application: myapp"
		if mock.titleCalls[0] != expectedTitle {
			t.Errorf("expected title %q, got %q", expectedTitle, mock.titleCalls[0])
		}
	}

	if len(mock.sectionCalls) != 1 {
		t.Errorf("expected 1 Section call, got %d", len(mock.sectionCalls))
	} else {
		expectedBody := "(Full implementation coming soon)"
		if mock.sectionCalls[0].Body != expectedBody {
			t.Errorf("expected section body %q, got %q", expectedBody, mock.sectionCalls[0].Body)
		}
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
			mock := &mockRenderer{}
			rendererFactory := func(*cobra.Command) renderer.Renderer {
				return mock
			}
			flusher := func(*cobra.Command, renderer.Renderer) {}

			newCmd := NewNewCommand(rendererFactory, flusher)
			cobraCmd := newCmd.Command()
			cobraCmd.SetOut(new(bytes.Buffer))
			cobraCmd.SetErr(new(bytes.Buffer))

			cobraCmd.SetArgs([]string{tt.projectName})
			if err := cobraCmd.Execute(); err != nil {
				t.Fatalf("execution failed: %v", err)
			}

			if len(mock.titleCalls) != 1 {
				t.Fatalf("expected 1 Title call, got %d", len(mock.titleCalls))
			}

			if mock.titleCalls[0] != tt.wantTitle {
				t.Errorf("expected title %q, got %q", tt.wantTitle, mock.titleCalls[0])
			}
		})
	}
}

func TestNewCommand_RendererFactoryCalledWithCommand(t *testing.T) {
	var capturedCmd *cobra.Command
	rendererFactory := func(cmd *cobra.Command) renderer.Renderer {
		capturedCmd = cmd
		return &mockRenderer{}
	}
	flusher := func(*cobra.Command, renderer.Renderer) {}

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
	mock := &mockRenderer{}
	rendererFactory := func(*cobra.Command) renderer.Renderer {
		return mock
	}

	var capturedCmd *cobra.Command
	var capturedRenderer renderer.Renderer
	flusher := func(cmd *cobra.Command, r renderer.Renderer) {
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

	if capturedRenderer != mock {
		t.Error("flusher not called with correct renderer")
	}
}

func TestNewCommand_CommandDescriptions(t *testing.T) {
	rendererFactory := func(*cobra.Command) renderer.Renderer {
		return &mockRenderer{}
	}
	flusher := func(*cobra.Command, renderer.Renderer) {}

	newCmd := NewNewCommand(rendererFactory, flusher)
	cobraCmd := newCmd.Command()

	// Verify Long description mentions key technologies
	keyTechnologies := []string{"templ", "SQLC", "production-ready", "Go"}
	for _, tech := range keyTechnologies {
		if !contains(cobraCmd.Long, tech) {
			t.Errorf("Long description missing mention of %q", tech)
		}
	}

	// Verify Example includes basic usage
	if !contains(cobraCmd.Example, "tracks new myapp") {
		t.Error("Example missing basic usage pattern")
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
