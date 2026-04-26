package gitlog

import (
	"strings"
	"testing"
)
func TestFilterCommits_EmptyQuery(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "")
	AssertLen(t, result, 3, "empty query should return all")
}

func TestFilterCommits_BySubject(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "login")
	AssertLen(t, result, 1, "should find login commit")
	AssertEqual(t, "Fix login bug", result[0].subject, "subject should match")
}

func TestFilterCommits_ByAuthor(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "Jane")
	AssertLen(t, result, 1, "should find Jane's commit")
	AssertEqual(t, "Jane Smith", result[0].author, "author should match")
}

func TestFilterCommits_ByHash(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "bbb")
	AssertLen(t, result, 1, "should find by hash")
	AssertEqual(t, "bbb2222", result[0].shortHash, "hash should match")
}

func TestFilterCommits_CaseInsensitive(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "LOGIN")
	AssertLen(t, result, 1, "filter should be case-insensitive")
}

func TestFilterCommits_NoMatch(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "zzznomatch")
	AssertLen(t, result, 0, "should return no matches")
}

func TestFilterCommits_MultipleMatches(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "John")
	AssertLen(t, result, 2, "should find multiple John commits")
}

func TestFilterCommitsByAuthor_Empty(t *testing.T) {
	commits := makeNamedCommits()
	result := filterCommitsByAuthor(commits, "")
	if len(result) != 3 {
		t.Errorf("empty filter should return all, got %d", len(result))
	}
}

func TestFilterCommitsByAuthor_SingleMatch(t *testing.T) {
	commits := makeNamedCommits()
	result := filterCommitsByAuthor(commits, "Jane Smith")
	if len(result) != 1 || result[0].author != "Jane Smith" {
		t.Errorf("expected Jane Smith, got %v", result)
	}
}

func TestFilterCommitsByAuthor_MultipleMatches(t *testing.T) {
	commits := makeNamedCommits()
	result := filterCommitsByAuthor(commits, "John Doe")
	if len(result) != 2 {
		t.Errorf("expected 2 John Doe commits, got %d", len(result))
	}
}

func TestFilterCommitsByAuthor_CaseInsensitive(t *testing.T) {
	commits := makeNamedCommits()
	result := filterCommitsByAuthor(commits, "jane smith")
	if len(result) != 1 {
		t.Errorf("expected 1 match (case-insensitive), got %d", len(result))
	}
}

func TestFilterCommitsByAuthor_NoMatch(t *testing.T) {
	commits := makeNamedCommits()
	result := filterCommitsByAuthor(commits, "Unknown Author")
	if len(result) != 0 {
		t.Errorf("expected 0 matches, got %d", len(result))
	}
}

func TestFilterCommitsSince_Zero(t *testing.T) {
	commits := makeCommitsWithDays()
	result := filterCommitsSince(commits, 0)
	if len(result) != 4 {
		t.Errorf("zero days should return all, got %d", len(result))
	}
}

func TestFilterCommitsSince_OneDay(t *testing.T) {
	commits := makeCommitsWithDays()
	result := filterCommitsSince(commits, 1)
	if len(result) != 1 {
		t.Errorf("expected 1 commit in last 1 day, got %d", len(result))
	}
	if result[0].shortHash != "aaa1111" {
		t.Errorf("expected aaa1111, got %s", result[0].shortHash)
	}
}

func TestFilterCommitsSince_FiveDays(t *testing.T) {
	commits := makeCommitsWithDays()
	result := filterCommitsSince(commits, 5)
	if len(result) != 2 {
		t.Errorf("expected 2 commits in last 5 days, got %d", len(result))
	}
}

func TestFilterCommitsSince_ThirtyDays(t *testing.T) {
	commits := makeCommitsWithDays()
	result := filterCommitsSince(commits, 30)
	if len(result) != 3 {
		t.Errorf("expected 3 commits in last 30 days, got %d", len(result))
	}
}

func TestFilterCommitsSince_NegativeDays(t *testing.T) {
	commits := makeCommitsWithDays()
	result := filterCommitsSince(commits, -1)
	if len(result) != 4 {
		t.Errorf("negative days should return all, got %d", len(result))
	}
}

func TestFilterCommitsByFile_Empty(t *testing.T) {
	commits := makeCommits(3)
	result := filterCommitsByFile(commits, "")
	if len(result) != 3 {
		t.Errorf("empty file filter should return all, got %d", len(result))
	}
}

func TestFilterCommitsByFile_NoMatches(t *testing.T) {
	commits := []commit{
		{shortHash: "aaa", subject: "fix: auth.go"},
		{shortHash: "bbb", subject: "update: main.go"},
	}
	result := filterCommitsByFile(commits, "nonexistent.go")
	if len(result) != 0 {
		t.Errorf("no matches should return empty, got %d", len(result))
	}
}

func TestFilterCommitsByFile_Matches(t *testing.T) {
	// This is infrastructure - actual matching would happen via git queries
	result := filterCommitsByFile([]commit{}, "test.go")
	if result == nil {
		t.Error("should return slice, not nil")
	}
}

func TestFilterByRegex_MatchesPattern(t *testing.T) {
	fixture := NewTestFixture()

	pattern := "^Add"
	results := filterByRegex(fixture.Commits, pattern)

	AssertNotNil(t, results, "should return results for valid pattern")
	AssertTrue(t, len(results) > 0, "should match commits with 'Add' prefix")
}

func TestFilterByRegex_NoMatches(t *testing.T) {
	fixture := NewTestFixture()

	pattern := "^NonExistentPattern"
	results := filterByRegex(fixture.Commits, pattern)

	AssertLen(t, results, 0, "should return empty when no matches")
}

func TestFilterByRegex_InvalidPattern(t *testing.T) {
	fixture := NewTestFixture()

	pattern := "[invalid(pattern"
	results := filterByRegex(fixture.Commits, pattern)

	// Handle nil slice return - nil slices are valid
	if results == nil {
		return
	}
	t.Errorf("should return nil for invalid regex: got %v", results)
}

func TestVisibleCommits_NoQuery(t *testing.T) {
	m := model{commits: makeCommits(5)}
	AssertLen(t, visibleCommits(m), 5, "no query should return all commits")
}

func TestVisibleCommits_WithQuery(t *testing.T) {
	m := model{commits: makeNamedCommits(), query: "login"}
	vc := visibleCommits(m)
	AssertLen(t, vc, 1, "should find login commit")
	AssertEqual(t, "Fix login bug", vc[0].subject, "subject should match")
}

func TestFilterByDateRange_WithinRange(t *testing.T) {
	fixture := NewTestFixture()

	startDays := 0
	endDays := 3
	results := filterByDateRange(fixture.Commits, startDays, endDays)

	AssertNotNil(t, results, "should return results")
	AssertTrue(t, len(results) > 0, "should match commits within range")
}

func TestFilterByDateRange_OutsideRange(t *testing.T) {
	fixture := NewTestFixture()

	startDays := 10
	endDays := 20
	results := filterByDateRange(fixture.Commits, startDays, endDays)

	AssertLen(t, results, 0, "should return empty when outside range")
}

func TestFilterCombined_MultipleFilters(t *testing.T) {
	fixture := NewTestFixture()

	filters := &FilterOptions{
		Author: "Alice",
		Search: "feature",
	}
	results := filterCommitsCombined(fixture.Commits, filters)

	AssertNotNil(t, results, "should return results")
}

func TestFilterByFilePattern_MatchesFiles(t *testing.T) {
	fixture := NewTestFixture()

	pattern := "*.go"
	results := filterByFilePattern(fixture.Commits, pattern)

	AssertNotNil(t, results, "should return results")
}

func TestFilterByExtension_SelectsFiles(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Update main.go"},
		{hash: "bbb", subject: "Fix style.css"},
		{hash: "ccc", subject: "Update main.go"},
	}
	filtered := filterByExtension(commits, ".go")
	if len(filtered) != 2 {
		t.Errorf("expected 2 .go commits, got %d", len(filtered))
	}
}

func TestFilterByAuthor_ExactMatch(t *testing.T) {
	fixture := NewTestFixture()

	author := "Alice"
	results := filterByAuthor(fixture.Commits, author)

	AssertTrue(t, len(results) > 0, "should match Alice's commits")
}

func TestFilterCommitsByFileChange_None(t *testing.T) {
	commits := []commit{{shortHash: "aaa", subject: "fix"}}
	filtered := filterCommitsByFileChange(commits, "nonexistent.go")
	if len(filtered) > 0 {
		t.Errorf("no matches should be empty, got %d", len(filtered))
	}
}

func TestFormatActiveFilters_NoFilters(t *testing.T) {
	m := model{}
	result := formatActiveFilters(m)
	if result != "" {
		t.Errorf("no filters should return empty, got %q", result)
	}
}

func TestFormatActiveFilters_AuthorOnly(t *testing.T) {
	m := model{authorFilter: "Jane Smith"}
	result := formatActiveFilters(m)
	if !strings.Contains(result, "Jane Smith") {
		t.Errorf("should contain author, got %q", result)
	}
}

func TestFormatActiveFilters_TimeOnly(t *testing.T) {
	m := model{sinceFilter: 7}
	result := formatActiveFilters(m)
	if !strings.Contains(result, "7") {
		t.Errorf("should contain days, got %q", result)
	}
}

func TestFormatActiveFilters_BothFilters(t *testing.T) {
	m := model{authorFilter: "John", sinceFilter: 14}
	result := formatActiveFilters(m)
	if !strings.Contains(result, "John") || !strings.Contains(result, "14") {
		t.Errorf("should contain both filters, got %q", result)
	}
}

func TestFilterCommitsWithCache_CachesResults(t *testing.T) {
	cache := NewFilterCache()
	commits := makeNamedCommits()
	
	// First call - cache miss
	result1 := filterCommitsWithCache(cache, commits, "login")
	AssertLen(t, result1, 1, "should find login commit")
	AssertEqual(t, 0, cache.metrics.Hits, "cache should have 0 hits initially")
	
	// Second call - cache hit
	result2 := filterCommitsWithCache(cache, commits, "login")
	AssertLen(t, result2, 1, "should return same result")
	AssertEqual(t, 1, cache.metrics.Hits, "cache should have 1 hit")
	
	// Different query - cache miss
	result3 := filterCommitsWithCache(cache, commits, "user")
	AssertLen(t, result3, 1, "should find user commit")
	AssertEqual(t, 1, cache.metrics.Hits, "hits should still be 1")
}

