package gitlog

import (
	"testing"
)

func TestFilterCommits_EmptyQuery(t *testing.T) {
	commits := []commit{
		{shortHash: "abc", author: "Alice", subject: "Fix bug"},
		{shortHash: "def", author: "Bob", subject: "Add feature"},
	}

	result := filterCommits(commits, "")

	if len(result) != 2 {
		t.Errorf("empty query should return all commits, got %d", len(result))
	}
}

func TestFilterCommits_MatchesSubject(t *testing.T) {
	commits := []commit{
		{shortHash: "abc", author: "Alice", subject: "Fix login bug"},
		{shortHash: "def", author: "Bob", subject: "Add user model"},
	}

	result := filterCommits(commits, "login")

	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
	if result[0].subject != "Fix login bug" {
		t.Errorf("expected 'Fix login bug', got %q", result[0].subject)
	}
}

func TestFilterCommits_MatchesAuthor(t *testing.T) {
	commits := []commit{
		{shortHash: "abc", author: "Alice", subject: "Fix bug"},
		{shortHash: "def", author: "Bob", subject: "Add feature"},
	}

	result := filterCommits(commits, "alice")

	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
	if result[0].author != "Alice" {
		t.Errorf("expected author 'Alice', got %q", result[0].author)
	}
}

func TestFilterCommits_MatchesHash(t *testing.T) {
	commits := []commit{
		{shortHash: "abc1111", author: "Alice", subject: "Fix bug"},
		{shortHash: "def2222", author: "Bob", subject: "Add feature"},
	}

	result := filterCommits(commits, "abc")

	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
	if result[0].shortHash != "abc1111" {
		t.Errorf("expected hash 'abc1111', got %q", result[0].shortHash)
	}
}

func TestFilterCommitsByAuthor_ExactMatch(t *testing.T) {
	commits := []commit{
		{shortHash: "abc", author: "Alice", subject: "Fix bug"},
		{shortHash: "def", author: "Bob", subject: "Add feature"},
		{shortHash: "ghi", author: "Alice", subject: "Update docs"},
	}

	result := filterCommitsByAuthor(commits, "Alice")

	if len(result) != 2 {
		t.Errorf("expected 2 results for Alice, got %d", len(result))
	}
}

func TestFilterCommitsByAuthor_EmptyAuthor(t *testing.T) {
	commits := []commit{
		{shortHash: "abc", author: "Alice", subject: "Fix bug"},
	}

	result := filterCommitsByAuthor(commits, "")

	if len(result) != 1 {
		t.Errorf("empty author should return all commits, got %d", len(result))
	}
}

func TestFilterCommitsSince_WithinRange(t *testing.T) {
	commits := []commit{
		{shortHash: "abc", author: "Alice", when: "2 days ago"},
		{shortHash: "def", author: "Bob", when: "5 days ago"},
		{shortHash: "ghi", author: "Charlie", when: "10 days ago"},
	}

	result := filterCommitsSince(commits, 7)

	if len(result) != 2 {
		t.Errorf("expected 2 commits within 7 days, got %d", len(result))
	}
}

func TestFilterCommitsSince_NoDaysFilter(t *testing.T) {
	commits := []commit{
		{shortHash: "abc", author: "Alice", when: "2 days ago"},
		{shortHash: "def", author: "Bob", when: "100 days ago"},
	}

	result := filterCommitsSince(commits, 0)

	if len(result) != 2 {
		t.Errorf("zero days should return all commits, got %d", len(result))
	}
}

func TestIsWithinDays_DaysUnit(t *testing.T) {
	if !isWithinDays("5 days ago", 10) {
		t.Error("5 days ago should be within 10 days")
	}
	if isWithinDays("15 days ago", 10) {
		t.Error("15 days ago should not be within 10 days")
	}
}

func TestIsWithinDays_WeeksUnit(t *testing.T) {
	if !isWithinDays("2 weeks ago", 20) {
		t.Error("2 weeks ago (14 days) should be within 20 days")
	}
	if isWithinDays("3 weeks ago", 20) {
		t.Error("3 weeks ago (21 days) should not be within 20 days")
	}
}

func TestFilterByExtension_MatchesExtension(t *testing.T) {
	commits := []commit{
		{shortHash: "abc", subject: ".go"},
		{shortHash: "def", subject: ".js"},
		{shortHash: "ghi", subject: ".go"},
	}

	result := filterByExtension(commits, ".go")

	if len(result) != 2 {
		t.Errorf("expected 2 .go commits, got %d", len(result))
	}
}

func TestFilterByRegex_ValidPattern(t *testing.T) {
	commits := []commit{
		{shortHash: "abc", subject: "Fix login bug"},
		{shortHash: "def", subject: "Add user model"},
		{shortHash: "ghi", subject: "Fix logout bug"},
	}

	result := filterByRegex(commits, "Fix.*bug")

	if len(result) != 2 {
		t.Errorf("expected 2 matches for regex, got %d", len(result))
	}
}

func TestFilterByRegex_InvalidPattern(t *testing.T) {
	commits := []commit{
		{shortHash: "abc", subject: "Fix bug"},
	}

	result := filterByRegex(commits, "[invalid")

	if result != nil {
		t.Errorf("invalid regex should return nil, got %v", result)
	}
}

func TestGroupCommits_ByAuthor(t *testing.T) {
	commits := []commit{
		{shortHash: "abc", author: "Alice", subject: "Fix bug"},
		{shortHash: "def", author: "Bob", subject: "Add feature"},
		{shortHash: "ghi", author: "Alice", subject: "Update docs"},
	}

	groups := groupCommits(commits, "author")

	if len(groups) < 2 {
		t.Errorf("expected at least 2 groups, got %d", len(groups))
	}
}
