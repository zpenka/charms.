package gitlog

import (
	"strings"
	"testing"
)
func TestParseCommits_Empty(t *testing.T) {
	if parseCommits("") != nil {
		t.Error("expected nil for empty input")
	}
	if parseCommits("   \n\n  ") != nil {
		t.Error("expected nil for whitespace input")
	}
}

func TestParseCommits_SubjectPreserved(t *testing.T) {
	// SplitN(5) keeps the subject intact even if it contained NUL bytes (unlikely but safe)
	input := "hash1234567890123456789012345678901234\x00h1234567\x00Author\x002h ago\x00subject with extra\n"
	commits := parseCommits(input)
	if len(commits) != 1 {
		t.Fatalf("expected 1, got %d", len(commits))
	}
	if commits[0].subject != "subject with extra" {
		t.Errorf("subject: got %q", commits[0].subject)
	}
}

func TestParseCommits_SkipsMalformed(t *testing.T) {
	input := "not-enough-fields\nok1234567890123456789012345678901234\x00ok12345\x00Auth\x001h ago\x00Good commit\n"
	commits := parseCommits(input)
	if len(commits) != 1 {
		t.Fatalf("expected 1 valid commit, got %d", len(commits))
	}
}

func TestParseDiff_Added(t *testing.T) {
	lines := parseDiff("+added line")
	if len(lines) != 1 || lines[0].kind != lineAdded {
		t.Errorf("expected lineAdded")
	}
}

func TestParseDiff_Removed(t *testing.T) {
	lines := parseDiff("-removed line")
	if len(lines) != 1 || lines[0].kind != lineRemoved {
		t.Errorf("expected lineRemoved")
	}
}

func TestParseDiff_Hunk(t *testing.T) {
	lines := parseDiff("@@ -1,5 +1,8 @@ func main()")
	if len(lines) != 1 || lines[0].kind != lineHunk {
		t.Errorf("expected lineHunk")
	}
}

func TestParseDiff_Meta(t *testing.T) {
	for _, text := range []string{
		"diff --git a/foo b/foo",
		"index abc..def 100644",
		"--- a/foo",
		"+++ b/foo",
		"new file mode 100644",
		"deleted file mode 100644",
		"similarity index 90%",
		"rename from a/b",
	} {
		lines := parseDiff(text)
		if lines[0].kind != lineMeta {
			t.Errorf("expected lineMeta for %q, got %d", text, lines[0].kind)
		}
	}
}

func TestParseDiff_Context(t *testing.T) {
	lines := parseDiff(" context line")
	if lines[0].kind != lineContext {
		t.Error("expected lineContext")
	}
}

func TestParseDiff_TextPreserved(t *testing.T) {
	lines := parseDiff("+hello world")
	if lines[0].text != "+hello world" {
		t.Errorf("text not preserved: %q", lines[0].text)
	}
}

func TestParseBlameLine_Valid(t *testing.T) {
	line := "abc1234a (John Doe        2024-01-15   42) func main() {"
	bl, ok := parseBlameLine(line)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if bl.shortHash != "abc1234" {
		t.Errorf("shortHash: got %q", bl.shortHash)
	}
	if bl.author != "John Doe" {
		t.Errorf("author: got %q", bl.author)
	}
	if bl.date != "2024-01-15" {
		t.Errorf("date: got %q", bl.date)
	}
	if bl.lineNum != 42 {
		t.Errorf("lineNum: got %d", bl.lineNum)
	}
	if bl.text != "func main() {" {
		t.Errorf("text: got %q", bl.text)
	}
}

func TestParseBlameLine_BoundaryCommit(t *testing.T) {
	// git prefixes boundary commits with ^
	line := "^abc1234 (Author Name      2024-01-01    1) first line"
	bl, ok := parseBlameLine(line)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if strings.HasPrefix(bl.shortHash, "^") {
		t.Error("shortHash should not start with ^")
	}
}

func TestParseBlameLine_Invalid(t *testing.T) {
	_, ok := parseBlameLine("not a blame line at all")
	if ok {
		t.Error("expected ok=false for malformed line")
	}
}

func TestParseBlameLine_EmptyContent(t *testing.T) {
	line := "abc1234a (Author     2024-01-01    1)"
	bl, ok := parseBlameLine(line)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if bl.text != "" {
		t.Errorf("expected empty text, got %q", bl.text)
	}
}

func TestParseBranches_Simple(t *testing.T) {
	input := "  main\n* develop\n  remotes/origin/main\n"
	branches := parseBranches(input)
	AssertLen(t, branches, 3, "should parse 3 branches")
	AssertEqual(t, "main", branches[0], "first branch should be main")
	AssertEqual(t, "develop", branches[1], "second branch should be develop")
}

func TestParseBranches_Empty(t *testing.T) {
	if parseBranches("") != nil {
		t.Error("expected nil for empty input")
	}
}

func TestParseBranches_SkipsRefPointers(t *testing.T) {
	input := "  main\n  remotes/origin/HEAD -> origin/main\n  remotes/origin/main\n"
	branches := parseBranches(input)
	for _, b := range branches {
		if strings.Contains(b, "->") {
			t.Errorf("should skip ref pointer lines, got %q", b)
		}
	}
}

func TestParseBranches_SkipsEmpty(t *testing.T) {
	input := "  main\n\n  develop\n"
	branches := parseBranches(input)
	if len(branches) != 2 {
		t.Errorf("expected 2, got %d", len(branches))
	}
}

func TestParseCurrentBranch_Found(t *testing.T) {
	input := "  main\n* develop\n  remotes/origin/main\n"
	if got := parseCurrentBranch(input); got != "develop" {
		t.Errorf("expected develop, got %q", got)
	}
}

func TestParseCurrentBranch_Detached(t *testing.T) {
	input := "* (HEAD detached at abc1234)\n  main\n"
	got := parseCurrentBranch(input)
	// detached HEAD is still returned as-is so the model can display it
	if got == "" {
		t.Error("expected non-empty for detached HEAD")
	}
}

func TestParseCurrentBranch_None(t *testing.T) {
	if got := parseCurrentBranch("  main\n  develop\n"); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestParseBlame_MultiLine(t *testing.T) {
	input := "abc1234a (John Doe   2024-01-15    1) line one\n" +
		"def5678b (Jane Smith  2024-01-16    2) line two\n"
	lines := parseBlame(input)
	if len(lines) != 2 {
		t.Fatalf("expected 2, got %d", len(lines))
	}
	if lines[0].lineNum != 1 || lines[1].lineNum != 2 {
		t.Errorf("wrong line numbers: %d %d", lines[0].lineNum, lines[1].lineNum)
	}
}

func TestParseBlame_SkipsMalformed(t *testing.T) {
	input := "abc1234a (Author  2024-01-01  1) good\nbad line\ndef5678b (Author  2024-01-02  2) also good\n"
	lines := parseBlame(input)
	if len(lines) != 2 {
		t.Errorf("expected 2 valid lines, got %d", len(lines))
	}
}

func TestParseCount_Empty(t *testing.T) {
	if parseCount("") != 1 {
		t.Error("empty should return 1")
	}
}

func TestParseCount_Valid(t *testing.T) {
	if parseCount("5") != 5 {
		t.Errorf("expected 5, got %d", parseCount("5"))
	}
	if parseCount("12") != 12 {
		t.Errorf("expected 12, got %d", parseCount("12"))
	}
}

func TestParseCount_Zero(t *testing.T) {
	if parseCount("0") != 1 {
		t.Error("zero should return 1")
	}
}

func TestParseCount_Capped(t *testing.T) {
	if parseCount("9999") != 200 {
		t.Errorf("expected cap at 200, got %d", parseCount("9999"))
	}
}

func TestParseCount_Invalid(t *testing.T) {
	if parseCount("abc") != 1 {
		t.Error("invalid should return 1")
	}
}

func TestParseFileItems_Empty(t *testing.T) {
	items := parseFileItemsFromDiff([]diffLine{
		{lineContext, "context"},
		{lineAdded, "+added"},
	})
	if len(items) != 0 {
		t.Errorf("expected 0, got %d", len(items))
	}
}

func TestParseFileItems_Single(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "diff --git a/foo.go b/foo.go"},
		{lineMeta, "index abc..def 100644"},
		{lineAdded, "+hello"},
	}
	items := parseFileItemsFromDiff(lines)
	if len(items) != 1 {
		t.Fatalf("expected 1, got %d", len(items))
	}
	if items[0].path != "foo.go" {
		t.Errorf("path: got %q", items[0].path)
	}
	if items[0].diffIdx != 0 {
		t.Errorf("diffIdx: got %d", items[0].diffIdx)
	}
}

func TestParseFileItems_Multiple(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "diff --git a/a.go b/a.go"},
		{lineContext, "context"},
		{lineMeta, "diff --git a/path/to/b.go b/path/to/b.go"},
		{lineAdded, "+new"},
	}
	items := parseFileItemsFromDiff(lines)
	AssertLen(t, items, 2, "should parse 2 files")
	AssertEqual(t, "a.go", items[0].path, "first path should match")
	AssertEqual(t, 0, items[0].diffIdx, "first diffIdx should be 0")
	AssertEqual(t, "path/to/b.go", items[1].path, "second path should match")
	AssertEqual(t, 2, items[1].diffIdx, "second diffIdx should be 2")
}

func TestParseFileItems_SkipsNonDiffGit(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "--- a/foo.go"},
		{lineMeta, "+++ b/foo.go"},
		{lineMeta, "diff --git a/bar.go b/bar.go"},
	}
	items := parseFileItemsFromDiff(lines)
	AssertLen(t, items, 1, "should only parse diff --git lines")
}

func TestParseGitReferences_None(t *testing.T) {
	refs := parseGitReferences("just a commit message")
	if len(refs) != 0 {
		t.Errorf("expected no refs, got %d", len(refs))
	}
}

func TestParseGitReferences_IssueNumber(t *testing.T) {
	refs := parseGitReferences("fix #123 for login")
	if len(refs) != 1 || refs[0] != "123" {
		t.Errorf("expected #123, got %v", refs)
	}
}

func TestParseGitReferences_Multiple(t *testing.T) {
	refs := parseGitReferences("fixes #123 and closes #456")
	if len(refs) != 2 {
		t.Errorf("expected 2 refs, got %d", len(refs))
	}
}

func TestParseGitReferences_FixesKeyword(t *testing.T) {
	refs := parseGitReferences("fixes #789")
	if len(refs) != 1 {
		t.Errorf("expected 1 ref, got %d", len(refs))
	}
}

func TestParseHunks_Empty(t *testing.T) {
	hunks := parseHunks([]diffLine{})
	if len(hunks) != 0 {
		t.Errorf("empty diff should have 0 hunks, got %d", len(hunks))
	}
}

func TestParseHunks_Single(t *testing.T) {
	lines := []diffLine{
		{lineHunk, "@@ -1,5 +1,8 @@ func main()"},
		{lineAdded, "+new line"},
	}
	hunks := parseHunks(lines)
	if len(hunks) != 1 {
		t.Errorf("expected 1 hunk, got %d", len(hunks))
	}
}

func TestParseHunks_Multiple(t *testing.T) {
	lines := []diffLine{
		{lineHunk, "@@ -1,5 +1,8 @@ func main()"},
		{lineAdded, "+line"},
		{lineHunk, "@@ -20,3 +25,5 @@ other func"},
		{lineRemoved, "-removed"},
	}
	hunks := parseHunks(lines)
	if len(hunks) != 2 {
		t.Errorf("expected 2 hunks, got %d", len(hunks))
	}
}

func TestParseDateRange_Empty(t *testing.T) {
	start, end, err := parseDateRange("")
	if err != nil {
		t.Errorf("empty range should not error, got %v", err)
	}
	if start != nil || end != nil {
		t.Error("empty range should return nil dates")
	}
}

func TestParseDateRange_SingleDate(t *testing.T) {
	start, _, err := parseDateRange("2024-01-15")
	if err != nil {
		t.Errorf("single date should not error, got %v", err)
	}
	if start == nil {
		t.Error("start date should be set")
	}
}

func TestParseDateRange_Range(t *testing.T) {
	start, end, err := parseDateRange("2024-01-01..2024-01-31")
	if err != nil {
		t.Errorf("date range should not error, got %v", err)
	}
	if start == nil {
		t.Error("start date should be set")
	}
	if end == nil {
		t.Error("end date should be set")
	}
}

func TestParseDateRange_InvalidFormat(t *testing.T) {
	_, _, err := parseDateRange("not-a-date")
	if err == nil {
		t.Error("invalid date should return error")
	}
}

func TestParseTags_Empty(t *testing.T) {
	tags := parseTags("")
	if len(tags) != 0 {
		t.Errorf("empty input should return empty slice, got %d", len(tags))
	}
}

func TestParseTags_SingleTag(t *testing.T) {
	tags := parseTags("v1.0.0\n")
	if len(tags) != 1 || tags[0] != "v1.0.0" {
		t.Errorf("expected [v1.0.0], got %v", tags)
	}
}

func TestParseTags_MultipleTagsWithHashes(t *testing.T) {
	input := "abc1234 v1.0.0\ndef5678 v1.0.1\n"
	tags := parseTags(input)
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestParseStashList_Empty(t *testing.T) {
	stashes := parseStashList("")
	if len(stashes) != 0 {
		t.Errorf("empty input should have no stashes, got %d", len(stashes))
	}
}

func TestParseStashList_Single(t *testing.T) {
	input := "stash@{0}: WIP on main: abc1234 commit message\n"
	stashes := parseStashList(input)
	if len(stashes) != 1 {
		t.Errorf("expected 1 stash, got %d", len(stashes))
	}
	if stashes[0].name != "stash@{0}" {
		t.Errorf("wrong stash name, got %q", stashes[0].name)
	}
}

func TestParseStashList_Multiple(t *testing.T) {
	input := "stash@{0}: WIP on main: abc1234 msg1\nstash@{1}: WIP on feature: def5678 msg2\n"
	stashes := parseStashList(input)
	if len(stashes) != 2 {
		t.Errorf("expected 2 stashes, got %d", len(stashes))
	}
}

func TestParseReflog_Empty(t *testing.T) {
	entries := parseReflog("")
	if len(entries) != 0 {
		t.Errorf("empty input should have no entries, got %d", len(entries))
	}
}

func TestParseReflog_Single(t *testing.T) {
	input := "abc1234 HEAD@{0}: commit: initial commit\n"
	entries := parseReflog(input)
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestParseRebaseSequence_Empty(t *testing.T) {
	seq := parseRebaseSequence([]commit{})
	if len(seq) != 0 {
		t.Errorf("empty commits should have no sequence, got %d", len(seq))
	}
}

func TestParseRebaseSequence_Linear(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "first"},
		{hash: "bbb", subject: "second"},
		{hash: "ccc", subject: "third"},
	}
	seq := parseRebaseSequence(commits)
	if len(seq) != 3 {
		t.Errorf("expected 3 operations, got %d", len(seq))
	}
}

func TestParseSubmodules_ExtractsInfo(t *testing.T) {
	configOutput := "[submodule \"lib1\"]\npath = lib1\nurl = https://github.com/user/lib1\n"
	submodules := parseSubmodules(configOutput)
	if len(submodules) == 0 {
		t.Error("should parse submodules")
	}
}

func TestDiffParsing_CacheHitRate(t *testing.T) {
	cache := newDiffCache(100)
	lines := []diffLine{{lineAdded, "+test"}}

	cache.set("abc", lines)
	cache.set("def", lines)
	cache.get("abc") // First hit
	cache.set("abc", lines)
	cache.get("abc") // Second hit

	hits := cache.getHitCount()
	if hits != 2 {
		t.Errorf("expected 2 hits, got %d", hits)
	}
}

func TestSafeParseCommitGraph_EmptyList(t *testing.T) {
	graph := safeParseCommitGraph(nil)
	if graph == nil {
		t.Error("should return empty slice, not nil")
	}
}

func TestParseCommitGraph_Empty(t *testing.T) {
	graph := parseCommitGraph([]commit{})
	if len(graph) != 0 {
		t.Errorf("empty commits should have no graph nodes, got %d", len(graph))
	}
}

func TestParseCommitGraph_Linear(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "first"},
		{hash: "bbb", subject: "second"},
		{hash: "ccc", subject: "third"},
	}
	graph := parseCommitGraph(commits)
	if len(graph) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(graph))
	}
}

