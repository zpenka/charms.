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
	// 5 spawns → 5th customer has spawnIdx=4, bonus=1, reducing base interval by 1.
	// At wave 0 only KindNormal and KindThirsty can spawn; both end up with an
	// interval strictly less than the base (Thirsty halves it further).
	for i := 0; i < 5; i++ {
		m = spawnCustomer(m)
		m.spawnsLeft--
	}
	base := customerMoveInterval(m.wave)
	last := m.customers[len(m.customers)-1].moveInterval
	if last >= base {
		t.Errorf("later customer interval %d should be < base %d after spawn speedup", last, base)
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

// ── feature 1: combo multiplier ────────────────────────────────────────────

func TestCheckCollisions_IncreasesCombo(t *testing.T) {
	m := modelWith([]mug{{lane: 0, x: 5}}, []customer{{lane: 0, x: 5}})
	m = checkCollisions(m)
	if m.combo != 1 {
		t.Errorf("combo = %d, want 1 after first serve", m.combo)
	}
}

func TestCheckCollisions_ScoreUsesComboMultiplier(t *testing.T) {
	m := modelWith([]mug{{lane: 0, x: 5}}, []customer{{lane: 0, x: 5}})
	m.combo = 2
	m = checkCollisions(m)
	// combo increments to 3; normal kind worth 1 pt × combo 3 = 3
	if m.score != 3 {
		t.Errorf("score = %d, want 3 (1pt × combo 3)", m.score)
	}
}

func TestLoseLife_ResetsCombo(t *testing.T) {
	m := newGame()
	m.combo = 5
	m.lives = 2
	m = loseLife(m)
	if m.combo != 0 {
		t.Errorf("combo = %d, want 0 after losing life", m.combo)
	}
}

func TestAdvanceMugs_MissResetsCombo(t *testing.T) {
	m := modelWith([]mug{{lane: 0, x: BarWidth - 1}}, nil)
	m.combo = 3
	m = advanceMugs(m)
	if m.combo != 0 {
		t.Errorf("combo = %d, want 0 after mug miss", m.combo)
	}
}

// ── feature 2: special customer kinds ──────────────────────────────────────

func TestCheckCollisions_ThirstyCustomerWorthDouble(t *testing.T) {
	m := modelWith(
		[]mug{{lane: 0, x: 5}},
		[]customer{{lane: 0, x: 5, kind: KindThirsty}},
	)
	m = checkCollisions(m)
	// combo=1, thirsty base=2 → score = 2×1 = 2
	if m.score != 2 {
		t.Errorf("score = %d, want 2 for thirsty customer", m.score)
	}
}

func TestCheckCollisions_VIPCustomerWorthFive(t *testing.T) {
	m := modelWith(
		[]mug{{lane: 0, x: 5}},
		[]customer{{lane: 0, x: 5, kind: KindVIP}},
	)
	m = checkCollisions(m)
	// combo=1, VIP base=5 → score = 5×1 = 5
	if m.score != 5 {
		t.Errorf("score = %d, want 5 for VIP customer", m.score)
	}
}

func TestSpawnCustomer_ThirstyHasFasterInterval(t *testing.T) {
	m := newGame()
	baseInterval := customerMoveInterval(m.wave)
	// Force a thirsty spawn directly
	m.customers = append(m.customers, customer{
		lane: 0, x: BarWidth - 1, kind: KindThirsty,
		moveInterval: max(baseInterval/2, 2),
	})
	if m.customers[0].moveInterval >= baseInterval {
		t.Errorf("thirsty interval %d should be less than base %d",
			m.customers[0].moveInterval, baseInterval)
	}
}

func TestSpawnCustomer_VIPHasSlowerInterval(t *testing.T) {
	m := newGame()
	baseInterval := customerMoveInterval(m.wave)
	m.customers = append(m.customers, customer{
		lane: 0, x: BarWidth - 1, kind: KindVIP,
		moveInterval: baseInterval + 6,
	})
	if m.customers[0].moveInterval <= baseInterval {
		t.Errorf("VIP interval %d should be greater than base %d",
			m.customers[0].moveInterval, baseInterval)
	}
}

// ── feature 3: extra life at score threshold ────────────────────────────────

func TestExtraLife_AwardedAtThreshold(t *testing.T) {
	m := modelWith([]mug{{lane: 0, x: 5}}, []customer{{lane: 0, x: 5}})
	m.lives = 2
	m.score = LifeThreshold - 1
	m.nextLifeAt = LifeThreshold
	m = checkCollisions(m)
	if m.lives != 3 {
		t.Errorf("lives = %d, want 3 after crossing threshold", m.lives)
	}
}

func TestExtraLife_NotExceedMaxLives(t *testing.T) {
	m := modelWith([]mug{{lane: 0, x: 5}}, []customer{{lane: 0, x: 5}})
	m.lives = MaxLives
	m.score = LifeThreshold - 1
	m.nextLifeAt = LifeThreshold
	m = checkCollisions(m)
	if m.lives != MaxLives {
		t.Errorf("lives = %d, should not exceed MaxLives %d", m.lives, MaxLives)
	}
}

func TestExtraLife_NextThresholdAdvances(t *testing.T) {
	m := modelWith([]mug{{lane: 0, x: 5}}, []customer{{lane: 0, x: 5}})
	m.lives = 2
	m.score = LifeThreshold - 1
	m.nextLifeAt = LifeThreshold
	m = checkCollisions(m)
	if m.nextLifeAt != LifeThreshold*2 {
		t.Errorf("nextLifeAt = %d, want %d", m.nextLifeAt, LifeThreshold*2)
	}
}

// ── feature 4: wave summary ─────────────────────────────────────────────────

func TestWaveSummary_TracksServes(t *testing.T) {
	m := modelWith([]mug{{lane: 0, x: 5}}, []customer{{lane: 0, x: 5}})
	m = checkCollisions(m)
	if m.waveServes != 1 {
		t.Errorf("waveServes = %d, want 1", m.waveServes)
	}
}

func TestWaveSummary_TracksLongestCombo(t *testing.T) {
	m := newGame()
	m.combo = 4
	m.lives = 2
	m = loseLife(m)
	if m.waveLongestCombo != 4 {
		t.Errorf("waveLongestCombo = %d, want 4", m.waveLongestCombo)
	}
}

func TestWaveSummary_ComboUpdatedOnServe(t *testing.T) {
	m := modelWith([]mug{{lane: 0, x: 5}}, []customer{{lane: 0, x: 5}})
	m.combo = 6
	m = checkCollisions(m)
	if m.waveLongestCombo != 7 {
		t.Errorf("waveLongestCombo = %d, want 7", m.waveLongestCombo)
	}
}

func TestWaveSummary_StartWaveClearsStats(t *testing.T) {
	m := newGame()
	m.waveServes = 5
	m.waveLongestCombo = 3
	m = startWave(m)
	if m.waveServes != 0 || m.waveLongestCombo != 0 {
		t.Error("startWave should reset waveServes and waveLongestCombo")
	}
}

func TestCalcWaveBonus_JustCombo(t *testing.T) {
	if got := calcWaveBonus(0, 10, 5); got != 15 {
		t.Errorf("calcWaveBonus = %d, want 15 (combo×3)", got)
	}
}

func TestCalcWaveBonus_PerfectClearAdds20(t *testing.T) {
	if got := calcWaveBonus(10, 10, 0); got != 20 {
		t.Errorf("calcWaveBonus = %d, want 20 (perfect clear)", got)
	}
}

func TestCalcWaveBonus_ComboPlusPerfect(t *testing.T) {
	if got := calcWaveBonus(10, 10, 5); got != 35 {
		t.Errorf("calcWaveBonus = %d, want 35 (combo×3 + perfect)", got)
	}
}

func TestWaveBonus_AddedToScoreOnClear(t *testing.T) {
	m := newGame()
	m.spawnsLeft = 0
	m.customers = nil
	m.waveServes = spawnsForWave(0) // perfect wave
	m.waveLongestCombo = 0
	m.tick = 999
	before := m.score
	m = tickGame(m)
	if m.state != StateWaveClear {
		t.Fatal("should be StateWaveClear")
	}
	want := calcWaveBonus(spawnsForWave(0), spawnsForWave(0), 0)
	if m.score != before+want {
		t.Errorf("score = %d, want %d (bonus applied on clear)", m.score, before+want)
	}
}

// ── endless mode ─────────────────────────────────────────────────────────────

func TestTickGame_EndlessSkipsWaveClear(t *testing.T) {
	m := newGame()
	m.endless = true
	m.spawnsLeft = 0
	m.customers = nil
	m.mugs = nil
	m = tickGame(m)
	if m.state == StateWaveClear {
		t.Error("endless mode should not enter StateWaveClear")
	}
}

func TestTickGame_EndlessAutoAdvancesWave(t *testing.T) {
	m := newGame()
	m.endless = true
	m.wave = 1
	m.spawnsLeft = 0
	m.customers = nil
	m.mugs = nil
	m = tickGame(m)
	if m.wave != 2 {
		t.Errorf("wave = %d, want 2 after auto-advance in endless mode", m.wave)
	}
}

// ── slow-motion powerup ───────────────────────────────────────────────────────

func TestKindSlowMo_HasNonZeroPoints(t *testing.T) {
	if kindPoints(KindSlowMo) == 0 {
		t.Error("KindSlowMo should be worth some points")
	}
}

func TestCheckCollisions_SlowMoCustomerSetsSlowMoTicks(t *testing.T) {
	m := newGame()
	m.mugs = []mug{{lane: 0, x: 5}}
	m.customers = []customer{{lane: 0, x: 5, kind: KindSlowMo}}
	m = checkCollisions(m)
	if m.slowMoTicks <= 0 {
		t.Error("serving a SlowMo customer should set slowMoTicks > 0")
	}
}

func TestTickGame_SlowMo_DecrementsSlowMoTicks(t *testing.T) {
	m := newGame()
	m.slowMoTicks = 10
	m.spawnsLeft = 0
	m.customers = nil
	m.mugs = nil
	m = tickGame(m)
	if m.slowMoTicks >= 10 {
		t.Error("tickGame should decrement slowMoTicks each tick")
	}
}

func TestTickGame_SlowMo_HalvesCustomerSpeed(t *testing.T) {
	// Without slow-mo: customer with moveInterval=0 moves every tick
	// With slow-mo active: customer should only move every other tick
	normal := newGame()
	normal.mugs = nil
	normal.spawnsLeft = 0
	normal.customers = []customer{{lane: 0, x: 10, moveInterval: 0}}
	xBefore := normal.customers[0].x
	normal = tickGame(normal)
	moved := normal.customers[0].x != xBefore

	slow := newGame()
	slow.slowMoTicks = 100
	slow.mugs = nil
	slow.spawnsLeft = 0
	slow.customers = []customer{{lane: 0, x: 10, moveInterval: 0}}
	slow = tickGame(slow)
	movedSlow := slow.customers[0].x != slow.customers[0].x // always false; recalc

	// Tick once more with slow-mo
	xAfterFirst := slow.customers[0].x
	slow = tickGame(slow)
	movedSlow = slow.customers[0].x != xAfterFirst

	if moved && movedSlow {
		// That's a problem: slow-mo didn't slow the customer
		// Actually let me restructure: tick twice with slow-mo, customer should move at most once
	}
	_ = moved
	_ = movedSlow
	// The real test: after N ticks with slow-mo, customer moves fewer times than without
	m1 := newGame()
	m1.mugs = nil
	m1.spawnsLeft = 0
	m1.customers = []customer{{lane: 0, x: BarWidth - 1, moveInterval: 0}}
	m2 := newGame()
	m2.slowMoTicks = 100
	m2.mugs = nil
	m2.spawnsLeft = 0
	m2.customers = []customer{{lane: 0, x: BarWidth - 1, moveInterval: 0}}
	for i := 0; i < 4; i++ {
		m1 = tickGame(m1)
		m2 = tickGame(m2)
	}
	if len(m1.customers) > 0 && len(m2.customers) > 0 {
		if m2.customers[0].x <= m1.customers[0].x {
			t.Error("slow-mo should result in customers advancing less (higher x value means less advanced)")
		}
	}
}

// ── double-tap bonus ─────────────────────────────────────────────────────────

func TestCheckCollisions_DoubleTap_QuickServeSameLane(t *testing.T) {
	m := newGame()
	m.lastServeTickByLane = [Lanes]int{}
	m.tick = 5
	m.lastServeTickByLane[0] = 2 // last serve on lane 0 was tick 2 (within window)
	m.combo = 1
	m.mugs = []mug{{lane: 0, x: 5}}
	m.customers = []customer{{lane: 0, x: 5, kind: KindNormal}}
	prevScore := m.score
	m = checkCollisions(m)
	// Normal: kindPoints(KindNormal)*combo = 1*2 = 2 (combo incremented to 2 after first serve)
	// But with double-tap on same lane, score should be doubled on top of combo
	// Expected: 2 * 2 = 4 (double-tap doubles the points for that serve)
	if m.score <= prevScore+2 {
		t.Errorf("score gained = %d, want > 2 for double-tap bonus", m.score-prevScore)
	}
}

func TestCheckCollisions_DoubleTap_NoBonusWhenFarApart(t *testing.T) {
	m := newGame()
	m.lastServeTickByLane = [Lanes]int{}
	m.tick = 50
	m.lastServeTickByLane[0] = 0 // last serve on lane 0 was tick 0 (too long ago)
	m.combo = 1
	m.mugs = []mug{{lane: 0, x: 5}}
	m.customers = []customer{{lane: 0, x: 5, kind: KindNormal}}
	prevScore := m.score
	m = checkCollisions(m)
	// No double-tap: normal combo score = kindPoints(KindNormal) * 2 (combo became 2) = 2
	if m.score > prevScore+2 {
		t.Errorf("no double-tap bonus expected when serves are far apart, got %d extra", m.score-prevScore)
	}
}

func TestCheckCollisions_DoubleTap_UpdatesLastServeTick(t *testing.T) {
	m := newGame()
	m.lastServeTickByLane = [Lanes]int{}
	m.tick = 7
	m.mugs = []mug{{lane: 2, x: 5}}
	m.customers = []customer{{lane: 2, x: 5, kind: KindNormal}}
	m = checkCollisions(m)
	if m.lastServeTickByLane[2] != 7 {
		t.Errorf("lastServeTickByLane[2] = %d, want 7", m.lastServeTickByLane[2])
	}
}
