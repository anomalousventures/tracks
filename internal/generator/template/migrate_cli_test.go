package template

import (
	"testing"

	"github.com/anomalousventures/tracks/internal/templates"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderMigrateCLITemplate(t *testing.T) string {
	t.Helper()
	renderer := NewRenderer(templates.FS)

	data := TemplateData{
		ModuleName: "github.com/test/app",
	}

	result, err := renderer.Render("cmd/migrate/main.go.tmpl", data)
	require.NoError(t, err)
	return result
}

func TestMigrateCLITemplateRenders(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.NotEmpty(t, result, "template should render")
}

func TestMigrateCLIPackageMain(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.Contains(t, result, "package main", "should be main package")
}

func TestMigrateCLIImportsConfigAndDB(t *testing.T) {
	result := renderMigrateCLITemplate(t)

	assert.Contains(t, result, `"github.com/test/app/internal/config"`, "should import config package")
	assert.Contains(t, result, `"github.com/test/app/internal/db"`, "should import db package")
}

func TestMigrateCLIImportsStandardLibrary(t *testing.T) {
	result := renderMigrateCLITemplate(t)

	assert.Contains(t, result, `"context"`, "should import context")
	assert.Contains(t, result, `"database/sql"`, "should import database/sql")
	assert.Contains(t, result, `"fmt"`, "should import fmt")
	assert.Contains(t, result, `"os"`, "should import os")
}

func TestMigrateCLIHasMainFunction(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.Contains(t, result, "func main()", "should have main function")
}

func TestMigrateCLIHasRunFunction(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.Contains(t, result, "func run() error", "should have run function returning error")
}

func TestMigrateCLIHandlesUpCommand(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.Contains(t, result, `case "up":`, "should handle up command")
	assert.Contains(t, result, "migrateUp(ctx, database)", "should call migrateUp")
}

func TestMigrateCLIHandlesDownCommand(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.Contains(t, result, `case "down":`, "should handle down command")
	assert.Contains(t, result, "migrateDown(ctx, database)", "should call migrateDown")
}

func TestMigrateCLIHandlesStatusCommand(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.Contains(t, result, `case "status":`, "should handle status command")
	assert.Contains(t, result, "migrateStatus(ctx, database)", "should call migrateStatus")
}

func TestMigrateCLICallsDBMigrateFunctions(t *testing.T) {
	result := renderMigrateCLITemplate(t)

	assert.Contains(t, result, "db.MigrateUp(ctx, database)", "should call db.MigrateUp")
	assert.Contains(t, result, "db.MigrateDown(ctx, database)", "should call db.MigrateDown")
	assert.Contains(t, result, "db.MigrateStatus(ctx, database)", "should call db.MigrateStatus")
}

func TestMigrateCLILoadsConfig(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.Contains(t, result, "config.Load()", "should load config")
}

func TestMigrateCLIConnectsToDatabase(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.Contains(t, result, "db.New(ctx, cfg.Database)", "should connect to database")
}

func TestMigrateCLIClosesDatabase(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.Contains(t, result, "database.Close()", "should close database")
	assert.Contains(t, result, "failed to close database", "should handle close error")
}

func TestMigrateCLIHasPrintUsage(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.Contains(t, result, "func printUsage()", "should have printUsage function")
}

func TestMigrateCLIUsageShowsCommands(t *testing.T) {
	result := renderMigrateCLITemplate(t)

	assert.Contains(t, result, "up", "usage should mention up command")
	assert.Contains(t, result, "down", "usage should mention down command")
	assert.Contains(t, result, "status", "usage should mention status command")
}

func TestMigrateCLIHandlesMissingCommand(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.Contains(t, result, `"missing command"`, "should error on missing command")
}

func TestMigrateCLIHandlesUnknownCommand(t *testing.T) {
	result := renderMigrateCLITemplate(t)
	assert.Contains(t, result, `"unknown command:`, "should error on unknown command")
}

func TestMigrateCLIUsesSqlDB(t *testing.T) {
	result := renderMigrateCLITemplate(t)

	assert.Contains(t, result, "func migrateUp(ctx context.Context, database *sql.DB)", "migrateUp should accept *sql.DB")
	assert.Contains(t, result, "func migrateDown(ctx context.Context, database *sql.DB)", "migrateDown should accept *sql.DB")
	assert.Contains(t, result, "func migrateStatus(ctx context.Context, database *sql.DB)", "migrateStatus should accept *sql.DB")
}
