package chess

import (
	"testing"

	"github.com/notnil/chess"
)

// TestHighlight_ExecuteMoveSetslastFrom verifies that after a move,
// lastFrom and lastTo are set to the move's squares.
func TestHighlight_ExecuteMoveSetslastFrom(t *testing.T) {
	m := newModel()
	m.executeMove(chess.E2, chess.E4)

	if m.lastFrom == nil {
		t.Fatal("expected lastFrom to be set after a move, got nil")
	}
	if *m.lastFrom != chess.E2 {
		t.Errorf("lastFrom = %v, want E2", *m.lastFrom)
	}
	if m.lastTo == nil {
		t.Fatal("expected lastTo to be set after a move, got nil")
	}
	if *m.lastTo != chess.E4 {
		t.Errorf("lastTo = %v, want E4", *m.lastTo)
	}
}

// TestHighlight_TakebackClearsHighlight verifies that after takeback,
// both lastFrom and lastTo are nil.
func TestHighlight_TakebackClearsHighlight(t *testing.T) {
	m := newModel()
	m.executeMove(chess.E2, chess.E4)

	if m.lastFrom == nil || m.lastTo == nil {
		t.Fatal("precondition failed: lastFrom/lastTo should be set after move")
	}

	m.takeback()

	if m.lastFrom != nil {
		t.Errorf("expected lastFrom to be nil after takeback, got %v", *m.lastFrom)
	}
	if m.lastTo != nil {
		t.Errorf("expected lastTo to be nil after takeback, got %v", *m.lastTo)
	}
}

// TestHighlight_ViewDiffersWithLastMove verifies that a model with lastFrom set
// renders differently than one without it.
func TestHighlight_ViewDiffersWithLastMove(t *testing.T) {
	// Model without any last move
	m1 := newModel()

	// Model with a last move set
	m2 := newModel()
	m2.executeMove(chess.E2, chess.E4)
	// lastFrom/lastTo should now be set

	view1 := m1.View()
	view2 := m2.View()

	if view1 == view2 {
		t.Error("view with lastFrom set should differ from view without lastFrom")
	}
}

// TestHighlight_ComputerMoveSetslastFromTo verifies that when the computer move
// message is processed, lastFrom and lastTo are set correctly.
func TestHighlight_ComputerMoveSetslastFromTo(t *testing.T) {
	m := newModel()
	m.thinking = true

	// Find e2-e4 move
	var mv *chess.Move
	for _, v := range m.game.ValidMoves() {
		if v.S1() == chess.E2 && v.S2() == chess.E4 {
			mv = v
			break
		}
	}
	if mv == nil {
		t.Fatal("could not find e2-e4 in valid moves")
	}

	updated, _ := m.Update(computerMoveMsg{mv})
	got := updated.(model)

	if got.lastFrom == nil {
		t.Fatal("expected lastFrom to be set after computer move, got nil")
	}
	if *got.lastFrom != chess.E2 {
		t.Errorf("lastFrom = %v, want E2", *got.lastFrom)
	}
	if got.lastTo == nil {
		t.Fatal("expected lastTo to be set after computer move, got nil")
	}
	if *got.lastTo != chess.E4 {
		t.Errorf("lastTo = %v, want E4", *got.lastTo)
	}
}
