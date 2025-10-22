package ui

import "github.com/charmbracelet/lipgloss"

// Theme provides consistent Lip Gloss styles for console and TUI output modes.
//
// The theme defines 5 core styles for common CLI output patterns:
//   - Title: Bold purple for prominent headings and titles
//   - Success: Green for positive feedback and completion messages
//   - Error: Red for error messages and failures
//   - Warning: Orange for warnings and cautionary messages
//   - Muted: Gray for secondary or subtle text
//
// Lip Gloss automatically respects the NO_COLOR and CLICOLOR environment
// variables for accessibility. When NO_COLOR is set, all styles will render
// without ANSI color codes while preserving text formatting like bold.
//
// Color Palette:
//   - Title: #7D56F4 (Purple) - Bold, attention-grabbing
//   - Success: #04B575 (Green) - Positive, affirmative
//   - Error: #FF4672 (Red) - Alerts, problems
//   - Warning: #FFA657 (Orange) - Caution, non-critical issues
//   - Muted: #626262 (Gray) - Subdued, secondary information
//
// Example usage:
//
//	// Render a success message
//	fmt.Println(Theme.Success.Render("✓ Project created successfully"))
//
//	// Render an error with additional context
//	fmt.Fprintln(os.Stderr, Theme.Error.Render("Error:"), "file not found")
//
//	// Combine styles in output
//	fmt.Printf("%s\n%s\n",
//	    Theme.Title.Render("Configuration"),
//	    Theme.Muted.Render("Using default settings"))
//
// The Theme is used by Renderer implementations to provide consistent
// visual styling across all output modes.
var Theme = struct {
	// Title is used for prominent headings and section titles.
	// Bold purple (#7D56F4) to draw attention.
	Title lipgloss.Style

	// Success is used for positive feedback and completion messages.
	// Green (#04B575) to indicate successful operations.
	Success lipgloss.Style

	// Error is used for error messages and failure notifications.
	// Red (#FF4672) to alert users to problems.
	Error lipgloss.Style

	// Warning is used for warnings and cautionary messages.
	// Orange (#FFA657) to indicate non-critical issues requiring attention.
	Warning lipgloss.Style

	// Muted is used for secondary or less important information.
	// Gray (#626262) to de-emphasize supplementary text.
	Muted lipgloss.Style
}{
	Title:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")),
	Success: lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575")),
	Error:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4672")),
	Warning: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA657")),
	Muted:   lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")),
}
