package main

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
)

func TestFormatMoveHistory_EmptyAtStart(t *testing.T) {
	g := chess.NewGame()
	if got := formatMoveHistory(g); len(got) != 0 {
		t.Errorf("expected empty history at start, got %d entries", len(got))
	}
}

func TestFormatMoveHistory_AfterWhiteMove(t *testing.T) {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{{chess.E2, chess.E4}})
	lines := formatMoveHistory(g)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "1.") {
		t.Errorf("line should contain move number: %q", lines[0])
	}
	if !strings.Contains(lines[0], "e4") {
		t.Errorf("line should contain move e4: %q", lines[0])
	}
}

func TestFormatMoveHistory_AfterBothMoves(t *testing.T) {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.E2, chess.E4},
		{chess.E7, chess.E5},
	})
	lines := formatMoveHistory(g)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line for a full move pair, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "e4") {
		t.Errorf("line should contain e4: %q", lines[0])
	}
	if !strings.Contains(lines[0], "e5") {
		t.Errorf("line should contain e5: %q", lines[0])
	}
}

func TestFormatMoveHistory_MultipleFullMoves(t *testing.T) {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.E2, chess.E4},
		{chess.E7, chess.E5},
		{chess.G1, chess.F3},
		{chess.B8, chess.C6},
	})
	lines := formatMoveHistory(g)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines for 2 full moves, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "1.") {
		t.Errorf("first line should show move 1: %q", lines[0])
	}
	if !strings.Contains(lines[1], "2.") {
		t.Errorf("second line should show move 2: %q", lines[1])
	}
}

func TestView_ShowsMoveHistoryAfterMoves(t *testing.T) {
	m := newModel()
	m.cursor = [2]int{6, 4}
	updated, _ := m.handleSelect()
	m = updated.(model)
	m.cursor = [2]int{4, 4}
	updated, _ = m.handleSelect()
	m = updated.(model)

	view := m.View()
	if !strings.Contains(view, "e4") {
		t.Error("view should show move history containing e4")
	}
	if !strings.Contains(view, "1.") {
		t.Error("view should show move number in history")
	}
}
