package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"
)

var (
	lightSq       = lipgloss.NewStyle().Background(lipgloss.Color("#F0D9B5")).Foreground(lipgloss.Color("#1a1a1a"))
	darkSq        = lipgloss.NewStyle().Background(lipgloss.Color("#B58863")).Foreground(lipgloss.Color("#1a1a1a"))
	cursorSq      = lipgloss.NewStyle().Background(lipgloss.Color("#5BA3FF")).Foreground(lipgloss.Color("#ffffff"))
	selectedSq    = lipgloss.NewStyle().Background(lipgloss.Color("#7FBF3F")).Foreground(lipgloss.Color("#ffffff"))
	validLight    = lipgloss.NewStyle().Background(lipgloss.Color("#D8C87A")).Foreground(lipgloss.Color("#555555"))
	validDark     = lipgloss.NewStyle().Background(lipgloss.Color("#9E7A46")).Foreground(lipgloss.Color("#222222"))
	lastMoveLight = lipgloss.NewStyle().Background(lipgloss.Color("#CEB97A")).Foreground(lipgloss.Color("#1a1a1a"))
	lastMoveDark  = lipgloss.NewStyle().Background(lipgloss.Color("#A07840")).Foreground(lipgloss.Color("#ffffff"))
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	msgStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
)

var glyphs = map[chess.PieceType][2]string{
	chess.King:   {"♔", "♚"},
	chess.Queen:  {"♕", "♛"},
	chess.Rook:   {"♖", "♜"},
	chess.Bishop: {"♗", "♝"},
	chess.Knight: {"♘", "♞"},
	chess.Pawn:   {"♙", "♟"},
}

type computerMoveMsg struct{ move *chess.Move }

type model struct {
	game          *chess.Game
	cursor        [2]int
	selected      *[2]int
	validDests    map[chess.Square]bool
	message       string
	modeSelect    bool
	diffSelect    bool
	colorSelect   bool
	vsComputer    bool
	computerColor chess.Color
	thinking      bool
	lastFrom      *chess.Square
	lastTo        *chess.Square
	difficulty    int
}

func newModel() model {
	return model{
		game:       chess.NewGame(),
		cursor:     [2]int{7, 4},
		validDests: make(map[chess.Square]bool),
		message:    "White's turn",
	}
}

func computeMove(g *chess.Game, depth int) tea.Cmd {
	snapshot := g.Clone()
	return func() tea.Msg {
		return computerMoveMsg{bestMoveAtDepth(snapshot, depth)}
	}
}

func toSquare(row, col int) chess.Square {
	return chess.Square((7-row)*8 + col)
}

func isLight(row, col int) bool {
	return ((7-row)+col)%2 == 1
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmsg, ok := msg.(computerMoveMsg); ok {
		if cmsg.move != nil {
			m.executeMove(cmsg.move.S1(), cmsg.move.S2())
		}
		m.thinking = false
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.modeSelect {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "1":
				m.modeSelect = false
				m.message = "White's turn"
			case "2":
				m.modeSelect = false
				m.diffSelect = true
			}
			return m, nil
		}

		if m.diffSelect {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "1":
				m.diffSelect = false
				m.colorSelect = true
				m.difficulty = 1
			case "2":
				m.diffSelect = false
				m.colorSelect = true
				m.difficulty = 2
			case "3":
				m.diffSelect = false
				m.colorSelect = true
				m.difficulty = 3
			}
			return m, nil
		}

		if m.colorSelect {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "w", "W":
				m.colorSelect = false
				m.vsComputer = true
				m.computerColor = chess.Black
				m.message = "White's turn"
			case "b", "B":
				m.colorSelect = false
				m.vsComputer = true
				m.computerColor = chess.White
				m.thinking = true
				m.message = "Computer is thinking..."
				return m, computeMove(m.game, depthForDifficulty(m.difficulty))
			}
			return m, nil
		}

		if m.thinking {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor[0] > 0 {
				m.cursor[0]--
			}
		case "down", "j":
			if m.cursor[0] < 7 {
				m.cursor[0]++
			}
		case "left", "h":
			if m.cursor[1] > 0 {
				m.cursor[1]--
			}
		case "right", "l":
			if m.cursor[1] < 7 {
				m.cursor[1]++
			}
		case "enter", " ":
			return m.handleSelect()
		case "esc":
			m.selected = nil
			m.validDests = make(map[chess.Square]bool)
			m.message = turnMsg(m.game)
		}
	}
	return m, nil
}

func (m model) handleSelect() (tea.Model, tea.Cmd) {
	if m.game.Outcome() != chess.NoOutcome {
		return m, nil
	}

	sq := toSquare(m.cursor[0], m.cursor[1])
	piece := m.game.Position().Board().Piece(sq)
	turn := m.game.Position().Turn()

	if m.selected == nil {
		if piece == chess.NoPiece || piece.Color() != turn {
			m.message = "Select one of your pieces"
			return m, nil
		}
		sel := m.cursor
		m.selected = &sel
		m.validDests = validDestsFor(m.game, sq)
		m.message = fmt.Sprintf("%s selected — choose destination  (Esc to cancel)", sq)
		return m, nil
	}

	// Re-select own piece
	if piece != chess.NoPiece && piece.Color() == turn && sq != toSquare((*m.selected)[0], (*m.selected)[1]) {
		sel := m.cursor
		m.selected = &sel
		m.validDests = validDestsFor(m.game, sq)
		m.message = fmt.Sprintf("%s selected — choose destination  (Esc to cancel)", sq)
		return m, nil
	}

	if !m.validDests[sq] {
		m.selected = nil
		m.validDests = make(map[chess.Square]bool)
		m.message = "Invalid move — select one of your pieces"
		return m, nil
	}

	fromSq := toSquare((*m.selected)[0], (*m.selected)[1])
	m.executeMove(fromSq, sq)
	m.selected = nil
	m.validDests = make(map[chess.Square]bool)

	if m.vsComputer && m.game.Position().Turn() == m.computerColor && m.game.Outcome() == chess.NoOutcome {
		m.thinking = true
		m.message = "Computer is thinking..."
		return m, computeMove(m.game, depthForDifficulty(m.difficulty))
	}
	return m, nil
}

func validDestsFor(g *chess.Game, from chess.Square) map[chess.Square]bool {
	dests := make(map[chess.Square]bool)
	for _, mv := range g.ValidMoves() {
		if mv.S1() == from {
			dests[mv.S2()] = true
		}
	}
	return dests
}

func (m *model) executeMove(from, to chess.Square) {
	var chosen *chess.Move
	for _, mv := range m.game.ValidMoves() {
		if mv.S1() == from && mv.S2() == to {
			if chosen == nil || mv.Promo() == chess.Queen {
				chosen = mv
			}
		}
	}
	if chosen == nil {
		return
	}
	m.game.Move(chosen)

	f := chosen.S1()
	m.lastFrom = &f
	t := chosen.S2()
	m.lastTo = &t

	switch m.game.Outcome() {
	case chess.WhiteWon:
		m.message = "Checkmate! White wins!  (q to quit)"
	case chess.BlackWon:
		m.message = "Checkmate! Black wins!  (q to quit)"
	case chess.Draw:
		m.message = "Draw!  (q to quit)"
	default:
		if chosen.HasTag(chess.Check) {
			m.message = "Check!  " + turnMsg(m.game)
		} else {
			m.message = turnMsg(m.game)
		}
	}
}

func turnMsg(g *chess.Game) string {
	if g.Position().Turn() == chess.White {
		return "White's turn"
	}
	return "Black's turn"
}

func (m model) View() string {
	if m.modeSelect {
		var sb strings.Builder
		sb.WriteString("\n ")
		sb.WriteString(titleStyle.Render("Chess"))
		sb.WriteString("\n\n")
		sb.WriteString("  How do you want to play?\n\n")
		sb.WriteString("  [1]  Two player\n")
		sb.WriteString("  [2]  vs Computer\n\n")
		sb.WriteString("  " + msgStyle.Render("Press 1 or 2") + "\n\n")
		return sb.String()
	}

	if m.diffSelect {
		var sb strings.Builder
		sb.WriteString("\n ")
		sb.WriteString(titleStyle.Render("Chess"))
		sb.WriteString("\n\n")
		sb.WriteString("  Select difficulty\n\n")
		sb.WriteString("  [1]  Easy\n")
		sb.WriteString("  [2]  Medium\n")
		sb.WriteString("  [3]  Hard\n\n")
		sb.WriteString("  " + msgStyle.Render("Press 1, 2, or 3") + "\n\n")
		return sb.String()
	}

	if m.colorSelect {
		var sb strings.Builder
		sb.WriteString("\n ")
		sb.WriteString(titleStyle.Render("Chess"))
		sb.WriteString("\n\n")
		sb.WriteString("  You play as...\n\n")
		sb.WriteString("  [W]hite  or  [B]lack\n\n")
		sb.WriteString("  " + msgStyle.Render("Press W or B") + "\n\n")
		return sb.String()
	}

	var sb strings.Builder

	sb.WriteString("\n ")
	sb.WriteString(titleStyle.Render("Chess"))
	sb.WriteString("\n\n")
	sb.WriteString("    a  b  c  d  e  f  g  h\n")

	board := m.game.Position().Board()

	for row := 0; row < 8; row++ {
		rank := 7 - row
		sb.WriteString(fmt.Sprintf("  %d ", rank+1))

		for col := 0; col < 8; col++ {
			sq := chess.Square(rank*8 + col)
			piece := board.Piece(sq)
			light := isLight(row, col)

			isCursor := m.cursor[0] == row && m.cursor[1] == col
			isSelected := m.selected != nil && (*m.selected)[0] == row && (*m.selected)[1] == col
			isValidDest := m.validDests[sq]
			isLastMove := (m.lastFrom != nil && *m.lastFrom == sq) || (m.lastTo != nil && *m.lastTo == sq)

			var cell string
			if piece == chess.NoPiece {
				if isValidDest {
					cell = " · "
				} else {
					cell = "   "
				}
			} else {
				idx := 0
				if piece.Color() == chess.Black {
					idx = 1
				}
				cell = " " + glyphs[piece.Type()][idx] + " "
			}

			var style lipgloss.Style
			switch {
			case isCursor:
				style = cursorSq
			case isSelected:
				style = selectedSq
			case isValidDest && light:
				style = validLight
			case isValidDest && !light:
				style = validDark
			case isLastMove && light:
				style = lastMoveLight
			case isLastMove && !light:
				style = lastMoveDark
			case light:
				style = lightSq
			default:
				style = darkSq
			}

			sb.WriteString(style.Render(cell))
		}

		sb.WriteString(fmt.Sprintf(" %d\n", rank+1))
	}

	sb.WriteString("    a  b  c  d  e  f  g  h\n\n")
	sb.WriteString(" " + msgStyle.Render(m.message) + "\n\n")

	if lines := formatMoveHistory(m.game); len(lines) > 0 {
		start := 0
		if len(lines) > 8 {
			start = len(lines) - 8
		}
		for _, l := range lines[start:] {
			sb.WriteString(" " + msgStyle.Render(l) + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(" ↑↓←→ / hjkl  move cursor\n")
	sb.WriteString(" Enter / Space  select / move\n")
	sb.WriteString(" Esc  cancel selection   q  quit\n\n")

	return sb.String()
}

func main() {
	m := newModel()
	m.modeSelect = true
	m.message = ""
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
