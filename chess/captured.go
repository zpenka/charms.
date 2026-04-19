package chess

import (
	"github.com/notnil/chess"
)

var startingCounts = map[chess.PieceType]int{
	chess.King:   1,
	chess.Queen:  1,
	chess.Rook:   2,
	chess.Bishop: 2,
	chess.Knight: 2,
	chess.Pawn:   8,
}

// capturedPieces returns the pieces captured by white and by black.
// byWhite = black pieces missing from the board (white took them).
// byBlack = white pieces missing from the board (black took them).
func capturedPieces(g *chess.Game) (byWhite, byBlack []chess.Piece) {
	board := g.Position().Board()

	current := map[chess.Piece]int{}
	for sq := chess.A1; sq <= chess.H8; sq++ {
		p := board.Piece(sq)
		if p != chess.NoPiece {
			current[p]++
		}
	}

	for _, pt := range []chess.PieceType{chess.Queen, chess.Rook, chess.Bishop, chess.Knight, chess.Pawn} {
		wp := chess.NewPiece(pt, chess.White)
		bp := chess.NewPiece(pt, chess.Black)
		missing := startingCounts[pt] - current[wp]
		for i := 0; i < missing; i++ {
			byBlack = append(byBlack, wp)
		}
		missing = startingCounts[pt] - current[bp]
		for i := 0; i < missing; i++ {
			byWhite = append(byWhite, bp)
		}
	}
	return byWhite, byBlack
}

func pieceGlyph(p chess.Piece) string {
	idx := 0
	if p.Color() == chess.Black {
		idx = 1
	}
	return glyphs[p.Type()][idx]
}
