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

// serve animation view

func TestView_ShowsServeAnimation(t *testing.T) {
	m := newGame()
	m.serveAnims = []serveAnim{{lane: 0, x: 5, frames: 2}}
	if !strings.Contains(m.View(), "*") {
		t.Error("view should show * when a serve animation is active")
	}
}

// leaderboard

func TestUpdate_GameOverSpaceGoesToLeaderboard(t *testing.T) {
	m := newGame()
	m.state = StateGameOver
	m.scorePath = "" // no file I/O; loadScores on "" returns nil
	updated, _ := m.Update(key(" "))
	if updated.(model).state != StateLeaderboard {
		t.Errorf("state = %v, want StateLeaderboard", updated.(model).state)
	}
}

func TestUpdate_LeaderboardSpaceGoesToModeSelect(t *testing.T) {
	m := newGame()
	m.state = StateLeaderboard
	updated, _ := m.Update(key(" "))
	if updated.(model).state != StateModeSelect {
		t.Errorf("state = %v, want StateModeSelect", updated.(model).state)
	}
}

func TestView_LeaderboardShowsScores(t *testing.T) {
	m := newGame()
	m.state = StateLeaderboard
	m.scores = []ScoreEntry{{Score: 120, Wave: 4}, {Score: 80, Wave: 2}}
	v := m.View()
	if !strings.Contains(v, "120") {
		t.Error("leaderboard should show top score 120")
	}
	if !strings.Contains(v, "80") {
		t.Error("leaderboard should show score 80")
	}
}

func TestView_GameOverShowsLeaderboardPrompt(t *testing.T) {
	m := newGame()
	m.state = StateGameOver
	if !strings.Contains(m.View(), "leaderboard") {
		t.Error("game over view should mention leaderboard")
	}
}

// combo

func TestView_ShowsComboWhenActive(t *testing.T) {
	m := newGame()
	m.combo = 3
	if !strings.Contains(m.View(), "Combo") {
		t.Error("view should show Combo label when combo > 1")
	}
}

func TestView_HidesComboWhenZero(t *testing.T) {
	m := newGame()
	m.combo = 0
	if strings.Contains(m.View(), "Combo") {
		t.Error("view should not show Combo when combo is 0")
	}
}

// special customers

func TestView_ShowsThirstyCustomer(t *testing.T) {
	m := newGame()
	m.customers = []customer{{lane: 0, x: 5, kind: KindThirsty}}
	if !strings.Contains(m.View(), "!") {
		t.Error("thirsty customer should render as !")
	}
}

func TestView_ShowsVIPCustomer(t *testing.T) {
	m := newGame()
	m.customers = []customer{{lane: 0, x: 5, kind: KindVIP}}
	if !strings.Contains(m.View(), "$") {
		t.Error("VIP customer should render as $")
	}
}

// wave summary

func TestView_WaveClearShowsServes(t *testing.T) {
	m := newGame()
	m.state = StateWaveClear
	m.waveServes = 11
	if !strings.Contains(m.View(), "11") {
		t.Error("wave clear view should show serve count")
	}
}

func TestView_WaveClearShowsLongestCombo(t *testing.T) {
	m := newGame()
	m.state = StateWaveClear
	m.waveLongestCombo = 9
	if !strings.Contains(m.View(), "9") {
		t.Error("wave clear view should show longest combo")
	}
}

func TestView_WaveClearShowsBonus(t *testing.T) {
	m := newGame()
	m.state = StateWaveClear
	m.waveBonus = 35
	if !strings.Contains(m.View(), "35") {
		t.Error("wave clear view should show wave bonus")
	}
}

// ── endless mode UI ───────────────────────────────────────────────────────────

func TestView_ModeSelectShowsOptions(t *testing.T) {
	m := newGame()
	m.state = StateModeSelect
	view := m.View()
	if !strings.Contains(view, "Waves") {
		t.Error("mode select should show 'Waves' option")
	}
	if !strings.Contains(view, "Endless") {
		t.Error("mode select should show 'Endless' option")
	}
}

func TestUpdate_ModeSelect1ChoosesWaves(t *testing.T) {
	m := newGame()
	m.state = StateModeSelect
	updated, _ := m.Update(key("1"))
	got := updated.(model)
	if got.state == StateModeSelect {
		t.Error("pressing 1 in mode select should leave StateModeSelect")
	}
	if got.endless {
		t.Error("pressing 1 should choose waves (endless=false)")
	}
}

func TestUpdate_ModeSelect2ChoosesEndless(t *testing.T) {
	m := newGame()
	m.state = StateModeSelect
	updated, _ := m.Update(key("2"))
	got := updated.(model)
	if got.state == StateModeSelect {
		t.Error("pressing 2 in mode select should leave StateModeSelect")
	}
	if !got.endless {
		t.Error("pressing 2 should choose endless mode")
	}
}
