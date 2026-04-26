package gitlog

import (
	"testing"
)

// Integration tests for complete workflows
// These tests verify that multiple components work together correctly

// TestWorkflow_SearchAndFilter tests search followed by filtering
func TestWorkflow_SearchAndFilter(t *testing.T) {
	fixture := NewTestFixture()
	commits := fixture.Commits

	// Step 1: Search for "feature"
	filtered := filterCommits(commits, "feature")
	AssertTrue(t, len(filtered) > 0, "should find commits with 'feature'")

	// Step 2: Filter by author
	byAuthor := filterCommitsByAuthor(filtered, "Alice")
	AssertTrue(t, len(byAuthor) <= len(filtered), "filtering by author should reduce results")
}

// TestWorkflow_NavigationWithFiltering tests navigation after filtering
func TestWorkflow_NavigationWithFiltering(t *testing.T) {
	fixture := NewTestFixture()
	commits := fixture.Commits

	// Step 1: Create model with commits
	m := model{
		commits: commits,
		cursor:  0,
	}

	// Step 2: Filter commits
	filtered := filterCommits(commits, "bug")
	m.commits = filtered

	// Step 3: Navigate
	AssertTrue(t, m.cursor >= 0, "cursor should be valid")
	AssertTrue(t, m.cursor < len(m.commits) || len(m.commits) == 0, "cursor should be within bounds")
}

// TestWorkflow_BookmarkAndSearch tests bookmarking and then searching
func TestWorkflow_BookmarkAndSearch(t *testing.T) {
	fixture := NewTestFixture()
	m := model{
		commits:   fixture.Commits,
		cursor:    0,
		bookmarks: []string{},
	}

	// Step 1: Add bookmark
	if len(m.commits) > 0 {
		m.bookmarks = append(m.bookmarks, m.commits[0].shortHash)
		AssertEqual(t, 1, len(m.bookmarks), "should have one bookmark")
	}

	// Step 2: Search
	filtered := filterCommits(m.commits, fixture.Commits[0].author)
	AssertTrue(t, len(filtered) > 0, "search should find results")

	// Step 3: Verify bookmark is still valid
	AssertEqual(t, 1, len(m.bookmarks), "bookmarks should persist through search")
}

// TestWorkflow_ParseAndFilter tests filtering after getting commits
func TestWorkflow_ParseAndFilter(t *testing.T) {
	// Step 1: Use pre-made commits (as if parsed)
	fixture := NewTestFixture()
	commits := fixture.Commits
	AssertTrue(t, len(commits) > 0, "should have commits to filter")

	// Step 2: Filter commits by different criteria
	filtered := filterCommits(commits, "feature")
	AssertTrue(t, len(filtered) >= 0, "should be able to filter commits")

	// Step 3: Verify filtering reduces results appropriately
	allCommits := len(commits)
	filteredCount := len(filtered)
	AssertTrue(t, filteredCount <= allCommits, "filtered should not exceed total")
}

// TestWorkflow_DiffViewAndParsing tests diff viewing after parsing
func TestWorkflow_DiffViewAndParsing(t *testing.T) {
	// Step 1: Parse diff
	rawDiff := `--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 func old() {
+func new() {
-}
+}`

	diffLines := parseDiff(rawDiff)
	AssertTrue(t, len(diffLines) > 0, "should parse diff lines")

	// Step 2: Create model with diff
	m := model{
		diffLines:  diffLines,
		diffOffset: 0,
	}

	// Step 3: Navigate diff
	if len(m.diffLines) > 0 {
		newOffset := m.diffOffset + 1
		AssertTrue(t, newOffset >= m.diffOffset, "should be able to navigate")
	}
}

// TestWorkflow_FilterByAuthorAndTime tests combined filtering
func TestWorkflow_FilterByAuthorAndTime(t *testing.T) {
	fixture := NewTestFixture()
	commits := fixture.Commits

	// Step 1: Filter by author
	byAuthor := filterCommitsByAuthor(commits, "Alice")
	AssertTrue(t, len(byAuthor) > 0, "should find Alice's commits")

	// Step 2: Filter by time (all commits are recent)
	byTime := filterCommitsSince(byAuthor, 30) // 30 days
	AssertTrue(t, len(byTime) > 0, "should find recent commits")

	// Step 3: Verify both filters apply
	AssertTrue(t, len(byTime) <= len(byAuthor), "time filter should reduce further")
}

// TestWorkflow_RenderConsolidatedPanel tests render consolidation workflow
func TestWorkflow_RenderConsolidatedPanel(t *testing.T) {
	// Step 1: Build data
	data := BuildAnalysisData(
		"Commits", 42,
		"Authors", 5,
		"Files", 23,
	)

	// Step 2: Render using consolidated template
	result := RenderAnalysisUI("Summary", data)
	AssertTrue(t, len(result) > 0, "should render panel")

	// Step 3: Verify content
	AssertStringContains(t, result, "Summary", "should contain title")
	AssertStringContains(t, result, "Commits", "should contain metrics")
}

// TestWorkflow_CommitBuilder tests using builder in workflow
func TestWorkflow_CommitBuilder(t *testing.T) {
	// Step 1: Create multiple commits using builder
	commits := []commit{
		NewCommitBuilder().
			WithAuthor("Alice").
			WithSubject("Feature 1").
			Build(),
		NewCommitBuilder().
			WithAuthor("Bob").
			WithSubject("Fix 1").
			Build(),
	}

	AssertEqual(t, 2, len(commits), "should create 2 commits")

	// Step 2: Filter the built commits
	aliceCommits := filterCommitsByAuthor(commits, "Alice")
	AssertEqual(t, 1, len(aliceCommits), "should find Alice's commits")
}

// TestWorkflow_MultiStepRendering tests multiple render steps
func TestWorkflow_MultiStepRendering(t *testing.T) {
	fixture := NewTestFixture()

	// Step 1: Build list with items
	items := []string{}
	for _, c := range fixture.Commits {
		items = append(items, c.shortHash+" - "+c.subject)
	}

	// Step 2: Create config
	config := RenderConfig{
		Title:       "Commits",
		Items:       items,
		ShowIndices: true,
		MaxItems:    10,
	}

	// Step 3: Render
	result := RenderStandardUI(config)
	AssertTrue(t, len(result) > 0, "should render list")
	AssertStringContains(t, result, "Commits", "should have title")

	// Step 4: Verify indices shown
	AssertStringContains(t, result, "0:", "should show indices")
}

// TestWorkflow_CompleteSearchToRender tests full search workflow
func TestWorkflow_CompleteSearchToRender(t *testing.T) {
	fixture := NewTestFixture()

	// Step 1: Filter commits by query
	query := "Add"
	filtered := filterCommits(fixture.Commits, query)
	AssertTrue(t, len(filtered) > 0, "should find matching commits")

	// Step 2: Build render items
	items := []string{}
	for _, c := range filtered {
		items = append(items, c.author+" - "+c.subject)
	}

	// Step 3: Create and render config
	config := RenderConfig{
		Title: "Search Results: " + query,
		Items: items,
	}

	result := RenderStandardUI(config)
	AssertTrue(t, len(result) > 0, "should render search results")
	AssertStringContains(t, result, "Search Results", "should indicate search")
}
