package gitlog

import (
	"strings"
	"testing"
)
func TestCommitStats_Empty(t *testing.T) {
	stats := commitStats([]diffLine{})
	AssertEqual(t, 0, stats.filesChanged, "empty diff should have 0 files")
	AssertEqual(t, 0, stats.insertions, "empty diff should have 0 insertions")
	AssertEqual(t, 0, stats.deletions, "empty diff should have 0 deletions")
}

func TestCommitStats_CountsFiles(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "diff --git a/foo.go b/foo.go"},
		{lineAdded, "+line"},
		{lineMeta, "diff --git a/bar.go b/bar.go"},
		{lineRemoved, "-line"},
	}
	stats := commitStats(lines)
	AssertEqual(t, 2, stats.filesChanged, "should count 2 files")
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
	AssertEqual(t, 2, stats.insertions, "should count 2 insertions")
	AssertEqual(t, 3, stats.deletions, "should count 3 deletions")
}

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

func TestHandleGoToCommitInput_InvalidHash(t *testing.T) {
	commits := []commit{{shortHash: "abc1234", hash: "abc1234567890"}}
	m := model{commits: commits, cursor: 0}
	m = handleGoToCommitInput(m, "zzz9999")
	if m.cursor != 0 {
		t.Errorf("should stay at 0 for invalid hash, got %d", m.cursor)
	}
}

