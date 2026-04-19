package game2048

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFD93D"))
	scoreStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	msgStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	alertStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D")).Bold(true)
)

func tileStyle(v int) lipgloss.Style {
	base := lipgloss.NewStyle().Bold(true).Width(6).Align(lipgloss.Center)
	switch {
	case v == 0:
		return base.Foreground(lipgloss.Color("#444444"))
	case v <= 2:
		return base.Foreground(lipgloss.Color("#DDDDDD"))
	case v <= 4:
		return base.Foreground(lipgloss.Color("#FFD93D"))
	case v <= 8:
		return base.Foreground(lipgloss.Color("#F39C12"))
	case v <= 16:
		return base.Foreground(lipgloss.Color("#E67E22"))
	case v <= 32:
		return base.Foreground(lipgloss.Color("#E74C3C"))
	case v <= 64:
		return base.Foreground(lipgloss.Color("#FF4500"))
	case v <= 128:
		return base.Foreground(lipgloss.Color("#FFD700"))
	case v <= 256:
		return base.Foreground(lipgloss.Color("#ADFF2F"))
	case v <= 512:
		return base.Foreground(lipgloss.Color("#2ECC71"))
	case v <= 1024:
		return base.Foreground(lipgloss.Color("#4ECDC4"))
	default:
		return base.Foreground(lipgloss.Color("#FF6B6B"))
	}
}

func cellText(v int) string {
	if v == 0 {
		return "·"
	}
	return fmt.Sprintf("%d", v)
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		if m.state == StateTargetSelect {
			switch km.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "1":
				m.targetTile = 512
				m.state = StatePlaying
				return m, nil
			case "2":
				m.targetTile = 1024
				m.state = StatePlaying
				return m, nil
			case "3":
				m.targetTile = 2048
				m.state = StatePlaying
				return m, nil
			case "4":
				m.targetTile = 4096
				m.state = StatePlaying
				return m, nil
			}
			return m, nil
		}
		switch km.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left", "a":
			return applyMove(m, DirLeft), nil
		case "right", "d":
			return applyMove(m, DirRight), nil
		case "up", "w":
			return applyMove(m, DirUp), nil
		case "down", "s":
			return applyMove(m, DirDown), nil
		case "z":
			return undoMove(m), nil
		case " ", "enter":
			switch m.state {
			case StateWon:
				m.continued = true
				m.state = StatePlaying
				return m, nil
			case StateGameOver:
				entries := addScore(loadScores(m.scorePath), m.score)
				saveScores(m.scorePath, entries)
				m.scores = entries
				m.state = StateLeaderboard
				return m, nil
			case StateLeaderboard:
				return newGameWithScores(m.scores, m.scorePath), nil
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var sb strings.Builder

	sb.WriteString("\n ")
	sb.WriteString(titleStyle.Render("2048"))
	sb.WriteString("  ")
	sb.WriteString(scoreStyle.Render(fmt.Sprintf("Score: %d", m.score)))
	sb.WriteString("  ")
	sb.WriteString(scoreStyle.Render(fmt.Sprintf("Best tile: %d", maxTile(m.board))))
	if m.allTimeBest > 0 {
		sb.WriteString("  ")
		sb.WriteString(msgStyle.Render(fmt.Sprintf("Best: %d", m.allTimeBest)))
	}
	sb.WriteString("\n\n")

	// Top border
	sep := borderStyle.Render("┼──────")
	topBorder := borderStyle.Render("┌──────") + strings.Repeat(borderStyle.Render("┬──────"), BoardSize-1) + borderStyle.Render("┐")
	midBorder := borderStyle.Render("├──────") + strings.Repeat(sep, BoardSize-1) + borderStyle.Render("┤")
	botBorder := borderStyle.Render("└──────") + strings.Repeat(borderStyle.Render("┴──────"), BoardSize-1) + borderStyle.Render("┘")

	sb.WriteString(" ")
	sb.WriteString(topBorder)
	sb.WriteString("\n")

	for r := 0; r < BoardSize; r++ {
		sb.WriteString(" ")
		sb.WriteString(borderStyle.Render("│"))
		for c := 0; c < BoardSize; c++ {
			v := m.board[r][c]
			sb.WriteString(tileStyle(v).Render(cellText(v)))
			sb.WriteString(borderStyle.Render("│"))
		}
		sb.WriteString("\n")
		if r < BoardSize-1 {
			sb.WriteString(" ")
			sb.WriteString(midBorder)
			sb.WriteString("\n")
		}
	}

	sb.WriteString(" ")
	sb.WriteString(botBorder)
	sb.WriteString("\n\n")

	switch m.state {
	case StateTargetSelect:
		sb.WriteString(alertStyle.Render(" Choose your target tile:"))
		sb.WriteString("\n\n")
		sb.WriteString(msgStyle.Render("  [1]  512\n  [2]  1024\n  [3]  2048\n  [4]  4096"))
		sb.WriteString("\n\n")
	case StateWon:
		goal := m.targetTile
		if goal == 0 {
			goal = 2048
		}
		sb.WriteString(alertStyle.Render(fmt.Sprintf(" You reached %d!", goal)))
		sb.WriteString("\n ")
		sb.WriteString(msgStyle.Render("Space to keep going  q to quit"))
		sb.WriteString("\n\n")
	case StateGameOver:
		sb.WriteString(alertStyle.Render(fmt.Sprintf(" Game over!  Final score: %d", m.score)))
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
			line := fmt.Sprintf(" %d.  %6d pts", i+1, e.Score)
			if e.Score == m.score {
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
		sb.WriteString(msgStyle.Render(" ↑↓←→ / wasd  move   z  undo   q  quit"))
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
