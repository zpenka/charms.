package snake

import "math/rand"

const (
	Width  = 40
	Height = 20
)

type gameState int

const (
	StatePlaying   gameState = iota
	StateGameOver
	StateLeaderboard
)

type dir int

const (
	DirRight dir = iota
	DirUp
	DirLeft
	DirDown
)

type pos struct{ x, y int }

type model struct {
	snake     []pos
	dir       dir
	nextDir   dir
	food      pos
	score     int
	tick      int
	state     gameState
	scores    []ScoreEntry
	scorePath string
}

func newGame() model {
	path := defaultScorePath()
	return newGameWithScores(loadScores(path), path)
}

func newGameWithScores(scores []ScoreEntry, path string) model {
	cx, cy := Width/2, Height/2
	snake := []pos{{cx, cy}, {cx - 1, cy}, {cx - 2, cy}}
	m := model{
		snake:     snake,
		dir:       DirRight,
		nextDir:   DirRight,
		state:     StatePlaying,
		scores:    scores,
		scorePath: path,
	}
	return placeFood(m)
}

func moveEvery(score int) int {
	switch {
	case score >= 30:
		return 3
	case score >= 20:
		return 4
	case score >= 10:
		return 5
	case score >= 5:
		return 6
	default:
		return 8
	}
}

func isReversal(a, b dir) bool {
	return (a == DirRight && b == DirLeft) ||
		(a == DirLeft && b == DirRight) ||
		(a == DirUp && b == DirDown) ||
		(a == DirDown && b == DirUp)
}

func changeDir(m model, d dir) model {
	if isReversal(d, m.dir) {
		return m
	}
	m.nextDir = d
	return m
}

func placeFood(m model) model {
	occupied := make(map[pos]bool, len(m.snake))
	for _, p := range m.snake {
		occupied[p] = true
	}
	for {
		p := pos{rand.Intn(Width), rand.Intn(Height)}
		if !occupied[p] {
			m.food = p
			return m
		}
	}
}

func moveSnake(m model) model {
	head := m.snake[0]
	var next pos
	switch m.dir {
	case DirRight:
		next = pos{head.x + 1, head.y}
	case DirLeft:
		next = pos{head.x - 1, head.y}
	case DirUp:
		next = pos{head.x, head.y - 1}
	case DirDown:
		next = pos{head.x, head.y + 1}
	}

	// wall collision
	if next.x < 0 || next.x >= Width || next.y < 0 || next.y >= Height {
		m.state = StateGameOver
		return m
	}

	// self collision
	for _, p := range m.snake {
		if p == next {
			m.state = StateGameOver
			return m
		}
	}

	if next == m.food {
		// grow: prepend head, keep tail
		m.snake = append([]pos{next}, m.snake...)
		m.score++
		m = placeFood(m)
	} else {
		// move: prepend head, drop tail
		m.snake = append([]pos{next}, m.snake[:len(m.snake)-1]...)
	}
	return m
}

func tickGame(m model) model {
	if m.state != StatePlaying {
		return m
	}
	m.tick++
	if m.tick%moveEvery(m.score) == 0 {
		m.dir = m.nextDir
		m = moveSnake(m)
	}
	return m
}
