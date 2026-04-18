package chess

import "github.com/notnil/chess"

// takeback undoes the last move (or the last two moves in vs-computer mode,
// so the human player is always left with a position to play).
func (m *model) takeback() {
	moves := m.game.Moves()
	if len(moves) == 0 {
		return
	}

	// In vs-computer mode, undo both the computer's reply and the player's move
	// so the human is returned to their own turn.
	undoCount := 1
	if m.vsComputer && len(moves) >= 2 {
		undoCount = 2
	}

	// Rebuild the game from scratch up to len(moves)-undoCount
	target := len(moves) - undoCount
	positions := m.game.Positions()
	newGame := chess.NewGame(chess.UseNotation(chess.AlgebraicNotation{}))
	allMoves := m.game.Moves()
	for i := 0; i < target; i++ {
		pos := positions[i]
		_ = pos
		newGame.Move(allMoves[i])
	}
	m.game = newGame

	// Clear the last-move highlight.
	m.lastFrom = nil
	m.lastTo = nil

	m.selected = nil
	m.validDests = make(map[chess.Square]bool)
	m.message = turnMsg(m.game)
}
