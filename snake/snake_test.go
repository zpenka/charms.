package snake

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
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestView_RendersSnakeHead(t *testing.T) {
	m := newGame()
	if !strings.Contains(m.View(), "@") {
		t.Error("view should render @ for snake head")
	}
}

func TestView_RendersFood(t *testing.T) {
	m := newGame()
	if !strings.Contains(m.View(), "*") {
		t.Error("view should render * for food")
	}
}

func TestView_ShowsScore(t *testing.T) {
	m := newGame()
	m.score = 7
	if !strings.Contains(m.View(), "7") {
		t.Error("view should show score")
	}
}

func TestView_ShowsGameOver(t *testing.T) {
	m := newGame()
	m.state = StateGameOver
	if !strings.Contains(m.View(), "Game over") {
		t.Error("view should show Game over message")
	}
}

func TestUpdate_UpArrowChangesDir(t *testing.T) {
	m := newGame()
	updated, _ := m.Update(key("up"))
	if updated.(model).nextDir != DirUp {
		t.Errorf("nextDir = %v, want DirUp", updated.(model).nextDir)
	}
}

func TestUpdate_DownArrowChangesDir(t *testing.T) {
	m := newGame()
	updated, _ := m.Update(key("down"))
	if updated.(model).nextDir != DirDown {
		t.Errorf("nextDir = %v, want DirDown", updated.(model).nextDir)
	}
}

func TestUpdate_LeftArrowChangesDir(t *testing.T) {
	m := newGame()
	m.dir = DirUp
	m.nextDir = DirUp
	updated, _ := m.Update(key("left"))
	if updated.(model).nextDir != DirLeft {
		t.Errorf("nextDir = %v, want DirLeft", updated.(model).nextDir)
	}
}

func TestUpdate_GameOverSpaceGoesToLeaderboard(t *testing.T) {
	m := newGame()
	m.state = StateGameOver
	m.scorePath = ""
	updated, _ := m.Update(key(" "))
	if updated.(model).state != StateLeaderboard {
		t.Errorf("state = %v, want StateLeaderboard", updated.(model).state)
	}
}

func TestUpdate_LeaderboardSpaceStartsNewGame(t *testing.T) {
	m := newGame()
	m.state = StateLeaderboard
	updated, _ := m.Update(key(" "))
	if updated.(model).state != StatePlaying {
		t.Errorf("state = %v, want StatePlaying", updated.(model).state)
	}
}

func TestView_LeaderboardShowsScores(t *testing.T) {
	m := newGame()
	m.state = StateLeaderboard
	m.scores = []ScoreEntry{{Score: 42}}
	if !strings.Contains(m.View(), "42") {
		t.Error("leaderboard should show score 42")
	}
}

func TestView_LeaderboardShowsGameOverPrompt(t *testing.T) {
	m := newGame()
	m.state = StateGameOver
	if !strings.Contains(m.View(), "leaderboard") {
		t.Error("game over view should mention leaderboard")
	}
}
