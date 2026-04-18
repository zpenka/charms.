package chess

import (
	"math"

	"github.com/notnil/chess"
)

const searchDepth = 4

var pieceVal = map[chess.PieceType]int{
	chess.Pawn:   100,
	chess.Knight: 320,
	chess.Bishop: 330,
	chess.Rook:   500,
	chess.Queen:  900,
	chess.King:   20000,
}

// Piece-square tables indexed by chess.Square (A1=0 .. H8=63).
// For black pieces, mirror with sq^56 (flips rank).
// Values are bonuses in centipawns added to the piece's material value.

var pawnPST = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0, // rank 1 (pawns never here)
	5, 10, 10, -20, -20, 10, 10, 5, // rank 2
	5, -5, -10, 0, 0, -10, -5, 5, // rank 3
	0, 0, 0, 20, 20, 0, 0, 0, // rank 4
	5, 5, 10, 25, 25, 10, 5, 5, // rank 5
	10, 10, 20, 30, 30, 20, 10, 10, // rank 6
	50, 50, 50, 50, 50, 50, 50, 50, // rank 7
	0, 0, 0, 0, 0, 0, 0, 0, // rank 8
}

var knightPST = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50, // rank 1
	-40, -20, 0, 5, 5, 0, -20, -40, // rank 2
	-30, 0, 10, 15, 15, 10, 0, -30, // rank 3
	-30, 5, 15, 20, 20, 15, 5, -30, // rank 4
	-30, 0, 15, 20, 20, 15, 0, -30, // rank 5
	-30, 5, 10, 15, 15, 10, 5, -30, // rank 6
	-40, -20, 0, 5, 5, 0, -20, -40, // rank 7
	-50, -40, -30, -30, -30, -30, -40, -50, // rank 8
}

var bishopPST = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20, // rank 1
	-10, 0, 0, 0, 0, 0, 0, -10, // rank 2
	-10, 0, 5, 10, 10, 5, 0, -10, // rank 3
	-10, 5, 5, 10, 10, 5, 5, -10, // rank 4
	-10, 0, 10, 10, 10, 10, 0, -10, // rank 5
	-10, 10, 10, 10, 10, 10, 10, -10, // rank 6
	-10, 5, 0, 0, 0, 0, 5, -10, // rank 7
	-20, -10, -10, -10, -10, -10, -10, -20, // rank 8
}

var rookPST = [64]int{
	0, 0, 0, 5, 5, 0, 0, 0, // rank 1
	-5, 0, 0, 0, 0, 0, 0, -5, // rank 2
	-5, 0, 0, 0, 0, 0, 0, -5, // rank 3
	-5, 0, 0, 0, 0, 0, 0, -5, // rank 4
	-5, 0, 0, 0, 0, 0, 0, -5, // rank 5
	-5, 0, 0, 0, 0, 0, 0, -5, // rank 6
	5, 10, 10, 10, 10, 10, 10, 5, // rank 7
	0, 0, 0, 0, 0, 0, 0, 0, // rank 8
}

var queenPST = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20, // rank 1
	-10, 0, 5, 0, 0, 0, 0, -10, // rank 2
	-10, 5, 5, 5, 5, 5, 0, -10, // rank 3
	0, 0, 5, 5, 5, 5, 0, -5, // rank 4
	-5, 0, 5, 5, 5, 5, 0, -5, // rank 5
	-10, 0, 5, 5, 5, 5, 0, -10, // rank 6
	-10, 0, 0, 0, 0, 0, 0, -10, // rank 7
	-20, -10, -10, -5, -5, -10, -10, -20, // rank 8
}

var kingPST = [64]int{
	20, 30, 10, 0, 0, 10, 30, 20, // rank 1 (castled positions rewarded)
	20, 20, 0, 0, 0, 0, 20, 20, // rank 2
	-10, -20, -20, -20, -20, -20, -20, -10, // rank 3
	-20, -30, -30, -40, -40, -30, -30, -20, // rank 4
	-30, -40, -40, -50, -50, -40, -40, -30, // rank 5
	-30, -40, -40, -50, -50, -40, -40, -30, // rank 6
	-30, -40, -40, -50, -50, -40, -40, -30, // rank 7
	-30, -40, -40, -50, -50, -40, -40, -30, // rank 8
}

var pst = map[chess.PieceType]*[64]int{
	chess.Pawn:   &pawnPST,
	chess.Knight: &knightPST,
	chess.Bishop: &bishopPST,
	chess.Rook:   &rookPST,
	chess.Queen:  &queenPST,
	chess.King:   &kingPST,
}

func evaluate(pos *chess.Position) int {
	score := 0
	board := pos.Board()
	for sq := 0; sq < 64; sq++ {
		p := board.Piece(chess.Square(sq))
		if p == chess.NoPiece {
			continue
		}
		table := pst[p.Type()]
		idx := sq
		if p.Color() == chess.Black {
			idx = sq ^ 56
			score -= pieceVal[p.Type()] + table[idx]
		} else {
			score += pieceVal[p.Type()] + table[idx]
		}
	}
	return score
}

// orderMoves puts captures before quiet moves to improve alpha-beta pruning.
func orderMoves(pos *chess.Position, moves []*chess.Move) []*chess.Move {
	captures := make([]*chess.Move, 0, len(moves))
	quiet := make([]*chess.Move, 0, len(moves))
	board := pos.Board()
	for _, mv := range moves {
		if board.Piece(mv.S2()) != chess.NoPiece {
			captures = append(captures, mv)
		} else {
			quiet = append(quiet, mv)
		}
	}
	return append(captures, quiet...)
}

func minimax(pos *chess.Position, depth, alpha, beta int) int {
	status := pos.Status()
	if status == chess.Checkmate {
		if pos.Turn() == chess.White {
			return math.MinInt32/2 + (searchDepth - depth)
		}
		return math.MaxInt32/2 - (searchDepth - depth)
	}
	if status != chess.NoMethod {
		return 0
	}
	if depth == 0 {
		return evaluate(pos)
	}

	moves := orderMoves(pos, pos.ValidMoves())
	if pos.Turn() == chess.White {
		best := math.MinInt32
		for _, mv := range moves {
			score := minimax(pos.Update(mv), depth-1, alpha, beta)
			if score > best {
				best = score
			}
			if score > alpha {
				alpha = score
			}
			if beta <= alpha {
				break
			}
		}
		return best
	}
	best := math.MaxInt32
	for _, mv := range moves {
		score := minimax(pos.Update(mv), depth-1, alpha, beta)
		if score < best {
			best = score
		}
		if score < beta {
			beta = score
		}
		if beta <= alpha {
			break
		}
	}
	return best
}

func bestMoveAtDepth(g *chess.Game, depth int) *chess.Move {
	pos := g.Position()
	moves := orderMoves(pos, pos.ValidMoves())
	if len(moves) == 0 {
		return nil
	}

	var best *chess.Move
	if pos.Turn() == chess.White {
		bestScore := math.MinInt32
		for _, mv := range moves {
			score := minimax(pos.Update(mv), depth-1, math.MinInt32, math.MaxInt32)
			if score > bestScore {
				bestScore = score
				best = mv
			}
		}
	} else {
		bestScore := math.MaxInt32
		for _, mv := range moves {
			score := minimax(pos.Update(mv), depth-1, math.MinInt32, math.MaxInt32)
			if score < bestScore {
				bestScore = score
				best = mv
			}
		}
	}
	return best
}

func bestMove(g *chess.Game) *chess.Move {
	return bestMoveAtDepth(g, searchDepth)
}

func depthForDifficulty(d int) int {
	return d + 1
}
