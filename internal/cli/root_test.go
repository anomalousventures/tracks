package cli

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func setupViperWithNoColor(t *testing.T, rootCmd *cobra.Command) *viper.Viper {
	t.Helper()
	v := viper.New()
	if err := v.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color")); err != nil {
		t.Fatalf("failed to bind no-color flag: %v", err)
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		v.SetDefault("no-color", true)
	}
	return v
}

func TestGlobalFlagsExist(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{"json flag exists", "json"},
		{"no-color flag exists", "no-color"},
		{"interactive flag exists", "interactive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
			rootCmd := NewRootCmd(build)
			flag := rootCmd.PersistentFlags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("flag --%s does not exist", tt.flagName)
			}
		})
	}
}

func TestGlobalFlagsArePersistent(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{"json flag is persistent", "json"},
		{"no-color flag is persistent", "no-color"},
		{"interactive flag is persistent", "interactive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
			rootCmd := NewRootCmd(build)
			persistentFlag := rootCmd.PersistentFlags().Lookup(tt.flagName)
			localFlag := rootCmd.Flags().Lookup(tt.flagName)

			if persistentFlag == nil {
				t.Errorf("flag --%s is not persistent", tt.flagName)
			}
			if localFlag != nil {
				t.Errorf("flag --%s should only be in PersistentFlags, not local Flags", tt.flagName)
			}
		})
	}
}

func TestGlobalFlagsDefaultValues(t *testing.T) {
	build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
	rootCmd := NewRootCmd(build)
	v := viper.New()
	ctx := WithViper(context.Background(), v)
	rootCmd.SetContext(ctx)

	config := GetConfig(rootCmd)

	if config.JSON {
		t.Error("JSON flag should default to false")
	}
	if config.NoColor {
		t.Error("NoColor flag should default to false")
	}
	if config.Interactive {
		t.Error("Interactive flag should default to false")
	}
}

func TestGlobalFlagsParsing(t *testing.T) {
	tests := []struct {
		name        string
		setupViper  func(*viper.Viper)
		wantJSON    bool
		wantNoColor bool
		wantInteractive bool
	}{
		{
			name:        "no flags set",
			setupViper:  func(v *viper.Viper) {},
			wantJSON:    false,
			wantNoColor: false,
			wantInteractive: false,
		},
		{
			name:        "json flag only",
			setupViper:  func(v *viper.Viper) { v.Set("json", true) },
			wantJSON:    true,
			wantNoColor: false,
			wantInteractive: false,
		},
		{
			name:        "no-color flag only",
			setupViper:  func(v *viper.Viper) { v.Set("no-color", true) },
			wantJSON:    false,
			wantNoColor: true,
			wantInteractive: false,
		},
		{
			name:        "interactive flag only",
			setupViper:  func(v *viper.Viper) { v.Set("interactive", true) },
			wantJSON:    false,
			wantNoColor: false,
			wantInteractive: true,
		},
		{
			name: "all flags set",
			setupViper: func(v *viper.Viper) {
				v.Set("json", true)
				v.Set("no-color", true)
				v.Set("interactive", true)
			},
			wantJSON:    true,
			wantNoColor: true,
			wantInteractive: true,
		},
		{
			name: "json and no-color",
			setupViper: func(v *viper.Viper) {
				v.Set("json", true)
				v.Set("no-color", true)
			},
			wantJSON:    true,
			wantNoColor: true,
			wantInteractive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			tt.setupViper(v)

			cmd := &cobra.Command{
				Use: "test",
				Run: func(cmd *cobra.Command, args []string) {},
			}

			ctx := WithViper(context.Background(), v)
			cmd.SetContext(ctx)

			config := GetConfig(cmd)

			if config.JSON != tt.wantJSON {
				t.Errorf("JSON = %v, want %v", config.JSON, tt.wantJSON)
			}
			if config.NoColor != tt.wantNoColor {
				t.Errorf("NoColor = %v, want %v", config.NoColor, tt.wantNoColor)
			}
			if config.Interactive != tt.wantInteractive {
				t.Errorf("Interactive = %v, want %v", config.Interactive, tt.wantInteractive)
			}
		})
	}
}

func TestGlobalFlagsAvailableToSubcommands(t *testing.T) {
	v := viper.New()
	v.Set("json", true)
	v.Set("no-color", true)

	subCmd := &cobra.Command{
		Use: "subtest",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	ctx := WithViper(context.Background(), v)
	subCmd.SetContext(ctx)

	config := GetConfig(subCmd)

	if !config.JSON {
		t.Error("JSON flag should be available to subcommands")
	}
	if !config.NoColor {
		t.Error("NoColor flag should be available to subcommands")
	}
}

func TestGlobalFlagsHelpText(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		wantContains string
	}{
		{
			name:         "json flag help",
			flagName:     "json",
			wantContains: "JSON",
		},
		{
			name:         "no-color flag help",
			flagName:     "no-color",
			wantContains: "color",
		},
		{
			name:         "interactive flag help",
			flagName:     "interactive",
			wantContains: "interactive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
			rootCmd := NewRootCmd(build)
			flag := rootCmd.PersistentFlags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag --%s not found", tt.flagName)
			}

			usage := flag.Usage
			if !strings.Contains(strings.ToLower(usage), strings.ToLower(tt.wantContains)) {
				t.Errorf("flag --%s usage %q should contain %q", tt.flagName, usage, tt.wantContains)
			}
		})
	}
}

func TestFlagDescriptionsAreHelpful(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		wantContains string
	}{
		{
			name:         "json flag mentions scripting",
			flagName:     "json",
			wantContains: "scripting",
		},
		{
			name:         "no-color flag mentions NO_COLOR",
			flagName:     "no-color",
			wantContains: "NO_COLOR",
		},
		{
			name:         "interactive flag mentions TTY",
			flagName:     "interactive",
			wantContains: "TTY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
			rootCmd := NewRootCmd(build)
			flag := rootCmd.PersistentFlags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("flag --%s not found", tt.flagName)
			}

			usage := flag.Usage
			if !strings.Contains(usage, tt.wantContains) {
				t.Errorf("flag --%s usage should contain helpful context about %q, got: %q", tt.flagName, tt.wantContains, usage)
			}
		})
	}
}

func TestVersionCommand(t *testing.T) {
	var buf bytes.Buffer

	build := BuildInfo{Version: "v1.0.0", Commit: "abc123", Date: "2025-10-19"}
	rootCmd := NewRootCmd(build)
	rootCmd.SetOut(&buf)

	v := viper.New()
	ctx := WithViper(context.Background(), v)
	rootCmd.SetArgs([]string{"version"})

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Tracks") {
		t.Error("version output should contain 'Tracks'")
	}
}

func TestNewCommand(t *testing.T) {
	var buf bytes.Buffer

	build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
	rootCmd := NewRootCmd(build)
	rootCmd.SetOut(&buf)

	v := viper.New()
	ctx := WithViper(context.Background(), v)
	rootCmd.SetArgs([]string{"new", "testproject"})

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		t.Fatalf("new command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "testproject") {
		t.Error("new command output should contain project name")
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name    string
		version string
		commit  string
		date    string
	}{
		{
			name:    "with version info",
			version: "v1.0.0",
			commit:  "abc123",
			date:    "2025-10-19",
		},
		{
			name:    "with dev version",
			version: "dev",
			commit:  "none",
			date:    "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Execute(tt.version, tt.commit, tt.date)

			if err != nil {
				t.Errorf("Execute should not error with help flag, got: %v", err)
			}
		})
	}
}

func TestExecuteWithCommand(t *testing.T) {
	err := Execute("v1.0.0", "abc123", "2025-10-19")
	if err != nil {
		t.Errorf("Execute with version command should not error, got: %v", err)
	}
}

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name  string
		build BuildInfo
		want  string
	}{
		{
			name:  "returns version when not dev",
			build: BuildInfo{Version: "v1.2.3", Commit: "abc", Date: "2025"},
			want:  "v1.2.3",
		},
		{
			name:  "returns dev when version is dev",
			build: BuildInfo{Version: "dev", Commit: "none", Date: "unknown"},
			want:  "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.build.getVersion()
			if got != tt.want {
				t.Errorf("getVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRootCommandUsage(t *testing.T) {
	build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
	rootCmd := NewRootCmd(build)

	use := rootCmd.Use
	if use != "tracks" {
		t.Errorf("root command Use = %q, want %q", use, "tracks")
	}

	if rootCmd.Short == "" {
		t.Error("root command should have a Short description")
	}

	if rootCmd.Long == "" {
		t.Error("root command should have a Long description")
	}

	if rootCmd.Example == "" {
		t.Error("root command should have Example usage")
	}
}

func TestRootCommandExamples(t *testing.T) {
	build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
	rootCmd := NewRootCmd(build)

	examples := rootCmd.Example
	requiredExamples := []string{
		"tracks new myapp",
		"tracks version",
		"tracks --json version",
	}

	for _, example := range requiredExamples {
		if !strings.Contains(examples, example) {
			t.Errorf("Example field should contain %q", example)
		}
	}
}

func TestCommandDescriptionsAreDetailed(t *testing.T) {
	build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
	rootCmd := NewRootCmd(build)

	versionCmd := rootCmd.Commands()[0]
	if versionCmd.Long == "" {
		t.Error("version command should have a Long description")
	}

	newCmd := rootCmd.Commands()[1]
	if newCmd.Long == "" {
		t.Error("new command should have a Long description")
	}

	if !strings.Contains(newCmd.Long, "templ") {
		t.Error("new command Long description should mention templ")
	}
	if !strings.Contains(newCmd.Long, "SQLC") {
		t.Error("new command Long description should mention SQLC")
	}
}

func TestVersionCommandDetails(t *testing.T) {
	expectedVersion := "v1.2.3"
	expectedCommit := "abc123def"
	expectedDate := "2025-10-19"

	var buf bytes.Buffer

	build := BuildInfo{Version: expectedVersion, Commit: expectedCommit, Date: expectedDate}
	rootCmd := NewRootCmd(build)
	rootCmd.SetOut(&buf)

	v := viper.New()
	ctx := WithViper(context.Background(), v)
	rootCmd.SetArgs([]string{"version"})

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, expectedVersion) {
		t.Errorf("version output should contain version %q", expectedVersion)
	}
	if !strings.Contains(output, expectedCommit) {
		t.Errorf("version output should contain commit %q", expectedCommit)
	}
	if !strings.Contains(output, expectedDate) {
		t.Errorf("version output should contain date %q", expectedDate)
	}
}

func TestNewCommandMissingArg(t *testing.T) {
	var buf bytes.Buffer

	build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
	rootCmd := NewRootCmd(build)
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	v := viper.New()
	ctx := WithViper(context.Background(), v)
	rootCmd.SetArgs([]string{"new"})

	err := rootCmd.ExecuteContext(ctx)
	if err == nil {
		t.Error("new command without project name should error")
	}
}

func TestNewCommandWithFlags(t *testing.T) {
	var buf bytes.Buffer

	build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
	rootCmd := NewRootCmd(build)
	rootCmd.SetOut(&buf)

	v := viper.New()
	if err := v.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json")); err != nil {
		t.Fatalf("failed to bind json flag: %v", err)
	}
	ctx := WithViper(context.Background(), v)
	rootCmd.SetArgs([]string{"--json", "new", "testproject"})

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		t.Fatalf("new command with flags failed: %v", err)
	}

	config := GetConfig(rootCmd)
	if !config.JSON {
		t.Error("JSON flag should be set when passed to new command")
	}

	output := buf.String()
	if !strings.Contains(output, "testproject") {
		t.Error("new command output should contain project name")
	}
}

func TestConfigStoredInContext(t *testing.T) {
	build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
	rootCmd := NewRootCmd(build)

	v := viper.New()
	if err := v.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json")); err != nil {
		t.Fatalf("failed to bind json flag: %v", err)
	}
	if err := v.BindPFlag("interactive", rootCmd.PersistentFlags().Lookup("interactive")); err != nil {
		t.Fatalf("failed to bind interactive flag: %v", err)
	}
	if err := v.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color")); err != nil {
		t.Fatalf("failed to bind no-color flag: %v", err)
	}

	ctx := WithViper(context.Background(), v)
	rootCmd.SetArgs([]string{"--json", "--interactive", "version"})

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		t.Fatalf("command execution failed: %v", err)
	}

	config := GetConfig(rootCmd)
	if !config.JSON {
		t.Error("JSON flag should be set in context")
	}
	if !config.Interactive {
		t.Error("Interactive flag should be set in context")
	}
	if config.NoColor {
		t.Error("NoColor flag should not be set")
	}
}

func TestGetConfigWithoutContext(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	cmd.SetContext(context.Background())

	config := GetConfig(cmd)

	if config.JSON {
		t.Error("JSON should default to false when no Viper in context")
	}
	if config.NoColor {
		t.Error("NoColor should default to false when no Viper in context")
	}
	if config.Interactive {
		t.Error("Interactive should default to false when no Viper in context")
	}
}

func TestGetViper(t *testing.T) {
	t.Run("returns viper from context", func(t *testing.T) {
		v := viper.New()
		v.Set("test-key", "test-value")

		cmd := &cobra.Command{Use: "test"}
		ctx := context.WithValue(context.Background(), viperKey{}, v)
		cmd.SetContext(ctx)

		result := GetViper(cmd)
		if result.GetString("test-key") != "test-value" {
			t.Error("GetViper should return Viper instance from context")
		}
	})

	t.Run("returns new viper when not in context", func(t *testing.T) {
		cmd := &cobra.Command{Use: "test"}
		cmd.SetContext(context.Background())

		result := GetViper(cmd)
		if result == nil {
			t.Error("GetViper should return new Viper instance when none in context")
		}
	})
}

func TestEnvironmentVariables(t *testing.T) {
	t.Run("TRACKS_JSON sets json flag", func(t *testing.T) {
		t.Setenv("TRACKS_JSON", "true")

		if err := Execute("v1.0.0", "abc123", "2025-10-19"); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
	})

	t.Run("TRACKS_NO_COLOR sets no-color flag", func(t *testing.T) {
		t.Setenv("TRACKS_NO_COLOR", "true")

		if err := Execute("v1.0.0", "abc123", "2025-10-19"); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
	})

	t.Run("TRACKS_INTERACTIVE sets interactive flag", func(t *testing.T) {
		t.Setenv("TRACKS_INTERACTIVE", "true")

		if err := Execute("v1.0.0", "abc123", "2025-10-19"); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
	})

	t.Run("NO_COLOR standard env var sets no-color flag", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")

		if err := Execute("v1.0.0", "abc123", "2025-10-19"); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
	})

	t.Run("NO_COLOR with empty value sets no-color flag", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")

		build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
		rootCmd := NewRootCmd(build)
		v := setupViperWithNoColor(t, rootCmd)
		ctx := WithViper(context.Background(), v)
		rootCmd.SetContext(ctx)

		config := GetConfig(rootCmd)
		if !config.NoColor {
			t.Error("NO_COLOR with empty value should set NoColor flag")
		}
	})

	t.Run("flags override NO_COLOR environment variable", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")

		build := BuildInfo{Version: "dev", Commit: "none", Date: "unknown"}
		rootCmd := NewRootCmd(build)
		v := setupViperWithNoColor(t, rootCmd)
		ctx := WithViper(context.Background(), v)
		rootCmd.SetArgs([]string{"--no-color=false"})

		if err := rootCmd.ExecuteContext(ctx); err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		config := GetConfig(rootCmd)
		if config.NoColor {
			t.Error("CLI flag should override NO_COLOR environment variable")
		}
	})
}
