package chess

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
)

// scholars mate puts white king in check
func scholarsMateGame() *chess.Game {
	g := chess.NewGame()
	// e4
	g.Move(mustMove(g, chess.E2, chess.E4))
	// e5
	g.Move(mustMove(g, chess.E7, chess.E5))
	// Bc4
	g.Move(mustMove(g, chess.F1, chess.C4))
	// Nc6
	g.Move(mustMove(g, chess.B8, chess.C6))
	// Qh5
	g.Move(mustMove(g, chess.D1, chess.H5))
	// Nf6? (blunder, allows Qxf7#)
	g.Move(mustMove(g, chess.G8, chess.F6))
	return g
}

func mustMove(g *chess.Game, from, to chess.Square) *chess.Move {
	for _, mv := range g.ValidMoves() {
		if mv.S1() == from && mv.S2() == to {
			return mv
		}
	}
	panic("move not found")
}

func TestInCheckSquare_NotInCheck(t *testing.T) {
	g := chess.NewGame()
	sq, ok := inCheckSquare(g)
	if ok {
		t.Errorf("expected no check, got square %v", sq)
	}
}

func TestInCheckSquare_DetectsCheck(t *testing.T) {
	g := scholarsMateGame()
	// Qxf7+ — black king on e8 is in check
	g.Move(mustMove(g, chess.H5, chess.F7))
	_, ok := inCheckSquare(g)
	if !ok {
		t.Error("expected check to be detected after Qxf7+")
	}
}

func TestInCheckSquare_ReturnsKingSquare(t *testing.T) {
	g := scholarsMateGame()
	g.Move(mustMove(g, chess.H5, chess.F7))
	sq, _ := inCheckSquare(g)
	// black king should be on e8
	if sq != chess.E8 {
		t.Errorf("king square = %v, want e8", sq)
	}
}

func TestView_CheckKingHighlightedText(t *testing.T) {
	g := scholarsMateGame()
	g.Move(mustMove(g, chess.H5, chess.F7))
	m := newModel()
	m.game = g
	// Should not panic and should include the board
	v := m.View()
	if !strings.Contains(v, "Chess") {
		t.Error("view should contain board header")
	}
}
