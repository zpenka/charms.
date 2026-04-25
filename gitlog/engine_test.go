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
