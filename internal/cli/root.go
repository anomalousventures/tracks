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

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Config holds the global CLI configuration.
type Config struct {
	JSON        bool
	NoColor     bool
	Interactive bool
}

// viperKey is used as a type-safe key for storing Viper in context.
type viperKey struct{}

// NewRootCmd creates a new root command with all flags and subcommands configured.
// This returns a fresh command instance to avoid cross-test state coupling.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "tracks",
		Short: "A productive web framework for Go",
		Long: `Tracks is a code-generating web framework for Go.

It generates complete applications with type-safe templates (templ),
type-safe SQL (SQLC), built-in authentication/authorization,
and an interactive TUI for code generation.

Generates idiomatic Go code you'd write yourself. No magic, full control.`,
	}

	rootCmd.PersistentFlags().Bool("json", false, "Output in JSON format")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")
	rootCmd.PersistentFlags().Bool("interactive", false, "Force interactive TUI mode")

	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(newCmd())

	return rootCmd
}

// Execute initializes and runs the root command with build information.
// It creates a fresh command instance and Viper configuration, binds CLI flags,
// sets up environment variable support, and makes the configuration available
// via context to all commands.
func Execute(versionStr, commitStr, dateStr string) error {
	rootCmd := NewRootCmd()

	version = versionStr
	commit = commitStr
	date = dateStr
	rootCmd.Version = getVersion()

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

	ctx := context.WithValue(context.Background(), viperKey{}, v)

	return rootCmd.ExecuteContext(ctx)
}

// GetViper extracts the Viper instance from the command's context.
// Returns a new Viper instance if none is found in context (useful for testing).
func GetViper(cmd *cobra.Command) *viper.Viper {
	if v := cmd.Context().Value(viperKey{}); v != nil {
		return v.(*viper.Viper)
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

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "Tracks %s\n", getVersion())
			fmt.Fprintf(cmd.OutOrStdout(), "Commit: %s\n", commit)
			fmt.Fprintf(cmd.OutOrStdout(), "Built: %s\n", date)
		},
	}
}

func newCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new Tracks application",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectName := args[0]
			fmt.Fprintf(cmd.OutOrStdout(), "Creating new Tracks application: %s\n", projectName)
			fmt.Fprintln(cmd.OutOrStdout(), "(Full implementation coming soon)")
		},
	}
}

func getVersion() string {
	if version != "dev" {
		return version
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			return info.Main.Version
		}
	}

	return "dev"
}
