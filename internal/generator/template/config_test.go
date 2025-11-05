package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/anomalousventures/tracks/tests/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderConfigTemplate(t *testing.T, envPrefix string) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)
	data := TemplateData{EnvPrefix: envPrefix}
	result, err := renderer.Render("internal/config/config.go.tmpl", data)
	require.NoError(t, err)
	return result
}

func TestConfigTemplate(t *testing.T) {
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
			result := renderConfigTemplate(t, tt.envPrefix)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, `v.SetEnvPrefix("`+tt.envPrefix+`")`)
		})
	}
}

func TestConfigValidGoCode(t *testing.T) {
	result := renderConfigTemplate(t, "APP")
	helpers.AssertValidGoCode(t, result, "config.go")
}

func TestConfigImports(t *testing.T) {
	result := renderConfigTemplate(t, "APP")

	helpers.AssertContainsAll(t, result, []string{
		`"fmt"`,
		`"os"`,
		`"time"`,
		`"github.com/spf13/viper"`,
	})
}

func TestConfigStructDefinitions(t *testing.T) {
	result := renderConfigTemplate(t, "APP")

	tests := []struct {
		name  string
		items []string
	}{
		{"struct declarations", []string{
			"type Config struct",
			"type ServerConfig struct",
			"type DatabaseConfig struct",
			"type LoggingConfig struct",
		}},
		{"config fields", []string{
			"Environment string",
			"Server      ServerConfig",
			"Database    DatabaseConfig",
			"Logging     LoggingConfig",
		}},
		{"server fields", []string{
			"Port            string",
			"ReadTimeout     time.Duration",
			"WriteTimeout    time.Duration",
			"IdleTimeout     time.Duration",
			"ShutdownTimeout time.Duration",
		}},
		{"database fields", []string{
			"URL             string",
			"ConnectTimeout  time.Duration",
			"MaxOpenConns    int",
			"MaxIdleConns    int",
			"ConnMaxLifetime time.Duration",
		}},
		{"logging fields", []string{
			"Level  string",
			"Format string",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.AssertContainsAll(t, result, tt.items)
		})
	}
}

func TestConfigLoadFunction(t *testing.T) {
	result := renderConfigTemplate(t, "APP")

	helpers.AssertContainsAll(t, result, []string{
		"func Load() (*Config, error)",
		"v := viper.New()",
		`v.SetConfigFile(".env")`,
		"v.ReadInConfig()",
		"v.AutomaticEnv()",
		"v.Unmarshal(&cfg)",
	})
}

func TestConfigDefaults(t *testing.T) {
	result := renderConfigTemplate(t, "APP")

	helpers.AssertContainsAll(t, result, []string{
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
	})
}

func TestConfigEnvPrefixInterpolation(t *testing.T) {
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
			result := renderConfigTemplate(t, tt.envPrefix)

			assert.Contains(t, result, `v.SetEnvPrefix("`+tt.envPrefix+`")`)

			if tt.envPrefix != "APP" {
				assert.NotContains(t, result, `v.SetEnvPrefix("APP")`)
			}
		})
	}
}

func TestConfigMapstructureTags(t *testing.T) {
	result := renderConfigTemplate(t, "APP")

	helpers.AssertContainsAll(t, result, []string{
		"`mapstructure:\"environment\"`",
		"`mapstructure:\"server\"`",
		"`mapstructure:\"database\"`",
		"`mapstructure:\"logging\"`",
		"`mapstructure:\"port\"`",
		"`mapstructure:\"read_timeout\"`",
		"`mapstructure:\"url\"`",
		"`mapstructure:\"level\"`",
		"`mapstructure:\"format\"`",
	})
}
