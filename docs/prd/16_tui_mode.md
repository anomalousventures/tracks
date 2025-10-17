# TUI Mode

**[← Back to Summary](./0_summary.md)**

## Overview

Tracks provides a rich Terminal User Interface (TUI) that launches by default when running `tracks` without arguments. Built with Bubble Tea, it offers an interactive dashboard for monitoring, debugging, and code generation with real-time updates and visual feedback.

## Goals

- Interactive dashboard for monitoring and code generation
- Real-time log streaming and filtering
- Visual job queue management
- Database schema inspection
- Zero-config launch experience

## User Stories

- As a developer, I want to launch TUI by just typing `tracks`
- As a developer, I want to monitor logs in real-time
- As a developer, I want to generate code through interactive forms
- As a developer, I want to inspect job queue status
- As a developer, I want to detect schema drift visually

## TUI Architecture

```go
// cmd/tracks/main.go
package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/spf13/cobra"
)

func main() {
    if len(os.Args) == 1 {
        // No arguments - launch TUI
        runTUI()
        return
    }

    // Otherwise, execute CLI commands
    rootCmd := &cobra.Command{
        Use:   "tracks",
        Short: "A Rails-like Go web framework",
    }

    rootCmd.AddCommand(
        newCmd(),
        generateCmd(),
        dbCmd(),
        devCmd(),
        testCmd(),
        buildCmd(),
    )

    rootCmd.Execute()
}

func runTUI() {
    // Initialize with project detection
    config := detectProject()

    p := tea.NewProgram(
        NewDashboard(config),
        tea.WithAltScreen(),        // Use alternate screen buffer
        tea.WithMouseCellMotion(),  // Enable mouse support
        tea.WithFPS(30),            // Smooth animations
    )

    if _, err := p.Run(); err != nil {
        fmt.Printf("Error running TUI: %v\n", err)
        os.Exit(1)
    }
}
```

## TUI Model Structure

```go
// internal/tui/dashboard.go
package tui

import (
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/gofrs/uuid/v5"
)

type Dashboard struct {
    // Navigation
    activeTab     int
    tabs          []string

    // Screens
    overview      *OverviewScreen
    logs          *LogsScreen
    generator     *GeneratorScreen
    database      *DatabaseScreen
    jobs          *JobsScreen
    monitoring    *MonitoringScreen

    // State
    width         int
    height        int
    connected     bool
    lastUpdate    time.Time

    // Real-time data
    metrics       *Metrics
    logBuffer     *LogBuffer
    jobQueue      *JobQueue
}

func NewDashboard(config *Config) *Dashboard {
    return &Dashboard{
        tabs: []string{
            "Overview",
            "Logs",
            "Generate",
            "Database",
            "Jobs",
            "Monitoring",
        },
        activeTab:  0,
        overview:   NewOverviewScreen(),
        logs:       NewLogsScreen(),
        generator:  NewGeneratorScreen(),
        database:   NewDatabaseScreen(),
        jobs:       NewJobsScreen(),
        monitoring: NewMonitoringScreen(),
        metrics:    NewMetrics(),
        logBuffer:  NewLogBuffer(1000),
        jobQueue:   NewJobQueue(),
    }
}
```

## TUI Screens

### 1. Dashboard Overview

```text
┌─ Tracks Dashboard ─────────────────────────────────────────────┐
│ App: myapp | Env: development | Uptime: 2h 34m | v1.2.3       │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│ HTTP Requests (last 5min)                                      │
│ ████████████████████░░░░░░░░ 234 req/min                      │
│                                                                 │
│ Response Times                                                 │
│ P50: 23ms | P95: 89ms | P99: 234ms                            │
│                                                                 │
│ Database                                                       │
│ ██████████████░░░░░░░░░░░░░░ 1,234 queries (23ms avg)        │
│ Connections: 8/20 | Slow queries: 2                           │
│                                                                 │
│ Job Queue                                                      │
│ ▶ Pending: 12 | ⚡ Processing: 3 | ✗ Failed: 0                │
│                                                                 │
│ System Resources                                               │
│ CPU: 12% | Memory: 234MB/512MB | Goroutines: 45              │
│                                                                 │
│ Recent Errors: 0 | Warnings: 2                                 │
│                                                                 │
│ [Tab] Switch Screen | [r] Refresh | [q] Quit                   │
└────────────────────────────────────────────────────────────────┘
```

### 2. Logs Screen with Filtering

```text
┌─ Logs ──────────────────────────────────────────────────────────┐
│ Filter: [all ▼] Level: [info ▼] Search: [_________]    [Clear] │
├─────────────────────────────────────────────────────────────────┤
│ 14:23:45 INFO  [http] GET /users 200 23ms                      │
│ 14:23:46 DEBUG [repo] user_repo.list 45 rows 12ms              │
│ 14:23:47 INFO  [http] POST /users 201 156ms                    │
│ 14:23:48 INFO  [jobs] send_welcome_email enqueued              │
│ 14:23:49 ERROR [validation] email format invalid                │
│ 14:23:50 WARN  [cache] miss for key user:123                   │
│ 14:23:51 INFO  [auth] login success user=john@example.com      │
│ 14:23:52 DEBUG [db] transaction committed in 45ms              │
│ 14:23:53 INFO  [http] GET /dashboard 200 67ms                  │
│                                                                 │
│ ───────────────────────────────────────────────────────────────│
│ [↑↓] Navigate [/] Search [f] Filter [l] Level [c] Clear        │
│ [Space] Pause [Enter] Details [Tab] Next Screen [q] Quit       │
└─────────────────────────────────────────────────────────────────┘
```

### 3. Interactive Code Generator

```text
┌─ Code Generator ────────────────────────────────────────────────┐
│ Select Type: [▼ Handler]                                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ Resource Name: [UserProfile_______]                            │
│                                                                 │
│ ☑ Generate Handler                                             │
│ ☑ Generate Service                                             │
│ ☑ Generate Repository                                          │
│ ☐ Generate Migration                                           │
│ ☑ Generate Tests                                               │
│                                                                 │
│ Handler Methods:                                                │
│ ☑ Index (GET /)      ☑ Show (GET /:id)                        │
│ ☑ Create (POST /)    ☑ Update (PUT /:id)                      │
│ ☐ Delete (DELETE /:id)                                         │
│                                                                 │
│ Fields:                                                         │
│ ┌─────────────┬──────────┬──────────┬─────────┐              │
│ │ Name        │ Type     │ Required │ Index   │              │
│ ├─────────────┼──────────┼──────────┼─────────┤              │
│ │ user_id     │ uuid     │ ✓        │ ✓       │              │
│ │ display_name│ string   │ ✓        │         │              │
│ │ bio         │ text     │          │         │              │
│ │ avatar_url  │ string   │          │         │              │
│ └─────────────┴──────────┴──────────┴─────────┘              │
│                                                                 │
│ [↑↓] Navigate [Space] Toggle [Tab] Next Field                  │
│ [Enter] Generate [Esc] Cancel                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 4. Database Inspector

```text
┌─ Database Inspector ────────────────────────────────────────────┐
│ Driver: go-libsql | Database: myapp.db | Tables: 12           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ Tables                        │ Schema: users                   │
│ ┌───────────────┬──────────┐ │ ┌──────────┬─────────┬───────┐│
│ │ Name          │ Records   │ │ │ Column   │ Type    │ Index ││
│ ├───────────────┼──────────┤ │ ├──────────┼─────────┼───────┤│
│ │ ▶ users       │ 1,234     │ │ │ id       │ TEXT    │ PK    ││
│ │   sessions    │ 456       │ │ │ email    │ TEXT    │ UNIQUE││
│ │   roles       │ 5         │ │ │ username │ TEXT    │ UNIQUE││
│ │   permissions │ 23        │ │ │ name     │ TEXT    │       ││
│ │   user_roles  │ 1,456     │ │ │ verified │ BOOLEAN │       ││
│ │   audit_logs  │ 12,345    │ │ │ created  │ DATETIME│ INDEX ││
│ └───────────────┴──────────┘ │ └──────────┴─────────┴───────┘│
│                               │                                 │
│ Migrations                    │ Indexes: 4                      │
│ Current: 0023_add_avatar.sql │ Foreign Keys: 2                 │
│ Pending: 0                   │ Size: 12.3 MB                   │
│                               │                                 │
│ [↑↓] Navigate [Enter] Details [m] Migrate [s] SQL Console      │
└─────────────────────────────────────────────────────────────────┘
```

### 5. Job Queue Monitor

```text
┌─ Job Queue Monitor ─────────────────────────────────────────────┐
│ Provider: memory | Workers: 4 | Uptime: 2h 34m                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ Queue Status                                                    │
│ ┌────────────────┬──────────┬───────────┬──────────┐         │
│ │ Queue          │ Pending  │ Processing│ Failed   │         │
│ ├────────────────┼──────────┼───────────┼──────────┤         │
│ │ default        │ 12       │ 3         │ 0        │         │
│ │ emails         │ 45       │ 2         │ 1        │         │
│ │ exports        │ 2        │ 1         │ 0        │         │
│ │ cleanup        │ 0        │ 0         │ 0        │         │
│ └────────────────┴──────────┴───────────┴──────────┘         │
│                                                                 │
│ Active Jobs                                                     │
│ ┌───────────────────────┬────────────┬─────────────┐         │
│ │ Job ID                │ Type       │ Progress    │         │
│ ├───────────────────────┼────────────┼─────────────┤         │
│ │ 01234567-89ab-cdef... │ SendEmail  │ ████░░ 67%  │         │
│ │ fedcba98-7654-3210... │ ExportCSV  │ ██░░░░ 34%  │         │
│ │ abcdef12-3456-7890... │ ProcessImg │ █████░ 89%  │         │
│ └───────────────────────┴────────────┴─────────────┘         │
│                                                                 │
│ Recent Failures                                                │
│ 14:22:15 SendEmail: SMTP connection timeout                    │
│                                                                 │
│ [r] Retry Failed [c] Clear Failed [p] Pause [Tab] Next         │
└─────────────────────────────────────────────────────────────────┘
```

### 6. Performance Monitoring

```text
┌─ Performance Monitor ───────────────────────────────────────────┐
│ Sampling: 1s | Buffer: 5min | Auto-refresh: ON                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ Request Latency (ms)                                           │
│ 300 ┤                                                          │
│ 250 ┤              ╱╲                                          │
│ 200 ┤             ╱  ╲                    ╱╲                  │
│ 150 ┤        ╱╲  ╱    ╲                  ╱  ╲                 │
│ 100 ┤    ╱╲ ╱  ╲╱      ╲  ╱╲  ╱╲       ╱    ╲               │
│  50 ┤╱╲ ╱  ╲                ╲╱  ╲ ╱╲ ╱        ╲ ╱╲          │
│   0 └────────────────────────────────────────────────────────│
│     14:20      14:21      14:22      14:23      14:24        │
│                                                                 │
│ Database Queries/sec                                           │
│ 100 ┤                                                          │
│  80 ┤    ████  ████                    ████                   │
│  60 ┤████    ██    ████  ████  ████████    ████              │
│  40 ┤                  ██    ██                ████          │
│  20 ┤                                              ████       │
│   0 └────────────────────────────────────────────────────────│
│                                                                 │
│ Top Endpoints (by latency)                                     │
│ GET /api/reports/export     234ms  ████████████████████████   │
│ POST /api/images/process    156ms  ████████████████           │
│ GET /api/users/search        89ms  ███████████                │
│ GET /dashboard               45ms  ██████                     │
│                                                                 │
│ [s] Sample Rate [b] Buffer Size [e] Export [Tab] Next          │
└─────────────────────────────────────────────────────────────────┘
```

## Real-time Updates

```go
// internal/tui/realtime.go
package tui

import (
    "time"
    tea "github.com/charmbracelet/bubbletea"
)

// Messages for real-time updates
type LogMsg struct {
    Level   string
    Message string
    Time    time.Time
}

type MetricMsg struct {
    RequestRate float64
    ErrorRate   float64
    P50Latency  time.Duration
    P95Latency  time.Duration
    P99Latency  time.Duration
}

type JobUpdateMsg struct {
    ID       string
    Status   string
    Progress int
}

// Ticker for periodic updates
func tickEvery(duration time.Duration) tea.Cmd {
    return tea.Every(duration, func(t time.Time) tea.Msg {
        return TickMsg(t)
    })
}

// WebSocket connection for real-time data
func (d *Dashboard) connectWebSocket() tea.Cmd {
    return func() tea.Msg {
        ws, err := websocket.Dial("ws://localhost:8080/ws/tui")
        if err != nil {
            return ErrorMsg{err}
        }

        go d.handleWebSocketMessages(ws)
        return ConnectedMsg{}
    }
}

func (d *Dashboard) handleWebSocketMessages(ws *websocket.Conn) {
    for {
        var msg interface{}
        if err := websocket.JSON.Receive(ws, &msg); err != nil {
            d.Send(DisconnectedMsg{})
            return
        }

        switch v := msg.(type) {
        case map[string]interface{}:
            if v["type"] == "log" {
                d.Send(LogMsg{
                    Level:   v["level"].(string),
                    Message: v["message"].(string),
                    Time:    time.Now(),
                })
            }
        }
    }
}
```

## Keyboard Shortcuts

```go
// internal/tui/keys.go
package tui

import (
    "github.com/charmbracelet/bubbles/key"
)

type KeyMap struct {
    // Navigation
    Up       key.Binding
    Down     key.Binding
    Left     key.Binding
    Right    key.Binding
    PageUp   key.Binding
    PageDown key.Binding

    // Actions
    Select   key.Binding
    Back     key.Binding
    Quit     key.Binding
    Help     key.Binding

    // Screen-specific
    Filter   key.Binding
    Search   key.Binding
    Refresh  key.Binding
    Clear    key.Binding
    Export   key.Binding
}

var DefaultKeyMap = KeyMap{
    Up: key.NewBinding(
        key.WithKeys("up", "k"),
        key.WithHelp("↑/k", "up"),
    ),
    Down: key.NewBinding(
        key.WithKeys("down", "j"),
        key.WithHelp("↓/j", "down"),
    ),
    Select: key.NewBinding(
        key.WithKeys("enter", " "),
        key.WithHelp("enter/space", "select"),
    ),
    Quit: key.NewBinding(
        key.WithKeys("q", "ctrl+c"),
        key.WithHelp("q", "quit"),
    ),
    Filter: key.NewBinding(
        key.WithKeys("f"),
        key.WithHelp("f", "filter"),
    ),
    Search: key.NewBinding(
        key.WithKeys("/"),
        key.WithHelp("/", "search"),
    ),
}
```

## Configuration

```yaml
# .tracks/tui.yaml
tui:
  # Connection settings
  api:
    endpoint: "http://localhost:8080"
    websocket: "ws://localhost:8080/ws/tui"

  # Display preferences
  display:
    theme: "dark"  # dark, light, auto
    colors:
      primary: "#00ADD8"
      success: "#00C853"
      warning: "#FFB300"
      error: "#D50000"

  # Update intervals (ms)
  refresh:
    metrics: 1000
    logs: 100
    jobs: 2000

  # Buffer sizes
  buffers:
    logs: 1000
    metrics: 300

  # Feature flags
  features:
    animations: true
    mouse: true
    sound: false
```

## Advanced Features

### Split View Mode

```go
// Support for split view layouts
type SplitView struct {
    left  tea.Model
    right tea.Model
    split float64  // 0.0 to 1.0
}

func (s *SplitView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.MouseMsg:
        if msg.Type == tea.MouseLeft {
            // Handle split resize
            s.split = float64(msg.X) / float64(s.width)
        }
    }

    // Update both panels
    leftCmd := s.left.Update(msg)
    rightCmd := s.right.Update(msg)

    return s, tea.Batch(leftCmd, rightCmd)
}
```

### Export Functionality

```go
// Export logs, metrics, or reports
func (d *Dashboard) exportData(format string) tea.Cmd {
    return func() tea.Msg {
        var data []byte
        var filename string

        switch d.activeTab {
        case LogsTab:
            data = d.logs.Export(format)
            filename = fmt.Sprintf("logs_%s.%s",
                time.Now().Format("20060102_150405"), format)

        case MetricsTab:
            data = d.monitoring.Export(format)
            filename = fmt.Sprintf("metrics_%s.%s",
                time.Now().Format("20060102_150405"), format)
        }

        if err := os.WriteFile(filename, data, 0644); err != nil {
            return ErrorMsg{err}
        }

        return ExportedMsg{filename}
    }
}
```

## Next Steps

- Continue to [Deployment →](./17_deployment.md)
- Back to [← MCP Server](./15_mcp_server.md)
- Return to [Summary](./0_summary.md)
