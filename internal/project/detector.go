package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

const (
	tracksConfigFile   = ".tracks.yaml"
	templUIConfigFile  = ".templui.json"
)

type detector struct{}

// NewDetector creates a new ProjectDetector implementation.
func NewDetector() interfaces.ProjectDetector {
	return &detector{}
}

func (d *detector) Detect(ctx context.Context, startDir string) (*interfaces.TracksProject, string, error) {
	logger := zerolog.Ctx(ctx)

	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve path: %w", err)
	}

	dir := absDir
	for {
		configPath := filepath.Join(dir, tracksConfigFile)
		if _, err := os.Stat(configPath); err == nil {
			logger.Debug().
				Str("path", configPath).
				Msg("found .tracks.yaml")

			proj, err := d.loadConfig(configPath)
			if err != nil {
				return nil, "", err
			}
			return proj, dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil, "", ErrNotTracksProject
}

func (d *detector) HasTemplUIConfig(ctx context.Context, projectDir string) bool {
	configPath := filepath.Join(projectDir, templUIConfigFile)
	_, err := os.Stat(configPath)
	return err == nil
}

func (d *detector) loadConfig(configPath string) (*interfaces.TracksProject, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", tracksConfigFile, err)
	}

	var config tracksConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", tracksConfigFile, err)
	}

	return &interfaces.TracksProject{
		Name:       config.Project.Name,
		ModulePath: config.Project.ModulePath,
		DBDriver:   config.Project.DatabaseDriver,
	}, nil
}

// tracksConfig matches the .tracks.yaml file structure.
type tracksConfig struct {
	SchemaVersion string `yaml:"schema_version"`
	Project       struct {
		Name                string `yaml:"name"`
		ModulePath          string `yaml:"module_path"`
		TracksVersion       string `yaml:"tracks_version"`
		LastUpgradedVersion string `yaml:"last_upgraded_version"`
		DatabaseDriver      string `yaml:"database_driver"`
		EnvPrefix           string `yaml:"env_prefix"`
	} `yaml:"project"`
}
