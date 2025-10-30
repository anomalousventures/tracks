package cli

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/anomalousventures/tracks/internal/cli/commands"
	"github.com/anomalousventures/tracks/internal/cli/interfaces"
	"github.com/anomalousventures/tracks/internal/cli/renderer"
	"github.com/anomalousventures/tracks/internal/cli/ui"
	"github.com/anomalousventures/tracks/internal/generator"
	"github.com/anomalousventures/tracks/internal/validation"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
	Verbose     bool
	Quiet       bool
	LogLevel    string
}

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
func NewRootCmd(build BuildInfo) (*cobra.Command, error) {
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
		Version: build.GetVersion(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			v := GetViper(cmd)
			if v.GetBool("verbose") && v.GetBool("quiet") {
				return fmt.Errorf("--verbose and --quiet flags are mutually exclusive")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			r := NewRendererFromCommand(cmd)

			r.Section(interfaces.Section{
				Body: "Interactive TUI mode coming in Phase 4. Use --help for available commands.",
			})

			FlushRenderer(cmd, r)
		},
	}

	rootCmd.PersistentFlags().Bool("json", false, "Output in JSON format (useful for scripting)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output (respects NO_COLOR env var)")
	rootCmd.PersistentFlags().Bool("interactive", false, "Force interactive TUI mode even in non-TTY environments")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output (shows detailed information)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Quiet mode (suppress non-error output)")

	// Configure viper to read flags and environment variables
	v := viper.New()
	if err := v.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json")); err != nil {
		return nil, fmt.Errorf("failed to bind json flag: %w", err)
	}
	if err := v.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color")); err != nil {
		return nil, fmt.Errorf("failed to bind no-color flag: %w", err)
	}
	if err := v.BindPFlag("interactive", rootCmd.PersistentFlags().Lookup("interactive")); err != nil {
		return nil, fmt.Errorf("failed to bind interactive flag: %w", err)
	}
	if err := v.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		return nil, fmt.Errorf("failed to bind verbose flag: %w", err)
	}
	if err := v.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet")); err != nil {
		return nil, fmt.Errorf("failed to bind quiet flag: %w", err)
	}

	v.SetEnvPrefix("TRACKS")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		v.SetDefault("no-color", true)
	}
	if os.Getenv("TRACKS_LOG_LEVEL") == "" {
		v.SetDefault("log-level", "off")
	}

	// Wire up dependencies following ADR-001 (Dependency Injection)
	// Logger is configured from viper (which reads TRACKS_LOG_LEVEL env var)
	// Defaults to disabled per ADR-003 (Context Propagation)
	logLevelStr := v.GetString("log-level")
	logLevel, err := zerolog.ParseLevel(logLevelStr)
	if err != nil {
		logLevel = zerolog.Disabled
	}
	logger := zerolog.New(os.Stderr).Level(logLevel)

	// Create validator from validation package (moved in Phase 3)
	validator := validation.NewValidator(logger)

	// Generator will be implemented in Phase 3+
	// Using noop implementation that returns "not yet implemented" errors
	projectGenerator := generator.NewNoopGenerator()

	// Make viper available through context (ADR-003)
	ctx := WithViper(context.Background(), v)
	rootCmd.SetContext(ctx)

	versionCmd := commands.NewVersionCommand(build, NewRendererFromCommand, FlushRenderer)
	rootCmd.AddCommand(versionCmd.Command())

	newCmd := commands.NewNewCommand(validator, projectGenerator, NewRendererFromCommand, FlushRenderer)
	rootCmd.AddCommand(newCmd.Command())

	return rootCmd, nil
}

// Execute initializes and runs the root command with build information.
// NewRootCmd handles all configuration setup (viper, logger, dependencies)
// and attaches context before returning. This function simply creates the
// command and executes it.
func Execute(versionStr, commitStr, dateStr string) error {
	build := BuildInfo{
		Version: versionStr,
		Commit:  commitStr,
		Date:    dateStr,
	}

	rootCmd, err := NewRootCmd(build)
	if err != nil {
		// Log technical error to stderr
		fmt.Fprintf(os.Stderr, "Error initializing CLI: %v\n", err)
		// Return user-friendly error
		return fmt.Errorf("failed to initialize tracks CLI - please report this issue at https://github.com/anomalousventures/tracks/issues")
	}
	return rootCmd.Execute()
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
		Verbose:     v.GetBool("verbose"),
		Quiet:       v.GetBool("quiet"),
		LogLevel:    v.GetString("log-level"),
	}
}

// NewRendererFromCommand creates an appropriate renderer based on command configuration.
func NewRendererFromCommand(cmd *cobra.Command) interfaces.Renderer {
	cfg := GetConfig(cmd)

	uiMode := ui.DetectMode(ui.UIConfig{
		Mode:        ui.ModeAuto,
		JSON:        cfg.JSON,
		NoColor:     cfg.NoColor,
		Interactive: cfg.Interactive,
	})

	if uiMode == ui.ModeJSON {
		return renderer.NewJSONRenderer(cmd.OutOrStdout())
	}

	// Set NO_COLOR env var if --no-color flag is set, so Lip Gloss respects it
	if cfg.NoColor {
		os.Setenv("NO_COLOR", "1")
	}

	return renderer.NewConsoleRenderer(cmd.OutOrStdout())
}

// FlushRenderer flushes the renderer and handles errors by writing to stderr and exiting.
func FlushRenderer(cmd *cobra.Command, r interfaces.Renderer) {
	if err := r.Flush(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", err)
		os.Exit(1)
	}
}

func (b BuildInfo) GetVersion() string {
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

func (b BuildInfo) GetCommit() string {
	return b.Commit
}

func (b BuildInfo) GetDate() string {
	return b.Date
}
