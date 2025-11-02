package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateProjectDirectories creates the complete directory structure for a new project.
// Uses ProjectConfig.OutputPath and ProjectConfig.ProjectName to determine the base path.
// All directories are created with 0755 permissions for cross-platform compatibility.
func CreateProjectDirectories(config ProjectConfig) error {
	projectRoot := filepath.Join(config.OutputPath, config.ProjectName)

	directories := []string{
		filepath.Join(projectRoot, "cmd", "server"),
		filepath.Join(projectRoot, "internal", "interfaces"),
		filepath.Join(projectRoot, "internal", "domain", "health"),
		filepath.Join(projectRoot, "internal", "http", "handlers"),
		filepath.Join(projectRoot, "internal", "routes"),
		filepath.Join(projectRoot, "db", "migrations"),
		filepath.Join(projectRoot, "db", "queries"),
		filepath.Join(projectRoot, "db", "generated"),
		filepath.Join(projectRoot, "test", "mocks"),
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
