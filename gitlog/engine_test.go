package gitlog

import (
	"strings"
	"testing"
)

func makeCommits(n int) []commit {
	var cs []commit
	for i := 0; i < n; i++ {
		cs = append(cs, commit{
			hash:      strings.Repeat("a", 40),
			shortHash: "abc1234",
			author:    "Test User",
			when:      "1d ago",
			subject:   "commit message",
		})
	}
	return cs
}

func makeDiffLines(n int) []diffLine {
	lines := make([]diffLine, n)
	for i := range lines {
		lines[i] = diffLine{kind: lineContext, text: "context"}
	}
	return lines
}

// --- parseCommits ---

func TestParseCommits(t *testing.T) {
	input := "abc1234def5678901234567890123456789012\x00abc1234\x00John Doe\x002 days ago\x00Fix login bug\n" +
		"xyz9876qwerty12345678901234567890123456\x00xyz9876\x00Jane Smith\x005 days ago\x00Add user model\n"
	commits := parseCommits(input)
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
	c := commits[0]
	if c.shortHash != "abc1234" {
		t.Errorf("shortHash: got %q", c.shortHash)
	}
	if c.author != "John Doe" {
		t.Errorf("author: got %q", c.author)
	}
	if c.when != "2 days ago" {
		t.Errorf("when: got %q", c.when)
	}
	if c.subject != "Fix login bug" {
		t.Errorf("subject: got %q", c.subject)
	}
	if commits[1].shortHash != "xyz9876" {
		t.Errorf("second shortHash: got %q", commits[1].shortHash)
	}
}

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

// --- parseDiff ---

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

// --- truncate ---

func TestTruncate_Short(t *testing.T) {
	if truncate("hello", 10) != "hello" {
		t.Error("should not truncate")
	}
}

func TestTruncate_Exact(t *testing.T) {
	if truncate("hello", 5) != "hello" {
		t.Error("exact length should not truncate")
	}
}

func TestTruncate_Long(t *testing.T) {
	got := truncate("hello world", 7)
	if got != "hello …" {
		t.Errorf("got %q", got)
	}
}

func TestTruncate_One(t *testing.T) {
	if truncate("hi", 1) != "…" {
		t.Errorf("got %q", truncate("hi", 1))
	}
}

func TestTruncate_Zero(t *testing.T) {
	if truncate("hi", 0) != "" {
		t.Errorf("got %q", truncate("hi", 0))
	}
}

// --- firstWord ---

func TestFirstWord_WithSpace(t *testing.T) {
	if firstWord("John Doe") != "John" {
		t.Errorf("got %q", firstWord("John Doe"))
	}
}

func TestFirstWord_NoSpace(t *testing.T) {
	if firstWord("John") != "John" {
		t.Errorf("got %q", firstWord("John"))
	}
}

// --- cursor navigation ---

func TestMoveCursorDown(t *testing.T) {
	m := model{commits: makeCommits(3), cursor: 0}
	m = moveCursorDown(m)
	if m.cursor != 1 {
		t.Errorf("expected 1, got %d", m.cursor)
	}
}

func TestMoveCursorDown_AtEnd(t *testing.T) {
	m := model{commits: makeCommits(3), cursor: 2}
	m = moveCursorDown(m)
	if m.cursor != 2 {
		t.Errorf("expected 2, got %d", m.cursor)
	}
}

func TestMoveCursorDown_ResetsDiffOffset(t *testing.T) {
	m := model{commits: makeCommits(3), cursor: 0, diffOffset: 10}
	m = moveCursorDown(m)
	if m.diffOffset != 0 {
		t.Error("diffOffset should reset on commit change")
	}
}

func TestMoveCursorUp(t *testing.T) {
	m := model{commits: makeCommits(3), cursor: 2}
	m = moveCursorUp(m)
	if m.cursor != 1 {
		t.Errorf("expected 1, got %d", m.cursor)
	}
}

func TestMoveCursorUp_AtStart(t *testing.T) {
	m := model{commits: makeCommits(3), cursor: 0}
	m = moveCursorUp(m)
	if m.cursor != 0 {
		t.Errorf("expected 0, got %d", m.cursor)
	}
}

func TestMoveCursorUp_ResetsDiffOffset(t *testing.T) {
	m := model{commits: makeCommits(3), cursor: 2, diffOffset: 5}
	m = moveCursorUp(m)
	if m.diffOffset != 0 {
		t.Error("diffOffset should reset on commit change")
	}
}

// --- diff scrolling ---

func TestScrollDiffDown(t *testing.T) {
	m := model{diffLines: makeDiffLines(50), diffOffset: 0, height: 30}
	m = scrollDiffDown(m, 5)
	if m.diffOffset != 5 {
		t.Errorf("expected 5, got %d", m.diffOffset)
	}
}

func TestScrollDiffUp(t *testing.T) {
	m := model{diffLines: makeDiffLines(50), diffOffset: 10, height: 30}
	m = scrollDiffUp(m, 3)
	if m.diffOffset != 7 {
		t.Errorf("expected 7, got %d", m.diffOffset)
	}
}

func TestScrollDiffDown_ClampsToMax(t *testing.T) {
	// height=30 → diffPanelHeight = 30-7 = 23, max = 50-23 = 27
	m := model{diffLines: makeDiffLines(50), diffOffset: 0, height: 30}
	m = scrollDiffDown(m, 1000)
	panelH := diffPanelHeight(m)
	expected := len(m.diffLines) - panelH
	if m.diffOffset != expected {
		t.Errorf("expected clamped to %d, got %d", expected, m.diffOffset)
	}
}

func TestScrollDiffUp_ClampsToZero(t *testing.T) {
	m := model{diffLines: makeDiffLines(50), diffOffset: 5, height: 30}
	m = scrollDiffUp(m, 1000)
	if m.diffOffset != 0 {
		t.Errorf("expected 0, got %d", m.diffOffset)
	}
}

func TestScrollDiffDown_FitsInPanel(t *testing.T) {
	// fewer lines than panel height — can't scroll
	m := model{diffLines: makeDiffLines(5), diffOffset: 0, height: 30}
	m = scrollDiffDown(m, 10)
	if m.diffOffset != 0 {
		t.Errorf("expected 0 when content fits panel, got %d", m.diffOffset)
	}
}

// --- switchPanel ---

func TestSwitchPanel_TowardsDiff(t *testing.T) {
	m := model{focus: panelList}
	m = switchPanel(m)
	if m.focus != panelDiff {
		t.Error("expected panelDiff")
	}
}

func TestSwitchPanel_TowardsList(t *testing.T) {
	m := model{focus: panelDiff}
	m = switchPanel(m)
	if m.focus != panelList {
		t.Error("expected panelList")
	}
}

// --- panel sizing ---

func TestListPanelWidth_Minimum(t *testing.T) {
	if listPanelWidth(40) < 36 {
		t.Error("should be at least 36")
	}
}

func TestListPanelWidth_Maximum(t *testing.T) {
	if listPanelWidth(300) > 52 {
		t.Error("should be at most 52")
	}
}

func TestListPanelWidth_ThirdOfWidth(t *testing.T) {
	w := listPanelWidth(120)
	if w < 36 || w > 52 {
		t.Errorf("unexpected width %d for total=120", w)
	}
}

func TestDiffPanelWidth(t *testing.T) {
	total := 120
	lw := listPanelWidth(total)
	dw := diffPanelWidth(total)
	if lw+dw+1 != total {
		t.Errorf("lw(%d) + dw(%d) + 1 != %d", lw, dw, total)
	}
}

func TestDiffPanelHeight_Normal(t *testing.T) {
	m := model{height: 40}
	h := diffPanelHeight(m)
	if h != 33 { // 40 - 7
		t.Errorf("expected 33, got %d", h)
	}
}

func TestDiffPanelHeight_Minimum(t *testing.T) {
	m := model{height: 5}
	if diffPanelHeight(m) < 5 {
		t.Error("should be at least 5")
	}
}

// --- parseFileItems ---

func TestParseFileItems_Empty(t *testing.T) {
	items := parseFileItems([]diffLine{
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
	items := parseFileItems(lines)
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
	items := parseFileItems(lines)
	if len(items) != 2 {
		t.Fatalf("expected 2, got %d", len(items))
	}
	if items[0].path != "a.go" {
		t.Errorf("first path: got %q", items[0].path)
	}
	if items[0].diffIdx != 0 {
		t.Errorf("first diffIdx: got %d", items[0].diffIdx)
	}
	if items[1].path != "path/to/b.go" {
		t.Errorf("second path: got %q", items[1].path)
	}
	if items[1].diffIdx != 2 {
		t.Errorf("second diffIdx: got %d", items[1].diffIdx)
	}
}

func TestParseFileItems_SkipsNonDiffGit(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "--- a/foo.go"},
		{lineMeta, "+++ b/foo.go"},
		{lineMeta, "diff --git a/bar.go b/bar.go"},
	}
	items := parseFileItems(lines)
	if len(items) != 1 {
		t.Fatalf("expected 1 (only diff --git lines), got %d", len(items))
	}
}

// --- filterCommits ---

func makeNamedCommits() []commit {
	return []commit{
		{shortHash: "aaa1111", author: "John Doe", subject: "Fix login bug"},
		{shortHash: "bbb2222", author: "Jane Smith", subject: "Add user model"},
		{shortHash: "ccc3333", author: "John Doe", subject: "Update README"},
	}
}

func TestFilterCommits_EmptyQuery(t *testing.T) {
	commits := makeNamedCommits()
	result := filterCommits(commits, "")
	if len(result) != 3 {
		t.Errorf("empty query should return all, got %d", len(result))
	}
}

func TestFilterCommits_BySubject(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "login")
	if len(result) != 1 || result[0].subject != "Fix login bug" {
		t.Errorf("expected Fix login bug, got %v", result)
	}
}

func TestFilterCommits_ByAuthor(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "Jane")
	if len(result) != 1 || result[0].author != "Jane Smith" {
		t.Errorf("expected Jane Smith commit, got %v", result)
	}
}

func TestFilterCommits_ByHash(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "bbb")
	if len(result) != 1 || result[0].shortHash != "bbb2222" {
		t.Errorf("expected bbb2222, got %v", result)
	}
}

func TestFilterCommits_CaseInsensitive(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "LOGIN")
	if len(result) != 1 {
		t.Errorf("filter should be case-insensitive, got %d", len(result))
	}
}

func TestFilterCommits_NoMatch(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "zzznomatch")
	if len(result) != 0 {
		t.Errorf("expected 0 results, got %d", len(result))
	}
}

func TestFilterCommits_MultipleMatches(t *testing.T) {
	result := filterCommits(makeNamedCommits(), "John")
	if len(result) != 2 {
		t.Errorf("expected 2 John Doe commits, got %d", len(result))
	}
}

// --- visibleCommits ---

func TestVisibleCommits_NoQuery(t *testing.T) {
	m := model{commits: makeCommits(5)}
	if len(visibleCommits(m)) != 5 {
		t.Error("no query should return all commits")
	}
}

func TestVisibleCommits_WithQuery(t *testing.T) {
	m := model{commits: makeNamedCommits(), query: "login"}
	vc := visibleCommits(m)
	if len(vc) != 1 || vc[0].subject != "Fix login bug" {
		t.Errorf("unexpected result: %v", vc)
	}
}

// --- scrollToDiffLine ---

func TestScrollToDiffLine(t *testing.T) {
	m := model{diffLines: makeDiffLines(50), diffOffset: 0, height: 30}
	m = scrollToDiffLine(m, 10)
	if m.diffOffset != 10 {
		t.Errorf("expected 10, got %d", m.diffOffset)
	}
}

func TestScrollToDiffLine_ClampsToMax(t *testing.T) {
	// height=30 → panelH=23, max=50-23=27
	m := model{diffLines: makeDiffLines(50), height: 30}
	m = scrollToDiffLine(m, 1000)
	panelH := diffPanelHeight(m)
	expected := len(m.diffLines) - panelH
	if m.diffOffset != expected {
		t.Errorf("expected %d, got %d", expected, m.diffOffset)
	}
}

func TestScrollToDiffLine_ClampsToZero(t *testing.T) {
	m := model{diffLines: makeDiffLines(50), height: 30}
	m = scrollToDiffLine(m, -5)
	if m.diffOffset != 0 {
		t.Errorf("expected 0, got %d", m.diffOffset)
	}
}

// --- toggleFileView ---

func TestToggleFileView_Show(t *testing.T) {
	m := model{showFiles: false}
	m = toggleFileView(m)
	if !m.showFiles {
		t.Error("expected showFiles=true")
	}
}

func TestToggleFileView_Hide(t *testing.T) {
	m := model{showFiles: true, fileCursor: 3}
	m = toggleFileView(m)
	if m.showFiles {
		t.Error("expected showFiles=false")
	}
	if m.fileCursor != 0 {
		t.Error("expected fileCursor reset to 0 on hide")
	}
}
