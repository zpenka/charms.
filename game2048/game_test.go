package game2048

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func press(s string) tea.KeyMsg {
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

func TestView_ShowsScore(t *testing.T) {
	m := newGame()
	m.score = 512
	if !strings.Contains(m.View(), "512") {
		t.Error("view should show score")
	}
}

func TestView_ShowsTileValue(t *testing.T) {
	m := newGame()
	m.board = board{}
	m.board[0][0] = 64
	if !strings.Contains(m.View(), "64") {
		t.Error("view should show tile value 64")
	}
}

func TestView_ShowsGameOver(t *testing.T) {
	m := newGame()
	m.state = StateGameOver
	if !strings.Contains(m.View(), "Game over") {
		t.Error("view should show Game over message")
	}
}

func TestView_ShowsWin(t *testing.T) {
	m := newGame()
	m.state = StateWon
	if !strings.Contains(m.View(), "2048") {
		t.Error("win view should mention 2048")
	}
}

func TestView_ShowsLeaderboardPromptOnGameOver(t *testing.T) {
	m := newGame()
	m.state = StateGameOver
	if !strings.Contains(m.View(), "leaderboard") {
		t.Error("game over view should mention leaderboard")
	}
}

func TestUpdate_LeftArrowAppliesMove(t *testing.T) {
	m := newGame()
	m.board = board{}
	m.board[0][3] = 2
	updated, _ := m.Update(press("left"))
	// tile should slide to column 0
	if updated.(model).board[0][0] != 2 {
		t.Error("left arrow should move tile to leftmost column")
	}
}

func TestUpdate_RightArrowAppliesMove(t *testing.T) {
	m := newGame()
	m.board = board{}
	m.board[0][0] = 2
	updated, _ := m.Update(press("right"))
	// tile should slide to column 3
	if updated.(model).board[0][3] != 2 {
		t.Error("right arrow should move tile to rightmost column")
	}
}

func TestUpdate_UpArrowAppliesMove(t *testing.T) {
	m := newGame()
	m.board = board{}
	m.board[3][0] = 2
	updated, _ := m.Update(press("up"))
	// tile should slide to row 0
	if updated.(model).board[0][0] != 2 {
		t.Error("up arrow should move tile to top row")
	}
}

func TestUpdate_DownArrowAppliesMove(t *testing.T) {
	m := newGame()
	m.board = board{}
	m.board[0][0] = 2
	updated, _ := m.Update(press("down"))
	// tile should slide to row 3
	if updated.(model).board[3][0] != 2 {
		t.Error("down arrow should move tile to bottom row")
	}
}

func TestUpdate_WonSpaceContinues(t *testing.T) {
	m := newGame()
	m.state = StateWon
	updated, _ := m.Update(press(" "))
	if updated.(model).state != StatePlaying {
		t.Errorf("state = %v, want StatePlaying after continuing from win", updated.(model).state)
	}
	if !updated.(model).continued {
		t.Error("continued should be true after pressing Space on win screen")
	}
}

func TestUpdate_GameOverSpaceGoesToLeaderboard(t *testing.T) {
	m := newGame()
	m.state = StateGameOver
	m.scorePath = ""
	updated, _ := m.Update(press(" "))
	if updated.(model).state != StateLeaderboard {
		t.Errorf("state = %v, want StateLeaderboard", updated.(model).state)
	}
}

func TestUpdate_LeaderboardSpaceStartsNewGame(t *testing.T) {
	m := newGame()
	m.state = StateLeaderboard
	updated, _ := m.Update(press(" "))
	if updated.(model).state != StatePlaying {
		t.Errorf("state = %v, want StatePlaying", updated.(model).state)
	}
}

func TestView_LeaderboardShowsScores(t *testing.T) {
	m := newGame()
	m.state = StateLeaderboard
	m.scores = []ScoreEntry{{Score: 4096}}
	if !strings.Contains(m.View(), "4096") {
		t.Error("leaderboard should show score 4096")
	}
}
