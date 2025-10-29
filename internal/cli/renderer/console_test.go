package renderer

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestNewConsoleRenderer(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	if renderer == nil {
		t.Fatal("NewConsoleRenderer should return non-nil renderer")
	}

	if renderer.out != &buf {
		t.Error("NewConsoleRenderer should use provided writer")
	}
}

func TestConsoleRendererImplementsInterface(t *testing.T) {
	var _ interfaces.Renderer = &ConsoleRenderer{}
}

func TestConsoleRendererTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{"simple title", "Welcome"},
		{"title with spaces", "Hello World"},
		{"empty title", ""},
		{"title with special chars", "Project: test-app!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewConsoleRenderer(&buf)

			renderer.Title(tt.title)

			output := buf.String()
			if tt.title != "" && !strings.Contains(output, tt.title) {
				t.Errorf("Title should contain %q\nGot: %q", tt.title, output)
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

	renderer.Title("Test Title")

	output := buf.String()

	if !strings.Contains(output, "Test Title") {
		t.Error("Output should contain the title text")
	}

	if !strings.Contains(output, "\033[") {
		t.Error("Output should contain ANSI escape codes when colors are enabled")
	}
}

func TestConsoleRendererSection(t *testing.T) {
	tests := []struct {
		name    string
		section interfaces.Section
		want    []string
	}{
		{
			name:    "section with title and body",
			section: interfaces.Section{Title: "Config", Body: "Using Chi router"},
			want:    []string{"Config", "Using Chi router"},
		},
		{
			name:    "section with title only",
			section: interfaces.Section{Title: "Settings"},
			want:    []string{"Settings"},
		},
		{
			name:    "section with body only",
			section: interfaces.Section{Body: "Database configured"},
			want:    []string{"Database configured"},
		},
		{
			name:    "empty section",
			section: interfaces.Section{},
			want:    []string{},
		},
		{
			name:    "section with multiline body",
			section: interfaces.Section{Title: "Info", Body: "Line 1\nLine 2"},
			want:    []string{"Info", "Line 1", "Line 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewConsoleRenderer(&buf)

			renderer.Section(tt.section)

			output := buf.String()
			for _, expected := range tt.want {
				if !strings.Contains(output, expected) {
					t.Errorf("Section should contain %q\nGot: %q", expected, output)
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

	section := interfaces.Section{
		Title: "Section Title",
		Body:  "Section body content",
	}

	renderer.Section(section)

	output := buf.String()

	if !strings.Contains(output, "Section Title") {
		t.Error("Output should contain section title")
	}

	if !strings.Contains(output, "Section body content") {
		t.Error("Output should contain section body")
	}

	if !strings.Contains(output, "\033[") {
		t.Error("Output should contain ANSI escape codes for the title")
	}
}

func TestConsoleRendererFlush(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	err := renderer.Flush()

	if err != nil {
		t.Errorf("Flush should not return error, got: %v", err)
	}
}

func TestConsoleRendererTableStub(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := interfaces.Table{
		Headers: []string{"Name", "Value"},
		Rows: [][]string{
			{"key1", "value1"},
		},
	}

	renderer.Table(table)
}

func TestConsoleRendererProgressStub(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := interfaces.ProgressSpec{
		Label: "Downloading",
		Total: 100,
	}

	progress := renderer.Progress(spec)

	if progress == nil {
		t.Error("Progress should return non-nil Progress")
	}

	progress.Increment(50)
	progress.Done()
}

func TestConsoleRendererOutputGoesToWriter(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	renderer.Title("Test")
	renderer.Section(interfaces.Section{Title: "Section", Body: "Body"})

	output := buf.String()

	if output == "" {
		t.Error("Output should be written to the provided writer")
	}
}

func TestConsoleRendererMultipleOperations(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	renderer.Title("Title")
	renderer.Section(interfaces.Section{Body: "Content"})
	renderer.Flush()

	output := buf.String()

	if !strings.Contains(output, "Title") {
		t.Error("Output should contain title")
	}

	if !strings.Contains(output, "Content") {
		t.Error("Output should contain section content")
	}
}

func TestConsoleRendererRespectsNOCOLOR(t *testing.T) {
	originalNOCOLOR := os.Getenv("NO_COLOR")
	defer func() {
		if originalNOCOLOR != "" {
			os.Setenv("NO_COLOR", originalNOCOLOR)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	os.Setenv("NO_COLOR", "1")

	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	renderer.Title("Test Title")

	output := buf.String()

	if !strings.Contains(output, "Test Title") {
		t.Error("Output should still contain the text content")
	}
}

func TestConsoleRendererTable(t *testing.T) {
	tests := []struct {
		name             string
		table interfaces.Table
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "simple table with headers and rows",
			table: interfaces.Table{
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
			table: interfaces.Table{
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
			table: interfaces.Table{
				Headers: []string{"Short", "Very Long Header Name"},
				Rows: [][]string{
					{"A", "B"},
					{"Long content here", "X"},
				},
			},
			shouldContain: []string{"Short", "Very Long Header Name", "A", "B", "Long content here", "X"},
		},
		{
			name: "table with extra columns in rows - extra columns ignored",
			table: interfaces.Table{
				Headers: []string{"Name", "Status"},
				Rows: [][]string{
					{"item1", "active", "ignored"},
					{"item2", "pending", "also-ignored", "more-ignored"},
				},
			},
			shouldContain:    []string{"Name", "Status", "item1", "active", "item2", "pending"},
			shouldNotContain: []string{"ignored", "also-ignored", "more-ignored"},
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

			for _, notExpected := range tt.shouldNotContain {
				if strings.Contains(output, notExpected) {
					t.Errorf("Table output should not contain %q\nGot: %q", notExpected, output)
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
		name              string
		table interfaces.Table
		shouldHaveOutput  bool
		shouldContain     []string
		shouldNotContain  []string
	}{
		{
			name: "empty table with no rows - headers displayed",
			table: interfaces.Table{
				Headers: []string{"Col1", "Col2"},
				Rows:    [][]string{},
			},
			shouldHaveOutput: true,
			shouldContain:    []string{"Col1", "Col2"},
		},
		{
			name: "table with no headers and no rows",
			table: interfaces.Table{
				Headers: []string{},
				Rows:    [][]string{},
			},
			shouldHaveOutput: false,
		},
		{
			name: "table with empty row slices",
			table: interfaces.Table{
				Headers: []string{},
				Rows:    [][]string{{}},
			},
			shouldHaveOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewConsoleRenderer(&buf)

			renderer.Table(tt.table)

			output := buf.String()

			if tt.shouldHaveOutput && output == "" {
				t.Error("Expected output but got empty string")
			}

			if !tt.shouldHaveOutput && output != "" {
				t.Errorf("Expected no output but got: %q", output)
			}

			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Output should contain %q\nGot: %q", expected, output)
				}
			}

			for _, notExpected := range tt.shouldNotContain {
				if strings.Contains(output, notExpected) {
					t.Errorf("Output should not contain %q\nGot: %q", notExpected, output)
				}
			}
		})
	}
}

func TestConsoleRendererTableWithEmptyCells(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := interfaces.Table{
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

	table := interfaces.Table{
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
	if len(lines) < 3 {
		t.Errorf("Table should have at least 3 lines (header + 2 rows), got %d", len(lines))
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

	table := interfaces.Table{
		Headers: []string{"Header1", "Header2"},
		Rows: [][]string{
			{"data1", "data2"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	if !strings.Contains(output, "\033[") {
		t.Error("Table headers should use Theme colors (ANSI codes present)")
	}

	if !strings.Contains(output, "Header1") || !strings.Contains(output, "Header2") {
		t.Error("Table should contain headers")
	}
}

func TestConsoleRendererTableRespectsNOCOLOR(t *testing.T) {
	originalNOCOLOR := os.Getenv("NO_COLOR")
	defer func() {
		if originalNOCOLOR != "" {
			os.Setenv("NO_COLOR", originalNOCOLOR)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	os.Setenv("NO_COLOR", "1")

	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := interfaces.Table{
		Headers: []string{"Header1", "Header2"},
		Rows: [][]string{
			{"data1", "data2"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	if !strings.Contains(output, "Header1") || !strings.Contains(output, "Header2") {
		t.Error("Table should still render headers with NO_COLOR")
	}

	if !strings.Contains(output, "data1") || !strings.Contains(output, "data2") {
		t.Error("Table should still render data with NO_COLOR")
	}
}

func TestConsoleRendererTableSpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := interfaces.Table{
		Headers: []string{"Emoji", "Symbols", "Spaces"},
		Rows: [][]string{
			{"✓", "→", "  spaced  "},
			{"★", "←", "multi word"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	specialChars := []string{"✓", "→", "★", "←", "spaced", "multi word"}
	for _, char := range specialChars {
		if !strings.Contains(output, char) {
			t.Errorf("Table should handle special character %q", char)
		}
	}
}

func TestConsoleRendererTableSingleColumn(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := interfaces.Table{
		Headers: []string{"OnlyColumn"},
		Rows: [][]string{
			{"value1"},
			{"value2"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	if !strings.Contains(output, "OnlyColumn") {
		t.Error("Single column table should render header")
	}

	if !strings.Contains(output, "value1") || !strings.Contains(output, "value2") {
		t.Error("Single column table should render all rows")
	}
}

func TestConsoleRendererTableSingleRow(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := interfaces.Table{
		Headers: []string{"Col1", "Col2", "Col3"},
		Rows: [][]string{
			{"a", "b", "c"},
		},
	}

	renderer.Table(table)

	output := buf.String()

	for _, val := range []string{"Col1", "Col2", "Col3", "a", "b", "c"} {
		if !strings.Contains(output, val) {
			t.Errorf("Single row table should contain %q", val)
		}
	}
}

func TestConsoleRendererTableWritesToWriter(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := interfaces.Table{
		Headers: []string{"Test"},
		Rows:    [][]string{{"data"}},
	}

	renderer.Table(table)

	if buf.Len() == 0 {
		t.Error("Table should write to the provided writer")
	}
}

func TestConsoleRendererTableNoHeaders(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	table := interfaces.Table{
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

func TestConsoleRendererProgressReturnsImplementation(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := interfaces.ProgressSpec{
		Label: "Downloading",
		Total: 100,
	}

	progress := renderer.Progress(spec)

	if progress == nil {
		t.Fatal("Progress should return non-nil Progress")
	}

	if _, ok := progress.(*ConsoleProgress); !ok {
		t.Error("Progress should return ConsoleProgress implementation")
	}
}

func TestConsoleProgressIncrement(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := interfaces.ProgressSpec{
		Label: "Processing",
		Total: 100,
	}

	progress := renderer.Progress(spec)
	progress.Increment(25)

	output := buf.String()

	if output == "" {
		t.Error("Increment should write progress bar to output")
	}

	if !strings.Contains(output, "\r") {
		t.Error("Progress output should contain \"\r\" for in-place updates")
	}
}

func TestConsoleProgressDone(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := interfaces.ProgressSpec{
		Label: "Uploading",
		Total: 100,
	}

	progress := renderer.Progress(spec)
	progress.Increment(100)
	progress.Done()

	output := buf.String()

	if !strings.Contains(output, "\n") {
		t.Error("Done should add a newline to complete the progress bar")
	}
}

func TestConsoleProgressInPlaceUpdates(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := interfaces.ProgressSpec{
		Label: "Loading",
		Total: 100,
	}

	progress := renderer.Progress(spec)
	progress.Increment(10)

	firstOutput := buf.String()

	progress.Increment(20)
	secondOutput := buf.String()

	if !strings.Contains(firstOutput, "\r") {
		t.Error("First increment should use \"\r\" for in-place update")
	}

	if !strings.Contains(secondOutput, "\r") {
		t.Error("Second increment should use \"\r\" for in-place update")
	}

	if len(secondOutput) <= len(firstOutput) {
		t.Error("Second increment should append to output (multiple \"\r\" lines)")
	}
}

func TestConsoleProgressMultipleIncrements(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := interfaces.ProgressSpec{
		Label: "Copying",
		Total: 100,
	}

	progress := renderer.Progress(spec)

	progress.Increment(10)
	progress.Increment(20)
	progress.Increment(30)

	output := buf.String()

	rCount := strings.Count(output, "\r")
	if rCount < 3 {
		t.Errorf("Multiple increments should produce multiple \"\r\" lines, got %d", rCount)
	}
}

func TestConsoleProgressCompletesAt100Percent(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := interfaces.ProgressSpec{
		Label: "Finalizing",
		Total: 100,
	}

	progress := renderer.Progress(spec)
	progress.Increment(100)
	progress.Done()

	output := buf.String()

	if output == "" {
		t.Error("Completed progress should have output")
	}

	if !strings.Contains(output, "\n") {
		t.Error("Completed progress should have newline from Done()")
	}
}

func TestConsoleProgressZeroTotal(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := interfaces.ProgressSpec{
		Label: "Testing",
		Total: 0,
	}

	progress := renderer.Progress(spec)
	progress.Increment(10)

	output := buf.String()

	if strings.Contains(output, "NaN") || strings.Contains(output, "Inf") {
		t.Error("Zero total should not produce NaN or Inf in output")
	}
}

func TestConsoleProgressIncrementBeyondTotal(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := interfaces.ProgressSpec{
		Label: "Overflowing",
		Total: 100,
	}

	progress := renderer.Progress(spec)
	progress.Increment(150)

	output := buf.String()

	if output == "" {
		t.Error("Progress beyond total should still produce output")
	}
}

func TestConsoleProgressDoneMultipleTimes(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := interfaces.ProgressSpec{
		Label: "Repeating",
		Total: 100,
	}

	progress := renderer.Progress(spec)
	progress.Increment(100)
	progress.Done()

	firstOutput := buf.String()
	firstNewlineCount := strings.Count(firstOutput, "\n")

	progress.Done()
	secondOutput := buf.String()
	secondNewlineCount := strings.Count(secondOutput, "\n")

	if secondNewlineCount > firstNewlineCount {
		t.Error("Calling Done() multiple times should be idempotent (no additional newlines)")
	}
}

func TestConsoleProgressDisplaysLabel(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := interfaces.ProgressSpec{
		Label: "Downloading",
		Total: 100,
	}

	progress := renderer.Progress(spec)
	progress.Increment(50)

	output := buf.String()

	if !strings.Contains(output, "Downloading") {
		t.Error("Progress output should contain the label")
	}

	if !strings.Contains(output, ":") {
		t.Error("Progress output should contain colon separator after label")
	}
}

func TestConsoleProgressWithoutLabel(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewConsoleRenderer(&buf)

	spec := interfaces.ProgressSpec{
		Label: "",
		Total: 100,
	}

	progress := renderer.Progress(spec)
	progress.Increment(50)

	output := buf.String()

	if output == "" {
		t.Error("Progress output should not be empty even without label")
	}

	if !strings.Contains(output, "\r") {
		t.Error("Progress output should contain carriage return for in-place updates")
	}
}
