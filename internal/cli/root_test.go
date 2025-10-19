package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

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
			persistentFlag := rootCmd.PersistentFlags().Lookup(tt.flagName)
			localFlag := rootCmd.Flags().Lookup(tt.flagName)

			if persistentFlag == nil {
				t.Errorf("flag --%s is not persistent", tt.flagName)
			}
			if localFlag != nil && !localFlag.Changed {
				t.Errorf("flag --%s should only be in PersistentFlags, not local Flags", tt.flagName)
			}
		})
	}
}

func TestGlobalFlagsDefaultValues(t *testing.T) {
	resetGlobalConfig()

	config := GetGlobalConfig()

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
		args        []string
		wantJSON    bool
		wantNoColor bool
		wantInteractive bool
	}{
		{
			name:        "no flags set",
			args:        []string{},
			wantJSON:    false,
			wantNoColor: false,
			wantInteractive: false,
		},
		{
			name:        "json flag only",
			args:        []string{"--json"},
			wantJSON:    true,
			wantNoColor: false,
			wantInteractive: false,
		},
		{
			name:        "no-color flag only",
			args:        []string{"--no-color"},
			wantJSON:    false,
			wantNoColor: true,
			wantInteractive: false,
		},
		{
			name:        "interactive flag only",
			args:        []string{"--interactive"},
			wantJSON:    false,
			wantNoColor: false,
			wantInteractive: true,
		},
		{
			name:        "all flags set",
			args:        []string{"--json", "--no-color", "--interactive"},
			wantJSON:    true,
			wantNoColor: true,
			wantInteractive: true,
		},
		{
			name:        "json and no-color",
			args:        []string{"--json", "--no-color"},
			wantJSON:    true,
			wantNoColor: true,
			wantInteractive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetGlobalConfig()

			cmd := &cobra.Command{
				Use: "test",
				Run: func(cmd *cobra.Command, args []string) {},
			}
			rootCmd.AddCommand(cmd)
			defer rootCmd.RemoveCommand(cmd)

			rootCmd.SetArgs(append(tt.args, "test"))

			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("failed to execute command: %v", err)
			}

			config := GetGlobalConfig()

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
	resetGlobalConfig()

	subCmd := &cobra.Command{
		Use: "subtest",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	rootCmd.AddCommand(subCmd)
	defer rootCmd.RemoveCommand(subCmd)

	rootCmd.SetArgs([]string{"--json", "--no-color", "subtest"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("failed to execute command: %v", err)
	}

	config := GetGlobalConfig()

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

func TestVersionCommand(t *testing.T) {
	resetGlobalConfig()

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Tracks") {
		t.Error("version output should contain 'Tracks'")
	}
}

func TestNewCommand(t *testing.T) {
	resetGlobalConfig()

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"new", "testproject"})

	if err := rootCmd.Execute(); err != nil {
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
			resetGlobalConfig()
			rootCmd.SetArgs([]string{"--help"})

			err := Execute(tt.version, tt.commit, tt.date)

			if err != nil {
				t.Errorf("Execute should not error with help flag, got: %v", err)
			}
		})
	}
}

func TestExecuteWithCommand(t *testing.T) {
	resetGlobalConfig()

	rootCmd.SetArgs([]string{"version"})

	err := Execute("v1.0.0", "abc123", "2025-10-19")
	if err != nil {
		t.Errorf("Execute with version command should not error, got: %v", err)
	}
}

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func()
		want        string
	}{
		{
			name: "returns version when not dev",
			setupFunc: func() {
				version = "v1.2.3"
			},
			want: "v1.2.3",
		},
		{
			name: "returns dev when version is dev",
			setupFunc: func() {
				version = "dev"
			},
			want: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()
			got := getVersion()
			if got != tt.want {
				t.Errorf("getVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRootCommandUsage(t *testing.T) {
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
}

func TestVersionCommandDetails(t *testing.T) {
	resetGlobalConfig()
	version = "v1.2.3"
	commit = "abc123def"
	date = "2025-10-19"

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"version"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, version) {
		t.Errorf("version output should contain version %q", version)
	}
	if !strings.Contains(output, commit) {
		t.Errorf("version output should contain commit %q", commit)
	}
	if !strings.Contains(output, date) {
		t.Errorf("version output should contain date %q", date)
	}
}

func TestNewCommandMissingArg(t *testing.T) {
	resetGlobalConfig()

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"new"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("new command without project name should error")
	}
}

func TestNewCommandWithFlags(t *testing.T) {
	resetGlobalConfig()

	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"--json", "new", "testproject"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("new command with flags failed: %v", err)
	}

	config := GetGlobalConfig()
	if !config.JSON {
		t.Error("JSON flag should be set when passed to new command")
	}

	output := buf.String()
	if !strings.Contains(output, "testproject") {
		t.Error("new command output should contain project name")
	}
}

func resetGlobalConfig() {
	globalConfig = GlobalConfig{}
	rootCmd.SetArgs([]string{})
	version = "dev"
	commit = "none"
	date = "unknown"
}
