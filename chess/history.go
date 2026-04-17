package main

import (
	"fmt"

	"github.com/notnil/chess"
)

var an = chess.AlgebraicNotation{}

// formatMoveHistory returns one string per full move pair, e.g. " 1.  e4   e5".
// An incomplete final pair (white moved, black hasn't yet) gets its own line.
func formatMoveHistory(g *chess.Game) []string {
	moves := g.Moves()
	positions := g.Positions()
	if len(moves) == 0 {
		return nil
	}

	var lines []string
	for i := 0; i < len(moves); i += 2 {
		moveNum := i/2 + 1
		white := an.Encode(positions[i], moves[i])
		if i+1 < len(moves) {
			black := an.Encode(positions[i+1], moves[i+1])
			lines = append(lines, fmt.Sprintf("%2d. %-6s %s", moveNum, white, black))
		} else {
			lines = append(lines, fmt.Sprintf("%2d. %s", moveNum, white))
		}
	}
	return lines
}
