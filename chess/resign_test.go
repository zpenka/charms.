package chess

import (
	"strings"
	"testing"
)

func TestResign_RKeyResigns(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(key("r"))
	got := updated.(model)
	if !got.resigned {
		t.Error("pressing r should set resigned = true")
	}
}

func TestResign_MessageShowsWinner(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(key("r"))
	got := updated.(model)
	if !strings.Contains(got.message, "resigns") {
		t.Errorf("message should contain 'resigns', got %q", got.message)
	}
}

func TestResign_BlocksMoves(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(key("r"))
	m = updated.(model)
	before := m.game.Position().Hash()
	m.cursor = [2]int{6, 4}
	updated, _ = m.Update(key("enter"))
	got := updated.(model)
	if got.game.Position().Hash() != before {
		t.Error("moves should be blocked after resignation")
	}
}

func TestResign_BlockedWhileThinking(t *testing.T) {
	m := newModel()
	m.thinking = true
	updated, _ := m.Update(key("r"))
	got := updated.(model)
	if got.resigned {
		t.Error("r should be blocked while computer is thinking")
	}
}

func TestResign_NoOpIfAlreadyResigned(t *testing.T) {
	m := newModel()
	m.resigned = true
	m.message = "original"
	updated, _ := m.Update(key("r"))
	got := updated.(model)
	if got.message != "original" {
		t.Error("second r press should not change message")
	}
}

func TestView_ContainsResignHint(t *testing.T) {
	if v := newModel().View(); !strings.Contains(v, "resign") {
		t.Error("view should show resign keyboard hint")
	}
}
