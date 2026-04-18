package tapper

const (
	Lanes    = 4
	BarWidth = 24
	MaxLives = 3
)

type gameState int

const (
	StatePlaying   gameState = iota
	StateWaveClear
	StateGameOver
	StateLeaderboard
)

type mug struct{ lane, x int }

type customer struct {
	lane, x      int
	retreating   bool
	moveInterval int
}

type serveAnim struct{ lane, x, frames int }

type model struct {
	bartender   int
	mugs        []mug
	customers   []customer
	serveAnims  []serveAnim
	score       int
	lives       int
	wave        int
	tick        int
	state       gameState
	spawnsLeft  int
	spawnTimer  int
	nextLane    int
	paused      bool
	flashFrames int
	scores      []ScoreEntry
	scorePath   string
}

func newGame() model {
	path := defaultScorePath()
	return newGameWithScores(loadScores(path), path)
}

func newGameWithScores(scores []ScoreEntry, path string) model {
	m := model{lives: MaxLives, scores: scores, scorePath: path}
	return startWave(m)
}

func startWave(m model) model {
	m.mugs = nil
	m.customers = nil
	m.serveAnims = nil
	m.tick = 0
	m.spawnsLeft = spawnsForWave(m.wave)
	m.spawnTimer = spawnInterval(m.wave)
	m.nextLane = 0
	m.state = StatePlaying
	return m
}

func spawnsForWave(wave int) int {
	return 8 + wave*2
}

func spawnInterval(wave int) int {
	switch {
	case wave <= 0:
		return 30
	case wave == 1:
		return 25
	case wave == 2:
		return 20
	case wave == 3:
		return 15
	default:
		return 12
	}
}

func mugMoveInterval(_ int) int { return 2 }

func customerMoveInterval(wave int) int {
	intervals := []int{12, 9, 7, 5, 4}
	if wave < len(intervals) {
		return intervals[wave]
	}
	return 3
}

// tap fires a mug on the bartender's lane. A second mug is allowed if the
// first has already passed the halfway point.
func tap(m model) model {
	for _, mg := range m.mugs {
		if mg.lane == m.bartender && mg.x < BarWidth/2 {
			return m
		}
	}
	m.mugs = append(m.mugs, mug{lane: m.bartender, x: 0})
	return m
}

func advanceMugs(m model) model {
	var keep []mug
	for _, mg := range m.mugs {
		mg.x++
		if mg.x >= BarWidth {
			m = loseLife(m)
			if m.state != StatePlaying {
				return m
			}
		} else {
			keep = append(keep, mg)
		}
	}
	m.mugs = keep
	return m
}

// advanceCustomers moves each customer according to its own moveInterval.
// moveInterval==0 means always move (used by direct-construction test fixtures).
func advanceCustomers(m model) model {
	var keep []customer
	for _, c := range m.customers {
		if c.moveInterval > 0 && m.tick%c.moveInterval != 0 {
			keep = append(keep, c)
			continue
		}
		if c.retreating {
			c.x++
			if c.x < BarWidth {
				keep = append(keep, c)
			}
		} else {
			c.x--
			keep = append(keep, c)
		}
	}
	m.customers = keep
	return m
}

func checkBreaches(m model) model {
	var keep []customer
	for _, c := range m.customers {
		if !c.retreating && c.x < 0 {
			m = loseLife(m)
			if m.state != StatePlaying {
				return m
			}
		} else {
			keep = append(keep, c)
		}
	}
	m.customers = keep
	return m
}

func checkCollisions(m model) model {
	var remainingMugs []mug
	for _, mg := range m.mugs {
		hit := false
		for i := range m.customers {
			c := &m.customers[i]
			if !c.retreating && mg.lane == c.lane && mg.x == c.x {
				c.retreating = true
				m.score++
				m.serveAnims = append(m.serveAnims, serveAnim{lane: mg.lane, x: mg.x, frames: 3})
				hit = true
				break
			}
		}
		if !hit {
			remainingMugs = append(remainingMugs, mg)
		}
	}
	m.mugs = remainingMugs
	return m
}

func tickServeAnims(m model) model {
	var keep []serveAnim
	for _, a := range m.serveAnims {
		a.frames--
		if a.frames > 0 {
			keep = append(keep, a)
		}
	}
	m.serveAnims = keep
	return m
}

// spawnCustomer spawns the next customer, assigning a faster moveInterval to
// later spawns within the wave (every 4 spawns, the interval drops by 1).
func spawnCustomer(m model) model {
	lane := m.nextLane % Lanes
	m.nextLane++
	spawnIdx := spawnsForWave(m.wave) - m.spawnsLeft
	interval := customerMoveInterval(m.wave)
	if bonus := spawnIdx / 4; bonus > 0 {
		interval -= bonus
		if interval < 2 {
			interval = 2
		}
	}
	m.customers = append(m.customers, customer{lane: lane, x: BarWidth - 1, moveInterval: interval})
	return m
}

func loseLife(m model) model {
	m.lives--
	if m.lives <= 0 {
		m.state = StateGameOver
		return m
	}
	m.mugs = nil
	m.customers = nil
	m.spawnTimer = spawnInterval(m.wave)
	m.flashFrames = 4
	return m
}

func tickGame(m model) model {
	if m.paused || m.state != StatePlaying {
		return m
	}
	if m.flashFrames > 0 {
		m.flashFrames--
		return m
	}
	m.tick++

	if m.tick%mugMoveInterval(m.wave) == 0 {
		m = advanceMugs(m)
		if m.state != StatePlaying {
			return m
		}
		m = checkCollisions(m)
	}

	// Per-customer intervals: advance every tick; each customer decides itself.
	m = advanceCustomers(m)
	m = checkBreaches(m)
	if m.state != StatePlaying {
		return m
	}
	m = checkCollisions(m)

	m = tickServeAnims(m)

	m.spawnTimer--
	if m.spawnTimer <= 0 && m.spawnsLeft > 0 {
		m = spawnCustomer(m)
		m.spawnsLeft--
		m.spawnTimer = spawnInterval(m.wave)
	}

	if m.spawnsLeft == 0 && len(m.customers) == 0 {
		m.state = StateWaveClear
		m.mugs = nil
	}

	return m
}
