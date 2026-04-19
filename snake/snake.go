package snake

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	scoreStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
	headStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4")).Bold(true)
	bodyStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#2ECC71"))
	foodStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true)
	bonusFoodStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D")).Bold(true)
	obstacleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	borderStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	msgStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	alertStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D")).Bold(true)
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
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "w", "k":
			return changeDir(m, DirUp), nil
		case "down", "s", "j":
			return changeDir(m, DirDown), nil
		case "left", "a", "h":
			return changeDir(m, DirLeft), nil
		case "right", "d", "l":
			return changeDir(m, DirRight), nil
		case " ", "enter":
			switch m.state {
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
	sb.WriteString(titleStyle.Render("Snake"))
	sb.WriteString("  ")
	sb.WriteString(scoreStyle.Render(fmt.Sprintf("Score: %d", m.score)))
	sb.WriteString("  ")
	sb.WriteString(msgStyle.Render(fmt.Sprintf("Length: %d", len(m.snake))))
	if m.ghostTicks > 0 {
		sb.WriteString("  ")
		sb.WriteString(alertStyle.Render("GHOST"))
	}
	sb.WriteString("\n\n")

	// Build cell lookup
	headPos := m.snake[0]
	bodySet := make(map[pos]bool, len(m.snake)-1)
	for _, p := range m.snake[1:] {
		bodySet[p] = true
	}
	obsSet := make(map[pos]bool, len(m.obstacles))
	for _, p := range m.obstacles {
		obsSet[p] = true
	}

	// Top border
	sb.WriteString(" ")
	sb.WriteString(borderStyle.Render("┌" + strings.Repeat("─", Width) + "┐"))
	sb.WriteString("\n")

	for y := 0; y < Height; y++ {
		sb.WriteString(" ")
		sb.WriteString(borderStyle.Render("│"))
		for x := 0; x < Width; x++ {
			p := pos{x, y}
			switch {
			case p == headPos:
				sb.WriteString(headStyle.Render("@"))
			case bodySet[p]:
				sb.WriteString(bodyStyle.Render("o"))
			case m.bonusFoodActive && p == m.bonusFood:
				sb.WriteString(bonusFoodStyle.Render("$"))
			case p == m.food:
				sb.WriteString(foodStyle.Render("*"))
			case obsSet[p]:
				sb.WriteString(obstacleStyle.Render("█"))
			default:
				sb.WriteString(" ")
			}
		}
		sb.WriteString(borderStyle.Render("│"))
		sb.WriteString("\n")
	}

	// Bottom border
	sb.WriteString(" ")
	sb.WriteString(borderStyle.Render("└" + strings.Repeat("─", Width) + "┘"))
	sb.WriteString("\n\n")

	switch m.state {
	case StateGameOver:
		sb.WriteString(alertStyle.Render(fmt.Sprintf(" Game over!  Final length: %d", len(m.snake))))
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
			line := fmt.Sprintf(" %d.  length %d", i+1, e.Score)
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
		sb.WriteString(msgStyle.Render(" ↑↓←→ / wasd / hjkl  move   q  quit"))
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
