# Phase 4c: Test Suite Reorganization Strategy

## Overview
Phase 4c reorganizes 411+ tests from the monolithic `engine_test.go` (4290 lines) into focused feature-specific test files, improving maintainability and test discoverability.

## Current State
- **Total Tests**: 411 test functions  
- **Current File**: engine_test.go (4290 lines)
- **Helper Functions**: Extracted to engine_test_minimal.go (46 lines)
  - `makeCommits(n int) []commit`
  - `makeDiffLines(n int) []diffLine`
  - `makeNamedCommits() []commit`
  - `makeCommitsWithDays() []commit`

## Recommended Organization Structure

### Phase 4c-1: Core Infrastructure Tests (Priority 1)
Essential, foundational tests that other tests may depend on:

#### parsing_test.go (~55 tests)
Tests for all input parsing functions:
- ParseCommits, ParseDiff, ParseBlameLine, ParseBranches
- ParseCurrentBranch, ParseBlame, ParseCount, ParseFileItems
- ParseGitReferences, ParseHunks, ParseDateRange, ParseTags
- ParseStashList, ParseReflog, ParseRebaseSequence, ParseSubmodules
- DiffParsing, SafeParseCommitGraph, ParseCommitGraph

**Rationale**: Parsing is foundational; all other tests depend on these functions working correctly.

#### filtering_test.go (~33 tests)
Tests for commit filtering and selection:
- FilterCommits (7 tests) - Main filtering function
- FilterCommitsByAuthor (5 tests) - Author-based filtering
- FilterCommitsSince (5 tests) - Time-based filtering
- FilterCommitsByFile, FilterByRegex, VisibleCommits
- FormatActiveFilters, FilterByDateRange, FilterCombined
- FilterByFilePattern, FilterByExtension, FilterCommitsByFileChange

**Rationale**: Filtering is used by UI and analysis features; critical for functionality.

#### navigation_test.go (~44 tests)
Tests for UI navigation and cursor control:
- KeyBinding (12 tests) - Keyboard input handling
- MoveCursor (6 tests) - Cursor up/down navigation
- ScrollDiff (8 tests) - Diff panel scrolling
- SwitchPanel (2 tests) - Panel focus switching
- ListPanelWidth, DiffPanelWidth, DiffPanelHeight (6 tests) - Panel sizing
- MiniMapPosition (3 tests) - Minimap position calculation
- ToggleFileView, ToggleBranchView (4 tests) - View toggles
- NavigationHistory (6 tests) - Navigation history tracking

**Rationale**: UI navigation is core to user interaction; frequently modified.

#### utilities_test.go (~22 tests)
Tests for utility and helper functions:
- Truncate (5 tests) - String truncation with ellipsis
- FirstWord (2 tests) - Word extraction
- CopyAsPatch, ExtractReviewers, ExtractCoAuthors (6 tests)
- DiffCache, RegexCache, CircularBuffer (6 tests) - Caching/performance
- CurrentFile (1 test) - Current file tracking

**Rationale**: Utilities are reused across codebase; good candidates for early extraction.

### Phase 4c-2: Feature Tests (Priority 2)
Feature-specific tests that depend on core infrastructure:

#### keybinding_test.go (~12 tests)
Keyboard binding and shortcut tests (already in navigation_test.go grouping)

#### bookmarks_test.go (~5 tests)
Bookmark management tests

#### rendering_test.go (~50+ tests)
UI rendering function tests (RenderUI variants, RenderBookmarkMarker, etc.)

### Phase 4c-3: Integration & Analysis Tests (Priority 3)
Complex tests that combine multiple components:

#### git_integration_test.go (~37 tests)
- CommitStats, AuthorStats - Statistics calculation
- GetMergeParents, IsMergeCommit - Merge detection
- NavigateAlongGraph, BuildFileHistory - Graph operations
- CurrentFile, BisectFindCulprit - State tracking

#### commit_operations_test.go (~27 tests)
- GenerateCommitMessage, GoToCommit, ResetToCommit
- AmendLastCommit, SelectForCherryPick
- PreviewRebase, RevertCommit, SquashCommit

#### analysis_test.go (~80+ tests)
Complex analysis and calculation functions:
- CalculateBisectProgress, CalculateVelocity, CalculateProductivity
- DetectLanguage, DetectAnomalies, DetectHotspots
- IdentifyUncoveredChanges, AnalyzeCodeOwnership
- TimeBasedStats, AuthorActivityHeatmap

### Phase 4c-4: Advanced & Miscellaneous (Priority 4)
Remaining tests for specialized features:

#### advanced_workflows_test.go (~40 tests)
- WorkflowTemplate, CreateNamedStash, BuildTimeline
- BuildFlameGraph, BuildDependencyGraph
- Semantic search and specialized features

#### miscellaneous_test.go (~25 tests)
- Edge cases and specialized utility tests
- Tests that don't fit other categories

## Implementation Strategy

### Step 1: Preparation
- [ ] Review this plan with team
- [ ] Identify any test interdependencies
- [ ] Plan test helper consolidation

### Step 2: Core Files (Phase 4c-1)
Create in this order (each can be verified independently):
1. **parsing_test.go** - No dependencies on other tests
2. **utilities_test.go** - No dependencies on other tests
3. **filtering_test.go** - Depends on parsing (OK, both in Phase 1)
4. **navigation_test.go** - Depends on test fixtures

### Step 3: Dependent Files (Phase 4c-2)
5. **keybinding_test.go** - Depends on navigation
6. **bookmarks_test.go** - Self-contained
7. **rendering_test.go** - Depends on parsing + utilities

### Step 4: Complex Files (Phase 4c-3)
8. **git_integration_test.go** - Depends on parsing + filtering
9. **commit_operations_test.go** - Depends on git operations
10. **analysis_test.go** - Depends on all of above

### Step 5: Finalization (Phase 4c-4)
11. **advanced_workflows_test.go** - Advanced features
12. **miscellaneous_test.go** - Cleanup
13. **engine_test.go** - Retain only shared helpers (if not consolidated)

## Migration Approach

### Option A: Parallel Testing (Recommended)
1. Create new test files with tests + helpers
2. Keep engine_test.go intact during development
3. Run both simultaneously with `go test ./gitlog`
4. Remove engine_test.go tests once all files verified
5. One atomic commit with all reorganized files

### Option B: Sequential Migration
1. Create new test files one category at a time
2. Remove extracted tests from engine_test.go
3. Verify tests after each extraction
4. Multiple commits, one per file

### Option C: Consolidated Helpers
1. Extract all helper functions to shared file (if not using engine_test_minimal.go)
2. Reference from all test files via imports
3. Reduce duplication of helper code

## Verification Steps

For each new test file:
```bash
# Compile check
go test ./gitlog -v -run TestFilePattern

# Full suite compatibility
go test ./gitlog -v

# Coverage verification
go test ./gitlog -cover
```

## File Organization Summary

| File | Tests | Dependency | Phase |
|------|-------|-----------|-------|
| parsing_test.go | ~55 | None | 1 |
| filtering_test.go | ~33 | Parsing | 1 |
| navigation_test.go | ~44 | Fixtures | 1 |
| utilities_test.go | ~22 | None | 1 |
| keybinding_test.go | ~12 | Navigation | 2 |
| bookmarks_test.go | ~5 | None | 2 |
| rendering_test.go | ~50 | Parsing+Utils | 2 |
| git_integration_test.go | ~37 | Parsing+Filter | 3 |
| commit_operations_test.go | ~27 | Git Ops | 3 |
| analysis_test.go | ~80 | All Core | 3 |
| advanced_workflows_test.go | ~40 | Advanced | 4 |
| miscellaneous_test.go | ~25 | Various | 4 |

## Benefits of Reorganization

1. **Improved Discoverability**: Tests organized by feature/function
2. **Better Maintainability**: Smaller files easier to navigate
3. **Faster CI**: Parallel test runs for independent files
4. **Easier Refactoring**: Clear test boundaries simplify code changes
5. **Reduced Cognitive Load**: Focused test files easier to understand
6. **Better Documentation**: File structure documents codebase organization

## Rollback Plan

If issues arise:
1. Revert to engine_test.go
2. All tests still pass (no functional changes)
3. Zero risk to production code

## Notes

- Helper functions can be consolidated in engine_test_minimal.go or duplicated per file
- Consider using shared helpers to reduce duplication (4 helpers × 12 files = significant duplication)
- engine_test_helpers.go already contains TestFixture and assertion helpers (separate from extraction helpers)
- Total estimated effort: 2-4 hours for full implementation
- Low risk: Only test organization changes, no production code modifications
