package tapper

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6B6B"))
	livesStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	scoreStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
	barStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	mugStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4")).Bold(true)
	custSafeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#2ECC71")).Bold(true)
	custWarnStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F39C12")).Bold(true)
	custDangerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E74C3C")).Bold(true)
	thirstyStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500")).Bold(true)
	vipStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
	bartenderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#A29BFE")).Bold(true)
	msgStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	alertStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D")).Bold(true)
	flashStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#E74C3C")).Bold(true)
	pauseStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#A29BFE")).Bold(true)
	serveStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)
	comboStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
)

type tickMsg struct{}

func doTick() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m model) Init() tea.Cmd { return doTick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return tickGame(m), doTick()
	case tea.KeyMsg:
		if m.state == StateModeSelect {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "1":
				m.endless = false
				return startWave(m), nil
			case "2":
				m.endless = true
				return startWave(m), nil
			}
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "p":
			if m.state == StatePlaying {
				m.paused = !m.paused
			}
		case "up", "k":
			if m.state == StatePlaying && !m.paused && m.bartender > 0 {
				m.bartender--
			}
		case "down", "j":
			if m.state == StatePlaying && !m.paused && m.bartender < Lanes-1 {
				m.bartender++
			}
		case " ", "enter":
			switch m.state {
			case StatePlaying:
				if !m.paused {
					return tap(m), nil
				}
			case StateWaveClear:
				m.wave++
				return startWave(m), nil
			case StateGameOver:
				entries := addScore(loadScores(m.scorePath), m.score, m.wave+1)
				saveScores(m.scorePath, entries)
				m.scores = entries
				m.state = StateLeaderboard
				return m, nil
			case StateLeaderboard:
				ng := newGameWithScores(m.scores, m.scorePath)
				ng.state = StateModeSelect
				return ng, nil
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var sb strings.Builder

	if m.state == StateModeSelect {
		sb.WriteString("\n ")
		sb.WriteString(titleStyle.Render("Tapper"))
		sb.WriteString("\n\n")
		sb.WriteString("  Choose a mode:\n\n")
		sb.WriteString("  [1]  Waves    (8 waves, then done)\n")
		sb.WriteString("  [2]  Endless  (keep going forever)\n\n")
		sb.WriteString("  " + msgStyle.Render("Press 1 or 2") + "\n\n")
		return sb.String()
	}

	sb.WriteString("\n ")
	sb.WriteString(titleStyle.Render("Tapper"))
	sb.WriteString("\n\n ")
	sb.WriteString(livesStyle.Render(strings.Repeat("♥ ", m.lives) + strings.Repeat("♡ ", MaxLives-m.lives)))
	sb.WriteString(" ")
	sb.WriteString(scoreStyle.Render(fmt.Sprintf("Score: %d", m.score)))
	sb.WriteString("  ")
	sb.WriteString(msgStyle.Render(fmt.Sprintf("Wave: %d", m.wave+1)))
	sb.WriteString("  ")
	sb.WriteString(msgStyle.Render(fmt.Sprintf("▸ %d remaining", m.spawnsLeft+len(m.customers))))
	if m.combo > 1 {
		sb.WriteString("  ")
		sb.WriteString(comboStyle.Render(fmt.Sprintf("Combo ×%d", m.combo)))
	}
	sb.WriteString("\n\n")

	if m.paused {
		sb.WriteString(pauseStyle.Render(" ── PAUSED ──"))
		sb.WriteString("\n\n")
	}

	// Build lookup maps for rendering
	mugAt := make(map[[2]int]bool)
	for _, mg := range m.mugs {
		mugAt[[2]int{mg.lane, mg.x}] = true
	}
	type custRender struct {
		kind customerKind
	}
	custAt := make(map[[2]int]custRender)
	for _, c := range m.customers {
		custAt[[2]int{c.lane, c.x}] = custRender{c.kind}
	}
	serveAt := make(map[[2]int]bool)
	for _, a := range m.serveAnims {
		serveAt[[2]int{a.lane, a.x}] = true
	}

	for lane := 0; lane < Lanes; lane++ {
		if m.bartender == lane {
			sb.WriteString(bartenderStyle.Render(" ☻ "))
		} else {
			sb.WriteString("   ")
		}
		sb.WriteString(barStyle.Render("│"))

		for x := 0; x < BarWidth; x++ {
			key := [2]int{lane, x}
			switch {
			case m.flashFrames > 0:
				sb.WriteString(flashStyle.Render("×"))
			case serveAt[key]:
				sb.WriteString(serveStyle.Render("*"))
			case mugAt[key]:
				sb.WriteString(mugStyle.Render("o"))
			case custAt[key] != (custRender{}):
				cr := custAt[key]
				switch cr.kind {
				case KindThirsty:
					sb.WriteString(thirstyStyle.Render("!"))
				case KindVIP:
					sb.WriteString(vipStyle.Render("$"))
				default:
					var cs lipgloss.Style
					switch {
					case x >= BarWidth*2/3:
						cs = custSafeStyle
					case x >= BarWidth/3:
						cs = custWarnStyle
					default:
						cs = custDangerStyle
					}
					sb.WriteString(cs.Render("@"))
				}
			default:
				sb.WriteString(barStyle.Render("·"))
			}
		}

		sb.WriteString(barStyle.Render("│"))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	switch m.state {
	case StateWaveClear:
		total := spawnsForWave(m.wave)
		sb.WriteString(alertStyle.Render(fmt.Sprintf(" Wave %d complete!", m.wave+1)))
		sb.WriteString("\n\n")
		sb.WriteString(msgStyle.Render(fmt.Sprintf("  Served:  %d / %d", m.waveServes, total)))
		if m.waveServes == total {
			sb.WriteString(alertStyle.Render("  perfect!"))
		}
		sb.WriteString("\n")
		sb.WriteString(msgStyle.Render(fmt.Sprintf("  Combo:   %dx best", m.waveLongestCombo)))
		sb.WriteString("\n")
		sb.WriteString(alertStyle.Render(fmt.Sprintf("  Bonus:   +%d", m.waveBonus)))
		sb.WriteString("\n\n ")
		sb.WriteString(msgStyle.Render("Press Space to continue"))
		sb.WriteString("\n\n")
	case StateGameOver:
		sb.WriteString(livesStyle.Render(fmt.Sprintf(" Game over!  Final score: %d", m.score)))
		sb.WriteString("\n ")
		sb.WriteString(msgStyle.Render("Press Space for leaderboard  q to quit"))
		sb.WriteString("\n\n")
	case StateLeaderboard:
		sb.WriteString(alertStyle.Render(" Top Scores"))
		sb.WriteString("\n\n")
		if len(m.scores) == 0 {
			sb.WriteString(msgStyle.Render(" No scores yet."))
			sb.WriteString("\n")
		}
		for i, e := range m.scores {
			line := fmt.Sprintf(" %d.  %6d pts  wave %d", i+1, e.Score, e.Wave)
			if e.Score == m.score && e.Wave == m.wave+1 {
				sb.WriteString(alertStyle.Render(line))
			} else {
				sb.WriteString(msgStyle.Render(line))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n ")
		sb.WriteString(msgStyle.Render("Space to play again  q to quit"))
		sb.WriteString("\n\n")
	default:
		if m.paused {
			sb.WriteString(msgStyle.Render(" p  unpause   q  quit"))
		} else {
			sb.WriteString(msgStyle.Render(" ↑↓ / jk  move   Space  tap   p  pause   q  quit"))
		}
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func Run() {
	p := tea.NewProgram(newGame(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
