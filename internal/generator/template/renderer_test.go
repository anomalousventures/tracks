package template

import (
	"embed"
	"testing"
)

// TestRendererInterface verifies that TemplateRenderer implements the Renderer interface
func TestRendererInterface(t *testing.T) {
	var _ Renderer = (*TemplateRenderer)(nil)
}

// TestNewRenderer tests the NewRenderer constructor function
func TestNewRenderer(t *testing.T) {
	var testFS embed.FS
	renderer := NewRenderer(testFS)

	if renderer == nil {
		t.Fatal("NewRenderer returned nil")
	}

	tr, ok := renderer.(*TemplateRenderer)
	if !ok {
		t.Fatal("NewRenderer did not return a *TemplateRenderer")
	}

	if tr.fs != testFS {
		t.Error("TemplateRenderer.fs not set correctly")
	}
}

// TestRendererInterfaceMethods verifies the Renderer interface has the expected methods
func TestRendererInterfaceMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{"Render method exists", "Render"},
		{"RenderToFile method exists", "RenderToFile"},
		{"Validate method exists", "Validate"},
	}

	var testFS embed.FS
	renderer := NewRenderer(testFS)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.method {
			case "Render":
				_, err := renderer.Render("test.tmpl", TemplateData{})
				if err == nil {
					t.Error("expected error for non-existent template, got nil")
				}
			case "RenderToFile":
				err := renderer.RenderToFile("test.tmpl", TemplateData{}, "/tmp/test")
				if err == nil {
					t.Error("expected error for non-existent template, got nil")
				}
			case "Validate":
				err := renderer.Validate("test.tmpl")
				if err == nil {
					t.Error("expected error for non-existent template, got nil")
				}
			}
		})
	}
}
