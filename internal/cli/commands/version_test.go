package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/spf13/cobra"
)

// mockBuildInfo implements the BuildInfo interface for testing.
type mockBuildInfo struct {
	version string
	commit  string
	date    string
}

func (m *mockBuildInfo) GetVersion() string {
	return m.version
}

func (m *mockBuildInfo) GetCommit() string {
	return m.commit
}

func (m *mockBuildInfo) GetDate() string {
	return m.date
}

// Test helpers for reducing boilerplate

// setupVersionTestCommand creates a VersionCommand with default mocks and returns the cobra command
// configured with output buffers. Use this for tests that don't need to inspect mock calls.
func setupVersionTestCommand(build interfaces.BuildInfo) *cobra.Command {
	factory := func(*cobra.Command) interfaces.Renderer {
		return &mockRenderer{}
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}
	cmd := NewVersionCommand(build, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))
	return cobraCmd
}

// setupVersionTestCommandWithMock returns command and mock for inspection.
// Use this when you need to verify renderer method calls.
func setupVersionTestCommandWithMock(build interfaces.BuildInfo) (*cobra.Command, *mockRenderer) {
	mock := &mockRenderer{}
	factory := func(*cobra.Command) interfaces.Renderer {
		return mock
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}
	cmd := NewVersionCommand(build, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))
	return cobraCmd, mock
}

func TestNewVersionCommand(t *testing.T) {
	build := &mockBuildInfo{
		version: "v1.0.0",
		commit:  "abc123",
		date:    "2025-10-29",
	}

	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return &mockRenderer{}
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	cmd := NewVersionCommand(build, rendererFactory, flusher)

	if cmd == nil {
		t.Fatal("NewVersionCommand returned nil")
	}

	if cmd.newRenderer == nil {
		t.Error("newRenderer field not set")
	}

	if cmd.flushRenderer == nil {
		t.Error("flushRenderer field not set")
	}

	// Verify build info is stored correctly
	if cmd.build.GetVersion() != build.GetVersion() {
		t.Errorf("expected build.GetVersion() %q, got %q", build.GetVersion(), cmd.build.GetVersion())
	}
	if cmd.build.GetCommit() != build.GetCommit() {
		t.Errorf("expected build.GetCommit() %q, got %q", build.GetCommit(), cmd.build.GetCommit())
	}
	if cmd.build.GetDate() != build.GetDate() {
		t.Errorf("expected build.GetDate() %q, got %q", build.GetDate(), cmd.build.GetDate())
	}
}

func TestVersionCommand_Command(t *testing.T) {
	build := &mockBuildInfo{
		version: "v1.0.0",
		commit:  "abc123",
		date:    "2025-10-29",
	}

	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return &mockRenderer{}
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	versionCmd := NewVersionCommand(build, rendererFactory, flusher)
	cobraCmd := versionCmd.Command()

	if cobraCmd == nil {
		t.Fatal("Command() returned nil")
	}

	if cobraCmd.Use != "version" {
		t.Errorf("expected Use 'version', got %q", cobraCmd.Use)
	}

	if cobraCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cobraCmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cobraCmd.Run == nil {
		t.Error("Run is nil, expected function")
	}
}

func TestVersionCommand_CommandUsage(t *testing.T) {
	build := &mockBuildInfo{
		version: "v1.0.0",
		commit:  "abc123",
		date:    "2025-10-29",
	}

	cobraCmd := setupVersionTestCommand(build)

	// Test that it works with no arguments
	cobraCmd.SetArgs([]string{})
	if err := cobraCmd.Execute(); err != nil {
		t.Errorf("expected no error when no arguments provided, got %v", err)
	}
}

func TestVersionCommand_Run(t *testing.T) {
	build := &mockBuildInfo{
		version: "v1.0.0",
		commit:  "abc123",
		date:    "2025-10-29",
	}

	mock := &mockRenderer{}
	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mock
	}

	flusherCalled := false
	flusher := func(cmd *cobra.Command, r interfaces.Renderer) {
		flusherCalled = true
		if r != mock {
			t.Error("flusher called with different renderer")
		}
	}

	versionCmd := NewVersionCommand(build, rendererFactory, flusher)
	cobraCmd := versionCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	// Execute with no arguments
	cobraCmd.SetArgs([]string{})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Verify renderer was called
	if len(mock.titleCalls) != 1 {
		t.Errorf("expected 1 Title call, got %d", len(mock.titleCalls))
	} else {
		expectedTitle := "Tracks v1.0.0"
		if mock.titleCalls[0] != expectedTitle {
			t.Errorf("expected title %q, got %q", expectedTitle, mock.titleCalls[0])
		}
	}

	if len(mock.sectionCalls) != 1 {
		t.Errorf("expected 1 Section call, got %d", len(mock.sectionCalls))
	} else {
		expectedBody := "Commit: abc123\nBuilt: 2025-10-29"
		if mock.sectionCalls[0].Body != expectedBody {
			t.Errorf("expected section body %q, got %q", expectedBody, mock.sectionCalls[0].Body)
		}
	}

	if !flusherCalled {
		t.Error("flusher was not called")
	}
}

func TestVersionCommand_RunWithDifferentBuildInfo(t *testing.T) {
	tests := []struct {
		name          string
		build         interfaces.BuildInfo
		wantTitle     string
		wantBodyParts []string
	}{
		{
			name: "release version",
			build: &mockBuildInfo{
				version: "v1.2.3",
				commit:  "abc123def456",
				date:    "2025-10-29T12:00:00Z",
			},
			wantTitle:     "Tracks v1.2.3",
			wantBodyParts: []string{"Commit: abc123def456", "Built: 2025-10-29T12:00:00Z"},
		},
		{
			name: "dev version",
			build: &mockBuildInfo{
				version: "dev",
				commit:  "local",
				date:    "unknown",
			},
			wantTitle:     "Tracks dev",
			wantBodyParts: []string{"Commit: local", "Built: unknown"},
		},
		{
			name: "empty commit and date",
			build: &mockBuildInfo{
				version: "v0.1.0",
				commit:  "",
				date:    "",
			},
			wantTitle:     "Tracks v0.1.0",
			wantBodyParts: []string{"Commit: ", "Built: "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cobraCmd, mock := setupVersionTestCommandWithMock(tt.build)

			cobraCmd.SetArgs([]string{})
			if err := cobraCmd.Execute(); err != nil {
				t.Fatalf("execution failed: %v", err)
			}

			if len(mock.titleCalls) != 1 {
				t.Fatalf("expected 1 Title call, got %d", len(mock.titleCalls))
			}

			if mock.titleCalls[0] != tt.wantTitle {
				t.Errorf("expected title %q, got %q", tt.wantTitle, mock.titleCalls[0])
			}

			if len(mock.sectionCalls) != 1 {
				t.Fatalf("expected 1 Section call, got %d", len(mock.sectionCalls))
			}

			// Check that all expected parts are present in the body
			body := mock.sectionCalls[0].Body
			for _, part := range tt.wantBodyParts {
				if !strings.Contains(body, part) {
					t.Errorf("expected body to contain %q, got %q", part, body)
				}
			}
		})
	}
}

func TestVersionCommand_RendererFactoryCalledWithCommand(t *testing.T) {
	build := &mockBuildInfo{
		version: "v1.0.0",
		commit:  "abc123",
		date:    "2025-10-29",
	}

	var capturedCmd *cobra.Command
	rendererFactory := func(cmd *cobra.Command) interfaces.Renderer {
		capturedCmd = cmd
		return &mockRenderer{}
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	versionCmd := NewVersionCommand(build, rendererFactory, flusher)
	cobraCmd := versionCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if capturedCmd != cobraCmd {
		t.Error("renderer factory not called with correct command")
	}
}

func TestVersionCommand_FlusherCalledWithCommandAndRenderer(t *testing.T) {
	build := &mockBuildInfo{
		version: "v1.0.0",
		commit:  "abc123",
		date:    "2025-10-29",
	}

	mock := &mockRenderer{}
	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mock
	}

	var capturedCmd *cobra.Command
	var capturedRenderer interfaces.Renderer
	flusher := func(cmd *cobra.Command, r interfaces.Renderer) {
		capturedCmd = cmd
		capturedRenderer = r
	}

	versionCmd := NewVersionCommand(build, rendererFactory, flusher)
	cobraCmd := versionCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{})
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

func TestVersionCommand_BuildInfoGetVersionCalled(t *testing.T) {
	build := &mockBuildInfo{
		version: "v2.5.0",
		commit:  "xyz789",
		date:    "2025-11-01",
	}

	cobraCmd, mock := setupVersionTestCommandWithMock(build)

	cobraCmd.SetArgs([]string{})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Verify that GetVersion() result is used in the title
	if len(mock.titleCalls) != 1 {
		t.Fatalf("expected 1 Title call, got %d", len(mock.titleCalls))
	}

	expectedTitle := "Tracks " + build.GetVersion()
	if mock.titleCalls[0] != expectedTitle {
		t.Errorf("expected title %q, got %q", expectedTitle, mock.titleCalls[0])
	}
}

func TestVersionCommand_CommandDescriptions(t *testing.T) {
	build := &mockBuildInfo{
		version: "v1.0.0",
		commit:  "abc123",
		date:    "2025-10-29",
	}

	cobraCmd := setupVersionTestCommand(build)

	// Verify Long description mentions key information
	keyPhrases := []string{"version number", "commit", "build date"}
	for _, phrase := range keyPhrases {
		if !contains(cobraCmd.Long, phrase) {
			t.Errorf("Long description missing mention of %q", phrase)
		}
	}

	// Verify Short description is meaningful
	if !contains(cobraCmd.Short, "version") {
		t.Error("Short description should mention 'version'")
	}
}
