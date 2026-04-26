package gitlog

import (
	"testing"
)

// Tests for critical stash and reflog functionality

func TestRenderStashView_Empty(t *testing.T) {
	stashes := []stashEntry{}
	result := renderStashView(stashes, 80)
	AssertNotNil(t, result, "should handle empty stash")
}

func TestRenderStashView_SingleStash(t *testing.T) {
	stashes := []stashEntry{
		{name: "stash@{0}", branch: "main", subject: "WIP on main", hash: "abc1111"},
	}
	result := renderStashView(stashes, 80)
	AssertNotNil(t, result, "should render stash")
	AssertStringContains(t, result, "stash@{0}", "should show stash name")
	AssertStringContains(t, result, "main", "should show branch")
}

func TestRenderStashView_MultipleStashes(t *testing.T) {
	stashes := []stashEntry{
		{name: "stash@{0}", branch: "main", subject: "WIP on main: feature", hash: "aaa1111"},
		{name: "stash@{1}", branch: "develop", subject: "WIP on develop: bugfix", hash: "bbb2222"},
		{name: "stash@{2}", branch: "master", subject: "WIP on master: refactor", hash: "ccc3333"},
	}
	result := renderStashView(stashes, 80)
	AssertNotNil(t, result, "should render multiple stashes")
	AssertStringContains(t, result, "main", "should show first branch")
	AssertStringContains(t, result, "develop", "should show second branch")
}

func TestRenderStashView_NarrowWidth(t *testing.T) {
	stashes := []stashEntry{
		{name: "stash@{0}", branch: "main", subject: "WIP with a very long subject name that should be truncated", hash: "abc1111"},
	}
	result := renderStashView(stashes, 40)
	AssertNotNil(t, result, "should handle narrow width")
}

func TestRenderReflogView_Empty(t *testing.T) {
	entries := []reflogEntry{}
	result := renderReflogView(entries, 80)
	AssertNotNil(t, result, "should handle empty reflog")
}

func TestRenderReflogView_SingleEntry(t *testing.T) {
	entries := []reflogEntry{
		{hash: "abc1111", action: "checkout", message: "main", date: "5 minutes ago"},
	}
	result := renderReflogView(entries, 80)
	AssertNotNil(t, result, "should render entry")
	AssertStringContains(t, result, "checkout", "should show action")
	AssertStringContains(t, result, "main", "should show message")
}

func TestRenderReflogView_MultipleEntries(t *testing.T) {
	entries := []reflogEntry{
		{hash: "aaa1111", action: "checkout", message: "feature-x", date: "1 minute ago"},
		{hash: "bbb2222", action: "reset", message: "abc1234", date: "5 minutes ago"},
		{hash: "ccc3333", action: "commit", message: "def5678", date: "10 minutes ago"},
		{hash: "ddd4444", action: "rebase", message: "main", date: "1 hour ago"},
	}
	result := renderReflogView(entries, 80)
	AssertNotNil(t, result, "should render multiple")
	AssertStringContains(t, result, "checkout", "should show checkout")
	AssertStringContains(t, result, "reset", "should show reset")
	AssertStringContains(t, result, "rebase", "should show rebase")
}

func TestRenderReflogView_NarrowWidth(t *testing.T) {
	entries := []reflogEntry{
		{hash: "abc1111", action: "checkout", message: "very-long-branch-name-that-should-wrap", date: "now"},
	}
	result := renderReflogView(entries, 40)
	AssertNotNil(t, result, "should handle narrow width")
}

func TestStashToCommitLike_ConvertStash(t *testing.T) {
	stash := stashEntry{
		name:    "stash@{0}",
		branch:  "main",
		subject: "WIP on main",
		hash:    "abc1111",
	}
	result := stashToCommitLike(stash)
	AssertEqual(t, result.subject, "WIP on main", "should use stash subject")
	AssertNotEqual(t, result.hash, "", "should have hash")
}

func TestStashToCommitLike_EmptySubject(t *testing.T) {
	stash := stashEntry{
		name:    "stash@{0}",
		branch:  "main",
		subject: "",
		hash:    "abc1111",
	}
	result := stashToCommitLike(stash)
	AssertNotNil(t, result, "should handle empty subject")
}

func TestReflogToCommitLike_ConvertEntry(t *testing.T) {
	entry := reflogEntry{
		hash:    "abc1111",
		action:  "checkout",
		message: "main",
		date:    "5 minutes ago",
	}
	result := reflogToCommitLike(entry)
	AssertEqual(t, result.subject, "main", "should use message as subject")
	AssertEqual(t, result.author, "checkout", "should use action as author")
	AssertNotEqual(t, result.hash, "", "should have hash")
}

func TestReflogToCommitLike_EmptyAction(t *testing.T) {
	entry := reflogEntry{
		hash:    "abc1111",
		action:  "",
		message: "main",
		date:    "now",
	}
	result := reflogToCommitLike(entry)
	AssertNotNil(t, result, "should handle empty action")
}

func TestStashEntry_Structure(t *testing.T) {
	entry := stashEntry{
		name:    "stash@{5}",
		branch:  "develop",
		subject: "WIP on branch",
		hash:    "abc1111",
	}
	AssertEqual(t, entry.name, "stash@{5}", "name should be set")
	AssertEqual(t, entry.branch, "develop", "branch should be set")
	AssertEqual(t, entry.subject, "WIP on branch", "subject should be set")
}

func TestReflogEntry_Structure(t *testing.T) {
	entry := reflogEntry{
		hash:    "abc1111",
		action:  "merge",
		message: "develop",
		date:    "3 days ago",
	}
	AssertEqual(t, entry.hash, "abc1111", "hash should be set")
	AssertEqual(t, entry.action, "merge", "action should be set")
	AssertEqual(t, entry.message, "develop", "message should be set")
}

func TestRenderStashView_WithBranchInfo(t *testing.T) {
	stashes := []stashEntry{
		{name: "stash@{0}", branch: "feature/login", subject: "WIP: auth work", hash: "abc1111"},
		{name: "stash@{1}", branch: "hotfix/security", subject: "WIP: security", hash: "def2222"},
	}
	result := renderStashView(stashes, 80)
	AssertNotNil(t, result, "should render with branch info")
}

func TestRenderReflogView_ComplexActions(t *testing.T) {
	entries := []reflogEntry{
		{hash: "aaa1111", action: "cherry-pick", message: "abc1234", date: "now"},
		{hash: "bbb2222", action: "rebase -i", message: "main", date: "1h ago"},
		{hash: "ccc3333", action: "pull", message: "origin main", date: "2h ago"},
	}
	result := renderReflogView(entries, 80)
	AssertNotNil(t, result, "should handle complex actions")
}
