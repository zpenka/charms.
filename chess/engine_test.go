package main

import (
	"testing"

	"github.com/notnil/chess"
)

// evaluate

func TestEvaluate_StartingPosition(t *testing.T) {
	if got := evaluate(chess.StartingPosition()); got != 0 {
		t.Errorf("starting position score = %d, want 0", got)
	}
}

func TestEvaluate_ExtraWhitePawn(t *testing.T) {
	fen, err := chess.FEN("8/8/8/8/8/8/P7/4K2k w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	g := chess.NewGame(fen)
	if got := evaluate(g.Position()); got != 100 {
		t.Errorf("extra white pawn score = %d, want 100", got)
	}
}

func TestEvaluate_ExtraBlackRook(t *testing.T) {
	fen, err := chess.FEN("r6k/8/8/8/8/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	g := chess.NewGame(fen)
	if got := evaluate(g.Position()); got != -500 {
		t.Errorf("extra black rook score = %d, want -500", got)
	}
}

func TestEvaluate_WhiteAheadByKnight(t *testing.T) {
	fen, err := chess.FEN("7k/8/8/8/8/8/8/2N1K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	g := chess.NewGame(fen)
	if got := evaluate(g.Position()); got != 320 {
		t.Errorf("white extra knight score = %d, want 320", got)
	}
}

func TestEvaluate_SymmetricPositionIsZero(t *testing.T) {
	// Any position where both sides have identical material sums to 0.
	fen, err := chess.FEN("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	g := chess.NewGame(fen)
	if got := evaluate(g.Position()); got != 0 {
		t.Errorf("symmetric rook+king position score = %d, want 0", got)
	}
}

// bestMove

func TestBestMove_FindsCheckmateInOne(t *testing.T) {
	// After 1.f3 e5 2.g4 it is Black's turn; Qh4# is the only checkmate.
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.F2, chess.F3},
		{chess.E7, chess.E5},
		{chess.G2, chess.G4},
	})
	mv := bestMove(g)
	if mv == nil {
		t.Fatal("bestMove returned nil")
	}
	if mv.S1() != chess.D8 || mv.S2() != chess.H4 {
		t.Errorf("expected Qh4 (d8->h4), got %v->%v", mv.S1(), mv.S2())
	}
}

func TestBestMove_TakesHangingQueen(t *testing.T) {
	// White queen on d1 can take undefended black queen on d8.
	fen, err := chess.FEN("3q3k/8/8/8/8/8/8/3QK3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	g := chess.NewGame(fen)
	mv := bestMove(g)
	if mv == nil {
		t.Fatal("bestMove returned nil")
	}
	if mv.S1() != chess.D1 || mv.S2() != chess.D8 {
		t.Errorf("expected Qxd8 (d1->d8), got %v->%v", mv.S1(), mv.S2())
	}
}

func TestBestMove_ReturnsNilForFinishedGame(t *testing.T) {
	g := foolsMate()
	if mv := bestMove(g); mv != nil {
		t.Errorf("expected nil for finished game, got %v", mv)
	}
}

func TestBestMove_ReturnsAMoveForNormalPosition(t *testing.T) {
	g := chess.NewGame()
	if mv := bestMove(g); mv == nil {
		t.Error("bestMove should return a move from the starting position")
	}
}
