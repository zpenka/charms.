package main

import (
	"math"

	"github.com/notnil/chess"
)

const searchDepth = 3

var pieceVal = map[chess.PieceType]int{
	chess.Pawn:   100,
	chess.Knight: 320,
	chess.Bishop: 330,
	chess.Rook:   500,
	chess.Queen:  900,
	chess.King:   20000,
}

func evaluate(pos *chess.Position) int {
	score := 0
	board := pos.Board()
	for sq := 0; sq < 64; sq++ {
		p := board.Piece(chess.Square(sq))
		if p == chess.NoPiece {
			continue
		}
		v := pieceVal[p.Type()]
		if p.Color() == chess.White {
			score += v
		} else {
			score -= v
		}
	}
	return score
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

	moves := pos.ValidMoves()
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

func bestMove(g *chess.Game) *chess.Move {
	pos := g.Position()
	moves := pos.ValidMoves()
	if len(moves) == 0 {
		return nil
	}

	var best *chess.Move
	if pos.Turn() == chess.White {
		bestScore := math.MinInt32
		for _, mv := range moves {
			score := minimax(pos.Update(mv), searchDepth-1, math.MinInt32, math.MaxInt32)
			if score > bestScore {
				bestScore = score
				best = mv
			}
		}
	} else {
		bestScore := math.MaxInt32
		for _, mv := range moves {
			score := minimax(pos.Update(mv), searchDepth-1, math.MinInt32, math.MaxInt32)
			if score < bestScore {
				bestScore = score
				best = mv
			}
		}
	}
	return best
}
