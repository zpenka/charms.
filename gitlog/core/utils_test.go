package gitlog

import (
	"testing"
)

func TestTruncate_NoTruncation(t *testing.T) {
	result := truncate("hello", 10)

	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncate_WithTruncation(t *testing.T) {
	result := truncate("hello world", 5)

	if result != "hell…" {
		t.Errorf("expected 'hell…', got %q", result)
	}
}

func TestTruncate_ExactLength(t *testing.T) {
	result := truncate("hello", 5)

	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestTruncate_SingleCharMax(t *testing.T) {
	result := truncate("hello", 1)

	if result != "…" {
		t.Errorf("expected '…', got %q", result)
	}
}

func TestTruncate_ZeroMax(t *testing.T) {
	result := truncate("hello", 0)

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestFirstWord_WithSpaces(t *testing.T) {
	result := firstWord("hello world test")

	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestFirstWord_NoSpaces(t *testing.T) {
	result := firstWord("hello")

	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
}

func TestFirstWord_Empty(t *testing.T) {
	result := firstWord("")

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestParseCount_DefaultEmpty(t *testing.T) {
	result := parseCount("")

	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}

func TestParseCount_ValidNumber(t *testing.T) {
	result := parseCount("42")

	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
}

func TestParseCount_ClampedToMax(t *testing.T) {
	result := parseCount("500")

	if result != 200 {
		t.Errorf("expected 200 (clamped), got %d", result)
	}
}

func TestParseCount_NegativeDefaultsToOne(t *testing.T) {
	result := parseCount("-5")

	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}
}

func TestParseGitReferences_Single(t *testing.T) {
	result := parseGitReferences("Fix bug #123")

	if len(result) != 1 {
		t.Errorf("expected 1 reference, got %d", len(result))
	}
	if result[0] != "123" {
		t.Errorf("expected '123', got %q", result[0])
	}
}

func TestParseGitReferences_Multiple(t *testing.T) {
	result := parseGitReferences("Fix #123 and resolve #456")

	if len(result) != 2 {
		t.Errorf("expected 2 references, got %d", len(result))
	}
}

func TestParseGitReferences_Duplicates(t *testing.T) {
	result := parseGitReferences("Fix #123 and #123 again")

	if len(result) != 1 {
		t.Errorf("expected 1 unique reference, got %d", len(result))
	}
}

func TestParseBranches_Single(t *testing.T) {
	output := "  main\n  develop"
	result := parseBranches(output)

	if len(result) != 2 {
		t.Errorf("expected 2 branches, got %d", len(result))
	}
	if result[0] != "main" {
		t.Errorf("expected 'main', got %q", result[0])
	}
}

func TestParseBranches_SkipsCurrent(t *testing.T) {
	output := "* main\n  develop"
	result := parseBranches(output)

	if len(result) != 2 {
		t.Errorf("expected 2 branches, got %d", len(result))
	}
	if result[0] != "main" {
		t.Errorf("expected 'main' without *, got %q", result[0])
	}
}

func TestParseCurrentBranch_Found(t *testing.T) {
	output := "  develop\n* main\n  feature/new"
	result := parseCurrentBranch(output)

	if result != "main" {
		t.Errorf("expected 'main', got %q", result)
	}
}

func TestParseCurrentBranch_NotFound(t *testing.T) {
	output := "  develop\n  main"
	result := parseCurrentBranch(output)

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestParseBlameLine_Valid(t *testing.T) {
	line := "abc1234 (John Doe         2026-04-20    42) code here"
	bl, ok := parseBlameLine(line)

	if !ok {
		t.Error("should parse valid blame line")
	}
	if bl.shortHash != "abc1234" {
		t.Errorf("expected 'abc1234', got %q", bl.shortHash)
	}
	if bl.author != "John Doe" {
		t.Errorf("expected 'John Doe', got %q", bl.author)
	}
	if bl.date != "2026-04-20" {
		t.Errorf("expected '2026-04-20', got %q", bl.date)
	}
	if bl.lineNum != 42 {
		t.Errorf("expected 42, got %d", bl.lineNum)
	}
}

func TestParseBlameLine_Invalid(t *testing.T) {
	line := "this is not a blame line"
	_, ok := parseBlameLine(line)

	if ok {
		t.Error("should not parse invalid blame line")
	}
}

func TestParseBlame_Multiple(t *testing.T) {
	output := `abc1234 (John Doe         2026-04-20    42) code
def5678 (Jane Smith       2026-04-19    43) more code`
	result := parseBlame(output)

	if len(result) != 2 {
		t.Errorf("expected 2 blame lines, got %d", len(result))
	}
	if result[0].author != "John Doe" {
		t.Errorf("expected 'John Doe', got %q", result[0].author)
	}
}

func TestIsMergeCommit_WithMergeBranch(t *testing.T) {
	lines := []diffLine{
		{lineContext, "some context"},
		{lineMeta, "Merge branch 'feature' into main"},
	}

	result := isMergeCommit(lines)

	if !result {
		t.Error("should detect merge commit with 'Merge branch'")
	}
}

func TestIsMergeCommit_WithMergeColon(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "Merge: abc def"},
	}

	result := isMergeCommit(lines)

	if !result {
		t.Error("should detect merge commit with 'Merge:'")
	}
}

func TestIsMergeCommit_NotMerge(t *testing.T) {
	lines := []diffLine{
		{lineAdded, "+ some code"},
		{lineContext, "context"},
	}

	result := isMergeCommit(lines)

	if result {
		t.Error("should not detect non-merge as merge commit")
	}
}
