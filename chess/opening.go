package chess

import (
	"strings"

	"github.com/notnil/chess"
)

// openings maps a move-sequence prefix (space-separated UCI notation) to a name.
var openings = []struct {
	moves string
	name  string
}{
	// 5-move lines (more specific — checked first)
	{"e2e4 e7e5 g1f3 b8c6 f1c4", "Italian Game"},
	{"e2e4 e7e5 g1f3 b8c6 f1b5", "Ruy Lopez"},
	{"e2e4 e7e5 g1f3 b8c6 d2d4", "Scotch Game"},
	{"e2e4 e7e5 f2f4", "King's Gambit"},
	{"d2d4 d7d5 c2c4 e7e6", "Queen's Gambit Declined"},
	{"d2d4 d7d5 c2c4 c7c6", "Slav Defense"},
	{"d2d4 g8f6 c2c4 g7g6", "King's Indian Defense"},
	{"d2d4 g8f6 c2c4 e7e6", "Nimzo/Queen's Indian"},
	{"e2e4 c7c5 g1f3", "Sicilian Defense"},
	{"e2e4 e7e6", "French Defense"},
	{"e2e4 c7c6", "Caro-Kann Defense"},
	{"e2e4 d7d5", "Scandinavian Defense"},
	// 2-move lines (broad)
	{"e2e4 e7e5", "Open Game"},
	{"e2e4 c7c5", "Sicilian Defense"},
	{"d2d4 d7d5 c2c4", "Queen's Gambit"},
	{"d2d4 d7d5", "Closed Game"},
	{"d2d4 g8f6", "Indian Defenses"},
	{"e2e4", "King's Pawn Opening"},
	{"d2d4", "Queen's Pawn Opening"},
	{"c2c4", "English Opening"},
	{"g1f3", "Réti Opening"},
}

// openingName returns the name of the detected opening, or "" if unknown.
func openingName(g *chess.Game) string {
	moves := g.Moves()
	if len(moves) == 0 {
		return ""
	}

	// Build a UCI prefix string from the actual moves played
	var parts []string
	for _, mv := range moves {
		parts = append(parts, mv.String())
	}
	played := strings.Join(parts, " ")

	for _, op := range openings {
		if strings.HasPrefix(played, op.moves) || played == op.moves {
			return op.name
		}
	}
	return ""
}
