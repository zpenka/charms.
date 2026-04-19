package game2048

import "math/rand"

const BoardSize = 4

type gameState int

const (
	StatePlaying   gameState = iota
	StateWon
	StateGameOver
	StateLeaderboard
)

type direction int

const (
	DirLeft  direction = iota
	DirRight
	DirUp
	DirDown
)

type board [BoardSize][BoardSize]int

type model struct {
	board     board
	score     int
	state     gameState
	continued bool
	hasPrev   bool
	prevBoard board
	prevScore int
	scores    []ScoreEntry
	scorePath string
}

func newGame() model {
	path := defaultScorePath()
	return newGameWithScores(loadScores(path), path)
}

func newGameWithScores(scores []ScoreEntry, path string) model {
	var b board
	b = addTile(b)
	b = addTile(b)
	return model{
		board:     b,
		state:     StatePlaying,
		scores:    scores,
		scorePath: path,
	}
}

// slideRow compacts and merges a single row leftward.
// Returns the new row and the score gained from merges.
func slideRow(row [BoardSize]int) ([BoardSize]int, int) {
	// compact: collect non-zero values
	var vals []int
	for _, v := range row {
		if v != 0 {
			vals = append(vals, v)
		}
	}
	// merge adjacent equal pairs (each pair merges once)
	score := 0
	var merged []int
	for i := 0; i < len(vals); i++ {
		if i+1 < len(vals) && vals[i] == vals[i+1] {
			merged = append(merged, vals[i]*2)
			score += vals[i] * 2
			i++ // skip the consumed tile
		} else {
			merged = append(merged, vals[i])
		}
	}
	// pad right with zeros
	var result [BoardSize]int
	for i, v := range merged {
		result[i] = v
	}
	return result, score
}

func transpose(b board) board {
	var r board
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			r[i][j] = b[j][i]
		}
	}
	return r
}

func reverseRows(b board) board {
	var r board
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			r[i][j] = b[i][BoardSize-1-j]
		}
	}
	return r
}

// slide moves all tiles in direction d and returns the new board, score, and
// whether any tile actually moved.
func slide(b board, d direction) (board, int, bool) {
	// Rotate so that the target direction becomes "left", slide, rotate back.
	var work board
	switch d {
	case DirRight:
		work = reverseRows(b)
	case DirUp:
		work = transpose(b)
	case DirDown:
		work = reverseRows(transpose(b))
	default:
		work = b
	}

	var result board
	total := 0
	for i := 0; i < BoardSize; i++ {
		row, score := slideRow(work[i])
		result[i] = row
		total += score
	}

	// Rotate back
	switch d {
	case DirRight:
		result = reverseRows(result)
	case DirUp:
		result = transpose(result)
	case DirDown:
		result = transpose(reverseRows(result))
	}

	return result, total, result != b
}

func addTile(b board) board {
	var empty [][2]int
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b[r][c] == 0 {
				empty = append(empty, [2]int{r, c})
			}
		}
	}
	if len(empty) == 0 {
		return b
	}
	cell := empty[rand.Intn(len(empty))]
	value := 2
	if rand.Intn(10) == 0 {
		value = 4
	}
	b[cell[0]][cell[1]] = value
	return b
}

func hasWon(b board) bool {
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b[r][c] >= 2048 {
				return true
			}
		}
	}
	return false
}

func canMove(b board) bool {
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b[r][c] == 0 {
				return true
			}
			v := b[r][c]
			if c+1 < BoardSize && b[r][c+1] == v {
				return true
			}
			if r+1 < BoardSize && b[r+1][c] == v {
				return true
			}
		}
	}
	return false
}

func undoMove(m model) model {
	if !m.hasPrev {
		return m
	}
	m.board = m.prevBoard
	m.score = m.prevScore
	m.hasPrev = false
	return m
}

func maxTile(b board) int {
	best := 0
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b[r][c] > best {
				best = b[r][c]
			}
		}
	}
	return best
}

func applyMove(m model, d direction) model {
	if m.state == StateGameOver {
		return m
	}
	if m.state == StateWon && !m.continued {
		return m
	}

	newBoard, gained, moved := slide(m.board, d)
	if !moved {
		return m
	}

	m.prevBoard = m.board
	m.prevScore = m.score
	m.hasPrev = true
	m.board = addTile(newBoard)
	m.score += gained

	if !m.continued && hasWon(m.board) {
		m.state = StateWon
		return m
	}

	if !canMove(m.board) {
		m.state = StateGameOver
	}

	return m
}
