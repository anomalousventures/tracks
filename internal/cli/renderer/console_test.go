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

func TestConsoleRendererTable(t *testing.T) {
	tests := []struct {
		name          string
		table         Table
		shouldContain []string
	}{
		{
			name: "simple table with headers and rows",
			table: Table{
				Headers: []string{"Name", "Type", "Status"},
				Rows: [][]string{
					{"user.go", "model", "created"},
					{"user_test.go", "test", "created"},
				},
			},
			shouldContain: []string{"Name", "Type", "Status", "user.go", "model", "created", "user_test.go", "test"},
		},
		{
			name: "table with two columns",
			table: Table{
				Headers: []string{"File", "Lines"},
				Rows: [][]string{
					{"main.go", "42"},
					{"helper.go", "128"},
				},
			},
			shouldContain: []string{"File", "Lines", "main.go", "42", "helper.go", "128"},
		},
		{
			name: "table with varying content lengths",
			table: Table{
				Headers: []string{"Short", "Very Long Header Name"},
				Rows: [][]string{
					{"A", "B"},
					{"Long content here", "X"},
				},
			},
			shouldContain: []string{"Short", "Very Long Header Name", "A", "B", "Long content here", "X"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewConsoleRenderer(&buf)

			renderer.Table(tt.table)

			output := buf.String()
			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Table output should contain %q\nGot: %q", expected, output)
				}
			}

			if output == "" {
				t.Error("Table should produce output for non-empty table")
			}
		})
	}
}

func TestConsoleRendererTableEmpty(t *testing.T) {
	tests := []struct {
		name  string
		table Table
	}{
		{
			name: "empty table with no rows",
			table: Table{
				Headers: []string{"Col1", "Col2"},
				Rows:    [][]string{},
			},
		},
		{
			name: "table with no headers and no rows",
			table: Table{
				Headers: []string{},
				Rows:    [][]string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewConsoleRenderer(&buf)

			renderer.Table(tt.table)

			output := buf.String()
			if len(tt.table.Headers) == 0 {
				if len(tt.table.Rows) == 0 && output != "" {
					t.Error("Empty table should produce no output")
				}
			}
		})
	}
}

func TestConsoleRendererTableWithEmptyCells(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := Table{
		Headers: []string{"Name", "Value", "Status"},
		Rows: [][]string{
			{"item1", "", "active"},
			{"", "42", ""},
			{"item3", "value3", "pending"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	requiredContent := []string{"Name", "Value", "Status", "item1", "active", "42", "item3", "value3", "pending"}
	for _, content := range requiredContent {
		if !strings.Contains(output, content) {
			t.Errorf("Table should contain %q\nGot: %q", content, output)
		}
	}
}

func TestConsoleRendererTableAlignment(t *testing.T) {
	originalProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(originalProfile)
	lipgloss.SetColorProfile(termenv.TrueColor)

	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := Table{
		Headers: []string{"Short", "Medium Length", "X"},
		Rows: [][]string{
			{"A", "B", "C"},
			{"Very long content", "D", "E"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	if output == "" {
		t.Fatal("Table should produce output")
	}

	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) < 2 {
		t.Errorf("Table should have at least 2 lines (header + 1 row), got %d", len(lines))
	}

	for _, content := range []string{"Short", "Medium Length", "X", "A", "B", "C", "Very long content", "D", "E"} {
		if !strings.Contains(output, content) {
			t.Errorf("Table should contain %q", content)
		}
	}
}

func TestConsoleRendererTableUsesTheme(t *testing.T) {
	originalProfile := lipgloss.ColorProfile()
	defer lipgloss.SetColorProfile(originalProfile)
	lipgloss.SetColorProfile(termenv.TrueColor)

	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := Table{
		Headers: []string{"Header1", "Header2"},
		Rows: [][]string{
			{"row1col1", "row1col2"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	if !strings.Contains(output, "\033[") {
		t.Error("Table headers should contain ANSI escape codes when colors are enabled")
	}

	if !strings.Contains(output, "Header1") || !strings.Contains(output, "Header2") {
		t.Error("Table should contain header text")
	}
}

func TestConsoleRendererTableRespectsNOCOLOR(t *testing.T) {
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

	table := Table{
		Headers: []string{"Name", "Status"},
		Rows: [][]string{
			{"test", "pass"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	if strings.Contains(output, "\033[") {
		t.Error("Table should not contain ANSI escape codes when NO_COLOR is set")
	}

	if !strings.Contains(output, "Name") || !strings.Contains(output, "Status") {
		t.Error("Table should still contain content when NO_COLOR is set")
	}

	if !strings.Contains(output, "test") || !strings.Contains(output, "pass") {
		t.Error("Table should contain row data when NO_COLOR is set")
	}
}

func TestConsoleRendererTableSpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := Table{
		Headers: []string{"Item", "Status", "Note"},
		Rows: [][]string{
			{"✓ Success", "✓", "All good 👍"},
			{"✗ Failed", "✗", "Error ⚠️"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	expectedContent := []string{"Item", "Status", "Note", "✓ Success", "✓", "All good 👍", "✗ Failed", "✗", "Error ⚠️"}
	for _, content := range expectedContent {
		if !strings.Contains(output, content) {
			t.Errorf("Table should contain %q\nGot: %q", content, output)
		}
	}
}

func TestConsoleRendererTableSingleColumn(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := Table{
		Headers: []string{"Files"},
		Rows: [][]string{
			{"main.go"},
			{"helper.go"},
			{"test.go"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	if !strings.Contains(output, "Files") {
		t.Error("Single column table should contain header")
	}

	for _, file := range []string{"main.go", "helper.go", "test.go"} {
		if !strings.Contains(output, file) {
			t.Errorf("Single column table should contain %q", file)
		}
	}
}

func TestConsoleRendererTableSingleRow(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := Table{
		Headers: []string{"Name", "Value", "Status"},
		Rows: [][]string{
			{"single", "row", "here"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	expected := []string{"Name", "Value", "Status", "single", "row", "here"}
	for _, content := range expected {
		if !strings.Contains(output, content) {
			t.Errorf("Single row table should contain %q\nGot: %q", content, output)
		}
	}
}

func TestConsoleRendererTableWritesToWriter(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := Table{
		Headers: []string{"Col1"},
		Rows:    [][]string{{"Data"}},
	}

	renderer.Table(table)

	if buf.Len() == 0 {
		t.Error("Table output should be written to the provided io.Writer")
	}

	output := buf.String()
	if !strings.Contains(output, "Col1") || !strings.Contains(output, "Data") {
		t.Errorf("Output should contain table content, got %q", output)
	}
}

func TestConsoleRendererTableNoHeaders(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := Table{
		Headers: []string{},
		Rows: [][]string{
			{"data1", "data2"},
			{"data3", "data4"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	for _, data := range []string{"data1", "data2", "data3", "data4"} {
		if !strings.Contains(output, data) {
			t.Errorf("Table without headers should still render rows, missing %q\nGot: %q", data, output)
		}
	}
}
