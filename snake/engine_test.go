package snake

import "testing"

// ── newGame ──────────────────────────────────────────────────────────────────

func TestNewGame_InitialState(t *testing.T) {
	m := newGame()
	if m.score != 0 {
		t.Errorf("score = %d, want 0", m.score)
	}
	if m.state != StatePlaying {
		t.Errorf("state = %v, want StatePlaying", m.state)
	}
	if len(m.snake) < 3 {
		t.Errorf("snake length = %d, want >= 3", len(m.snake))
	}
	if m.dir != DirRight {
		t.Errorf("dir = %v, want DirRight", m.dir)
	}
}

func TestNewGame_FoodNotOnSnake(t *testing.T) {
	m := newGame()
	for _, p := range m.snake {
		if p == m.food {
			t.Error("food should not overlap snake at start")
		}
	}
}

// ── moveSnake ────────────────────────────────────────────────────────────────

func TestMoveSnake_MovesHeadRight(t *testing.T) {
	m := newGame()
	m.food = pos{0, 0} // keep food out of the way
	head := m.snake[0]
	m = moveSnake(m)
	want := pos{head.x + 1, head.y}
	if m.snake[0] != want {
		t.Errorf("head = %v, want %v", m.snake[0], want)
	}
}

func TestMoveSnake_TailDropsOnMove(t *testing.T) {
	m := newGame()
	m.food = pos{0, 0}
	length := len(m.snake)
	m = moveSnake(m)
	if len(m.snake) != length {
		t.Errorf("length = %d, want %d (tail should drop)", len(m.snake), length)
	}
}

func TestMoveSnake_GrowsOnFood(t *testing.T) {
	m := newGame()
	length := len(m.snake)
	m.food = pos{m.snake[0].x + 1, m.snake[0].y}
	m = moveSnake(m)
	if len(m.snake) != length+1 {
		t.Errorf("length = %d, want %d (should grow on food)", len(m.snake), length+1)
	}
}

func TestMoveSnake_ScoreIncreasesOnFood(t *testing.T) {
	m := newGame()
	m.food = pos{m.snake[0].x + 1, m.snake[0].y}
	m = moveSnake(m)
	if m.score != 1 {
		t.Errorf("score = %d, want 1", m.score)
	}
}

func TestMoveSnake_WallCollision_Right(t *testing.T) {
	m := newGame()
	m.snake = []pos{{Width - 1, 5}}
	m.dir = DirRight
	m = moveSnake(m)
	if m.state != StateGameOver {
		t.Error("should be game over hitting right wall")
	}
}

func TestMoveSnake_WallCollision_Left(t *testing.T) {
	m := newGame()
	m.snake = []pos{{0, 5}}
	m.dir = DirLeft
	m.nextDir = DirLeft
	m = moveSnake(m)
	if m.state != StateGameOver {
		t.Error("should be game over hitting left wall")
	}
}

func TestMoveSnake_WallCollision_Top(t *testing.T) {
	m := newGame()
	m.snake = []pos{{5, 0}}
	m.dir = DirUp
	m.nextDir = DirUp
	m = moveSnake(m)
	if m.state != StateGameOver {
		t.Error("should be game over hitting top wall")
	}
}

func TestMoveSnake_WallCollision_Bottom(t *testing.T) {
	m := newGame()
	m.snake = []pos{{5, Height - 1}}
	m.dir = DirDown
	m.nextDir = DirDown
	m = moveSnake(m)
	if m.state != StateGameOver {
		t.Error("should be game over hitting bottom wall")
	}
}

func TestMoveSnake_SelfCollision(t *testing.T) {
	m := newGame()
	// Snake coiled so head (10,5) going right will hit body at (11,5)
	m.snake = []pos{{10, 5}, {10, 6}, {11, 6}, {11, 5}, {12, 5}}
	m.dir = DirRight
	m.nextDir = DirRight
	m = moveSnake(m)
	if m.state != StateGameOver {
		t.Error("should be game over on self collision")
	}
}

// ── changeDir ────────────────────────────────────────────────────────────────

func TestChangeDir_BuffersDirection(t *testing.T) {
	m := newGame() // dir=DirRight
	m = changeDir(m, DirUp)
	if m.nextDir != DirUp {
		t.Errorf("nextDir = %v, want DirUp", m.nextDir)
	}
}

func TestChangeDir_BlocksReversal(t *testing.T) {
	m := newGame() // dir=DirRight
	before := m.nextDir
	m = changeDir(m, DirLeft)
	if m.nextDir != before {
		t.Error("should block 180° reversal")
	}
}

func TestChangeDir_AllowsPerpendicular(t *testing.T) {
	m := newGame() // dir=DirRight
	m = changeDir(m, DirUp)
	if m.nextDir != DirUp {
		t.Error("should allow 90° direction change")
	}
}

func TestChangeDir_BlocksDownWhenMovingUp(t *testing.T) {
	m := newGame()
	m.dir = DirUp
	m.nextDir = DirUp
	before := m.nextDir
	m = changeDir(m, DirDown)
	if m.nextDir != before {
		t.Error("should block reversal from up to down")
	}
}

// ── moveEvery ────────────────────────────────────────────────────────────────

func TestMoveEvery_DecreasesWithScore(t *testing.T) {
	prev := moveEvery(0)
	for score := 5; score <= 30; score += 5 {
		cur := moveEvery(score)
		if cur > prev {
			t.Errorf("score %d: interval %d should not be greater than prev %d", score, cur, prev)
		}
		prev = cur
	}
}

func TestMoveEvery_HasMinimum(t *testing.T) {
	if moveEvery(9999) < 2 {
		t.Error("moveEvery should have a minimum of 2")
	}
}

// ── placeFood ────────────────────────────────────────────────────────────────

func TestPlaceFood_AvoidsSnake(t *testing.T) {
	m := newGame()
	// fill every cell except (0,0) with snake body
	var body []pos
	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			if x == 0 && y == 0 {
				continue
			}
			body = append(body, pos{x, y})
		}
	}
	m.snake = body
	m = placeFood(m)
	if m.food != (pos{0, 0}) {
		t.Errorf("food = %v, want {0,0} (only empty cell)", m.food)
	}
}

// ── tickGame ─────────────────────────────────────────────────────────────────

func TestTickGame_NoopWhenNotPlaying(t *testing.T) {
	m := newGame()
	m.state = StateGameOver
	before := m.tick
	m = tickGame(m)
	if m.tick != before {
		t.Error("tickGame should be noop when not StatePlaying")
	}
}

func TestTickGame_IncrementsTick(t *testing.T) {
	m := newGame()
	m = tickGame(m)
	if m.tick != 1 {
		t.Errorf("tick = %d, want 1", m.tick)
	}
}

func TestTickGame_MovesSnakeOnInterval(t *testing.T) {
	m := newGame()
	m.food = pos{0, 0}
	head := m.snake[0]
	interval := moveEvery(0)
	for i := 0; i < interval; i++ {
		m = tickGame(m)
	}
	if m.snake[0] == head {
		t.Error("snake should have moved after one full interval")
	}
}

func TestTickGame_AppliesBufferedDir(t *testing.T) {
	m := newGame() // moving right
	m.food = pos{0, 0}
	m = changeDir(m, DirUp)
	interval := moveEvery(0)
	for i := 0; i < interval; i++ {
		m = tickGame(m)
	}
	if m.dir != DirUp {
		t.Errorf("dir = %v, want DirUp after buffered change applied", m.dir)
	}
}
