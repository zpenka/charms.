package main

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
)

// TestDifficultySelect_PressTwoGoesToDiffSelect verifies that pressing "2" in
// modeSelect leads to diffSelect=true, not colorSelect.
func TestDifficultySelect_PressTwoGoesToDiffSelect(t *testing.T) {
	m := newModel()
	m.modeSelect = true

	updated, _ := m.Update(key("2"))
	got := updated.(model)

	if got.modeSelect {
		t.Error("modeSelect should be false after pressing 2")
	}
	if !got.diffSelect {
		t.Error("diffSelect should be true after pressing 2 in modeSelect")
	}
	if got.colorSelect {
		t.Error("colorSelect should NOT be true yet — must choose difficulty first")
	}
}

// TestDifficultySelect_PressOneSetsDifficultyOne verifies that pressing "1" in
// diffSelect sets difficulty=1 and transitions to colorSelect.
func TestDifficultySelect_PressOneSetsDifficultyOne(t *testing.T) {
	m := newModel()
	m.diffSelect = true

	updated, _ := m.Update(key("1"))
	got := updated.(model)

	if got.diffSelect {
		t.Error("diffSelect should be false after choosing difficulty")
	}
	if !got.colorSelect {
		t.Error("colorSelect should be true after choosing difficulty")
	}
	if got.difficulty != 1 {
		t.Errorf("difficulty = %d, want 1", got.difficulty)
	}
}

// TestDifficultySelect_PressTwoSetsDifficultyTwo verifies that pressing "2" in
// diffSelect sets difficulty=2.
func TestDifficultySelect_PressTwoSetsDifficultyTwo(t *testing.T) {
	m := newModel()
	m.diffSelect = true

	updated, _ := m.Update(key("2"))
	got := updated.(model)

	if got.diffSelect {
		t.Error("diffSelect should be false after choosing difficulty")
	}
	if !got.colorSelect {
		t.Error("colorSelect should be true after choosing difficulty")
	}
	if got.difficulty != 2 {
		t.Errorf("difficulty = %d, want 2", got.difficulty)
	}
}

// TestDifficultySelect_PressThreeSetsDifficultyThree verifies that pressing "3"
// in diffSelect sets difficulty=3.
func TestDifficultySelect_PressThreeSetsDifficultyThree(t *testing.T) {
	m := newModel()
	m.diffSelect = true

	updated, _ := m.Update(key("3"))
	got := updated.(model)

	if got.diffSelect {
		t.Error("diffSelect should be false after choosing difficulty")
	}
	if !got.colorSelect {
		t.Error("colorSelect should be true after choosing difficulty")
	}
	if got.difficulty != 3 {
		t.Errorf("difficulty = %d, want 3", got.difficulty)
	}
}

// TestDifficultySelect_ViewShowsOptions verifies the difficulty screen shows
// Easy, Medium, and Hard options.
func TestDifficultySelect_ViewShowsOptions(t *testing.T) {
	m := newModel()
	m.diffSelect = true
	view := m.View()

	for _, want := range []string{"Easy", "Medium", "Hard"} {
		if !strings.Contains(view, want) {
			t.Errorf("difficulty select view missing %q", want)
		}
	}
	if strings.Contains(view, "a  b  c") {
		t.Error("difficulty select view should not render the board")
	}
}

// TestBestMoveAtDepth_ReturnsMove verifies that bestMoveAtDepth returns a
// non-nil move for a normal starting position at depth 2.
func TestBestMoveAtDepth_ReturnsMove(t *testing.T) {
	g := chess.NewGame()
	mv := bestMoveAtDepth(g, 2)
	if mv == nil {
		t.Error("bestMoveAtDepth returned nil for starting position at depth 2")
	}
}

// TestDepthForDifficulty verifies the mapping from difficulty level to search depth.
func TestDepthForDifficulty(t *testing.T) {
	tests := []struct {
		difficulty int
		wantDepth  int
	}{
		{1, 2},
		{2, 3},
		{3, 4},
	}
	for _, tt := range tests {
		got := depthForDifficulty(tt.difficulty)
		if got != tt.wantDepth {
			t.Errorf("depthForDifficulty(%d) = %d, want %d", tt.difficulty, got, tt.wantDepth)
		}
	}
}

// TestComputeMove_UsesSelectedDifficulty verifies that after the full flow
// (modeSelect→2→diffSelect→1→colorSelect→W), the model has difficulty=1 and vsComputer=true.
func TestComputeMove_UsesSelectedDifficulty(t *testing.T) {
	m := newModel()
	m.modeSelect = true

	// Press "2" to choose vs computer — should go to diffSelect
	updated, _ := m.Update(key("2"))
	m = updated.(model)

	if !m.diffSelect {
		t.Fatal("expected diffSelect=true after pressing 2 in modeSelect")
	}

	// Press "1" to choose Easy difficulty — should go to colorSelect
	updated, _ = m.Update(key("1"))
	m = updated.(model)

	if !m.colorSelect {
		t.Fatal("expected colorSelect=true after choosing difficulty 1")
	}
	if m.difficulty != 1 {
		t.Fatalf("expected difficulty=1, got %d", m.difficulty)
	}

	// Press "W" to play as White — starts game vs computer
	updated, _ = m.Update(key("W"))
	m = updated.(model)

	if !m.vsComputer {
		t.Error("expected vsComputer=true after full setup flow")
	}
	if m.difficulty != 1 {
		t.Errorf("expected difficulty=1 preserved, got %d", m.difficulty)
	}
	if m.colorSelect {
		t.Error("colorSelect should be false after choosing color")
	}
}
