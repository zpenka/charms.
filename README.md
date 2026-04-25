# charms.

A collection of [Charm](https://charm.sh) TUI apps built with [Bubble Tea](https://github.com/charmbracelet/bubbletea): Chess, Tapper, Snake, 2048, and a Git Log browser.


<img width="636" height="347" alt="Screenshot 2026-04-20 at 10 16 26" src="https://github.com/user-attachments/assets/aeef5ba8-4d5e-492e-8ca5-f144f14de4c8" />


## Install

### Homebrew (macOS / Linux)

```
brew tap zpenka/tap
brew install charms
```

### Download a binary

Grab a pre-built binary from the [latest GitHub release](https://github.com/zpenka/charms./releases/latest), extract it, and put `charms` somewhere on your `$PATH`.

### Build from source

```
git clone https://github.com/zpenka/charms.
cd charms.
go build -o charms .
```

Requires Go 1.21+.

## Running

```
charms
```

A lobby appears showing all available games with a short description and your all-time best score for each. Press `q` or `Ctrl+C` to quit.

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


<img width="665" height="722" alt="Screenshot 2026-04-20 at 10 17 08" src="https://github.com/user-attachments/assets/7602dea0-2884-4341-825d-8b027be5e932" />


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
| `t` | Takeback (undo last move; undoes two moves when playing vs Computer) |
| `r` | Resign |
| `q` | Quit to lobby |

Valid move destinations are highlighted on the board. After every move, the from- and to-squares are tinted in amber. When the active king is in check, its square is highlighted red. The board auto-flips when you play as Black; press `f` at any time to toggle. Each player's clock counts down on their turn; clocks are displayed below the board. Move history is shown in algebraic notation below the board. The computer opponent uses a minimax engine with alpha-beta pruning, piece-square positional tables, and capture-first move ordering. Search depth is determined by the selected difficulty level (2–4 ply).

**Pawn promotion:** when you move a pawn to the back rank, a picker appears — press `Q`, `R`, `B`, or `N` to choose. The computer always promotes to a queen automatically.

**Captured pieces:** pieces taken by each side are listed below the clocks (e.g. `Captured by White: ♟ ♟`).

**Material score:** when one side is ahead in material, the HUD shows the advantage (e.g. `+3` for a knight up). Nothing is shown when material is equal.

**Opening name:** the HUD displays the detected opening name (e.g. "Italian Game", "Sicilian Defense") as long as the position matches a known opening line.

**PGN:** when the game ends, full PGN notation is shown below the board so you can review or copy the game.

---

### Tapper

A terminal take on the classic 1983 arcade game. Slide beer mugs down four bar lanes to serve customers walking in from the right. Miss a customer and your mug falls off the end — lose a life. Let a customer reach the bar — lose a life. Three lives per game.


<img width="543" height="430" alt="Screenshot 2026-04-20 at 10 17 24" src="https://github.com/user-attachments/assets/d0634c82-c73c-4934-85f9-b2406fd1c655" />



On launch, choose your mode:

- `1` — Waves (eight waves, then done)
- `2` — Endless (waves keep coming forever, no wave-clear screen)

**Controls:**

| Key | Action |
|-----|--------|
| `↑↓` / `jk` | Move bartender between lanes |
| `Space` / `Enter` | Tap (fire a mug) |
| `p` | Pause / unpause |
| `q` | Quit to lobby |

One mug per lane at a time — but once the first mug passes halfway, you can fire a second on the same lane. Customers tint green → yellow → red as they approach the bar. A `*` flashes at the delivery point on each successful serve. Later customers within a wave spawn progressively faster, ratcheting up pressure as the wave goes on. The HUD shows how many customers remain (queued + on-screen). Losing a life triggers a brief red flash that freezes the action.

**Special customers:** `!` Thirsty customers move at double speed and are worth 2× points. `$` VIP customers move slowly but are worth 5× points. `~` Slow-Mo customers move at normal speed, are worth 3× points, and trigger a 100-tick slow-motion effect when served — all customers advance at half speed while `SLOW MO` is shown in the HUD.

**Combo multiplier:** each consecutive serve without a miss increases your combo. Points scored = customer value × combo, so chaining serves across lanes pays off big. The active combo is shown in the HUD; it resets on any life lost.

**Double-tap bonus:** serving the same lane twice within 10 ticks doubles the points for the second serve.

**Extra life:** earn a heart back every 50 points, up to the 3-life maximum.

**Wave summary:** after each wave the clear screen shows your serve accuracy, best combo, and a wave bonus (combo×3, +20 for a perfect clear) added to your score.

After game over, scores are saved to `~/.local/share/charms/tapper_scores.json` and a leaderboard shows your top 5 all-time scores with the current run highlighted. The **best wave** reached across all runs is shown at the bottom of the leaderboard.

---

### Snake

The classic game. Eat food (`*`) to grow your snake. Don't hit obstacles or your own tail.


<img width="683" height="720" alt="Screenshot 2026-04-20 at 10 17 51" src="https://github.com/user-attachments/assets/80ff7f9c-7d43-4454-923f-9031050dc04b" />



**Controls:**

| Key | Action |
|-----|--------|
| `↑↓←→` / `wasd` / `hjkl` | Steer |
| `q` | Quit to lobby |

**Portal walls:** the edges of the board wrap around — exiting one side brings you out the other. No wall deaths.

**Obstacles:** each game spawns a set of `█` tiles scattered around the board. Running into one ends the game. A new obstacle is added for every 5 food items eaten, so the board gets progressively more dangerous.

**Bonus food:** a `$` tile appears every 30 ticks and expires after 40. Eating it scores 3 points and activates **Ghost mode** (`GHOST` shown in HUD) for 10 moves — in ghost mode the snake passes through obstacles without dying.

Speed increases as you grow. After game over, your length is saved to `~/.local/share/charms/snake_scores.json` and a leaderboard shows your top 5 runs.

---

### 2048

Slide all tiles in one direction with each keypress. Tiles with equal values merge into their sum. Reach your target tile to win — or keep going for a higher score.


<img width="583" height="445" alt="Screenshot 2026-04-20 at 10 18 15" src="https://github.com/user-attachments/assets/a9abf745-3e50-4324-875c-d13c0262ea17" />



**Target tile:** on the game start screen, choose your win condition:

- `1` — 512
- `2` — 1024
- `3` — 2048 (default)
- `4` — 4096

**Controls:**

| Key | Action |
|-----|--------|
| `↑↓←→` / `wasd` / `hjkl` | Slide tiles |
| `z` | Undo last move (one level) |
| `Space` | Continue after winning / confirm on end screens |
| `q` | Quit to lobby |

Each merge adds to your score (e.g. merging two 512s scores 1024). **Bonus tiles** occasionally spawn on the board — merging a bonus tile doubles the points scored for that merge (the bonus marker is consumed). The HUD shows your current score, highest tile on the board, and your **all-time best score** across all sessions. After game over, scores are saved to `~/.local/share/charms/2048_scores.json` and a leaderboard shows your top 5 runs.

---

### Git Log

A two-panel TUI for browsing the git history of the current repository — commit list on the left, coloured diff on the right.

**Layout:**

```
 git log  /path/to/repo

 Commits                    │ abc1234  Fix login bug
▶ abc1234  Fix login bug …  │ diff --git a/auth.go b/auth.go
  def5678  Add user model   │ --- a/auth.go
  xyz9876  Update README    │ +++ b/auth.go
  ...                       │ @@ -10,7 +10,9 @@
                            │ -   old code
                            │ +   new code
```

**Controls:**

| Key | Action |
|-----|--------|
| `j` / `↓` | Move to next commit |
| `k` / `↑` | Move to previous commit |
| `g` / `G` | Jump to top / bottom of commit list |
| `l` / `Tab` | Switch focus to diff panel |
| `h` / `Tab` | Switch focus back to commit list |
| `j` / `k` *(diff focused)* | Scroll diff one line |
| `d` / `u` *(diff focused)* | Scroll diff half a page |
| `g` / `G` *(diff focused)* | Jump to top / bottom of diff |
| `/` | Enter search mode — filter commits by message, author, or hash |
| `Esc` *(searching)* | Clear filter and exit search |
| `Enter` *(searching)* | Confirm filter and exit search |
| `f` | Toggle file list — shows files changed in the current commit |
| `j` / `k` *(file list)* | Navigate files |
| `Enter` *(file list)* | Jump to that file's diff and switch to diff panel |
| `f` / `Esc` *(file list)* | Close file list |
| `y` | Copy the current commit's full hash to the clipboard |
| `e` | Open the current commit's diff in `$EDITOR` |
| `q` | Quit to lobby |

The diff panel shows `git show --stat --patch` output for the selected commit, colour-coded: green for additions, red for removals, cyan for hunk headers, grey for file metadata. The commit list auto-scrolls to keep the cursor centred. Up to 200 commits are loaded on launch; diffs are fetched asynchronously as you navigate.

**Search:** press `/` to enter search mode. Typing filters the commit list live by subject, author name, or short hash. `Esc` clears the filter; `Enter` keeps it and returns to normal navigation. The header shows `[/query] N` with the match count while a filter is active.

**File list:** press `f` to replace the commit list with the list of files changed in the current commit. Navigate with `j`/`k` and press `Enter` to jump directly to that file's section in the diff. Press `f` or `Esc` to return to the commit list.

**Clipboard:** `y` copies the full 40-character commit hash. Requires `pbcopy` (macOS), `wl-copy` (Wayland), `xclip`, or `xsel` to be installed.
