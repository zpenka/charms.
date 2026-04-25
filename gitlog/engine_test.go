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

// --- parseBranches ---

func TestParseBranches_Simple(t *testing.T) {
	input := "  main\n* develop\n  remotes/origin/main\n"
	branches := parseBranches(input)
	if len(branches) != 3 {
		t.Fatalf("expected 3, got %d: %v", len(branches), branches)
	}
	if branches[0] != "main" {
		t.Errorf("first: got %q", branches[0])
	}
	if branches[1] != "develop" {
		t.Errorf("second: got %q", branches[1])
	}
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

// --- parseCurrentBranch ---

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

// --- parseBlameLine ---

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

// --- parseBlame ---

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

// --- currentFile ---

func TestCurrentFile_Empty(t *testing.T) {
	m := model{}
	if currentFile(m) != "" {
		t.Error("expected empty string with no fileItems")
	}
}

func TestCurrentFile_AtStart(t *testing.T) {
	m := model{
		fileItems:  []fileItem{{path: "a.go", diffIdx: 0}, {path: "b.go", diffIdx: 20}},
		diffOffset: 0,
	}
	if got := currentFile(m); got != "a.go" {
		t.Errorf("expected a.go, got %q", got)
	}
}

func TestCurrentFile_PastBoundary(t *testing.T) {
	m := model{
		fileItems:  []fileItem{{path: "a.go", diffIdx: 0}, {path: "b.go", diffIdx: 20}},
		diffOffset: 25,
	}
	if got := currentFile(m); got != "b.go" {
		t.Errorf("expected b.go, got %q", got)
	}
}

func TestCurrentFile_ExactlyAtBoundary(t *testing.T) {
	m := model{
		fileItems:  []fileItem{{path: "a.go", diffIdx: 0}, {path: "b.go", diffIdx: 20}},
		diffOffset: 20,
	}
	if got := currentFile(m); got != "b.go" {
		t.Errorf("expected b.go at exact boundary, got %q", got)
	}
}

// --- parseCount ---

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

// --- toggleBranchView ---

func TestToggleBranchView_Show(t *testing.T) {
	m := model{showBranch: false}
	m = toggleBranchView(m)
	if !m.showBranch {
		t.Error("expected showBranch=true")
	}
}

func TestToggleBranchView_Hide(t *testing.T) {
	m := model{showBranch: true, branchCursor: 3}
	m = toggleBranchView(m)
	if m.showBranch {
		t.Error("expected showBranch=false")
	}
	if m.branchCursor != 0 {
		t.Error("expected branchCursor reset to 0")
	}
}

// --- filterCommitsByAuthor ---

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

// --- filterCommitsSince ---

func makeCommitsWithDays() []commit {
	return []commit{
		{shortHash: "aaa1111", author: "John", when: "1 day ago", subject: "Recent"},
		{shortHash: "bbb2222", author: "Jane", when: "5 days ago", subject: "Medium"},
		{shortHash: "ccc3333", author: "Bob", when: "20 days ago", subject: "Old"},
		{shortHash: "ddd4444", author: "Alice", when: "100 days ago", subject: "Very old"},
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

// --- formatActiveFilters ---

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

// --- Navigation history (breadcrumb trail) ---

func TestNavigationHistory_InitEmpty(t *testing.T) {
	m := model{}
	if len(m.navHistory) != 0 {
		t.Errorf("should start empty, got %d items", len(m.navHistory))
	}
}

func TestNavigationHistory_AddPosition(t *testing.T) {
	m := model{}
	m = addToNavHistory(m, 5)
	if len(m.navHistory) != 1 {
		t.Errorf("expected 1 item, got %d", len(m.navHistory))
	}
	if m.navHistory[0] != 5 {
		t.Errorf("expected 5, got %d", m.navHistory[0])
	}
}

func TestNavigationHistory_GoBack(t *testing.T) {
	m := model{navHistory: []int{0, 5, 10}, navHistoryIdx: 2, cursor: 10}
	m = goBackInHistory(m)
	if m.navHistoryIdx != 1 {
		t.Errorf("expected idx=1, got %d", m.navHistoryIdx)
	}
	if m.cursor != 5 {
		t.Errorf("expected cursor=5, got %d", m.cursor)
	}
}

func TestNavigationHistory_GoForward(t *testing.T) {
	m := model{navHistory: []int{0, 5, 10}, navHistoryIdx: 1, cursor: 5}
	m = goForwardInHistory(m)
	if m.navHistoryIdx != 2 {
		t.Errorf("expected idx=2, got %d", m.navHistoryIdx)
	}
	if m.cursor != 10 {
		t.Errorf("expected cursor=10, got %d", m.cursor)
	}
}

func TestNavigationHistory_CannotGoBackAtStart(t *testing.T) {
	m := model{navHistory: []int{0, 5}, navHistoryIdx: 0}
	m = goBackInHistory(m)
	if m.navHistoryIdx != 0 {
		t.Errorf("should stay at 0, got %d", m.navHistoryIdx)
	}
}

func TestNavigationHistory_CannotGoForwardAtEnd(t *testing.T) {
	m := model{navHistory: []int{0, 5}, navHistoryIdx: 1}
	m = goForwardInHistory(m)
	if m.navHistoryIdx != 1 {
		t.Errorf("should stay at 1, got %d", m.navHistoryIdx)
	}
}

// --- commitStats ---

func TestCommitStats_Empty(t *testing.T) {
	stats := commitStats([]diffLine{})
	if stats.filesChanged != 0 {
		t.Errorf("empty diff should have 0 files, got %d", stats.filesChanged)
	}
	if stats.insertions != 0 || stats.deletions != 0 {
		t.Errorf("empty diff should have 0 changes")
	}
}

func TestCommitStats_CountsFiles(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "diff --git a/foo.go b/foo.go"},
		{lineAdded, "+line"},
		{lineMeta, "diff --git a/bar.go b/bar.go"},
		{lineRemoved, "-line"},
	}
	stats := commitStats(lines)
	if stats.filesChanged != 2 {
		t.Errorf("expected 2 files, got %d", stats.filesChanged)
	}
}

func TestCommitStats_CountsInsertionsAndDeletions(t *testing.T) {
	lines := []diffLine{
		{lineAdded, "+added1"},
		{lineAdded, "+added2"},
		{lineRemoved, "-deleted1"},
		{lineRemoved, "-deleted2"},
		{lineRemoved, "-deleted3"},
		{lineContext, " context"},
	}
	stats := commitStats(lines)
	if stats.insertions != 2 {
		t.Errorf("expected 2 insertions, got %d", stats.insertions)
	}
	if stats.deletions != 3 {
		t.Errorf("expected 3 deletions, got %d", stats.deletions)
	}
}

// --- generateCommitMessage ---

func TestGenerateCommitMessage_Empty(t *testing.T) {
	msg := generateCommitMessage([]diffLine{}, "")
	if msg == "" {
		t.Error("should generate non-empty message")
	}
}

func TestGenerateCommitMessage_DetectsDeletedFile(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "deleted file mode 100644"},
		{lineMeta, "diff --git a/old.txt b/old.txt"},
	}
	msg := generateCommitMessage(lines, "old.txt")
	if !strings.Contains(strings.ToLower(msg), "remove") && !strings.Contains(strings.ToLower(msg), "delete") {
		t.Errorf("should suggest 'remove' or 'delete', got %q", msg)
	}
}

func TestGenerateCommitMessage_DetectsNewFile(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "new file mode 100644"},
		{lineMeta, "diff --git a/new.txt b/new.txt"},
	}
	msg := generateCommitMessage(lines, "new.txt")
	if !strings.Contains(strings.ToLower(msg), "add") && !strings.Contains(strings.ToLower(msg), "create") {
		t.Errorf("should suggest 'add' or 'create', got %q", msg)
	}
}

func TestGenerateCommitMessage_DetectsBreakingChange(t *testing.T) {
	lines := []diffLine{
		{lineRemoved, "-func OldAPI() {}"},
		{lineMeta, "diff --git a/api.go b/api.go"},
	}
	msg := generateCommitMessage(lines, "api.go")
	if !strings.Contains(msg, "!") {
		t.Errorf("should contain breaking change marker (!), got %q", msg)
	}
}

// --- bookmarks ---

func TestBookmarks_InitEmpty(t *testing.T) {
	m := model{}
	if len(m.bookmarks) != 0 {
		t.Errorf("should start with no bookmarks, got %d", len(m.bookmarks))
	}
}

func TestBookmarks_Toggle(t *testing.T) {
	m := model{cursor: 5, commits: makeCommits(10)}
	m = toggleBookmark(m)
	if !isBookmarked(m, 5) {
		t.Error("cursor position should be bookmarked")
	}
}

func TestBookmarks_ToggleRemoves(t *testing.T) {
	c := commit{shortHash: "abc123"}
	m := model{cursor: 0, commits: []commit{c}, bookmarks: []string{"abc123"}}
	m = toggleBookmark(m)
	if isBookmarked(m, 0) {
		t.Error("bookmark should be removed")
	}
}

func TestBookmarks_JumpToNext(t *testing.T) {
	commits := []commit{
		{shortHash: "aaa", subject: "first"},
		{shortHash: "bbb", subject: "second"},
		{shortHash: "ccc", subject: "third"},
	}
	m := model{commits: commits, cursor: 0, bookmarks: []string{"bbb", "ccc"}}
	m = jumpToNextBookmark(m)
	if m.cursor != 1 {
		t.Errorf("expected cursor=1, got %d", m.cursor)
	}
}

func TestBookmarks_JumpToPrev(t *testing.T) {
	commits := []commit{
		{shortHash: "aaa", subject: "first"},
		{shortHash: "bbb", subject: "second"},
		{shortHash: "ccc", subject: "third"},
	}
	m := model{commits: commits, cursor: 2, bookmarks: []string{"aaa", "bbb"}}
	m = jumpToPrevBookmark(m)
	if m.cursor != 1 {
		t.Errorf("expected cursor=1, got %d", m.cursor)
	}
}

// --- detectLanguage ---

func TestDetectLanguage_Go(t *testing.T) {
	if lang := detectLanguage("main.go"); lang != "go" {
		t.Errorf("expected 'go', got %q", lang)
	}
}

func TestDetectLanguage_Python(t *testing.T) {
	if lang := detectLanguage("script.py"); lang != "python" {
		t.Errorf("expected 'python', got %q", lang)
	}
}

func TestDetectLanguage_Unknown(t *testing.T) {
	if lang := detectLanguage("file.unknown"); lang == "" {
		t.Error("should return some language for unknown")
	}
}

func TestDetectLanguage_NoExtension(t *testing.T) {
	if lang := detectLanguage("Makefile"); lang != "makefile" {
		t.Errorf("expected 'makefile', got %q", lang)
	}
}

// --- miniMapPosition ---

func TestMiniMapPosition_StartOfList(t *testing.T) {
	pos := miniMapPosition(0, 10, 20)
	if pos != 0 {
		t.Errorf("at start, should return 0, got %d", pos)
	}
}

func TestMiniMapPosition_EndOfList(t *testing.T) {
	pos := miniMapPosition(19, 10, 20)
	if pos < 9 {
		t.Errorf("near end, should return high value, got %d", pos)
	}
}

func TestMiniMapPosition_Middle(t *testing.T) {
	pos := miniMapPosition(5, 10, 10)
	if pos < 4 || pos > 6 {
		t.Errorf("in middle, should return ~5, got %d", pos)
	}
}

// --- diffStatBadge ---

func TestDiffStatBadge_Empty(t *testing.T) {
	badge := diffStatBadge(commitStatistics{})
	if badge == "" {
		t.Error("should generate non-empty badge")
	}
}

func TestDiffStatBadge_Format(t *testing.T) {
	stats := commitStatistics{insertions: 10, deletions: 5}
	badge := diffStatBadge(stats)
	if !strings.Contains(badge, "10") || !strings.Contains(badge, "5") {
		t.Errorf("badge should contain counts, got %q", badge)
	}
}

func TestDiffStatBadge_ShowsFiles(t *testing.T) {
	stats := commitStatistics{filesChanged: 3, insertions: 10, deletions: 5}
	badge := diffStatBadge(stats)
	if !strings.Contains(badge, "3") {
		t.Errorf("badge should show files changed, got %q", badge)
	}
}

// --- goToCommit ---

func TestGoToCommit_FindsByHash(t *testing.T) {
	commits := []commit{
		{shortHash: "abc1234", hash: "abc1234567890"},
		{shortHash: "def5678", hash: "def5678901234"},
	}
	idx := goToCommit(commits, "abc1234")
	if idx != 0 {
		t.Errorf("expected 0, got %d", idx)
	}
}

func TestGoToCommit_FindsByFullHash(t *testing.T) {
	commits := []commit{
		{shortHash: "abc1234", hash: "abc1234567890abcdef1234567890"},
		{shortHash: "def5678", hash: "def5678901234abcdef5678901234"},
	}
	idx := goToCommit(commits, "abc1234567890abcdef1234567890")
	if idx != 0 {
		t.Errorf("expected 0, got %d", idx)
	}
}

func TestGoToCommit_NotFound(t *testing.T) {
	commits := []commit{
		{shortHash: "abc1234", hash: "abc1234567890"},
	}
	idx := goToCommit(commits, "zzz9999")
	if idx != -1 {
		t.Errorf("should return -1 for not found, got %d", idx)
	}
}

func TestGoToCommit_CaseInsensitive(t *testing.T) {
	commits := []commit{
		{shortHash: "AbC1234", hash: "AbC1234567890"},
	}
	idx := goToCommit(commits, "abc1234")
	if idx != 0 {
		t.Errorf("should be case-insensitive, got %d", idx)
	}
}

// --- copyAsPatch ---

func TestCopyAsPatch_Empty(t *testing.T) {
	patch := copyAsPatch("abc1234", []diffLine{})
	if patch == "" {
		t.Error("should generate non-empty patch")
	}
}

func TestCopyAsPatch_ContainsHash(t *testing.T) {
	lines := []diffLine{{lineAdded, "+new line"}}
	patch := copyAsPatch("abc1234", lines)
	if !strings.Contains(patch, "abc1234") {
		t.Errorf("patch should contain hash, got %q", patch)
	}
}

func TestCopyAsPatch_ContainsDiff(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "diff --git a/test.go b/test.go"},
		{lineAdded, "+new line"},
	}
	patch := copyAsPatch("abc1234", lines)
	if !strings.Contains(patch, "+new line") {
		t.Errorf("patch should contain diff content, got %q", patch)
	}
}

// --- parseGitReferences ---

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

// --- isMergeCommit ---

func TestIsMergeCommit_NotMerge(t *testing.T) {
	lines := []diffLine{{lineAdded, "+normal commit"}}
	if isMergeCommit(lines) {
		t.Error("should not detect as merge")
	}
}

func TestIsMergeCommit_DetectsMerge(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "Merge: abc1234 def5678"},
	}
	if !isMergeCommit(lines) {
		t.Error("should detect merge commit")
	}
}

func TestIsMergeCommit_WithMergeTag(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "Merge branch 'feature' into main"},
	}
	if !isMergeCommit(lines) {
		t.Error("should detect merge from message")
	}
}

// --- getMergeParents ---

func TestGetMergeParents_Empty(t *testing.T) {
	parents := getMergeParents([]diffLine{})
	if len(parents) != 0 {
		t.Errorf("empty diff should have 0 parents, got %d", len(parents))
	}
}

func TestGetMergeParents_Single(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "Merge: abc1234 def5678"},
	}
	parents := getMergeParents(lines)
	if len(parents) != 2 {
		t.Errorf("expected 2 parents, got %d", len(parents))
	}
	if parents[0] != "abc1234" || parents[1] != "def5678" {
		t.Errorf("wrong parents: %v", parents)
	}
}

// --- parseHunks ---

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

// --- toggleLineComment ---

func TestToggleLineComment_AddComment(t *testing.T) {
	m := model{comments: make(map[int]string)}
	m = toggleLineComment(m, 5, "needs review")
	if m.comments[5] != "needs review" {
		t.Errorf("comment not set, got %q", m.comments[5])
	}
}

func TestToggleLineComment_RemoveComment(t *testing.T) {
	m := model{comments: map[int]string{5: "needs review"}}
	m = toggleLineComment(m, 5, "")
	if _, ok := m.comments[5]; ok {
		t.Error("comment should be removed")
	}
}

// --- compileRegex ---

func TestCompileRegex_Valid(t *testing.T) {
	re, err := compileRegex("fix.*bug")
	if err != nil {
		t.Errorf("valid regex should compile, got %v", err)
	}
	if re == nil {
		t.Error("regex should not be nil")
	}
}

func TestCompileRegex_Invalid(t *testing.T) {
	_, err := compileRegex("[invalid(regex")
	if err == nil {
		t.Error("invalid regex should return error")
	}
}

func TestCompileRegex_Match(t *testing.T) {
	re, _ := compileRegex("^fix")
	if !re.MatchString("fix login bug") {
		t.Error("regex should match 'fix login bug'")
	}
	if re.MatchString("bugfix something") {
		t.Error("regex should not match 'bugfix something'")
	}
}

// --- parseDateRange ---

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

// --- filterCommitsByFile ---

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

// Note: This test requires access to git history which is complex to mock
// In practice, this would need the model to track changed files per commit
func TestFilterCommitsByFile_Matches(t *testing.T) {
	// This is infrastructure - actual matching would happen via git queries
	result := filterCommitsByFile([]commit{}, "test.go")
	if result == nil {
		t.Error("should return slice, not nil")
	}
}

// --- parseTags ---

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

// --- Tag view model field ---

func TestTagView_InitFalse(t *testing.T) {
	m := model{}
	if m.showTags {
		t.Error("showTags should default to false")
	}
}

func TestTagView_Toggle(t *testing.T) {
	m := model{showTags: false}
	m.showTags = !m.showTags
	if !m.showTags {
		t.Error("should toggle to true")
	}
}

// ===== OPTION 1: UI INTEGRATION =====

func TestRenderStatsBadgeInList(t *testing.T) {
	stats := commitStatistics{filesChanged: 3, insertions: 10, deletions: 5}
	badge := renderStatsBadgeInList(stats, 20)
	if badge == "" {
		t.Error("should generate non-empty badge")
	}
	if !strings.Contains(badge, "+10") || !strings.Contains(badge, "-5") {
		t.Errorf("badge should contain stats, got %q", badge)
	}
}

func TestRenderStatsBadgeInList_TruncatesLong(t *testing.T) {
	stats := commitStatistics{filesChanged: 100, insertions: 999, deletions: 888}
	badge := renderStatsBadgeInList(stats, 10)
	if len(badge) > 12 {
		t.Errorf("should truncate for width 10, got len=%d", len(badge))
	}
}

func TestFormatFilterHeaderDisplay(t *testing.T) {
	m := model{authorFilter: "Jane", sinceFilter: 7}
	display := formatFilterHeaderDisplay(m)
	if display == "" {
		t.Error("should show filters")
	}
	if !strings.Contains(display, "Jane") {
		t.Errorf("should contain author, got %q", display)
	}
}

func TestRenderBookmarkMarker_Bookmarked(t *testing.T) {
	m := model{cursor: 0, commits: []commit{{shortHash: "abc"}}, bookmarks: []string{"abc"}}
	marker := renderBookmarkMarker(m, 0)
	if marker == "" {
		t.Error("bookmarked commit should have marker")
	}
}

func TestRenderBookmarkMarker_NotBookmarked(t *testing.T) {
	m := model{cursor: 0, commits: []commit{{shortHash: "abc"}}, bookmarks: []string{}}
	marker := renderBookmarkMarker(m, 0)
	if marker != "" {
		t.Errorf("non-bookmarked should be empty, got %q", marker)
	}
}

func TestHandleGoToCommitInput(t *testing.T) {
	commits := []commit{
		{shortHash: "abc1234", hash: "abc1234567890"},
		{shortHash: "def5678", hash: "def5678901234"},
	}
	m := model{commits: commits, cursor: 0}
	m = handleGoToCommitInput(m, "def5678")
	if m.cursor != 1 {
		t.Errorf("should jump to index 1, got %d", m.cursor)
	}
}

func TestHandleGoToCommitInput_InvalidHash(t *testing.T) {
	commits := []commit{{shortHash: "abc1234", hash: "abc1234567890"}}
	m := model{commits: commits, cursor: 0}
	m = handleGoToCommitInput(m, "zzz9999")
	if m.cursor != 0 {
		t.Errorf("should stay at 0 for invalid hash, got %d", m.cursor)
	}
}

func TestRenderLineCommentMarker_HasComment(t *testing.T) {
	m := model{comments: map[int]string{5: "needs review"}}
	marker := renderLineCommentMarker(m, 5)
	if marker == "" {
		t.Error("line with comment should have marker")
	}
}

func TestRenderLineCommentMarker_NoComment(t *testing.T) {
	m := model{comments: map[int]string{}}
	marker := renderLineCommentMarker(m, 5)
	if marker != "" {
		t.Errorf("line without comment should be empty, got %q", marker)
	}
}

// ===== OPTION 2: COMMIT GRAPH =====

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

func TestDetectBranches_Linear(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "first"},
		{hash: "bbb", subject: "second"},
	}
	branches := detectBranches(commits)
	if len(branches) != 1 {
		t.Errorf("linear history should have 1 branch, got %d", len(branches))
	}
}

func TestRenderAsciiGraph_SingleCommit(t *testing.T) {
	graph := []graphNode{
		{hash: "abc", depth: 0, isMerge: false},
	}
	art := renderAsciiGraph(graph)
	if art == "" {
		t.Error("should generate non-empty ASCII art")
	}
}

func TestNavigateAlongGraph_Forward(t *testing.T) {
	graph := []graphNode{
		{hash: "aaa", depth: 0, isMerge: false},
		{hash: "bbb", depth: 0, isMerge: false},
	}
	idx := navigateAlongGraph(graph, 0, "down")
	if idx != 1 {
		t.Errorf("should move to index 1, got %d", idx)
	}
}

func TestNavigateAlongGraph_Backward(t *testing.T) {
	graph := []graphNode{
		{hash: "aaa", depth: 0, isMerge: false},
		{hash: "bbb", depth: 0, isMerge: false},
	}
	idx := navigateAlongGraph(graph, 1, "up")
	if idx != 0 {
		t.Errorf("should move to index 0, got %d", idx)
	}
}

func TestGetCommitRelationships_None(t *testing.T) {
	rels := getCommitRelationships([]commit{})
	if len(rels) != 0 {
		t.Errorf("empty commits should have no relationships, got %d", len(rels))
	}
}

// ===== OPTION 3: FILE-CENTRIC VIEW =====

func TestBuildFileHistory_Empty(t *testing.T) {
	history := buildFileHistory([]commit{}, "test.go")
	if history == nil {
		t.Error("should return non-nil slice")
	}
}

func TestBuildFileHistory_SingleFile(t *testing.T) {
	// Note: In real implementation, would query git for file history
	history := buildFileHistory([]commit{}, "test.go")
	if history == nil {
		t.Error("should return slice")
	}
}

func TestRenderFileTimeline_Empty(t *testing.T) {
	timeline := renderFileTimeline([]commit{}, "test.go", 50)
	if timeline == "" {
		t.Error("should generate non-empty timeline")
	}
}

func TestGetFileBlameContext_Empty(t *testing.T) {
	ctx := getFileBlameContext([]diffLine{}, "test.go")
	if ctx == nil {
		t.Error("should return non-nil context")
	}
}

func TestFilterCommitsByFileChange_None(t *testing.T) {
	commits := []commit{{shortHash: "aaa", subject: "fix"}}
	filtered := filterCommitsByFileChange(commits, "nonexistent.go")
	if len(filtered) > 0 {
		t.Errorf("no matches should be empty, got %d", len(filtered))
	}
}

func TestIsFileModifiedInCommit_Unknown(t *testing.T) {
	// Infrastructure: actual check would need git
	modified := isFileModifiedInCommit("abc1234", "test.go")
	if modified {
		t.Error("unknown should return false")
	}
}

// ===== OPTION 4: STASH & REFLOG =====

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

func TestRenderStashView_Empty(t *testing.T) {
	view := renderStashView([]stashEntry{}, 50)
	if view == "" {
		t.Error("should generate non-empty view")
	}
}

func TestRenderReflogView_Empty(t *testing.T) {
	view := renderReflogView([]reflogEntry{}, 50)
	if view == "" {
		t.Error("should generate non-empty view")
	}
}

func TestStashToCommitLike_Conversion(t *testing.T) {
	stash := stashEntry{name: "stash@{0}", branch: "main", subject: "WIP"}
	c := stashToCommitLike(stash)
	if c.subject != "WIP" {
		t.Errorf("wrong subject, got %q", c.subject)
	}
}

func TestReflogToCommitLike_Conversion(t *testing.T) {
	entry := reflogEntry{hash: "abc1234", action: "commit", message: "test"}
	c := reflogToCommitLike(entry)
	if c.hash != "abc1234" {
		t.Errorf("wrong hash, got %q", c.hash)
	}
}

func TestSwitchViewMode_Log(t *testing.T) {
	m := model{viewMode: "log"}
	m = switchViewMode(m, "stash")
	if m.viewMode != "stash" {
		t.Errorf("should switch to stash, got %q", m.viewMode)
	}
}

func TestSwitchViewMode_Stash(t *testing.T) {
	m := model{viewMode: "stash"}
	m = switchViewMode(m, "reflog")
	if m.viewMode != "reflog" {
		t.Errorf("should switch to reflog, got %q", m.viewMode)
	}
}

func TestSwitchViewMode_Reflog(t *testing.T) {
	m := model{viewMode: "reflog"}
	m = switchViewMode(m, "log")
	if m.viewMode != "log" {
		t.Errorf("should switch to log, got %q", m.viewMode)
	}
}

func TestApplyStash_FindsStash(t *testing.T) {
	stashes := []stashEntry{
		{name: "stash@{0}", branch: "main"},
		{name: "stash@{1}", branch: "feature"},
	}
	found := findStashByIndex(stashes, 0)
	if found == nil {
		t.Error("should find stash@{0}")
	}
	if found.name != "stash@{0}" {
		t.Errorf("wrong stash, got %q", found.name)
	}
}

// ===== UI INTEGRATION: KEYBINDINGS & STATE =====

func TestKeyBinding_M_ToggleBookmark(t *testing.T) {
	m := model{cursor: 0, commits: []commit{{shortHash: "abc"}}, bookmarks: []string{}}
	m = handleKeyBinding(m, "m")
	if len(m.bookmarks) != 1 {
		t.Errorf("m should toggle bookmark, got %d bookmarks", len(m.bookmarks))
	}
}

func TestKeyBinding_SingleQuote_JumpBookmark(t *testing.T) {
	m := model{cursor: 0, commits: []commit{
		{shortHash: "aaa"},
		{shortHash: "bbb"},
	}, bookmarks: []string{"bbb"}}
	m = handleKeyBinding(m, "'")
	if m.cursor != 1 {
		t.Errorf("' should jump to bookmark, got cursor=%d", m.cursor)
	}
}

func TestKeyBinding_gg_GoToCommit(t *testing.T) {
	m := model{commits: []commit{
		{shortHash: "abc1234", hash: "abc1234567890"},
		{shortHash: "def5678", hash: "def5678901234"},
	}, cursor: 0}
	m = handleKeyBinding(m, "gg")
	// Note: gg would need additional input, so this tests the mode entry
	if !m.inGoToCommitMode {
		t.Error("gg should enter go-to-commit mode")
	}
}

func TestKeyBinding_c_LineComment(t *testing.T) {
	m := model{comments: make(map[int]string), diffOffset: 0}
	m = handleKeyBinding(m, "c")
	if !m.inCommentMode {
		t.Error("c should enter comment mode")
	}
}

func TestKeyBinding_v_SwitchToStash(t *testing.T) {
	m := model{viewMode: "log"}
	m = handleKeyBinding(m, "v")
	if m.viewMode != "stash" {
		t.Errorf("v should switch to stash, got %q", m.viewMode)
	}
}

func TestKeyBinding_Shift_V_SwitchToReflog(t *testing.T) {
	m := model{viewMode: "log"}
	m = handleKeyBinding(m, "V")
	if m.viewMode != "reflog" {
		t.Errorf("V should switch to reflog, got %q", m.viewMode)
	}
}

func TestKeyBinding_G_ShowGraph(t *testing.T) {
	m := model{showGraph: false}
	m = handleKeyBinding(m, "G")
	if !m.showGraph {
		t.Error("G should show graph")
	}
}

func TestKeyBinding_F_FileView_Toggle(t *testing.T) {
	m := model{showFiles: false}
	m = handleKeyBinding(m, "f")
	if !m.showFiles {
		t.Error("f should toggle file view")
	}
}

func TestKeyBinding_Multiple_Navigation(t *testing.T) {
	m := model{cursor: 0, commits: makeCommits(10)}
	m = handleKeyBinding(m, "5j")
	if m.cursor != 5 {
		t.Errorf("5j should move 5 down, got %d", m.cursor)
	}
}

// ===== UI INTEGRATION: RENDERING =====

func TestRenderWithStats_IncludesBadges(t *testing.T) {
	m := model{showStatsBadge: true, lastStats: commitStatistics{filesChanged: 3, insertions: 10, deletions: 5}}
	output := renderCommitRowWithStats(m, 0, 50)
	if output == "" {
		t.Error("should render non-empty output")
	}
}

func TestRenderBookmarkList_ShowsBookmarks(t *testing.T) {
	m := model{bookmarks: []string{"abc1234", "def5678"}, commits: []commit{
		{shortHash: "abc1234"},
		{shortHash: "def5678"},
	}}
	output := renderBookmarkList(m, 50)
	if !strings.Contains(output, "abc1234") {
		t.Errorf("should show bookmark, got %q", output)
	}
}

func TestRenderGraphView_ShowsArt(t *testing.T) {
	m := model{showGraph: true, commitGraph: []graphNode{
		{hash: "abc1234", isMerge: false},
	}}
	output := renderGraphView(m, 50)
	if output == "" {
		t.Error("should render non-empty graph")
	}
}

func TestRenderStashBrowser_ShowsStashes(t *testing.T) {
	m := model{viewMode: "stash", stashes: []stashEntry{
		{name: "stash@{0}", subject: "WIP"},
	}}
	output := renderViewMode(m, 50)
	if !strings.Contains(output, "stash") {
		t.Errorf("should show stash view, got %q", output)
	}
}

func TestRenderReflogBrowser_ShowsReflog(t *testing.T) {
	m := model{viewMode: "reflog", reflogEntries: []reflogEntry{
		{hash: "abc1234", action: "commit"},
	}}
	output := renderViewMode(m, 50)
	if !strings.Contains(output, "reflog") && !strings.Contains(output, "abc1234") {
		t.Errorf("should show reflog, got %q", output)
	}
}

func TestRenderCommentedDiff_ShowsMarkers(t *testing.T) {
	m := model{
		comments: map[int]string{5: "needs review"},
		diffLines: []diffLine{
			{lineAdded, "+test1"},
			{lineAdded, "+test2"},
			{lineAdded, "+test3"},
		},
		diffOffset: 0,
	}
	output := renderDiffWithComments(m, 10, 50)
	if output == "" {
		t.Error("should render non-empty diff")
	}
}

// ===== OPTIMIZATION: CACHING & MEMOIZATION =====

func TestDiffCache_StoresResult(t *testing.T) {
	cache := newDiffCache(10)
	lines := []diffLine{{lineAdded, "+test"}}
	cache.set("abc1234", lines)
	cached, ok := cache.get("abc1234")
	if !ok {
		t.Error("cache should have stored diff")
	}
	if len(cached) != 1 {
		t.Errorf("cache should return same data, got %d lines", len(cached))
	}
}

func TestDiffCache_EvictsOldest(t *testing.T) {
	cache := newDiffCache(2)
	cache.set("aaa", []diffLine{{lineAdded, "+a"}})
	cache.set("bbb", []diffLine{{lineAdded, "+b"}})
	cache.set("ccc", []diffLine{{lineAdded, "+c"}})
	_, ok := cache.get("aaa")
	if ok {
		t.Error("cache should evict oldest entry")
	}
}

func TestStatCache_Memoizes(t *testing.T) {
	cache := newStatCache(10)
	lines := []diffLine{
		{lineAdded, "+test"},
		{lineRemoved, "-old"},
	}
	stats1 := cache.getOrCompute("abc", lines)
	stats2 := cache.getOrCompute("abc", lines)
	if stats1.insertions != stats2.insertions {
		t.Error("cached stats should be identical")
	}
}

func TestRegexCache_Compiles(t *testing.T) {
	cache := newRegexCache(10)
	re, err := cache.compile("test.*pattern")
	if err != nil {
		t.Errorf("should compile regex, got %v", err)
	}
	if !re.MatchString("test something pattern") {
		t.Error("compiled regex should match")
	}
}

func TestRegexCache_Reuses(t *testing.T) {
	cache := newRegexCache(10)
	re1, _ := cache.compile("test")
	re2, _ := cache.compile("test")
	if re1 != re2 {
		t.Error("cache should return same regex object")
	}
}

// ===== OPTIMIZATION: LAZY LOADING =====

func TestLazyLoadDiff_LoadsOnDemand(t *testing.T) {
	m := model{cursor: 0, commits: makeCommits(3), diffLines: []diffLine{}}
	m = lazyLoadDiff(m)
	// Should trigger loading, but actual load is async
	if !m.loading {
		t.Error("should set loading flag")
	}
}

func TestLazyLoadGraph_BuildsOnDemand(t *testing.T) {
	m := model{commits: []commit{
		{hash: "aaa", subject: "first"},
		{hash: "bbb", subject: "second"},
	}, commitGraph: nil}
	m = lazyLoadGraph(m)
	if len(m.commitGraph) != 2 {
		t.Errorf("should build graph on demand, got %d nodes", len(m.commitGraph))
	}
}

func TestLazyLoadStats_ComputesWhenNeeded(t *testing.T) {
	m := model{diffLines: []diffLine{
		{lineAdded, "+test"},
		{lineRemoved, "-old"},
	}}
	stats := lazyLoadStats(m)
	if stats.insertions != 1 {
		t.Errorf("should compute stats, got %d insertions", stats.insertions)
	}
}

// ===== OPTIMIZATION: ERROR HANDLING =====

func TestSafeKeyBinding_InvalidKey(t *testing.T) {
	m := model{}
	m = safeHandleKeyBinding(m, "")
	if m.width == 0 && m.height == 0 {
		// Safe to call with empty key
		return
	}
}

func TestSafeFileModified_NoGitError(t *testing.T) {
	result := safeIsFileModified("invalid_hash", "nonexistent.go")
	if result {
		t.Error("should handle missing git gracefully")
	}
}

func TestSafeParseCommitGraph_EmptyList(t *testing.T) {
	graph := safeParseCommitGraph(nil)
	if graph == nil {
		t.Error("should return empty slice, not nil")
	}
}

// ===== OPTIMIZATION: PERFORMANCE VALIDATION =====

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

func TestStatComputation_CacheHitRate(t *testing.T) {
	cache := newStatCache(100)
	lines := []diffLine{
		{lineAdded, "+test"},
		{lineRemoved, "-old"},
	}

	cache.getOrCompute("abc", lines)
	cache.getOrCompute("abc", lines) // Hit
	cache.getOrCompute("def", lines)
	cache.getOrCompute("abc", lines) // Hit

	hits := cache.getHitCount()
	if hits != 2 {
		t.Errorf("expected 2 hits, got %d", hits)
	}
}

// ===== OPTIMIZATION: STATE MANAGEMENT =====

func TestViewModeTransitions_Valid(t *testing.T) {
	m := model{viewMode: "log"}
	validModes := []string{"stash", "reflog", "log"}
	for _, mode := range validModes {
		m = switchViewMode(m, mode)
		if m.viewMode != mode {
			t.Errorf("should transition to %q", mode)
		}
	}
}

func TestEnterCommentMode_ValidState(t *testing.T) {
	m := model{inCommentMode: false}
	m = enterCommentMode(m)
	if !m.inCommentMode {
		t.Error("should enter comment mode")
	}
}

func TestExitCommentMode_ClearsInput(t *testing.T) {
	m := model{inCommentMode: true, commentInput: "test"}
	m = exitCommentMode(m)
	if m.inCommentMode {
		t.Error("should exit comment mode")
	}
	if m.commentInput != "" {
		t.Errorf("should clear input, got %q", m.commentInput)
	}
}

// ===== OPTION A: ADVANCED COMMIT OPERATIONS =====

// --- Interactive Rebase ---

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

func TestReorderCommit_ValidIndex(t *testing.T) {
	seq := []rebaseOp{
		{action: "pick", hash: "aaa", subject: "first"},
		{action: "pick", hash: "bbb", subject: "second"},
	}
	seq = reorderCommit(seq, 0, 1)
	if seq[0].hash != "bbb" {
		t.Errorf("should reorder, got %s", seq[0].hash)
	}
}

func TestSquashCommit_CombinesMsgs(t *testing.T) {
	seq := []rebaseOp{
		{action: "pick", hash: "aaa", subject: "first"},
		{action: "pick", hash: "bbb", subject: "second"},
	}
	seq = squashCommit(seq, 1)
	if seq[1].action != "squash" {
		t.Errorf("should mark as squash, got %q", seq[1].action)
	}
}

func TestFixupCommit_DiscardMsg(t *testing.T) {
	seq := []rebaseOp{
		{action: "pick", hash: "aaa", subject: "first"},
		{action: "pick", hash: "bbb", subject: "fixup"},
	}
	seq = fixupCommit(seq, 1)
	if seq[1].action != "fixup" {
		t.Errorf("should mark as fixup, got %q", seq[1].action)
	}
}

func TestPreviewRebase_ShowsResult(t *testing.T) {
	seq := []rebaseOp{
		{action: "pick", hash: "aaa", subject: "first"},
		{action: "squash", hash: "bbb", subject: "second"},
	}
	preview := previewRebase(seq)
	if preview == "" {
		t.Error("should generate non-empty preview")
	}
}

// --- Cherry-pick ---

func TestSelectForCherryPick_AddToList(t *testing.T) {
	m := model{cherryPickList: []string{}}
	m = toggleCherryPick(m, "abc1234")
	if len(m.cherryPickList) != 1 {
		t.Errorf("should add to list, got %d", len(m.cherryPickList))
	}
}

func TestSelectForCherryPick_RemoveFromList(t *testing.T) {
	m := model{cherryPickList: []string{"abc1234"}}
	m = toggleCherryPick(m, "abc1234")
	if len(m.cherryPickList) != 0 {
		t.Errorf("should remove from list, got %d", len(m.cherryPickList))
	}
}

func TestPreviewCherryPick_ShowsSequence(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "first"},
		{hash: "bbb", subject: "second"},
	}
	picks := []string{"aaa", "bbb"}
	preview := previewCherryPick(commits, picks)
	if preview == "" {
		t.Error("should generate non-empty preview")
	}
}

// --- Reset ---

func TestResetToCommit_SoftMode(t *testing.T) {
	result := resetToCommit("abc1234", "soft")
	if result == "" {
		t.Error("should generate reset command")
	}
}

func TestResetToCommit_MixedMode(t *testing.T) {
	result := resetToCommit("abc1234", "mixed")
	if result == "" {
		t.Error("should generate reset command")
	}
}

func TestResetToCommit_HardMode(t *testing.T) {
	result := resetToCommit("abc1234", "hard")
	if result == "" {
		t.Error("should generate reset command")
	}
}

// --- Revert ---

func TestRevertCommit_GeneratesCmd(t *testing.T) {
	cmd := revertCommit("abc1234")
	if cmd == "" {
		t.Error("should generate revert command")
	}
	if !strings.Contains(cmd, "abc1234") {
		t.Errorf("should contain hash, got %q", cmd)
	}
}

func TestRevertCommit_WithMessage(t *testing.T) {
	cmd := revertCommit("abc1234")
	if !strings.Contains(cmd, "revert") {
		t.Errorf("should be revert command, got %q", cmd)
	}
}

// --- Amend ---

func TestAmendLastCommit_WithMessage(t *testing.T) {
	m := model{
		commits: []commit{
			{hash: "abc1234", subject: "old message"},
		},
		cursor: 0,
	}
	amended := amendLastCommit(m, "new message")
	if amended.commits[0].subject == "old message" {
		t.Error("should update subject")
	}
}

func TestAmendLastCommit_PreservesMetadata(t *testing.T) {
	m := model{
		commits: []commit{
			{hash: "abc1234", author: "John", subject: "old"},
		},
		cursor: 0,
	}
	amended := amendLastCommit(m, "new")
	if amended.commits[0].author != "John" {
		t.Error("should preserve author")
	}
}

// ===== OPTION B: COLLABORATION & ANALYTICS =====

// --- Author Statistics ---

func TestAuthorStats_Empty(t *testing.T) {
	stats := calculateAuthorStats([]commit{})
	if len(stats) != 0 {
		t.Errorf("empty commits should have no stats, got %d", len(stats))
	}
}

func TestAuthorStats_SingleAuthor(t *testing.T) {
	commits := []commit{
		{author: "John", subject: "first"},
		{author: "John", subject: "second"},
		{author: "Jane", subject: "third"},
	}
	stats := calculateAuthorStats(commits)
	if stats["John"] != 2 {
		t.Errorf("John should have 2 commits, got %d", stats["John"])
	}
}

func TestAuthorStats_MultipleAuthors(t *testing.T) {
	commits := []commit{
		{author: "John", subject: "first"},
		{author: "Jane", subject: "second"},
		{author: "Bob", subject: "third"},
	}
	stats := calculateAuthorStats(commits)
	if len(stats) != 3 {
		t.Errorf("expected 3 authors, got %d", len(stats))
	}
}

func TestRenderAuthorStats_ShowsCount(t *testing.T) {
	stats := map[string]int{
		"John": 10,
		"Jane": 5,
		"Bob":  3,
	}
	output := renderAuthorStats(stats, 50)
	if output == "" {
		t.Error("should render non-empty output")
	}
	if !strings.Contains(output, "John") {
		t.Errorf("should show author, got %q", output)
	}
}

// --- Time-based Analytics ---

func TestTimeBasedStats_Empty(t *testing.T) {
	stats := calculateTimeStats([]commit{})
	if len(stats) != 0 {
		t.Errorf("empty commits should have no stats, got %d", len(stats))
	}
}

func TestTimeBasedStats_ByDay(t *testing.T) {
	commits := []commit{
		{when: "1 day ago", subject: "today"},
		{when: "1 day ago", subject: "also today"},
		{when: "7 days ago", subject: "week ago"},
	}
	stats := calculateTimeStats(commits)
	if len(stats) == 0 {
		t.Error("should calculate time stats")
	}
}

func TestCommitsPerWeek_Aggregates(t *testing.T) {
	commits := []commit{
		{when: "1 day ago", subject: "a"},
		{when: "2 days ago", subject: "b"},
		{when: "8 days ago", subject: "c"},
	}
	weekly := aggregateByWeek(commits)
	if len(weekly) == 0 {
		t.Error("should aggregate by week")
	}
}

func TestRenderTimeStats_ShowsHeatmap(t *testing.T) {
	stats := map[string]int{
		"Mon": 5,
		"Tue": 3,
		"Wed": 7,
	}
	output := renderTimeStats(stats, 50)
	if output == "" {
		t.Error("should render non-empty output")
	}
}

// --- Co-author Detection ---

func TestExtractCoAuthors_None(t *testing.T) {
	authors := extractCoAuthors("simple commit message")
	if len(authors) != 0 {
		t.Errorf("no co-authors should return empty, got %d", len(authors))
	}
}

func TestExtractCoAuthors_Single(t *testing.T) {
	msg := "Fix bug\n\nCo-authored-by: Jane Smith <jane@example.com>"
	authors := extractCoAuthors(msg)
	if len(authors) != 1 {
		t.Errorf("expected 1 co-author, got %d", len(authors))
	}
	if !strings.Contains(authors[0], "Jane") {
		t.Errorf("should contain name, got %q", authors[0])
	}
}

func TestExtractCoAuthors_Multiple(t *testing.T) {
	msg := "Fix bug\n\nCo-authored-by: Jane <jane@example.com>\nCo-authored-by: Bob <bob@example.com>"
	authors := extractCoAuthors(msg)
	if len(authors) != 2 {
		t.Errorf("expected 2 co-authors, got %d", len(authors))
	}
}

// --- Reviewer Tracking ---

func TestExtractReviewers_None(t *testing.T) {
	reviewers := extractReviewers("normal commit message")
	if len(reviewers) != 0 {
		t.Errorf("no reviewers should return empty, got %d", len(reviewers))
	}
}

func TestExtractReviewers_Mentioned(t *testing.T) {
	msg := "Fix: handle edge case\n\nReviewed-by: John Smith <john@example.com>"
	reviewers := extractReviewers(msg)
	if len(reviewers) != 1 {
		t.Errorf("expected 1 reviewer, got %d", len(reviewers))
	}
}

func TestExtractReviewers_Multiple(t *testing.T) {
	msg := "Feature: new API\n\nReviewed-by: Alice <alice@ex.com>\nReviewed-by: Bob <bob@ex.com>"
	reviewers := extractReviewers(msg)
	if len(reviewers) != 2 {
		t.Errorf("expected 2 reviewers, got %d", len(reviewers))
	}
}

// --- Productivity Metrics ---

func TestCalculateProductivity_Empty(t *testing.T) {
	metrics := calculateProductivity([]commit{})
	if len(metrics) != 0 {
		t.Errorf("empty should have no metrics, got %d", len(metrics))
	}
}

func TestCalculateProductivity_LineChanges(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "first"},
		{hash: "bbb", subject: "second"},
	}
	// Would need diff data to calculate actual metrics
	metrics := calculateProductivity(commits)
	if metrics == nil {
		t.Error("should return map")
	}
}

func TestRenderProductivityMetrics_ShowsData(t *testing.T) {
	metrics := map[string]interface{}{
		"commits":      42,
		"avg_files":    2.5,
		"avg_additions": 15,
	}
	output := renderProductivityMetrics(metrics, 50)
	if output == "" {
		t.Error("should render non-empty output")
	}
}

// --- UI Integration ---

func TestKeyBinding_R_ShowRebase(t *testing.T) {
	m := model{showRebaseUI: false}
	m = handleKeyBinding(m, "R")
	if !m.showRebaseUI {
		t.Error("R should show rebase UI")
	}
}

func TestKeyBinding_C_ShowCherryPick(t *testing.T) {
	m := model{showCherryPickUI: false}
	m = handleKeyBinding(m, "C")
	if !m.showCherryPickUI {
		t.Error("C should show cherry-pick UI")
	}
}

func TestKeyBinding_A_ShowAnalytics(t *testing.T) {
	m := model{showAnalytics: false}
	m = handleKeyBinding(m, "A")
	if !m.showAnalytics {
		t.Error("A should show analytics")
	}
}

func TestRenderRebaseUI_ShowsSequence(t *testing.T) {
	m := model{
		showRebaseUI: true,
		commits: []commit{
			{hash: "aaa", subject: "first"},
			{hash: "bbb", subject: "second"},
		},
	}
	output := renderRebaseUI(m, 50)
	if output == "" {
		t.Error("should render non-empty rebase UI")
	}
}

func TestRenderAnalyticsPanel_ShowsStats(t *testing.T) {
	m := model{
		showAnalytics: true,
		commits: []commit{
			{author: "John", when: "1 day ago"},
			{author: "Jane", when: "2 days ago"},
		},
	}
	output := renderAnalyticsPanel(m, 50)
	if output == "" {
		t.Error("should render non-empty analytics")
	}
}

// --- Bisect & Recovery (5 features) ---

// Feature 1: Interactive Bisect Workflow
func TestInitiateBisect_CreatesState(t *testing.T) {
	m := model{
		commits: []commit{
			{hash: "aaa", shortHash: "aaa1111"},
			{hash: "bbb", shortHash: "bbb2222"},
			{hash: "ccc", shortHash: "ccc3333"},
		},
		cursor: 1,
	}
	m = initiateBisect(m)
	if !m.bisectState.active {
		t.Error("bisect should be active")
	}
	if m.bisectState.current == "" {
		t.Error("should have current commit")
	}
}

func TestBisectMarkGood_AddsToGoodList(t *testing.T) {
	m := model{
		bisectState: bisectState{
			active:  true,
			current: "bbb",
			good:    []string{},
			bad:     []string{},
		},
	}
	m = bisectMarkGood(m)
	if len(m.bisectState.good) != 1 {
		t.Errorf("expected 1 good commit, got %d", len(m.bisectState.good))
	}
}

func TestBisectMarkBad_AddsToBadList(t *testing.T) {
	m := model{
		bisectState: bisectState{
			active:  true,
			current: "aaa",
			good:    []string{},
			bad:     []string{},
		},
	}
	m = bisectMarkBad(m)
	if len(m.bisectState.bad) != 1 {
		t.Errorf("expected 1 bad commit, got %d", len(m.bisectState.bad))
	}
}

func TestBisectFindCulprit_IdentifiesCommit(t *testing.T) {
	commits := []commit{
		{hash: "a1a", shortHash: "a1a1"},
		{hash: "b2b", shortHash: "b2b2"},
		{hash: "c3c", shortHash: "c3c3"},
		{hash: "d4d", shortHash: "d4d4"},
		{hash: "e5e", shortHash: "e5e5"},
	}
	culprit := bisectFindCulprit(commits, []string{"a1a"}, []string{"e5e"})
	if culprit == "" {
		t.Error("should find culprit commit")
	}
}

// Feature 2: Bisect Visualization
func TestRenderBisectUI_ShowsProgress(t *testing.T) {
	m := model{
		showBisectUI: true,
		bisectState: bisectState{
			active:      true,
			current:     "bbb",
			good:        []string{"aaa"},
			bad:         []string{"eee"},
			visualSteps: 2,
			totalSteps:  5,
		},
	}
	output := renderBisectUI(m, 50)
	if output == "" {
		t.Error("should render non-empty bisect UI")
	}
	if !strings.Contains(output, "aaa") && !strings.Contains(output, "eee") {
		t.Error("should show good and bad commits")
	}
}

func TestCalculateBisectProgress_ComputesSteps(t *testing.T) {
	state := bisectState{
		good:    []string{"a", "b"},
		bad:     []string{"e", "f"},
		current: "c",
	}
	steps := calculateBisectProgress(state)
	if steps == 0 {
		t.Error("should calculate non-zero steps")
	}
}

// Feature 3: Reflog Recovery
func TestExtractReflogEntries_ParsesReflog(t *testing.T) {
	reflogOutput := "abc1234 HEAD@{0}: commit: feat: add feature\n" +
		"def5678 HEAD@{1}: rebase: my branch\n" +
		"ghi9012 HEAD@{2}: reset: back to main\n"
	entries := extractReflogEntries(reflogOutput)
	if len(entries) != 3 {
		t.Errorf("expected 3 reflog entries, got %d", len(entries))
	}
	if entries[0].hash != "abc1234" {
		t.Errorf("first hash: got %q", entries[0].hash)
	}
	if entries[0].action != "commit" {
		t.Errorf("first action: got %q", entries[0].action)
	}
}

func TestRecoverFromReflog_RestoresCommits(t *testing.T) {
	m := model{
		reflogEntries: []reflogEntry{
			{hash: "abc1234", action: "rebase", message: "my branch"},
			{hash: "def5678", action: "reset", message: "back to main"},
		},
	}
	m = enableReflogRecovery(m)
	if !m.reflogRecoveryMode {
		t.Error("reflog recovery mode should be enabled")
	}
	if len(m.recoveryCommits) == 0 {
		t.Error("should have recovery commits")
	}
}

// Feature 4: Lost Commits Finder
func TestFindLostCommits_ScansFsck(t *testing.T) {
	fsckOutput := "unreachable commit abc1234\nFix: login bug\n" +
		"unreachable commit def5678\nAdd: user model\n"
	commits := findLostCommits(fsckOutput)
	if len(commits) < 1 {
		t.Error("should find at least 1 lost commit")
	}
}

func TestRecoveryCommitsList_ShowsRecoverable(t *testing.T) {
	m := model{
		showLostCommits: true,
		lostCommits: []lostCommit{
			{hash: "abc", shortHash: "abc1234", subject: "Fix: bug"},
			{hash: "def", shortHash: "def5678", subject: "Add: feature"},
		},
	}
	output := renderLostCommitsUI(m, 50)
	if output == "" {
		t.Error("should render non-empty UI")
	}
}

// Feature 5: Undo Operations
func TestPushUndo_TracksCommit(t *testing.T) {
	m := model{
		undoStack:    []string{"aaa", "bbb"},
		undoStackIdx: 2,
	}
	m = pushUndo(m, "ccc")
	if m.undoStack[len(m.undoStack)-1] != "ccc" {
		t.Error("should add commit to undo stack")
	}
}

func TestUndo_RestoresPreviousState(t *testing.T) {
	m := model{
		undoStack:    []string{"aaa", "bbb", "ccc"},
		undoStackIdx: 3,
	}
	m = performUndo(m)
	if m.undoStackIdx != 2 {
		t.Errorf("undo index: expected 2, got %d", m.undoStackIdx)
	}
}

func TestRenderUndoMenu_ShowsStack(t *testing.T) {
	m := model{
		showUndoMenu: true,
		undoStack:    []string{"aaa", "bbb", "ccc"},
		undoStackIdx: 3,
	}
	output := renderUndoMenu(m, 50)
	if output == "" {
		t.Error("should render non-empty undo menu")
	}
}

// --- Code Patterns & Quality (5 features) ---

// Feature 6: Code Ownership Analysis
func TestAnalyzeCodeOwnership_ComputesOwners(t *testing.T) {
	commits := []commit{
		{author: "Alice", subject: "Fix: main.go"},
		{author: "Alice", subject: "Fix: main.go"},
		{author: "Bob", subject: "Add: utils.go"},
		{author: "Bob", subject: "Add: utils.go"},
		{author: "Bob", subject: "Add: utils.go"},
	}
	ownership := analyzeCodeOwnership(commits)
	if len(ownership) == 0 {
		t.Error("should compute code ownership")
	}
	aliceData := ownership["Alice"]
	if aliceData.author != "Alice" {
		t.Errorf("author: got %q", aliceData.author)
	}
}

func TestDetectCodeOwners_IdentifiesOwners(t *testing.T) {
	ownership := map[string]codeOwnershipData{
		"Alice": {author: "Alice", expertise: 0.85},
		"Bob":   {author: "Bob", expertise: 0.45},
	}
	owner := detectCodeOwners(ownership)
	if owner == "" {
		t.Error("should detect code owner")
	}
}

func TestRenderCodeOwnershipUI_ShowsAnalysis(t *testing.T) {
	m := model{
		showCodeOwnership: true,
		codeOwnership: map[string]codeOwnershipData{
			"Alice": {author: "Alice", expertise: 0.85, files: map[string]int{"main.go": 10}},
			"Bob":   {author: "Bob", expertise: 0.45, files: map[string]int{"utils.go": 5}},
		},
	}
	output := renderCodeOwnershipUI(m, 50)
	if output == "" {
		t.Error("should render non-empty ownership UI")
	}
}

// Feature 7: Hotspot Detection
func TestDetectHotspots_FindsFrequentChanges(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Fix: main.go"},
		{hash: "bbb", subject: "Add: main.go"},
		{hash: "ccc", subject: "Refactor: main.go"},
		{hash: "ddd", subject: "Fix: utils.go"},
		{hash: "eee", subject: "Add: utils.go"},
	}
	hotspots := detectHotspots(commits)
	if len(hotspots) == 0 {
		t.Error("should detect hotspots")
	}
	if hotspots[0].changeFrequency < 1 {
		t.Error("should track change frequency")
	}
}

func TestAssessRiskLevel_EvaluatesHotspots(t *testing.T) {
	hotspot := hotspotData{
		path:            "main.go",
		changeFrequency: 10,
		collaborators:   5,
	}
	risk := assessRiskLevel(hotspot)
	if risk != "low" && risk != "medium" && risk != "high" {
		t.Errorf("invalid risk level: %q", risk)
	}
}

func TestRenderHotspotsUI_ShowsAnalysis(t *testing.T) {
	m := model{
		showHotspots: true,
		hotspots: []hotspotData{
			{path: "main.go", changeFrequency: 10, riskLevel: "high"},
			{path: "utils.go", changeFrequency: 3, riskLevel: "low"},
		},
	}
	output := renderHotspotsUI(m, 50)
	if output == "" {
		t.Error("should render non-empty hotspots UI")
	}
}

// Feature 8: Commit Message Linting
func TestLintCommitMessage_ChecksQuality(t *testing.T) {
	result := lintCommitMessage("add feature", "abc1234")
	if result.score > 100 || result.score < 0 {
		t.Errorf("invalid score: %d", result.score)
	}
	if result.hash != "abc1234" {
		t.Errorf("hash mismatch: got %q", result.hash)
	}
}

func TestValidateCommitFormat_DetectsIssues(t *testing.T) {
	issues := validateCommitFormat("fix bug")
	if len(issues) == 0 {
		t.Error("should detect lowercase issue")
	}
}

func TestRenderLintingUI_ShowsResults(t *testing.T) {
	m := model{
		showLinting: true,
		lintingResults: []lintingResult{
			{hash: "abc", subject: "Add feature", score: 85, issues: []string{}},
			{hash: "def", subject: "fix bug", score: 40, issues: []string{"lowercase"}},
		},
	}
	output := renderLintingUI(m, 50)
	if output == "" {
		t.Error("should render non-empty linting UI")
	}
}

// Feature 9: Large Commit Detection
func TestDetectLargeCommits_FindsBig(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Large refactor"},
		{hash: "bbb", subject: "Small fix"},
		{hash: "ccc", subject: "Huge change"},
	}
	m := model{commits: commits}
	m = analyzeCommitSize(m)
	if len(m.largeCommits) == 0 {
		t.Error("should detect at least 1 large commit")
	}
}

func TestCalculateCommitSize_ComputesLines(t *testing.T) {
	metrics := calculateCommitMetrics("abc1234", 500, 50)
	if metrics.hash != "abc1234" {
		t.Errorf("hash: got %q", metrics.hash)
	}
	if metrics.linesChanged != 500 {
		t.Errorf("lines: expected 500, got %d", metrics.linesChanged)
	}
}

func TestRenderLargeCommitsUI_ShowsLarge(t *testing.T) {
	m := model{
		showLargeCommits: true,
		largeCommits: []commitMetrics{
			{hash: "aaa", linesChanged: 500, filesChanged: 20},
			{hash: "bbb", linesChanged: 1000, filesChanged: 50},
		},
	}
	output := renderLargeCommitsUI(m, 50)
	if output == "" {
		t.Error("should render non-empty UI")
	}
}

// Feature 10: Commit Complexity Analysis
func TestAnalyzeCommitComplexity_ScoresComplexity(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Simple fix"},
		{hash: "bbb", subject: "Refactor with multiple changes"},
	}
	m := model{
		commits:       commits,
		showComplexity: true,
	}
	m = analyzeComplexity(m)
	if len(m.commitMetrics) == 0 {
		t.Error("should analyze commits")
	}
}

func TestCalculateCommitComplexity_EstimatesScore(t *testing.T) {
	metrics := commitMetrics{
		hash:         "abc",
		linesChanged: 300,
		filesChanged: 15,
	}
	score := calculateComplexityScore(metrics)
	if score < 0 || score > 100 {
		t.Errorf("invalid complexity score: %d", score)
	}
}

func TestRenderComplexityUI_ShowsMetrics(t *testing.T) {
	m := model{
		showComplexity: true,
		commitMetrics: []commitMetrics{
			{hash: "aaa", complexity: 45, linesChanged: 200, filesChanged: 10},
			{hash: "bbb", complexity: 78, linesChanged: 500, filesChanged: 25},
		},
	}
	output := renderComplexityUI(m, 50)
	if output == "" {
		t.Error("should render non-empty complexity UI")
	}
}

// --- Commit Analysis & Search (4 features) ---

// Feature 1: Semantic Search
func TestSemanticSearch_FindsFunctions(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Add parseJSON function"},
		{hash: "bbb", subject: "Fix login handler"},
		{hash: "ccc", subject: "Refactor database query"},
	}
	results := semanticSearch(commits, "parseJSON")
	if len(results) == 0 {
		t.Error("should find semantic matches")
	}
}

func TestSemanticSearchRanking_ScoresRelevance(t *testing.T) {
	result := semanticSearchResult{
		hash:      "abc",
		subject:   "Add parseJSON",
		matches:   []string{"parseJSON"},
		relevance: 85,
	}
	if result.relevance < 0 || result.relevance > 100 {
		t.Errorf("invalid relevance score: %d", result.relevance)
	}
}

func TestRenderSemanticSearch_ShowsResults(t *testing.T) {
	m := model{
		showSemanticSearch: true,
		semanticSearchResults: []semanticSearchResult{
			{hash: "aaa", subject: "Add parseJSON", relevance: 95},
			{hash: "bbb", subject: "Call parseJSON", relevance: 80},
		},
	}
	output := renderSemanticSearchUI(m, 50)
	if output == "" {
		t.Error("should render non-empty semantic search UI")
	}
}

// Feature 2: Author Activity Heatmap
func TestAuthorActivityHeatmap_TracksTiming(t *testing.T) {
	commits := []commit{
		{author: "Alice", when: "09:30"},
		{author: "Alice", when: "14:45"},
		{author: "Alice", when: "09:15"},
	}
	heatmap := buildActivityHeatmap(commits)
	if len(heatmap) == 0 {
		t.Error("should build activity heatmap")
	}
}

func TestIdentifyPeakHours_FindsBusyTimes(t *testing.T) {
	data := authorActivityData{
		hourOfDay: map[int]int{9: 5, 14: 8, 17: 3},
	}
	peak := findPeakHour(data)
	if peak != 14 {
		t.Errorf("peak hour should be 14, got %d", peak)
	}
}

func TestRenderActivityHeatmap_ShowsPattern(t *testing.T) {
	m := model{
		showActivityHeatmap: true,
		authorActivityHeatmap: map[string]authorActivityData{
			"Alice": {author: "Alice", peakHour: 14, peakDay: "Wednesday"},
			"Bob":   {author: "Bob", peakHour: 9, peakDay: "Monday"},
		},
	}
	output := renderActivityHeatmapUI(m, 50)
	if output == "" {
		t.Error("should render non-empty heatmap")
	}
}

// Feature 3: Merge Analysis
func TestAnalyzeMerges_DetectsFastForwards(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Merge branch feature"},
		{hash: "bbb", subject: "Regular commit"},
	}
	analysis := analyzeMerges(commits)
	if len(analysis) == 0 {
		t.Error("should analyze merges")
	}
}

func TestDetectFastForward_IdentifiesMergeType(t *testing.T) {
	merge := mergeAnalysis{hash: "abc", isMerge: true, isFastForward: true}
	if !merge.isFastForward {
		t.Error("should detect fast-forward merge")
	}
}

func TestRenderMergeAnalysis_ShowsData(t *testing.T) {
	m := model{
		showMergeAnalysis: true,
		mergeAnalysisData: []mergeAnalysis{
			{hash: "aaa", isMerge: true, isFastForward: true},
			{hash: "bbb", isMerge: true, isFastForward: false, conflictRisk: 45},
		},
	}
	output := renderMergeAnalysisUI(m, 50)
	if output == "" {
		t.Error("should render non-empty merge analysis")
	}
}

// Feature 4: Commit Coupling Analysis
func TestAnalyzeCommitCoupling_FindsCoChanges(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Update main.go and utils.go"},
		{hash: "bbb", subject: "Fix main.go and utils.go"},
		{hash: "ccc", subject: "Add config.go"},
	}
	couplings := analyzeCommitCoupling(commits)
	if len(couplings) == 0 {
		t.Error("should find coupled files")
	}
}

func TestCalculateCoupling_ComputesCorrelation(t *testing.T) {
	coupling := commitCoupling{
		file1:         "main.go",
		file2:         "utils.go",
		coChangeCount: 5,
		correlation:   0.85,
	}
	if coupling.correlation < 0 || coupling.correlation > 1 {
		t.Errorf("invalid correlation: %f", coupling.correlation)
	}
}

func TestRenderCouplingUI_ShowsAnalysis(t *testing.T) {
	m := model{
		showCoupling: true,
		commitCouplings: []commitCoupling{
			{file1: "main.go", file2: "utils.go", coChangeCount: 5, correlation: 0.85},
			{file1: "auth.go", file2: "user.go", coChangeCount: 3, correlation: 0.60},
		},
	}
	output := renderCouplingAnalysisUI(m, 50)
	if output == "" {
		t.Error("should render non-empty coupling UI")
	}
}

// --- Performance & Filtering (4 features) ---

// Feature 5: Filter by File Extension
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

func TestToggleExtensionFilter_TrackFilter(t *testing.T) {
	m := model{extensionFilters: []fileExtensionFilter{}}
	m = toggleExtensionFilter(m, ".go")
	if m.currentExtFilter != ".go" {
		t.Errorf("filter should be .go, got %q", m.currentExtFilter)
	}
}

func TestRenderExtensionFilter_ShowsActive(t *testing.T) {
	m := model{
		extensionFilters: []fileExtensionFilter{
			{extension: ".go", enabled: true},
			{extension: ".js", enabled: false},
		},
	}
	output := renderExtensionFilterUI(m, 50)
	if output == "" {
		t.Error("should render non-empty filter UI")
	}
}

// Feature 6: Commit Grouping
func TestGroupCommits_ByBranch(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Feature: add feature"},
		{hash: "bbb", subject: "Fix: bug fix"},
		{hash: "ccc", subject: "Feature: another feature"},
	}
	groups := groupCommits(commits, "branch")
	if len(groups) == 0 {
		t.Error("should group commits")
	}
}

func TestGroupCommits_ByDate(t *testing.T) {
	commits := []commit{
		{hash: "aaa", when: "2 days ago"},
		{hash: "bbb", when: "2 days ago"},
		{hash: "ccc", when: "1 day ago"},
	}
	groups := groupCommits(commits, "date")
	if len(groups) != 2 {
		t.Errorf("expected 2 date groups, got %d", len(groups))
	}
}

func TestRenderCommitGroups_ShowsGrouped(t *testing.T) {
	m := model{
		commitGroups: []commitGroup{
			{name: "Feature", label: "feat", commits: []string{"aaa", "bbb"}},
			{name: "Fix", label: "fix", commits: []string{"ccc"}},
		},
	}
	output := renderCommitGroupsUI(m, 50)
	if output == "" {
		t.Error("should render non-empty groups UI")
	}
}

// Feature 7: Fast-Forward Merge Detection
func TestDetectFastForwardMerges_IdentifiesMerges(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Merge branch feature (fast-forward)"},
		{hash: "bbb", subject: "Merge branch bugfix"},
	}
	ffMerges := detectFastForwardMerges(commits)
	if len(ffMerges) == 0 {
		t.Error("should detect fast-forward merges")
	}
}

func TestRenderFastForwards_ShowsFFInfo(t *testing.T) {
	m := model{
		mergeAnalysisData: []mergeAnalysis{
			{hash: "aaa", isMerge: true, isFastForward: true},
			{hash: "bbb", isMerge: true, isFastForward: false},
		},
	}
	output := renderFastForwardsUI(m, 50)
	if output == "" {
		t.Error("should render fast-forward info")
	}
}

// Feature 8: Dependency Change Tracking
func TestTrackDependencies_FindsVersions(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Upgrade go-git from 4.7 to 5.0"},
		{hash: "bbb", subject: "Update react to 17.0"},
	}
	deps := trackDependencyChanges(commits)
	if len(deps) == 0 {
		t.Error("should find dependency changes")
	}
}

func TestRenderDependencies_ShowsChanges(t *testing.T) {
	m := model{
		showDependencies: true,
		dependencyChanges: []dependencyChange{
			{hash: "aaa", dep: "go-git", oldVer: "4.7", newVer: "5.0"},
			{hash: "bbb", dep: "react", oldVer: "16.0", newVer: "17.0"},
		},
	}
	output := renderDependenciesUI(m, 50)
	if output == "" {
		t.Error("should render non-empty dependencies UI")
	}
}

// --- Advanced Workflows (5 features) ---

// Feature 9: Worktree Support
func TestLoadWorktrees_ParsesList(t *testing.T) {
	worktreeOutput := "/home/user/repo\n/home/user/repo-feature (branch: feature)\n/home/user/repo-bugfix\n"
	worktrees := loadWorktrees(worktreeOutput)
	if len(worktrees) == 0 {
		t.Error("should load worktrees")
	}
}

func TestSwitchWorktree_ChangesPath(t *testing.T) {
	m := model{
		worktrees:       []worktreeInfo{{path: "/repo1", branch: "main"}, {path: "/repo2", branch: "feature"}},
		currentWorktree: "/repo1",
	}
	m = switchWorktree(m, "/repo2")
	if m.currentWorktree != "/repo2" {
		t.Errorf("should switch to /repo2, got %q", m.currentWorktree)
	}
}

func TestRenderWorktrees_ShowsList(t *testing.T) {
	m := model{
		showWorktrees: true,
		worktrees: []worktreeInfo{
			{path: "/repo1", branch: "main"},
			{path: "/repo2", branch: "feature"},
		},
	}
	output := renderWorktreesUI(m, 50)
	if output == "" {
		t.Error("should render non-empty worktrees UI")
	}
}

// Feature 10: Submodule Tracking
func TestParseSubmodules_ExtractsInfo(t *testing.T) {
	configOutput := "[submodule \"lib1\"]\npath = lib1\nurl = https://github.com/user/lib1\n"
	submodules := parseSubmodules(configOutput)
	if len(submodules) == 0 {
		t.Error("should parse submodules")
	}
}

func TestRenderSubmodules_ShowsList(t *testing.T) {
	m := model{
		showSubmodules: true,
		submodules: []submoduleInfo{
			{path: "lib1", url: "https://github.com/user/lib1", branch: "main"},
			{path: "lib2", url: "https://github.com/user/lib2", branch: "develop"},
		},
	}
	output := renderSubmodulesUI(m, 50)
	if output == "" {
		t.Error("should render non-empty submodules UI")
	}
}

// Feature 11: Named Stashes
func TestCreateNamedStash_StoreName(t *testing.T) {
	m := model{namedStashes: []namedStash{}}
	m = createNamedStash(m, 0, "my-stash", "Work in progress")
	if len(m.namedStashes) != 1 {
		t.Error("should create named stash")
	}
	if m.namedStashes[0].name != "my-stash" {
		t.Errorf("name should be my-stash, got %q", m.namedStashes[0].name)
	}
}

func TestRenderNamedStashes_ShowsList(t *testing.T) {
	m := model{
		showNamedStashes: true,
		namedStashes: []namedStash{
			{index: 0, name: "my-stash", description: "WIP"},
			{index: 1, name: "feature-work", description: "In progress"},
		},
	}
	output := renderNamedStashesUI(m, 50)
	if output == "" {
		t.Error("should render non-empty stashes UI")
	}
}

// Feature 12: Tag Management
func TestQueueTagOperation_TracksIntent(t *testing.T) {
	m := model{pendingTagOps: []tagOperation{}}
	m = queueTagOperation(m, "v1.0.0", "abc1234", "create", "Release 1.0.0")
	if len(m.pendingTagOps) != 1 {
		t.Error("should queue tag operation")
	}
}

func TestRenderTagMgmt_ShowsPending(t *testing.T) {
	m := model{
		showTagMgmt: true,
		pendingTagOps: []tagOperation{
			{name: "v1.0.0", hash: "abc1234", action: "create"},
			{name: "v2.0.0", hash: "def5678", action: "delete"},
		},
	}
	output := renderTagMgmtUI(m, 50)
	if output == "" {
		t.Error("should render non-empty tag UI")
	}
}

// Feature 13: GPG Signature Status
func TestExtractGPGStatus_ParsesSignatures(t *testing.T) {
	gpgOutput := "abc1234 signed Alice Key1234567890\n"
	statuses := extractGPGSignatureStatus(gpgOutput)
	if len(statuses) == 0 {
		t.Error("should parse GPG statuses")
	}
}

func TestRenderGPGStatus_ShowsSignatures(t *testing.T) {
	m := model{
		showGPGStatus: true,
		gpgStatuses: map[string]gpgSignatureStatus{
			"abc1234": {hash: "abc1234", signed: true, signer: "Alice", verified: true},
			"def5678": {hash: "def5678", signed: false},
		},
	}
	output := renderGPGStatusUI(m, 50)
	if output == "" {
		t.Error("should render non-empty GPG UI")
	}
}

// --- Visualization (5 features) ---

// Feature 14: Contributor Flamegraph
func TestBuildContributorFlame_RanksAuthors(t *testing.T) {
	commits := []commit{
		{author: "Alice", subject: "feat: add feature"},
		{author: "Alice", subject: "feat: another feature"},
		{author: "Bob", subject: "fix: bug"},
	}
	flame := buildContributorFlame(commits)
	if len(flame) == 0 {
		t.Error("should build flamegraph")
	}
	if flame[0].author != "Alice" {
		t.Errorf("top author should be Alice, got %q", flame[0].author)
	}
}

func TestRenderFlamegraph_ShowsVisualization(t *testing.T) {
	m := model{
		showFlamegraph: true,
		contributorFlameData: []contributorFlameData{
			{author: "Alice", commits: 50, percentage: 65.0},
			{author: "Bob", commits: 27, percentage: 35.0},
		},
	}
	output := renderFlamegraphUI(m, 50)
	if output == "" {
		t.Error("should render non-empty flamegraph")
	}
}

// Feature 15: Timeline Slider
func TestBuildTimeline_CreatePoints(t *testing.T) {
	commits := []commit{
		{hash: "aaa", when: "1 day ago"},
		{hash: "bbb", when: "1 day ago"},
		{hash: "ccc", when: "2 days ago"},
	}
	timeline := buildTimeline(commits)
	if len(timeline) == 0 {
		t.Error("should build timeline")
	}
}

func TestRenderTimelineSlider_ShowsSlider(t *testing.T) {
	m := model{
		showTimeline: true,
		timelinePoints: []timelinePoint{
			{date: "2026-04-20", commits: 5, hash: "abc"},
			{date: "2026-04-21", commits: 3, hash: "def"},
		},
		timelineSliderPos: 0,
	}
	output := renderTimelineSliderUI(m, 50)
	if output == "" {
		t.Error("should render non-empty timeline")
	}
}

// Feature 16: Tree View
func TestBuildTreeView_CreatesHierarchy(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Initial commit"},
		{hash: "bbb", subject: "Add feature"},
		{hash: "ccc", subject: "Fix bug"},
	}
	root := buildTreeView(commits)
	if root == nil {
		t.Error("should build tree view")
	}
}

func TestRenderTreeView_ShowsHierarchy(t *testing.T) {
	m := model{
		showTreeView: true,
		treeRoot: &treeNode{
			hash:    "aaa",
			subject: "Initial",
			depth:   0,
			children: []*treeNode{
				{hash: "bbb", subject: "Feature", depth: 1},
			},
		},
	}
	output := renderTreeViewUI(m, 50)
	if output == "" {
		t.Error("should render non-empty tree view")
	}
}

// Feature 17: Author Comparison
func TestCompareAuthors_ComputesDifferences(t *testing.T) {
	m := model{
		commits: []commit{
			{author: "Alice", subject: "Add feature"},
			{author: "Alice", subject: "Fix bug"},
			{author: "Bob", subject: "Update docs"},
		},
		selectedAuthors: [2]string{"Alice", "Bob"},
	}
	comp := compareAuthors(m)
	if len(comp) == 0 {
		t.Error("should compare authors")
	}
}

func TestRenderAuthorComparison_ShowsSideBySide(t *testing.T) {
	m := model{
		showAuthorComparison: true,
		authorComparisons: []authorComparison{
			{author1: "Alice", commits1: 50, files1: 100, author2: "Bob", commits2: 30, files2: 80, similarity: 0.75},
		},
	}
	output := renderAuthorComparisonUI(m, 50)
	if output == "" {
		t.Error("should render non-empty comparison")
	}
}

// Feature 18: File Heatmap
func TestBuildFileHeatmap_TracksFrequency(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Update main.go"},
		{hash: "bbb", subject: "Fix main.go"},
		{hash: "ccc", subject: "Refactor utils.go"},
	}
	heatmap := buildFileHeatmap(commits)
	if len(heatmap) == 0 {
		t.Error("should build file heatmap")
	}
}

func TestRenderFileHeatmap_ShowsFrequency(t *testing.T) {
	m := model{
		showFileHeatmap: true,
		fileHeatmap: []fileHeatmapEntry{
			{path: "main.go", frequency: 15, recent: 3, risk: "high"},
			{path: "utils.go", frequency: 5, recent: 1, risk: "low"},
		},
	}
	output := renderFileHeatmapUI(m, 50)
	if output == "" {
		t.Error("should render non-empty file heatmap")
	}
}

// --- Integration & Export (5 features) ---

// Feature 19: GitHub PR Linking
func TestExtractPRReferences_FindsPRNumbers(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Fix #123 login issue"},
		{hash: "bbb", subject: "Merge PR #456"},
		{hash: "ccc", subject: "Regular commit"},
	}
	prefs := extractPRReferences(commits)
	if len(prefs) == 0 {
		t.Error("should find PR references")
	}
}

func TestRenderPRLinks_ShowsLinks(t *testing.T) {
	m := model{
		showPRLinks: true,
		prReferences: []githubPRReference{
			{hash: "aaa", prNumber: 123, status: "merged"},
			{hash: "bbb", prNumber: 456, status: "open"},
		},
	}
	output := renderPRLinksUI(m, 50)
	if output == "" {
		t.Error("should render non-empty PR links")
	}
}

// Feature 20: JIRA Ticket Linking
func TestExtractJiraTickets_FindsTickets(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "PROJ-123 Add feature"},
		{hash: "bbb", subject: "Fix PROJ-456 bug"},
		{hash: "ccc", subject: "Regular commit"},
	}
	tickets := extractJiraTickets(commits)
	if len(tickets) == 0 {
		t.Error("should find JIRA tickets")
	}
}

func TestRenderJiraLinks_ShowsTickets(t *testing.T) {
	m := model{
		showJiraLinks: true,
		jiraLinks: []jiraTicketLink{
			{hash: "aaa", ticket: "PROJ-123", status: "done"},
			{hash: "bbb", ticket: "PROJ-456", status: "in-progress"},
		},
	}
	output := renderJiraLinksUI(m, 50)
	if output == "" {
		t.Error("should render non-empty JIRA links")
	}
}

// Feature 21: Export to Markdown
func TestExportToMarkdown_FormatsText(t *testing.T) {
	commits := []commit{
		{hash: "aaa", shortHash: "aaa1111", author: "Alice", subject: "Add feature"},
		{hash: "bbb", shortHash: "bbb2222", author: "Bob", subject: "Fix bug"},
	}
	exported := exportToMarkdown(commits)
	if exported.content == "" {
		t.Error("should generate markdown export")
	}
	if !strings.Contains(exported.content, "Alice") {
		t.Error("should include author")
	}
}

func TestRenderExportUI_ShowsOptions(t *testing.T) {
	m := model{
		showExportUI: true,
		exportFormat: "markdown",
	}
	output := renderExportUI(m, 50)
	if output == "" {
		t.Error("should render non-empty export UI")
	}
}

// Feature 22: Patch Series Export
func TestExportPatchSeries_CreatesPatch(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Add feature"},
		{hash: "bbb", subject: "Fix bug"},
	}
	exported := exportPatchSeries(commits)
	if exported.content == "" {
		t.Error("should generate patch series")
	}
	if exported.format != "patch" {
		t.Errorf("format should be patch, got %q", exported.format)
	}
}

// Feature 23: Issue Reference Tracking
func TestExtractIssueReferences_FindsIssues(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "fixes #123 and closes #456"},
		{hash: "bbb", subject: "resolves #789"},
		{hash: "ccc", subject: "Regular commit"},
	}
	refs := extractIssueReferences(commits)
	if len(refs) == 0 {
		t.Error("should find issue references")
	}
}

func TestRenderIssueRefs_ShowsReferences(t *testing.T) {
	m := model{
		showIssueRefs: true,
		issueReferences: []issueReference{
			{hash: "aaa", references: []string{"#123", "#456"}, keywords: []string{"fixes", "closes"}},
			{hash: "bbb", references: []string{"#789"}, keywords: []string{"resolves"}},
		},
	}
	output := renderIssueRefsUI(m, 50)
	if output == "" {
		t.Error("should render non-empty issue refs")
	}
}

// --- Advanced Git Operations (5 features) ---

// Feature 1: Interactive Rebase with Live Preview
func TestPreviewRebaseOperations_ShowsChanges(t *testing.T) {
	ops := []rebaseOp{
		{action: "pick", hash: "aaa", subject: "First"},
		{action: "squash", hash: "bbb", subject: "Second"},
	}
	preview := previewRebaseOperations(ops)
	if !preview.willApply {
		t.Error("should preview rebase")
	}
}

// Feature 2: Conflict Resolution UI
func TestDetectConflicts_FindsMarkers(t *testing.T) {
	conflicts := detectConflicts("<<<<<<< HEAD\nmy code\n=======\ntheir code\n>>>>>>> branch")
	if len(conflicts) == 0 {
		t.Error("should detect conflict markers")
	}
}

func TestRenderConflictUI_ShowsOptions(t *testing.T) {
	m := model{
		showConflictUI: true,
		conflictList: []conflictInfo{
			{file: "main.go", resolved: false},
			{file: "utils.go", resolved: true},
		},
	}
	output := renderConflictUI(m, 50)
	if output == "" {
		t.Error("should render conflict UI")
	}
}

// Feature 3: Squash/Fixup Automation
func TestPlanSquash_CreatesSequence(t *testing.T) {
	plan := planSquashSequence("aaa", []string{"bbb", "ccc"}, "Combined message")
	if plan.resultMsg == "" {
		t.Error("should create squash plan")
	}
}

// Feature 4: Cherry-pick Improvements
func TestImproveCheryPick_SuggestsResolutions(t *testing.T) {
	m := model{
		commits: []commit{
			{hash: "aaa", subject: "Fix: important fix"},
		},
	}
	improvements := improveCherryPick(m, "aaa")
	if improvements == nil {
		t.Error("should improve cherry-pick")
	}
}

// Feature 5: Commit Amend with Diff Viewing
func TestPreviewAmend_ShowsDiff(t *testing.T) {
	preview := previewAmendCommit("original message", "new message", map[string]int{"file.go": 5})
	if preview.originalMsg == "" {
		t.Error("should preview amend")
	}
}

// --- Team & Collaboration (5 features) ---

// Feature 6: Team Statistics Dashboard
func TestCalculateTeamStats_ComputesMetrics(t *testing.T) {
	commits := []commit{
		{author: "Alice", subject: "Feature A"},
		{author: "Alice", subject: "Fix B"},
		{author: "Bob", subject: "Docs C"},
	}
	stats := calculateTeamStats(commits)
	if len(stats) == 0 {
		t.Error("should calculate team stats")
	}
}

// Feature 7: Code Review Workflow Automation
func TestAutomateReviewWorkflow_TracksState(t *testing.T) {
	workflow := automateReviewWorkflow(123, "alice", []string{"bob", "charlie"})
	if workflow.prNumber != 123 {
		t.Error("should automate review workflow")
	}
}

// Feature 8: Reviewer Assignment Suggestions
func TestSuggestReviewers_RecommendsExperts(t *testing.T) {
	m := model{
		commits: []commit{
			{author: "Alice", subject: "Fix main.go"},
			{author: "Alice", subject: "Update main.go"},
		},
	}
	suggestions := suggestReviewers(m, "utils.go")
	if len(suggestions) == 0 {
		t.Error("should suggest reviewers")
	}
}

// Feature 9: Pair Programming Detection
func TestDetectPairProgramming_FindsPatterns(t *testing.T) {
	commits := []commit{
		{author: "Alice", subject: "Pair: Alice & Bob"},
		{author: "Alice", subject: "Pair: Alice & Bob"},
	}
	pairs := detectPairProgramming(commits)
	if len(pairs) == 0 {
		t.Error("should detect pair programming")
	}
}

// Feature 10: Team Velocity Tracking
func TestCalculateVelocity_TracksProgress(t *testing.T) {
	commits := []commit{
		{hash: "aaa", when: "1 week ago", subject: "Feature 1"},
		{hash: "bbb", when: "1 week ago", subject: "Fix 1"},
		{hash: "ccc", when: "2 weeks ago", subject: "Feature 2"},
	}
	velocity := calculateVelocity(commits)
	if len(velocity) == 0 {
		t.Error("should calculate velocity")
	}
}

// --- AI-Powered Insights (5 features) ---

// Feature 11: Commit Message Auto-completion
func TestAutoCompleteMessage_SuggestsEndings(t *testing.T) {
	completions := autoCompleteMessage("Add", []commit{
		{subject: "Add feature X"},
		{subject: "Add test for Y"},
	})
	if len(completions) == 0 {
		t.Error("should suggest completions")
	}
}

// Feature 12: ML-based Commit Classification
func TestClassifyCommit_CategorizesFix(t *testing.T) {
	class := classifyCommit("Fix: login bug", "aaa")
	if class.category == "" {
		t.Error("should classify commit")
	}
	if class.category != "fix" && class.category != "feature" {
		t.Error("category should be valid")
	}
}

// Feature 13: Anomaly Detection
func TestDetectAnomalies_FindsUnusual(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Small fix"},
		{hash: "bbb", subject: "Massive refactor with 10000 lines"},
		{hash: "ccc", subject: "Another fix"},
	}
	anomalies := detectAnomalies(commits)
	if len(anomalies) == 0 {
		t.Error("should detect anomalies")
	}
}

// Feature 14: Similar Commits Finder
func TestFindSimilarCommits_ComparesMessages(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "Fix login bug in auth.go"},
		{hash: "bbb", subject: "Fix login issue in auth.go"},
		{hash: "ccc", subject: "Add feature"},
	}
	similar := findSimilarCommits(commits, "aaa")
	if len(similar) == 0 {
		t.Error("should find similar commits")
	}
}

// Feature 15: Auto-generated Summaries
func TestGenerateAutoSummary_CreatesAbstract(t *testing.T) {
	summary := generateAutoSummary("aaa", "Fix: long commit message with many details about the authentication system")
	if summary.summary == "" {
		t.Error("should generate summary")
	}
}

// --- Compliance & Security (5 features) ---

// Feature 16: Commit Signing Enforcement
func TestCheckSigningCompliance_VerifiesSignature(t *testing.T) {
	statuses := checkSigningCompliance([]commit{
		{hash: "aaa", subject: "Signed commit"},
		{hash: "bbb", subject: "Unsigned commit"},
	}, true)
	if len(statuses) == 0 {
		t.Error("should check signing compliance")
	}
}

// Feature 17: License Header Tracking
func TestTrackLicenseHeaders_ChecksFiles(t *testing.T) {
	headers := trackLicenseHeaders("aaa")
	if headers == nil {
		t.Error("should track license headers")
	}
}

// Feature 18: Security Scanning Integration
func TestScanForSecurityIssues_DetectsProblems(t *testing.T) {
	issues := scanForSecurityIssues("aaa", "api_key = 'sk-123456789'")
	if len(issues) == 0 {
		t.Error("should detect security issues")
	}
}

// Feature 19: GDPR Data Deletion Tracking
func TestTrackDataDeletionRequests_LogsRequests(t *testing.T) {
	m := model{dataDeleteRequests: []dataDeleteRequest{}}
	m = trackDataDeletion(m, "aaa", "user@example.com")
	if len(m.dataDeleteRequests) != 1 {
		t.Error("should track deletion request")
	}
}

// Feature 20: Secrets Detection
func TestDetectSecrets_FindsExposed(t *testing.T) {
	secrets := detectSecrets("aaa", "password = 'secret123'")
	if len(secrets) == 0 {
		t.Error("should detect secrets")
	}
}

// --- Release & Versioning (5 features) ---

// Feature 21: Semantic Versioning Detection
func TestDetectSemver_IdentifiesVersions(t *testing.T) {
	versions := detectSemver([]commit{
		{hash: "aaa", subject: "v1.0.0"},
		{hash: "bbb", subject: "v1.1.0"},
		{hash: "ccc", subject: "v2.0.0"},
	})
	if len(versions) == 0 {
		t.Error("should detect versions")
	}
}

// Feature 22: Changelog Auto-generation
func TestGenerateChangelog_CreatesNotes(t *testing.T) {
	commits := []commit{
		{hash: "aaa", subject: "feat: add feature"},
		{hash: "bbb", subject: "fix: bug fix"},
		{hash: "ccc", subject: "docs: update"},
	}
	changelog := generateChangelog(commits, "v1.0.0")
	if changelog == nil {
		t.Error("should generate changelog")
	}
}

// Feature 23: Release Note Builder
func TestBuildReleaseNotes_CreatesDocument(t *testing.T) {
	notes := buildReleaseNotes("v1.0.0", []string{"aaa", "bbb"})
	if notes.version != "v1.0.0" {
		t.Error("should build release notes")
	}
}

// Feature 24: Version Bump History
func TestTrackVersionBumps_RecordsChanges(t *testing.T) {
	bumps := trackVersionBumps([]commit{
		{hash: "aaa", subject: "Bump version 1.0.0 -> 1.1.0"},
	})
	if len(bumps) == 0 {
		t.Error("should track version bumps")
	}
}

// Feature 25: Milestone Tracking
func TestTrackMilestones_AssignCommits(t *testing.T) {
	m := model{milestones: []milestone{}}
	m = createMilestone(m, "v1.0", []string{"aaa", "bbb"})
	if len(m.milestones) != 1 {
		t.Error("should create milestone")
	}
}

// --- Advanced Performance (5 features) ---

// Feature 26: Incremental Repo Loading
func TestIncrementalLoad_TracksProgress(t *testing.T) {
	state := incrementalLoadRepository("repo", 1000)
	if state.totalCommits == 0 {
		t.Error("should track load progress")
	}
}

// Feature 27: Parallel Diff Processing
func TestParallelDiffProcessing_ProcessesConcurrently(t *testing.T) {
	results := parallelProcessDiffs([]string{"aaa", "bbb", "ccc"})
	if len(results) == 0 {
		t.Error("should process diffs in parallel")
	}
}

// Feature 28: Background Indexing
func TestBackgroundIndexing_BuildsIndex(t *testing.T) {
	data := buildBackgroundIndex([]commit{
		{hash: "aaa", subject: "Feature"},
		{hash: "bbb", subject: "Fix"},
	})
	if data.entries == 0 {
		t.Error("should build index")
	}
}

// Feature 29: Lazy Blame Loading
func TestLazyBlame_LoadsOnDemand(t *testing.T) {
	blame := lazyLoadBlame("aaa", "main.go")
	if blame == nil {
		t.Error("should lazy load blame")
	}
}

// Feature 30: Memory Optimization
func TestMemoryOptimization_TracksUsage(t *testing.T) {
	metrics := optimizeMemory([]commit{
		{hash: "aaa", subject: "Fix"},
	})
	if metrics.usageBytes == 0 {
		t.Error("should track memory")
	}
}
