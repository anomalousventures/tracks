package renderer

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewJSONRenderer(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewJSONRenderer(&buf)

	if renderer == nil {
		t.Fatal("NewJSONRenderer should return non-nil renderer")
	}

	if _, ok := interface{}(renderer).(*JSONRenderer); !ok {
		t.Error("NewJSONRenderer should return *JSONRenderer")
	}
}

func TestJSONRendererImplementsInterface(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewJSONRenderer(&buf)

	var _ Renderer = renderer

	if renderer == nil {
		t.Fatal("JSONRenderer should implement Renderer interface")
	}
}

func TestJSONRendererTitle(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewJSONRenderer(&buf)

	renderer.Title("Test Title")
	err := renderer.Flush()

	if err != nil {
		t.Fatalf("Flush should not return error: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}

	if result["title"] != "Test Title" {
		t.Errorf("Expected title 'Test Title', got %v", result["title"])
	}
}

func TestJSONRendererSection(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewJSONRenderer(&buf)

	renderer.Section(Section{
		Title: "Section 1",
		Body:  "Section body",
	})
	err := renderer.Flush()

	if err != nil {
		t.Fatalf("Flush should not return error: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}

	sections, ok := result["sections"].([]interface{})
	if !ok {
		t.Fatal("sections should be an array")
	}

	if len(sections) != 1 {
		t.Fatalf("Expected 1 section, got %d", len(sections))
	}

	section := sections[0].(map[string]interface{})
	if section["title"] != "Section 1" {
		t.Errorf("Expected section title 'Section 1', got %v", section["title"])
	}
	if section["body"] != "Section body" {
		t.Errorf("Expected section body 'Section body', got %v", section["body"])
	}
}

func TestJSONRendererTable(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewJSONRenderer(&buf)

	renderer.Table(Table{
		Headers: []string{"Name", "Status"},
		Rows: [][]string{
			{"file1.go", "created"},
			{"file2.go", "modified"},
		},
	})
	err := renderer.Flush()

	if err != nil {
		t.Fatalf("Flush should not return error: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}

	tables, ok := result["tables"].([]interface{})
	if !ok {
		t.Fatal("tables should be an array")
	}

	if len(tables) != 1 {
		t.Fatalf("Expected 1 table, got %d", len(tables))
	}

	table := tables[0].(map[string]interface{})
	headers := table["headers"].([]interface{})
	if len(headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(headers))
	}

	rows := table["rows"].([]interface{})
	if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}
}

func TestJSONRendererProgress(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewJSONRenderer(&buf)

	progress := renderer.Progress(ProgressSpec{
		Label: "Processing",
		Total: 100,
	})

	if progress == nil {
		t.Fatal("Progress should return non-nil Progress")
	}

	progress.Increment(50)
	progress.Increment(50)
	progress.Done()

	if buf.Len() > 0 {
		t.Error("Progress updates should not write to output (no-op implementation)")
	}
}

func TestJSONRendererFlush(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewJSONRenderer(&buf)

	err := renderer.Flush()
	if err != nil {
		t.Fatalf("Flush should not return error: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}
}

func TestJSONRendererFlushWithAllFields(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewJSONRenderer(&buf)

	renderer.Title("Project Created")
	renderer.Section(Section{Title: "Config", Body: "Using Chi"})
	renderer.Table(Table{
		Headers: []string{"File", "Status"},
		Rows:    [][]string{{"user.go", "created"}},
	})

	err := renderer.Flush()
	if err != nil {
		t.Fatalf("Flush should not return error: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}

	if result["title"] != "Project Created" {
		t.Errorf("Expected title 'Project Created', got %v", result["title"])
	}

	if _, ok := result["sections"]; !ok {
		t.Error("Expected sections field in output")
	}

	if _, ok := result["tables"]; !ok {
		t.Error("Expected tables field in output")
	}
}

func TestJSONRendererMultipleSections(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewJSONRenderer(&buf)

	renderer.Section(Section{Title: "Section 1", Body: "Body 1"})
	renderer.Section(Section{Title: "Section 2", Body: "Body 2"})
	renderer.Section(Section{Title: "Section 3", Body: "Body 3"})

	err := renderer.Flush()
	if err != nil {
		t.Fatalf("Flush should not return error: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}

	sections, ok := result["sections"].([]interface{})
	if !ok {
		t.Fatal("sections should be an array")
	}

	if len(sections) != 3 {
		t.Errorf("Expected 3 sections, got %d", len(sections))
	}
}

func TestJSONRendererMultipleTables(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewJSONRenderer(&buf)

	renderer.Table(Table{Headers: []string{"A"}, Rows: [][]string{{"1"}}})
	renderer.Table(Table{Headers: []string{"B"}, Rows: [][]string{{"2"}}})

	err := renderer.Flush()
	if err != nil {
		t.Fatalf("Flush should not return error: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}

	tables, ok := result["tables"].([]interface{})
	if !ok {
		t.Fatal("tables should be an array")
	}

	if len(tables) != 2 {
		t.Errorf("Expected 2 tables, got %d", len(tables))
	}
}

func TestJSONRendererEmptyFlush(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewJSONRenderer(&buf)

	err := renderer.Flush()
	if err != nil {
		t.Fatalf("Flush should not return error: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Flush should produce output even with no data")
	}

	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		t.Fatalf("Output should be valid JSON: %v", err)
	}
}

func TestJSONRendererOutputFormatted(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewJSONRenderer(&buf)

	renderer.Title("Test")
	err := renderer.Flush()

	if err != nil {
		t.Fatalf("Flush should not return error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "\n") {
		t.Error("JSON output should be formatted with newlines")
	}

	if !strings.Contains(output, "  ") {
		t.Error("JSON output should be indented")
	}
}
