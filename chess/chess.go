package chess

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/notnil/chess"
)

type boardScheme struct {
	name      string
	light     string
	dark      string
	moveLight string
	moveDark  string
}

var schemes = []boardScheme{
	{"Classic", "#F0D9B5", "#B58863", "#CEB97A", "#A07840"},
	{"Ocean",   "#C8DDF0", "#4A79A5", "#9BBFE0", "#2E5F85"},
	{"Mint",    "#C8EBC8", "#4A804A", "#96CC96", "#2E602E"},
	{"Dusk",    "#E0C8F0", "#7B5EA7", "#C8A0E0", "#5B3E87"},
}

var (
	cursorSq   = lipgloss.NewStyle().Background(lipgloss.Color("#5BA3FF")).Foreground(lipgloss.Color("#ffffff"))
	selectedSq = lipgloss.NewStyle().Background(lipgloss.Color("#7FBF3F")).Foreground(lipgloss.Color("#ffffff"))
	validLight = lipgloss.NewStyle().Background(lipgloss.Color("#D8C87A")).Foreground(lipgloss.Color("#555555"))
	validDark  = lipgloss.NewStyle().Background(lipgloss.Color("#9E7A46")).Foreground(lipgloss.Color("#222222"))
	hintLight  = lipgloss.NewStyle().Background(lipgloss.Color("#B57BFF")).Foreground(lipgloss.Color("#ffffff"))
	hintDark   = lipgloss.NewStyle().Background(lipgloss.Color("#8A50CC")).Foreground(lipgloss.Color("#ffffff"))
	checkSq    = lipgloss.NewStyle().Background(lipgloss.Color("#CC3333")).Foreground(lipgloss.Color("#ffffff"))
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	msgStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
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
type hintMsg struct{ move *chess.Move }

func computeHint(g *chess.Game, depth int) tea.Cmd {
	snapshot := g.Clone()
	return func() tea.Msg {
		return hintMsg{bestMoveAtDepth(snapshot, depth)}
	}
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func formatClock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

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
	promoting     bool
	promotionFrom chess.Square
	promotionTo   chess.Square
	flipped       bool
	hintFrom      *chess.Square
	hintTo        *chess.Square
	hinting           bool
	resigned          bool
	timeSelect        bool
	schemeSelect      bool
	schemeIdx         int
	pendingVsComputer bool
	whiteTime         time.Duration
	blackTime         time.Duration
	clockOn           bool
}

func (m model) boardSquare(row, col int) chess.Square {
	if m.flipped {
		return chess.Square(row*8 + (7 - col))
	}
	return chess.Square((7-row)*8 + col)
}

func (m model) boardIsLight(row, col int) bool {
	if m.flipped {
		return (row+(7-col))%2 == 1
	}
	return ((7-row)+col)%2 == 1
}

func newModel() model {
	return model{
		game:       chess.NewGame(),
		cursor:     [2]int{7, 4},
		validDests: make(map[chess.Square]bool),
		message:    "White's turn",
		whiteTime:  10 * time.Minute,
		blackTime:  10 * time.Minute,
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

func isPromotionMove(g *chess.Game, from, to chess.Square) bool {
	for _, mv := range g.ValidMoves() {
		if mv.S1() == from && mv.S2() == to && mv.Promo() != chess.NoPieceType {
			return true
		}
	}
	return false
}

func (m model) Init() tea.Cmd { return tick() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if cmsg, ok := msg.(computerMoveMsg); ok {
		if cmsg.move != nil {
			m.executeMove(cmsg.move.S1(), cmsg.move.S2())
		}
		m.thinking = false
		return m, nil
	}

	if hmsg, ok := msg.(hintMsg); ok {
		m.hinting = false
		if hmsg.move != nil {
			f := hmsg.move.S1()
			m.hintFrom = &f
			t := hmsg.move.S2()
			m.hintTo = &t
		}
		return m, nil
	}

	if _, ok := msg.(tickMsg); ok {
		if m.clockOn && m.game.Outcome() == chess.NoOutcome && !m.thinking && !m.modeSelect && !m.colorSelect {
			if m.game.Position().Turn() == chess.White {
				m.whiteTime -= time.Second
				if m.whiteTime <= 0 {
					m.whiteTime = 0
					m.message = "White loses on time!"
					return m, nil
				}
			} else {
				m.blackTime -= time.Second
				if m.blackTime <= 0 {
					m.blackTime = 0
					m.message = "Black loses on time!"
					return m, nil
				}
			}
		}
		return m, tick()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.modeSelect {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "1":
				m.modeSelect = false
				m.timeSelect = true
			case "2":
				m.modeSelect = false
				m.timeSelect = true
				m.pendingVsComputer = true
			}
			return m, nil
		}

		if m.timeSelect {
			times := map[string]time.Duration{
				"1": 1 * time.Minute,
				"2": 5 * time.Minute,
				"3": 10 * time.Minute,
				"4": 30 * time.Minute,
			}
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "1", "2", "3", "4":
				d := times[msg.String()]
				m.whiteTime = d
				m.blackTime = d
				m.timeSelect = false
				m.schemeSelect = true
			}
			return m, nil
		}

		if m.schemeSelect {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "1", "2", "3", "4":
				idx := int(msg.String()[0] - '1')
				if idx < len(schemes) {
					m.schemeIdx = idx
				}
				m.schemeSelect = false
				if m.pendingVsComputer {
					m.pendingVsComputer = false
					m.diffSelect = true
				} else {
					m.message = "White's turn"
				}
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
				m.flipped = true
				m.thinking = true
				m.message = "Computer is thinking..."
				return m, computeMove(m.game, depthForDifficulty(m.difficulty))
			}
			return m, nil
		}

		if m.promoting {
			switch msg.String() {
			case "q", "Q":
				m.promoting = false
				m.executePromotion(chess.Queen)
			case "r", "R":
				m.promoting = false
				m.executePromotion(chess.Rook)
			case "b", "B":
				m.promoting = false
				m.executePromotion(chess.Bishop)
			case "n", "N":
				m.promoting = false
				m.executePromotion(chess.Knight)
			case "ctrl+c":
				return m, tea.Quit
			}
			if !m.promoting && m.vsComputer && m.game.Position().Turn() == m.computerColor && m.game.Outcome() == chess.NoOutcome {
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
		case "f":
			m.flipped = !m.flipped
		case "?":
			m.hintFrom = nil
			m.hintTo = nil
			m.hinting = true
			return m, computeHint(m.game, 2)
		case "r":
			if !m.resigned && m.game.Outcome() == chess.NoOutcome {
				m.resigned = true
				if m.game.Position().Turn() == chess.White {
					m.message = "White resigns — Black wins!  (q to quit)"
				} else {
					m.message = "Black resigns — White wins!  (q to quit)"
				}
			}
		case "t":
			m.takeback()
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
	if m.game.Outcome() != chess.NoOutcome || m.resigned {
		return m, nil
	}

	sq := m.boardSquare(m.cursor[0], m.cursor[1])
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
	if piece != chess.NoPiece && piece.Color() == turn && sq != m.boardSquare((*m.selected)[0], (*m.selected)[1]) {
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

	fromSq := m.boardSquare((*m.selected)[0], (*m.selected)[1])

	if isPromotionMove(m.game, fromSq, sq) {
		m.promoting = true
		m.promotionFrom = fromSq
		m.promotionTo = sq
		m.selected = nil
		m.validDests = make(map[chess.Square]bool)
		m.message = "Promote pawn: Q R B N"
		return m, nil
	}

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
	m.clockOn = true

	f := chosen.S1()
	m.lastFrom = &f
	t := chosen.S2()
	m.lastTo = &t
	m.hintFrom = nil
	m.hintTo = nil

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

func (m *model) executePromotion(piece chess.PieceType) {
	for _, mv := range m.game.ValidMoves() {
		if mv.S1() == m.promotionFrom && mv.S2() == m.promotionTo && mv.Promo() == piece {
			m.game.Move(mv)
			switch m.game.Outcome() {
			case chess.WhiteWon:
				m.message = "Checkmate! White wins!  (q to quit)"
			case chess.BlackWon:
				m.message = "Checkmate! Black wins!  (q to quit)"
			case chess.Draw:
				m.message = "Draw!  (q to quit)"
			default:
				if mv.HasTag(chess.Check) {
					m.message = "Check!  " + turnMsg(m.game)
				} else {
					m.message = turnMsg(m.game)
				}
			}
			return
		}
	}
}

func inCheckSquare(g *chess.Game) (chess.Square, bool) {
	moves := g.Moves()
	if len(moves) == 0 {
		return chess.A1, false
	}
	if !moves[len(moves)-1].HasTag(chess.Check) {
		return chess.A1, false
	}
	pos := g.Position()
	turn := pos.Turn()
	board := pos.Board()
	for sq := chess.A1; sq <= chess.H8; sq++ {
		p := board.Piece(sq)
		if p.Type() == chess.King && p.Color() == turn {
			return sq, true
		}
	}
	return chess.A1, false
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

	if m.schemeSelect {
		var sb strings.Builder
		sb.WriteString("\n ")
		sb.WriteString(titleStyle.Render("Chess"))
		sb.WriteString("\n\n")
		sb.WriteString("  Choose a board color scheme\n\n")
		for i, s := range schemes {
			sb.WriteString(fmt.Sprintf("  [%d]  %s\n", i+1, s.name))
		}
		sb.WriteString("\n  " + msgStyle.Render("Press 1, 2, 3, or 4") + "\n\n")
		return sb.String()
	}

	if m.timeSelect {
		var sb strings.Builder
		sb.WriteString("\n ")
		sb.WriteString(titleStyle.Render("Chess"))
		sb.WriteString("\n\n")
		sb.WriteString("  Select time control\n\n")
		sb.WriteString("  [1]  Bullet    (1 min)\n")
		sb.WriteString("  [2]  Blitz     (5 min)\n")
		sb.WriteString("  [3]  Rapid     (10 min)\n")
		sb.WriteString("  [4]  Classical (30 min)\n\n")
		sb.WriteString("  " + msgStyle.Render("Press 1, 2, 3, or 4") + "\n\n")
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
	fileLabels := "    a  b  c  d  e  f  g  h\n"
	if m.flipped {
		fileLabels = "    h  g  f  e  d  c  b  a\n"
	}
	sb.WriteString(fileLabels)

	sc := schemes[m.schemeIdx]
	sqLight     := lipgloss.NewStyle().Background(lipgloss.Color(sc.light)).Foreground(lipgloss.Color("#1a1a1a"))
	sqDark      := lipgloss.NewStyle().Background(lipgloss.Color(sc.dark)).Foreground(lipgloss.Color("#1a1a1a"))
	sqMoveLight := lipgloss.NewStyle().Background(lipgloss.Color(sc.moveLight)).Foreground(lipgloss.Color("#1a1a1a"))
	sqMoveDark  := lipgloss.NewStyle().Background(lipgloss.Color(sc.moveDark)).Foreground(lipgloss.Color("#ffffff"))

	board := m.game.Position().Board()

	for row := 0; row < 8; row++ {
		var rank int
		if m.flipped {
			rank = row
		} else {
			rank = 7 - row
		}
		sb.WriteString(fmt.Sprintf("  %d ", rank+1))

		for col := 0; col < 8; col++ {
			sq := m.boardSquare(row, col)
			piece := board.Piece(sq)
			light := m.boardIsLight(row, col)

			isCursor := m.cursor[0] == row && m.cursor[1] == col
			isSelected := m.selected != nil && (*m.selected)[0] == row && (*m.selected)[1] == col
			isValidDest := m.validDests[sq]
			isLastMove := (m.lastFrom != nil && *m.lastFrom == sq) || (m.lastTo != nil && *m.lastTo == sq)
			isHint := (m.hintFrom != nil && *m.hintFrom == sq) || (m.hintTo != nil && *m.hintTo == sq)
			checkKingSq, inCheck := inCheckSquare(m.game)
			isCheckKing := inCheck && sq == checkKingSq

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
			case isCheckKing:
				style = checkSq
			case isValidDest && light:
				style = validLight
			case isValidDest && !light:
				style = validDark
			case isHint && light:
				style = hintLight
			case isHint && !light:
				style = hintDark
			case isLastMove && light:
				style = sqMoveLight
			case isLastMove && !light:
				style = sqMoveDark
			case light:
				style = sqLight
			default:
				style = sqDark
			}

			// Always render pieces in their own colour so white pieces aren't
			// made dark by a light-square foreground, and black pieces aren't
			// turned white by a last-move highlight's foreground.
			if piece != chess.NoPiece {
				if piece.Color() == chess.White {
					style = style.Foreground(lipgloss.Color("#FFFFFF"))
				} else {
					style = style.Foreground(lipgloss.Color("#1a1a1a"))
				}
			}

			sb.WriteString(style.Render(cell))
		}

		sb.WriteString(fmt.Sprintf(" %d\n", rank+1))
	}

	sb.WriteString(fileLabels + "\n")
	// message + material score on same line
	matStr := ""
	if ms := materialScore(m.game); ms > 0 {
		matStr = fmt.Sprintf("  +%d", ms)
	} else if ms < 0 {
		matStr = fmt.Sprintf("  %d", ms)
	}
	sb.WriteString(" " + msgStyle.Render(m.message+matStr) + "\n")
	// opening name
	if op := openingName(m.game); op != "" {
		sb.WriteString(" " + msgStyle.Render("Opening: "+op) + "\n")
	}
	sb.WriteString("\n")
	if m.hinting {
		sb.WriteString(" " + msgStyle.Render("Finding hint...") + "\n\n")
	}
	sb.WriteString(fmt.Sprintf("  White: %s   Black: %s\n\n",
		formatClock(m.whiteTime), formatClock(m.blackTime)))

	// Captured pieces
	byWhite, byBlack := capturedPieces(m.game)
	if len(byWhite)+len(byBlack) > 0 {
		sb.WriteString("  Captured by White:")
		for _, p := range byWhite {
			sb.WriteString(" " + pieceGlyph(p))
		}
		sb.WriteString("\n  Captured by Black:")
		for _, p := range byBlack {
			sb.WriteString(" " + pieceGlyph(p))
		}
		sb.WriteString("\n\n")
	}

	if m.promoting {
		sb.WriteString("  [Q]ueen  [R]ook  [B]ishop  [N]knight\n\n")
	}

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

	// PGN on game over
	if m.game.Outcome() != chess.NoOutcome || m.resigned {
		pgn := m.game.String()
		for _, line := range strings.Split(strings.TrimSpace(pgn), "\n") {
			sb.WriteString(" " + msgStyle.Render(line) + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(" ↑↓←→ / hjkl  move cursor\n")
	sb.WriteString(" Enter / Space  select / move\n")
	sb.WriteString(" Esc  cancel selection   q  quit\n")
	sb.WriteString(" f  flip board   ?  hint   r  resign   t  takeback\n\n")

	return sb.String()
}

func Run() {
	m := newModel()
	m.modeSelect = true
	m.message = ""
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
