package template

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigTemplate(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name      string
		envPrefix string
	}{
		{"default prefix", "APP"},
		{"custom prefix", "MYAPP"},
		{"service prefix", "USERSERVICE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				EnvPrefix: tt.envPrefix,
			}

			result, err := renderer.Render("internal/config/config.go.tmpl", data)
			require.NoError(t, err)
			assert.NotEmpty(t, result)

			assert.Contains(t, result, `v.SetEnvPrefix("`+tt.envPrefix+`")`)
		})
	}
}

func TestConfigValidGoCode(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		EnvPrefix: "APP",
	}

	result, err := renderer.Render("internal/config/config.go.tmpl", data)
	require.NoError(t, err)

	fset := token.NewFileSet()
	_, err = parser.ParseFile(fset, "config.go", result, parser.AllErrors)
	require.NoError(t, err, "generated config.go should be valid Go code")
}

func TestConfigImports(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		EnvPrefix: "APP",
	}

	result, err := renderer.Render("internal/config/config.go.tmpl", data)
	require.NoError(t, err)

	requiredImports := []string{
		`"fmt"`,
		`"os"`,
		`"time"`,
		`"github.com/spf13/viper"`,
	}

	for _, imp := range requiredImports {
		assert.Contains(t, result, imp, "should import %s", imp)
	}
}

func TestConfigStructFields(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		EnvPrefix: "APP",
	}

	result, err := renderer.Render("internal/config/config.go.tmpl", data)
	require.NoError(t, err)

	requiredStructs := []string{
		"type Config struct",
		"type ServerConfig struct",
		"type DatabaseConfig struct",
		"type LoggingConfig struct",
	}

	for _, structDef := range requiredStructs {
		assert.Contains(t, result, structDef, "should define %s", structDef)
	}

	configFields := []string{
		"Environment string",
		"Server      ServerConfig",
		"Database    DatabaseConfig",
		"Logging     LoggingConfig",
	}

	for _, field := range configFields {
		assert.Contains(t, result, field, "Config should have field: %s", field)
	}
}

func TestConfigServerFields(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		EnvPrefix: "APP",
	}

	result, err := renderer.Render("internal/config/config.go.tmpl", data)
	require.NoError(t, err)

	serverFields := []string{
		"Port            string",
		"ReadTimeout     time.Duration",
		"WriteTimeout    time.Duration",
		"IdleTimeout     time.Duration",
		"ShutdownTimeout time.Duration",
	}

	for _, field := range serverFields {
		assert.Contains(t, result, field, "ServerConfig should have field: %s", field)
	}
}

func TestConfigDatabaseFields(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		EnvPrefix: "APP",
	}

	result, err := renderer.Render("internal/config/config.go.tmpl", data)
	require.NoError(t, err)

	databaseFields := []string{
		"URL             string",
		"ConnectTimeout  time.Duration",
		"MaxOpenConns    int",
		"MaxIdleConns    int",
		"ConnMaxLifetime time.Duration",
	}

	for _, field := range databaseFields {
		assert.Contains(t, result, field, "DatabaseConfig should have field: %s", field)
	}
}

func TestConfigLoggingFields(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		EnvPrefix: "APP",
	}

	result, err := renderer.Render("internal/config/config.go.tmpl", data)
	require.NoError(t, err)

	loggingFields := []string{
		"Level  string",
		"Format string",
	}

	for _, field := range loggingFields {
		assert.Contains(t, result, field, "LoggingConfig should have field: %s", field)
	}
}

func TestConfigLoadFunction(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		EnvPrefix: "APP",
	}

	result, err := renderer.Render("internal/config/config.go.tmpl", data)
	require.NoError(t, err)

	assert.Contains(t, result, "func Load() (*Config, error)")
	assert.Contains(t, result, "v := viper.New()")
	assert.Contains(t, result, `v.SetConfigFile(".env")`)
	assert.Contains(t, result, "v.ReadInConfig()")
	assert.Contains(t, result, "v.AutomaticEnv()")
	assert.Contains(t, result, "v.Unmarshal(&cfg)")
}

func TestConfigDefaults(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		EnvPrefix: "APP",
	}

	result, err := renderer.Render("internal/config/config.go.tmpl", data)
	require.NoError(t, err)

	expectedDefaults := []string{
		`v.SetDefault("environment", "production")`,
		`v.SetDefault("server.port", ":8080")`,
		`v.SetDefault("server.read_timeout", "15s")`,
		`v.SetDefault("server.write_timeout", "15s")`,
		`v.SetDefault("server.idle_timeout", "60s")`,
		`v.SetDefault("server.shutdown_timeout", "30s")`,
		`v.SetDefault("database.connect_timeout", "10s")`,
		`v.SetDefault("database.max_open_conns", 25)`,
		`v.SetDefault("database.max_idle_conns", 5)`,
		`v.SetDefault("database.conn_max_lifetime", "5m")`,
		`v.SetDefault("logging.level", "info")`,
		`v.SetDefault("logging.format", "json")`,
	}

	for _, defaultVal := range expectedDefaults {
		assert.Contains(t, result, defaultVal, "should set default: %s", defaultVal)
	}
}

func TestConfigEnvPrefixInterpolation(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	tests := []struct {
		name      string
		envPrefix string
	}{
		{"APP prefix", "APP"},
		{"MYAPP prefix", "MYAPP"},
		{"USERSERVICE prefix", "USERSERVICE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := TemplateData{
				EnvPrefix: tt.envPrefix,
			}

			result, err := renderer.Render("internal/config/config.go.tmpl", data)
			require.NoError(t, err)

			assert.Contains(t, result, `v.SetEnvPrefix("`+tt.envPrefix+`")`)

			if tt.envPrefix != "APP" {
				assert.NotContains(t, result, `v.SetEnvPrefix("APP")`)
			}
		})
	}
}

func TestConfigMapstructureTags(t *testing.T) {
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		EnvPrefix: "APP",
	}

	result, err := renderer.Render("internal/config/config.go.tmpl", data)
	require.NoError(t, err)

	requiredTags := []string{
		"`mapstructure:\"environment\"`",
		"`mapstructure:\"server\"`",
		"`mapstructure:\"database\"`",
		"`mapstructure:\"logging\"`",
		"`mapstructure:\"port\"`",
		"`mapstructure:\"read_timeout\"`",
		"`mapstructure:\"url\"`",
		"`mapstructure:\"level\"`",
		"`mapstructure:\"format\"`",
	}

	for _, tag := range requiredTags {
		assert.Contains(t, result, tag, "should have mapstructure tag: %s", tag)
	}
}
