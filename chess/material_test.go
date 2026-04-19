package chess

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
)

func TestMaterialScore_InitialPositionIsZero(t *testing.T) {
	g := chess.NewGame()
	if got := materialScore(g); got != 0 {
		t.Errorf("materialScore = %d, want 0 at start", got)
	}
}

func TestMaterialScore_AfterWhiteCapturePawn(t *testing.T) {
	// 1.e4 d5 2.exd5 — white captures d5 pawn
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.E2, chess.E4},
		{chess.D7, chess.D5},
		{chess.E4, chess.D5},
	})
	if got := materialScore(g); got != 1 {
		t.Errorf("materialScore = %d, want +1 (white up a pawn)", got)
	}
}

func TestMaterialScore_AfterBlackCaptureKnight(t *testing.T) {
	// 1.e4 e5 2.Nf3 d6 3.d4 Bg4 4.h3 Bxf3 — black bishop takes white knight
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.E2, chess.E4},
		{chess.E7, chess.E5},
		{chess.G1, chess.F3},
		{chess.D7, chess.D6},
		{chess.D2, chess.D4},
		{chess.C8, chess.G4},
		{chess.H2, chess.H3},
		{chess.G4, chess.F3}, // Bxf3 — black takes white's knight
	})
	if got := materialScore(g); got != -3 {
		t.Errorf("materialScore = %d, want -3 (white down a knight)", got)
	}
}

func TestMaterialScore_IsPositiveWhenWhiteLeads(t *testing.T) {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.E2, chess.E4},
		{chess.D7, chess.D5},
		{chess.E4, chess.D5},
	})
	if materialScore(g) <= 0 {
		t.Error("material score should be positive when white is ahead")
	}
}

func TestView_ShowsMaterialAdvantageAfterCapture(t *testing.T) {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.E2, chess.E4},
		{chess.D7, chess.D5},
		{chess.E4, chess.D5},
	})
	m := model{game: g, validDests: make(map[chess.Square]bool)}
	view := m.View()
	if !strings.Contains(view, "+1") {
		t.Error("view should show +1 material advantage for white after pawn capture")
	}
}

func TestView_NoMaterialShownWhenEven(t *testing.T) {
	m := newModel()
	view := m.View()
	// When material is 0 (even), we don't show a material indicator at all
	if strings.Contains(view, "+0") || strings.Contains(view, "-0") {
		t.Error("view should not show +0 or -0 material when position is even")
	}
}
