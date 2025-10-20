package main

import (
	"fmt"
	"os"

	"github.com/anomalousventures/tracks/internal/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := cli.Execute(version, commit, date); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
