package generator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	generatorinterfaces "github.com/anomalousventures/tracks/internal/generator/interfaces"
	"github.com/anomalousventures/tracks/internal/generator/template"
	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/rs/zerolog"
)

type projectGenerator struct {
	renderer generatorinterfaces.TemplateRenderer
}

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

	// Phase 1: Pre-generate templates (files that don't import SQLC-generated code)
	preGenerateTemplates := map[string]string{
		".env.example.tmpl":                        ".env.example",
		".gitignore.tmpl":                          ".gitignore",
		".golangci.yml.tmpl":                       ".golangci.yml",
		".mockery.yaml.tmpl":                       ".mockery.yaml",
		".tracks.yaml.tmpl":                        ".tracks.yaml",
		".air.toml.tmpl":                           ".air.toml",
		"go.mod.tmpl":                              "go.mod",
		"Makefile.tmpl":                            "Makefile",
		"README.md.tmpl":                           "README.md",
		"sqlc.yaml.tmpl":                           "sqlc.yaml",
		"docker-compose.yml.tmpl":                  "docker-compose.yml",
		"internal/config/config.go.tmpl":           "internal/config/config.go",
		"internal/interfaces/health.go.tmpl":       "internal/interfaces/health.go",
		"internal/interfaces/logger.go.tmpl":       "internal/interfaces/logger.go",
		"internal/logging/logger.go.tmpl":          "internal/logging/logger.go",
		"internal/domain/health/service.go.tmpl":   "internal/domain/health/service.go",
		"internal/http/server.go.tmpl":             "internal/http/server.go",
		"internal/http/routes.go.tmpl":             "internal/http/routes.go",
		"internal/http/routes/routes.go.tmpl":      "internal/http/routes/routes.go",
		"internal/http/handlers/health.go.tmpl":    "internal/http/handlers/health.go",
		"internal/http/middleware/logging.go.tmpl": "internal/http/middleware/logging.go",
		"internal/db/db.go.tmpl":                   "internal/db/db.go",
		"internal/db/queries/.gitkeep.tmpl":        "internal/db/queries/.gitkeep",
		"internal/db/queries/health.sql.tmpl":      "internal/db/queries/health.sql",
	}

	// Phase 3: Post-generate templates (files that import SQLC-generated code or depend on it)
	postGenerateTemplates := map[string]string{
		"cmd/server/main.go.tmpl":                      "cmd/server/main.go",
		"internal/domain/health/repository.go.tmpl":    "internal/domain/health/repository.go",
	}

	testTemplates := map[string]string{
		"internal/config/config_test.go.tmpl":         "internal/config/config_test.go",
		"internal/logging/logger_test.go.tmpl":        "internal/logging/logger_test.go",
		"internal/domain/health/service_test.go.tmpl": "internal/domain/health/service_test.go",
		"internal/http/handlers/health_test.go.tmpl":  "internal/http/handlers/health_test.go",
	}

	// Phase 1: Render pre-generate templates
	logger.Info().
		Int("template_count", len(preGenerateTemplates)).
		Msg("rendering pre-generate templates")

	for templateName, outputFile := range preGenerateTemplates {
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

	logger.Info().Msg("pre-generate templates rendered successfully")

	logger.Info().Msg("tidying dependencies")
	tidyCmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	tidyCmd.Dir = projectRoot
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		logger.Warn().
			Err(err).
			Str("output", string(output)).
			Msg("go mod tidy failed - continuing anyway")
	} else {
		logger.Info().Msg("dependencies tidied")
	}

	logger.Info().Msg("downloading all dependencies including tools")
	downloadCmd := exec.CommandContext(ctx, "go", "mod", "download", "all")
	downloadCmd.Dir = projectRoot
	if output, err := downloadCmd.CombinedOutput(); err != nil {
		logger.Warn().
			Err(err).
			Str("output", string(output)).
			Msg("go mod download all failed - continuing anyway")
	} else {
		logger.Info().Msg("all dependencies downloaded and go.sum populated")
	}

	logger.Info().Msg("generating mocks and SQL code")
	generateCmd := exec.CommandContext(ctx, "make", "generate")
	generateCmd.Dir = projectRoot
	if output, err := generateCmd.CombinedOutput(); err != nil {
		logger.Error().
			Err(err).
			Str("output", string(output)).
			Msg("make generate failed")
		return fmt.Errorf("failed to generate mocks and SQL code: %w", err)
	} else {
		logger.Info().Msg("mocks and SQL code generated successfully")
	}

	// Phase 3: Render post-generate templates (files that import generated code)
	logger.Info().
		Int("template_count", len(postGenerateTemplates)).
		Msg("rendering post-generate templates")

	for templateName, outputFile := range postGenerateTemplates {
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

	logger.Info().Msg("post-generate templates rendered successfully")

	// Phase 4: Tidy dependencies again (post-generate templates may have new imports)
	logger.Info().Msg("tidying dependencies after post-generate templates")
	tidyCmd = exec.CommandContext(ctx, "go", "mod", "tidy")
	tidyCmd.Dir = projectRoot
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		logger.Warn().
			Err(err).
			Str("output", string(output)).
			Msg("go mod tidy (after post-generate) failed - continuing anyway")
	} else {
		logger.Info().Msg("dependencies tidied")
	}

	// Phase 5: Render test templates
	logger.Info().
		Int("template_count", len(testTemplates)).
		Msg("rendering test templates")

	for templateName, outputFile := range testTemplates {
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

	logger.Info().Msg("test templates rendered successfully")

	logger.Info().Msg("running go mod tidy to pick up test dependencies")
	tidyCmd = exec.CommandContext(ctx, "go", "mod", "tidy")
	tidyCmd.Dir = projectRoot
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		logger.Warn().
			Err(err).
			Str("output", string(output)).
			Msg("go mod tidy (after test templates) failed - continuing anyway")
	}

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
