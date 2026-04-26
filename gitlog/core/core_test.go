package gitlog

import (
	"testing"
)

// --- Type Tests ---

func TestCommitStruct_HasRequiredFields(t *testing.T) {
	c := commit{
		hash:      "abc123",
		shortHash: "abc",
		author:    "Alice",
		when:      "1 day ago",
		subject:   "Test commit",
	}

	if c.hash != "abc123" {
		t.Error("commit should have hash")
	}
	if c.author != "Alice" {
		t.Error("commit should have author")
	}
}

func TestDiffLineStruct_HasKindAndText(t *testing.T) {
	line := diffLine{
		kind: lineAdded,
		text: "new code",
	}

	if line.kind != lineAdded {
		t.Error("diffLine should have kind")
	}
	if line.text != "new code" {
		t.Error("diffLine should have text")
	}
}

func TestFileItemStruct_HasPathAndIndex(t *testing.T) {
	item := fileItem{
		path:    "main.go",
		diffIdx: 0,
	}

	if item.path != "main.go" {
		t.Error("fileItem should have path")
	}
	if item.diffIdx != 0 {
		t.Error("fileItem should have diffIdx")
	}
}

func TestBlameLineStruct_HasAllFields(t *testing.T) {
	blame := blameLine{
		shortHash: "abc1234",
		author:    "Alice",
		date:      "2026-04-20",
		lineNum:   42,
		text:      "code line",
	}

	if blame.shortHash != "abc1234" {
		t.Error("blameLine should have shortHash")
	}
	if blame.lineNum != 42 {
		t.Error("blameLine should have lineNum")
	}
}

// --- Parser Tests ---

func TestParseCommits_ValidInput(t *testing.T) {
	input := "abc1234def5678901234567890123456789012\x00abc1234\x00John Doe\x002 days ago\x00Fix login bug\n" +
		"xyz9876qwerty12345678901234567890123456\x00xyz9876\x00Jane Smith\x005 days ago\x00Add user model\n"

	commits := parseCommits(input)

	if len(commits) != 2 {
		t.Errorf("expected 2 commits, got %d", len(commits))
	}

	if commits[0].author != "John Doe" {
		t.Errorf("first commit author should be John Doe, got %s", commits[0].author)
	}

	if commits[1].when != "5 days ago" {
		t.Errorf("second commit when should be '5 days ago', got %s", commits[1].when)
	}
}

func TestParseCommits_EmptyInput(t *testing.T) {
	if parseCommits("") != nil {
		t.Error("parseCommits should return nil for empty input")
	}
}

func TestParseCommits_WhitespaceOnly(t *testing.T) {
	if parseCommits("   \n\n  ") != nil {
		t.Error("parseCommits should return nil for whitespace-only input")
	}
}

func TestParseCommits_MalformedLines(t *testing.T) {
	input := "not-enough-fields\nok1234567890123456789012345678901234\x00ok12345\x00Auth\x001h ago\x00Good commit\n"

	commits := parseCommits(input)

	if len(commits) != 1 {
		t.Errorf("should skip malformed lines, got %d commits", len(commits))
	}
}

func TestParseDiff_AddedLines(t *testing.T) {
	diff := "+added line\n+another added line"
	lines := parseDiff(diff)

	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	if lines[0].kind != lineAdded {
		t.Error("first line should be marked as added")
	}
}

func TestParseDiff_RemovedLines(t *testing.T) {
	diff := "-removed line"
	lines := parseDiff(diff)

	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}

	if lines[0].kind != lineRemoved {
		t.Error("line should be marked as removed")
	}
}

func TestParseDiff_ContextLines(t *testing.T) {
	diff := " context line"
	lines := parseDiff(diff)

	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}

	if lines[0].kind != lineContext {
		t.Error("line should be marked as context")
	}
}

func TestParseDiff_HunkHeaders(t *testing.T) {
	diff := "@@ -10,7 +10,9 @@"
	lines := parseDiff(diff)

	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}

	if lines[0].kind != lineHunk {
		t.Error("line should be marked as hunk header")
	}
}

func TestParseFileItems_CreatesItems(t *testing.T) {
	commits := []commit{
		{hash: "abc123", subject: "main.go"},
		{hash: "def456", subject: "util.go"},
	}

	items := parseFileItems(commits)

	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}

	if items[0].path != "main.go" {
		t.Errorf("first item path should be main.go, got %s", items[0].path)
	}

	if items[0].diffIdx != 0 {
		t.Errorf("first item diffIdx should be 0, got %d", items[0].diffIdx)
	}
}

func TestParseFileItems_EmptyCommits(t *testing.T) {
	items := parseFileItems([]commit{})

	if len(items) != 0 {
		t.Errorf("expected 0 items for empty commits, got %d", len(items))
	}
}

// --- Panel Enum Tests ---

func TestPanelEnum_HasListAndDiff(t *testing.T) {
	if panelList != 0 {
		t.Error("panelList should be 0")
	}

	if panelDiff != 1 {
		t.Error("panelDiff should be 1")
	}
}

// --- LineKind Enum Tests ---

func TestLineKindEnum_HasAllTypes(t *testing.T) {
	if lineContext != 0 {
		t.Error("lineContext should be 0")
	}

	if lineAdded != 1 {
		t.Error("lineAdded should be 1")
	}

	if lineRemoved != 2 {
		t.Error("lineRemoved should be 2")
	}

	if lineHunk != 3 {
		t.Error("lineHunk should be 3")
	}

	if lineMeta != 4 {
		t.Error("lineMeta should be 4")
	}
}
