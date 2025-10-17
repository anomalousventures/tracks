package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tracks",
		Short: "A productive web framework for Go",
		Long: `Tracks is a code-generating web framework for Go.

It generates complete applications with type-safe templates (templ),
type-safe SQL (SQLC), built-in authentication/authorization,
and an interactive TUI for code generation.

Generates idiomatic Go code you'd write yourself. No magic, full control.`,
		Version: getVersion(),
	}

	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(newCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Tracks %s\n", getVersion())
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Built: %s\n", date)
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
			fmt.Printf("Creating new Tracks application: %s\n", projectName)
			fmt.Println("(Full implementation coming soon)")
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
