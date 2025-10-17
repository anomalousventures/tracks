package main

import (
	"fmt"
	"os"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	fmt.Fprintf(os.Stderr, "Tracks MCP Server %s\n", version)
	fmt.Fprintf(os.Stderr, "Commit: %s\n", commit)
	fmt.Fprintf(os.Stderr, "Built: %s\n", date)
	fmt.Fprintln(os.Stderr, "\nMCP server implementation coming soon...")
	fmt.Fprintln(os.Stderr, "This will enable AI-powered development via Model Context Protocol")
	os.Exit(0)
}
