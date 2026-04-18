package chess

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
)

// promotionGame returns a game where it's White's turn and e7 pawn can promote to e8.
func promotionGame() *chess.Game {
	fen, _ := chess.FEN("8/4P3/8/8/8/8/8/4K1k1 w - - 0 1")
	return chess.NewGame(fen)
}

// TestIsPromotionMove_DetectsPromotion verifies that moving a pawn to the back rank is detected.
func TestIsPromotionMove_DetectsPromotion(t *testing.T) {
	if !isPromotionMove(promotionGame(), chess.E7, chess.E8) {
		t.Error("expected isPromotionMove to return true for pawn promotion move")
	}
}

// TestIsPromotionMove_NotPromotion verifies that a normal pawn move is not detected as promotion.
func TestIsPromotionMove_NotPromotion(t *testing.T) {
	if isPromotionMove(chess.NewGame(), chess.E2, chess.E4) {
		t.Error("expected isPromotionMove to return false for normal pawn move")
	}
}

// TestPromotion_EntersPromotingState verifies that moving a pawn to the back rank
// sets promoting = true instead of completing the move immediately.
func TestPromotion_EntersPromotingState(t *testing.T) {
	m := model{game: promotionGame(), validDests: make(map[chess.Square]bool), message: "White's turn"}

	// cursor on E7: rank 6, file 4 → row = 7-6 = 1, col = 4
	m.cursor = [2]int{1, 4}
	updated, _ := m.handleSelect()
	m = updated.(model)

	// cursor on E8: rank 7, file 4 → row = 7-7 = 0, col = 4
	m.cursor = [2]int{0, 4}
	updated, _ = m.handleSelect()
	got := updated.(model)

	if !got.promoting {
		t.Error("expected promoting = true after moving pawn to back rank")
	}
}

// TestPromotion_ViewShowsPicker verifies that when promoting = true, the view renders the picker.
func TestPromotion_ViewShowsPicker(t *testing.T) {
	m := model{game: promotionGame(), validDests: make(map[chess.Square]bool), promoting: true, message: "Promote pawn: Q R B N"}
	view := m.View()

	for _, choice := range []string{"Q", "R", "B", "N"} {
		if !strings.Contains(view, choice) {
			t.Errorf("view should contain %q in promotion picker, got:\n%s", choice, view)
		}
	}
}

// TestPromotion_PressQCompletesMove verifies that pressing q during promotion promotes to queen.
func TestPromotion_PressQCompletesMove(t *testing.T) {
	m := model{game: promotionGame(), validDests: make(map[chess.Square]bool), message: "White's turn"}

	// Navigate to E7, select it
	m.cursor = [2]int{1, 4}
	updated, _ := m.handleSelect()
	m = updated.(model)

	// Navigate to E8, confirm move (enters promoting state)
	m.cursor = [2]int{0, 4}
	updated, _ = m.handleSelect()
	m = updated.(model)

	if !m.promoting {
		t.Fatal("expected promoting = true before pressing q")
	}

	movesBefore := len(m.game.Moves())

	// Press q to choose queen
	updated, _ = m.Update(key("q"))
	got := updated.(model)

	if got.promoting {
		t.Error("promoting should be false after pressing q")
	}
	if len(got.game.Moves()) != movesBefore+1 {
		t.Errorf("expected one more move after promotion, got %d moves total", len(got.game.Moves()))
	}
	promoted := got.game.Position().Board().Piece(chess.E8)
	if promoted == chess.NoPiece {
		t.Fatal("expected promoted piece on e8")
	}
	if promoted.Type() != chess.Queen {
		t.Errorf("expected queen on e8, got %v", promoted.Type())
	}
}

// TestPromotion_PressRPromotesToRook verifies that pressing r during promotion promotes to rook.
func TestPromotion_PressRPromotesToRook(t *testing.T) {
	m := model{game: promotionGame(), validDests: make(map[chess.Square]bool), message: "White's turn"}

	m.cursor = [2]int{1, 4}
	updated, _ := m.handleSelect()
	m = updated.(model)

	m.cursor = [2]int{0, 4}
	updated, _ = m.handleSelect()
	m = updated.(model)

	if !m.promoting {
		t.Fatal("expected promoting = true before pressing r")
	}

	updated, _ = m.Update(key("r"))
	got := updated.(model)

	if got.promoting {
		t.Error("promoting should be false after pressing r")
	}
	promoted := got.game.Position().Board().Piece(chess.E8)
	if promoted == chess.NoPiece {
		t.Fatal("expected promoted piece on e8")
	}
	if promoted.Type() != chess.Rook {
		t.Errorf("expected rook on e8, got %v", promoted.Type())
	}
}

// TestPromotion_BlocksOtherInputWhilePromoting verifies that arrow keys do not move the cursor
// while the game is waiting for a promotion choice.
func TestPromotion_BlocksOtherInputWhilePromoting(t *testing.T) {
	m := model{
		game:       promotionGame(),
		validDests: make(map[chess.Square]bool),
		message:    "Promote pawn: Q R B N",
		promoting:  true,
		cursor:     [2]int{4, 4},
	}

	for _, k := range []string{"up", "down", "left", "right"} {
		updated, _ := m.Update(key(k))
		got := updated.(model)
		if got.cursor != [2]int{4, 4} {
			t.Errorf("key %q should not move cursor while promoting", k)
		}
	}
}

// TestPromotion_BlockedWhileThinking verifies that once promotion is complete (pressed Q)
// in vsComputer mode, thinking = true and a cmd is returned.
func TestPromotion_BlockedWhileThinking(t *testing.T) {
	m := model{
		game:          promotionGame(),
		validDests:    make(map[chess.Square]bool),
		message:       "Promote pawn: Q R B N",
		promoting:     true,
		promotionFrom: chess.E7,
		promotionTo:   chess.E8,
		vsComputer:    true,
		computerColor: chess.Black,
	}

	updated, cmd := m.Update(key("q"))
	got := updated.(model)

	if got.promoting {
		t.Error("promoting should be false after pressing q")
	}
	if !got.thinking {
		t.Error("expected thinking = true after promotion in vsComputer mode")
	}
	if cmd == nil {
		t.Error("expected a command to compute the computer's response")
	}
}
