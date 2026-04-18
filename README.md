# charms.

A collection of [Charm](https://charm.sh) TUI experiments built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Charms

### chess

A fully playable chess game in your terminal — two player or vs a computer opponent.

**Starting the game:**

```
cd chess
go run .
```

On launch, choose your mode:

- `1` — Two player (pass and play)
- `2` — vs Computer

When playing vs Computer, you will first choose a difficulty level:

- `1` — Easy (search depth 2 — faster, weaker)
- `2` — Medium (search depth 3 — balanced)
- `3` — Hard (search depth 4 — strongest)

Then choose your colour (`W` for White or `B` for Black).

**Controls:**

| Key | Action |
|-----|--------|
| `↑↓←→` / `hjkl` | Move cursor |
| `Enter` / `Space` | Select piece / confirm move |
| `Esc` | Cancel selection |
| `f` | Flip board perspective |
| `q` | Quit |

Valid move destinations are highlighted on the board. After every move, the from- and to-squares are tinted in amber. The board auto-flips when you play as Black; press `f` at any time to toggle. Move history is shown in algebraic notation below the board. The computer opponent uses a minimax engine with alpha-beta pruning, piece-square positional tables, and capture-first move ordering. Search depth is determined by the selected difficulty level (2–4 ply).

**Pawn promotion:** when you move a pawn to the back rank, a picker appears — press `Q`, `R`, `B`, or `N` to choose. The computer always promotes to a queen automatically.
