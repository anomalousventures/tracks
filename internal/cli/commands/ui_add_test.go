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

func setupUIAddTestCommand(t *testing.T) (*cobra.Command, *mocks.MockUIExecutor, *mocks.MockProjectDetector, *mocks.MockRenderer) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {
		mockRenderer.Flush()
	}

	cmd := NewUIAddCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	return cobraCmd, mockExecutor, mockDetector, mockRenderer
}

func TestUIAddCommand_Command(t *testing.T) {
	cobraCmd, _, _, _ := setupUIAddTestCommand(t)

	if cobraCmd == nil {
		t.Fatal("Command() returned nil")
	}

	if !strings.HasPrefix(cobraCmd.Use, "add") {
		t.Errorf("expected Use to start with 'add', got %q", cobraCmd.Use)
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

	forceFlag := cobraCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Fatal("--force flag not found")
	}
	if forceFlag.Shorthand != "f" {
		t.Errorf("expected --force shorthand 'f', got %q", forceFlag.Shorthand)
	}
}

func TestUIAddCommand_Run_NotInProject(t *testing.T) {
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

	cmd := NewUIAddCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"button"})
	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	mockDetector.AssertExpectations(t)
	mockRenderer.AssertExpectations(t)
}

func TestUIAddCommand_Run_Success(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	project := &interfaces.TracksProject{Name: "testapp"}
	mockDetector.On("Detect", mock.Anything, ".").
		Return(project, "/tmp/testapp", nil).Once()

	mockExecutor.On("Add", mock.Anything, "/tmp/testapp", "", []string{"button"}, false).
		Return(nil).Once()

	mockRenderer.On("Title", "Adding templUI components").Once()
	mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
		return strings.Contains(s.Body, "Added 1 component(s)") &&
			strings.Contains(s.Body, "button")
	})).Once()
	mockRenderer.On("Flush").Return(nil).Once()

	factory := func(*cobra.Command) interfaces.Renderer { return mockRenderer }
	flusher := func(*cobra.Command, interfaces.Renderer) { mockRenderer.Flush() }

	cmd := NewUIAddCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"button"})
	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	mockExecutor.AssertExpectations(t)
	mockDetector.AssertExpectations(t)
	mockRenderer.AssertExpectations(t)
}

func TestUIAddCommand_Run_WithForceFlag(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	project := &interfaces.TracksProject{Name: "testapp"}
	mockDetector.On("Detect", mock.Anything, ".").
		Return(project, "/tmp/testapp", nil).Once()

	mockExecutor.On("Add", mock.Anything, "/tmp/testapp", "", []string{"button"}, true).
		Return(nil).Once()

	mockRenderer.On("Title", mock.Anything).Once()
	mockRenderer.On("Section", mock.Anything).Once()
	mockRenderer.On("Flush").Return(nil).Once()

	factory := func(*cobra.Command) interfaces.Renderer { return mockRenderer }
	flusher := func(*cobra.Command, interfaces.Renderer) { mockRenderer.Flush() }

	cmd := NewUIAddCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"button", "--force"})
	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	mockExecutor.AssertExpectations(t)
}

func TestUIAddCommand_Run_MultipleComponents(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	project := &interfaces.TracksProject{Name: "testapp"}
	mockDetector.On("Detect", mock.Anything, ".").
		Return(project, "/tmp/testapp", nil).Once()

	mockExecutor.On("Add", mock.Anything, "/tmp/testapp", "", []string{"button", "card", "toast"}, false).
		Return(nil).Once()

	mockRenderer.On("Title", mock.Anything).Once()
	mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
		return strings.Contains(s.Body, "Added 3 component(s)")
	})).Once()
	mockRenderer.On("Flush").Return(nil).Once()

	factory := func(*cobra.Command) interfaces.Renderer { return mockRenderer }
	flusher := func(*cobra.Command, interfaces.Renderer) { mockRenderer.Flush() }

	cmd := NewUIAddCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"button", "card", "toast"})
	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	mockExecutor.AssertExpectations(t)
}

func TestUIAddCommand_Run_ExecutorError(t *testing.T) {
	mockExecutor := mocks.NewMockUIExecutor(t)
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	project := &interfaces.TracksProject{Name: "testapp"}
	mockDetector.On("Detect", mock.Anything, ".").
		Return(project, "/tmp/testapp", nil).Once()

	mockExecutor.On("Add", mock.Anything, "/tmp/testapp", "", []string{"button"}, false).
		Return(errors.New("component not found")).Once()

	mockRenderer.On("Title", mock.Anything).Once()
	mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
		return strings.Contains(s.Body, "Error:") && strings.Contains(s.Body, "component not found")
	})).Once()
	mockRenderer.On("Flush").Return(nil).Once()

	factory := func(*cobra.Command) interfaces.Renderer { return mockRenderer }
	flusher := func(*cobra.Command, interfaces.Renderer) { mockRenderer.Flush() }

	cmd := NewUIAddCommand(mockDetector, mockExecutor, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"button"})
	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	mockExecutor.AssertExpectations(t)
	mockRenderer.AssertExpectations(t)
}

func TestUIAddCommand_Run_NoComponents(t *testing.T) {
	cobraCmd, _, _, _ := setupUIAddTestCommand(t)

	cobraCmd.SetArgs([]string{})
	err := cobraCmd.Execute()
	if err == nil {
		t.Error("expected error for missing arguments")
	}
}

func TestHasScriptFunc(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		funcName string
		expected bool
	}{
		{
			name:     "templ declaration",
			content:  "templ ToastScript() {\n\t<script>\n\t</script>\n}",
			funcName: "ToastScript",
			expected: true,
		},
		{
			name:     "func declaration",
			content:  "func ButtonScript() templ.Component {\n\treturn nil\n}",
			funcName: "ButtonScript",
			expected: true,
		},
		{
			name:     "no script function",
			content:  "templ Button() {\n\t<button>Click</button>\n}",
			funcName: "ButtonScript",
			expected: false,
		},
		{
			name:     "different function name",
			content:  "templ CardScript() {\n\t<script>\n\t</script>\n}",
			funcName: "ButtonScript",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "component.templ")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			result := hasScriptFunc(tmpFile, tt.funcName)
			if result != tt.expected {
				t.Errorf("hasScriptFunc() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestInjectScriptCall(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		scriptCall string
		expected   string
	}{
		{
			name: "inject between markers",
			content: `<body>
			<!-- TRACKS:UI_SCRIPTS:BEGIN -->
			<!-- TRACKS:UI_SCRIPTS:END -->
		</body>`,
			scriptCall: "@ui.ToastScript()",
			expected: `<body>
			<!-- TRACKS:UI_SCRIPTS:BEGIN -->
			@ui.ToastScript()
			<!-- TRACKS:UI_SCRIPTS:END -->
		</body>`,
		},
		{
			name:       "no markers",
			content:    "<body></body>",
			scriptCall: "@ui.ToastScript()",
			expected:   "<body></body>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := injectScriptCall(tt.content, tt.scriptCall)
			if result != tt.expected {
				t.Errorf("injectScriptCall() =\n%s\n\nexpected:\n%s", result, tt.expected)
			}
		})
	}
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"button", "Button"},
		{"toast", "Toast"},
		{"", ""},
		{"B", "B"},
		{"BUTTON", "BUTTON"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := capitalizeFirst(tt.input)
			if result != tt.expected {
				t.Errorf("capitalizeFirst(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseRefFromUse(t *testing.T) {
	tests := []struct {
		name     string
		useStr   string
		calledAs string
		expected string
	}{
		{"no ref", "add", "add", ""},
		{"with ref", "add[@<ref>]", "add@v0.1.0", "v0.1.0"},
		{"just add", "add", "add", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRefFromUse(tt.useStr, tt.calledAs)
			if result != tt.expected {
				t.Errorf("parseRefFromUse(%q, %q) = %q, expected %q", tt.useStr, tt.calledAs, result, tt.expected)
			}
		})
	}
}
