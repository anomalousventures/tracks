package ui

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestThemeExists(t *testing.T) {
	testText := "test"

	if Theme.Title.Render(testText) == "" {
		t.Error("Theme.Title should be initialized and renderable")
	}
	if Theme.Success.Render(testText) == "" {
		t.Error("Theme.Success should be initialized and renderable")
	}
	if Theme.Error.Render(testText) == "" {
		t.Error("Theme.Error should be initialized and renderable")
	}
	if Theme.Warning.Render(testText) == "" {
		t.Error("Theme.Warning should be initialized and renderable")
	}
	if Theme.Muted.Render(testText) == "" {
		t.Error("Theme.Muted should be initialized and renderable")
	}
}

func TestThemeHasAllStyles(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"Title", Theme.Title},
		{"Success", Theme.Success},
		{"Error", Theme.Error},
		{"Warning", Theme.Warning},
		{"Muted", Theme.Muted},
	}

	testText := "test"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render(testText)
			if rendered == "" {
				t.Errorf("Theme.%s should be initialized and renderable", tt.name)
			}
			if !strings.Contains(rendered, testText) {
				t.Errorf("Theme.%s should render the input text", tt.name)
			}
		})
	}
}

func TestTitleStyleProperties(t *testing.T) {
	rendered := Theme.Title.Render("Test")
	if rendered == "" {
		t.Error("Title.Render should not return empty string")
	}

	if !strings.Contains(rendered, "Test") {
		t.Error("Title.Render should contain the input text")
	}
}

func TestSuccessStyleProperties(t *testing.T) {
	rendered := Theme.Success.Render("Success")
	if rendered == "" {
		t.Error("Success.Render should not return empty string")
	}

	if !strings.Contains(rendered, "Success") {
		t.Error("Success.Render should contain the input text")
	}
}

func TestErrorStyleProperties(t *testing.T) {
	rendered := Theme.Error.Render("Error")
	if rendered == "" {
		t.Error("Error.Render should not return empty string")
	}

	if !strings.Contains(rendered, "Error") {
		t.Error("Error.Render should contain the input text")
	}
}

func TestWarningStyleProperties(t *testing.T) {
	rendered := Theme.Warning.Render("Warning")
	if rendered == "" {
		t.Error("Warning.Render should not return empty string")
	}

	if !strings.Contains(rendered, "Warning") {
		t.Error("Warning.Render should contain the input text")
	}
}

func TestMutedStyleProperties(t *testing.T) {
	rendered := Theme.Muted.Render("Muted")
	if rendered == "" {
		t.Error("Muted.Render should not return empty string")
	}

	if !strings.Contains(rendered, "Muted") {
		t.Error("Muted.Render should contain the input text")
	}
}

func TestThemeWithNOCOLOR(t *testing.T) {
	originalNOCOLOR := os.Getenv("NO_COLOR")
	defer func() {
		if originalNOCOLOR != "" {
			os.Setenv("NO_COLOR", originalNOCOLOR)
		} else {
			os.Unsetenv("NO_COLOR")
		}
	}()

	tests := []struct {
		name     string
		envValue string
		style    lipgloss.Style
		text     string
	}{
		{
			name:     "NO_COLOR=1 with Title style",
			envValue: "1",
			style:    Theme.Title,
			text:     "Title",
		},
		{
			name:     "NO_COLOR=1 with Success style",
			envValue: "1",
			style:    Theme.Success,
			text:     "Success",
		},
		{
			name:     "NO_COLOR=1 with Error style",
			envValue: "1",
			style:    Theme.Error,
			text:     "Error",
		},
		{
			name:     "NO_COLOR=1 with Warning style",
			envValue: "1",
			style:    Theme.Warning,
			text:     "Warning",
		},
		{
			name:     "NO_COLOR=1 with Muted style",
			envValue: "1",
			style:    Theme.Muted,
			text:     "Muted",
		},
		{
			name:     "NO_COLOR empty with Title style",
			envValue: "",
			style:    Theme.Title,
			text:     "Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NO_COLOR", tt.envValue)

			rendered := tt.style.Render(tt.text)

			if rendered == "" {
				t.Error("Style.Render should not return empty string even with NO_COLOR")
			}

			if !strings.Contains(rendered, tt.text) {
				t.Error("Style.Render should contain the input text")
			}

			if tt.envValue != "" {
				if strings.Contains(rendered, "\033[") {
					t.Errorf("Style.Render should not contain ANSI escape codes when NO_COLOR=%s", tt.envValue)
				}
			}
		})
	}
}

func TestThemeStylesRenderText(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
		input string
	}{
		{"Title renders hello", Theme.Title, "hello"},
		{"Success renders world", Theme.Success, "world"},
		{"Error renders oops", Theme.Error, "oops"},
		{"Warning renders caution", Theme.Warning, "caution"},
		{"Muted renders subtle", Theme.Muted, "subtle"},
		{"Title renders empty", Theme.Title, ""},
		{"Title renders multiline", Theme.Title, "line1\nline2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered := tt.style.Render(tt.input)

			if tt.input != "" && !strings.Contains(rendered, tt.input) {
				t.Errorf("Rendered output should contain input text %q, got %q", tt.input, rendered)
			}
		})
	}
}

func TestThemeStylesAreIndependent(t *testing.T) {
	text := "Same Text"

	titleOutput := Theme.Title.Render(text)
	successOutput := Theme.Success.Render(text)
	errorOutput := Theme.Error.Render(text)
	warningOutput := Theme.Warning.Render(text)
	mutedOutput := Theme.Muted.Render(text)

	allOutputs := []string{titleOutput, successOutput, errorOutput, warningOutput, mutedOutput}

	for i, output := range allOutputs {
		if !strings.Contains(output, text) {
			t.Errorf("Output %d should contain the text %q", i, text)
		}
	}
}

func TestThemeCanBeUsedMultipleTimes(t *testing.T) {
	first := Theme.Title.Render("First")
	second := Theme.Title.Render("Second")
	third := Theme.Title.Render("Third")

	if !strings.Contains(first, "First") {
		t.Error("First render should contain 'First'")
	}
	if !strings.Contains(second, "Second") {
		t.Error("Second render should contain 'Second'")
	}
	if !strings.Contains(third, "Third") {
		t.Error("Third render should contain 'Third'")
	}
}
