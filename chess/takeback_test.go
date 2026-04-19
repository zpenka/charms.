package chess

import (
	"testing"

	"github.com/notnil/chess"
)

func TestTakeback_NopWhenNoMoves(t *testing.T) {
	m := newModel()
	before := m.game.Position().Hash()
	m.takeback()
	if m.game.Position().Hash() != before {
		t.Error("takeback with no moves should not change position")
	}
}

func TestTakeback_UndoesLastMove(t *testing.T) {
	m := newModel()
	m.executeMove(chess.E2, chess.E4)
	if m.game.Position().Turn() != chess.Black {
		t.Fatal("expected black's turn after e2-e4")
	}
	m.takeback()
	if m.game.Position().Turn() != chess.White {
		t.Errorf("turn = %v, want White after takeback", m.game.Position().Turn())
	}
	if m.game.Position().Board().Piece(chess.E4) != chess.NoPiece {
		t.Error("e4 should be empty after takeback")
	}
}

func TestTakeback_UndoesTwoMovesVsComputer(t *testing.T) {
	m := newModel()
	m.vsComputer = true
	m.computerColor = chess.Black
	m.executeMove(chess.E2, chess.E4) // white
	m.executeMove(chess.E7, chess.E5) // black (simulating computer reply)
	if len(m.game.Moves()) != 2 {
		t.Fatalf("expected 2 moves, got %d", len(m.game.Moves()))
	}
	m.takeback()
	if m.game.Position().Turn() != chess.White {
		t.Errorf("turn = %v, want White after 2-move takeback", m.game.Position().Turn())
	}
	if len(m.game.Moves()) != 0 {
		t.Errorf("moves = %d, want 0 after 2-move takeback", len(m.game.Moves()))
	}
}

func TestTakeback_ClearsLastMoveHighlight(t *testing.T) {
	m := newModel()
	m.executeMove(chess.E2, chess.E4)
	m.takeback()
	if m.lastFrom != nil || m.lastTo != nil {
		t.Error("takeback should clear lastFrom and lastTo highlights")
	}
}

func TestUpdate_TKeyCallsTakeback(t *testing.T) {
	m := newModel()
	m.executeMove(chess.E2, chess.E4)
	updated, _ := m.Update(key("t"))
	got := updated.(model)
	if got.game.Position().Turn() != chess.White {
		t.Error("t key should call takeback, restoring white's turn")
	}
}

func TestUpdate_TKeyNopWhenNoMoves(t *testing.T) {
	m := newModel()
	before := m.game.Position().Hash()
	updated, _ := m.Update(key("t"))
	if updated.(model).game.Position().Hash() != before {
		t.Error("t key with no moves should not change game state")
	}
}
