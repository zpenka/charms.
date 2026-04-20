package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"charms/chess"
	game2048 "charms/game2048"
	"charms/snake"
	"charms/tapper"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// version is set at build time by GoReleaser via -ldflags "-X main.version=..."
var version = "dev"

var (
	lobbyTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	lobbySubtitle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	lobbyActive   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true)
	lobbyInactive = lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDDDD"))
	lobbyDesc     = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	lobbyScore    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD93D"))
)

type game struct {
	name     string
	desc     string
	run      func()
	topScore func() string
}

var games = []game{
	{
		name:     "Chess",
		desc:     "two-player or vs computer, full clocks",
		run:      chess.Run,
		topScore: func() string { return "—" },
	},
	{
		name:     "Tapper",
		desc:     "serve beer before customers reach the bar",
		run:      tapper.Run,
		topScore: tapper.TopScore,
	},
	{
		name:     "Snake",
		desc:     "eat, grow, don't bite yourself",
		run:      snake.Run,
		topScore: snake.TopScore,
	},
	{
		name:     "2048",
		desc:     "slide tiles, merge numbers, reach 2048",
		run:      game2048.Run,
		topScore: game2048.TopScore,
	},
}

type lobbyModel struct {
	cursor int
	chosen int // -1 = none chosen yet
	width  int
	height int
}

func newLobbyModel() lobbyModel {
	return lobbyModel{chosen: -1}
}

func (m lobbyModel) Init() tea.Cmd { return nil }

func (m lobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = wsm.Width
		m.height = wsm.Height
		return m, nil
	}
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(games)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.chosen = m.cursor
			return m, tea.Quit
		default:
			if len(km.String()) == 1 && km.String() >= "1" && km.String() <= "9" {
				idx := int(km.String()[0] - '1')
				if idx < len(games) {
					m.chosen = idx
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}

func (m lobbyModel) View() string {
	var sb strings.Builder
	sb.WriteString("\n\n ")
	sb.WriteString(lobbyTitle.Render("charms."))
	sb.WriteString("\n\n ")
	sb.WriteString(lobbySubtitle.Render("what do you want to play?"))
	sb.WriteString("\n\n")

	for i, g := range games {
		score := g.topScore()
		active := i == m.cursor
		if active {
			sb.WriteString(fmt.Sprintf("  %s [%d]  %-8s  %s",
				lobbyActive.Render("►"),
				i+1,
				lobbyActive.Render(g.name),
				lobbyDesc.Render(g.desc),
			))
		} else {
			sb.WriteString(fmt.Sprintf("     [%d]  %-8s  %s",
				i+1,
				lobbyInactive.Render(g.name),
				lobbyDesc.Render(g.desc),
			))
		}
		if score != "—" {
			sb.WriteString("  ")
			sb.WriteString(lobbyScore.Render("Best: " + score))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n\n ")
	sb.WriteString(lobbySubtitle.Render("↑↓ / jk  navigate   Enter  play   q  quit"))
	sb.WriteString("\n\n")
	content := sb.String()
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("charms " + version)
		return
	}

	for {
		p := tea.NewProgram(newLobbyModel(), tea.WithAltScreen())
		result, err := p.Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		m := result.(lobbyModel)
		if m.chosen < 0 {
			break
		}
		games[m.chosen].run()
	}
}
