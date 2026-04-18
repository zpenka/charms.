# charms.

A collection of [Charm](https://charm.sh) TUI experiments built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Running

```
go run .
```

A lobby appears where you can pick which game to play. Press `q` or `Ctrl+C` to quit.

## Testing

Run all tests across every game:

```
go test ./...
```

Run tests for a specific game:

```
go test ./chess/...
```

## Games

### Chess

A fully playable chess game in your terminal — two player or vs a computer opponent.

On launch, choose your mode:

- `1` — Two player (pass and play)
- `2` — vs Computer

Then choose your time control:

- `1` — Bullet (1 min per side)
- `2` — Blitz (5 min per side)
- `3` — Rapid (10 min per side)
- `4` — Classical (30 min per side)

When playing vs Computer, you will then choose a difficulty level:

- `1` — Easy (search depth 2 — faster, weaker)
- `2` — Medium (search depth 3 — balanced)
- `3` — Hard (search depth 4 — strongest)

After the time control, choose a board color scheme:

- `1` — Classic (wood tones)
- `2` — Ocean (blue)
- `3` — Mint (green)
- `4` — Dusk (purple)

Then choose your colour (`W` for White or `B` for Black).

**Controls:**

| Key | Action |
|-----|--------|
| `↑↓←→` / `hjkl` | Move cursor |
| `Enter` / `Space` | Select piece / confirm move |
| `Esc` | Cancel selection |
| `f` | Flip board perspective |
| `?` | Hint (highlight engine's suggested move) |
| `r` | Resign |
| `q` | Quit to lobby |

Valid move destinations are highlighted on the board. After every move, the from- and to-squares are tinted in amber. When the active king is in check, its square is highlighted red. The board auto-flips when you play as Black; press `f` at any time to toggle. Each player's clock counts down on their turn; clocks are displayed below the board. Move history is shown in algebraic notation below the board. The computer opponent uses a minimax engine with alpha-beta pruning, piece-square positional tables, and capture-first move ordering. Search depth is determined by the selected difficulty level (2–4 ply).

**Pawn promotion:** when you move a pawn to the back rank, a picker appears — press `Q`, `R`, `B`, or `N` to choose. The computer always promotes to a queen automatically.

---

### Tapper

A terminal take on the classic 1983 arcade game. Slide beer mugs down four bar lanes to serve customers walking in from the right. Miss a customer and your mug falls off the end — lose a life. Let a customer reach the bar — lose a life. Three lives per game, eight waves to start.

**Controls:**

| Key | Action |
|-----|--------|
| `↑↓` / `jk` | Move bartender between lanes |
| `Space` / `Enter` | Tap (fire a mug) |
| `p` | Pause / unpause |
| `q` | Quit to lobby |

One mug per lane at a time — but once the first mug passes halfway, you can fire a second on the same lane. Customers tint green → yellow → red as they approach the bar. A `*` flashes at the delivery point on each successful serve. Later customers within a wave spawn progressively faster, ratcheting up pressure as the wave goes on. The HUD shows how many customers remain (queued + on-screen). Losing a life triggers a brief red flash that freezes the action.

After game over, scores are saved to `~/.local/share/charms/tapper_scores.json` and a leaderboard shows your top 5 all-time scores with the current run highlighted.
