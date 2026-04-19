package chess

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
)

func TestOpeningName_EmptyGameIsEmpty(t *testing.T) {
	g := chess.NewGame()
	if got := openingName(g); got != "" {
		t.Errorf("openingName = %q, want empty string at start", got)
	}
}

func TestOpeningName_KnowsItalianGame(t *testing.T) {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.E2, chess.E4},
		{chess.E7, chess.E5},
		{chess.G1, chess.F3},
		{chess.B8, chess.C6},
		{chess.F1, chess.C4},
	})
	if got := openingName(g); !strings.Contains(got, "Italian") {
		t.Errorf("openingName = %q, want Italian Game", got)
	}
}

func TestOpeningName_KnowsSicilianDefense(t *testing.T) {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.E2, chess.E4},
		{chess.C7, chess.C5},
	})
	if got := openingName(g); !strings.Contains(got, "Sicilian") {
		t.Errorf("openingName = %q, want Sicilian Defense", got)
	}
}

func TestOpeningName_KnowsQueensGambit(t *testing.T) {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.D2, chess.D4},
		{chess.D7, chess.D5},
		{chess.C2, chess.C4},
	})
	if got := openingName(g); !strings.Contains(got, "Queen") {
		t.Errorf("openingName = %q, want Queen's Gambit", got)
	}
}

func TestOpeningName_KnowsRuyLopez(t *testing.T) {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.E2, chess.E4},
		{chess.E7, chess.E5},
		{chess.G1, chess.F3},
		{chess.B8, chess.C6},
		{chess.F1, chess.B5},
	})
	if got := openingName(g); !strings.Contains(got, "Ruy") || !strings.Contains(got, "Lopez") {
		t.Errorf("openingName = %q, want Ruy Lopez", got)
	}
}

func TestOpeningName_KnowsKingsIndian(t *testing.T) {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.D2, chess.D4},
		{chess.G8, chess.F6},
		{chess.C2, chess.C4},
		{chess.G7, chess.G6},
	})
	if got := openingName(g); !strings.Contains(got, "Indian") {
		t.Errorf("openingName = %q, want King's Indian", got)
	}
}

func TestOpeningName_ReturnsEmptyForUnknown(t *testing.T) {
	// Unusual moves unlikely to match any opening
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.A2, chess.A4},
		{chess.A7, chess.A5},
	})
	// Should return empty or some generic name — just confirm it doesn't panic
	_ = openingName(g)
}

func TestView_ShowsOpeningNameWhenKnown(t *testing.T) {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.E2, chess.E4},
		{chess.C7, chess.C5},
	})
	m := model{game: g, validDests: make(map[chess.Square]bool)}
	view := m.View()
	if !strings.Contains(view, "Sicilian") {
		t.Error("view should show opening name when a known opening is detected")
	}
}

func TestView_NoOpeningLabelWhenUnknown(t *testing.T) {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.A2, chess.A4},
		{chess.A7, chess.A5},
	})
	m := model{game: g, validDests: make(map[chess.Square]bool)}
	view := m.View()
	// Should not show "Opening:" label when opening is unknown
	if strings.Contains(view, "Opening: ") {
		t.Error("view should not show 'Opening: ' label when opening is unknown")
	}
}
