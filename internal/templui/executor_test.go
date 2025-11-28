package templui

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
)

func TestNewExecutor(t *testing.T) {
	executor := NewExecutor()
	if executor == nil {
		t.Fatal("NewExecutor returned nil")
	}
}

func TestParseComponentList(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   []interfaces.UIComponent
	}{
		{
			name:   "empty output",
			output: "",
			want:   nil,
		},
		{
			name:   "whitespace only",
			output: "   \n\t\n  ",
			want:   nil,
		},
		{
			name:   "single component",
			output: "button",
			want: []interfaces.UIComponent{
				{Name: "button"},
			},
		},
		{
			name:   "multiple components",
			output: "button\ncard\nalert\n",
			want: []interfaces.UIComponent{
				{Name: "button"},
				{Name: "card"},
				{Name: "alert"},
			},
		},
		{
			name:   "components with whitespace",
			output: "  button  \n  card  \n",
			want: []interfaces.UIComponent{
				{Name: "button"},
				{Name: "card"},
			},
		},
		{
			name:   "components with empty lines",
			output: "button\n\ncard\n\nalert",
			want: []interfaces.UIComponent{
				{Name: "button"},
				{Name: "card"},
				{Name: "alert"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseComponentList(tt.output)

			if len(got) != len(tt.want) {
				t.Fatalf("parseComponentList() returned %d components, want %d", len(got), len(tt.want))
			}

			for i, component := range got {
				if component.Name != tt.want[i].Name {
					t.Errorf("component[%d].Name = %q, want %q", i, component.Name, tt.want[i].Name)
				}
			}
		})
	}
}

func TestExecutor_Add_EmptyComponents(t *testing.T) {
	executor := NewExecutor()

	err := executor.Add(t.Context(), ".", []string{}, false)
	if err == nil {
		t.Fatal("expected error for empty components, got nil")
	}

	if err.Error() != "at least one component name required" {
		t.Errorf("expected error message 'at least one component name required', got %q", err.Error())
	}
}
