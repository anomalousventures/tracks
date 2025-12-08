package database

import "testing"

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "postgres with credentials",
			input:    "postgres://myuser:mypassword@localhost:5432/mydb",
			expected: "postgres://****:****@localhost:5432/mydb",
		},
		{
			name:     "postgres with credentials and query params",
			input:    "postgres://user:pass@host:5432/db?sslmode=disable",
			expected: "postgres://****:****@host:5432/db?sslmode=disable",
		},
		{
			name:     "postgres without credentials",
			input:    "postgres://localhost:5432/mydb",
			expected: "postgres://localhost:5432/mydb",
		},
		{
			name:     "sqlite file URL",
			input:    "file:./data/test.db",
			expected: "file:./data/test.db",
		},
		{
			name:     "http URL without credentials",
			input:    "http://localhost:8080",
			expected: "http://localhost:8080",
		},
		{
			name:     "invalid URL returns placeholder",
			input:    "://invalid",
			expected: "[invalid URL]",
		},
		{
			name:     "empty string returns placeholder",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeURL(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
