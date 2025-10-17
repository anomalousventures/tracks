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
	fmt.Fprintln(os.Stderr, "MCP server implementation coming soon...")
	fmt.Fprintln(os.Stderr, "This will enable AI-powered development via Model Context Protocol")
	os.Exit(0)
}
