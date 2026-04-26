# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**charms.** is a collection of Charm TUI (Terminal User Interface) applications built with Bubble Tea. It includes 4 interactive games:
- **Chess**: Two-player or vs computer with full clock support and AI engine
- **Tapper**: Classic 1983 arcade game recreated in the terminal
- **Snake**: Portal-wrapped classic snake game with obstacles
- **2048**: Tile-merging puzzle game with undo and configurable targets

All games are accessed through a main lobby interface. Requires Go 1.21+.

## Build & Development

### Compiling the Application

Build the executable:
```bash
go build -o charms .
```

Build with version string (set via ldflags):
```bash
go build -ldflags "-X main.version=v1.0.0" -o charms .
```

Run the compiled binary:
```bash
./charms
```

### Running Tests

Run all tests across all packages:
```bash
go test ./...
```

Run tests for a specific game/package:
```bash
go test ./chess/...
go test ./gitlog/...
```

Run a single test:
```bash
go test ./gitlog -run TestNavigation_BasicMovement
```

Run tests with verbose output:
```bash
go test -v ./...
```

Run tests with coverage report:
```bash
go test -cover ./...
```

Generate detailed coverage HTML report for a package:
```bash
go test -coverprofile=coverage.out ./gitlog
go tool cover -html=coverage.out
```

### Code Quality

Format all Go code:
```bash
go fmt ./...
```

Check for common issues:
```bash
go vet ./...
```

## Architecture

### High-Level Structure

```
charms./
├── main.go                    # Entry point: lobby model and game selection
├── chess/                     # Chess game package (~24 files)
├── tapper/                    # Tapper arcade game package
├── snake/                     # Snake game package
├── game2048/                  # 2048 puzzle game package
├── go.mod / go.sum            # Dependencies
└── .goreleaser.yaml           # GoReleaser configuration
```

### Game Package Structure

Each game package follows this pattern:
- **Main file** (e.g., `chess.go`, `tapper.go`): Model and Update/View implementations for Bubble Tea
- **Logic files**: Game-specific logic split into focused modules (e.g., `engine.go` for AI, `clock.go` for timers)
- **Test files**: One `*_test.go` per logic file, plus integration tests
- **Utilities**: Helper functions and data structures (e.g., `opening.go`, `material.go` in chess)

**Entry Point**: Each game exposes a `Run()` function that initializes and runs a Bubble Tea program.


### Dependencies

Core Bubble Tea stack (see `go.mod`):
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling and layout
- `github.com/charmbracelet/colorprofile` - Color detection
- `github.com/notnil/chess` - Chess move validation and notation

All other dependencies are indirect (for color handling, input processing, terminal compatibility).

## Testing

### Test Organization

The test suite includes comprehensive tests across all game packages:

**Main Test Files**:
- `main_test.go` - Lobby and game selection tests
- `chess/main_test.go` - Comprehensive chess game tests
- `tapper/*_test.go` - Tapper game tests
- `snake/*_test.go` - Snake game tests
- `game2048/*_test.go` - 2048 game tests

### Test Patterns

**Table-Driven Tests**: Tests with multiple input scenarios:
```go
tests := []struct {
    name     string
    input    string
    expected string
}{
    {"scenario 1", "input1", "expected1"},
    {"scenario 2", "input2", "expected2"},
}
```

### Running Tests

See "Running Tests" in the Build & Development section above.

**Useful patterns**:
- Run only chess tests: `go test ./chess/...`
- Run only tests matching a pattern: `go test -run TestChess_* ./chess`
- Run tests for a specific function's coverage: `go test -cover ./chess`

## Code Patterns & Conventions

### Bubble Tea Integration

All games follow the Bubble Tea model pattern:
```go
type model struct {
    // state fields
    cursor int
    data   []Item
}

func (m model) Init() tea.Cmd { /* initialization */ }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { /* state updates */ }
func (m model) View() string { /* render UI */ }
```

**Keybinding handling**: Switch on `km.String()` (tea.KeyMsg) with support for:
- Arrow keys and hjkl for navigation
- vim-style motions (j/k for up/down, g/G for jump to top/bottom)
- Common keys (q to quit, Enter/Space to select, / for search)


### File Organization

- Keep logic separate from rendering (Update vs View)
- One concern per file when possible
- Tests live in the same package (suffix `_test.go`)
- Helpers and utilities in focused utility files
- Large models like `gitlog/engine.go` can contain multiple logical sections with clear comment boundaries

### Import Organization

Group imports by standard library, then external, then internal:
```go
import (
    "fmt"
    "os"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    "charms/gitlog/core"
)
```

## Recent Work & Context

### Extraction of grit

Recently extracted the git history browser from charms. as a standalone application:
- Now lives at `github.com/zpenka/grit`
- 283 comprehensive tests
- 312+ features for git history exploration
- See grit repository for full documentation

## Performance Considerations

### Memory Management

- Use buffering in string building (see `lipgloss.Place` calls)
- Cache large computations where appropriate
- Clean up resources in `Quit` commands

## Common Commands Summary

```bash
# Build and run
go build -o charms . && ./charms

# Testing
go test ./...                           # Run all tests
go test ./chess/... -v                  # Verbose chess tests
go test -run TestChess_* ./chess        # Run specific test
go test -cover ./...                    # Coverage summary
go test -coverprofile=cov.out ./...     # Generate coverage file
go tool cover -html=cov.out             # View coverage in browser

# Code quality
go fmt ./...                            # Format code
go vet ./...                            # Check for issues

# Build with version
go build -ldflags "-X main.version=v1.0.0" -o charms .
```

## Future Considerations

- Performance optimization for games
- Enhanced UI/UX features
- New game variants
- Cross-platform improvements
