package cli

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// BuildInfo contains version metadata for the CLI.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// Config holds the global CLI configuration.
type Config struct {
	JSON        bool
	NoColor     bool
	Interactive bool
}

// viperKey is used as a type-safe key for storing Viper in context.
type viperKey struct{}

// WithViper returns a new context with the provided Viper instance.
func WithViper(ctx context.Context, v *viper.Viper) context.Context {
	return context.WithValue(ctx, viperKey{}, v)
}

// ViperFromContext retrieves the Viper instance from the context, if present.
func ViperFromContext(ctx context.Context) *viper.Viper {
	if v := ctx.Value(viperKey{}); v != nil {
		if vv, ok := v.(*viper.Viper); ok {
			return vv
		}
	}
	return nil
}

// NewRootCmd creates a new root command with all flags and subcommands configured.
// This returns a fresh command instance to avoid cross-test state coupling.
func NewRootCmd(build BuildInfo) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "tracks",
		Short: "A productive web framework for Go",
		Long: `Tracks is a code-generating web framework for Go.

It generates complete applications with type-safe templates (templ),
type-safe SQL (SQLC), built-in authentication/authorization,
and an interactive TUI for code generation.

Generates idiomatic Go code you'd write yourself. No magic, full control.`,
		Example: `  # Create a new Tracks application
  tracks new myapp

  # Show version information
  tracks version

  # Get JSON output for scripting
  tracks --json version

  # View help for any command
  tracks help new`,
		Version: build.getVersion(),
	}

	rootCmd.PersistentFlags().Bool("json", false, "Output in JSON format (useful for scripting)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output (respects NO_COLOR env var)")
	rootCmd.PersistentFlags().Bool("interactive", false, "Force interactive TUI mode even in non-TTY environments")

	rootCmd.AddCommand(versionCmd(build))
	rootCmd.AddCommand(newCmd())

	return rootCmd
}

// Execute initializes and runs the root command with build information.
// It creates a fresh command instance and Viper configuration, binds CLI flags,
// sets up environment variable support, and makes the configuration available
// via context to all commands.
func Execute(versionStr, commitStr, dateStr string) error {
	build := BuildInfo{
		Version: versionStr,
		Commit:  commitStr,
		Date:    dateStr,
	}

	rootCmd := NewRootCmd(build)

	v := viper.New()

	if err := v.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json")); err != nil {
		return fmt.Errorf("failed to bind json flag: %w", err)
	}
	if err := v.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color")); err != nil {
		return fmt.Errorf("failed to bind no-color flag: %w", err)
	}
	if err := v.BindPFlag("interactive", rootCmd.PersistentFlags().Lookup("interactive")); err != nil {
		return fmt.Errorf("failed to bind interactive flag: %w", err)
	}

	v.SetEnvPrefix("TRACKS")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		v.SetDefault("no-color", true)
	}

	ctx := WithViper(context.Background(), v)

	return rootCmd.ExecuteContext(ctx)
}

// GetViper extracts the Viper instance from the command's context.
// Returns a new Viper instance if none is found in context (useful for testing).
func GetViper(cmd *cobra.Command) *viper.Viper {
	if v := ViperFromContext(cmd.Context()); v != nil {
		return v
	}
	return viper.New()
}

// GetConfig extracts the configuration from the command's Viper instance.
// This is the primary way commands should access configuration values.
func GetConfig(cmd *cobra.Command) Config {
	v := GetViper(cmd)
	return Config{
		JSON:        v.GetBool("json"),
		NoColor:     v.GetBool("no-color"),
		Interactive: v.GetBool("interactive"),
	}
}

func versionCmd(build BuildInfo) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Display the version number, git commit hash, and build date for this Tracks CLI binary.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "Tracks %s\n", build.getVersion())
			fmt.Fprintf(cmd.OutOrStdout(), "Commit: %s\n", build.Commit)
			fmt.Fprintf(cmd.OutOrStdout(), "Built: %s\n", build.Date)
		},
	}
}

func newCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Tracks application",
		Long: `Create a new Tracks application with the specified project name.

This command generates a complete Go web application with:
  - Proper project structure following Go best practices
  - Type-safe templates using templ
  - Type-safe SQL queries using SQLC
  - Built-in authentication and authorization (RBAC)
  - Development tooling (Makefile, hot-reload, linting)
  - Docker and CI/CD configurations

The generated application is production-ready and follows idiomatic Go patterns.`,
		Example: `  # Create a new application with default settings
  tracks new myapp

  # Future: Specify database driver
  tracks new myapp --db postgres

  # Future: Custom Go module path
  tracks new myapp --module github.com/myorg/myapp

  # Future: Skip git initialization
  tracks new myapp --no-git`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectName := args[0]
			fmt.Fprintf(cmd.OutOrStdout(), "Creating new Tracks application: %s\n", projectName)
			fmt.Fprintln(cmd.OutOrStdout(), "(Full implementation coming soon)")
		},
	}
}

func (b BuildInfo) getVersion() string {
	if b.Version != "dev" {
		return b.Version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			return info.Main.Version
		}
	}

	return "dev"
}
