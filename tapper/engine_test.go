package tapper

import "testing"

// helpers

func modelWith(mugs []mug, customers []customer) model {
	m := newGame()
	m.mugs = mugs
	m.customers = customers
	return m
}

// tap

func TestTap_CreatesMugOnBartenderLane(t *testing.T) {
	m := newGame()
	m.bartender = 2
	m = tap(m)
	if len(m.mugs) != 1 {
		t.Fatalf("want 1 mug, got %d", len(m.mugs))
	}
	if m.mugs[0].lane != 2 || m.mugs[0].x != 0 {
		t.Errorf("mug = %+v, want {lane:2 x:0}", m.mugs[0])
	}
}

func TestTap_NoDoubleOnSameLane(t *testing.T) {
	m := newGame()
	m.bartender = 1
	m = tap(m)
	m = tap(m)
	if len(m.mugs) != 1 {
		t.Errorf("want 1 mug (no double), got %d", len(m.mugs))
	}
}

func TestTap_DifferentLanesAllowed(t *testing.T) {
	m := newGame()
	m.bartender = 0
	m = tap(m)
	m.bartender = 1
	m = tap(m)
	if len(m.mugs) != 2 {
		t.Errorf("want 2 mugs on different lanes, got %d", len(m.mugs))
	}
}

// advanceMugs

func TestAdvanceMugs_MovesRight(t *testing.T) {
	m := modelWith([]mug{{lane: 0, x: 5}}, nil)
	m = advanceMugs(m)
	if m.mugs[0].x != 6 {
		t.Errorf("mug x = %d, want 6", m.mugs[0].x)
	}
}

func TestAdvanceMugs_MissLosesLife(t *testing.T) {
	m := modelWith([]mug{{lane: 0, x: BarWidth - 1}}, nil)
	before := m.lives
	m = advanceMugs(m)
	if m.lives != before-1 {
		t.Errorf("lives = %d, want %d", m.lives, before-1)
	}
}

func TestAdvanceMugs_MissedMugRemoved(t *testing.T) {
	m := modelWith([]mug{{lane: 0, x: BarWidth - 1}}, nil)
	m = advanceMugs(m)
	if len(m.mugs) != 0 {
		t.Errorf("missed mug should be removed, got %d mugs", len(m.mugs))
	}
}

func TestAdvanceMugs_GameOverWhenLivesExhausted(t *testing.T) {
	m := modelWith([]mug{{lane: 0, x: BarWidth - 1}}, nil)
	m.lives = 1
	m = advanceMugs(m)
	if m.state != StateGameOver {
		t.Error("should be game over when last life lost on miss")
	}
}

// advanceCustomers

func TestAdvanceCustomers_MovesLeft(t *testing.T) {
	m := modelWith(nil, []customer{{lane: 0, x: 10}})
	m = advanceCustomers(m)
	if m.customers[0].x != 9 {
		t.Errorf("customer x = %d, want 9", m.customers[0].x)
	}
}

func TestAdvanceCustomers_RetreatingMovesRight(t *testing.T) {
	m := modelWith(nil, []customer{{lane: 0, x: 10, retreating: true}})
	m = advanceCustomers(m)
	if m.customers[0].x != 11 {
		t.Errorf("retreating customer x = %d, want 11", m.customers[0].x)
	}
}

func TestAdvanceCustomers_RetreatedOffRightRemoved(t *testing.T) {
	m := modelWith(nil, []customer{{lane: 0, x: BarWidth - 1, retreating: true}})
	m = advanceCustomers(m)
	if len(m.customers) != 0 {
		t.Errorf("fully retreated customer should be removed, got %d", len(m.customers))
	}
}

// checkCollisions

func TestCheckCollisions_ServesMatchingMugAndCustomer(t *testing.T) {
	m := modelWith(
		[]mug{{lane: 1, x: 8}},
		[]customer{{lane: 1, x: 8}},
	)
	before := m.score
	m = checkCollisions(m)
	if m.score != before+1 {
		t.Errorf("score = %d, want %d", m.score, before+1)
	}
	if len(m.mugs) != 0 {
		t.Error("served mug should be removed")
	}
	if !m.customers[0].retreating {
		t.Error("served customer should be retreating")
	}
}

func TestCheckCollisions_NoHitDifferentLane(t *testing.T) {
	m := modelWith(
		[]mug{{lane: 0, x: 8}},
		[]customer{{lane: 1, x: 8}},
	)
	before := m.score
	m = checkCollisions(m)
	if m.score != before {
		t.Error("different lane should not score")
	}
}

func TestCheckCollisions_NoHitDifferentX(t *testing.T) {
	m := modelWith(
		[]mug{{lane: 0, x: 7}},
		[]customer{{lane: 0, x: 9}},
	)
	before := m.score
	m = checkCollisions(m)
	if m.score != before {
		t.Error("different position should not score")
	}
}

func TestCheckCollisions_RetreatingCustomerNotServed(t *testing.T) {
	m := modelWith(
		[]mug{{lane: 0, x: 8}},
		[]customer{{lane: 0, x: 8, retreating: true}},
	)
	before := m.score
	m = checkCollisions(m)
	if m.score != before {
		t.Error("retreating customer should not be served again")
	}
}

// checkBreaches

func TestCheckBreaches_CustomerAtNegativeXLosesLife(t *testing.T) {
	m := modelWith(nil, []customer{{lane: 0, x: -1}})
	before := m.lives
	m = checkBreaches(m)
	if m.lives != before-1 {
		t.Errorf("lives = %d, want %d", m.lives, before-1)
	}
}

func TestCheckBreaches_BreachedCustomerRemoved(t *testing.T) {
	m := modelWith(nil, []customer{{lane: 0, x: -1}})
	m = checkBreaches(m)
	if len(m.customers) != 0 {
		t.Errorf("breached customer should be removed, got %d", len(m.customers))
	}
}

func TestCheckBreaches_RetreatingCustomerAtNegativeXIgnored(t *testing.T) {
	// Retreating customers can't breach — they're going the other way
	m := modelWith(nil, []customer{{lane: 0, x: -1, retreating: true}})
	before := m.lives
	m = checkBreaches(m)
	if m.lives != before {
		t.Error("retreating customer at negative x should not lose a life")
	}
}

// loseLife

func TestLoseLife_Decrements(t *testing.T) {
	m := newGame()
	before := m.lives
	m = loseLife(m)
	if m.lives != before-1 {
		t.Errorf("lives = %d, want %d", m.lives, before-1)
	}
}

func TestLoseLife_ClearsMugsAndCustomers(t *testing.T) {
	m := modelWith([]mug{{0, 5}}, []customer{{lane: 0, x: 10}})
	m = loseLife(m)
	if len(m.mugs) != 0 || len(m.customers) != 0 {
		t.Error("loseLife should clear mugs and customers when lives remain")
	}
}

func TestLoseLife_GameOverAtZero(t *testing.T) {
	m := newGame()
	m.lives = 1
	m = loseLife(m)
	if m.state != StateGameOver {
		t.Error("should be game over when lives reach 0")
	}
}

// spawnCustomer

func TestSpawnCustomer_AddsAtRightEnd(t *testing.T) {
	m := newGame()
	m = spawnCustomer(m)
	if len(m.customers) != 1 {
		t.Fatalf("want 1 customer, got %d", len(m.customers))
	}
	if m.customers[0].x != BarWidth-1 {
		t.Errorf("spawn x = %d, want %d", m.customers[0].x, BarWidth-1)
	}
}

func TestSpawnCustomer_RoundRobinLanes(t *testing.T) {
	m := newGame()
	for i := 0; i < Lanes; i++ {
		m = spawnCustomer(m)
	}
	seen := make(map[int]bool)
	for _, c := range m.customers {
		seen[c.lane] = true
	}
	if len(seen) != Lanes {
		t.Errorf("want customers on all %d lanes, got %d distinct", Lanes, len(seen))
	}
}

// speed functions

func TestCustomerMoveInterval_DecreasesWithWave(t *testing.T) {
	prev := customerMoveInterval(0)
	for wave := 1; wave <= 4; wave++ {
		cur := customerMoveInterval(wave)
		if cur >= prev {
			t.Errorf("wave %d: interval %d should be less than wave %d: %d", wave, cur, wave-1, prev)
		}
		prev = cur
	}
}

func TestMugMoveInterval_Constant(t *testing.T) {
	for wave := 0; wave < 5; wave++ {
		if mugMoveInterval(wave) != mugMoveInterval(0) {
			t.Error("mug speed should be constant across waves")
		}
	}
}

func TestSpawnsForWave_IncreasesWithWave(t *testing.T) {
	if spawnsForWave(1) <= spawnsForWave(0) {
		t.Error("later waves should spawn more customers")
	}
}

// tickGame

func TestTickGame_DoesNothingWhenNotPlaying(t *testing.T) {
	m := newGame()
	m.state = StateGameOver
	before := m.tick
	m = tickGame(m)
	if m.tick != before {
		t.Error("tickGame should be a no-op when state is not StatePlaying")
	}
}

func TestTickGame_IncrementsTick(t *testing.T) {
	m := newGame()
	m = tickGame(m)
	if m.tick != 1 {
		t.Errorf("tick = %d, want 1", m.tick)
	}
}

func TestWaveClear_WhenAllCustomersServedAndNoSpawnsLeft(t *testing.T) {
	m := newGame()
	m.spawnsLeft = 0
	m.customers = nil
	m.mugs = nil
	// Force a tick that isn't a move tick to just trigger the wave-clear check
	m.tick = 999
	m = tickGame(m)
	if m.state != StateWaveClear {
		t.Error("should be wave clear when no spawns left and no customers")
	}
}

// newGame / startWave

func TestNewGame_InitialState(t *testing.T) {
	m := newGame()
	if m.lives != MaxLives {
		t.Errorf("lives = %d, want %d", m.lives, MaxLives)
	}
	if m.score != 0 {
		t.Errorf("score = %d, want 0", m.score)
	}
	if m.wave != 0 {
		t.Errorf("wave = %d, want 0", m.wave)
	}
	if m.state != StatePlaying {
		t.Errorf("state = %v, want StatePlaying", m.state)
	}
}

// pause

func TestTickGame_NoopWhenPaused(t *testing.T) {
	m := newGame()
	m.paused = true
	before := m.tick
	m = tickGame(m)
	if m.tick != before {
		t.Error("tickGame should not advance while paused")
	}
}

// flash

func TestLoseLife_SetsFlashFrames(t *testing.T) {
	m := newGame()
	m.lives = 2
	m = loseLife(m)
	if m.flashFrames == 0 {
		t.Error("loseLife should set flashFrames > 0 when lives remain")
	}
}

func TestLoseLife_NoFlashOnGameOver(t *testing.T) {
	m := newGame()
	m.lives = 1
	m = loseLife(m)
	if m.state != StateGameOver {
		t.Error("should be game over")
	}
}

func TestTickGame_FlashPausesEngine(t *testing.T) {
	m := newGame()
	m.flashFrames = 3
	before := m.tick
	m = tickGame(m)
	if m.tick != before {
		t.Error("tick should not advance while flashing")
	}
}

func TestTickGame_FlashDecrementsEachTick(t *testing.T) {
	m := newGame()
	m.flashFrames = 3
	m = tickGame(m)
	if m.flashFrames != 2 {
		t.Errorf("flashFrames = %d, want 2", m.flashFrames)
	}
}

// startWave

func TestStartWave_ClearsBarAndResetsSpawns(t *testing.T) {
	m := modelWith([]mug{{0, 5}}, []customer{{lane: 0, x: 10}})
	m.wave = 2
	m = startWave(m)
	if len(m.mugs) != 0 || len(m.customers) != 0 {
		t.Error("startWave should clear mugs and customers")
	}
	if m.spawnsLeft != spawnsForWave(2) {
		t.Errorf("spawnsLeft = %d, want %d", m.spawnsLeft, spawnsForWave(2))
	}
	if m.state != StatePlaying {
		t.Error("startWave should set state to StatePlaying")
	}
}

// multi-mug per lane

func TestTap_AllowsSecondMugWhenFirstPastHalfway(t *testing.T) {
	m := newGame()
	m.bartender = 0
	m.mugs = []mug{{lane: 0, x: BarWidth / 2}}
	m = tap(m)
	if len(m.mugs) != 2 {
		t.Errorf("want 2 mugs when first is past halfway, got %d", len(m.mugs))
	}
}

func TestTap_BlocksSecondMugWhenFirstBeforeHalfway(t *testing.T) {
	m := newGame()
	m.bartender = 0
	m.mugs = []mug{{lane: 0, x: BarWidth/2 - 1}}
	m = tap(m)
	if len(m.mugs) != 1 {
		t.Errorf("want 1 mug (blocked before halfway), got %d", len(m.mugs))
	}
}

// serve animation

func TestCheckCollisions_CreatesServeAnim(t *testing.T) {
	m := modelWith(
		[]mug{{lane: 0, x: 8}},
		[]customer{{lane: 0, x: 8}},
	)
	m = checkCollisions(m)
	if len(m.serveAnims) != 1 {
		t.Fatalf("want 1 serveAnim, got %d", len(m.serveAnims))
	}
	if m.serveAnims[0].lane != 0 || m.serveAnims[0].x != 8 {
		t.Errorf("serveAnim = %+v, want {lane:0 x:8}", m.serveAnims[0])
	}
	if m.serveAnims[0].frames <= 0 {
		t.Error("serveAnim frames should be > 0")
	}
}

func TestTickServeAnims_DecrementsFrames(t *testing.T) {
	m := newGame()
	m.serveAnims = []serveAnim{{lane: 0, x: 5, frames: 3}}
	m = tickServeAnims(m)
	if len(m.serveAnims) != 1 {
		t.Fatalf("anim should still exist, got %d", len(m.serveAnims))
	}
	if m.serveAnims[0].frames != 2 {
		t.Errorf("frames = %d, want 2", m.serveAnims[0].frames)
	}
}

func TestTickServeAnims_RemovesExpiredAnims(t *testing.T) {
	m := newGame()
	m.serveAnims = []serveAnim{{lane: 0, x: 5, frames: 1}}
	m = tickServeAnims(m)
	if len(m.serveAnims) != 0 {
		t.Errorf("expired anim should be removed, got %d", len(m.serveAnims))
	}
}

// per-customer speed

func TestSpawnCustomer_LaterSpawnsHaveFasterInterval(t *testing.T) {
	m := newGame()
	// spawn enough to trigger speedup (every 4 spawns)
	for i := 0; i < 5; i++ {
		m = spawnCustomer(m)
		m.spawnsLeft--
	}
	first := m.customers[0].moveInterval
	last := m.customers[len(m.customers)-1].moveInterval
	if last >= first {
		t.Errorf("later customer interval %d should be less than first %d", last, first)
	}
}

func TestAdvanceCustomers_RespectsPerCustomerInterval(t *testing.T) {
	m := newGame()
	m.tick = 5
	// moveInterval=10: tick 5 → 5%10 != 0, should NOT move
	m.customers = []customer{{lane: 0, x: 10, moveInterval: 10}}
	m = advanceCustomers(m)
	if m.customers[0].x != 10 {
		t.Errorf("customer should not move when tick%%interval != 0, x = %d", m.customers[0].x)
	}
	// moveInterval=5: tick 5 → 5%5 == 0, SHOULD move
	m.customers = []customer{{lane: 0, x: 10, moveInterval: 5}}
	m = advanceCustomers(m)
	if m.customers[0].x != 9 {
		t.Errorf("customer should move when tick%%interval == 0, x = %d", m.customers[0].x)
	}
}
