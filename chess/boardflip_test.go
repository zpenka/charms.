package main

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
)

// TestBoardFlip_DefaultNotFlipped verifies new models are not flipped.
func TestBoardFlip_DefaultNotFlipped(t *testing.T) {
	m := newModel()
	if m.flipped {
		t.Error("newModel().flipped should be false by default")
	}
}

// TestBoardFlip_FKeyToggles verifies f key toggles the flipped field.
func TestBoardFlip_FKeyToggles(t *testing.T) {
	m := newModel()

	updated, _ := m.Update(key("f"))
	got := updated.(model)
	if !got.flipped {
		t.Error("pressing f should set flipped = true")
	}

	updated, _ = got.Update(key("f"))
	got = updated.(model)
	if got.flipped {
		t.Error("pressing f again should set flipped = false")
	}
}

// TestBoardFlip_FKeyBlockedWhileThinking verifies f does not toggle while thinking.
func TestBoardFlip_FKeyBlockedWhileThinking(t *testing.T) {
	m := newModel()
	m.thinking = true

	updated, _ := m.Update(key("f"))
	got := updated.(model)
	if got.flipped {
		t.Error("f key should not toggle flipped while thinking")
	}
}

// TestBoardFlip_AutoFlipWhenPlayingBlack verifies auto-flip when player picks Black.
func TestBoardFlip_AutoFlipWhenPlayingBlack(t *testing.T) {
	m := newModel()
	m.modeSelect = true

	// Press 2 → timeSelect, then pick blitz to proceed
	updated, _ := m.Update(key("2"))
	m = updated.(model)
	updated, _ = m.Update(key("2"))
	m = updated.(model)

	// Press 2 to choose medium difficulty
	updated, _ = m.Update(key("2"))
	m = updated.(model)

	// Press B to play as Black
	updated, _ = m.Update(key("B"))
	got := updated.(model)

	if !got.flipped {
		t.Error("model should be flipped when player chooses to play as Black")
	}
}

// TestBoardFlip_ViewChangesWhenFlipped verifies the view changes with flip state.
func TestBoardFlip_ViewChangesWhenFlipped(t *testing.T) {
	unflipped := newModel()
	flipped := newModel()
	flipped.flipped = true

	uv := unflipped.View()
	fv := flipped.View()

	if uv == fv {
		t.Error("flipped and unflipped views should differ")
	}

	if !strings.Contains(uv, "a  b  c") {
		t.Errorf("unflipped view should contain 'a  b  c', got: %q", uv[:min(200, len(uv))])
	}
	if !strings.Contains(fv, "h  g  f") {
		t.Errorf("flipped view should contain 'h  g  f', got: %q", fv[:min(200, len(fv))])
	}
}

// TestBoardFlip_SquareAtCursorIsCorrect verifies boardSquare returns correct squares.
func TestBoardFlip_SquareAtCursorIsCorrect(t *testing.T) {
	m := newModel()
	mf := newModel()
	mf.flipped = true

	// Unflipped: cursor [7,4] = row=7, col=4 => rank=7-7=0, file=4 => E1
	if got := m.boardSquare(7, 4); got != chess.E1 {
		t.Errorf("unflipped boardSquare(7,4) = %v, want E1", got)
	}

	// Flipped: cursor [7,4] => rank=7, file=7-4=3 => square = 7*8+3 = 59 = D8
	want := chess.Square(59) // D8
	if got := mf.boardSquare(7, 4); got != want {
		t.Errorf("flipped boardSquare(7,4) = %v, want square 59 (D8)", got)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
