package renderer

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
)

func TestSectionStruct(t *testing.T) {
	tests := []struct {
		name      string
		section   interfaces.Section
		wantTitle string
		wantBody  string
	}{
		{
			name:      "zero value",
			section:   interfaces.Section{},
			wantTitle: "",
			wantBody:  "",
		},
		{
			name:      "with title only",
			section:   interfaces.Section{Title: "Test Title"},
			wantTitle: "Test Title",
			wantBody:  "",
		},
		{
			name:      "with body only",
			section:   interfaces.Section{Body: "Test body content"},
			wantTitle: "",
			wantBody:  "Test body content",
		},
		{
			name:      "with title and body",
			section:   interfaces.Section{Title: "Overview", Body: "This is a description"},
			wantTitle: "Overview",
			wantBody:  "This is a description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.section.Title != tt.wantTitle {
				t.Errorf("Section.Title = %q, want %q", tt.section.Title, tt.wantTitle)
			}
			if tt.section.Body != tt.wantBody {
				t.Errorf("Section.Body = %q, want %q", tt.section.Body, tt.wantBody)
			}
		})
	}
}

func TestTableStruct(t *testing.T) {
	tests := []struct {
		name        string
		table       interfaces.Table
		wantHeaders []string
		wantRows    [][]string
	}{
		{
			name:        "zero value",
			table:       interfaces.Table{},
			wantHeaders: nil,
			wantRows:    nil,
		},
		{
			name:        "with headers only",
			table:       interfaces.Table{Headers: []string{"Name", "Age"}},
			wantHeaders: []string{"Name", "Age"},
			wantRows:    nil,
		},
		{
			name: "with headers and rows",
			table: interfaces.Table{
				Headers: []string{"Name", "Age", "City"},
				Rows: [][]string{
					{"Alice", "30", "NYC"},
					{"Bob", "25", "LA"},
				},
			},
			wantHeaders: []string{"Name", "Age", "City"},
			wantRows: [][]string{
				{"Alice", "30", "NYC"},
				{"Bob", "25", "LA"},
			},
		},
		{
			name: "with empty rows",
			table: interfaces.Table{
				Headers: []string{"Col1", "Col2"},
				Rows:    [][]string{},
			},
			wantHeaders: []string{"Col1", "Col2"},
			wantRows:    [][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.table.Headers) != len(tt.wantHeaders) {
				t.Errorf("Table.Headers length = %d, want %d", len(tt.table.Headers), len(tt.wantHeaders))
			}
			for i, h := range tt.table.Headers {
				if h != tt.wantHeaders[i] {
					t.Errorf("Table.Headers[%d] = %q, want %q", i, h, tt.wantHeaders[i])
				}
			}

			if len(tt.table.Rows) != len(tt.wantRows) {
				t.Errorf("Table.Rows length = %d, want %d", len(tt.table.Rows), len(tt.wantRows))
			}
			for i, row := range tt.table.Rows {
				if len(row) != len(tt.wantRows[i]) {
					t.Errorf("Table.Rows[%d] length = %d, want %d", i, len(row), len(tt.wantRows[i]))
				}
				for j, cell := range row {
					if cell != tt.wantRows[i][j] {
						t.Errorf("Table.Rows[%d][%d] = %q, want %q", i, j, cell, tt.wantRows[i][j])
					}
				}
			}
		})
	}
}

func TestProgressSpecStruct(t *testing.T) {
	tests := []struct {
		name      string
		spec      interfaces.ProgressSpec
		wantLabel string
		wantTotal int64
	}{
		{
			name:      "zero value",
			spec:      interfaces.ProgressSpec{},
			wantLabel: "",
			wantTotal: 0,
		},
		{
			name:      "with label only",
			spec:      interfaces.ProgressSpec{Label: "Downloading"},
			wantLabel: "Downloading",
			wantTotal: 0,
		},
		{
			name:      "with total only",
			spec:      interfaces.ProgressSpec{Total: 100},
			wantLabel: "",
			wantTotal: 100,
		},
		{
			name:      "with label and total",
			spec:      interfaces.ProgressSpec{Label: "Processing files", Total: 42},
			wantLabel: "Processing files",
			wantTotal: 42,
		},
		{
			name:      "with large total",
			spec:      interfaces.ProgressSpec{Label: "Large operation", Total: 1000000},
			wantLabel: "Large operation",
			wantTotal: 1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.spec.Label != tt.wantLabel {
				t.Errorf("ProgressSpec.Label = %q, want %q", tt.spec.Label, tt.wantLabel)
			}
			if tt.spec.Total != tt.wantTotal {
				t.Errorf("ProgressSpec.Total = %d, want %d", tt.spec.Total, tt.wantTotal)
			}
		})
	}
}

type mockRenderer struct{}

func (m *mockRenderer) Title(s string)                                             {}
func (m *mockRenderer) Section(sec interfaces.Section)                             {}
func (m *mockRenderer) Table(t interfaces.Table)                                   {}
func (m *mockRenderer) Progress(spec interfaces.ProgressSpec) interfaces.Progress  { return &mockProgress{} }
func (m *mockRenderer) Flush() error                                               { return nil }

type mockProgress struct{}

func (m *mockProgress) Increment(n int64) {}
func (m *mockProgress) Done()              {}

func TestRendererInterface(t *testing.T) {
	var _ interfaces.Renderer = (*mockRenderer)(nil)
}

func TestProgressInterface(t *testing.T) {
	var _ interfaces.Progress = (*mockProgress)(nil)
}

func TestRendererInterfaceMethods(t *testing.T) {
	r := &mockRenderer{}

	t.Run("Title method exists", func(t *testing.T) {
		r.Title("test")
	})

	t.Run("Section method exists", func(t *testing.T) {
		r.Section(interfaces.Section{Title: "test", Body: "body"})
	})

	t.Run("Table method exists", func(t *testing.T) {
		r.Table(interfaces.Table{Headers: []string{"h1"}, Rows: [][]string{{"r1"}}})
	})

	t.Run("Progress method exists and returns Progress interface", func(t *testing.T) {
		p := r.Progress(interfaces.ProgressSpec{Label: "test", Total: 100})
		if p == nil {
			t.Error("Progress() returned nil")
		}
	})

	t.Run("Flush method exists", func(t *testing.T) {
		err := r.Flush()
		if err != nil {
			t.Errorf("Flush() returned unexpected error: %v", err)
		}
	})
}

func TestProgressInterfaceMethods(t *testing.T) {
	p := &mockProgress{}

	t.Run("Increment method exists", func(t *testing.T) {
		p.Increment(10)
		p.Increment(50)
		p.Increment(100)
	})

	t.Run("Done method exists", func(t *testing.T) {
		p.Done()
	})
}
