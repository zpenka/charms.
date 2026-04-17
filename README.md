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
- `2` — vs Computer (you play White)

**Controls:**

| Key | Action |
|-----|--------|
| `↑↓←→` / `hjkl` | Move cursor |
| `Enter` / `Space` | Select piece / confirm move |
| `Esc` | Cancel selection |
| `q` | Quit |

Valid move destinations are highlighted on the board. The computer opponent uses a depth-4 minimax engine with alpha-beta pruning, piece-square positional tables, and capture-first move ordering. Pawns auto-promote to queens.
