package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

func CreateProjectDirectories(config ProjectConfig) error {
	projectRoot := filepath.Join(config.OutputPath, config.ProjectName)

	directories := []string{
		filepath.Join(projectRoot, ".github", "workflows"),
		filepath.Join(projectRoot, "cmd", "server"),
		filepath.Join(projectRoot, "cmd", "migrate"),
		filepath.Join(projectRoot, "internal", "interfaces"),
		filepath.Join(projectRoot, "internal", "domain", "health"),
		filepath.Join(projectRoot, "internal", "http", "handlers"),
		filepath.Join(projectRoot, "internal", "http", "helpers"),
		filepath.Join(projectRoot, "internal", "http", "routes"),
		filepath.Join(projectRoot, "internal", "http", "views", "layouts"),
		filepath.Join(projectRoot, "internal", "http", "views", "pages"),
		filepath.Join(projectRoot, "internal", "http", "views", "components"),
		filepath.Join(projectRoot, "internal", "http", "views", "components", "ui"),
		filepath.Join(projectRoot, "internal", "http", "views", "components", "utils"),
		filepath.Join(projectRoot, "internal", "db", "migrations", "sqlite"),
		filepath.Join(projectRoot, "internal", "db", "migrations", "postgres"),
		filepath.Join(projectRoot, "internal", "db", "queries"),
		filepath.Join(projectRoot, "internal", "db", "generated"),
		filepath.Join(projectRoot, "tests", "mocks"),
		filepath.Join(projectRoot, "tests", "integration"),
		filepath.Join(projectRoot, "internal", "pkg", "identifier"),
		filepath.Join(projectRoot, "internal", "pkg", "slug"),
		filepath.Join(projectRoot, "internal", "assets", "web", "css"),
		filepath.Join(projectRoot, "internal", "assets", "web", "js"),
		filepath.Join(projectRoot, "internal", "assets", "web", "images"),
		filepath.Join(projectRoot, "internal", "assets", "dist", "css"),
		filepath.Join(projectRoot, "internal", "assets", "dist", "js"),
		filepath.Join(projectRoot, "internal", "assets", "dist", "images"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
