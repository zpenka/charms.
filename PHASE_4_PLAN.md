# Phase 4: Testing Reorganization Plan

## Objective
Reorganize and enhance the test suite for better maintainability, coverage, and clarity.

## Current State

### Test Statistics
- **Total Tests:** 565+ across 7 test files
- **Code Coverage:** 73.8%
- **Main Test File:** engine_test.go (545 tests)
- **Other Tests:** render_consolidation_test.go (11), render_migration_test.go (9), core module tests

### Test Categories Identified
1. **KeyBinding** (12 tests) - Keyboard input handling
2. **FilterCommits** (7 tests) - Commit filtering logic
3. **ParseDiff** (6 tests) - Diff parsing
4. **NavigationHistory** (6 tests) - Navigation history tracking
5. **Truncate** (5 tests) - String truncation utility
6. **ParseCount** (5 tests) - Count parsing
7. **FilterCommitsSince** (5 tests) - Time-based filtering
8. **FilterCommitsByAuthor** (5 tests) - Author filtering
9. **Bookmarks** (5 tests) - Bookmark functionality
10. **[20+ more categories]** - Various features

## Phase 4 Goals

### Goal 1: Test File Organization
**Organize tests into logical files by feature/component:**

- `core_tests/` directory
  - parser_test.go - Commit and diff parsing
  - filter_test.go - Filtering and search logic
  - utils_test.go - Utility functions (already exists)
  - types_test.go - Type definitions and helpers

- `engine_tests/` directory
  - navigation_test.go - Movement and cursor control
  - bookmarks_test.go - Bookmark features
  - search_test.go - Search and filter UI
  - clipboard_test.go - Copy/paste functionality
  - history_test.go - Navigation history
  - keybinding_test.go - Keyboard input

- `render_tests/` directory
  - render_consolidation_test.go (already exists)
  - render_migration_test.go (already exists)
  - render_helpers_test.go - Helper function tests
  - render_integration_test.go - Full render pipeline tests

### Goal 2: Improve Test Coverage
**Target areas for increased coverage:**

- Consolidated render functions - currently at ~73.8%
- Error handling paths
- Edge cases in parsing functions
- Integration between components

**Coverage targets:**
- Core module: 85%+
- Render functions: 80%+
- Overall: 80%+

### Goal 3: Enhance Test Documentation
- Add clear comments to test functions
- Document test fixtures and helpers
- Create test data builders for common scenarios
- Document expected behavior for edge cases

### Goal 4: Create Test Helpers Library
**Consolidate common test patterns:**

```go
// testdata.go - Test fixtures and builders
type TestFixture struct {
    Commits []commit
    Model   model
}

func NewTestFixture() TestFixture
func NewTestFixtureWithCommits(count int) TestFixture
func NewTestModel() model
```

**Assertion helpers (already exist in engine_test_helpers.go):**
- AssertEqual, AssertNotEqual
- AssertTrue, AssertFalse
- AssertStringContains, AssertStringNotContains
- Etc.

### Goal 5: Add Integration Tests
**Test complete workflows:**

- Filtering → Searching → Navigation
- Copy → Paste → Undo
- Bookmark → Search → Unbookmark
- Diff viewing → Comment → Save

### Goal 6: Performance Benchmarks
**Add benchmarks for critical paths:**

```go
// Benchmark filtering on large commit sets
BenchmarkFilterCommits_LargeSet
BenchmarkParseCommits_1000Commits
BenchmarkRenderPanel_ComplexData
```

### Goal 7: Test Documentation
**Create test documentation:**

- TEST_GUIDE.md - How to run, organize, and write tests
- Test patterns and best practices
- Coverage goals and tracking
- CI/CD integration guide

## Implementation Plan

### Phase 4a: Test Organization (Current)
1. ✅ Analyze current test structure
2. Create test directory structure
3. Create test helper library improvements
4. Move tests to organized files

### Phase 4b: Coverage Improvement
1. Identify coverage gaps using coverage report
2. Add tests for uncovered paths
3. Improve edge case coverage
4. Target 80%+ coverage

### Phase 4c: Integration Tests
1. Create integration test suite
2. Test multi-step workflows
3. Test error recovery
4. Test state transitions

### Phase 4d: Performance Optimization
1. Add benchmarks
2. Profile hot paths
3. Identify bottlenecks
4. Document performance targets

### Phase 4e: Documentation
1. Create TEST_GUIDE.md
2. Document test patterns
3. Add test data builders
4. Document coverage goals

## Test Organization Structure (Proposed)

```
gitlog/
├── core/
│   ├── filter.go
│   ├── filter_test.go (improved)
│   ├── parser.go
│   ├── parser_test.go (new)
│   ├── utils.go
│   └── utils_test.go
│
├── engine.go
├── engine_render_consolidation.go
│
├── test_helpers/
│   ├── assertions.go
│   ├── fixtures.go
│   ├── builders.go
│   └── mocks.go
│
└── tests/
    ├── navigation_test.go
    ├── bookmarks_test.go
    ├── search_test.go
    ├── render_consolidation_test.go
    ├── render_integration_test.go
    └── integration_test.go
```

## Success Metrics

- [ ] Tests organized into logical files (0 blockers in any file)
- [ ] 80%+ code coverage
- [ ] All integration tests passing
- [ ] Performance benchmarks established
- [ ] Test documentation complete
- [ ] No flaky tests
- [ ] CI/CD fully integrated

## Next Steps

1. **Create test helper library** - Consolidate and improve test utilities
2. **Analyze coverage** - Identify gaps and priority areas
3. **Reorganize tests** - Move to organized structure
4. **Add missing tests** - Improve coverage to 80%+
5. **Create integration tests** - Test workflows
6. **Add benchmarks** - Establish performance baselines
7. **Document** - Create comprehensive test guide
