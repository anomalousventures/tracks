package template

import (
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvExampleTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name        string
		projectName string
		dbDriver    string
		envPrefix   string
		expectedURL string
	}{
		{
			name:        "go-libsql driver",
			projectName: "myapp",
			dbDriver:    "go-libsql",
			envPrefix:   "APP",
			expectedURL: "APP_DATABASE_URL=file:./myapp.db",
		},
		{
			name:        "sqlite3 driver",
			projectName: "testapp",
			dbDriver:    "sqlite3",
			envPrefix:   "APP",
			expectedURL: "APP_DATABASE_URL=./testapp.db",
		},
		{
			name:        "postgres driver",
			projectName: "webapp",
			dbDriver:    "postgres",
			envPrefix:   "APP",
			expectedURL: "APP_DATABASE_URL=postgres://localhost/webapp?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				ProjectName: tt.projectName,
				DBDriver:    tt.dbDriver,
				EnvPrefix:   tt.envPrefix,
			}

			result, err := renderer.Render(".env.example.tmpl", data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			assert.Contains(t, result, "⚠️")
			assert.Contains(t, result, "WARNING")
			assert.Contains(t, result, "NEVER COMMIT")

			assert.Contains(t, result, tt.expectedURL)

			assert.Contains(t, result, "SECRET_KEY=")
			assert.Contains(t, result, "APP_ENVIRONMENT=")
			assert.Contains(t, result, "APP_LOGGING_LEVEL=")
			assert.Contains(t, result, "APP_LOGGING_FORMAT=")
			assert.Contains(t, result, "APP_SERVER_PORT=")
		})
	}
}

func TestEnvExampleSecurityWarning(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
		DBDriver:    "sqlite3",
		EnvPrefix:   "APP",
	}

	result, err := renderer.Render(".env.example.tmpl", data)
	require.NoError(t, err)

	expectedWarnings := []string{
		"⚠️",
		"WARNING",
		"NEVER COMMIT",
		".env",
		"VERSION CONTROL",
		"gitignored",
		"secrets",
	}

	for _, warning := range expectedWarnings {
		assert.Contains(t, result, warning, "should contain security warning: %s", warning)
	}
}

func TestEnvExampleRequiredVariables(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
		DBDriver:    "sqlite3",
		EnvPrefix:   "APP",
	}

	result, err := renderer.Render(".env.example.tmpl", data)
	require.NoError(t, err)

	requiredVariables := []string{
		"APP_ENVIRONMENT=",
		"APP_LOGGING_LEVEL=",
		"APP_LOGGING_FORMAT=",
		"APP_SERVER_PORT=",
		"APP_DATABASE_URL=",
		"SECRET_KEY=",
	}

	for _, variable := range requiredVariables {
		assert.Contains(t, result, variable, "should contain variable: %s", variable)
	}
}

func TestEnvExampleSecretKeyGuidance(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
		DBDriver:    "sqlite3",
		EnvPrefix:   "APP",
	}

	result, err := renderer.Render(".env.example.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "SECRET_KEY=")
	assert.Contains(t, result, "openssl rand")
	assert.Contains(t, result, "replace")
}

func TestEnvExampleDifferentProjectNames(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		projectName string
		driver      string
	}{
		{"myapp", "go-libsql"},
		{"cool-project", "sqlite3"},
		{"web-service", "postgres"},
	}

	for _, tt := range tests {
		t.Run(tt.projectName, func(t *testing.T) {
			data := TemplateData{
				ProjectName: tt.projectName,
				DBDriver:    tt.driver,
				EnvPrefix:   "APP",
			}

			result, err := renderer.Render(".env.example.tmpl", data)
			require.NoError(t, err)

			assert.Contains(t, result, tt.projectName, "database URL should contain project name")
		})
	}
}

func TestEnvExamplePostgresSecurityNote(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ProjectName: "myapp",
		DBDriver:    "postgres",
		EnvPrefix:   "APP",
	}

	result, err := renderer.Render(".env.example.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "production")
	assert.Contains(t, result, "sslmode")
}

func TestEnvExampleCustomPrefix(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name      string
		envPrefix string
	}{
		{"default APP prefix", "APP"},
		{"custom MYAPP prefix", "MYAPP"},
		{"custom USERSERVICE prefix", "USERSERVICE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				ProjectName: "myapp",
				DBDriver:    "sqlite3",
				EnvPrefix:   tt.envPrefix,
			}

			result, err := renderer.Render(".env.example.tmpl", data)
			require.NoError(t, err)

			// Verify prefix is used in variable names
			assert.Contains(t, result, tt.envPrefix+"_DATABASE_URL=")
			assert.Contains(t, result, tt.envPrefix+"_SERVER_PORT=")
			assert.Contains(t, result, tt.envPrefix+"_LOGGING_LEVEL=")
			assert.Contains(t, result, "Environment variable prefix: "+tt.envPrefix+"_")

			// Verify old prefix is NOT present when using custom prefix (that doesn't contain APP)
			if tt.envPrefix != "APP" && !strings.Contains(tt.envPrefix, "APP") {
				assert.NotContains(t, result, "APP_DATABASE_URL=")
				assert.NotContains(t, result, "APP_SERVER_PORT=")
			}
		})
	}
}
