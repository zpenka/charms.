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
- `2` — vs Computer, then choose your colour (`W` or `B`)

**Controls:**

| Key | Action |
|-----|--------|
| `↑↓←→` / `hjkl` | Move cursor |
| `Enter` / `Space` | Select piece / confirm move |
| `Esc` | Cancel selection |
| `q` | Quit |

Valid move destinations are highlighted on the board. The computer opponent uses a minimax engine with alpha-beta pruning and auto-promotes pawns to queens.
