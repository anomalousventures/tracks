package cli

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
)

func TestNewLogger(t *testing.T) {
	// Save original global level and restore after test
	originalLevel := zerolog.GlobalLevel()
	t.Cleanup(func() {
		zerolog.SetGlobalLevel(originalLevel)
	})

	tests := []struct {
		name     string
		logLevel string
		wantNil  bool
	}{
		{
			name:     "debug level",
			logLevel: "debug",
			wantNil:  false,
		},
		{
			name:     "info level",
			logLevel: "info",
			wantNil:  false,
		},
		{
			name:     "warn level",
			logLevel: "warn",
			wantNil:  false,
		},
		{
			name:     "error level",
			logLevel: "error",
			wantNil:  false,
		},
		{
			name:     "off level",
			logLevel: "off",
			wantNil:  false,
		},
		{
			name:     "invalid level defaults to disabled",
			logLevel: "invalid",
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.logLevel)
			var buf bytes.Buffer
			logger = logger.Output(&buf)
			logger.Info().Msg("test")

			if tt.logLevel == "off" || tt.logLevel == "invalid" {
				if buf.Len() > 0 {
					t.Errorf("Expected no output for level %q, got output", tt.logLevel)
				}
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected zerolog.Level
	}{
		{
			name:     "debug",
			level:    "debug",
			expected: zerolog.DebugLevel,
		},
		{
			name:     "info",
			level:    "info",
			expected: zerolog.InfoLevel,
		},
		{
			name:     "warn",
			level:    "warn",
			expected: zerolog.WarnLevel,
		},
		{
			name:     "error",
			level:    "error",
			expected: zerolog.ErrorLevel,
		},
		{
			name:     "off",
			level:    "off",
			expected: zerolog.Disabled,
		},
		{
			name:     "invalid defaults to disabled",
			level:    "invalid",
			expected: zerolog.Disabled,
		},
		{
			name:     "empty defaults to disabled",
			level:    "",
			expected: zerolog.Disabled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLogLevel(tt.level)
			if got != tt.expected {
				t.Errorf("parseLogLevel(%q) = %v, want %v", tt.level, got, tt.expected)
			}
		})
	}
}
