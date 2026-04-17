package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/notnil/chess"
)

// helpers

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
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func playMoves(g *chess.Game, pairs [][2]chess.Square) {
	for _, pair := range pairs {
		for _, mv := range g.ValidMoves() {
			if mv.S1() == pair[0] && mv.S2() == pair[1] {
				g.Move(mv)
				break
			}
		}
	}
}

func foolsMate() *chess.Game {
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.F2, chess.F3},
		{chess.E7, chess.E5},
		{chess.G2, chess.G4},
		{chess.D8, chess.H4},
	})
	return g
}

// toSquare

func TestToSquare(t *testing.T) {
	tests := []struct {
		row, col int
		want     chess.Square
		name     string
	}{
		{7, 0, chess.A1, "a1"},
		{7, 7, chess.H1, "h1"},
		{0, 0, chess.A8, "a8"},
		{0, 7, chess.H8, "h8"},
		{6, 4, chess.E2, "e2"},
		{1, 4, chess.E7, "e7"},
		{7, 4, chess.E1, "e1"},
		{0, 3, chess.D8, "d8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toSquare(tt.row, tt.col); got != tt.want {
				t.Errorf("toSquare(%d, %d) = %v, want %v", tt.row, tt.col, got, tt.want)
			}
		})
	}
}

func TestToSquare_RoundTrip(t *testing.T) {
	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			sq := toSquare(row, col)
			gotRank := int(sq) / 8
			gotFile := int(sq) % 8
			wantRank := 7 - row
			if gotRank != wantRank || gotFile != col {
				t.Errorf("toSquare(%d, %d): rank=%d file=%d, want rank=%d file=%d",
					row, col, gotRank, gotFile, wantRank, col)
			}
		}
	}
}

// isLight

func TestIsLight(t *testing.T) {
	tests := []struct {
		row, col int
		want     bool
		name     string
	}{
		{7, 0, false, "a1 is dark"},
		{7, 1, true, "b1 is light"},
		{6, 0, true, "a2 is light"},
		{6, 1, false, "b2 is dark"},
		{0, 0, true, "a8 is light"},
		{0, 7, false, "h8 is dark"},
		{7, 7, true, "h1 is light"},
		{0, 4, true, "e8 is light"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isLight(tt.row, tt.col); got != tt.want {
				t.Errorf("isLight(%d, %d) = %v, want %v", tt.row, tt.col, got, tt.want)
			}
		})
	}
}

func TestIsLight_AdjacentSquaresAlternate(t *testing.T) {
	for row := 0; row < 8; row++ {
		for col := 0; col < 7; col++ {
			if isLight(row, col) == isLight(row, col+1) {
				t.Errorf("adjacent squares (%d,%d) and (%d,%d) should alternate color",
					row, col, row, col+1)
			}
		}
	}
}

// newModel

func TestNewModel(t *testing.T) {
	m := newModel()

	if m.cursor != [2]int{7, 4} {
		t.Errorf("cursor = %v, want [7 4]", m.cursor)
	}
	if m.selected != nil {
		t.Error("expected no selection on new game")
	}
	if len(m.validDests) != 0 {
		t.Error("expected empty validDests on new game")
	}
	if m.message != "White's turn" {
		t.Errorf("message = %q, want %q", m.message, "White's turn")
	}
	if m.game == nil {
		t.Fatal("game should not be nil")
	}
	if m.game.Position().Turn() != chess.White {
		t.Error("expected white to move first")
	}
}

// turnMsg

func TestTurnMsg(t *testing.T) {
	g := chess.NewGame()
	if got := turnMsg(g); got != "White's turn" {
		t.Errorf("got %q, want %q", got, "White's turn")
	}

	g.Move(g.ValidMoves()[0])
	if got := turnMsg(g); got != "Black's turn" {
		t.Errorf("got %q, want %q", got, "Black's turn")
	}
}

// validDestsFor

func TestValidDestsFor_Pawn(t *testing.T) {
	g := chess.NewGame()
	dests := validDestsFor(g, chess.E2)
	if !dests[chess.E3] {
		t.Error("e2 pawn should be able to reach e3")
	}
	if !dests[chess.E4] {
		t.Error("e2 pawn should be able to reach e4")
	}
	if len(dests) != 2 {
		t.Errorf("e2 pawn: got %d destinations, want 2", len(dests))
	}
}

func TestValidDestsFor_Knight(t *testing.T) {
	g := chess.NewGame()
	dests := validDestsFor(g, chess.B1)
	if !dests[chess.A3] {
		t.Error("b1 knight should reach a3")
	}
	if !dests[chess.C3] {
		t.Error("b1 knight should reach c3")
	}
	if len(dests) != 2 {
		t.Errorf("b1 knight: got %d destinations, want 2", len(dests))
	}
}

func TestValidDestsFor_BlockedKing(t *testing.T) {
	g := chess.NewGame()
	dests := validDestsFor(g, chess.E1)
	if len(dests) != 0 {
		t.Errorf("e1 king should have 0 moves at start, got %d", len(dests))
	}
}

func TestValidDestsFor_EmptySquare(t *testing.T) {
	g := chess.NewGame()
	dests := validDestsFor(g, chess.E4)
	if len(dests) != 0 {
		t.Errorf("empty square should have 0 destinations, got %d", len(dests))
	}
}

// executeMove

func TestExecuteMove_NormalMove(t *testing.T) {
	m := newModel()
	m.executeMove(chess.E2, chess.E4)

	if m.message != "Black's turn" {
		t.Errorf("message = %q, want %q", m.message, "Black's turn")
	}
	if m.game.Position().Board().Piece(chess.E4) == chess.NoPiece {
		t.Error("expected pawn on e4")
	}
	if m.game.Position().Board().Piece(chess.E2) != chess.NoPiece {
		t.Error("expected e2 to be empty after move")
	}
}

func TestExecuteMove_InvalidMove(t *testing.T) {
	m := newModel()
	m.executeMove(chess.E2, chess.E5)

	if m.game.Position().Turn() != chess.White {
		t.Error("game state should not change after invalid move attempt")
	}
}

func TestExecuteMove_NoOpForNoMove(t *testing.T) {
	m := newModel()
	before := m.game.Position().Hash()
	m.executeMove(chess.E1, chess.E1)
	after := m.game.Position().Hash()
	if before != after {
		t.Error("self-move should not change game state")
	}
}

func TestExecuteMove_Check(t *testing.T) {
	// After 1.e4 e5 2.Bc4 — Bxf7+ is check (not mate)
	fen, err := chess.FEN("r1bqkbnr/pppp1ppp/2n5/4p3/2B1P3/8/PPPP1PPP/RNBQK1NR w KQkq - 2 3")
	if err != nil {
		t.Fatal(err)
	}
	g := chess.NewGame(fen)
	m := model{game: g, validDests: make(map[chess.Square]bool)}
	m.executeMove(chess.C4, chess.F7)

	if !strings.Contains(m.message, "Check!") {
		t.Errorf("expected check message, got %q", m.message)
	}
	if !strings.Contains(m.message, "Black's turn") {
		t.Errorf("expected black's turn in message, got %q", m.message)
	}
}

func TestExecuteMove_Checkmate(t *testing.T) {
	g := foolsMate()
	if g.Outcome() != chess.BlackWon {
		t.Fatalf("fool's mate should result in BlackWon, got %v", g.Outcome())
	}
}

func TestExecuteMove_CheckmateMessage(t *testing.T) {
	// Set up fool's mate: after 1.f3 e5 2.g4, black plays Qh4#
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.F2, chess.F3},
		{chess.E7, chess.E5},
		{chess.G2, chess.G4},
	})
	m := model{game: g, validDests: make(map[chess.Square]bool)}
	m.executeMove(chess.D8, chess.H4)

	if !strings.Contains(m.message, "Black wins") {
		t.Errorf("expected black wins message, got %q", m.message)
	}
}

func TestExecuteMove_WhiteWinsMessage(t *testing.T) {
	// Scholar's mate: 1.e4 e5 2.Bc4 Nc6 3.Qh5 Nf6 4.Qxf7#
	g := chess.NewGame()
	playMoves(g, [][2]chess.Square{
		{chess.E2, chess.E4},
		{chess.E7, chess.E5},
		{chess.F1, chess.C4},
		{chess.B8, chess.C6},
		{chess.D1, chess.H5},
		{chess.G8, chess.F6},
	})
	m := model{game: g, validDests: make(map[chess.Square]bool)}
	m.executeMove(chess.H5, chess.F7)

	if !strings.Contains(m.message, "White wins") {
		t.Errorf("expected white wins message, got %q", m.message)
	}
}

func TestExecuteMove_QueenPromotion(t *testing.T) {
	// Position with a pawn ready to promote on e7
	fen, err := chess.FEN("8/4P3/8/8/8/8/8/4K2k w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	g := chess.NewGame(fen)
	m := model{game: g, validDests: make(map[chess.Square]bool)}
	m.executeMove(chess.E7, chess.E8)

	promoted := m.game.Position().Board().Piece(chess.E8)
	if promoted == chess.NoPiece {
		t.Fatal("expected promoted piece on e8")
	}
	if promoted.Type() != chess.Queen {
		t.Errorf("pawn should promote to queen, got %v", promoted.Type())
	}
}

// Update — cursor movement

func TestUpdate_CursorMovement(t *testing.T) {
	tests := []struct {
		name       string
		start      [2]int
		key        string
		wantCursor [2]int
	}{
		{"up", [2]int{4, 4}, "up", [2]int{3, 4}},
		{"k alias for up", [2]int{4, 4}, "k", [2]int{3, 4}},
		{"down", [2]int{4, 4}, "down", [2]int{5, 4}},
		{"j alias for down", [2]int{4, 4}, "j", [2]int{5, 4}},
		{"left", [2]int{4, 4}, "left", [2]int{4, 3}},
		{"h alias for left", [2]int{4, 4}, "h", [2]int{4, 3}},
		{"right", [2]int{4, 4}, "right", [2]int{4, 5}},
		{"l alias for right", [2]int{4, 4}, "l", [2]int{4, 5}},
		{"up clamped at top", [2]int{0, 4}, "up", [2]int{0, 4}},
		{"down clamped at bottom", [2]int{7, 4}, "down", [2]int{7, 4}},
		{"left clamped at left edge", [2]int{4, 0}, "left", [2]int{4, 0}},
		{"right clamped at right edge", [2]int{4, 7}, "right", [2]int{4, 7}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newModel()
			m.cursor = tt.start
			updated, _ := m.Update(key(tt.key))
			if got := updated.(model).cursor; got != tt.wantCursor {
				t.Errorf("cursor = %v, want %v", got, tt.wantCursor)
			}
		})
	}
}

func TestUpdate_QuitKeys(t *testing.T) {
	for _, k := range []string{"q", "ctrl+c"} {
		t.Run(k, func(t *testing.T) {
			m := newModel()
			_, cmd := m.Update(key(k))
			if cmd == nil {
				t.Fatalf("key %q should return a quit command", k)
			}
			if _, ok := cmd().(tea.QuitMsg); !ok {
				t.Errorf("key %q: expected QuitMsg, got %T", k, cmd())
			}
		})
	}
}

func TestUpdate_EscClearsSelection(t *testing.T) {
	m := newModel()
	sel := [2]int{6, 4}
	m.selected = &sel
	m.validDests = map[chess.Square]bool{chess.E3: true, chess.E4: true}

	updated, cmd := m.Update(key("esc"))
	got := updated.(model)

	if cmd != nil {
		t.Error("esc should not return a command")
	}
	if got.selected != nil {
		t.Error("esc should clear selection")
	}
	if len(got.validDests) != 0 {
		t.Error("esc should clear validDests")
	}
	if got.message != "White's turn" {
		t.Errorf("message = %q, want %q", got.message, "White's turn")
	}
}

func TestUpdate_UnknownKeyNoOp(t *testing.T) {
	m := newModel()
	m.cursor = [2]int{3, 3}
	updated, cmd := m.Update(key("x"))
	got := updated.(model)

	if cmd != nil {
		t.Error("unknown key should not return a command")
	}
	if got.cursor != [2]int{3, 3} {
		t.Error("unknown key should not move cursor")
	}
}

// handleSelect

func TestHandleSelect_EmptySquare(t *testing.T) {
	m := newModel()
	m.cursor = [2]int{4, 4} // e4 — empty at start
	updated, _ := m.handleSelect()
	got := updated.(model)

	if got.message != "Select one of your pieces" {
		t.Errorf("message = %q", got.message)
	}
	if got.selected != nil {
		t.Error("should not select empty square")
	}
}

func TestHandleSelect_OpponentPiece(t *testing.T) {
	m := newModel()
	m.cursor = [2]int{1, 4} // e7 — black pawn, white's turn
	updated, _ := m.handleSelect()
	got := updated.(model)

	if got.message != "Select one of your pieces" {
		t.Errorf("message = %q", got.message)
	}
	if got.selected != nil {
		t.Error("should not select opponent's piece")
	}
}

func TestHandleSelect_OwnPiece(t *testing.T) {
	m := newModel()
	m.cursor = [2]int{6, 4} // e2 — white pawn
	updated, _ := m.handleSelect()
	got := updated.(model)

	if got.selected == nil {
		t.Fatal("expected piece to be selected")
	}
	if *got.selected != [2]int{6, 4} {
		t.Errorf("selected = %v, want [6 4]", *got.selected)
	}
	if !got.validDests[chess.E3] || !got.validDests[chess.E4] {
		t.Error("expected e3 and e4 as valid destinations for e2 pawn")
	}
	if !strings.Contains(got.message, "e2") {
		t.Errorf("message should reference the selected square, got %q", got.message)
	}
}

func TestHandleSelect_ValidDestination(t *testing.T) {
	m := newModel()
	// Select e2 pawn
	m.cursor = [2]int{6, 4}
	updated, _ := m.handleSelect()
	m = updated.(model)
	// Move to e4
	m.cursor = [2]int{4, 4}
	updated, _ = m.handleSelect()
	got := updated.(model)

	if got.selected != nil {
		t.Error("selection should be cleared after move")
	}
	if len(got.validDests) != 0 {
		t.Error("validDests should be cleared after move")
	}
	if got.game.Position().Turn() != chess.Black {
		t.Error("expected black's turn after white's move")
	}
	if got.game.Position().Board().Piece(chess.E4) == chess.NoPiece {
		t.Error("expected pawn on e4")
	}
}

func TestHandleSelect_InvalidDestination(t *testing.T) {
	m := newModel()
	// Select e2 pawn
	m.cursor = [2]int{6, 4}
	updated, _ := m.handleSelect()
	m = updated.(model)
	// Try to move to e5 (invalid for one-step pawn)
	m.cursor = [2]int{3, 4}
	updated, _ = m.handleSelect()
	got := updated.(model)

	if got.selected != nil {
		t.Error("selection should be cleared after invalid destination")
	}
	if !strings.Contains(got.message, "Invalid move") {
		t.Errorf("message = %q", got.message)
	}
	if got.game.Position().Turn() != chess.White {
		t.Error("game state should not change")
	}
}

func TestHandleSelect_ReselectionOfOwnPiece(t *testing.T) {
	m := newModel()
	// Select e2 pawn
	m.cursor = [2]int{6, 4}
	updated, _ := m.handleSelect()
	m = updated.(model)
	// Re-select d2 pawn without moving
	m.cursor = [2]int{6, 3}
	updated, _ = m.handleSelect()
	got := updated.(model)

	if got.selected == nil || *got.selected != [2]int{6, 3} {
		t.Errorf("expected d2 selected, got %v", got.selected)
	}
	if !got.validDests[chess.D3] || !got.validDests[chess.D4] {
		t.Error("expected d3/d4 as destinations for d2 pawn")
	}
}

func TestHandleSelect_GameOverNoOp(t *testing.T) {
	m := model{
		game:       foolsMate(),
		cursor:     [2]int{6, 4},
		validDests: make(map[chess.Square]bool),
		message:    "Checkmate! Black wins!  (q to quit)",
	}
	updated, _ := m.handleSelect()
	got := updated.(model)

	if got.selected != nil {
		t.Error("should not select anything when game is over")
	}
	if got.message != "Checkmate! Black wins!  (q to quit)" {
		t.Errorf("message should not change, got %q", got.message)
	}
}

func TestHandleSelect_EnterThenEscThenMove(t *testing.T) {
	m := newModel()
	// Select e2
	m.cursor = [2]int{6, 4}
	updated, _ := m.handleSelect()
	m = updated.(model)
	// Esc cancels
	updated, _ = m.Update(key("esc"))
	m = updated.(model)
	if m.selected != nil {
		t.Error("expected selection cleared by esc")
	}
	// Now select d2
	m.cursor = [2]int{6, 3}
	updated, _ = m.handleSelect()
	got := updated.(model)
	if got.selected == nil || *got.selected != [2]int{6, 3} {
		t.Error("expected d2 selected after esc and re-selection")
	}
}

// View

func TestView_ContainsFileLabels(t *testing.T) {
	view := newModel().View()
	for _, f := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		if !strings.Contains(view, f) {
			t.Errorf("view missing file label %q", f)
		}
	}
}

func TestView_ContainsRankLabels(t *testing.T) {
	view := newModel().View()
	for _, r := range []string{"1", "2", "3", "4", "5", "6", "7", "8"} {
		if !strings.Contains(view, r) {
			t.Errorf("view missing rank label %q", r)
		}
	}
}

func TestView_ContainsTitle(t *testing.T) {
	if !strings.Contains(newModel().View(), "Chess") {
		t.Error("view missing 'Chess' title")
	}
}

func TestView_ContainsMessage(t *testing.T) {
	if !strings.Contains(newModel().View(), "White's turn") {
		t.Error("view missing turn message")
	}
}

func TestView_ContainsAllPieceGlyphs(t *testing.T) {
	view := newModel().View()
	for _, g := range []string{"♔", "♕", "♖", "♗", "♘", "♙", "♚", "♛", "♜", "♝", "♞", "♟"} {
		if !strings.Contains(view, g) {
			t.Errorf("view missing piece glyph %q", g)
		}
	}
}

func TestView_ValidDestsShowDot(t *testing.T) {
	m := newModel()
	m.cursor = [2]int{6, 4}
	updated, _ := m.handleSelect()
	if !strings.Contains(updated.(model).View(), "·") {
		t.Error("view should show · for valid move destinations")
	}
}

func TestView_ContainsKeyboardHints(t *testing.T) {
	view := newModel().View()
	for _, hint := range []string{"hjkl", "Enter", "Esc", "quit"} {
		if !strings.Contains(view, hint) {
			t.Errorf("view missing keyboard hint %q", hint)
		}
	}
}

func TestView_UpdatedMessageAfterMove(t *testing.T) {
	m := newModel()
	m.cursor = [2]int{6, 4}
	updated, _ := m.handleSelect()
	m = updated.(model)
	m.cursor = [2]int{4, 4}
	updated, _ = m.handleSelect()

	if !strings.Contains(updated.(model).View(), "Black's turn") {
		t.Error("view should show black's turn after white's move")
	}
}
