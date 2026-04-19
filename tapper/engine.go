package tapper

import "math/rand"

const (
	Lanes         = 4
	BarWidth      = 24
	MaxLives      = 3
	LifeThreshold = 50
)

type gameState int

const (
	StateModeSelect gameState = iota
	StatePlaying
	StateWaveClear
	StateGameOver
	StateLeaderboard
)

type customerKind int

const (
	KindNormal  customerKind = iota
	KindThirsty              // moves faster, worth 2× points
	KindVIP                  // moves slower, worth 5× points
	KindSlowMo               // triggers slow-motion when served, worth 3× points
)

const (
	SlowMoDuration    = 100 // ticks of slow-motion effect (~5 seconds)
	DoubleTapWindow   = 10  // ticks between serves on the same lane for double-tap bonus
)

type mug struct{ lane, x int }

type customer struct {
	lane, x      int
	retreating   bool
	moveInterval int
	kind         customerKind
}

type serveAnim struct{ lane, x, frames int }

type model struct {
	bartender           int
	mugs                []mug
	customers           []customer
	serveAnims          []serveAnim
	score               int
	lives               int
	wave                int
	tick                int
	state               gameState
	spawnsLeft          int
	spawnTimer          int
	nextLane            int
	paused              bool
	flashFrames         int
	combo               int
	waveLongestCombo    int
	waveServes          int
	waveBonus           int
	nextLifeAt          int
	endless             bool
	slowMoTicks         int
	lastServeTickByLane [Lanes]int
	scores              []ScoreEntry
	scorePath           string
}

func newGame() model {
	path := defaultScorePath()
	return newGameWithScores(loadScores(path), path)
}

func newGameWithScores(scores []ScoreEntry, path string) model {
	m := model{
		lives:      MaxLives,
		nextLifeAt: LifeThreshold,
		scores:     scores,
		scorePath:  path,
	}
	return startWave(m)
}

func startWave(m model) model {
	m.mugs = nil
	m.customers = nil
	m.serveAnims = nil
	m.tick = 0
	m.combo = 0
	m.waveServes = 0
	m.waveLongestCombo = 0
	m.waveBonus = 0
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

func kindPoints(k customerKind) int {
	switch k {
	case KindThirsty:
		return 2
	case KindVIP:
		return 5
	case KindSlowMo:
		return 3
	default:
		return 1
	}
}

// tap fires a mug on the bartender's lane. A second mug is allowed once the
// first has passed the halfway point.
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
// moveInterval==0 means always move (used by test fixtures).
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
				m.combo++
				if m.combo > m.waveLongestCombo {
					m.waveLongestCombo = m.combo
				}
				pts := kindPoints(c.kind) * m.combo
				// Double-tap bonus: quick consecutive serve on same lane
				if m.tick-m.lastServeTickByLane[mg.lane] <= DoubleTapWindow &&
					m.lastServeTickByLane[mg.lane] > 0 {
					pts *= 2
				}
				m.lastServeTickByLane[mg.lane] = m.tick
				m.score += pts
				m.waveServes++
				if m.score >= m.nextLifeAt {
					if m.lives < MaxLives {
						m.lives++
					}
					m.nextLifeAt += LifeThreshold
				}
				// Slow-mo powerup
				if c.kind == KindSlowMo {
					m.slowMoTicks = SlowMoDuration
				}
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

func kindForWave(wave int) customerKind {
	r := rand.Intn(100)
	switch {
	case wave >= 2 && r < 10:
		return KindVIP
	case wave >= 1 && r < 15:
		return KindSlowMo
	case r < 20:
		return KindThirsty
	default:
		return KindNormal
	}
}

// spawnCustomer spawns the next customer with kind-adjusted speed. Later
// spawns within the wave get progressively faster (every 4 spawns, −1 tick).
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
	kind := kindForWave(m.wave)
	switch kind {
	case KindThirsty:
		if interval > 2 {
			interval /= 2
		}
	case KindVIP:
		interval += 6
	}
	m.customers = append(m.customers, customer{
		lane:         lane,
		x:            BarWidth - 1,
		moveInterval: interval,
		kind:         kind,
	})
	return m
}

func loseLife(m model) model {
	if m.combo > m.waveLongestCombo {
		m.waveLongestCombo = m.combo
	}
	m.combo = 0
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

func calcWaveBonus(serves, total, longestCombo int) int {
	b := longestCombo * 3
	if serves == total {
		b += 20
	}
	return b
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

	if m.slowMoTicks > 0 {
		m.slowMoTicks--
	}

	if m.tick%mugMoveInterval(m.wave) == 0 {
		m = advanceMugs(m)
		if m.state != StatePlaying {
			return m
		}
		m = checkCollisions(m)
	}

	// During slow-mo, customers only advance every other tick
	if m.slowMoTicks > 0 && m.tick%2 != 0 {
		return m
	}

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
		total := spawnsForWave(m.wave)
		if m.combo > m.waveLongestCombo {
			m.waveLongestCombo = m.combo
		}
		m.waveBonus = calcWaveBonus(m.waveServes, total, m.waveLongestCombo)
		m.score += m.waveBonus
		m.mugs = nil
		if m.endless {
			m.wave++
			m = startWave(m)
		} else {
			m.state = StateWaveClear
		}
	}

	return m
}
