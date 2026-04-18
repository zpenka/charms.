package chess

import (
	"strings"
	"testing"
	"time"

	"github.com/notnil/chess"
)

func TestClock_InitialTime(t *testing.T) {
	m := newModel()
	if m.whiteTime != 10*time.Minute {
		t.Errorf("whiteTime = %v, want %v", m.whiteTime, 10*time.Minute)
	}
	if m.blackTime != 10*time.Minute {
		t.Errorf("blackTime = %v, want %v", m.blackTime, 10*time.Minute)
	}
}

func TestClock_InitReturnsTickCmd(t *testing.T) {
	m := newModel()
	if m.Init() == nil {
		t.Error("Init() should return a tick command, got nil")
	}
}

func TestClock_TickDecrementsActivePlayer(t *testing.T) {
	m := newModel()
	m.clockOn = true
	// Play one move so it's Black's turn
	playMoves(m.game, [][2]chess.Square{{chess.E2, chess.E4}})

	updated, cmd := m.Update(tickMsg(time.Now()))
	got := updated.(model)

	if got.blackTime != 10*time.Minute-time.Second {
		t.Errorf("blackTime = %v, want %v", got.blackTime, 10*time.Minute-time.Second)
	}
	if got.whiteTime != 10*time.Minute {
		t.Errorf("whiteTime = %v, want %v", got.whiteTime, 10*time.Minute)
	}
	if cmd == nil {
		t.Error("expected tick to return another tick command")
	}
}

func TestClock_TickDoesNotRunWhenClockOff(t *testing.T) {
	m := newModel()
	m.clockOn = false

	updated, _ := m.Update(tickMsg(time.Now()))
	got := updated.(model)

	if got.whiteTime != 10*time.Minute {
		t.Errorf("whiteTime changed when clock is off: %v", got.whiteTime)
	}
	if got.blackTime != 10*time.Minute {
		t.Errorf("blackTime changed when clock is off: %v", got.blackTime)
	}
}

func TestClock_TickDoesNotRunAfterGameOver(t *testing.T) {
	m := newModel()
	m.clockOn = true
	m.game = foolsMate()

	updated, _ := m.Update(tickMsg(time.Now()))
	got := updated.(model)

	if got.whiteTime != 10*time.Minute {
		t.Errorf("whiteTime changed after game over: %v", got.whiteTime)
	}
	if got.blackTime != 10*time.Minute {
		t.Errorf("blackTime changed after game over: %v", got.blackTime)
	}
}

func TestClock_TickDoesNotRunWhileThinking(t *testing.T) {
	m := newModel()
	m.clockOn = true
	m.thinking = true

	updated, _ := m.Update(tickMsg(time.Now()))
	got := updated.(model)

	if got.whiteTime != 10*time.Minute {
		t.Errorf("whiteTime changed while thinking: %v", got.whiteTime)
	}
	if got.blackTime != 10*time.Minute {
		t.Errorf("blackTime changed while thinking: %v", got.blackTime)
	}
}

func TestClock_ExecuteMoveStartsClock(t *testing.T) {
	m := newModel()
	if m.clockOn {
		t.Error("clockOn should be false before any moves")
	}
	m.executeMove(chess.E2, chess.E4)
	if !m.clockOn {
		t.Error("clockOn should be true after first move")
	}
}

func TestClock_TimeoutMessage(t *testing.T) {
	m := newModel()
	m.clockOn = true
	m.whiteTime = time.Second

	updated, _ := m.Update(tickMsg(time.Now()))
	got := updated.(model)

	if !strings.Contains(got.message, "time") {
		t.Errorf("expected timeout message to contain 'time', got %q", got.message)
	}
}

func TestFormatClock(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{10 * time.Minute, "10:00"},
		{90 * time.Second, "1:30"},
		{0, "0:00"},
	}
	for _, tt := range tests {
		got := formatClock(tt.d)
		if got != tt.want {
			t.Errorf("formatClock(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestClock_ViewShowsClocks(t *testing.T) {
	view := newModel().View()
	if !strings.Contains(view, "White:") {
		t.Error("view should contain 'White:' clock")
	}
	if !strings.Contains(view, "Black:") {
		t.Error("view should contain 'Black:' clock")
	}
}
