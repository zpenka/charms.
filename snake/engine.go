package snake

import "math/rand"

const (
	Width            = 40
	Height           = 20
	GhostDuration    = 10 // moves of ghost mode after eating bonus food
	BonusFoodDuration = 40 // ticks before bonus food expires
	BonusSpawnEvery  = 30  // ticks between bonus food spawn attempts
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
	snake           []pos
	dir             dir
	nextDir         dir
	food            pos
	bonusFood       pos
	bonusFoodActive bool
	bonusFoodTicks  int
	ghostTicks      int
	obstacles       []pos
	score           int
	tick            int
	state           gameState
	scores          []ScoreEntry
	scorePath       string
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
	m = placeObstacles(m)
	return placeFood(m)
}

func placeObstacles(m model) model {
	// Build a set of positions occupied by the starting snake
	occupied := make(map[pos]bool, len(m.snake))
	for _, p := range m.snake {
		occupied[p] = true
	}
	// Scatter ~10% of cells as obstacle tiles, avoiding the snake start area
	count := (Width * Height) / 10
	var obs []pos
	for len(obs) < count {
		p := pos{rand.Intn(Width), rand.Intn(Height)}
		if !occupied[p] {
			obs = append(obs, p)
			occupied[p] = true
		}
	}
	m.obstacles = obs
	return m
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
	occupied := make(map[pos]bool, len(m.snake)+len(m.obstacles))
	for _, p := range m.snake {
		occupied[p] = true
	}
	for _, p := range m.obstacles {
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

func addObstacle(m model) model {
	occupied := make(map[pos]bool, len(m.snake)+len(m.obstacles)+1)
	for _, p := range m.snake {
		occupied[p] = true
	}
	for _, p := range m.obstacles {
		occupied[p] = true
	}
	occupied[m.food] = true
	for {
		p := pos{rand.Intn(Width), rand.Intn(Height)}
		if !occupied[p] {
			m.obstacles = append(m.obstacles, p)
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

	// portal walls: wrap around edges
	next.x = (next.x + Width) % Width
	next.y = (next.y + Height) % Height

	// obstacle collision — skipped in ghost mode
	if m.ghostTicks <= 0 {
		for _, p := range m.obstacles {
			if p == next {
				m.state = StateGameOver
				return m
			}
		}
	}

	// self collision
	for _, p := range m.snake {
		if p == next {
			m.state = StateGameOver
			return m
		}
	}

	if m.bonusFoodActive && next == m.bonusFood {
		m.score += 3
		m.bonusFoodActive = false
		m.ghostTicks = GhostDuration
		m.snake = append([]pos{next}, m.snake...) // bonus food grows snake
	} else {
		if m.ghostTicks > 0 {
			m.ghostTicks--
		}
		if next == m.food {
			m.snake = append([]pos{next}, m.snake...)
			m.score++
			if m.score%5 == 0 {
				m = addObstacle(m)
			}
			m = placeFood(m)
		} else {
			m.snake = append([]pos{next}, m.snake[:len(m.snake)-1]...)
		}
	}
	return m
}

func placeBonusFood(m model) model {
	occupied := make(map[pos]bool, len(m.snake)+len(m.obstacles)+1)
	for _, p := range m.snake {
		occupied[p] = true
	}
	for _, p := range m.obstacles {
		occupied[p] = true
	}
	occupied[m.food] = true
	for {
		p := pos{rand.Intn(Width), rand.Intn(Height)}
		if !occupied[p] {
			m.bonusFood = p
			m.bonusFoodActive = true
			m.bonusFoodTicks = BonusFoodDuration
			return m
		}
	}
}

func tickGame(m model) model {
	if m.state != StatePlaying {
		return m
	}
	m.tick++

	if m.bonusFoodActive {
		m.bonusFoodTicks--
		if m.bonusFoodTicks <= 0 {
			m.bonusFoodActive = false
		}
	} else if m.tick%BonusSpawnEvery == 0 {
		m = placeBonusFood(m)
	}

	if m.tick%moveEvery(m.score) == 0 {
		m.dir = m.nextDir
		m = moveSnake(m)
	}
	return m
}
