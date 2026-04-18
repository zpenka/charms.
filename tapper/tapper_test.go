package tapper

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestUpdate_PTogglesPause(t *testing.T) {
	m := newGame()
	updated, _ := m.Update(key("p"))
	if !updated.(model).paused {
		t.Error("pressing p should pause")
	}
	updated, _ = updated.(model).Update(key("p"))
	if updated.(model).paused {
		t.Error("pressing p again should unpause")
	}
}

func TestUpdate_PauseBlocksMovement(t *testing.T) {
	m := newGame()
	m.paused = true
	m.bartender = 1
	updated, _ := m.Update(key("up"))
	if updated.(model).bartender != 1 {
		t.Error("movement should be blocked while paused")
	}
}

func TestUpdate_PauseBlocksTap(t *testing.T) {
	m := newGame()
	m.paused = true
	updated, _ := m.Update(key(" "))
	if len(updated.(model).mugs) != 0 {
		t.Error("tap should be blocked while paused")
	}
}

func TestView_ShowsPausedIndicator(t *testing.T) {
	m := newGame()
	m.paused = true
	if !strings.Contains(m.View(), "PAUSED") {
		t.Error("view should show PAUSED when paused")
	}
}

func TestView_ShowsSpawnsRemaining(t *testing.T) {
	m := newGame()
	m.spawnsLeft = 7
	v := m.View()
	if !strings.Contains(v, "7") {
		t.Error("view should display remaining spawns count")
	}
}

func TestView_ShowsFlashIndicator(t *testing.T) {
	m := newGame()
	m.flashFrames = 3
	if !strings.Contains(m.View(), "×") {
		t.Error("view should show × flash indicator when flashFrames > 0")
	}
}
