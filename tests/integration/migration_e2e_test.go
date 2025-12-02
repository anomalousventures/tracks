package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMigrationTest(t *testing.T, driver string) (projectRoot string, cleanup func()) {
	t.Helper()

	tmpDir := t.TempDir()
	projectName := "testapp"

	cfg := generator.ProjectConfig{
		ProjectName:    projectName,
		ModulePath:     fmt.Sprintf("github.com/test/%s-app", driver),
		DatabaseDriver: driver,
		EnvPrefix:      "APP",
		InitGit:        false,
		OutputPath:     tmpDir,
	}

	gen := generator.NewProjectGenerator()
	ctx := context.Background()
	err := gen.Generate(ctx, cfg)
	require.NoError(t, err, "project generation should succeed")

	projectRoot = filepath.Join(tmpDir, projectName)

	tidyCmd, cancel := cmdWithTimeout(longTimeout, "go", "mod", "tidy")
	tidyCmd.Dir = projectRoot
	defer cancel()
	output, err := tidyCmd.CombinedOutput()
	if err != nil {
		t.Logf("go mod tidy output:\n%s", string(output))
	}
	require.NoError(t, err, "go mod tidy should succeed")

	cleanup = func() {}

	switch driver {
	case "sqlite3":
		dataDir := filepath.Join(projectRoot, "data")
		err = os.MkdirAll(dataDir, 0755)
		require.NoError(t, err, "should create data directory")

	case "postgres", "go-libsql":
		composeUpCmd, cancel := dockerCmdWithTimeout(longTimeout, "compose", "up", "-d")
		composeUpCmd.Dir = projectRoot
		defer cancel()
		output, err = composeUpCmd.CombinedOutput()
		if err != nil {
			t.Logf("docker compose up output:\n%s", string(output))
		}
		require.NoError(t, err, "docker compose up should succeed")

		t.Log("Waiting for database to be ready...")
		time.Sleep(5 * time.Second)

		cleanup = func() {
			t.Log("Stopping docker compose services...")
			composeDownCmd, cancel := dockerCmdWithTimeout(mediumTimeout, "compose", "down", "-v")
			defer cancel()
			composeDownCmd.Dir = projectRoot
			_ = composeDownCmd.Run()
		}
	}

	return projectRoot, cleanup
}

func getDBURL(driver, projectRoot string) string {
	switch driver {
	case "sqlite3":
		return fmt.Sprintf("file:%s/data/test.db", projectRoot)
	case "postgres":
		return "postgres://testapp:testapp@localhost:5432/testapp?sslmode=disable"
	case "go-libsql":
		return "http://localhost:8081"
	default:
		return ""
	}
}

func TestMigrationE2E_UpAppliesMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tests := []struct {
		name   string
		driver string
	}{
		{"sqlite3", "sqlite3"},
		{"postgres", "postgres"},
		{"go-libsql", "go-libsql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot, cleanup := setupMigrationTest(t, tt.driver)
			defer cleanup()

			dbURL := getDBURL(tt.driver, projectRoot)

			migrateCmd, cancel := cmdWithTimeout(e2eTimeout, "go", "run", "./cmd/migrate", "up")
			migrateCmd.Dir = projectRoot
			migrateCmd.Env = append(os.Environ(), "APP_DATABASE_URL="+dbURL)
			defer cancel()

			output, err := migrateCmd.CombinedOutput()
			outputStr := string(output)
			if err != nil {
				t.Logf("migrate up output:\n%s", outputStr)
			}
			require.NoError(t, err, "migrate up should succeed")

			assert.Contains(t, outputStr, "Applied", "should show applied migrations")
			assert.Contains(t, outputStr, "migration", "should mention migration")
		})
	}
}

func TestMigrationE2E_UpIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tests := []struct {
		name   string
		driver string
	}{
		{"sqlite3", "sqlite3"},
		{"postgres", "postgres"},
		{"go-libsql", "go-libsql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot, cleanup := setupMigrationTest(t, tt.driver)
			defer cleanup()

			dbURL := getDBURL(tt.driver, projectRoot)

			migrateCmd1, cancel1 := cmdWithTimeout(e2eTimeout, "go", "run", "./cmd/migrate", "up")
			migrateCmd1.Dir = projectRoot
			migrateCmd1.Env = append(os.Environ(), "APP_DATABASE_URL="+dbURL)
			defer cancel1()

			output1, err := migrateCmd1.CombinedOutput()
			if err != nil {
				t.Logf("first migrate up output:\n%s", string(output1))
			}
			require.NoError(t, err, "first migrate up should succeed")

			migrateCmd2, cancel2 := cmdWithTimeout(e2eTimeout, "go", "run", "./cmd/migrate", "up")
			migrateCmd2.Dir = projectRoot
			migrateCmd2.Env = append(os.Environ(), "APP_DATABASE_URL="+dbURL)
			defer cancel2()

			output2, err := migrateCmd2.CombinedOutput()
			outputStr := string(output2)
			if err != nil {
				t.Logf("second migrate up output:\n%s", outputStr)
			}
			require.NoError(t, err, "second migrate up should succeed")

			assert.Contains(t, outputStr, "up to date", "second run should show database is up to date")
		})
	}
}

func TestMigrationE2E_DownRollsBack(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tests := []struct {
		name   string
		driver string
	}{
		{"sqlite3", "sqlite3"},
		{"postgres", "postgres"},
		{"go-libsql", "go-libsql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot, cleanup := setupMigrationTest(t, tt.driver)
			defer cleanup()

			dbURL := getDBURL(tt.driver, projectRoot)

			migrateUpCmd, cancelUp := cmdWithTimeout(e2eTimeout, "go", "run", "./cmd/migrate", "up")
			migrateUpCmd.Dir = projectRoot
			migrateUpCmd.Env = append(os.Environ(), "APP_DATABASE_URL="+dbURL)
			defer cancelUp()

			output, err := migrateUpCmd.CombinedOutput()
			if err != nil {
				t.Logf("migrate up output:\n%s", string(output))
			}
			require.NoError(t, err, "migrate up should succeed")

			migrateDownCmd, cancelDown := cmdWithTimeout(e2eTimeout, "go", "run", "./cmd/migrate", "down")
			migrateDownCmd.Dir = projectRoot
			migrateDownCmd.Env = append(os.Environ(), "APP_DATABASE_URL="+dbURL)
			defer cancelDown()

			output, err = migrateDownCmd.CombinedOutput()
			outputStr := string(output)
			if err != nil {
				t.Logf("migrate down output:\n%s", outputStr)
			}
			require.NoError(t, err, "migrate down should succeed")

			assert.Contains(t, outputStr, "Rolled back", "should show rolled back migration")
		})
	}
}

func TestMigrationE2E_Status(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	tests := []struct {
		name   string
		driver string
	}{
		{"sqlite3", "sqlite3"},
		{"postgres", "postgres"},
		{"go-libsql", "go-libsql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot, cleanup := setupMigrationTest(t, tt.driver)
			defer cleanup()

			dbURL := getDBURL(tt.driver, projectRoot)

			statusCmd1, cancel1 := cmdWithTimeout(e2eTimeout, "go", "run", "./cmd/migrate", "status")
			statusCmd1.Dir = projectRoot
			statusCmd1.Env = append(os.Environ(), "APP_DATABASE_URL="+dbURL)
			defer cancel1()

			output1, err := statusCmd1.CombinedOutput()
			outputStr1 := string(output1)
			if err != nil {
				t.Logf("first status output:\n%s", outputStr1)
			}
			require.NoError(t, err, "first status should succeed")

			assert.Contains(t, outputStr1, "pending", "should show pending migrations before applying")

			migrateUpCmd, cancelUp := cmdWithTimeout(e2eTimeout, "go", "run", "./cmd/migrate", "up")
			migrateUpCmd.Dir = projectRoot
			migrateUpCmd.Env = append(os.Environ(), "APP_DATABASE_URL="+dbURL)
			defer cancelUp()

			output, err := migrateUpCmd.CombinedOutput()
			if err != nil {
				t.Logf("migrate up output:\n%s", string(output))
			}
			require.NoError(t, err, "migrate up should succeed")

			statusCmd2, cancel2 := cmdWithTimeout(e2eTimeout, "go", "run", "./cmd/migrate", "status")
			statusCmd2.Dir = projectRoot
			statusCmd2.Env = append(os.Environ(), "APP_DATABASE_URL="+dbURL)
			defer cancel2()

			output2, err := statusCmd2.CombinedOutput()
			outputStr2 := string(output2)
			if err != nil {
				t.Logf("second status output:\n%s", outputStr2)
			}
			require.NoError(t, err, "second status should succeed")

			assert.Contains(t, outputStr2, "applied", "should show applied migrations after applying")
			assert.NotContains(t, outputStr2, "pending", "should not show pending after applying")
		})
	}
}

func TestMigrationE2E_MakeTargets(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	projectRoot, cleanup := setupMigrationTest(t, "sqlite3")
	defer cleanup()

	dbURL := getDBURL("sqlite3", projectRoot)

	makeUpCmd, cancel := cmdWithTimeout(e2eTimeout, "make", "migrate-up")
	makeUpCmd.Dir = projectRoot
	makeUpCmd.Env = append(os.Environ(), "APP_DATABASE_URL="+dbURL)
	defer cancel()

	output, err := makeUpCmd.CombinedOutput()
	outputStr := string(output)
	if err != nil {
		t.Logf("make migrate-up output:\n%s", outputStr)
	}
	require.NoError(t, err, "make migrate-up should succeed")

	assert.Contains(t, outputStr, "Applied", "should show applied migrations")
}
