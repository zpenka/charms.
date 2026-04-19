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

// ── target tile selector ──────────────────────────────────────────────────────

func TestView_TargetSelectShowsOptions(t *testing.T) {
	m := newGame()
	m.state = StateTargetSelect
	view := m.View()
	if !strings.Contains(view, "512") {
		t.Error("target select view should show 512 option")
	}
	if !strings.Contains(view, "4096") {
		t.Error("target select view should show 4096 option")
	}
}

func TestUpdate_TargetSelect1Chooses512(t *testing.T) {
	m := newGame()
	m.state = StateTargetSelect
	updated, _ := m.Update(press("1"))
	got := updated.(model)
	if got.targetTile != 512 {
		t.Errorf("targetTile = %d, want 512", got.targetTile)
	}
	if got.state != StatePlaying {
		t.Errorf("state = %v, want StatePlaying after choosing target", got.state)
	}
}

func TestUpdate_TargetSelect2Chooses1024(t *testing.T) {
	m := newGame()
	m.state = StateTargetSelect
	updated, _ := m.Update(press("2"))
	got := updated.(model)
	if got.targetTile != 1024 {
		t.Errorf("targetTile = %d, want 1024", got.targetTile)
	}
}

func TestUpdate_TargetSelect3Chooses2048(t *testing.T) {
	m := newGame()
	m.state = StateTargetSelect
	updated, _ := m.Update(press("3"))
	got := updated.(model)
	if got.targetTile != 2048 {
		t.Errorf("targetTile = %d, want 2048", got.targetTile)
	}
}

func TestUpdate_TargetSelect4Chooses4096(t *testing.T) {
	m := newGame()
	m.state = StateTargetSelect
	updated, _ := m.Update(press("4"))
	got := updated.(model)
	if got.targetTile != 4096 {
		t.Errorf("targetTile = %d, want 4096", got.targetTile)
	}
}

// ── all-time high score HUD ───────────────────────────────────────────────────

func TestView_ShowsAllTimeHighScore(t *testing.T) {
	m := newGame()
	m.allTimeBest = 9876
	view := m.View()
	if !strings.Contains(view, "9876") {
		t.Error("view should show all-time high score in HUD")
	}
}

func TestView_AllTimeHighScoreLabel(t *testing.T) {
	m := newGame()
	m.allTimeBest = 100
	view := m.View()
	if !strings.Contains(view, "Best") {
		t.Error("view should show 'Best' label for all-time high score")
	}
}

func TestUpdate_ZKeyUndoes(t *testing.T) {
	m := newGame()
	m.board = board{}
	m.board[0][3] = 2
	prevBoard := m.board
	updated, _ := m.Update(press("left"))
	m = updated.(model)
	updated2, _ := m.Update(press("z"))
	got := updated2.(model)
	if got.board != prevBoard {
		t.Error("z key should undo the last move")
	}
}

func TestView_ShowsBestTileInHUD(t *testing.T) {
	m := newGame()
	m.board = board{}
	m.board[0][0] = 512
	view := m.View()
	if !strings.Contains(view, "Best") {
		t.Error("view HUD should show 'Best' tile label")
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
