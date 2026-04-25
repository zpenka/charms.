# Git Log Architecture & Implementation Guide

## Project Overview

A comprehensive terminal UI for browsing git history with 113 integrated features organized across 6 major categories.

### Statistics
- **Total Features**: 113
- **Test Coverage**: 371+ tests
- **Type Definitions**: 100+ types  
- **Functions**: 200+ functions
- **Lines of Code**: 8,200+ lines

## Architecture Layers

### Layer 1: Core Types (engine.go, engine_types.go)
- `commit`: Core commit representation
- `model`: Central state management (200+ fields)
- `diffLine`, `lineKind`: Diff parsing & representation
- `panel`: UI panel management

### Layer 2: Feature Types
Organized by category with dedicated types:
- **Bisect & Recovery**: `bisectState`, `lostCommit`, `conflictInfo`
- **Code Quality**: `codeOwnershipData`, `hotspotData`, `commitMetrics`
- **Analysis**: `semanticSearchResult`, `authorActivityData`, `mergeAnalysis`
- **Workflows**: `worktreeInfo`, `namedStash`, `tagOperation`
- **AI Insights**: `commitClassification`, `anomalyData`, `similarCommit`
- **Performance**: `repoLoadState`, `indexData`, `memoryMetrics`

### Layer 3: Core Functions (engine.go)
**Parsing Functions**:
- `parseCommits()` - Parse git log output
- `parseDiff()` - Parse unified diffs
- `parseFileItems()` - Extract file changes

**Navigation Functions**:
- `moveCursorDown()`, `moveCursorUp()` - Vertical navigation
- `switchPanel()` - UI panel switching
- `scrollDiffDown()`, `scrollDiffUp()` - Diff scrolling

**Filtering Functions**:
- `filterCommits()` - Search/filter commits
- `visibleCommits()` - Apply all active filters

### Layer 4: Feature Functions (engine.go)
Organized by feature area:

**Bisect & Recovery** (5 functions):
- `initiateBisect()` - Start bisect workflow
- `bisectMarkGood/Bad()` - Mark commits
- `bisectFindCulprit()` - Find buggy commit

**Code Quality** (10 functions):
- `analyzeCodeOwnership()` - Identify code owners
- `detectHotspots()` - Find frequently-changed areas
- `lintCommitMessage()` - Validate message quality

**Analysis** (15 functions):
- `semanticSearch()` - Find commits by content
- `buildActivityHeatmap()` - Track author timing
- `analyzeMerges()` - Classify merge types

**Team & Collaboration** (8 functions):
- `calculateTeamStats()` - Team metrics
- `suggestReviewers()` - Reviewer recommendations
- `detectPairProgramming()` - Find pair sessions

**AI Insights** (8 functions):
- `autoCompleteMessage()` - Message suggestions
- `classifyCommit()` - Categorize commits
- `detectAnomalies()` - Find unusual patterns

**Compliance & Security** (8 functions):
- `checkSigningCompliance()` - GPG verification
- `scanForSecurityIssues()` - Find exposed secrets
- `detectSecrets()` - Secret scanning

**Release & Versioning** (8 functions):
- `detectSemver()` - Find version tags
- `generateChangelog()` - Create changelog
- `trackVersionBumps()` - Record version changes

**Performance** (8 functions):
- `incrementalLoadRepository()` - Progressive loading
- `parallelProcessDiffs()` - Concurrent processing
- `lazyLoadBlame()` - On-demand blame loading

### Layer 5: UI Rendering (engine.go, engine_render_consolidation.go)
**Consolidated Rendering**:
- `RenderStandardUI()` - Template for standard UI
- `RenderAnalysisUI()` - Analysis data display
- `RenderDataGrid()` - Tabular output
- `RenderMetricBar()` - Visual metrics
- `RenderComparisonTable()` - Side-by-side data

**Feature-Specific Rendering** (30+ functions):
- All `render*UI()` functions follow consistent pattern
- Each feature has dedicated rendering function
- UI output through model fields

### Layer 6: Utilities & Optimization (engine_render_consolidation.go, engine_optimization.go)
**Rendering Config**:
- `RenderConfig` struct for consistent UI

**Performance Patterns**:
- `CacheMetrics` - Cache performance tracking
- `LazyLoader` - Deferred initialization
- `MemoryPool` - Object pooling
- `CircularBuffer` - Fixed-size ring buffer
- `RateLimiter` - Operation rate limiting
- `BatchProcessor` - Bulk operation handling
- `Metrics` - General performance metrics

### Layer 7: Testing (engine_test.go, engine_test_helpers.go)
**Test Organization** (371+ tests):
- Grouped by feature category
- Core parsing tests (30+ tests)
- Navigation & filtering (20+ tests)
- Feature-specific tests (250+ tests)

**Test Helpers**:
- `TestFixture` - Reusable test data
- `Assert*` functions - Assertion helpers
- `TestCategory` - Test organization

## State Management

### Model Struct (200+ fields)
Central state holder in `model`:

**Navigation State**:
- `cursor` - Current position in commit list
- `focus` - Active panel
- `diffOffset` - Diff scrolling position

**Filter State**:
- `query` - Search query
- `authorFilter` - Author filter
- `sinceFilter` - Time-based filter
- `extensionFilters` - File type filters

**Feature State**:
- `show*` boolean flags (50+) - Feature visibility
- `*Data` fields - Feature-specific data
- `*History` fields - Historical data for trends

**Cache State**:
- `dcache` - Diff line cache
- `scache` - Statistics cache
- `recache` - Regex pattern cache

## Data Flow

### Commit Loading
```
git log output
  ↓
parseCommits() → []commit
  ↓
filterCommits() + visibleCommits() → filtered commits
  ↓
model.commits (state)
  ↓
Render functions → UI output
```

### Diff Processing
```
git show --stat --patch
  ↓
parseDiff() → []diffLine
  ↓
dcache (memoization)
  ↓
renderDiffPanel() → colored output
```

### Feature Computation
```
model.commits (input)
  ↓
Feature function (e.g., analyzeCodeOwnership)
  ↓
Typed result (e.g., map[string]codeOwnershipData)
  ↓
model.codeOwnership (state)
  ↓
render*UI() → formatted output
```

## Caching Strategy

### Diff Cache (LRU)
- Stores parsed `[]diffLine` by commit hash
- Configured maximum size
- Tracks hit rate for optimization
- Automatic eviction of oldest entries

### Statistics Cache (LRU)
- Caches `commitStatistics` (files, additions, deletions)
- Avoids redundant computation
- Hit rate tracking

### Regex Cache
- Compiles and caches `*regexp.Regexp` patterns
- First-compile optimization
- Improves search performance

## Performance Optimizations

### Lazy Loading
- Diffs loaded on demand (not on startup)
- Feature data computed when activated
- Graph visualization built on demand

### Incremental Loading
- Large repos load commits progressively
- Non-blocking UI during loading
- Progress percentage display

### Parallel Processing
- Diffs processed concurrently
- Multiple commits analyzed in parallel
- Significant speedup for large histories

### Memory Management
- Fixed-size caches prevent unbounded growth
- Object pooling reduces allocations
- Circular buffers for fixed-window data
- Lazy initialization for features

## Testing Strategy

### Test Organization (5 categories)
1. **Core Tests** (30+) - Parsing, navigation, filtering
2. **Feature Tests** (300+) - One per feature implementation
3. **Integration Tests** - Multi-feature interactions
4. **Performance Tests** - Cache & optimization verification
5. **Regression Tests** - Prevent known bugs

### Test Pattern
```go
func Test<Feature>_<Scenario>(t *testing.T) {
    // Setup
    m := model{...}
    
    // Execute
    result := feature(m)
    
    // Assert
    if result == nil {
        t.Error("expected result")
    }
}
```

## Extending the System

### Adding a New Feature
1. Define type in feature category
2. Add model fields with `show*` flag
3. Create feature function with TDD tests
4. Create `render*UI()` function
5. Add keybinding to `handleKeyBinding()`
6. Update README.md documentation

### Example: Code Hotspots
```go
// 1. Type definition
type hotspotData struct {
    path string
    changeFrequency int
    riskLevel string
}

// 2. Model fields
type model struct {
    hotspots []hotspotData
    showHotspots bool
}

// 3. Feature function
func detectHotspots(commits []commit) []hotspotData {
    // Implementation
}

// 4. Render function
func renderHotspotsUI(m model, width int) string {
    // Implementation
}

// 5. Keybinding
case "H":
    m.showHotspots = !m.showHotspots

// 6. Tests
func TestDetectHotspots_FindsFrequent(t *testing.T) {
    // Implementation
}
```

## Configuration & Customization

### Model Initialization
```go
m := newModel(repoPath)  // Default state
```

### Feature Activation
```go
m.showAnalytics = true   // Show analytics panel
m = handleKeyBinding(m, "A")  // Toggle via keybinding
```

### Cache Configuration
```go
dcache := &diffCache{maxSize: 100}
scache := &statCache{maxSize: 200}
recache := &regexCache{maxSize: 50}
```

## Dependencies & Imports

### Standard Library
- `fmt` - String formatting
- `regexp` - Pattern matching
- `strconv` - Type conversion
- `strings` - String manipulation
- `time` - Time operations

### External
- None (zero dependencies)

## File Structure

```
gitlog/
├── engine.go                          (Main: 3500+ lines)
├── engine_render_consolidation.go     (UI: 200 lines)
├── engine_optimization.go              (Performance: 250 lines)
├── engine_test_helpers.go              (Tests: 150 lines)
├── engine_test.go                      (Tests: 3970 lines)
├── gitlog.go                           (Package main)
├── ARCHITECTURE.md                     (This file)
├── REFACTORING.md                      (Refactoring plan)
└── README.md                           (User documentation)
```

## Summary

The git log browser is architected as a layered system with clear separation between:
- **Type definitions** (100+ types for 113 features)
- **Core functions** (200+ functions for operations)
- **Rendering** (30+ functions for UI)
- **Optimization** (caching, lazy loading, parallelization)
- **Testing** (371 tests covering all features)

The model struct serves as the central state hub, with features adding fields to support their functionality. All code follows consistent patterns for extensibility and maintainability.
