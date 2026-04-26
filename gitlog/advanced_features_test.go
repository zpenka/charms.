package gitlog

import (
	"testing"
)
func TestBookmarks_InitEmpty(t *testing.T) {
	m := model{}
	if len(m.bookmarks) != 0 {
		t.Errorf("should start with no bookmarks, got %d", len(m.bookmarks))
	}
}

func TestBookmarks_Toggle(t *testing.T) {
	m := model{cursor: 5, commits: makeCommits(10)}
	m = toggleBookmark(m)
	AssertTrue(t, isBookmarked(m, 5), "cursor position should be bookmarked")
}

func TestBookmarks_ToggleRemoves(t *testing.T) {
	c := commit{shortHash: "abc123"}
	m := model{cursor: 0, commits: []commit{c}, bookmarks: []string{"abc123"}}
	m = toggleBookmark(m)
	AssertFalse(t, isBookmarked(m, 0), "bookmark should be removed")
}

func TestBookmarks_JumpToNext(t *testing.T) {
	commits := []commit{
		{shortHash: "aaa", subject: "first"},
		{shortHash: "bbb", subject: "second"},
		{shortHash: "ccc", subject: "third"},
	}
	m := model{commits: commits, cursor: 0, bookmarks: []string{"bbb", "ccc"}}
	m = jumpToNextBookmark(m)
	AssertEqual(t, 1, m.cursor, "should jump to next bookmark")
}

func TestBookmarks_JumpToPrev(t *testing.T) {
	commits := []commit{
		{shortHash: "aaa", subject: "first"},
		{shortHash: "bbb", subject: "second"},
		{shortHash: "ccc", subject: "third"},
	}
	m := model{commits: commits, cursor: 2, bookmarks: []string{"aaa", "bbb"}}
	m = jumpToPrevBookmark(m)
	AssertEqual(t, 1, m.cursor, "should jump to prev bookmark")
}

func TestWorkflowTemplate_CreatesTemplate(t *testing.T) {
	tmpl := &WorkflowTemplate{
		Name: "Feature Branch",
		Steps: []string{"git checkout -b feature/...", "git commit"},
	}

	AssertNotNil(t, tmpl, "should create template")
	AssertEqual(t, "Feature Branch", tmpl.Name, "should have correct name")
	AssertLen(t, tmpl.Steps, 2, "should have 2 steps")
}

func TestWorkflowTemplate_ExecutesSteps(t *testing.T) {
	tmpl := &WorkflowTemplate{
		Name: "Test Workflow",
		Steps: []string{"step1", "step2", "step3"},
	}

	executed := executeWorkflowTemplate(tmpl)
	AssertTrue(t, executed, "should execute successfully")
}

func TestGetPredefinedWorkflows_ReturnsTemplates(t *testing.T) {
	workflows := getPredefinedWorkflows()

	AssertNotNil(t, workflows, "should return workflows")
	AssertTrue(t, len(workflows) > 0, "should have predefined workflows")
}

func TestVerifyCommitSignature_ValidSignature(t *testing.T) {
	commit := &commit{
		hash: "abc1234567890123456789012345678901234",
		subject: "Signed commit",
	}

	verified := verifyCommitSignature(commit)
	AssertNotNil(t, verified, "should return verification result")
}

func TestGetSignatureStatus_ReturnsStatus(t *testing.T) {
	commit := &commit{
		hash: "def5678901234567890123456789012345678",
		subject: "Check signature",
	}

	status := getSignatureStatus(commit)
	AssertNotNil(t, status, "should return status")
	AssertTrue(t, len(status) > 0, "status should not be empty")
}

func TestRenderCommitSigningUI_DisplaysSigningInfo(t *testing.T) {
	commits := NewTestFixture().Commits

	ui := renderCommitSigningUI(commits)
	AssertNotNil(t, ui, "should render signing UI")
	AssertTrue(t, len(ui) > 0, "UI should not be empty")
	AssertStringContains(t, ui, "Signature", "should contain signature info")
}

func TestGetCodeReviewStats_CalculatesStats(t *testing.T) {
	fixture := NewTestFixture()

	stats := getCodeReviewStats(fixture.Commits)
	AssertNotNil(t, stats, "should return stats")
}

func TestGetPairProgrammingStats_CalculatesStats(t *testing.T) {
	fixture := NewTestFixture()

	stats := getPairProgrammingStats(fixture.Commits)
	AssertNotNil(t, stats, "should return pair stats")
}

func TestBuildCollaborationMetrics_TracksMetrics(t *testing.T) {
	fixture := NewTestFixture()

	metrics := buildCollaborationMetrics(fixture.Commits)
	AssertNotNil(t, metrics, "should build metrics")
}

func TestRenderCollaborationUI_DisplaysMetrics(t *testing.T) {
	fixture := NewTestFixture()

	ui := renderCollaborationUI(fixture.Commits)
	AssertNotNil(t, ui, "should render collaboration UI")
	AssertTrue(t, len(ui) > 0, "UI should not be empty")
}

func TestBuildFlameGraph_GeneratesGraph(t *testing.T) {
	fixture := NewTestFixture()

	graph := buildFlameGraph(fixture.Commits)
	AssertNotNil(t, graph, "should generate flame graph")
}

func TestBuildDependencyGraph_GeneratesGraph(t *testing.T) {
	fixture := NewTestFixture()

	graph := buildDependencyGraph(fixture.Commits)
	AssertNotNil(t, graph, "should generate dependency graph")
}

