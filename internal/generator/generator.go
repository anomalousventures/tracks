package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	generatorinterfaces "github.com/anomalousventures/tracks/internal/generator/interfaces"
	"github.com/anomalousventures/tracks/internal/generator/template"
	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/rs/zerolog"
)

// ProjectConfig and other generator types remain in this package.
// The ProjectGenerator interface has moved to internal/cli/interfaces/generator.go
// following ADR-002 (interfaces defined by consumers).

type projectGenerator struct {
	renderer generatorinterfaces.TemplateRenderer
}

// NewProjectGenerator returns a fully functional project generator that orchestrates
// directory creation, template rendering, and git initialization.
func NewProjectGenerator() interfaces.ProjectGenerator {
	return &projectGenerator{
		renderer: template.NewRenderer(templates.FS),
	}
}

func (g *projectGenerator) Generate(ctx context.Context, cfg any) error {
	logger := zerolog.Ctx(ctx)

	projectCfg, ok := cfg.(ProjectConfig)
	if !ok {
		logger.Error().
			Str("type", fmt.Sprintf("%T", cfg)).
			Msg("invalid config type")
		return fmt.Errorf("invalid config type: expected ProjectConfig, got %T", cfg)
	}

	projectRoot := filepath.Join(projectCfg.OutputPath, projectCfg.ProjectName)

	logger.Info().
		Str("project", projectCfg.ProjectName).
		Str("path", projectRoot).
		Msg("starting project generation")

	data := template.TemplateData{
		ModuleName:  projectCfg.ModulePath,
		ProjectName: projectCfg.ProjectName,
		DBDriver:    projectCfg.DatabaseDriver,
		GoVersion:   "1.25",
		Year:        time.Now().Year(),
		EnvPrefix:   projectCfg.EnvPrefix,
	}

	logger.Info().
		Str("project", projectCfg.ProjectName).
		Str("path", projectRoot).
		Msg("creating project directories")

	if err := CreateProjectDirectories(projectCfg); err != nil {
		logger.Error().
			Err(err).
			Str("path", projectRoot).
			Msg("failed to create directories")
		return fmt.Errorf("failed to create project directories: %w", err)
	}

	templates := map[string]string{
		".env.example.tmpl":                          ".env.example",
		".gitignore.tmpl":                            ".gitignore",
		".golangci.yml.tmpl":                         ".golangci.yml",
		".mockery.yaml.tmpl":                         ".mockery.yaml",
		".tracks.yaml.tmpl":                          ".tracks.yaml",
		"go.mod.tmpl":                                "go.mod",
		"Makefile.tmpl":                              "Makefile",
		"README.md.tmpl":                             "README.md",
		"sqlc.yaml.tmpl":                             "sqlc.yaml",
		"cmd/server/main.go.tmpl":                    "cmd/server/main.go",
		"internal/config/config.go.tmpl":             "internal/config/config.go",
		"internal/interfaces/health.go.tmpl":         "internal/interfaces/health.go",
		"internal/interfaces/logger.go.tmpl":         "internal/interfaces/logger.go",
		"internal/logging/logger.go.tmpl":            "internal/logging/logger.go",
		"internal/domain/health/service.go.tmpl":     "internal/domain/health/service.go",
		"internal/http/server.go.tmpl":               "internal/http/server.go",
		"internal/http/routes.go.tmpl":               "internal/http/routes.go",
		"internal/http/routes/routes.go.tmpl":        "internal/http/routes/routes.go",
		"internal/http/handlers/health.go.tmpl":      "internal/http/handlers/health.go",
		"internal/http/middleware/logging.go.tmpl":   "internal/http/middleware/logging.go",
		"db/db.go.tmpl":                              "db/db.go",
	}

	logger.Info().
		Int("template_count", len(templates)).
		Msg("rendering templates")

	for templateName, outputFile := range templates {
		outputPath := filepath.Join(projectRoot, outputFile)

		logger.Debug().
			Str("template", templateName).
			Str("output", outputPath).
			Msg("rendering template")

		if err := g.renderer.RenderToFile(templateName, data, outputPath); err != nil {
			logger.Error().
				Err(err).
				Str("template", templateName).
				Str("output", outputPath).
				Msg("template rendering failed")
			return fmt.Errorf("failed to render %s: %w", templateName, err)
		}
	}

	logger.Info().Msg("all templates rendered successfully")

	if projectCfg.InitGit {
		logger.Info().
			Str("path", projectRoot).
			Msg("initializing git repository")

		if err := InitializeGit(ctx, projectRoot, false); err != nil {
			logger.Warn().
				Err(err).
				Str("path", projectRoot).
				Msg("git initialization failed - continuing without git")
		} else {
			logger.Info().Msg("git repository initialized")
		}
	} else {
		logger.Debug().Msg("skipping git initialization (--no-git)")
	}

	logger.Info().
		Str("project", projectCfg.ProjectName).
		Str("path", projectRoot).
		Msg("project generation complete")

	return nil
}

func (g *projectGenerator) Validate(cfg any) error {
	projectCfg, ok := cfg.(ProjectConfig)
	if !ok {
		return fmt.Errorf("invalid config type: expected ProjectConfig, got %T", cfg)
	}

	projectRoot := filepath.Join(projectCfg.OutputPath, projectCfg.ProjectName)

	if _, err := os.Stat(projectRoot); err == nil {
		return fmt.Errorf("directory '%s' already exists", projectRoot)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check directory: %w", err)
	}

	return nil
}
