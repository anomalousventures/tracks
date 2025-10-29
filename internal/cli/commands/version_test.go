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
func setupVersionTestCommand(t *testing.T, version, commit, date string) *cobra.Command {
	mockBuild := mocks.NewMockBuildInfo(t)
	mockBuild.On("GetVersion").Return(version).Maybe()
	mockBuild.On("GetCommit").Return(commit).Maybe()
	mockBuild.On("GetDate").Return(date).Maybe()
	mockRenderer := mocks.NewMockRenderer(t)
	mockRenderer.On("Title", mock.Anything).Return().Maybe()
	mockRenderer.On("Section", mock.Anything).Return().Maybe()
	mockRenderer.On("Flush").Return(nil).Maybe()

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {
		// Actually call Flush for tests that execute
		mockRenderer.Flush()
	}
	cmd := NewVersionCommand(mockBuild, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))
	return cobraCmd
}

// Use this when you need to verify renderer or build info method calls.
func setupVersionTestCommandWithMock(t *testing.T, version, commit, date string) (*cobra.Command, *mocks.MockRenderer, *mocks.MockBuildInfo) {
	mockBuild := mocks.NewMockBuildInfo(t)
	mockBuild.On("GetVersion").Return(version).Maybe()
	mockBuild.On("GetCommit").Return(commit).Maybe()
	mockBuild.On("GetDate").Return(date).Maybe()

	mockRenderer := mocks.NewMockRenderer(t)

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}
	cmd := NewVersionCommand(mockBuild, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))
	return cobraCmd, mockRenderer, mockBuild
}

func TestNewVersionCommand(t *testing.T) {
	mockBuild := mocks.NewMockBuildInfo(t)
	mockBuild.On("GetVersion").Return("v1.0.0")
	mockBuild.On("GetCommit").Return("abc123")
	mockBuild.On("GetDate").Return("2025-10-29")

	mockRenderer := mocks.NewMockRenderer(t)
	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	cmd := NewVersionCommand(mockBuild, rendererFactory, flusher)

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
	if cmd.build.GetVersion() != "v1.0.0" {
		t.Errorf("expected build.GetVersion() %q, got %q", "v1.0.0", cmd.build.GetVersion())
	}
	if cmd.build.GetCommit() != "abc123" {
		t.Errorf("expected build.GetCommit() %q, got %q", "abc123", cmd.build.GetCommit())
	}
	if cmd.build.GetDate() != "2025-10-29" {
		t.Errorf("expected build.GetDate() %q, got %q", "2025-10-29", cmd.build.GetDate())
	}
}

func TestVersionCommand_Command(t *testing.T) {
	mockBuild := mocks.NewMockBuildInfo(t)

	mockRenderer := mocks.NewMockRenderer(t)
	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	versionCmd := NewVersionCommand(mockBuild, rendererFactory, flusher)
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
	cobraCmd := setupVersionTestCommand(t, "v1.0.0", "abc123", "2025-10-29")

	// Test that it works with no arguments
	cobraCmd.SetArgs([]string{})
	if err := cobraCmd.Execute(); err != nil {
		t.Errorf("expected no error when no arguments provided, got %v", err)
	}
}

func TestVersionCommand_Run(t *testing.T) {
	mockBuild := mocks.NewMockBuildInfo(t)
	mockBuild.On("GetVersion").Return("v1.0.0")
	mockBuild.On("GetCommit").Return("abc123")
	mockBuild.On("GetDate").Return("2025-10-29")

	mockRenderer := mocks.NewMockRenderer(t)
	mockRenderer.On("Title", "Tracks v1.0.0").Once()
	mockRenderer.On("Section", interfaces.Section{Body: "Commit: abc123\nBuilt: 2025-10-29"}).Once()

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

	versionCmd := NewVersionCommand(mockBuild, rendererFactory, flusher)
	cobraCmd := versionCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	// Execute with no arguments
	cobraCmd.SetArgs([]string{})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !flusherCalled {
		t.Error("flusher was not called")
	}

	// Mock expectations are automatically verified in cleanup
}

func TestVersionCommand_RunWithDifferentBuildInfo(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		commit        string
		date          string
		wantTitle     string
		wantBodyParts []string
	}{
		{
			name:          "release version",
			version:       "v1.2.3",
			commit:        "abc123def456",
			date:          "2025-10-29T12:00:00Z",
			wantTitle:     "Tracks v1.2.3",
			wantBodyParts: []string{"Commit: abc123def456", "Built: 2025-10-29T12:00:00Z"},
		},
		{
			name:          "dev version",
			version:       "dev",
			commit:        "local",
			date:          "unknown",
			wantTitle:     "Tracks dev",
			wantBodyParts: []string{"Commit: local", "Built: unknown"},
		},
		{
			name:          "empty commit and date",
			version:       "v0.1.0",
			commit:        "",
			date:          "",
			wantTitle:     "Tracks v0.1.0",
			wantBodyParts: []string{"Commit: ", "Built: "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cobraCmd, mockRenderer, _ := setupVersionTestCommandWithMock(t, tt.version, tt.commit, tt.date)
			mockRenderer.On("Title", tt.wantTitle).Once()
			mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
				for _, part := range tt.wantBodyParts {
					if !strings.Contains(s.Body, part) {
						return false
					}
				}
				return true
			})).Once()

			cobraCmd.SetArgs([]string{})
			if err := cobraCmd.Execute(); err != nil {
				t.Fatalf("execution failed: %v", err)
			}
		})
	}
}

func TestVersionCommand_RendererFactoryCalledWithCommand(t *testing.T) {
	mockBuild := mocks.NewMockBuildInfo(t)
	mockBuild.On("GetVersion").Return("v1.0.0")
	mockBuild.On("GetCommit").Return("abc123")
	mockBuild.On("GetDate").Return("2025-10-29")

	mockRenderer := mocks.NewMockRenderer(t)
	mockRenderer.On("Title", mock.Anything).Once()
	mockRenderer.On("Section", mock.Anything).Once()

	var capturedCmd *cobra.Command
	rendererFactory := func(cmd *cobra.Command) interfaces.Renderer {
		capturedCmd = cmd
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	versionCmd := NewVersionCommand(mockBuild, rendererFactory, flusher)
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
	mockBuild := mocks.NewMockBuildInfo(t)
	mockBuild.On("GetVersion").Return("v1.0.0")
	mockBuild.On("GetCommit").Return("abc123")
	mockBuild.On("GetDate").Return("2025-10-29")

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

	versionCmd := NewVersionCommand(mockBuild, rendererFactory, flusher)
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

	if capturedRenderer != mockRenderer {
		t.Error("flusher not called with correct renderer")
	}
}

func TestVersionCommand_BuildInfoGetVersionCalled(t *testing.T) {
	cobraCmd, mockRenderer, _ := setupVersionTestCommandWithMock(t, "v2.5.0", "xyz789", "2025-11-01")
	expectedTitle := "Tracks v2.5.0"
	mockRenderer.On("Title", expectedTitle).Once()
	mockRenderer.On("Section", mock.Anything).Once()

	cobraCmd.SetArgs([]string{})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Mock expectations are automatically verified in cleanup
}

func TestVersionCommand_CommandDescriptions(t *testing.T) {
	cobraCmd := setupVersionTestCommand(t, "v1.0.0", "abc123", "2025-10-29")

	// Verify Long description mentions key information
	keyPhrases := []string{"version number", "commit", "build date"}
	for _, phrase := range keyPhrases {
		if !strings.Contains(cobraCmd.Long, phrase) {
			t.Errorf("Long description missing mention of %q", phrase)
		}
	}

	// Verify Short description is meaningful
	if !strings.Contains(cobraCmd.Short, "version") {
		t.Error("Short description should mention 'version'")
	}
}
