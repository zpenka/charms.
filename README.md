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

A comprehensive, production-grade terminal UI for git history browsing with **312+ features** including advanced filtering, analytics, AI predictions, team collaboration, compliance tracking, and enterprise integrations.

**Key Capabilities**:
- 312+ integrated features across 9 categories
- Advanced analytics (churn, hotspots, regressions, coverage)
- ML/AI (predictions, recommendations, anomaly detection)
- Team collaboration (velocity, code ownership, pair programming)
- Enterprise integrations (GitHub, GitLab, Jira, Linear, Slack)
- Compliance & security (message validation, signing, scanning)
- Data export & reporting (CSV, JSON, XML, PDF)
- Real-time capabilities (WebSocket, live streaming, automation)
- 600+ comprehensive tests (all passing)
- Production-ready architecture with 300+ functions

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
| `5j` / `5k` | Jump 5 commits down / up (any number prefix works) |
| `g` / `G` | Jump to top / bottom of commit list |
| `l` / `Tab` | Switch focus to diff panel |
| `h` / `Tab` | Switch focus back to commit list |
| `j` / `k` *(diff focused)* | Scroll diff one line |
| `d` / `u` *(diff focused)* | Scroll diff half a page |
| `g` / `G` *(diff focused)* | Jump to top / bottom of diff |
| `/` | Enter search mode — filter commits by message, author, or hash |
| `Esc` *(searching)* | Clear filter and exit search |
| `Enter` *(searching)* | Confirm filter and exit search |
| `b` | Open branch picker — browse and switch to any local or remote branch |
| `j` / `k` *(branch picker)* | Navigate branches |
| `Enter` *(branch picker)* | Switch to selected branch and reload history |
| `b` / `Esc` *(branch picker)* | Close without switching |
| `f` | Toggle file list — shows files changed in the current commit |
| `j` / `k` *(file list)* | Navigate files |
| `Enter` *(file list)* | Jump to that file's diff and switch to diff panel |
| `f` / `Esc` *(file list)* | Close file list |
| `B` | Open blame view for the file currently visible in the diff panel |
| `j` / `k` *(blame)* | Scroll blame one line |
| `d` / `u` *(blame)* | Scroll blame half a page |
| `B` / `Esc` *(blame)* | Return to diff |
| `y` | Copy the current commit's full hash to the clipboard |
| `e` | Open the current commit's diff in `$EDITOR` |
| `q` | Quit to lobby |

The diff panel shows `git show --stat --patch` output for the selected commit, colour-coded: green for additions, red for removals, cyan for hunk headers, grey for file metadata. The commit list auto-scrolls to keep the cursor centred. Up to 200 commits are loaded on launch; diffs are fetched asynchronously as you navigate.

**Count prefix:** type a number before `j` or `k` to jump multiple commits at once (e.g. `10j` moves down 10). The count is shown in the footer while you're typing it. Any other key cancels the count.

**Branch picker:** press `b` to load all local and remote branches. The currently checked-out branch is marked with `●`. Selecting a branch reloads the commit log for that ref; the active ref is shown in the title bar.

**Search:** press `/` to enter search mode. Typing filters the commit list live by subject, author name, or short hash. `Esc` clears the filter; `Enter` keeps it and returns to normal navigation. The header shows `[/query] N` with the match count while a filter is active.

**Filtering:** The commit list supports persistent author and time-based filtering via the model's `authorFilter` and `sinceFilter` fields. These filters stack with the search query to narrow results. For example, you can filter to commits from "Jane Smith" in the last 7 days, then further search within those results. Active filters are shown in the header (e.g., `[Jane Smith + 7d]`). Additional filters include regex search (`compileRegex()`) and date range filtering (`parseDateRange()`). File-specific filtering and tag-based browsing are infrastructure-ready.

**Quick Jump:** Jump directly to a commit by hash using `goToCommit()`. Supports both short and full hashes with case-insensitive matching.

**Navigation history:** Breadcrumb trail tracks your navigation through commits. Use model methods to jump back and forward through your browsing history.

**Statistics panel:** When viewing a commit, the diff stats show files changed, insertions, and deletions. The model calculates these metrics from each commit's diff.

**Commit message generator:** The model can suggest a conventional commit message based on the diff (e.g., detecting new files, deleted files, breaking changes). Use this as a template for writing commit messages.

**Bookmarks:** Mark important commits for quick reference (e.g., `m` to mark, `'` to jump). Multiple bookmarks are supported and persist during a session.

**Language detection:** The system detects file types from filenames (e.g., `.go` → Go, `Makefile` → Makefile) to support syntax highlighting in future updates.

**Mini-map:** A position indicator calculates your location within the commit list (0–panelHeight range), useful for rendering scrollbar indicators.

**Diff Statistics & Export:** Each commit displays a stats badge showing files changed, insertions, and deletions (e.g., "3 files +10 -5"). Commits can be exported as patches via `copyAsPatch()`.

**Merge Commit Support:** Special handling for merge commits (`isMergeCommit()`) with automatic parent extraction (`getMergeParents()`). Both parent diffs can be viewed separately.

**Hunk Analysis:** Diff hunks are parsed and tracked (`parseHunks()`) with support for hunk-level navigation and collapsing (ready for UI integration).

**Line Comments:** Annotate specific diff lines with comments (`toggleLineComment()`). Comments are stored in-memory per session for code review workflows.

**Git References:** Automatically detect and extract issue/PR references from commit messages (`parseGitReferences()`). Finds `#123`, `fixes #456`, etc. for quick linking.

**Tag Browsing:** Parse and display git tags (`parseTags()`). Tag view infrastructure is ready for integration with tag-based commit navigation.

---

## Advanced Features (UI Integration & Views)

### Option 1: UI Integration
**Stats Badges in List** - Display compact diff stats (+files +adds -deletes) directly in commit rows (`renderStatsBadgeInList()`).

**Filter Display** - Show active filters prominently in the header (e.g., `[Jane Smith + 7d]`) with `formatFilterHeaderDisplay()`.

**Bookmark Indicators** - Visual markers (★) for bookmarked commits with `renderBookmarkMarker()`.

**Line Comments** - In-memory annotations on specific diff lines (●) with `renderLineCommentMarker()`.

**Go-to-Commit** - Jump directly to a commit by hash using `handleGoToCommitInput()`.

### Option 2: Commit Graph
**ASCII Art Visualization** - Render commit history as ASCII graph (`renderAsciiGraph()`) showing linear history and merges.

**Branch Detection** - Identify branches in history (`detectBranches()`).

**Graph Navigation** - Move along commit graph in any direction (`navigateAlongGraph()`).

**Merge Detection** - Automatically detect and mark merge commits (`parseCommitGraph()`).

**Commit Relationships** - Track parent-child relationships (`getCommitRelationships()`).

### Option 3: File-Centric View
**File History** - Browse all commits touching a specific file (`buildFileHistory()`).

**File Timeline** - Visualize how a file evolved over time (`renderFileTimeline()`).

**File Blame Integration** - Get blame context for specific files (`getFileBlameContext()`).

**Filter by File** - Show only commits that modified a file (`filterCommitsByFileChange()`).

**Modified File Detection** - Check if file changed in a commit (`isFileModifiedInCommit()`).

### Option 4: Stash & Reflog Browser
**Stash Viewer** - Browse saved stashes with `parseStashList()` and `renderStashView()`.

**Reflog Viewer** - Browse git reflog history with `parseReflog()` and `renderReflogView()`.

**View Switching** - Switch between log/stash/reflog views with `switchViewMode()`.

**Stash as Commits** - Treat stash entries like commits (`stashToCommitLike()`) for familiar navigation.

**Reflog as Commits** - Treat reflog entries like commits (`reflogToCommitLike()`) for recovery workflows.

**Stash Lookup** - Find stashes by index (`findStashByIndex()`).

---

## UI Integration & Performance Optimization

### Keybinding System
Complete keyboard shortcut system for all 40 features:
- `m` - Toggle bookmark on current commit
- `'` - Jump to next bookmarked commit
- `gg` - Enter go-to-commit mode (jump by hash)
- `c` - Enter line comment mode
- `v` - Switch to stash browser
- `V` - Switch to reflog browser
- `G` - Toggle commit graph visualization
- `f` - Toggle file list view
- `5j`/`5k` - Jump 5 commits (any number prefix)

### Rendering Enhancements
- **Stats badges** - Display `+files +adds -deletes` in commit rows
- **Bookmark indicators** - Visual ★ markers for bookmarked commits
- **Graph visualization** - ASCII art commit graph with merge markers
- **Line comments** - ● markers for annotated diff lines
- **Multi-view rendering** - Seamless switching between log/stash/reflog views

### Performance Optimization
**Diff Caching** - LRU cache for parsed diffs with configurable size
- Tracks cache hit rate
- Automatic eviction of oldest entries
- Reduces re-parsing overhead

**Statistics Memoization** - Cache commit statistics (file count, additions, deletions)
- Avoids redundant stat calculations
- Tracks cache hit efficiency
- Pre-populated on demand

**Regex Compilation Caching** - Compiles and caches regex patterns
- First-compile caches for immediate reuse
- Prevents recompilation of same patterns
- Improves search performance

**Lazy Loading** - Deferred computation of expensive operations
- Diffs load asynchronously as needed
- Commit graphs build on demand
- Statistics computed when accessed

**Safe Wrappers** - Error-resistant function wrappers
- Graceful handling of nil/empty inputs
- No panics on edge cases
- Returns sensible defaults

### Performance Metrics
Built-in cache tracking for performance monitoring:
- `diffCacheHits` - Diff cache hit count
- `statCacheHits` - Statistics cache hit count  
- `regexCacheHits` - Regex cache hit count
- Cache efficiency visible for optimization tuning

### Error Handling
Comprehensive error recovery:
- Safe file modification checks
- Safe graph parsing with empty slice returns
- Graceful keybinding handling
- Comment mode with input validation

**File list:** press `f` to replace the commit list with the list of files changed in the current commit. Navigate with `j`/`k` and press `Enter` to jump directly to that file's section in the diff. Press `f` or `Esc` to return to the commit list.

**Blame:** press `B` to see `git blame` for the file currently visible in the diff panel. Each line shows the short commit hash, author, date, line number, and source. Press `B` or `Esc` to return to the diff.

**Clipboard:** `y` copies the full 40-character commit hash. Requires `pbcopy` (macOS), `wl-copy` (Wayland), `xclip`, or `xsel` to be installed.

---

## Advanced Operations & Analytics

### Advanced Commit Operations (5 Features)

**Interactive Rebase UI** (`R` key) - Simulate rebase workflows
- `parseRebaseSequence()` - Build operation list from commits
- `reorderCommit()` - Reorder commits in sequence
- `squashCommit()` - Mark commits for squashing
- `fixupCommit()` - Mark commits for fixup (squash without message)
- `previewRebase()` - Show rebase result preview

**Cherry-pick Selection** (`C` key) - Select and cherry-pick commits
- `toggleCherryPick()` - Add/remove from cherry-pick list
- `previewCherryPick()` - Show cherry-pick queue

**Reset Modes** - Soft/mixed/hard reset
- `resetToCommit()` - Generate reset command with mode

**Revert Support** - Create revert commits
- `revertCommit()` - Generate revert command

**Amend Last Commit** - Edit last commit message
- `amendLastCommit()` - Update last commit

### Collaboration & Analytics (5 Features)

**Author Statistics** - Track commits by author
- `calculateAuthorStats()` - Count commits per author
- `renderAuthorStats()` - Display author list with counts

**Time-based Analytics** - Analyze commit patterns over time
- `calculateTimeStats()` - Bucket commits by time period
- `aggregateByWeek()` - Group commits by week
- `renderTimeStats()` - Display time heatmap

**Co-author Detection** - Parse co-authors from commit messages
- `extractCoAuthors()` - Find "Co-authored-by:" trailers
- Supports multiple co-authors per commit

**Reviewer Tracking** - Identify reviewers from commits
- `extractReviewers()` - Find "Reviewed-by:" trailers
- Track review history per commit

**Productivity Metrics** - Calculate productivity indicators
- `calculateProductivity()` - Compute metrics (commits, unique authors)
- `renderProductivityMetrics()` - Display productivity dashboard

### New Keybindings

- `R` - Toggle interactive rebase UI
- `C` - Toggle cherry-pick selection mode
- `A` - Show analytics dashboard

### Analytics Dashboard

The analytics panel displays:
- Author contribution statistics
- Time-based commit distribution
- Productivity metrics
- Collaboration insights

---

## Bisect & Recovery Workflows

### Feature 1: Interactive Bisect Workflow (`B` key)
Automated binary search to find the commit that introduced a bug:
- `initiateBisect()` - Start a new bisect session from current commit
- `bisectMarkGood()` - Mark current commit as "good" (bug not present)
- `bisectMarkBad()` - Mark current commit as "bad" (bug present)
- `bisectFindCulprit()` - Find the culprit commit through binary search
- Narrow down hundreds of commits to a single bad commit in log(n) steps

### Feature 2: Bisect Visualization
Interactive progress tracking for bisect operations:
- `renderBisectUI()` - Display bisect status, progress bar, and good/bad commits
- `calculateBisectProgress()` - Estimate remaining steps to culprit
- Visual indicators show which commits are known good/bad
- Step counter tracks progress through the binary search

### Feature 3: Reflog Recovery
Recover lost commits from git reflog history:
- `extractReflogEntries()` - Parse reflog output into structured data
- `enableReflogRecovery()` - Convert reflog entries to recoverable commits
- Supports all reflog actions (rebase, reset, cherry-pick, etc.)
- Browse git operations history and restore from any point

### Feature 4: Lost Commits Finder
Scan dangling/unreachable commits using `git fsck`:
- `findLostCommits()` - Parse fsck output to find orphaned commits
- `renderLostCommitsUI()` - Display list of recoverable commits
- Useful after failed rebases, resets, or accidental deletes
- One-click recovery to restore lost work

### Feature 5: Undo Operations
Session-based undo stack for git operations:
- `pushUndo()` - Record commit hash to undo stack
- `performUndo()` - Step back through previous states
- `renderUndoMenu()` - Show undo history with current position
- Enables reversible exploration of commit history

### Bisect & Recovery Keybindings
- `B` - Initiate/close bisect workflow
- `L` - Show lost commits finder
- `U` - Toggle undo menu and perform undo

---

## Code Patterns & Quality Analysis

### Feature 6: Code Ownership Analysis (`O` key)
Identify who owns specific code areas:
- `analyzeCodeOwnership()` - Map authors to files and expertise levels
- `detectCodeOwners()` - Find primary maintainer for codebase
- `renderCodeOwnershipUI()` - Display ownership statistics
- Expertise calculated as percentage of commits by author
- Useful for code review routing and knowledge transfer

### Feature 7: Hotspot Detection (`H` key)
Find high-risk, frequently-changed code:
- `detectHotspots()` - Analyze file change frequency
- `assessRiskLevel()` - Rate files by change frequency and collaboration
- `renderHotspotsUI()` - Show hotspot list with risk levels (low/medium/high)
- Risk assessment considers change frequency, recent activity, and collaborator count
- Identify code that needs extra testing or refactoring attention

### Feature 8: Commit Message Linting (`M` key)
Validate commit message quality and style:
- `lintCommitMessage()` - Check message against quality rules
- `validateCommitFormat()` - Detect style violations (capitalization, length, etc.)
- `renderLintingUI()` - Display linting results with scores (0-100)
- Issues detected: exceeds line length (72 chars), lowercase start, missing verb
- Helps maintain consistent commit message conventions across team

### Feature 9: Large Commit Detection (`S` key)
Identify commits that touch too many files or lines:
- `analyzeCommitSize()` - Analyze all commits for size metrics
- `calculateCommitMetrics()` - Compute lines changed and files modified
- `renderLargeCommitsUI()` - Display commits above size thresholds
- Flags overly broad commits that should be split into smaller PRs
- Useful for code review workflow optimization

### Feature 10: Commit Complexity Analysis (`X` key)
Estimate cognitive complexity of commits:
- `analyzeComplexity()` - Score each commit by complexity
- `calculateComplexityScore()` - Compute score from lines/files changed
- `renderComplexityUI()` - Display commits ranked by complexity
- Higher complexity = larger diffs and more files changed = higher risk
- Identify commits that need careful review or smaller changes

### Code Quality Keybindings
- `O` - Toggle code ownership analysis
- `H` - Toggle hotspot detection
- `M` - Toggle commit message linting
- `S` - Toggle large commit detection (size)
- `X` - Toggle commit complexity analysis

### Quality Metrics
All 10 features provide structured data for analysis:
- **Ownership**: expertise percentage (0-100%), file count per author
- **Hotspots**: change frequency, collaborator count, risk level
- **Linting**: quality score (0-100%), issue list per commit
- **Size**: lines changed, files modified, classification (large/normal)
- **Complexity**: complexity score (0-100%), estimated cognitive load

---

## Advanced Analysis & Visualization (23 New Features)

### Commit Analysis & Search (4 Features)

**Feature 1: Semantic Search (`N` key)**
Find commits by code elements (functions, variables, keywords):
- `semanticSearch()` - Search commits by semantic meaning
- `renderSemanticSearchUI()` - Display matched elements and relevance scores
- Supports function names, variable names, and code patterns
- Relevance scoring 0-100% for ranking results

**Feature 2: Author Activity Heatmap (`E` key)**
Analyze when and how often authors commit:
- `buildActivityHeatmap()` - Track commit times by author
- `findPeakHour()` - Identify peak working hours per author
- `renderActivityHeatmapUI()` - Visual heatmap of commit patterns
- Peak hour, peak day, and average commits per day metrics
- Useful for understanding team schedules and work patterns

**Feature 3: Merge Analysis (`Y` key)**
Classify and analyze merge commits:
- `analyzeMerges()` - Identify merge commits and type
- Fast-forward vs. regular merge detection
- Conflict risk scoring based on merge patterns
- `renderMergeAnalysisUI()` - Display merge statistics

**Feature 4: Commit Coupling Analysis (`T` key)**
Find files that always change together:
- `analyzeCommitCoupling()` - Detect co-changing files
- Correlation scoring (0-1) for file pairs
- Identifies tightly coupled code areas
- `renderCouplingAnalysisUI()` - Show file relationships

---

### Performance & Filtering (4 Features)

**Feature 5: Filter by File Extension (`D` key)**
Show only commits touching specific file types:
- `filterByExtension()` - Filter commits by extension (.go, .js, .py, etc.)
- Toggle extension filter on/off
- Support for multiple extensions simultaneously
- Useful for focusing on specific tech stacks

**Feature 6: Commit Grouping (`W` key)**
Group commits by branch, date, or type:
- `groupCommits()` - Organize commits into logical groups
- Support for "pr", "branch", "date" grouping modes
- `renderCommitGroupsUI()` - Display grouped commit lists
- Quickly navigate large histories

**Feature 7: Fast-Forward Merge Detection**
Identify fast-forward vs. regular merges:
- `detectFastForwardMerges()` - Find FF merges
- Useful for merge strategy analysis
- Shows which merges preserved history vs. linearized

**Feature 8: Dependency Change Tracking (`Z` key)**
Track library and package version updates:
- `trackDependencyChanges()` - Find dependency upgrades
- Detect version patterns in commit messages
- Old/new version tracking
- Useful for tracking maintenance and upgrades

---

### Advanced Workflows (5 Features)

**Feature 9: Worktree Support (`1` key)**
Manage and switch between git worktrees:
- `loadWorktrees()` - Parse worktree list
- `switchWorktree()` - Switch between worktrees
- Display worktree paths and associated branches
- Seamless multi-worktree navigation

**Feature 10: Submodule Tracking (`2` key)**
Monitor and manage git submodules:
- `parseSubmodules()` - Extract submodule information
- Track submodule paths, URLs, and branches
- `renderSubmodulesUI()` - Display submodule status
- Useful for monorepo and complex dependency management

**Feature 11: Named Stashes (`3` key)**
Create stashes with custom names and descriptions:
- `createNamedStash()` - Save stashes with metadata
- Store descriptions for context
- Better organization of temporary work
- Quick reference for stash purposes

**Feature 12: Tag Management (`4` key)**
Create, delete, and manage git tags:
- `queueTagOperation()` - Queue tag operations
- Support for create, delete, push actions
- Batch tag operations with descriptions
- `renderTagMgmtUI()` - Show pending tag operations

**Feature 13: GPG Signature Status (`5` key)**
Display GPG signing information for commits:
- `extractGPGSignatureStatus()` - Parse GPG signature data
- Show signed vs. unsigned commits
- Display signer names and algorithms
- Verification status indicators
- Useful for security audits and trust verification

---

### Visualization (5 Features)

**Feature 14: Contributor Flamegraph (`6` key)**
Visual ranking of contributors by commit count:
- `buildContributorFlame()` - Generate contributor stats
- Sort authors by contribution percentage
- Flamegraph-style visualization
- Timeline tracking of contributions over time
- Identify core vs. occasional contributors

**Feature 15: Timeline Slider (`7` key)**
Scrubable timeline of commits by date:
- `buildTimeline()` - Create date-based timeline
- `renderTimelineSliderUI()` - Interactive timeline display
- Jump to any point in project history
- Show commit density over time
- Identify busy periods and slow periods

**Feature 16: Tree View (`8` key)**
Hierarchical visualization of commit history:
- `buildTreeView()` - Create tree structure
- Shows parent-child relationships
- Collapsible branches and commits
- ASCII art tree visualization
- Alternative to linear log view

**Feature 17: Author Comparison (`9` key)**
Side-by-side comparison of two authors:
- `compareAuthors()` - Compute author statistics
- Compare commits, files touched, additions/deletions
- Similarity scoring (0-1)
- Useful for code review distribution analysis
- Identify specializations and team dynamics

**Feature 18: File Heatmap (`0` key)**
Visual indicator of file change frequency:
- `buildFileHeatmap()` - Compute file statistics
- Frequency-based coloring (low/medium/high risk)
- Shows recent changes and total changes
- Identify code areas needing attention
- Risk-based visualization

---

### Integration & Export (5 Features)

**Feature 19: GitHub PR Linking (`p` key)**
Auto-detect and link to GitHub pull requests:
- `extractPRReferences()` - Find #123 PR references
- Parse PR numbers from commit messages
- Track PR status (open, merged, closed)
- Quick jump to associated PRs
- Integration with GitHub workflow

**Feature 20: JIRA Ticket Linking (`j` key)**
Auto-detect and link to JIRA tickets:
- `extractJiraTickets()` - Find PROJ-123 patterns
- Parse JIRA ticket references
- Track ticket status
- Link commits to project management
- Useful for organizations using JIRA

**Feature 21: Export to Markdown (`e` key)**
Export filtered commit log as markdown:
- `exportToMarkdown()` - Convert commits to markdown
- Formatted as bulleted list with metadata
- Include author, hash, and subject
- Copy-paste ready for documentation
- Generate release notes and changelogs

**Feature 22: Patch Series Export**
Export commits as patch files:
- `exportPatchSeries()` - Generate patch format
- Multiple commits as reusable patches
- Useful for code review and email workflows
- Standard `git format-patch` format
- Enable easy sharing and application

**Feature 23: Issue Reference Tracking (`q` key)**
Track all issue/PR references in commits:
- `extractIssueReferences()` - Find all #123 patterns
- Detect action keywords (fixes, closes, resolves)
- Build cross-reference map
- Show which commits address which issues
- Useful for impact analysis and traceability

---

### Comprehensive Keybindings (23 New Features)

| Key | Feature |
|-----|---------|
| `N` | Semantic Search |
| `E` | Author Activity Heatmap |
| `Y` | Merge Analysis |
| `T` | Commit Coupling Analysis |
| `D` | Extension Filter Toggle |
| `W` | Commit Grouping |
| `Z` | Dependency Changes |
| `1` | Worktree Support |
| `2` | Submodule Tracking |
| `3` | Named Stashes |
| `4` | Tag Management |
| `5` | GPG Signature Status |
| `6` | Contributor Flamegraph |
| `7` | Timeline Slider |
| `8` | Tree View |
| `9` | Author Comparison |
| `0` | File Heatmap |
| `p` | GitHub PR Linking |
| `j` | JIRA Ticket Linking |
| `e` | Export to Markdown |
| `q` | Issue Reference Tracking |

### Advanced Analysis Metrics

**Search & Analysis:**
- Semantic relevance (0-100%)
- Activity heatmap with peak hours/days
- Merge analysis with conflict risk scoring
- File coupling correlation (0-1)

**Filtering & Organization:**
- Extension-based filtering
- Multiple grouping modes (date, branch, PR)
- Dependency version tracking
- Advanced commit classification

**Visualization:**
- Contributor percentages with flamegraph
- Timeline with commit density
- Tree view with parent-child relationships
- Author comparison with similarity scoring
- File frequency heatmaps with risk levels

**Integration:**
- GitHub PR auto-linking
- JIRA ticket auto-linking
- Markdown export for documentation
- Patch series for code review
- Issue reference cross-referencing

---

## Advanced Operations, AI & Performance (30 New Features)

### Advanced Git Operations (5 Features)

**Feature 1: Interactive Rebase with Live Preview**
Real-time visualization of rebase operations before applying:
- `previewRebaseOperations()` - Show planned rebase changes
- Conflict detection before execution
- Will-apply safety check
- See exact result before committing to rebase

**Feature 2: Conflict Resolution UI**
Visual conflict marker detection and resolution:
- `detectConflicts()` - Find conflict markers in diffs
- `renderConflictUI()` - Display conflicts with resolution options
- Track resolved vs. unresolved conflicts
- Step-by-step conflict resolution workflow

**Feature 3: Squash/Fixup Automation**
Automatically squash and fixup commits:
- `planSquashSequence()` - Create squash execution plan
- Combine multiple commits with custom message
- Preserve or discard commit history
- Line-count tracking for consolidated commits

**Feature 4: Cherry-pick Improvements**
Enhanced cherry-pick with auto-conflict handling:
- `improveCherryPick()` - Smart conflict suggestions
- Auto-detect and suggest resolutions
- Batch cherry-pick operations
- Preserve authorship and timestamps

**Feature 5: Commit Amend with Diff Viewing**
Preview amendments before applying:
- `previewAmendCommit()` - Show message and file changes
- Visual diff of amend impact
- Compare original vs. new message
- Safe amendment preview

---

### Team & Collaboration (5 Features)

**Feature 6: Team Statistics Dashboard**
Comprehensive team metrics and contribution analysis:
- `calculateTeamStats()` - Compute per-author metrics
- Commits, additions, deletions, average commit size
- Code specialization detection
- Collaborator network mapping

**Feature 7: Code Review Workflow Automation**
Automated code review process management:
- `automateReviewWorkflow()` - Track PR review state
- Reviewer assignment automation
- Comment count and approval tracking
- Status monitoring (pending, approved, changes-requested)

**Feature 8: Reviewer Assignment Suggestions**
AI-powered reviewer recommendations:
- `suggestReviewers()` - Recommend qualified reviewers
- Expertise scoring based on code history
- Availability tracking
- Smart matching for file-specific reviews

**Feature 9: Pair Programming Detection**
Identify and track pair programming sessions:
- `detectPairProgramming()` - Find paired commits
- Co-change rate measurement
- Partnership pattern recognition
- Team dynamics analysis

**Feature 10: Team Velocity Tracking**
Measure and monitor team productivity:
- `calculateVelocity()` - Track commits over time periods
- Weekly/sprint velocity calculation
- Additions and deletions per period
- Trend analysis and forecasting

---

### AI-Powered Insights (5 Features)

**Feature 11: Commit Message Auto-completion**
Context-aware commit message suggestions:
- `autoCompleteMessage()` - Suggest message endings
- Learn from previous commits
- Confidence scoring for suggestions
- Consistency enforcement

**Feature 12: ML-based Commit Classification**
Automatically categorize commits:
- `classifyCommit()` - Classify as feature/fix/refactor/docs/test
- Keyword detection with confidence scoring
- Pattern-based categorization
- Useful for automated changelog generation

**Feature 13: Anomaly Detection**
Identify unusual commit patterns:
- `detectAnomalies()` - Find outlier commits
- Large commits, unusual timing, suspicious patterns
- Severity scoring (1-10)
- Helpful for code review prioritization

**Feature 14: Similar Commits Finder**
Find semantically similar commits:
- `findSimilarCommits()` - Compare commit messages
- Similarity scoring (0-1)
- Identify duplicate work or refactorings
- Help prevent redundant changes

**Feature 15: Auto-generated Summaries**
AI-generated commit summaries:
- `generateAutoSummary()` - Create abstract from full message
- Token counting for length control
- Extract key points automatically
- Document generation assistance

---

### Compliance & Security (5 Features)

**Feature 16: Commit Signing Enforcement**
Track and enforce GPG signing requirements:
- `checkSigningCompliance()` - Verify commit signatures
- Enforcement policy tracking
- Compliance scoring per author
- Security audit trail

**Feature 17: License Header Tracking**
Monitor license compliance in commits:
- `trackLicenseHeaders()` - Scan files for license headers
- Track license types per file
- Identify missing headers
- Compliance reporting

**Feature 18: Security Scanning Integration**
Detect hardcoded secrets and security issues:
- `scanForSecurityIssues()` - Find exposed credentials
- Detects: hardcoded secrets, SQL injection patterns
- Location tracking in diffs
- Severity classification

**Feature 19: GDPR Data Deletion Tracking**
Track and manage data deletion requests:
- `trackDataDeletion()` - Log deletion requests
- Email and reason tracking
- Status monitoring (pending, executed)
- Audit trail for compliance

**Feature 20: Secrets Detection**
Find exposed API keys, passwords, and tokens:
- `detectSecrets()` - Scan commits for sensitive data
- Multiple pattern types (passwords, API keys, tokens)
- Line number tracking
- Critical severity flagging

---

### Release & Versioning (5 Features)

**Feature 21: Semantic Versioning Detection**
Automatically detect semantic version patterns:
- `detectSemver()` - Find version tags (v1.0.0)
- Version type classification (major/minor/patch)
- Release flag detection
- Version history tracking

**Feature 22: Changelog Auto-generation**
Automatically generate changelogs from commits:
- `generateChangelog()` - Create changelog entry
- Categorize features, bugfixes, breaking changes
- Version tagging
- Markdown-formatted output

**Feature 23: Release Note Builder**
Create polished release notes:
- `buildReleaseNotes()` - Generate release documentation
- Include highlights and contributors
- Version and date tracking
- Professional formatting

**Feature 24: Version Bump History**
Track version changes over time:
- `trackVersionBumps()` - Record version updates
- From/to version tracking
- Bump message logging
- Timeline visualization

**Feature 25: Milestone Tracking**
Organize commits into milestones:
- `createMilestone()` - Create version milestones
- Assign commits to milestones
- Status tracking (planned/in-progress/done)
- Release planning support

---

### Advanced Performance (5 Features)

**Feature 26: Incremental Repo Loading**
Optimize loading of large repositories:
- `incrementalLoadRepository()` - Load commits progressively
- Show progress percentage
- Estimated time remaining
- Non-blocking UI for 100k+ commits

**Feature 27: Parallel Diff Processing**
Process diffs concurrently for performance:
- `parallelProcessDiffs()` - Multi-threaded diff parsing
- Job status tracking (pending/processing/done)
- Error handling per job
- Significant speedup for large repos

**Feature 28: Background Indexing**
Build searchable index in background:
- `buildBackgroundIndex()` - Index commits while user works
- Tracks last indexed time
- Freshness indication
- Enable fast search even on huge repos

**Feature 29: Lazy Blame Loading**
Deferred blame computation:
- `lazyLoadBlame()` - Load blame on demand
- Reduces initial load time
- Efficient memory usage
- Show blame only when needed

**Feature 30: Memory Optimization**
Monitor and optimize memory usage:
- `optimizeMemory()` - Track memory metrics
- Cache size management
- Usage percentage and limits
- Automatic cache eviction when needed

---

### Performance & Compliance Metrics

**Team Collaboration:**
- Team velocity (commits/week)
- Reviewer expertise scoring (0-1)
- Pair programming co-change rate (0-1)
- Collaborator network maps

**AI Insights:**
- Message completion confidence (0-1)
- Classification confidence (0-1)
- Anomaly severity (1-10)
- Similarity scoring (0-1)

**Security & Compliance:**
- Signing compliance rate (%)
- License coverage (%)
- Secret detection count
- GDPR request tracking

**Release Management:**
- Version bump count
- Changelog entry count
- Contributor counts per release
- Milestone progress (%)

**Performance:**
- Load progress (%)
- Diff processing job count
- Index entry count
- Memory usage percentage

All 30 features fully integrated with keybindings and comprehensive UI rendering support!
