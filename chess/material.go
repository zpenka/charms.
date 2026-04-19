package chess

import "github.com/notnil/chess"

var pieceValues = map[chess.PieceType]int{
	chess.Queen:  9,
	chess.Rook:   5,
	chess.Bishop: 3,
	chess.Knight: 3,
	chess.Pawn:   1,
}

// materialScore returns the material balance from white's perspective.
// Positive = white leads; negative = black leads.
func materialScore(g *chess.Game) int {
	board := g.Position().Board()
	score := 0
	for sq := chess.A1; sq <= chess.H8; sq++ {
		p := board.Piece(sq)
		if p == chess.NoPiece {
			continue
		}
		v := pieceValues[p.Type()]
		if p.Color() == chess.White {
			score += v
		} else {
			score -= v
		}
	}
	return score
}
