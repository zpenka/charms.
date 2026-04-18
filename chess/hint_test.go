package main

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
)

func TestHint_HKeyRequestsHint(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(key("?"))
	got := updated.(model)
	if !got.hinting {
		t.Error("pressing h should set hinting = true")
	}
}

func TestHint_BlockedWhileThinking(t *testing.T) {
	m := newModel()
	m.thinking = true
	updated, _ := m.Update(key("?"))
	got := updated.(model)
	if got.hinting {
		t.Error("h should not trigger hint while thinking")
	}
}

func TestHint_HintMsgSetsSquares(t *testing.T) {
	m := newModel()
	mv := chess.NewGame().ValidMoves()[0]
	updated, _ := m.Update(hintMsg{mv})
	got := updated.(model)
	if got.hintFrom == nil || *got.hintFrom != mv.S1() {
		t.Errorf("hintFrom = %v, want %v", got.hintFrom, mv.S1())
	}
	if got.hintTo == nil || *got.hintTo != mv.S2() {
		t.Errorf("hintTo = %v, want %v", got.hintTo, mv.S2())
	}
	if got.hinting {
		t.Error("hinting should be false after hintMsg received")
	}
}

func TestHint_ExecuteMovesClearsHint(t *testing.T) {
	m := newModel()
	f := chess.E2
	t2 := chess.E4
	m.hintFrom = &f
	m.hintTo = &t2
	m.executeMove(chess.E2, chess.E4)
	if m.hintFrom != nil || m.hintTo != nil {
		t.Error("executeMove should clear hintFrom/hintTo")
	}
}

func TestHint_ViewShowsFindingHint(t *testing.T) {
	m := newModel()
	m.hinting = true
	if v := m.View(); !strings.Contains(v, "Finding hint") {
		t.Error("view should show 'Finding hint...' while hinting")
	}
}

func TestComputeHint_ReturnsCmd(t *testing.T) {
	if computeHint(chess.NewGame(), 2) == nil {
		t.Error("computeHint should return a non-nil Cmd")
	}
}

func TestHint_NilMoveInHintMsgIsHandled(t *testing.T) {
	m := newModel()
	updated, _ := m.Update(hintMsg{nil})
	got := updated.(model)
	if got.hintFrom != nil || got.hintTo != nil {
		t.Error("nil move in hintMsg should leave hint squares nil")
	}
}
