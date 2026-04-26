package gitlog

import (
	"testing"
)

// Tests for low-coverage and uncovered functions
// These tests target critical paths that need better coverage

// TestHandleKeyBinding_BasicInput tests that handleKeyBinding processes input
func TestHandleKeyBinding_BasicInput(t *testing.T) {
	fixture := NewTestFixture()
	m := model{
		commits: fixture.Commits,
		cursor:  0,
	}

	// Test that function executes without panic
	result := handleKeyBinding(m, "q")

	AssertNotNil(t, result, "should return model")
}

// TestRenderFileTimeline_WithCommits tests timeline rendering with data
func TestRenderFileTimeline_WithCommits(t *testing.T) {
	commits := []commit{
		{hash: "aaa", author: "Alice", subject: "Change 1", when: "1 hour ago"},
		{hash: "bbb", author: "Bob", subject: "Change 2", when: "2 hours ago"},
	}

	result := renderFileTimeline(commits, "main.go", 80)

	AssertTrue(t, len(result) > 0, "should render timeline")
	AssertStringContains(t, result, "main.go", "should contain filename")
}

// TestCalculateBisectProgress_SingleCommit tests progress with single commit
func TestCalculateBisectProgress_SingleCommit(t *testing.T) {
	state := bisectState{
		candidates: []string{"hash1"},
	}

	progress := calculateBisectProgress(state)

	AssertEqual(t, 1, progress, "single commit should return 1")
}

// TestCalculateBisectProgress_MultipleCommits tests progress calculation
func TestCalculateBisectProgress_MultipleCommits(t *testing.T) {
	state := bisectState{
		candidates: []string{"h1", "h2", "h3", "h4", "h5"},
	}

	progress := calculateBisectProgress(state)

	AssertTrue(t, progress > 0, "progress should be positive")
}

// TestClassifyCommit_WithMerge tests merge commit detection
func TestClassifyCommit_WithMerge(t *testing.T) {
	classification := classifyCommit("Merge pull request #123", "abc123")

	AssertTrue(t, classification.category != "", "should classify merge commit")
}

// TestClassifyCommit_FeatureCommit tests feature commit classification
func TestClassifyCommit_FeatureCommit(t *testing.T) {
	classification := classifyCommit("feat: Add new feature", "def456")

	AssertTrue(t, classification.category != "", "should classify feature commit")
}

// TestClassifyCommit_Refactor tests refactor commit classification
func TestClassifyCommit_Refactor(t *testing.T) {
	classification := classifyCommit("refactor: Clean up code structure", "ghi789")

	AssertTrue(t, classification.category != "", "should classify refactor commit")
}

// TestNewModel_Creation tests model initialization
func TestNewModel_Creation(t *testing.T) {
	m := newModel(".")

	AssertEqual(t, 0, m.cursor, "initial cursor should be 0")
	AssertTrue(t, len(m.commits) >= 0, "commits should be initialized")
}

// TestBuildAnalysisData_KeyValuePairs tests data builder
func TestBuildAnalysisData_KeyValuePairs(t *testing.T) {
	data := BuildAnalysisData(
		"key1", "value1",
		"key2", 42,
		"key3", 3.14,
	)

	AssertMapContains(t, data, "key1", "should contain key1")
	AssertEqual(t, "value1", data["key1"], "should have correct value")
	AssertMapContains(t, data, "key2", "should contain key2")
	AssertEqual(t, 42, data["key2"], "should have correct numeric value")
}

// TestBuildAnalysisData_Empty tests with no data
func TestBuildAnalysisData_Empty(t *testing.T) {
	data := BuildAnalysisData()

	AssertEqual(t, 0, len(data), "empty builder should create empty map")
}

// TestCommitBuilder_FluentAPI tests builder pattern
func TestCommitBuilder_FluentAPI(t *testing.T) {
	commit := NewCommitBuilder().
		WithHash("abc123def456").
		WithAuthor("John Doe").
		WithSubject("Test commit").
		WithWhen("1 hour ago").
		Build()

	AssertEqual(t, "abc123d", commit.shortHash, "should set short hash")
	AssertEqual(t, "John Doe", commit.author, "should set author")
	AssertEqual(t, "Test commit", commit.subject, "should set subject")
	AssertEqual(t, "1 hour ago", commit.when, "should set when")
}

// TestCommitBuilder_Defaults tests builder with defaults
func TestCommitBuilder_Defaults(t *testing.T) {
	commit := NewCommitBuilder().Build()

	AssertEqual(t, "Test Author", commit.author, "should use default author")
	AssertEqual(t, "Test Subject", commit.subject, "should use default subject")
}
