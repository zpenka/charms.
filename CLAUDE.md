# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**charms.** is a collection of Charm TUI (Terminal User Interface) applications built with Bubble Tea. It includes 5 interactive games:
- **Chess**: Two-player or vs computer with full clock support and AI engine
- **Tapper**: Classic 1983 arcade game recreated in the terminal
- **Snake**: Portal-wrapped classic snake game with obstacles
- **2048**: Tile-merging puzzle game with undo and configurable targets
- **Git Log**: Production-grade git history browser with 312+ features, advanced analytics, and ML/AI capabilities

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
├── gitlog/                    # Git log browser package (most complex)
│   ├── core/                  # Core data structures and utilities
│   ├── engine.go              # Main model logic (~158K LOC)
│   ├── gitlog.go              # TUI rendering and main interface
│   └── [multiple test files]  # 600+ comprehensive tests
├── go.mod / go.sum            # Dependencies
├── .goreleaser.yaml           # GoReleaser configuration
└── TEST_GUIDE.md              # Comprehensive testing documentation
```

### Game Package Structure

Each game package follows this pattern:
- **Main file** (e.g., `chess.go`, `tapper.go`): Model and Update/View implementations for Bubble Tea
- **Logic files**: Game-specific logic split into focused modules (e.g., `engine.go` for AI, `clock.go` for timers)
- **Test files**: One `*_test.go` per logic file, plus integration tests
- **Utilities**: Helper functions and data structures (e.g., `opening.go`, `material.go` in chess)

**Entry Point**: Each game exposes a `Run()` function that initializes and runs a Bubble Tea program.

### Git Log Browser Architecture

The Git Log browser is the most complex component with production-grade features:

**Core Module** (`gitlog/core/`):
- `types.go`: Core data structures (Commit, Filter, Analysis types)
- `filter.go`: Filtering logic (author, time, text search, regex)
- `utils.go`: Utility functions and helpers
- `parser.go`: Parsing utilities for git output
- Tests for each module with comprehensive coverage

**Engine** (`gitlog/engine.go`):
- 158K LOC containing the model state machine
- 300+ functions implementing features across 9 categories:
  - Navigation (commit traversal, searching, bookmarks)
  - Filtering (author, date, extension, text)
  - Analysis (hotspots, code ownership, complexity)
  - Operations (rebase, cherry-pick, amend, bisect)
  - Collaboration (team velocity, pair programming detection)
  - AI features (classification, anomaly detection, summaries)
  - Compliance (signing, secrets scanning)
  - Release (semantic versioning, changelog generation)
  - Export (markdown, patches, JIRA linking)

**UI** (`gitlog/gitlog.go`):
- Rendering logic using lipgloss for styling
- Diff panel, commit list, branch picker, file list, blame view
- Status bar and help text
- Input handling for all keybindings

### Dependencies

Core Bubble Tea stack (see `go.mod`):
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling and layout
- `github.com/charmbracelet/colorprofile` - Color detection
- `github.com/notnil/chess` - Chess move validation and notation

All other dependencies are indirect (for color handling, input processing, terminal compatibility).

## Testing

### Test Organization

The test suite has 600+ tests with 74.7% code coverage (target: 80%+):

**Main Test Files**:
- `main_test.go` - Lobby and game selection tests
- `chess/main_test.go` - Comprehensive chess game tests
- `gitlog/engine_test.go` - Main engine tests (545+ tests)
- `gitlog/parsing_test.go` - Git parsing tests
- `gitlog/navigation_test.go` - Navigation and UI tests
- `gitlog/filtering_test.go` - Filter logic tests
- `gitlog/core/*_test.go` - Core module tests
- `gitlog/integration_test.go` - End-to-end workflow tests
- `gitlog/coverage_gaps_test.go` - Tests targeting low-coverage functions
- `gitlog/*_test.go` - Feature-specific tests

### Test Patterns

**CommitBuilder Pattern** (for gitlog tests):
```go
commit := NewCommitBuilder()
    .WithHash("abc123")
    .WithAuthor("John Doe")
    .WithSubject("Test feature X")
    .WithWhen("2 hours ago")
    .WithBody("Detailed description")
    .Build()
```

**Assertion Helpers** (see `gitlog/engine_test_helpers.go`):
- `requireCommitsEqual()` - Compare commits
- `requireDiffEqual()` - Compare diffs
- `requireAnalysisDataEqual()` - Compare analysis output

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
- Run only gitlog tests: `go test ./gitlog/...`
- Run only tests matching a pattern: `go test -run TestNavigation ./gitlog`
- Run tests for a specific function's coverage: `go test -cover ./gitlog`

### Coverage Goals

**Low-Coverage Targets** (from TEST_GUIDE.md):
- `handleKeyBinding` - Target: 30%+ (current: 23.7%)
- `renderFileTimeline` - Target: 40%+ (current: 33.3%)
- `calculateBisectProgress` - Target: 45%+ (current: 37.5%)
- `classifyCommit` - Target: 50%+ (current: 40%)
- Model initialization - Target: 10%+ (current: 0%)

When writing new tests, focus on:
1. Gap-filling (target the low-coverage functions above)
2. Edge cases (boundary conditions, empty inputs, malformed data)
3. Integration workflows (multi-step user interactions)
4. Error paths (invalid inputs, system failures)

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

### Git Integration (Git Log Browser)

The git log browser uses the `git` CLI via `os/exec`:
- `git log --oneline` and `git log --name-status` for commit listing
- `git show` for diff viewing
- `git blame` for blame view
- `git branch -a` for branch listing
- `git checkout` for branch switching

All git command output is parsed into internal data structures (see `gitlog/core/types.go`).

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

### Phase 4: Test Suite Reorganization (Latest)

Recently completed improvements to the test suite:
- Added 35+ new tests (600+ total)
- Improved coverage from 73.4% → 74.7%
- Created `CommitBuilder` pattern for cleaner test fixtures
- Added 12 gap-filling tests targeting low-coverage functions
- Created 10 integration tests for complete workflows
- Generated comprehensive `TEST_GUIDE.md`

Key files from Phase 4:
- `PHASE_4_SUMMARY.md` - Overview of Phase 4 work
- `TEST_GUIDE.md` - Complete testing reference
- `gitlog/engine_test_helpers.go` - CommitBuilder and helpers

### Git Log Browser Features

The git log browser has evolved significantly and now includes:
- 312+ integrated features across 9 categories
- Advanced filtering (author, date, extension, regex)
- Analytics (hotspots, code ownership, complexity)
- AI features (classification, anomaly detection, summaries)
- Team collaboration (velocity, pair programming)
- Security (signing compliance, secrets detection)
- Release management (changelog generation, version tracking)
- Export capabilities (markdown, patches, JIRA linking)

See `README.md` for complete feature documentation.

## Performance Considerations

### Git Log Browser Optimization

For the git log browser (handling large repositories):
- **Lazy loading**: Diffs load asynchronously, up to 200 commits on startup
- **Caching**: Diff caching (LRU), statistics memoization, regex compilation caching
- **Performance metrics**: Track cache hits for optimization tuning
- **Large repo support**: Incremental loading and parallel diff processing ready

### Memory Management

- Use buffering in string building (see `lipgloss.Place` calls)
- Cache large computations (see gitlog cache implementations)
- Clean up resources in `Quit` commands

## Common Commands Summary

```bash
# Build and run
go build -o charms . && ./charms

# Testing
go test ./...                           # Run all tests
go test ./gitlog/... -v                 # Verbose gitlog tests
go test -run TestNavigation ./gitlog    # Run specific test
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

From `PHASE_4_PLAN.md`:
- Continue targeting 80%+ coverage (currently 74.7%)
- Focus gap-filling tests on identified low-coverage functions
- Consider UI integration tests for game rendering logic
- Build out integration test suite for complex workflows
