# Test Guide - Git Log Browser

Comprehensive guide for testing, writing tests, and improving test coverage.

## Overview

- **Total Tests**: 600+
- **Code Coverage**: 74.7% (Target: 80%+)
- **Test Files**: 7
- **Test Framework**: Go `testing` package

## Test Files Organization

### Core Module Tests

- **core/core_test.go** - Core type and utility tests
- **core/filter_test.go** - Filtering and search logic
- **core/utils_test.go** - Utility function tests

### Engine Tests

- **engine_test.go** (545 tests) - Main test file (to be reorganized)
  - Navigation tests
  - Filtering tests
  - Parsing tests
  - Bookmark tests
  - And 20+ more categories

- **engine_test_helpers.go** - Assertion helpers and test fixtures
  - TestFixture struct
  - Assert* functions
  - CommitBuilder pattern

- **coverage_gaps_test.go** - Tests for low-coverage functions
  - Tests for functions with < 50% coverage
  - Improving coverage toward 80%

### Render Tests

- **render_consolidation_test.go** - Consolidation template tests
  - RenderStandardUI tests
  - RenderAnalysisUI tests
  - RenderDataGrid tests
  - RenderMetricBar tests
  - RenderSummaryStats tests
  - RenderErrorList tests
  - RenderComparisonTable tests

- **render_migration_test.go** - Migration pattern tests
  - Shows how to migrate from old patterns
  - Documents refactoring patterns

### Integration Tests

- **integration_test.go** - Complete workflow tests
  - Search and filtering workflows
  - Navigation after filtering
  - Bookmarking workflows
  - Multi-component interactions

## Running Tests

### Run All Tests

```bash
go test ./gitlog -v
```

### Run Specific Test

```bash
go test ./gitlog -run TestFilterCommits -v
```

### Run Tests with Coverage

```bash
go test ./gitlog -coverprofile=coverage.out
go tool cover -html=coverage.out  # View in browser
```

### Run Tests by Category

```bash
# Navigation tests
go test ./gitlog -run "Test.*Navigation" -v

# Filter tests
go test ./gitlog -run "Test.*Filter" -v

# Render tests
go test ./gitlog -run "Test.*Render" -v
```

## Test Patterns

### 1. Basic Unit Test

```go
func TestFunctionName_Behavior(t *testing.T) {
    // Arrange
    fixture := NewTestFixture()
    
    // Act
    result := someFunction(fixture.Commits)
    
    // Assert
    AssertTrue(t, len(result) > 0, "should return results")
}
```

### 2. Using Test Fixture

```go
func TestWithFixture(t *testing.T) {
    fixture := NewTestFixture()
    
    // fixture.Commits - 5 sample commits
    // fixture.CommitCount - 5
    // fixture.AuthorCount - 4
    // fixture.SampleFiles - common file names
    // fixture.SampleAuthors - ["Alice", "Bob", ...]
}
```

### 3. Using CommitBuilder

```go
func TestWithBuilder(t *testing.T) {
    commit := NewCommitBuilder().
        WithHash("abc123def456").
        WithAuthor("John Doe").
        WithSubject("Test Subject").
        WithWhen("1 hour ago").
        Build()
    
    AssertEqual(t, "John Doe", commit.author, "author should match")
}
```

### 4. Integration Test

```go
func TestWorkflow_SearchAndFilter(t *testing.T) {
    // Step 1: Setup
    fixture := NewTestFixture()
    
    // Step 2: Execute workflow
    filtered := filterCommits(fixture.Commits, "feature")
    
    // Step 3: Verify integration
    AssertTrue(t, len(filtered) > 0, "should find results")
}
```

## Assertion Functions

### Basic Assertions

- `AssertEqual(t, expected, actual, message)` - Check equality
- `AssertNotEqual(t, a, b, message)` - Check inequality
- `AssertTrue(t, condition, message)` - Check boolean true
- `AssertFalse(t, condition, message)` - Check boolean false
- `AssertNil(t, value, message)` - Check nil
- `AssertNotNil(t, value, message)` - Check not nil

### String Assertions

- `AssertStringContains(t, str, substr, message)` - Check substring
- `AssertStringNotContains(t, str, substr, message)` - Check not substring

### Collection Assertions

- `AssertLen(t, items, expectedLen, message)` - Check length
- `AssertSliceContains(t, items, item, message)` - Check slice contains value
- `AssertMapContains(t, m, key, message)` - Check map contains key

### Range Assertions

- `AssertIntRange(t, value, min, max, message)` - Check int in range
- `AssertFloatRange(t, value, min, max, message)` - Check float in range

## Test Coverage Goals

### Current Status
- **Overall**: 74.7%
- **Core Module**: ~85%
- **Engine**: ~73%
- **Render**: ~75%

### Target Coverage by Component

| Component | Current | Target |
|-----------|---------|--------|
| Core Parsing | 95% | 95%+ |
| Core Filtering | 90% | 95%+ |
| Navigation | 85% | 90%+ |
| Rendering | 75% | 80%+ |
| Keybindings | 50% | 70%+ |
| Analytics | 70% | 80%+ |
| Overall | 74.7% | 80%+ |

## Low Coverage Functions (Priority Targets)

### Critical (< 50% coverage)

- `handleKeyBinding` (23.7%) - Keyboard input dispatcher
- `renderFileTimeline` (33.3%) - File history rendering
- `calculateBisectProgress` (37.5%) - Bisect progress calculation
- `classifyCommit` (40.0%) - Commit classification

### Important (50-70% coverage)

- `addToNavHistory` (80.0%) - Navigation history
- `parseFileItemsFromDiff` (90.0%) - File parsing from diff

### Helpers (0% coverage - often not directly tested)

- `newModel` - Model initialization
- `BuildAnalysisData` - Data builder helper
- `CommitBuilder` methods - Builder pattern

## Best Practices

### 1. Test Naming

- Start with `Test`
- Include function name: `TestFunctionName`
- Include behavior: `TestFilterCommits_ByAuthor`
- Use underscores to separate: `TestFunctionName_BehaviorDescription`

### 2. Clear Assertions

```go
// Good
AssertEqual(t, 5, len(results), "should return 5 results")

// Bad
AssertEqual(t, 5, len(results), "results")
```

### 3. Test Organization

- Group related tests together
- Use `// Test category` comments
- Keep tests focused (one behavior per test)
- Use table-driven tests for multiple scenarios

### 4. Setup and Cleanup

```go
func TestSomething(t *testing.T) {
    // Setup
    fixture := NewTestFixture()
    
    // Test
    result := doSomething(fixture)
    
    // Assert
    AssertTrue(t, result, "should work")
    
    // Cleanup (if needed)
    // t.Cleanup(func() { /* cleanup */ })
}
```

### 5. Test Data

- Use `NewTestFixture()` for standard test data
- Use `NewCommitBuilder()` for custom commits
- Keep test data small and focused
- Use meaningful names (Alice, Bob, Charlie)

## Adding New Tests

### Step 1: Identify Coverage Gap

```bash
go test ./gitlog -coverprofile=coverage.out
go tool cover -func=coverage.out | grep "0\.0%"
```

### Step 2: Create Test

```go
// Place in appropriate test file
func TestNewFunction_Behavior(t *testing.T) {
    fixture := NewTestFixture()
    result := newFunction(fixture.Commits)
    AssertTrue(t, result != nil, "should not be nil")
}
```

### Step 3: Run Test

```bash
go test ./gitlog -run TestNewFunction -v
```

### Step 4: Check Coverage

```bash
go test ./gitlog -coverprofile=coverage.out
go tool cover -func=coverage.out | grep NewFunction
```

## Common Test Scenarios

### Scenario 1: Testing Filtering

```go
func TestFilter_FindsMatches(t *testing.T) {
    commits := []commit{
        {subject: "Add feature", author: "Alice"},
        {subject: "Fix bug", author: "Bob"},
    }
    
    result := filterCommits(commits, "feature")
    AssertEqual(t, 1, len(result), "should find feature")
}
```

### Scenario 2: Testing Navigation

```go
func TestNavigation_UpdatesCursor(t *testing.T) {
    m := model{cursor: 0, commits: fixture.Commits}
    
    m = moveCursorDown(m)
    AssertTrue(t, m.cursor >= 0, "cursor should be valid")
}
```

### Scenario 3: Testing Rendering

```go
func TestRender_ProducesOutput(t *testing.T) {
    config := RenderConfig{Title: "Test", Items: []string{"a", "b"}}
    result := RenderStandardUI(config)
    
    AssertStringContains(t, result, "Test", "should have title")
    AssertStringContains(t, result, "a", "should have items")
}
```

## Debugging Tests

### Run Single Test with Output

```bash
go test ./gitlog -run TestName -v -count=1
```

### Add Debug Output

```go
t.Logf("Debug: %v", variableName)  // Only shown if test fails
t.Errorf("Error: %v", variableName) // Fails the test
```

### Use -race Flag

```bash
go test ./gitlog -race  # Detect race conditions
```

## Performance Testing

### Benchmark Example

```go
func BenchmarkFilterCommits(b *testing.B) {
    commits := makeTestCommits(1000)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        filterCommits(commits, "test")
    }
}
```

### Run Benchmarks

```bash
go test ./gitlog -bench . -benchmem
```

## Continuous Integration

### Pre-commit

```bash
go test ./gitlog -v
```

### CI Pipeline

1. Run all tests
2. Check coverage (must be >= 74.7%)
3. Run benchmarks
4. Report results

## Future Improvements

- [ ] Reorganize engine_test.go into separate files
- [ ] Increase coverage to 80%+
- [ ] Add performance benchmarks
- [ ] Add fuzzing tests
- [ ] Create test matrix for different scenarios
- [ ] Add property-based tests

## Resources

- Go testing package: https://golang.org/pkg/testing/
- Table-driven tests: https://github.com/golang/go/wiki/TableDrivenTests
- Test coverage best practices: https://golang.org/doc/tutorial/add-a-test
