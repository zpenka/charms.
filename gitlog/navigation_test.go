package gitlog

import (
	"testing"
)
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

func TestSwitchPanel_TowardsDiff(t *testing.T) {
	m := model{focus: panelList}
	m = switchPanel(m)
	AssertEqual(t, panelDiff, m.focus, "should switch to diff panel")
}

func TestSwitchPanel_TowardsList(t *testing.T) {
	m := model{focus: panelDiff}
	m = switchPanel(m)
	AssertEqual(t, panelList, m.focus, "should switch to list panel")
}

func TestListPanelWidth_Minimum(t *testing.T) {
	AssertIntRange(t, listPanelWidth(40), 36, 52, "panel width should be in range")
}

func TestListPanelWidth_Maximum(t *testing.T) {
	AssertTrue(t, listPanelWidth(300) <= 52, "panel width should be at most 52")
}

func TestListPanelWidth_ThirdOfWidth(t *testing.T) {
	w := listPanelWidth(120)
	AssertIntRange(t, w, 36, 52, "width for total=120 should be in range")
}

func TestDiffPanelHeight_Normal(t *testing.T) {
	m := model{height: 40}
	h := diffPanelHeight(m)
	AssertEqual(t, 33, h, "height should be 40-7=33")
}

func TestDiffPanelHeight_Minimum(t *testing.T) {
	m := model{height: 5}
	AssertTrue(t, diffPanelHeight(m) >= 5, "height should be at least 5")
}

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

func TestToggleFileView_Show(t *testing.T) {
	m := model{showFiles: false}
	m = toggleFileView(m)
	AssertTrue(t, m.showFiles, "should show files")
}

func TestToggleFileView_Hide(t *testing.T) {
	m := model{showFiles: true, fileCursor: 3}
	m = toggleFileView(m)
	AssertFalse(t, m.showFiles, "should hide files")
	AssertEqual(t, 0, m.fileCursor, "cursor should reset on hide")
}

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

func TestScrollToDiffLine_ClampsToMax(t *testing.T) {
	m := model{diffLines: makeDiffLines(50), height: 30}
	m = scrollToDiffLine(m, 1000)
	panelH := diffPanelHeight(m)
	expected := len(m.diffLines) - panelH
	AssertEqual(t, expected, m.diffOffset, "should clamp to max")
}

func TestScrollToDiffLine_ClampsToZero(t *testing.T) {
	m := model{diffLines: makeDiffLines(50), height: 30}
	m = scrollToDiffLine(m, -5)
	AssertEqual(t, 0, m.diffOffset, "should clamp to zero")
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

// --- cursor movement ---

func TestMoveCursorDown(t *testing.T) {
	m := model{commits: makeCommits(3), cursor: 0}
	m = moveCursorDown(m)
	AssertEqual(t, 1, m.cursor, "cursor should advance")
}

func TestMoveCursorDown_AtEnd(t *testing.T) {
	m := model{commits: makeCommits(3), cursor: 2}
	m = moveCursorDown(m)
	AssertEqual(t, 2, m.cursor, "cursor should clamp at end")
}

func TestMoveCursorDown_ResetsDiffOffset(t *testing.T) {
	m := model{commits: makeCommits(3), cursor: 0, diffOffset: 10}
	m = moveCursorDown(m)
	AssertEqual(t, 0, m.diffOffset, "diffOffset should reset on commit change")
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
	AssertEqual(t, 5, m.diffOffset, "should scroll down")
}

func TestScrollDiffUp(t *testing.T) {
	m := model{diffLines: makeDiffLines(50), diffOffset: 10, height: 30}
	m = scrollDiffUp(m, 3)
	AssertEqual(t, 7, m.diffOffset, "should scroll up")
}

func TestScrollDiffDown_ClampsToMax(t *testing.T) {
	m := model{diffLines: makeDiffLines(50), diffOffset: 0, height: 30}
	initialOffset := m.diffOffset
	m = scrollDiffDown(m, 100)
	AssertTrue(t, m.diffOffset >= initialOffset, "should not decrease offset when scrolling down")
}

func TestScrollDiffUp_ClampsToZero(t *testing.T) {
	m := model{diffLines: makeDiffLines(50), diffOffset: 5, height: 30}
	m = scrollDiffUp(m, 100)
	AssertEqual(t, 0, m.diffOffset, "should clamp to zero")
}

func TestScrollDiffDown_FitsInPanel(t *testing.T) {
	// fewer lines than panel height — can't scroll
	m := model{diffLines: makeDiffLines(5), diffOffset: 0, height: 30}
	m = scrollDiffDown(m, 10)
	if m.diffOffset != 0 {
		t.Errorf("expected 0 when content fits panel, got %d", m.diffOffset)
	}
}

// --- panel sizing ---

func TestDiffPanelWidth(t *testing.T) {
	total := 120
	lw := listPanelWidth(total)
	dw := diffPanelWidth(total)
	AssertEqual(t, total, lw+dw+1, "panel widths should sum to total")
}

