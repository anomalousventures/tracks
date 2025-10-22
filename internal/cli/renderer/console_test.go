package renderer

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/cli/ui"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestNewConsoleRenderer(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	if renderer == nil {
		t.Fatal("NewConsoleRenderer should return a non-nil renderer")
	}

	if renderer.out == nil {
		t.Error("ConsoleRenderer.out should be set to the provided writer")
	}
}

func TestConsoleRendererImplementsInterface(t *testing.T) {
	var buf bytes.Buffer
	var _ Renderer = NewConsoleRenderer(&buf)
}

func TestConsoleRendererTitle(t *testing.T) {
	originalProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(originalProfile)
	lipgloss.SetColorProfile(termenv.TrueColor)

	tests := []struct {
		name  string
		title string
	}{
		{"simple title", "Welcome"},
		{"title with spaces", "Project Created"},
		{"empty title", ""},
		{"title with special chars", "Success! ✓"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewConsoleRenderer(&buf)

			renderer.Title(tt.title)

			output := buf.String()
			if !strings.Contains(output, tt.title) {
				t.Errorf("Title output should contain %q, got %q", tt.title, output)
			}

			if !strings.HasSuffix(output, "\n") {
				t.Error("Title output should end with newline")
			}
		})
	}
}

func TestConsoleRendererTitleUsesTheme(t *testing.T) {
	originalProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(originalProfile)
	lipgloss.SetColorProfile(termenv.TrueColor)

	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	renderer.Title("Test")

	output := buf.String()
	expected := ui.Theme.Title.Render("Test") + "\n"

	if output != expected {
		t.Errorf("Title should use Theme.Title style\nGot:      %q\nExpected: %q", output, expected)
	}
}

func TestConsoleRendererSection(t *testing.T) {
	tests := []struct {
		name           string
		section        Section
		shouldContain  []string
		shouldNotMatch string
	}{
		{
			name: "section with title and body",
			section: Section{
				Title: "Configuration",
				Body:  "Using default settings",
			},
			shouldContain: []string{"Configuration", "Using default settings"},
		},
		{
			name: "section with title only",
			section: Section{
				Title: "Empty Section",
				Body:  "",
			},
			shouldContain: []string{"Empty Section"},
		},
		{
			name: "section with body only",
			section: Section{
				Title: "",
				Body:  "Just some content",
			},
			shouldContain: []string{"Just some content"},
		},
		{
			name: "empty section",
			section: Section{
				Title: "",
				Body:  "",
			},
			shouldContain:  []string{},
			shouldNotMatch: "should be empty or just newlines",
		},
		{
			name: "section with multiline body",
			section: Section{
				Title: "Details",
				Body:  "Line 1\nLine 2\nLine 3",
			},
			shouldContain: []string{"Details", "Line 1", "Line 2", "Line 3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewConsoleRenderer(&buf)

			renderer.Section(tt.section)

			output := buf.String()
			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Section output should contain %q\nGot: %q", expected, output)
				}
			}
		})
	}
}

func TestConsoleRendererSectionUsesTheme(t *testing.T) {
	originalProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(originalProfile)
	lipgloss.SetColorProfile(termenv.TrueColor)

	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	section := Section{
		Title: "Test Section",
		Body:  "Test body",
	}

	renderer.Section(section)

	output := buf.String()

	expectedTitle := ui.Theme.Title.Render("Test Section")
	if !strings.Contains(output, expectedTitle) {
		t.Error("Section title should use Theme.Title style")
	}

	if !strings.Contains(output, "Test body") {
		t.Error("Section body should be rendered")
	}
}

func TestConsoleRendererFlush(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	err := renderer.Flush()

	if err != nil {
		t.Errorf("Flush should return nil for console renderer, got %v", err)
	}
}

func TestConsoleRendererTableStub(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := Table{
		Headers: []string{"Name", "Type"},
		Rows:    [][]string{{"user.go", "model"}},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Table stub should not panic, got: %v", r)
		}
	}()

	renderer.Table(table)
}

func TestConsoleRendererProgressStub(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := ProgressSpec{
		Label: "Processing",
		Total: 100,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Progress stub should not panic, got: %v", r)
		}
	}()

	progress := renderer.Progress(spec)

	if progress == nil {
		t.Error("Progress should return a non-nil Progress interface")
	}

	progress.Increment(10)
	progress.Done()
}

func TestConsoleRendererOutputGoesToWriter(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	renderer.Title("Test Title")

	if buf.Len() == 0 {
		t.Error("Output should be written to the provided io.Writer")
	}

	output := buf.String()
	if !strings.Contains(output, "Test Title") {
		t.Errorf("Output should contain the rendered title, got %q", output)
	}
}

func TestConsoleRendererMultipleOperations(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	renderer.Title("Main Title")
	renderer.Section(Section{
		Title: "Section 1",
		Body:  "Content 1",
	})
	renderer.Section(Section{
		Title: "Section 2",
		Body:  "Content 2",
	})

	output := buf.String()

	expectedParts := []string{
		"Main Title",
		"Section 1",
		"Content 1",
		"Section 2",
		"Content 2",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("Output should contain %q\nGot: %q", part, output)
		}
	}
}

func TestConsoleRendererRespectsNOCOLOR(t *testing.T) {
	originalNOCOLOR := os.Getenv("NO_COLOR")
	originalProfile := lipgloss.ColorProfile()
	defer func() {
		if originalNOCOLOR != "" {
			os.Setenv("NO_COLOR", originalNOCOLOR)
		} else {
			os.Unsetenv("NO_COLOR")
		}
		lipgloss.SetColorProfile(originalProfile)
	}()

	os.Setenv("NO_COLOR", "1")
	lipgloss.SetColorProfile(termenv.Ascii)

	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	renderer.Title("Test")

	output := buf.String()

	if strings.Contains(output, "\033[") {
		t.Error("Output should not contain ANSI escape codes when NO_COLOR is set")
	}

	if !strings.Contains(output, "Test") {
		t.Error("Output should still contain the text content")
	}
}
