package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/anomalousventures/tracks/internal/validation"
	"github.com/anomalousventures/tracks/tests/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func setupTestCommand(t *testing.T) *cobra.Command {
	mockValidator := mocks.NewMockValidator(t)
	mockGenerator := mocks.NewMockProjectGenerator(t)
	mockRenderer := mocks.NewMockRenderer(t)
	mockRenderer.On("Title", mock.Anything).Return().Maybe()
	mockRenderer.On("Section", mock.Anything).Return().Maybe()
	mockRenderer.On("Flush").Return(nil).Maybe()

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {
		mockRenderer.Flush()
	}
	cmd := NewNewCommand(mockValidator, mockGenerator, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))
	return cobraCmd
}

func setupTestCommandWithMock(t *testing.T) (*cobra.Command, *mocks.MockRenderer) {
	mockValidator := mocks.NewMockValidator(t)
	mockGenerator := mocks.NewMockProjectGenerator(t)
	mockRenderer := mocks.NewMockRenderer(t)

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}
	cmd := NewNewCommand(mockValidator, mockGenerator, factory, flusher)
	cobraCmd := cmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))
	return cobraCmd, mockRenderer
}

func TestNewNewCommand(t *testing.T) {
	mockRenderer := mocks.NewMockRenderer(t)
	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	mockValidator := mocks.NewMockValidator(t)
	mockGenerator := mocks.NewMockProjectGenerator(t)
	cmd := NewNewCommand(mockValidator, mockGenerator, rendererFactory, flusher)

	if cmd == nil {
		t.Fatal("NewNewCommand returned nil")
	}

	// Verify the command is properly configured by checking its cobra.Command
	cobraCmd := cmd.Command()
	if cobraCmd == nil {
		t.Fatal("Command() returned nil")
	}
	if cobraCmd.Use != "new [project-name]" {
		t.Errorf("Expected Use='new [project-name]', got Use='%s'", cobraCmd.Use)
	}
}

func TestNewCommand_Command(t *testing.T) {
	mockValidator := mocks.NewMockValidator(t)
	mockGenerator := mocks.NewMockProjectGenerator(t)
	mockRenderer := mocks.NewMockRenderer(t)
	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	newCmd := NewNewCommand(mockValidator, mockGenerator, rendererFactory, flusher)
	cobraCmd := newCmd.Command()

	if cobraCmd == nil {
		t.Fatal("Command() returned nil")
	}

	if cobraCmd.Use != "new [project-name]" {
		t.Errorf("expected Use 'new [project-name]', got %q", cobraCmd.Use)
	}

	if cobraCmd.Short == "" {
		t.Error("Short description is empty")
	}

	if cobraCmd.Long == "" {
		t.Error("Long description is empty")
	}

	if cobraCmd.Example == "" {
		t.Error("Example is empty")
	}

	if cobraCmd.RunE == nil {
		t.Error("RunE is nil, expected function")
	}
}

func TestNewCommand_CommandUsage(t *testing.T) {
	cobraCmd := setupTestCommand(t)

	cobraCmd.SetArgs([]string{})
	if err := cobraCmd.Execute(); err == nil {
		t.Error("expected error when no arguments provided, got nil")
	}

	cobraCmd.SetArgs([]string{"project1", "project2"})
	if err := cobraCmd.Execute(); err == nil {
		t.Error("expected error when too many arguments provided, got nil")
	}
}

func TestNewCommand_Run(t *testing.T) {
	mockValidator := mocks.NewMockValidator(t)
	mockGenerator := mocks.NewMockProjectGenerator(t)
	mockRenderer := mocks.NewMockRenderer(t)

	mockValidator.On("ValidateProjectName", mock.Anything, "myapp").Return(nil).Once()
	mockValidator.On("ValidateDatabaseDriver", mock.Anything, "go-libsql").Return(nil).Once()
	mockRenderer.On("Title", "Creating new Tracks application: myapp").Once()
	mockRenderer.On("Section", mock.Anything).Once()

	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}

	flusherCalled := false
	flusher := func(cmd *cobra.Command, r interfaces.Renderer) {
		flusherCalled = true
		if r != mockRenderer {
			t.Error("flusher called with different renderer")
		}
	}

	newCmd := NewNewCommand(mockValidator, mockGenerator, rendererFactory, flusher)
	cobraCmd := newCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"myapp"})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if !flusherCalled {
		t.Error("flusher was not called")
	}

	mockValidator.AssertExpectations(t)
}

func TestNewCommand_RunWithDifferentProjectNames(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		wantTitle   string
	}{
		{
			name:        "simple name",
			projectName: "myapp",
			wantTitle:   "Creating new Tracks application: myapp",
		},
		{
			name:        "name with hyphens",
			projectName: "my-awesome-app",
			wantTitle:   "Creating new Tracks application: my-awesome-app",
		},
		{
			name:        "single char name",
			projectName: "x",
			wantTitle:   "Creating new Tracks application: x",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := mocks.NewMockValidator(t)
			mockGenerator := mocks.NewMockProjectGenerator(t)
			mockRenderer := mocks.NewMockRenderer(t)

			mockValidator.On("ValidateProjectName", mock.Anything, tt.projectName).Return(nil).Once()
			mockValidator.On("ValidateDatabaseDriver", mock.Anything, "go-libsql").Return(nil).Once()
			mockRenderer.On("Title", tt.wantTitle).Once()
			mockRenderer.On("Section", mock.Anything).Once()

			factory := func(*cobra.Command) interfaces.Renderer {
				return mockRenderer
			}
			flusher := func(*cobra.Command, interfaces.Renderer) {}

			cmd := NewNewCommand(mockValidator, mockGenerator, factory, flusher)
			cobraCmd := cmd.Command()
			cobraCmd.SetOut(new(bytes.Buffer))
			cobraCmd.SetErr(new(bytes.Buffer))

			cobraCmd.SetArgs([]string{tt.projectName})
			if err := cobraCmd.Execute(); err != nil {
				t.Fatalf("execution failed: %v", err)
			}

			mockValidator.AssertExpectations(t)
		})
	}
}

func TestNewCommand_RendererFactoryCalledWithCommand(t *testing.T) {
	mockValidator := mocks.NewMockValidator(t)
	mockGenerator := mocks.NewMockProjectGenerator(t)
	mockRenderer := mocks.NewMockRenderer(t)

	mockValidator.On("ValidateProjectName", mock.Anything, "testapp").Return(nil).Once()
	mockValidator.On("ValidateDatabaseDriver", mock.Anything, "go-libsql").Return(nil).Once()
	mockRenderer.On("Title", mock.Anything).Once()
	mockRenderer.On("Section", mock.Anything).Once()

	var capturedCmd *cobra.Command
	rendererFactory := func(cmd *cobra.Command) interfaces.Renderer {
		capturedCmd = cmd
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	newCmd := NewNewCommand(mockValidator, mockGenerator, rendererFactory, flusher)
	cobraCmd := newCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"testapp"})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if capturedCmd != cobraCmd {
		t.Error("renderer factory not called with correct command")
	}

	mockValidator.AssertExpectations(t)
}

func TestNewCommand_FlusherCalledWithCommandAndRenderer(t *testing.T) {
	mockValidator := mocks.NewMockValidator(t)
	mockGenerator := mocks.NewMockProjectGenerator(t)
	mockRenderer := mocks.NewMockRenderer(t)

	mockValidator.On("ValidateProjectName", mock.Anything, "testapp").Return(nil).Once()
	mockValidator.On("ValidateDatabaseDriver", mock.Anything, "go-libsql").Return(nil).Once()
	mockRenderer.On("Title", mock.Anything).Once()
	mockRenderer.On("Section", mock.Anything).Once()

	rendererFactory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}

	var capturedCmd *cobra.Command
	var capturedRenderer interfaces.Renderer
	flusher := func(cmd *cobra.Command, r interfaces.Renderer) {
		capturedCmd = cmd
		capturedRenderer = r
	}

	newCmd := NewNewCommand(mockValidator, mockGenerator, rendererFactory, flusher)
	cobraCmd := newCmd.Command()
	cobraCmd.SetOut(new(bytes.Buffer))
	cobraCmd.SetErr(new(bytes.Buffer))

	cobraCmd.SetArgs([]string{"testapp"})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if capturedCmd != cobraCmd {
		t.Error("flusher not called with correct command")
	}

	if capturedRenderer != mockRenderer {
		t.Error("flusher not called with correct renderer")
	}

	mockValidator.AssertExpectations(t)
}

func TestNewCommand_CommandDescriptions(t *testing.T) {
	cobraCmd := setupTestCommand(t)

	keyTechnologies := []string{"templ", "SQLC", "production-ready", "Go"}
	for _, tech := range keyTechnologies {
		if !strings.Contains(cobraCmd.Long, tech) {
			t.Errorf("Long description missing mention of %q", tech)
		}
	}

	if !strings.Contains(cobraCmd.Example, "tracks new myapp") {
		t.Error("Example missing basic usage pattern")
	}
}

func TestNewCommand_FlagDefaults(t *testing.T) {
	mockValidator := mocks.NewMockValidator(t)
	mockGenerator := mocks.NewMockProjectGenerator(t)
	mockRenderer := mocks.NewMockRenderer(t)

	factory := func(*cobra.Command) interfaces.Renderer {
		return mockRenderer
	}
	flusher := func(*cobra.Command, interfaces.Renderer) {}

	cmd := NewNewCommand(mockValidator, mockGenerator, factory, flusher)
	cobraCmd := cmd.Command()

	dbFlag := cobraCmd.Flags().Lookup("db")
	if dbFlag == nil {
		t.Fatal("--db flag not found")
	}
	if dbFlag.DefValue != "go-libsql" {
		t.Errorf("expected --db default 'go-libsql', got %q", dbFlag.DefValue)
	}

	moduleFlag := cobraCmd.Flags().Lookup("module")
	if moduleFlag == nil {
		t.Fatal("--module flag not found")
	}
	if moduleFlag.DefValue != "" {
		t.Errorf("expected --module default '', got %q", moduleFlag.DefValue)
	}

	noGitFlag := cobraCmd.Flags().Lookup("no-git")
	if noGitFlag == nil {
		t.Fatal("--no-git flag not found")
	}
	if noGitFlag.DefValue != "false" {
		t.Errorf("expected --no-git default 'false', got %q", noGitFlag.DefValue)
	}
}

func TestNewCommand_FlagParsing(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantDB         string
		wantModule     string
		wantNoGit      bool
		setupValidator func(*mocks.MockValidator)
	}{
		{
			name:       "default flags",
			args:       []string{"myapp"},
			wantDB:     "go-libsql",
			wantModule: "example.com/myapp",
			wantNoGit:  false,
			setupValidator: func(v *mocks.MockValidator) {
				v.On("ValidateProjectName", mock.Anything, "myapp").Return(nil).Once()
				v.On("ValidateDatabaseDriver", mock.Anything, "go-libsql").Return(nil).Once()
			},
		},
		{
			name:       "db flag postgres",
			args:       []string{"myapp", "--db", "postgres"},
			wantDB:     "postgres",
			wantModule: "example.com/myapp",
			wantNoGit:  false,
			setupValidator: func(v *mocks.MockValidator) {
				v.On("ValidateProjectName", mock.Anything, "myapp").Return(nil).Once()
				v.On("ValidateDatabaseDriver", mock.Anything, "postgres").Return(nil).Once()
			},
		},
		{
			name:       "db flag sqlite3",
			args:       []string{"myapp", "--db", "sqlite3"},
			wantDB:     "sqlite3",
			wantModule: "example.com/myapp",
			wantNoGit:  false,
			setupValidator: func(v *mocks.MockValidator) {
				v.On("ValidateProjectName", mock.Anything, "myapp").Return(nil).Once()
				v.On("ValidateDatabaseDriver", mock.Anything, "sqlite3").Return(nil).Once()
			},
		},
		{
			name:       "module flag provided",
			args:       []string{"myapp", "--module", "github.com/user/myapp"},
			wantDB:     "go-libsql",
			wantModule: "github.com/user/myapp",
			wantNoGit:  false,
			setupValidator: func(v *mocks.MockValidator) {
				v.On("ValidateProjectName", mock.Anything, "myapp").Return(nil).Once()
				v.On("ValidateDatabaseDriver", mock.Anything, "go-libsql").Return(nil).Once()
				v.On("ValidateModulePath", mock.Anything, "github.com/user/myapp").Return(nil).Once()
			},
		},
		{
			name:       "no-git flag true",
			args:       []string{"myapp", "--no-git"},
			wantDB:     "go-libsql",
			wantModule: "example.com/myapp",
			wantNoGit:  true,
			setupValidator: func(v *mocks.MockValidator) {
				v.On("ValidateProjectName", mock.Anything, "myapp").Return(nil).Once()
				v.On("ValidateDatabaseDriver", mock.Anything, "go-libsql").Return(nil).Once()
			},
		},
		{
			name:       "combined flags",
			args:       []string{"myapp", "--db", "postgres", "--module", "github.com/org/project", "--no-git"},
			wantDB:     "postgres",
			wantModule: "github.com/org/project",
			wantNoGit:  true,
			setupValidator: func(v *mocks.MockValidator) {
				v.On("ValidateProjectName", mock.Anything, "myapp").Return(nil).Once()
				v.On("ValidateDatabaseDriver", mock.Anything, "postgres").Return(nil).Once()
				v.On("ValidateModulePath", mock.Anything, "github.com/org/project").Return(nil).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := mocks.NewMockValidator(t)
			mockGenerator := mocks.NewMockProjectGenerator(t)
			mockRenderer := mocks.NewMockRenderer(t)

			tt.setupValidator(mockValidator)
			mockRenderer.On("Title", mock.Anything).Return().Once()
			mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
				return strings.Contains(s.Body, "Database: "+tt.wantDB) &&
					strings.Contains(s.Body, "Module: "+tt.wantModule)
			})).Return().Once()

			factory := func(*cobra.Command) interfaces.Renderer {
				return mockRenderer
			}
			flusher := func(*cobra.Command, interfaces.Renderer) {}

			cmd := NewNewCommand(mockValidator, mockGenerator, factory, flusher)
			cobraCmd := cmd.Command()
			cobraCmd.SetOut(new(bytes.Buffer))
			cobraCmd.SetErr(new(bytes.Buffer))

			cobraCmd.SetArgs(tt.args)
			if err := cobraCmd.Execute(); err != nil {
				t.Fatalf("execution failed: %v", err)
			}

			mockValidator.AssertExpectations(t)
			mockRenderer.AssertExpectations(t)
		})
	}
}

func TestNewCommand_ValidationIntegration(t *testing.T) {
	t.Run("project name validation error", func(t *testing.T) {
		mockValidator := mocks.NewMockValidator(t)
		mockGenerator := mocks.NewMockProjectGenerator(t)
		mockRenderer := mocks.NewMockRenderer(t)

		mockValidator.On("ValidateProjectName", mock.Anything, "MyApp").
			Return(&validation.ValidationError{Field: "project_name", Message: "must be lowercase"}).Once()

		factory := func(*cobra.Command) interfaces.Renderer {
			return mockRenderer
		}
		flusher := func(*cobra.Command, interfaces.Renderer) {}

		cmd := NewNewCommand(mockValidator, mockGenerator, factory, flusher)
		cobraCmd := cmd.Command()
		cobraCmd.SetOut(new(bytes.Buffer))
		cobraCmd.SetErr(new(bytes.Buffer))

		cobraCmd.SetArgs([]string{"MyApp"})
		err := cobraCmd.Execute()
		if err == nil {
			t.Fatal("expected error for invalid project name, got nil")
		}

		if !strings.Contains(err.Error(), "invalid project name") {
			t.Errorf("expected error message to contain 'invalid project name', got %q", err.Error())
		}

		mockValidator.AssertExpectations(t)
	})

	t.Run("database driver validation error", func(t *testing.T) {
		mockValidator := mocks.NewMockValidator(t)
		mockGenerator := mocks.NewMockProjectGenerator(t)
		mockRenderer := mocks.NewMockRenderer(t)

		mockValidator.On("ValidateProjectName", mock.Anything, "myapp").Return(nil).Once()
		mockValidator.On("ValidateDatabaseDriver", mock.Anything, "mysql").
			Return(&validation.ValidationError{Field: "database_driver", Message: "unsupported driver"}).Once()

		factory := func(*cobra.Command) interfaces.Renderer {
			return mockRenderer
		}
		flusher := func(*cobra.Command, interfaces.Renderer) {}

		cmd := NewNewCommand(mockValidator, mockGenerator, factory, flusher)
		cobraCmd := cmd.Command()
		cobraCmd.SetOut(new(bytes.Buffer))
		cobraCmd.SetErr(new(bytes.Buffer))

		cobraCmd.SetArgs([]string{"myapp", "--db", "mysql"})
		err := cobraCmd.Execute()
		if err == nil {
			t.Fatal("expected error for invalid database driver, got nil")
		}

		if !strings.Contains(err.Error(), "invalid database driver") {
			t.Errorf("expected error message to contain 'invalid database driver', got %q", err.Error())
		}

		mockValidator.AssertExpectations(t)
	})

	t.Run("module path validation error", func(t *testing.T) {
		mockValidator := mocks.NewMockValidator(t)
		mockGenerator := mocks.NewMockProjectGenerator(t)
		mockRenderer := mocks.NewMockRenderer(t)

		mockValidator.On("ValidateProjectName", mock.Anything, "myapp").Return(nil).Once()
		mockValidator.On("ValidateDatabaseDriver", mock.Anything, "go-libsql").Return(nil).Once()
		mockValidator.On("ValidateModulePath", mock.Anything, "invalid path").
			Return(&validation.ValidationError{Field: "module_path", Message: "invalid format"}).Once()

		factory := func(*cobra.Command) interfaces.Renderer {
			return mockRenderer
		}
		flusher := func(*cobra.Command, interfaces.Renderer) {}

		cmd := NewNewCommand(mockValidator, mockGenerator, factory, flusher)
		cobraCmd := cmd.Command()
		cobraCmd.SetOut(new(bytes.Buffer))
		cobraCmd.SetErr(new(bytes.Buffer))

		cobraCmd.SetArgs([]string{"myapp", "--module", "invalid path"})
		err := cobraCmd.Execute()
		if err == nil {
			t.Fatal("expected error for invalid module path, got nil")
		}

		if !strings.Contains(err.Error(), "invalid module path") {
			t.Errorf("expected error message to contain 'invalid module path', got %q", err.Error())
		}

		mockValidator.AssertExpectations(t)
	})

	t.Run("context propagation through validation", func(t *testing.T) {
		mockValidator := mocks.NewMockValidator(t)
		mockGenerator := mocks.NewMockProjectGenerator(t)
		mockRenderer := mocks.NewMockRenderer(t)
		mockRenderer.On("Title", mock.Anything).Return().Maybe()
		mockRenderer.On("Section", mock.Anything).Return().Maybe()

		mockValidator.On("ValidateProjectName", mock.MatchedBy(func(ctx interface{}) bool {
			return ctx != nil
		}), "myapp").Return(nil).Once()

		mockValidator.On("ValidateDatabaseDriver", mock.MatchedBy(func(ctx interface{}) bool {
			return ctx != nil
		}), "go-libsql").Return(nil).Once()

		factory := func(*cobra.Command) interfaces.Renderer {
			return mockRenderer
		}
		flusher := func(*cobra.Command, interfaces.Renderer) {}

		cmd := NewNewCommand(mockValidator, mockGenerator, factory, flusher)
		cobraCmd := cmd.Command()
		cobraCmd.SetOut(new(bytes.Buffer))
		cobraCmd.SetErr(new(bytes.Buffer))

		cobraCmd.SetArgs([]string{"myapp"})
		if err := cobraCmd.Execute(); err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		mockValidator.AssertExpectations(t)
	})

	t.Run("module auto-generation when not provided", func(t *testing.T) {
		mockValidator := mocks.NewMockValidator(t)
		mockGenerator := mocks.NewMockProjectGenerator(t)
		mockRenderer := mocks.NewMockRenderer(t)
		mockRenderer.On("Title", mock.Anything).Return().Maybe()
		mockRenderer.On("Section", mock.MatchedBy(func(s interfaces.Section) bool {
			return strings.Contains(s.Body, "Module: example.com/myapp")
		})).Return().Maybe()

		mockValidator.On("ValidateProjectName", mock.Anything, "myapp").Return(nil).Once()
		mockValidator.On("ValidateDatabaseDriver", mock.Anything, "go-libsql").Return(nil).Once()

		factory := func(*cobra.Command) interfaces.Renderer {
			return mockRenderer
		}
		flusher := func(*cobra.Command, interfaces.Renderer) {}

		cmd := NewNewCommand(mockValidator, mockGenerator, factory, flusher)
		cobraCmd := cmd.Command()
		cobraCmd.SetOut(new(bytes.Buffer))
		cobraCmd.SetErr(new(bytes.Buffer))

		cobraCmd.SetArgs([]string{"myapp"})
		if err := cobraCmd.Execute(); err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		mockValidator.AssertExpectations(t)
		mockRenderer.AssertExpectations(t)
	})

	t.Run("module validation called when provided", func(t *testing.T) {
		mockValidator := mocks.NewMockValidator(t)
		mockGenerator := mocks.NewMockProjectGenerator(t)
		mockRenderer := mocks.NewMockRenderer(t)
		mockRenderer.On("Title", mock.Anything).Return().Maybe()
		mockRenderer.On("Section", mock.Anything).Return().Maybe()

		mockValidator.On("ValidateProjectName", mock.Anything, "myapp").Return(nil).Once()
		mockValidator.On("ValidateDatabaseDriver", mock.Anything, "go-libsql").Return(nil).Once()
		mockValidator.On("ValidateModulePath", mock.Anything, "github.com/user/myapp").Return(nil).Once()

		factory := func(*cobra.Command) interfaces.Renderer {
			return mockRenderer
		}
		flusher := func(*cobra.Command, interfaces.Renderer) {}

		cmd := NewNewCommand(mockValidator, mockGenerator, factory, flusher)
		cobraCmd := cmd.Command()
		cobraCmd.SetOut(new(bytes.Buffer))
		cobraCmd.SetErr(new(bytes.Buffer))

		cobraCmd.SetArgs([]string{"myapp", "--module", "github.com/user/myapp"})
		if err := cobraCmd.Execute(); err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		mockValidator.AssertExpectations(t)
	})

	t.Run("module validation not called when auto-generated", func(t *testing.T) {
		mockValidator := mocks.NewMockValidator(t)
		mockGenerator := mocks.NewMockProjectGenerator(t)
		mockRenderer := mocks.NewMockRenderer(t)
		mockRenderer.On("Title", mock.Anything).Return().Maybe()
		mockRenderer.On("Section", mock.Anything).Return().Maybe()

		mockValidator.On("ValidateProjectName", mock.Anything, "myapp").Return(nil).Once()
		mockValidator.On("ValidateDatabaseDriver", mock.Anything, "go-libsql").Return(nil).Once()

		factory := func(*cobra.Command) interfaces.Renderer {
			return mockRenderer
		}
		flusher := func(*cobra.Command, interfaces.Renderer) {}

		cmd := NewNewCommand(mockValidator, mockGenerator, factory, flusher)
		cobraCmd := cmd.Command()
		cobraCmd.SetOut(new(bytes.Buffer))
		cobraCmd.SetErr(new(bytes.Buffer))

		cobraCmd.SetArgs([]string{"myapp"})
		if err := cobraCmd.Execute(); err != nil {
			t.Fatalf("execution failed: %v", err)
		}

		mockValidator.AssertExpectations(t)
	})
}
