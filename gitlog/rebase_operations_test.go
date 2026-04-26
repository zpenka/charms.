package gitlog

import (
	"testing"
)

// Tests for critical rebase operations

func TestReorderCommit_MoveDown(t *testing.T) {
	seq := []rebaseOp{
		{action: "pick", hash: "aaa", subject: "First"},
		{action: "pick", hash: "bbb", subject: "Second"},
		{action: "pick", hash: "ccc", subject: "Third"},
	}

	result := reorderCommit(seq, 0, 1)
	AssertNotNil(t, result, "reorder should return sequence")
	AssertTrue(t, len(result) == 3, "should preserve length")
	AssertEqual(t, result[0].hash, "bbb", "first should move to second position")
}

func TestReorderCommit_MoveUp(t *testing.T) {
	seq := []rebaseOp{
		{action: "pick", hash: "aaa", subject: "First"},
		{action: "pick", hash: "bbb", subject: "Second"},
		{action: "pick", hash: "ccc", subject: "Third"},
	}

	result := reorderCommit(seq, 2, 0)
	AssertNotNil(t, result, "reorder should work")
	AssertTrue(t, len(result) == 3, "should preserve length")
}

func TestReorderCommit_NoChange(t *testing.T) {
	seq := []rebaseOp{
		{action: "pick", hash: "aaa", subject: "A"},
		{action: "pick", hash: "bbb", subject: "B"},
	}

	result := reorderCommit(seq, 0, 0)
	AssertNotNil(t, result, "should handle same index")
	AssertTrue(t, len(result) == 2, "should preserve length")
}

func TestSquashCommit_MarkForSquash(t *testing.T) {
	seq := []rebaseOp{
		{action: "pick", hash: "aaa", subject: "A"},
		{action: "pick", hash: "bbb", subject: "B"},
		{action: "pick", hash: "ccc", subject: "C"},
	}

	result := squashCommit(seq, 1)
	AssertNotNil(t, result, "squash should return sequence")
	AssertTrue(t, len(result) == 3, "should preserve length")
	AssertEqual(t, result[1].action, "squash", "should mark as squash")
	AssertEqual(t, result[0].action, "pick", "should not change first")
}

func TestFixupCommit_MarkForFixup(t *testing.T) {
	seq := []rebaseOp{
		{action: "pick", hash: "aaa", subject: "A"},
		{action: "pick", hash: "bbb", subject: "B"},
		{action: "pick", hash: "ccc", subject: "C"},
	}

	result := fixupCommit(seq, 2)
	AssertNotNil(t, result, "fixup should return sequence")
	AssertTrue(t, len(result) == 3, "should preserve length")
	AssertEqual(t, result[2].action, "fixup", "should mark as fixup")
}

func TestPreviewRebase_WithOperations(t *testing.T) {
	seq := []rebaseOp{
		{action: "pick", hash: "aaa1111", subject: "Commit A"},
		{action: "squash", hash: "bbb2222", subject: "Commit B"},
		{action: "fixup", hash: "ccc3333", subject: "Commit C"},
	}

	result := previewRebase(seq)
	AssertNotNil(t, result, "preview should return string")
	AssertStringContains(t, result, "squash", "should show squash operation")
	AssertStringContains(t, result, "fixup", "should show fixup operation")
}

func TestPreviewRebase_Empty(t *testing.T) {
	seq := []rebaseOp{}

	result := previewRebase(seq)
	AssertNotNil(t, result, "preview should return value")
}

func TestPreviewCherryPick_WithCommits(t *testing.T) {
	commits := []commit{
		{hash: "aaa1111", shortHash: "aaa1111", subject: "Feature X", author: "Alice", when: "1h ago"},
		{hash: "bbb2222", shortHash: "bbb2222", subject: "Feature Y", author: "Bob", when: "2h ago"},
	}
	picks := []string{"aaa1111"}

	result := previewCherryPick(commits, picks)
	AssertNotNil(t, result, "preview should return string")
	AssertStringContains(t, result, "Feature X", "should include picked commit")
}

func TestPreviewCherryPick_Empty(t *testing.T) {
	commits := []commit{}
	picks := []string{}

	result := previewCherryPick(commits, picks)
	AssertNotNil(t, result, "preview should handle empty")
}

func TestRevertCommit_WithHash(t *testing.T) {
	hash := "abc1111"
	result := revertCommit(hash)
	AssertNotNil(t, result, "revert should return string")
}

func TestRevertCommit_EmptyHash(t *testing.T) {
	result := revertCommit("")
	AssertNotNil(t, result, "revert should handle empty")
}

func TestRebaseOp_Structure(t *testing.T) {
	op := rebaseOp{
		action:  "pick",
		hash:    "abc1111",
		subject: "Test commit",
	}
	AssertEqual(t, op.action, "pick", "action should be set")
	AssertEqual(t, op.hash, "abc1111", "hash should be set")
	AssertEqual(t, op.subject, "Test commit", "subject should be set")
}

func TestRebaseSequence_MultipleActions(t *testing.T) {
	ops := []rebaseOp{
		{action: "pick", hash: "aaa", subject: "A"},
		{action: "squash", hash: "bbb", subject: "B"},
		{action: "fixup", hash: "ccc", subject: "C"},
		{action: "reword", hash: "ddd", subject: "D"},
	}

	AssertTrue(t, len(ops) == 4, "should have 4 operations")
	AssertEqual(t, ops[1].action, "squash", "second should be squash")
	AssertEqual(t, ops[2].action, "fixup", "third should be fixup")
}

func TestSquashCommit_MultipleSquashes(t *testing.T) {
	seq := []rebaseOp{
		{action: "pick", hash: "aaa", subject: "A"},
		{action: "pick", hash: "bbb", subject: "B"},
		{action: "pick", hash: "ccc", subject: "C"},
	}

	result := squashCommit(seq, 1)
	result = squashCommit(result, 2)
	AssertEqual(t, result[1].action, "squash", "should mark index 1")
	AssertEqual(t, result[2].action, "squash", "should mark index 2")
}
