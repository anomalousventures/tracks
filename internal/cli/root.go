package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type GlobalConfig struct {
	JSON        bool
	NoColor     bool
	Interactive bool
}

var globalConfig GlobalConfig

func init() {
	rootCmd.PersistentFlags().BoolVar(&globalConfig.JSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&globalConfig.NoColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVar(&globalConfig.Interactive, "interactive", false, "Force interactive TUI mode")
}

var rootCmd = &cobra.Command{
	Use:   "tracks",
	Short: "A productive web framework for Go",
	Long: `Tracks is a code-generating web framework for Go.

It generates complete applications with type-safe templates (templ),
type-safe SQL (SQLC), built-in authentication/authorization,
and an interactive TUI for code generation.

Generates idiomatic Go code you'd write yourself. No magic, full control.`,
	Version: getVersion(),
}

func init() {
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(newCmd())
}

func Execute(versionStr, commitStr, dateStr string) error {
	version = versionStr
	commit = commitStr
	date = dateStr
	rootCmd.Version = getVersion()
	return rootCmd.Execute()
}

func GetGlobalConfig() GlobalConfig {
	return globalConfig
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
