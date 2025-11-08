package generator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderSuccessOutput(t *testing.T) {
	output := SuccessOutput{
		ProjectName:    "myapp",
		ProjectPath:    "/home/user/myapp",
		ModulePath:     "github.com/user/myapp",
		DatabaseDriver: "postgres",
		GitInitialized: true,
		NoColor:        false,
	}

	result := RenderSuccessOutput(output)

	assert.Contains(t, result, "myapp")
	assert.Contains(t, result, "/home/user/myapp")
	assert.Contains(t, result, "github.com/user/myapp")
	assert.Contains(t, result, "postgres")
	assert.Contains(t, result, "initialized")
	assert.Contains(t, result, "cd myapp")
	assert.Contains(t, result, "make test")
	assert.Contains(t, result, "make dev")
}

func TestRenderSuccessOutput_GitNotInitialized(t *testing.T) {
	output := SuccessOutput{
		ProjectName:    "myapp",
		ProjectPath:    "/home/user/myapp",
		ModulePath:     "github.com/user/myapp",
		DatabaseDriver: "go-libsql",
		GitInitialized: false,
		NoColor:        false,
	}

	result := RenderSuccessOutput(output)

	assert.Contains(t, result, "not initialized")
	assert.NotContains(t, result, "Git:           initialized")
}

func TestRenderSuccessOutput_NoColor(t *testing.T) {
	output := SuccessOutput{
		ProjectName:    "testapp",
		ProjectPath:    "/tmp/testapp",
		ModulePath:     "example.com/testapp",
		DatabaseDriver: "sqlite3",
		GitInitialized: true,
		NoColor:        true,
	}

	result := RenderSuccessOutput(output)

	assert.Contains(t, result, "testapp")
	assert.Contains(t, result, "/tmp/testapp")
	assert.Contains(t, result, "example.com/testapp")
	assert.Contains(t, result, "sqlite3")
	assert.Contains(t, result, "initialized")
	assert.Contains(t, result, "cd testapp")

	assert.NotContains(t, result, "\x1b[")
}

func TestRenderSuccessPlain(t *testing.T) {
	output := SuccessOutput{
		ProjectName:    "plainapp",
		ProjectPath:    "/home/plainapp",
		ModulePath:     "github.com/test/plainapp",
		DatabaseDriver: "postgres",
		GitInitialized: false,
		NoColor:        true,
	}

	result := renderSuccessPlain(output)

	assert.Contains(t, result, "✓ Project 'plainapp' created successfully!")
	assert.Contains(t, result, "Location:      /home/plainapp")
	assert.Contains(t, result, "Module:        github.com/test/plainapp")
	assert.Contains(t, result, "Database:      postgres")
	assert.Contains(t, result, "Git:           not initialized")
	assert.Contains(t, result, "Next steps:")
	assert.Contains(t, result, "1. cd plainapp")
	assert.Contains(t, result, "2. make test")
	assert.Contains(t, result, "3. make dev")
}

func TestRenderSuccessOutput_AllDatabaseDrivers(t *testing.T) {
	drivers := []string{"go-libsql", "sqlite3", "postgres"}

	for _, driver := range drivers {
		t.Run(driver, func(t *testing.T) {
			output := SuccessOutput{
				ProjectName:    "testapp",
				ProjectPath:    "/tmp/testapp",
				ModulePath:     "example.com/testapp",
				DatabaseDriver: driver,
				GitInitialized: true,
				NoColor:        false,
			}

			result := RenderSuccessOutput(output)
			assert.Contains(t, result, driver)
		})
	}
}

func TestGetAbsolutePath(t *testing.T) {
	tests := []struct {
		name        string
		basePath    string
		projectName string
	}{
		{
			name:        "current directory",
			basePath:    ".",
			projectName: "myapp",
		},
		{
			name:        "explicit path",
			basePath:    "/tmp",
			projectName: "testapp",
		},
		{
			name:        "relative path",
			basePath:    "../projects",
			projectName: "demo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetAbsolutePath(tt.basePath, tt.projectName)
			assert.NoError(t, err)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, tt.projectName)
			assert.True(t, strings.HasPrefix(result, "/") || strings.Contains(result, ":\\"))
		})
	}
}

func TestRenderSuccessOutput_NextStepsOrder(t *testing.T) {
	output := SuccessOutput{
		ProjectName:    "myapp",
		ProjectPath:    "/home/user/myapp",
		ModulePath:     "github.com/user/myapp",
		DatabaseDriver: "postgres",
		GitInitialized: true,
		NoColor:        true,
	}

	result := RenderSuccessOutput(output)

	cdIndex := strings.Index(result, "cd myapp")
	testIndex := strings.Index(result, "make test")
	devIndex := strings.Index(result, "make dev")

	assert.True(t, cdIndex < testIndex, "cd should come before make test")
	assert.True(t, testIndex < devIndex, "make test should come before make dev")
}
