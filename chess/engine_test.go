package chess

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
	if got := evaluate(g.Position()); got <= 0 {
		t.Errorf("extra white pawn score = %d, want > 0", got)
	}
}

func TestEvaluate_ExtraBlackRook(t *testing.T) {
	fen, err := chess.FEN("r6k/8/8/8/8/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	g := chess.NewGame(fen)
	if got := evaluate(g.Position()); got >= 0 {
		t.Errorf("extra black rook score = %d, want < 0", got)
	}
}

func TestEvaluate_WhiteAheadByKnight(t *testing.T) {
	fen, err := chess.FEN("7k/8/8/8/8/8/8/2N1K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	g := chess.NewGame(fen)
	if got := evaluate(g.Position()); got <= 0 {
		t.Errorf("white extra knight score = %d, want > 0", got)
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

// searchDepth

func TestSearchDepth_IsFour(t *testing.T) {
	if searchDepth != 4 {
		t.Errorf("searchDepth = %d, want 4", searchDepth)
	}
}

// positional evaluation (piece-square tables)

func TestEvaluate_CentralKnightScoresHigherThanCornerKnight(t *testing.T) {
	fenCenter, err := chess.FEN("4k3/8/8/8/4N3/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	fenCorner, err := chess.FEN("4k3/8/8/8/8/8/8/N3K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	scoreCenter := evaluate(chess.NewGame(fenCenter).Position())
	scoreCorner := evaluate(chess.NewGame(fenCorner).Position())
	if scoreCenter <= scoreCorner {
		t.Errorf("central knight (%d) should score higher than corner knight (%d)", scoreCenter, scoreCorner)
	}
}

func TestEvaluate_CentralPawnScoresHigherThanEdgePawn(t *testing.T) {
	fenCenter, err := chess.FEN("4k3/8/8/8/4P3/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	fenEdge, err := chess.FEN("4k3/8/8/8/P7/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	scoreCenter := evaluate(chess.NewGame(fenCenter).Position())
	scoreEdge := evaluate(chess.NewGame(fenEdge).Position())
	if scoreCenter <= scoreEdge {
		t.Errorf("central pawn (%d) should score higher than edge pawn (%d)", scoreCenter, scoreEdge)
	}
}

func TestEvaluate_AdvancedPawnScoresHigherThanStartingPawn(t *testing.T) {
	fenAdvanced, err := chess.FEN("4k3/8/8/4P3/8/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	fenStart, err := chess.FEN("4k3/8/8/8/8/8/4P3/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	scoreAdvanced := evaluate(chess.NewGame(fenAdvanced).Position())
	scoreStart := evaluate(chess.NewGame(fenStart).Position())
	if scoreAdvanced <= scoreStart {
		t.Errorf("advanced pawn (%d) should score higher than starting pawn (%d)", scoreAdvanced, scoreStart)
	}
}

func TestEvaluate_PositionalBonusIsSymmetric(t *testing.T) {
	// White knight on E4 and black knight on E5 should produce equal and opposite scores.
	fenWhite, err := chess.FEN("4k3/8/8/8/4N3/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	fenBlack, err := chess.FEN("4k3/8/8/4n3/8/8/8/4K3 w - - 0 1") // E5 mirrors E4
	if err != nil {
		t.Fatal(err)
	}
	scoreWhite := evaluate(chess.NewGame(fenWhite).Position())
	scoreBlack := evaluate(chess.NewGame(fenBlack).Position())
	if scoreWhite != -scoreBlack {
		t.Errorf("positional bonus not symmetric: white knight = %d, black knight = %d (want negatives of each other)", scoreWhite, scoreBlack)
	}
}
