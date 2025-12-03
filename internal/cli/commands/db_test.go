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

func setupDBTestCommand(t *testing.T) (*cobra.Command, *mocks.MockProjectDetector, *mocks.MockRenderer) {
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

	cmd := NewDBCommand(mockDetector, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	return cobraCmd, mockDetector, mockRenderer
}

func TestNewDBCommand(t *testing.T) {
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	dbCmd := NewDBCommand(mockDetector, factory, flusher)

	if dbCmd == nil {
		t.Fatal("NewDBCommand returned nil")
	}

	if dbCmd.detector != mockDetector {
		t.Error("detector not set correctly")
	}

	if dbCmd.newRenderer == nil {
		t.Error("newRenderer not set")
	}

	if dbCmd.flushRenderer == nil {
		t.Error("flushRenderer not set")
	}
}

func TestDBCommand_Command(t *testing.T) {
	mockDetector := mocks.NewMockProjectDetector(t)
	mockRenderer := mocks.NewMockRenderer(t)

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	dbCmd := NewDBCommand(mockDetector, factory, flusher)
	cobraCmd := dbCmd.Command()

	if cobraCmd == nil {
		t.Fatal("Command() returned nil")
	}

	if cobraCmd.Use != "db" {
		t.Errorf("expected Use 'db', got %q", cobraCmd.Use)
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
}

func TestDBCommand_Run_NoFlags(t *testing.T) {
	cobraCmd, _, _ := setupDBTestCommand(t)

	var outBuf bytes.Buffer
	cobraCmd.SetOut(&outBuf)
	cobraCmd.SetArgs([]string{})

	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	output := outBuf.String()
	if !strings.Contains(output, "db") {
		t.Error("expected help output to contain 'db'")
	}
}

func TestDBCommand_CommandDescriptions(t *testing.T) {
	cobraCmd, _, _ := setupDBTestCommand(t)

	keyPhrases := []string{"Database", "migrations", "Tracks"}
	for _, phrase := range keyPhrases {
		if !strings.Contains(cobraCmd.Long, phrase) {
			t.Errorf("Long description missing mention of %q", phrase)
		}
	}

	if !strings.Contains(cobraCmd.Example, "tracks db migrate") {
		t.Error("Example missing migrate usage pattern")
	}

	if !strings.Contains(cobraCmd.Example, "tracks db rollback") {
		t.Error("Example missing rollback usage pattern")
	}

	if !strings.Contains(cobraCmd.Example, "tracks db status") {
		t.Error("Example missing status usage pattern")
	}

	if !strings.Contains(cobraCmd.Example, "tracks db reset") {
		t.Error("Example missing reset usage pattern")
	}
}
