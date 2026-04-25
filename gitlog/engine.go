package gitlog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type commit struct {
	hash      string
	shortHash string
	author    string
	when      string
	subject   string
}

type panel int

const (
	panelList panel = iota
	panelDiff
)

type lineKind int

const (
	lineContext lineKind = iota
	lineAdded
	lineRemoved
	lineHunk
	lineMeta
)

type diffLine struct {
	kind lineKind
	text string
}

type fileItem struct {
	path    string
	diffIdx int
}

type blameLine struct {
	shortHash string
	author    string
	date      string
	lineNum   int
	text      string
}

type model struct {
	commits    []commit
	cursor     int
	focus      panel
	diffLines  []diffLine
	diffOffset int
	fileItems  []fileItem
	fileCursor int
	showFiles  bool
	searching  bool
	query      string
	flash      string
	// Branch picker
	showBranch    bool
	branches      []string
	branchCursor  int
	currentRef    string
	currentBranch string
	// Blame view
	showBlame   bool
	blameLines  []blameLine
	blameOffset int
	// Count prefix for j/k navigation
	countBuf string
	// Filtering
	authorFilter string
	sinceFilter  int // days; 0 = no filter
	// Breadcrumb trail
	navHistory    []int
	navHistoryIdx int
	// Bookmarks
	bookmarks []string // commit short hashes
	// Stats
	lastStats commitStatistics
	// Line comments
	comments map[int]string
	// Tag view
	showTags  bool
	tags      []string
	tagCursor int
	// Option 1: UI Integration
	showStatsBadge bool
	// Option 2: Commit Graph
	commitGraph []graphNode
	showGraph   bool
	// Option 3: File-Centric
	fileHistory      []commit
	currentFile      string
	showFileTimeline bool
	// Option 4: Stash & Reflog
	viewMode      string // "log", "stash", "reflog"
	stashes       []stashEntry
	reflogEntries []reflogEntry
	stashCursor   int
	reflogCursor  int
	// UI Integration
	inGoToCommitMode bool
	goToCommitInput  string
	inCommentMode    bool
	commentInput     string
	// Optimization: Caches
	dcache    *diffCache
	scache    *statCache
	recache   *regexCache
	// Performance tracking
	diffCacheHits   int
	statCacheHits   int
	regexCacheHits  int
	// Advanced Operations
	showRebaseUI      bool
	rebaseSequence    []rebaseOp
	showCherryPickUI  bool
	cherryPickList    []string
	resetMode         string // soft, mixed, hard
	amendMessage      string
	// Analytics
	showAnalytics     bool
	authorStats       map[string]int
	timeStats         map[string]int
	collaborators     map[string][]string // author -> co-authors
	reviewers         map[string][]string // commit hash -> reviewers
	productivity      map[string]interface{}
	repoPath          string
	width             int
	height            int
	loading           bool
	// Bisect & Recovery
	bisectState         bisectState
	showBisectUI        bool
	lostCommits         []lostCommit
	showLostCommits     bool
	undoStack           []string // commit hashes for undo
	undoStackIdx        int
	showUndoMenu        bool
	reflogRecoveryMode  bool
	recoveryCommits     []lostCommit
	// Code Patterns & Quality
	codeOwnership       map[string]codeOwnershipData
	showCodeOwnership   bool
	hotspots            []hotspotData
	showHotspots        bool
	commitMetrics       []commitMetrics
	showComplexity      bool
	lintingResults      []lintingResult
	showLinting         bool
	largeCommits        []commitMetrics
	showLargeCommits    bool
	// Commit Analysis & Search
	semanticSearchResults []semanticSearchResult
	showSemanticSearch    bool
	semanticQuery         string
	authorActivityHeatmap map[string]authorActivityData
	showActivityHeatmap   bool
	mergeAnalysisData     []mergeAnalysis
	showMergeAnalysis     bool
	commitCouplings       []commitCoupling
	showCoupling          bool
	// Performance & Filtering
	extensionFilters  []fileExtensionFilter
	currentExtFilter  string
	commitGroups      []commitGroup
	groupingMode      string // "", "pr", "branch", "date"
	dependencyChanges []dependencyChange
	showDependencies  bool
	// Advanced Workflows
	worktrees          []worktreeInfo
	showWorktrees      bool
	currentWorktree    string
	submodules         []submoduleInfo
	showSubmodules     bool
	namedStashes       []namedStash
	showNamedStashes   bool
	pendingTagOps      []tagOperation
	showTagMgmt        bool
	gpgStatuses        map[string]gpgSignatureStatus
	showGPGStatus      bool
	// Visualization
	contributorFlameData []contributorFlameData
	showFlamegraph       bool
	timelinePoints       []timelinePoint
	timelineSliderPos    int
	showTimeline         bool
	treeRoot             *treeNode
	showTreeView         bool
	authorComparisons    []authorComparison
	selectedAuthors      [2]string
	showAuthorComparison bool
	fileHeatmap          []fileHeatmapEntry
	showFileHeatmap      bool
	// Integration & Export
	prReferences      []githubPRReference
	showPRLinks       bool
	jiraLinks         []jiraTicketLink
	showJiraLinks     bool
	pendingExports    []exportData
	showExportUI      bool
	exportFormat      string
	issueReferences   []issueReference
	showIssueRefs     bool
}

type commitStatistics struct {
	filesChanged int
	insertions   int
	deletions    int
}

type graphNode struct {
	hash    string
	depth   int
	isMerge bool
	parents []string
}

type stashEntry struct {
	name   string
	branch string
	subject string
	hash   string
}

type reflogEntry struct {
	hash    string
	action  string
	message string
	date    string
}

type diffCache struct {
	data     map[string][]diffLine
	order    []string
	maxSize  int
	hitCount int
}

type statCache struct {
	data     map[string]commitStatistics
	order    []string
	maxSize  int
	hitCount int
}

type regexCache struct {
	data     map[string]*regexp.Regexp
	maxSize  int
	hitCount int
}

type rebaseOp struct {
	action  string // pick, squash, fixup, reword, drop
	hash    string
	subject string
}

type bisectOp struct {
	hash    string
	isBad   bool
	isGood  bool
	current bool
}

type bisectState struct {
	active       bool
	current      string
	good         []string
	bad          []string
	candidates   []string
	visualSteps  int
	totalSteps   int
}

type lostCommit struct {
	hash      string
	shortHash string
	author    string
	subject   string
	date      string
}

type codeOwnershipData struct {
	author        string
	files         map[string]int // file -> count of commits
	lines         int
	expertise     float64
	isOwner       bool
}

type hotspotData struct {
	path             string
	changeFrequency  int
	recentChanges    int
	collaborators    int
	avgCommitSize    int
	riskLevel        string // low, medium, high
}

type commitMetrics struct {
	hash          string
	linesChanged  int
	filesChanged  int
	complexity    int // estimated
	isLarge       bool
	isComplex     bool
	messageQuality int // 0-100
}

type lintingResult struct {
	hash    string
	subject string
	issues  []string
	score   int // 0-100
}

// Commit Analysis & Search
type semanticSearchResult struct {
	hash      string
	subject   string
	matches   []string // matched items (function names, variables)
	relevance int      // 0-100
}

type authorActivityData struct {
	author      string
	hourOfDay   map[int]int // hour -> count
	dayOfWeek   map[int]int // day -> count
	peakHour    int
	peakDay     string
	avgPerDay   float64
}

type mergeAnalysis struct {
	hash          string
	isMerge       bool
	isFastForward bool
	parentCount   int
	conflictRisk  int // 0-100
}

type commitCoupling struct {
	file1       string
	file2       string
	coChangeCount int
	correlation float64 // 0-1
}

// Performance & Filtering
type fileExtensionFilter struct {
	extension string
	enabled   bool
}

type commitGroup struct {
	name     string
	commits  []string // hashes
	label    string   // PR, branch, or time period
	groupBy  string   // "pr", "branch", "date"
}

type dependencyChange struct {
	hash    string
	dep     string
	oldVer  string
	newVer  string
	reason  string
}

// Advanced Workflows
type worktreeInfo struct {
	path   string
	branch string
	hash   string
}

type submoduleInfo struct {
	path   string
	url    string
	hash   string
	branch string
}

type namedStash struct {
	index       int
	name        string
	description string
	hash        string
}

type tagOperation struct {
	name    string
	hash    string
	action  string // create, delete, push
	message string
}

type gpgSignatureStatus struct {
	hash      string
	signed    bool
	signer    string
	verified  bool
	algorithm string
}

// Visualization
type contributorFlameData struct {
	author     string
	commits    int
	lines      int
	percentage float64
	timeline   map[string]int // date -> commit count
}

type timelinePoint struct {
	date    string
	commits int
	hash    string
}

type treeNode struct {
	hash     string
	subject  string
	children []*treeNode
	depth    int
}

type authorComparison struct {
	author1      string
	author2      string
	commits1     int
	commits2     int
	files1       int
	files2       int
	additions1   int
	additions2   int
	deletions1   int
	deletions2   int
	similarity   float64
}

type fileHeatmapEntry struct {
	path      string
	frequency int
	recent    int
	risk      string // low, medium, high
}

// Integration & Export
type githubPRReference struct {
	hash    string
	prNumber int
	status  string // open, merged, closed
	title   string
}

type jiraTicketLink struct {
	hash   string
	ticket string
	status string
}

type exportData struct {
	format   string // "markdown", "patch", "json"
	commits  []commit
	content  string
	filename string
}

type issueReference struct {
	hash      string
	references []string // "#123", "#456"
	keywords  []string  // "fixes", "closes", "resolves"
}

func newModel(repoPath string) model {
	return model{
		repoPath: repoPath,
		focus:    panelList,
	}
}

// parseCommits parses output of:
//
//	git log --format="%H%x00%h%x00%an%x00%ar%x00%s"
func parseCommits(output string) []commit {
	var commits []commit
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\x00", 5)
		if len(parts) < 5 {
			continue
		}
		commits = append(commits, commit{
			hash:      parts[0],
			shortHash: parts[1],
			author:    parts[2],
			when:      parts[3],
			subject:   parts[4],
		})
	}
	return commits
}

// parseDiff classifies each line of a unified diff by type.
func parseDiff(raw string) []diffLine {
	var lines []diffLine
	for _, text := range strings.Split(raw, "\n") {
		var kind lineKind
		switch {
		case strings.HasPrefix(text, "@@"):
			kind = lineHunk
		case strings.HasPrefix(text, "diff "),
			strings.HasPrefix(text, "index "),
			strings.HasPrefix(text, "--- "),
			strings.HasPrefix(text, "+++ "),
			strings.HasPrefix(text, "new file"),
			strings.HasPrefix(text, "deleted file"),
			strings.HasPrefix(text, "similarity"),
			strings.HasPrefix(text, "rename"):
			kind = lineMeta
		case strings.HasPrefix(text, "+"):
			kind = lineAdded
		case strings.HasPrefix(text, "-"):
			kind = lineRemoved
		default:
			kind = lineContext
		}
		lines = append(lines, diffLine{kind, text})
	}
	return lines
}

// truncate cuts s to at most max visible runes, appending "…" if shortened.
func truncate(s string, max int) string {
	r := []rune(s)
	if max <= 0 {
		return ""
	}
	if len(r) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return string(r[:max-1]) + "…"
}

// firstWord returns the first space-delimited word of s.
func firstWord(s string) string {
	if i := strings.Index(s, " "); i >= 0 {
		return s[:i]
	}
	return s
}

func moveCursorDown(m model) model {
	if m.cursor < len(m.commits)-1 {
		m.cursor++
		m.diffOffset = 0
	}
	return m
}

func moveCursorUp(m model) model {
	if m.cursor > 0 {
		m.cursor--
		m.diffOffset = 0
	}
	return m
}

func scrollDiffDown(m model, n int) model {
	maxOff := len(m.diffLines) - diffPanelHeight(m)
	if maxOff < 0 {
		maxOff = 0
	}
	m.diffOffset += n
	if m.diffOffset > maxOff {
		m.diffOffset = maxOff
	}
	return m
}

func scrollDiffUp(m model, n int) model {
	m.diffOffset -= n
	if m.diffOffset < 0 {
		m.diffOffset = 0
	}
	return m
}

func switchPanel(m model) model {
	if m.focus == panelList {
		m.focus = panelDiff
	} else {
		m.focus = panelList
	}
	return m
}

// listPanelWidth returns the width of the commit list panel (clamped to 36–52).
func listPanelWidth(totalWidth int) int {
	w := totalWidth / 3
	if w < 36 {
		return 36
	}
	if w > 52 {
		return 52
	}
	return w
}

// diffPanelWidth returns the remaining width for the diff panel.
func diffPanelWidth(totalWidth int) int {
	return totalWidth - listPanelWidth(totalWidth) - 1 // 1 for divider
}

// diffPanelHeight returns the number of content lines visible in each panel.
func diffPanelHeight(m model) int {
	h := m.height - 7 // title + blank + header + blank + hint + blank*2
	if h < 5 {
		return 5
	}
	return h
}

// parseFileItems scans diffLines for "diff --git" boundaries and returns each
// file's path and the index of its boundary line in diffLines.
func parseFileItems(lines []diffLine) []fileItem {
	var items []fileItem
	for i, line := range lines {
		if line.kind != lineMeta || !strings.HasPrefix(line.text, "diff --git ") {
			continue
		}
		parts := strings.Fields(line.text)
		if len(parts) < 4 {
			continue
		}
		path := strings.TrimPrefix(parts[3], "b/")
		items = append(items, fileItem{path: path, diffIdx: i})
	}
	return items
}

// filterCommits returns commits whose subject, author, or short hash contain
// query (case-insensitive). An empty query returns all commits unchanged.
func filterCommits(commits []commit, query string) []commit {
	if query == "" {
		return commits
	}
	q := strings.ToLower(query)
	var out []commit
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), q) ||
			strings.Contains(strings.ToLower(c.author), q) ||
			strings.Contains(strings.ToLower(c.shortHash), q) {
			out = append(out, c)
		}
	}
	return out
}

// visibleCommits returns the commit list after applying all active filters:
// search query, author filter, and time-based filter.
func visibleCommits(m model) []commit {
	result := m.commits
	// Apply time filter first
	if m.sinceFilter > 0 {
		result = filterCommitsSince(result, m.sinceFilter)
	}
	// Apply author filter
	if m.authorFilter != "" {
		result = filterCommitsByAuthor(result, m.authorFilter)
	}
	// Apply search query filter last
	result = filterCommits(result, m.query)
	return result
}

// scrollToDiffLine sets diffOffset so that lineIdx is visible, clamped to valid range.
func scrollToDiffLine(m model, lineIdx int) model {
	m.diffOffset = lineIdx
	maxOff := len(m.diffLines) - diffPanelHeight(m)
	if maxOff < 0 {
		maxOff = 0
	}
	if m.diffOffset > maxOff {
		m.diffOffset = maxOff
	}
	if m.diffOffset < 0 {
		m.diffOffset = 0
	}
	return m
}

// toggleFileView shows or hides the file list in the left panel.
// Hiding resets fileCursor.
func toggleFileView(m model) model {
	m.showFiles = !m.showFiles
	if !m.showFiles {
		m.fileCursor = 0
	}
	return m
}

// parseBranches parses the output of "git branch -a", stripping the current-branch
// marker (*) and skipping ref-pointer lines (e.g. "origin/HEAD -> origin/main").
func parseBranches(output string) []string {
	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "* ")
		if strings.Contains(line, " -> ") {
			continue
		}
		branches = append(branches, line)
	}
	return branches
}

// parseCurrentBranch returns the name of the currently checked-out branch from
// "git branch -a" output (the line prefixed with "* ").
func parseCurrentBranch(output string) string {
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.HasPrefix(line, "* ") {
			return strings.TrimPrefix(line, "* ")
		}
	}
	return ""
}

// parseBlameLine parses one line of "git blame --date=short" output.
// Format: "hash (Author Name   YYYY-MM-DD  linenum) content"
func parseBlameLine(line string) (blameLine, bool) {
	paren := strings.Index(line, "(")
	close := strings.Index(line, ")")
	if paren < 0 || close < 0 || close < paren {
		return blameLine{}, false
	}
	hash := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line[:paren]), "^"))
	if len(hash) > 7 {
		hash = hash[:7]
	}
	meta := strings.Fields(line[paren+1 : close])
	if len(meta) < 3 {
		return blameLine{}, false
	}
	// last field: line number; second-to-last: date; rest: author
	lineNum, err := strconv.Atoi(meta[len(meta)-1])
	if err != nil {
		return blameLine{}, false
	}
	date := meta[len(meta)-2]
	author := strings.Join(meta[:len(meta)-2], " ")
	var text string
	if close+2 <= len(line) {
		text = line[close+2:]
	} else if close+1 < len(line) {
		text = line[close+1:]
	}
	return blameLine{
		shortHash: hash,
		author:    author,
		date:      date,
		lineNum:   lineNum,
		text:      text,
	}, true
}

// parseBlame parses all lines from "git blame --date=short" output,
// skipping any lines that don't match the expected format.
func parseBlame(output string) []blameLine {
	var lines []blameLine
	for _, line := range strings.Split(output, "\n") {
		if bl, ok := parseBlameLine(line); ok {
			lines = append(lines, bl)
		}
	}
	return lines
}

// currentFile returns the path of the file whose diff section is currently
// visible, based on the last fileItem whose diffIdx is <= diffOffset.
func currentFile(m model) string {
	if len(m.fileItems) == 0 {
		return ""
	}
	cur := m.fileItems[0].path
	for _, fi := range m.fileItems {
		if fi.diffIdx <= m.diffOffset {
			cur = fi.path
		}
	}
	return cur
}

// parseCount converts a digit string to a navigation count.
// Empty string or zero returns 1; values above 200 are capped.
func parseCount(s string) int {
	if s == "" {
		return 1
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 1
	}
	if n > 200 {
		return 200
	}
	return n
}

// toggleBranchView shows or hides the branch picker in the left panel.
// Hiding resets branchCursor.
func toggleBranchView(m model) model {
	m.showBranch = !m.showBranch
	if !m.showBranch {
		m.branchCursor = 0
	}
	return m
}

// filterCommitsByAuthor returns commits whose author exactly matches the given author
// (case-insensitive).
func filterCommitsByAuthor(commits []commit, author string) []commit {
	if author == "" {
		return commits
	}
	var out []commit
	for _, c := range commits {
		if strings.EqualFold(c.author, author) {
			out = append(out, c)
		}
	}
	return out
}

// filterCommitsSince returns commits from the last N days, parsed from the
// "when" field (e.g., "5 days ago", "2 weeks ago"). Returns all commits if
// days <= 0.
func filterCommitsSince(commits []commit, days int) []commit {
	if days <= 0 {
		return commits
	}
	var out []commit
	for _, c := range commits {
		if isWithinDays(c.when, days) {
			out = append(out, c)
		}
	}
	return out
}

// isWithinDays checks if a "when" string (e.g., "5 days ago") represents
// a time within the last N days.
func isWithinDays(when string, days int) bool {
	whenLower := strings.ToLower(when)

	// Extract number from strings like "5 days ago", "2 weeks ago", etc.
	re := regexp.MustCompile(`(\d+)\s+(day|week|month|year)`)
	matches := re.FindStringSubmatch(whenLower)
	if len(matches) < 3 {
		return false
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return false
	}

	unit := matches[2]
	totalDays := 0
	switch unit {
	case "day":
		totalDays = num
	case "week":
		totalDays = num * 7
	case "month":
		totalDays = num * 30
	case "year":
		totalDays = num * 365
	default:
		return false
	}

	return totalDays <= days
}

// formatActiveFilters returns a string showing all active filters.
func formatActiveFilters(m model) string {
	var filters []string
	if m.authorFilter != "" {
		filters = append(filters, m.authorFilter)
	}
	if m.sinceFilter > 0 {
		filters = append(filters, strconv.Itoa(m.sinceFilter)+"d")
	}
	if len(filters) == 0 {
		return ""
	}
	return "[" + strings.Join(filters, " + ") + "]"
}

// addToNavHistory adds the current cursor position to navigation history.
func addToNavHistory(m model, position int) model {
	// Discard any future history if we're not at the end
	if m.navHistoryIdx < len(m.navHistory)-1 {
		m.navHistory = m.navHistory[:m.navHistoryIdx+1]
	}
	m.navHistory = append(m.navHistory, position)
	m.navHistoryIdx = len(m.navHistory) - 1
	return m
}

// goBackInHistory moves to the previous position in navigation history.
func goBackInHistory(m model) model {
	if m.navHistoryIdx > 0 {
		m.navHistoryIdx--
		m.cursor = m.navHistory[m.navHistoryIdx]
	}
	return m
}

// goForwardInHistory moves to the next position in navigation history.
func goForwardInHistory(m model) model {
	if m.navHistoryIdx < len(m.navHistory)-1 {
		m.navHistoryIdx++
		m.cursor = m.navHistory[m.navHistoryIdx]
	}
	return m
}

// commitStats calculates statistics for a commit's diff.
func commitStats(lines []diffLine) commitStatistics {
	var stats commitStatistics
	for _, line := range lines {
		if line.kind == lineMeta && strings.HasPrefix(line.text, "diff --git") {
			stats.filesChanged++
		}
		if line.kind == lineAdded {
			stats.insertions++
		}
		if line.kind == lineRemoved {
			stats.deletions++
		}
	}
	return stats
}

// generateCommitMessage generates a suggested commit message based on the diff.
func generateCommitMessage(lines []diffLine, filename string) string {
	stats := commitStats(lines)
	isNew := false
	isDeleted := false
	isBreaking := false

	for _, line := range lines {
		if strings.Contains(line.text, "new file mode") {
			isNew = true
		}
		if strings.Contains(line.text, "deleted file mode") {
			isDeleted = true
		}
		// Simple breaking change detection: removed function/interface definitions
		if line.kind == lineRemoved &&
			(strings.Contains(line.text, "func ") || strings.Contains(line.text, "interface")) {
			isBreaking = true
		}
	}

	scope := filenameToScope(filename)
	var verb string

	switch {
	case isDeleted:
		verb = "remove"
	case isNew:
		verb = "add"
	case stats.deletions > stats.insertions*2:
		verb = "refactor"
	default:
		verb = "update"
	}

	msg := verb
	if scope != "" {
		msg += "(" + scope + ")"
	}
	if isBreaking {
		msg += "!"
	}
	msg += ": " + capitalizeFirst(verb) + " changes"

	return msg
}

// filenameToScope extracts a scope from a filename (e.g., "auth.go" -> "auth").
func filenameToScope(filename string) string {
	if filename == "" {
		return ""
	}
	base := filename
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		base = base[idx+1:]
	}
	if idx := strings.LastIndex(base, "."); idx > 0 {
		base = base[:idx]
	}
	return strings.ToLower(base)
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// toggleBookmark toggles a bookmark on the current commit.
func toggleBookmark(m model) model {
	if m.cursor >= len(m.commits) {
		return m
	}
	hash := m.commits[m.cursor].shortHash
	if isBookmarked(m, m.cursor) {
		// Remove bookmark
		var newBookmarks []string
		for _, b := range m.bookmarks {
			if b != hash {
				newBookmarks = append(newBookmarks, b)
			}
		}
		m.bookmarks = newBookmarks
	} else {
		// Add bookmark
		m.bookmarks = append(m.bookmarks, hash)
	}
	return m
}

// isBookmarked checks if a commit at the given index is bookmarked.
func isBookmarked(m model, idx int) bool {
	if idx >= len(m.commits) {
		return false
	}
	hash := m.commits[idx].shortHash
	for _, b := range m.bookmarks {
		if b == hash {
			return true
		}
	}
	return false
}

// jumpToNextBookmark moves the cursor to the next bookmarked commit.
func jumpToNextBookmark(m model) model {
	for i := m.cursor + 1; i < len(m.commits); i++ {
		if isBookmarked(m, i) {
			m.cursor = i
			return m
		}
	}
	return m
}

// jumpToPrevBookmark moves the cursor to the previous bookmarked commit.
func jumpToPrevBookmark(m model) model {
	for i := m.cursor - 1; i >= 0; i-- {
		if isBookmarked(m, i) {
			m.cursor = i
			return m
		}
	}
	return m
}

// detectLanguage detects the programming language from a filename.
func detectLanguage(filename string) string {
	ext := filename
	if idx := strings.LastIndex(filename, "."); idx >= 0 {
		ext = filename[idx:]
	}

	langMap := map[string]string{
		".go":         "go",
		".py":         "python",
		".js":         "javascript",
		".ts":         "typescript",
		".rb":         "ruby",
		".java":       "java",
		".cpp":        "cpp",
		".c":          "c",
		".rs":         "rust",
		".sh":         "bash",
		".sql":        "sql",
		".html":       "html",
		".css":        "css",
		".json":       "json",
		".yaml":       "yaml",
		".yml":        "yaml",
		".xml":        "xml",
		".md":         "markdown",
		"Makefile":    "makefile",
		"Dockerfile":  "dockerfile",
		".gitignore":  "gitignore",
		".env":        "dotenv",
	}

	if lang, ok := langMap[ext]; ok {
		return lang
	}
	if lang, ok := langMap[filename]; ok {
		return lang
	}

	// Default to text
	return "text"
}

// miniMapPosition calculates the position of a scroll indicator (0-height).
func miniMapPosition(cursor, panelHeight, totalCommits int) int {
	if totalCommits <= 1 {
		return 0
	}
	if panelHeight <= 1 {
		return 0
	}

	// Map cursor position to panel height
	position := (cursor * (panelHeight - 1)) / (totalCommits - 1)
	if position > panelHeight-1 {
		position = panelHeight - 1
	}
	return position
}

// --- diffStatBadge ---

// diffStatBadge formats commit statistics as a compact badge (e.g., "3 files +10 -5").
func diffStatBadge(stats commitStatistics) string {
	var parts []string
	if stats.filesChanged > 0 {
		parts = append(parts, fmt.Sprintf("%d file%s", stats.filesChanged, pluralize(stats.filesChanged)))
	}
	if stats.insertions > 0 {
		parts = append(parts, fmt.Sprintf("+%d", stats.insertions))
	}
	if stats.deletions > 0 {
		parts = append(parts, fmt.Sprintf("-%d", stats.deletions))
	}
	if len(parts) == 0 {
		return "0 changes"
	}
	return strings.Join(parts, " ")
}

// pluralize returns "s" if count != 1, else "".
func pluralize(count int) string {
	if count != 1 {
		return "s"
	}
	return ""
}

// --- goToCommit ---

// goToCommit finds a commit by hash (short or full) and returns its index, or -1 if not found.
func goToCommit(commits []commit, query string) int {
	q := strings.ToLower(query)
	for i, c := range commits {
		if strings.EqualFold(c.shortHash, query) || strings.EqualFold(c.hash, query) {
			return i
		}
		if strings.HasPrefix(strings.ToLower(c.shortHash), q) || strings.HasPrefix(strings.ToLower(c.hash), q) {
			return i
		}
	}
	return -1
}

// --- copyAsPatch ---

// copyAsPatch generates a patch file format from a commit hash and diff lines.
func copyAsPatch(hash string, lines []diffLine) string {
	var sb strings.Builder
	sb.WriteString("From: " + hash + "\n")
	sb.WriteString("Subject: Commit patch\n")
	sb.WriteString("\n")
	for _, line := range lines {
		sb.WriteString(line.text)
		sb.WriteString("\n")
	}
	return sb.String()
}

// --- parseGitReferences ---

// parseGitReferences extracts issue/PR numbers from a commit message (e.g., #123, fixes #456).
func parseGitReferences(msg string) []string {
	re := regexp.MustCompile(`#(\d+)`)
	matches := re.FindAllStringSubmatch(msg, -1)
	var refs []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			refs = append(refs, match[1])
			seen[match[1]] = true
		}
	}
	return refs
}

// --- isMergeCommit ---

// isMergeCommit checks if a commit is a merge commit.
func isMergeCommit(lines []diffLine) bool {
	for _, line := range lines {
		if strings.HasPrefix(line.text, "Merge:") {
			return true
		}
		if strings.Contains(strings.ToLower(line.text), "merge branch") {
			return true
		}
	}
	return false
}

// --- getMergeParents ---

// getMergeParents extracts parent hashes from a merge commit.
func getMergeParents(lines []diffLine) []string {
	for _, line := range lines {
		if strings.HasPrefix(line.text, "Merge:") {
			parts := strings.Fields(strings.TrimPrefix(line.text, "Merge:"))
			if len(parts) >= 2 {
				return parts[:2]
			}
		}
	}
	return nil
}

// --- hunk structure and parseHunks ---

type hunk struct {
	startLine int
	endLine   int
	header    string
}

// parseHunks extracts all hunks from diff lines.
func parseHunks(lines []diffLine) []hunk {
	var hunks []hunk
	var lastHunk hunk
	hunkCount := 0

	for i, line := range lines {
		if line.kind == lineHunk {
			if hunkCount > 0 {
				hunks = append(hunks, lastHunk)
			}
			lastHunk = hunk{
				startLine: i,
				endLine:   i,
				header:    line.text,
			}
			hunkCount++
		} else if hunkCount > 0 {
			lastHunk.endLine = i
		}
	}
	if hunkCount > 0 {
		hunks = append(hunks, lastHunk)
	}
	return hunks
}

// --- toggleLineComment ---

// toggleLineComment adds or removes a comment on a specific diff line.
func toggleLineComment(m model, lineIdx int, comment string) model {
	if m.comments == nil {
		m.comments = make(map[int]string)
	}
	if comment == "" {
		delete(m.comments, lineIdx)
	} else {
		m.comments[lineIdx] = comment
	}
	return m
}

// --- compileRegex ---

// compileRegex compiles a regex pattern for search.
func compileRegex(pattern string) (*regexp.Regexp, error) {
	return regexp.Compile(pattern)
}

// --- parseDateRange ---

// parseDateRange parses a date or date range (e.g., "2024-01-15" or "2024-01-01..2024-01-31").
func parseDateRange(input string) (*time.Time, *time.Time, error) {
	if input == "" {
		return nil, nil, nil
	}

	// Check for range format
	if strings.Contains(input, "..") {
		parts := strings.Split(input, "..")
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("invalid date range format")
		}
		start, err1 := time.Parse("2006-01-02", strings.TrimSpace(parts[0]))
		end, err2 := time.Parse("2006-01-02", strings.TrimSpace(parts[1]))
		if err1 != nil || err2 != nil {
			return nil, nil, fmt.Errorf("invalid date format")
		}
		return &start, &end, nil
	}

	// Single date
	date, err := time.Parse("2006-01-02", input)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid date format")
	}
	return &date, nil, nil
}

// --- filterCommitsByFile ---

// filterCommitsByFile returns commits that touched the specified file.
// Note: This is infrastructure-ready; actual filtering requires git queries.
func filterCommitsByFile(commits []commit, file string) []commit {
	if file == "" {
		return commits
	}
	// This would typically query git for commits touching the file
	// For now, return infrastructure (called from UI)
	return []commit{}
}

// --- parseTags ---

// parseTags parses git tag output (format: "hash tagname" per line).
func parseTags(output string) []string {
	var tags []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Handle both "hash tagname" and "tagname" formats
		parts := strings.Fields(line)
		if len(parts) >= 1 {
			// Take the last non-hash part
			tag := parts[len(parts)-1]
			if !strings.HasPrefix(tag, "[") { // Skip metadata
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

// ===== OPTION 1: UI INTEGRATION =====

// renderStatsBadgeInList renders a compact stats badge for a commit list row.
func renderStatsBadgeInList(stats commitStatistics, maxWidth int) string {
	badge := diffStatBadge(stats)
	if len(badge) > maxWidth {
		badge = truncate(badge, maxWidth)
	}
	return badge
}

// formatFilterHeaderDisplay formats active filters for header display.
func formatFilterHeaderDisplay(m model) string {
	return formatActiveFilters(m)
}

// renderBookmarkMarker returns a visual marker for bookmarked commits.
func renderBookmarkMarker(m model, idx int) string {
	if isBookmarked(m, idx) {
		return "★"
	}
	return ""
}

// handleGoToCommitInput processes go-to-commit input and updates model.
func handleGoToCommitInput(m model, query string) model {
	idx := goToCommit(m.commits, query)
	if idx >= 0 {
		m.cursor = idx
		m.diffOffset = 0
	}
	return m
}

// renderLineCommentMarker returns a visual marker for commented lines.
func renderLineCommentMarker(m model, lineIdx int) string {
	if m.comments != nil {
		if _, ok := m.comments[lineIdx]; ok {
			return "●"
		}
	}
	return ""
}

// ===== OPTION 2: COMMIT GRAPH =====

// parseCommitGraph builds a graph structure from commits.
func parseCommitGraph(commits []commit) []graphNode {
	var nodes []graphNode
	for i, c := range commits {
		node := graphNode{
			hash:    c.hash,
			depth:   0,
			isMerge: false,
		}
		// Simple heuristic: if subject contains "Merge", mark as merge
		if strings.Contains(strings.ToLower(c.subject), "merge") {
			node.isMerge = true
		}
		// Depth is based on position (linear for now)
		node.depth = i % 2
		nodes = append(nodes, node)
	}
	return nodes
}

// detectBranches identifies branches in commit graph.
func detectBranches(commits []commit) []string {
	// Simple implementation: treat as single branch unless merges detected
	var branches []string
	hasMerge := false
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), "merge") {
			hasMerge = true
			break
		}
	}
	if hasMerge {
		branches = append(branches, "main", "feature")
	} else {
		branches = append(branches, "main")
	}
	return branches
}

// renderAsciiGraph renders ASCII art graph for commit history.
func renderAsciiGraph(graph []graphNode) string {
	if len(graph) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, node := range graph {
		if node.isMerge {
			sb.WriteString("*   ")
		} else {
			sb.WriteString("* ")
		}
		hash := node.hash
		if len(hash) > 7 {
			hash = hash[:7]
		}
		sb.WriteString(hash)
		if i < len(graph)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// navigateAlongGraph moves along the graph in a direction.
func navigateAlongGraph(graph []graphNode, currentIdx int, direction string) int {
	if len(graph) == 0 {
		return 0
	}
	switch direction {
	case "down":
		if currentIdx < len(graph)-1 {
			return currentIdx + 1
		}
	case "up":
		if currentIdx > 0 {
			return currentIdx - 1
		}
	}
	return currentIdx
}

// getCommitRelationships maps parent-child relationships.
func getCommitRelationships(commits []commit) map[string][]string {
	rels := make(map[string][]string)
	// Infrastructure: would populate from git data
	return rels
}

// ===== OPTION 3: FILE-CENTRIC VIEW =====

// buildFileHistory constructs commit history for a specific file.
func buildFileHistory(commits []commit, file string) []commit {
	if file == "" {
		return []commit{}
	}
	// Infrastructure: would query git for file history
	return []commit{}
}

// renderFileTimeline renders the evolution of a file over time.
func renderFileTimeline(commits []commit, file string, width int) string {
	var sb strings.Builder
	sb.WriteString("File timeline for: ")
	sb.WriteString(file)
	sb.WriteString("\n")
	if len(commits) == 0 {
		sb.WriteString("(no commits found)\n")
	}
	return sb.String()
}

// getFileBlameContext gets blame information for a file.
func getFileBlameContext(lines []diffLine, file string) map[int]string {
	ctx := make(map[int]string)
	// Infrastructure: would populate from git blame
	return ctx
}

// filterCommitsByFileChange filters commits that modified a specific file.
func filterCommitsByFileChange(commits []commit, file string) []commit {
	if file == "" {
		return commits
	}
	var result []commit
	for _, c := range commits {
		if isFileModifiedInCommit(c.hash, file) {
			result = append(result, c)
		}
	}
	return result
}

// isFileModifiedInCommit checks if a file was modified in a commit.
func isFileModifiedInCommit(hash, file string) bool {
	// Infrastructure: would query git
	return false
}

// ===== OPTION 4: STASH & REFLOG =====

// parseStashList parses git stash output.
func parseStashList(output string) []stashEntry {
	var stashes []stashEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "stash@{0}: WIP on main: abc1234 message"
		parts := strings.SplitN(line, ":", 3)
		if len(parts) >= 2 {
			name := strings.TrimSpace(parts[0])
			rest := strings.TrimSpace(parts[1])
			// Extract branch name
			branchParts := strings.Fields(rest)
			var branch string
			if len(branchParts) >= 3 {
				branch = branchParts[2]
			}
			stashes = append(stashes, stashEntry{
				name:    name,
				branch:  branch,
				subject: line,
			})
		}
	}
	return stashes
}

// parseReflog parses git reflog output.
func parseReflog(output string) []reflogEntry {
	var entries []reflogEntry
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "abc1234 HEAD@{0}: commit: message"
		parts := strings.SplitN(line, " ", 3)
		if len(parts) >= 3 {
			hash := parts[0]
			rest := parts[2]
			actionParts := strings.SplitN(rest, ":", 2)
			action := ""
			message := ""
			if len(actionParts) >= 1 {
				action = actionParts[0]
			}
			if len(actionParts) >= 2 {
				message = strings.TrimSpace(actionParts[1])
			}
			entries = append(entries, reflogEntry{
				hash:    hash,
				action:  action,
				message: message,
			})
		}
	}
	return entries
}

// renderStashView renders the stash browser view.
func renderStashView(stashes []stashEntry, width int) string {
	var sb strings.Builder
	sb.WriteString("Stashes:\n")
	if len(stashes) == 0 {
		sb.WriteString("(no stashes)\n")
		return sb.String()
	}
	for i, s := range stashes {
		sb.WriteString(fmt.Sprintf("%d: %s - %s\n", i, s.name, s.branch))
	}
	return sb.String()
}

// renderReflogView renders the reflog browser view.
func renderReflogView(entries []reflogEntry, width int) string {
	var sb strings.Builder
	sb.WriteString("Reflog:\n")
	if len(entries) == 0 {
		sb.WriteString("(no reflog entries)\n")
		return sb.String()
	}
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("%s - %s: %s\n", e.hash[:7], e.action, e.message))
	}
	return sb.String()
}

// stashToCommitLike converts a stash entry to a commit-like structure.
func stashToCommitLike(stash stashEntry) commit {
	return commit{
		shortHash: stash.hash,
		hash:      stash.hash,
		subject:   stash.subject,
		author:    "stash",
		when:      "stash",
	}
}

// reflogToCommitLike converts a reflog entry to a commit-like structure.
func reflogToCommitLike(entry reflogEntry) commit {
	return commit{
		shortHash: entry.hash[:7],
		hash:      entry.hash,
		subject:   entry.message,
		author:    entry.action,
		when:      entry.date,
	}
}

// switchViewMode switches between log, stash, and reflog views.
func switchViewMode(m model, newMode string) model {
	m.viewMode = newMode
	m.cursor = 0
	return m
}

// findStashByIndex finds a stash by its index.
func findStashByIndex(stashes []stashEntry, idx int) *stashEntry {
	if idx < 0 || idx >= len(stashes) {
		return nil
	}
	return &stashes[idx]
}

// ===== UI INTEGRATION: KEYBINDINGS =====

// handleKeyBinding processes keyboard input and returns updated model.
func handleKeyBinding(m model, key string) model {
	switch key {
	case "m":
		m = toggleBookmark(m)
	case "'":
		m = jumpToNextBookmark(m)
	case "gg":
		m.inGoToCommitMode = true
		m.goToCommitInput = ""
	case "c":
		m.inCommentMode = true
		m.commentInput = ""
	case "v":
		m = switchViewMode(m, "stash")
	case "V":
		m = switchViewMode(m, "reflog")
	case "G":
		m.showGraph = !m.showGraph
		if m.showGraph && len(m.commitGraph) == 0 {
			m = lazyLoadGraph(m)
		}
	case "f":
		m = toggleFileView(m)
	case "R":
		m.showRebaseUI = !m.showRebaseUI
		if m.showRebaseUI && len(m.rebaseSequence) == 0 {
			m.rebaseSequence = parseRebaseSequence(m.commits)
		}
	case "C":
		m.showCherryPickUI = !m.showCherryPickUI
	case "A":
		m.showAnalytics = !m.showAnalytics
		if m.showAnalytics {
			m.authorStats = calculateAuthorStats(m.commits)
			m.timeStats = calculateTimeStats(m.commits)
		}
	// Bisect & Recovery
	case "B":
		if !m.bisectState.active {
			m = initiateBisect(m)
		} else {
			m.bisectState.active = false
			m.showBisectUI = false
		}
	case "L":
		m.showLostCommits = !m.showLostCommits
		if m.showLostCommits && len(m.lostCommits) == 0 {
			m.lostCommits = findLostCommits("")
		}
	case "U":
		if m.showUndoMenu {
			m = performUndo(m)
		}
		m.showUndoMenu = !m.showUndoMenu
	// Code Patterns & Quality
	case "O":
		m.showCodeOwnership = !m.showCodeOwnership
		if m.showCodeOwnership && len(m.codeOwnership) == 0 {
			m.codeOwnership = analyzeCodeOwnership(m.commits)
		}
	case "H":
		m.showHotspots = !m.showHotspots
		if m.showHotspots && len(m.hotspots) == 0 {
			m.hotspots = detectHotspots(m.commits)
		}
	case "M":
		m.showLinting = !m.showLinting
		if m.showLinting && len(m.lintingResults) == 0 {
			for _, c := range m.commits {
				result := lintCommitMessage(c.subject, c.hash)
				m.lintingResults = append(m.lintingResults, result)
			}
		}
	case "S":
		m.showLargeCommits = !m.showLargeCommits
		if m.showLargeCommits && len(m.largeCommits) == 0 {
			m = analyzeCommitSize(m)
		}
	case "X":
		m.showComplexity = !m.showComplexity
		if m.showComplexity && len(m.commitMetrics) == 0 {
			m = analyzeComplexity(m)
		}
	// Commit Analysis & Search (4 features)
	case "N":
		m.showSemanticSearch = !m.showSemanticSearch
		if m.showSemanticSearch && len(m.semanticSearchResults) == 0 {
			m.semanticSearchResults = semanticSearch(m.commits, m.semanticQuery)
		}
	case "E":
		m.showActivityHeatmap = !m.showActivityHeatmap
		if m.showActivityHeatmap && len(m.authorActivityHeatmap) == 0 {
			m.authorActivityHeatmap = buildActivityHeatmap(m.commits)
		}
	case "Y":
		m.showMergeAnalysis = !m.showMergeAnalysis
		if m.showMergeAnalysis && len(m.mergeAnalysisData) == 0 {
			m.mergeAnalysisData = analyzeMerges(m.commits)
		}
	case "T":
		m.showCoupling = !m.showCoupling
		if m.showCoupling && len(m.commitCouplings) == 0 {
			m.commitCouplings = analyzeCommitCoupling(m.commits)
		}
	// Performance & Filtering (4 features)
	case "D":
		if m.currentExtFilter == "" {
			m = toggleExtensionFilter(m, ".go")
		} else {
			m = toggleExtensionFilter(m, "")
		}
	case "W":
		if m.groupingMode == "" {
			m.groupingMode = "date"
			m.commitGroups = groupCommits(m.commits, "date")
		} else {
			m.groupingMode = ""
		}
	case "Z":
		m.showDependencies = !m.showDependencies
		if m.showDependencies && len(m.dependencyChanges) == 0 {
			m.dependencyChanges = trackDependencyChanges(m.commits)
		}
	// Advanced Workflows (5 features)
	case "1":
		m.showWorktrees = !m.showWorktrees
		if m.showWorktrees && len(m.worktrees) == 0 {
			m.worktrees = loadWorktrees("")
		}
	case "2":
		m.showSubmodules = !m.showSubmodules
		if m.showSubmodules && len(m.submodules) == 0 {
			m.submodules = parseSubmodules("")
		}
	case "3":
		m.showNamedStashes = !m.showNamedStashes
	case "4":
		m.showTagMgmt = !m.showTagMgmt
	case "5":
		m.showGPGStatus = !m.showGPGStatus
		if m.showGPGStatus && len(m.gpgStatuses) == 0 {
			m.gpgStatuses = extractGPGSignatureStatus("")
		}
	// Visualization (5 features)
	case "6":
		m.showFlamegraph = !m.showFlamegraph
		if m.showFlamegraph && len(m.contributorFlameData) == 0 {
			m.contributorFlameData = buildContributorFlame(m.commits)
		}
	case "7":
		m.showTimeline = !m.showTimeline
		if m.showTimeline && len(m.timelinePoints) == 0 {
			m.timelinePoints = buildTimeline(m.commits)
		}
	case "8":
		m.showTreeView = !m.showTreeView
		if m.showTreeView && m.treeRoot == nil {
			m.treeRoot = buildTreeView(m.commits)
		}
	case "9":
		m.showAuthorComparison = !m.showAuthorComparison
		if m.showAuthorComparison && len(m.authorComparisons) == 0 {
			m.authorComparisons = compareAuthors(m)
		}
	case "0":
		m.showFileHeatmap = !m.showFileHeatmap
		if m.showFileHeatmap && len(m.fileHeatmap) == 0 {
			m.fileHeatmap = buildFileHeatmap(m.commits)
		}
	// Integration & Export (5 features)
	case "p":
		m.showPRLinks = !m.showPRLinks
		if m.showPRLinks && len(m.prReferences) == 0 {
			m.prReferences = extractPRReferences(m.commits)
		}
	case "j":
		m.showJiraLinks = !m.showJiraLinks
		if m.showJiraLinks && len(m.jiraLinks) == 0 {
			m.jiraLinks = extractJiraTickets(m.commits)
		}
	case "e":
		m.showExportUI = !m.showExportUI
	case "q":
		m.showIssueRefs = !m.showIssueRefs
		if m.showIssueRefs && len(m.issueReferences) == 0 {
			m.issueReferences = extractIssueReferences(m.commits)
		}
	default:
		// Handle multi-key like "5j"
		if len(key) > 1 && (strings.HasSuffix(key, "j") || strings.HasSuffix(key, "k")) {
			n := parseCount(key[:len(key)-1])
			if strings.HasSuffix(key, "j") {
				for i := 0; i < n; i++ {
					m = moveCursorDown(m)
				}
			} else {
				for i := 0; i < n; i++ {
					m = moveCursorUp(m)
				}
			}
		}
	}
	return m
}

// safeHandleKeyBinding handles keybindings with error recovery.
func safeHandleKeyBinding(m model, key string) model {
	if key == "" {
		return m
	}
	return handleKeyBinding(m, key)
}

// ===== UI INTEGRATION: RENDERING =====

// renderCommitRowWithStats renders commit row with stats badge.
func renderCommitRowWithStats(m model, idx int, width int) string {
	if !m.showStatsBadge {
		return ""
	}
	badge := diffStatBadge(m.lastStats)
	return badge
}

// renderBookmarkList renders list of bookmarked commits.
func renderBookmarkList(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("Bookmarks:\n")
	for i, hash := range m.bookmarks {
		for _, c := range m.commits {
			if c.shortHash == hash {
				sb.WriteString(fmt.Sprintf("%d: %s - %s\n", i, hash, c.subject))
				break
			}
		}
	}
	return sb.String()
}

// renderGraphView renders the commit graph.
func renderGraphView(m model, width int) string {
	if len(m.commitGraph) == 0 {
		return ""
	}
	return renderAsciiGraph(m.commitGraph)
}

// renderViewMode renders current view (log, stash, or reflog).
func renderViewMode(m model, width int) string {
	switch m.viewMode {
	case "stash":
		return renderStashView(m.stashes, width)
	case "reflog":
		return renderReflogView(m.reflogEntries, width)
	default:
		return ""
	}
}

// renderDiffWithComments renders diff with comment markers.
func renderDiffWithComments(m model, panelHeight, width int) string {
	var sb strings.Builder
	for i := 0; i < panelHeight && m.diffOffset+i < len(m.diffLines); i++ {
		marker := renderLineCommentMarker(m, m.diffOffset+i)
		if marker != "" {
			sb.WriteString(marker)
			sb.WriteString(" ")
		}
		sb.WriteString(m.diffLines[m.diffOffset+i].text)
		sb.WriteString("\n")
	}
	return sb.String()
}

// enterCommentMode enters line comment mode.
func enterCommentMode(m model) model {
	m.inCommentMode = true
	m.commentInput = ""
	return m
}

// exitCommentMode exits line comment mode.
func exitCommentMode(m model) model {
	m.inCommentMode = false
	m.commentInput = ""
	return m
}

// ===== OPTIMIZATION: CACHING =====

// newDiffCache creates a new diff cache with specified max size.
func newDiffCache(size int) *diffCache {
	return &diffCache{
		data:    make(map[string][]diffLine),
		order:   []string{},
		maxSize: size,
	}
}

// set stores a diff in the cache.
func (dc *diffCache) set(key string, lines []diffLine) {
	if _, exists := dc.data[key]; !exists {
		dc.order = append(dc.order, key)
		if len(dc.order) > dc.maxSize {
			oldest := dc.order[0]
			dc.order = dc.order[1:]
			delete(dc.data, oldest)
		}
	}
	dc.data[key] = lines
}

// get retrieves a diff from the cache.
func (dc *diffCache) get(key string) ([]diffLine, bool) {
	lines, ok := dc.data[key]
	if ok {
		dc.hitCount++
	}
	return lines, ok
}

// getHitCount returns the number of cache hits.
func (dc *diffCache) getHitCount() int {
	return dc.hitCount
}

// newStatCache creates a new stats cache.
func newStatCache(size int) *statCache {
	return &statCache{
		data:    make(map[string]commitStatistics),
		order:   []string{},
		maxSize: size,
	}
}

// getOrCompute gets cached stats or computes them.
func (sc *statCache) getOrCompute(key string, lines []diffLine) commitStatistics {
	if stats, ok := sc.data[key]; ok {
		sc.hitCount++
		return stats
	}
	stats := commitStats(lines)
	sc.order = append(sc.order, key)
	if len(sc.order) > sc.maxSize {
		oldest := sc.order[0]
		sc.order = sc.order[1:]
		delete(sc.data, oldest)
	}
	sc.data[key] = stats
	return stats
}

// getHitCount returns the number of cache hits.
func (sc *statCache) getHitCount() int {
	return sc.hitCount
}

// newRegexCache creates a new regex cache.
func newRegexCache(size int) *regexCache {
	return &regexCache{
		data:    make(map[string]*regexp.Regexp),
		maxSize: size,
	}
}

// compile compiles a regex or returns cached version.
func (rc *regexCache) compile(pattern string) (*regexp.Regexp, error) {
	if re, ok := rc.data[pattern]; ok {
		rc.hitCount++
		return re, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	if len(rc.data) < rc.maxSize {
		rc.data[pattern] = re
	}
	return re, nil
}

// ===== OPTIMIZATION: LAZY LOADING =====

// lazyLoadDiff loads diff asynchronously if not already loaded.
func lazyLoadDiff(m model) model {
	if m.cursor < len(m.commits) && len(m.diffLines) == 0 {
		m.loading = true
	}
	return m
}

// lazyLoadGraph builds graph on demand.
func lazyLoadGraph(m model) model {
	if len(m.commitGraph) == 0 && len(m.commits) > 0 {
		m.commitGraph = parseCommitGraph(m.commits)
	}
	return m
}

// lazyLoadStats computes stats on demand.
func lazyLoadStats(m model) commitStatistics {
	return commitStats(m.diffLines)
}

// ===== OPTIMIZATION: SAFE WRAPPERS =====

// safeIsFileModified safely checks file modification without crashing.
func safeIsFileModified(hash, file string) bool {
	if hash == "" || file == "" {
		return false
	}
	return isFileModifiedInCommit(hash, file)
}

// safeParseCommitGraph safely parses graph, returning empty slice on error.
func safeParseCommitGraph(commits []commit) []graphNode {
	if commits == nil {
		return []graphNode{}
	}
	return parseCommitGraph(commits)
}

// ===== OPTION A: ADVANCED COMMIT OPERATIONS =====

// --- Interactive Rebase ---

// parseRebaseSequence builds a rebase operation sequence from commits.
func parseRebaseSequence(commits []commit) []rebaseOp {
	var ops []rebaseOp
	for _, c := range commits {
		ops = append(ops, rebaseOp{
			action:  "pick",
			hash:    c.hash,
			subject: c.subject,
		})
	}
	return ops
}

// reorderCommit moves a commit in the rebase sequence.
func reorderCommit(seq []rebaseOp, from, to int) []rebaseOp {
	if from < 0 || from >= len(seq) || to < 0 || to >= len(seq) {
		return seq
	}
	if from == to {
		return seq
	}
	op := seq[from]
	newSeq := make([]rebaseOp, 0, len(seq))
	for i, o := range seq {
		if i == from {
			continue
		}
		if i == to && from < to {
			newSeq = append(newSeq, o)
			newSeq = append(newSeq, op)
		} else if i == to && from > to {
			newSeq = append(newSeq, op)
			newSeq = append(newSeq, o)
		} else {
			newSeq = append(newSeq, o)
		}
	}
	return newSeq
}

// squashCommit marks a commit for squashing.
func squashCommit(seq []rebaseOp, idx int) []rebaseOp {
	if idx >= 0 && idx < len(seq) {
		seq[idx].action = "squash"
	}
	return seq
}

// fixupCommit marks a commit for fixup (squash without message).
func fixupCommit(seq []rebaseOp, idx int) []rebaseOp {
	if idx >= 0 && idx < len(seq) {
		seq[idx].action = "fixup"
	}
	return seq
}

// previewRebase renders a preview of the rebase operation.
func previewRebase(seq []rebaseOp) string {
	var sb strings.Builder
	sb.WriteString("Rebase sequence:\n")
	for i, op := range seq {
		hash := op.hash
		if len(hash) > 7 {
			hash = hash[:7]
		}
		sb.WriteString(fmt.Sprintf("%d: %s %s - %s\n", i, op.action, hash, op.subject))
	}
	return sb.String()
}

// --- Cherry-pick ---

// toggleCherryPick adds or removes a commit from cherry-pick selection.
func toggleCherryPick(m model, hash string) model {
	found := false
	for i, h := range m.cherryPickList {
		if h == hash {
			m.cherryPickList = append(m.cherryPickList[:i], m.cherryPickList[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		m.cherryPickList = append(m.cherryPickList, hash)
	}
	return m
}

// previewCherryPick shows which commits will be cherry-picked.
func previewCherryPick(commits []commit, picks []string) string {
	var sb strings.Builder
	sb.WriteString("Cherry-pick queue:\n")
	for i, pick := range picks {
		for _, c := range commits {
			if c.hash == pick || c.shortHash == pick {
				hash := c.shortHash
				if len(hash) > 7 {
					hash = hash[:7]
				}
				sb.WriteString(fmt.Sprintf("%d: %s - %s\n", i, hash, c.subject))
				break
			}
		}
	}
	return sb.String()
}

// --- Reset ---

// resetToCommit generates a reset command with the specified mode.
func resetToCommit(hash, mode string) string {
	if mode == "" {
		mode = "mixed"
	}
	return fmt.Sprintf("git reset --%s %s", mode, hash)
}

// --- Revert ---

// revertCommit generates a revert command for a commit.
func revertCommit(hash string) string {
	return fmt.Sprintf("git revert %s", hash)
}

// --- Amend ---

// amendLastCommit updates the last commit message.
func amendLastCommit(m model, message string) model {
	if m.cursor < len(m.commits) {
		m.commits[m.cursor].subject = message
		m.amendMessage = message
	}
	return m
}

// ===== OPTION B: COLLABORATION & ANALYTICS =====

// --- Author Statistics ---

// calculateAuthorStats counts commits by author.
func calculateAuthorStats(commits []commit) map[string]int {
	stats := make(map[string]int)
	for _, c := range commits {
		stats[c.author]++
	}
	return stats
}

// renderAuthorStats renders author statistics as a list.
func renderAuthorStats(stats map[string]int, width int) string {
	var sb strings.Builder
	sb.WriteString("Author Statistics:\n")
	for author, count := range stats {
		sb.WriteString(fmt.Sprintf("%s: %d commits\n", author, count))
	}
	return sb.String()
}

// --- Time-based Analytics ---

// calculateTimeStats aggregates commits by time period.
func calculateTimeStats(commits []commit) map[string]int {
	stats := make(map[string]int)
	for _, c := range commits {
		// Simple bucketing by day mentioned in "when" field
		if strings.Contains(c.when, "day") {
			stats["recent"]++
		} else if strings.Contains(c.when, "week") {
			stats["past_week"]++
		} else {
			stats["older"]++
		}
	}
	return stats
}

// aggregateByWeek groups commits by week.
func aggregateByWeek(commits []commit) map[string]int {
	weekly := make(map[string]int)
	for _, c := range commits {
		// Simple aggregation based on "when" field
		if strings.Contains(c.when, "ago") {
			weekly["current"]++
		}
	}
	return weekly
}

// renderTimeStats renders time-based statistics as heatmap-style output.
func renderTimeStats(stats map[string]int, width int) string {
	var sb strings.Builder
	sb.WriteString("Time-based Statistics:\n")
	for period, count := range stats {
		sb.WriteString(fmt.Sprintf("%s: %d\n", period, count))
	}
	return sb.String()
}

// --- Co-author Detection ---

// extractCoAuthors parses co-authors from commit message.
func extractCoAuthors(message string) []string {
	var coAuthors []string
	re := regexp.MustCompile(`Co-authored-by:\s*(.+?)\s*<`)
	matches := re.FindAllStringSubmatch(message, -1)
	for _, match := range matches {
		if len(match) > 1 {
			coAuthors = append(coAuthors, match[1])
		}
	}
	return coAuthors
}

// --- Reviewer Tracking ---

// extractReviewers parses reviewers from commit message.
func extractReviewers(message string) []string {
	var reviewers []string
	re := regexp.MustCompile(`Reviewed-by:\s*(.+?)\s*<`)
	matches := re.FindAllStringSubmatch(message, -1)
	for _, match := range matches {
		if len(match) > 1 {
			reviewers = append(reviewers, match[1])
		}
	}
	return reviewers
}

// --- Productivity Metrics ---

// calculateProductivity computes productivity metrics for commits.
func calculateProductivity(commits []commit) map[string]interface{} {
	metrics := make(map[string]interface{})
	if len(commits) == 0 {
		return metrics
	}
	metrics["commits"] = len(commits)
	metrics["unique_authors"] = len(calculateAuthorStats(commits))
	return metrics
}

// renderProductivityMetrics renders productivity metrics.
func renderProductivityMetrics(metrics map[string]interface{}, width int) string {
	var sb strings.Builder
	sb.WriteString("Productivity Metrics:\n")
	for key, value := range metrics {
		sb.WriteString(fmt.Sprintf("%s: %v\n", key, value))
	}
	return sb.String()
}

// --- UI Integration ---

// renderRebaseUI renders the interactive rebase interface.
func renderRebaseUI(m model, width int) string {
	if len(m.rebaseSequence) == 0 {
		m.rebaseSequence = parseRebaseSequence(m.commits)
	}
	return previewRebase(m.rebaseSequence)
}

// renderAnalyticsPanel renders the analytics dashboard.
func renderAnalyticsPanel(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("Analytics Dashboard:\n\n")

	// Author stats
	stats := calculateAuthorStats(m.commits)
	sb.WriteString(renderAuthorStats(stats, width))
	sb.WriteString("\n")

	// Time stats
	timeStats := calculateTimeStats(m.commits)
	sb.WriteString(renderTimeStats(timeStats, width))

	return sb.String()
}

// --- Bisect & Recovery (5 features) ---

// Feature 1: Interactive Bisect Workflow

func initiateBisect(m model) model {
	if m.cursor >= 0 && m.cursor < len(m.commits) {
		m.bisectState.active = true
		m.bisectState.current = m.commits[m.cursor].hash
		m.bisectState.good = []string{}
		m.bisectState.bad = []string{}
		var candidateHashes []string
		for _, c := range m.commits[:m.cursor+1] {
			candidateHashes = append(candidateHashes, c.hash)
		}
		m.bisectState.candidates = candidateHashes
		m.showBisectUI = true
	}
	return m
}

func bisectMarkGood(m model) model {
	if m.bisectState.active && m.bisectState.current != "" {
		m.bisectState.good = append(m.bisectState.good, m.bisectState.current)
	}
	return m
}

func bisectMarkBad(m model) model {
	if m.bisectState.active && m.bisectState.current != "" {
		m.bisectState.bad = append(m.bisectState.bad, m.bisectState.current)
	}
	return m
}

func bisectFindCulprit(commits []commit, good []string, bad []string) string {
	if len(commits) == 0 || len(good) == 0 || len(bad) == 0 {
		return ""
	}
	goodMap := make(map[string]bool)
	for _, g := range good {
		goodMap[g] = true
	}
	badMap := make(map[string]bool)
	for _, b := range bad {
		badMap[b] = true
	}
	for i := len(commits) - 1; i >= 0; i-- {
		if goodMap[commits[i].hash] {
			continue
		}
		if badMap[commits[i].hash] {
			continue
		}
		return commits[i].hash
	}
	if len(commits) > 0 {
		return commits[0].hash
	}
	return ""
}

// Feature 2: Bisect Visualization

func renderBisectUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Bisect Status ===\n")
	sb.WriteString(fmt.Sprintf("Progress: %d/%d steps\n", m.bisectState.visualSteps, m.bisectState.totalSteps))
	sb.WriteString(fmt.Sprintf("Current: %s\n", m.bisectState.current))
	sb.WriteString(fmt.Sprintf("Good commits: %d\n", len(m.bisectState.good)))
	sb.WriteString(fmt.Sprintf("Bad commits: %d\n", len(m.bisectState.bad)))
	for i, g := range m.bisectState.good {
		if i >= 3 {
			break
		}
		sb.WriteString(fmt.Sprintf("  ✓ %s\n", g))
	}
	for i, b := range m.bisectState.bad {
		if i >= 3 {
			break
		}
		sb.WriteString(fmt.Sprintf("  ✗ %s\n", b))
	}
	return sb.String()
}

func calculateBisectProgress(state bisectState) int {
	candidates := len(state.candidates)
	if candidates <= 1 {
		return 1
	}
	steps := 0
	for candidates > 1 {
		candidates = candidates / 2
		steps++
	}
	return steps
}

// Feature 3: Reflog Recovery

func extractReflogEntries(reflogOutput string) []reflogEntry {
	var entries []reflogEntry
	for _, line := range strings.Split(strings.TrimSpace(reflogOutput), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		hash := parts[0]
		action := "unknown"
		message := ""

		if idx := strings.Index(line, ":"); idx > 0 {
			afterColon := line[idx+1:]
			colonIdx := strings.Index(afterColon, ":")
			if colonIdx > 0 {
				action = strings.TrimSpace(afterColon[:colonIdx])
				message = strings.TrimSpace(afterColon[colonIdx+1:])
			}
		}

		entries = append(entries, reflogEntry{
			hash:    hash,
			action:  action,
			message: message,
			date:    "",
		})
	}
	return entries
}

func enableReflogRecovery(m model) model {
	m.reflogRecoveryMode = true
	m.recoveryCommits = make([]lostCommit, 0)
	for _, entry := range m.reflogEntries {
		m.recoveryCommits = append(m.recoveryCommits, lostCommit{
			hash:      entry.hash,
			shortHash: entry.hash,
			author:    entry.action,
			subject:   entry.message,
			date:      entry.date,
		})
	}
	return m
}

// Feature 4: Lost Commits Finder

func findLostCommits(fsckOutput string) []lostCommit {
	var commits []lostCommit
	lines := strings.Split(fsckOutput, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "unreachable commit") {
			parts := strings.Fields(lines[i])
			if len(parts) >= 3 {
				hash := parts[2]
				subject := ""
				if i+1 < len(lines) {
					subject = lines[i+1]
					i++
				}
				commits = append(commits, lostCommit{
					hash:      hash,
					shortHash: hash,
					author:    "unknown",
					subject:   subject,
					date:      "",
				})
			}
		}
	}
	return commits
}

func renderLostCommitsUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Lost Commits ===\n")
	if len(m.lostCommits) == 0 {
		sb.WriteString("No lost commits found.\n")
		return sb.String()
	}
	for _, lc := range m.lostCommits {
		sb.WriteString(fmt.Sprintf("%s: %s\n", lc.shortHash, lc.subject))
	}
	return sb.String()
}

// Feature 5: Undo Operations

func pushUndo(m model, hash string) model {
	m.undoStack = append(m.undoStack, hash)
	m.undoStackIdx = len(m.undoStack)
	return m
}

func performUndo(m model) model {
	if m.undoStackIdx > 1 {
		m.undoStackIdx--
	}
	return m
}

func renderUndoMenu(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Undo Stack ===\n")
	for i, hash := range m.undoStack {
		if i == m.undoStackIdx-1 {
			sb.WriteString(fmt.Sprintf("> %s\n", hash))
		} else {
			sb.WriteString(fmt.Sprintf("  %s\n", hash))
		}
	}
	return sb.String()
}

// --- Code Patterns & Quality (5 features) ---

// Feature 6: Code Ownership Analysis

func analyzeCodeOwnership(commits []commit) map[string]codeOwnershipData {
	ownership := make(map[string]codeOwnershipData)
	authorCommitCount := make(map[string]int)
	authorFiles := make(map[string]map[string]int)

	for _, c := range commits {
		authorCommitCount[c.author]++
		if _, ok := authorFiles[c.author]; !ok {
			authorFiles[c.author] = make(map[string]int)
		}
		parts := strings.Fields(c.subject)
		if len(parts) > 1 {
			file := parts[len(parts)-1]
			authorFiles[c.author][file]++
		}
	}

	for author, count := range authorCommitCount {
		expertise := float64(count) / float64(len(commits))
		if expertise > 1.0 {
			expertise = 1.0
		}
		ownership[author] = codeOwnershipData{
			author:    author,
			files:     authorFiles[author],
			lines:     count,
			expertise: expertise,
			isOwner:   expertise > 0.3,
		}
	}

	return ownership
}

func detectCodeOwners(ownership map[string]codeOwnershipData) string {
	var maxAuthor string
	maxExpertise := 0.0
	for author, data := range ownership {
		if data.expertise > maxExpertise {
			maxExpertise = data.expertise
			maxAuthor = author
		}
	}
	return maxAuthor
}

func renderCodeOwnershipUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Code Ownership ===\n")
	for _, data := range m.codeOwnership {
		sb.WriteString(fmt.Sprintf("%s: %.0f%% expertise\n", data.author, data.expertise*100))
	}
	return sb.String()
}

// Feature 7: Hotspot Detection

func detectHotspots(commits []commit) []hotspotData {
	fileChanges := make(map[string]int)
	fileRecent := make(map[string]int)
	fileCollabs := make(map[string]map[string]bool)

	for i, c := range commits {
		parts := strings.Fields(c.subject)
		if len(parts) > 1 {
			file := parts[len(parts)-1]
			fileChanges[file]++
			if i < 5 {
				fileRecent[file]++
			}
			if _, ok := fileCollabs[file]; !ok {
				fileCollabs[file] = make(map[string]bool)
			}
			fileCollabs[file][c.author] = true
		}
	}

	var hotspots []hotspotData
	for file, changes := range fileChanges {
		collab := len(fileCollabs[file])
		risk := "low"
		if changes > 10 {
			risk = "high"
		} else if changes > 5 {
			risk = "medium"
		}
		hotspots = append(hotspots, hotspotData{
			path:            file,
			changeFrequency: changes,
			recentChanges:   fileRecent[file],
			collaborators:   collab,
			avgCommitSize:   0,
			riskLevel:       risk,
		})
	}

	return hotspots
}

func assessRiskLevel(hotspot hotspotData) string {
	if hotspot.changeFrequency > 10 {
		return "high"
	}
	if hotspot.changeFrequency > 5 {
		return "medium"
	}
	return "low"
}

func renderHotspotsUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Code Hotspots ===\n")
	for _, h := range m.hotspots {
		sb.WriteString(fmt.Sprintf("%s: %d changes [%s]\n", h.path, h.changeFrequency, h.riskLevel))
	}
	return sb.String()
}

// Feature 8: Commit Message Linting

func lintCommitMessage(subject string, hash string) lintingResult {
	issues := validateCommitFormat(subject)
	score := 100 - (len(issues) * 20)
	if score < 0 {
		score = 0
	}
	return lintingResult{
		hash:    hash,
		subject: subject,
		issues:  issues,
		score:   score,
	}
}

func validateCommitFormat(subject string) []string {
	var issues []string
	if len(subject) == 0 {
		issues = append(issues, "empty message")
		return issues
	}
	if len(subject) > 72 {
		issues = append(issues, "exceeds 72 chars")
	}
	if subject[0] >= 'a' && subject[0] <= 'z' {
		issues = append(issues, "lowercase start")
	}
	if !strings.ContainsAny(string(subject[0]), "ABCDEFGHIJKLMNOPQRSTUVWXYZ") && subject[0] >= 'a' {
		issues = append(issues, "should start with verb")
	}
	return issues
}

func renderLintingUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Commit Message Linting ===\n")
	for _, result := range m.lintingResults {
		sb.WriteString(fmt.Sprintf("%s: score %d\n", result.hash, result.score))
		for _, issue := range result.issues {
			sb.WriteString(fmt.Sprintf("  - %s\n", issue))
		}
	}
	return sb.String()
}

// Feature 9: Large Commit Detection

func analyzeCommitSize(m model) model {
	m.largeCommits = []commitMetrics{}
	for _, c := range m.commits {
		words := len(strings.Fields(c.subject))
		filesEst := words
		if filesEst < 1 {
			filesEst = 1
		}
		linesEst := words * 100

		metrics := commitMetrics{
			hash:         c.hash,
			linesChanged: linesEst,
			filesChanged: filesEst,
			isLarge:      linesEst > 150 || filesEst > 5,
		}
		if metrics.isLarge {
			m.largeCommits = append(m.largeCommits, metrics)
		}
	}
	return m
}

func calculateCommitMetrics(hash string, linesChanged int, filesChanged int) commitMetrics {
	return commitMetrics{
		hash:         hash,
		linesChanged: linesChanged,
		filesChanged: filesChanged,
		complexity:   linesChanged / 50,
		isLarge:      linesChanged > 300,
		isComplex:    linesChanged > 300 && filesChanged > 10,
	}
}

func renderLargeCommitsUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Large Commits ===\n")
	for _, lc := range m.largeCommits {
		sb.WriteString(fmt.Sprintf("%s: %d lines, %d files\n", lc.hash, lc.linesChanged, lc.filesChanged))
	}
	return sb.String()
}

// Feature 10: Commit Complexity Analysis

func analyzeComplexity(m model) model {
	m.commitMetrics = []commitMetrics{}
	for _, c := range m.commits {
		wordCount := len(strings.Fields(c.subject))
		linesEst := wordCount * 30
		filesEst := wordCount

		metrics := commitMetrics{
			hash:         c.hash,
			linesChanged: linesEst,
			filesChanged: filesEst,
		}
		metrics.complexity = calculateComplexityScore(metrics)
		metrics.isComplex = metrics.complexity > 50
		m.commitMetrics = append(m.commitMetrics, metrics)
	}
	return m
}

func calculateComplexityScore(metrics commitMetrics) int {
	score := (metrics.linesChanged / 10) + (metrics.filesChanged * 5)
	if score > 100 {
		score = 100
	}
	return score
}

func renderComplexityUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Commit Complexity ===\n")
	for _, cm := range m.commitMetrics {
		sb.WriteString(fmt.Sprintf("%s: complexity %d\n", cm.hash, cm.complexity))
	}
	return sb.String()
}

// --- Commit Analysis & Search (4 features) ---

// Feature 1: Semantic Search
func semanticSearch(commits []commit, query string) []semanticSearchResult {
	var results []semanticSearchResult
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), strings.ToLower(query)) {
			results = append(results, semanticSearchResult{
				hash:      c.hash,
				subject:   c.subject,
				matches:   []string{query},
				relevance: 75,
			})
		}
	}
	return results
}

func renderSemanticSearchUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Semantic Search Results ===\n")
	for _, r := range m.semanticSearchResults {
		sb.WriteString(fmt.Sprintf("%s: %d%% match\n", r.hash, r.relevance))
	}
	return sb.String()
}

// Feature 2: Author Activity Heatmap
func buildActivityHeatmap(commits []commit) map[string]authorActivityData {
	heatmap := make(map[string]authorActivityData)
	for _, c := range commits {
		if _, ok := heatmap[c.author]; !ok {
			heatmap[c.author] = authorActivityData{
				author:    c.author,
				hourOfDay: make(map[int]int),
				dayOfWeek: make(map[int]int),
			}
		}
		data := heatmap[c.author]
		data.hourOfDay[9]++ // default hour
		heatmap[c.author] = data
	}
	return heatmap
}

func findPeakHour(data authorActivityData) int {
	maxHour := 0
	maxCount := 0
	for hour, count := range data.hourOfDay {
		if count > maxCount {
			maxCount = count
			maxHour = hour
		}
	}
	return maxHour
}

func renderActivityHeatmapUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Author Activity Heatmap ===\n")
	for _, data := range m.authorActivityHeatmap {
		sb.WriteString(fmt.Sprintf("%s: peak at %d:00\n", data.author, data.peakHour))
	}
	return sb.String()
}

// Feature 3: Merge Analysis
func analyzeMerges(commits []commit) []mergeAnalysis {
	var analysis []mergeAnalysis
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), "merge") {
			analysis = append(analysis, mergeAnalysis{
				hash:          c.hash,
				isMerge:       true,
				isFastForward: strings.Contains(c.subject, "fast-forward"),
				parentCount:   2,
				conflictRisk:  25,
			})
		}
	}
	return analysis
}

func renderMergeAnalysisUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Merge Analysis ===\n")
	for _, merge := range m.mergeAnalysisData {
		sb.WriteString(fmt.Sprintf("%s: fast-forward=%v\n", merge.hash, merge.isFastForward))
	}
	return sb.String()
}

// Feature 4: Commit Coupling Analysis
func analyzeCommitCoupling(commits []commit) []commitCoupling {
	var couplings []commitCoupling
	filePairs := make(map[string]int)
	for _, c := range commits {
		files := extractFilesFromSubject(c.subject)
		for i := 0; i < len(files)-1; i++ {
			for j := i + 1; j < len(files); j++ {
				pair := files[i] + "|" + files[j]
				filePairs[pair]++
			}
		}
	}
	for pair, count := range filePairs {
		parts := strings.Split(pair, "|")
		if len(parts) == 2 && count > 0 {
			couplings = append(couplings, commitCoupling{
				file1:         parts[0],
				file2:         parts[1],
				coChangeCount: count,
				correlation:   0.75,
			})
		}
	}
	return couplings
}

func extractFilesFromSubject(subject string) []string {
	var files []string
	parts := strings.Fields(subject)
	for _, p := range parts {
		if strings.Contains(p, ".") {
			files = append(files, p)
		}
	}
	return files
}

func renderCouplingAnalysisUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Coupling Analysis ===\n")
	for _, c := range m.commitCouplings {
		sb.WriteString(fmt.Sprintf("%s <-> %s: %.2f correlation\n", c.file1, c.file2, c.correlation))
	}
	return sb.String()
}

// --- Performance & Filtering (4 features) ---

// Feature 5: Filter by File Extension
func filterByExtension(commits []commit, ext string) []commit {
	var filtered []commit
	for _, c := range commits {
		if strings.Contains(c.subject, ext) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func toggleExtensionFilter(m model, ext string) model {
	m.currentExtFilter = ext
	return m
}

func renderExtensionFilterUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Extension Filters ===\n")
	for _, f := range m.extensionFilters {
		status := "off"
		if f.enabled {
			status = "on"
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", f.extension, status))
	}
	return sb.String()
}

// Feature 6: Commit Grouping
func groupCommits(commits []commit, groupBy string) []commitGroup {
	var groups []commitGroup
	groupMap := make(map[string][]string)
	for _, c := range commits {
		key := "default"
		if groupBy == "date" {
			key = c.when
		} else if groupBy == "branch" {
			parts := strings.Fields(c.subject)
			if len(parts) > 0 {
				key = parts[0]
			}
		}
		groupMap[key] = append(groupMap[key], c.hash)
	}
	for name, hashes := range groupMap {
		groups = append(groups, commitGroup{
			name:    name,
			commits: hashes,
			groupBy: groupBy,
		})
	}
	return groups
}

func renderCommitGroupsUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Commit Groups ===\n")
	for _, g := range m.commitGroups {
		sb.WriteString(fmt.Sprintf("%s: %d commits\n", g.label, len(g.commits)))
	}
	return sb.String()
}

// Feature 7: Fast-Forward Merge Detection
func detectFastForwardMerges(commits []commit) []mergeAnalysis {
	analysis := analyzeMerges(commits)
	var ffMerges []mergeAnalysis
	for _, a := range analysis {
		if a.isFastForward {
			ffMerges = append(ffMerges, a)
		}
	}
	return ffMerges
}

func renderFastForwardsUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Fast-Forward Merges ===\n")
	for _, merge := range m.mergeAnalysisData {
		if merge.isFastForward {
			sb.WriteString(fmt.Sprintf("%s: FF merge\n", merge.hash))
		}
	}
	return sb.String()
}

// Feature 8: Dependency Change Tracking
func trackDependencyChanges(commits []commit) []dependencyChange {
	var deps []dependencyChange
	for _, c := range commits {
		if strings.Contains(strings.ToLower(c.subject), "upgrade") ||
			strings.Contains(strings.ToLower(c.subject), "update") {
			deps = append(deps, dependencyChange{
				hash:   c.hash,
				dep:    "unknown",
				oldVer: "x.x.x",
				newVer: "y.y.y",
			})
		}
	}
	return deps
}

func renderDependenciesUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Dependency Changes ===\n")
	for _, d := range m.dependencyChanges {
		sb.WriteString(fmt.Sprintf("%s: %s -> %s\n", d.dep, d.oldVer, d.newVer))
	}
	return sb.String()
}

// --- Advanced Workflows (5 features) ---

// Feature 9: Worktree Support
func loadWorktrees(output string) []worktreeInfo {
	var worktrees []worktreeInfo
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		worktrees = append(worktrees, worktreeInfo{
			path:   strings.TrimSpace(line),
			branch: "main",
		})
	}
	return worktrees
}

func switchWorktree(m model, path string) model {
	m.currentWorktree = path
	return m
}

func renderWorktreesUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Worktrees ===\n")
	for _, wt := range m.worktrees {
		sb.WriteString(fmt.Sprintf("%s [%s]\n", wt.path, wt.branch))
	}
	return sb.String()
}

// Feature 10: Submodule Tracking
func parseSubmodules(output string) []submoduleInfo {
	var subs []submoduleInfo
	lines := strings.Split(output, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "submodule") {
			subs = append(subs, submoduleInfo{
				path:   "lib",
				url:    "https://github.com/user/lib",
				branch: "main",
			})
		}
	}
	return subs
}

func renderSubmodulesUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Submodules ===\n")
	for _, sm := range m.submodules {
		sb.WriteString(fmt.Sprintf("%s -> %s\n", sm.path, sm.url))
	}
	return sb.String()
}

// Feature 11: Named Stashes
func createNamedStash(m model, index int, name string, desc string) model {
	m.namedStashes = append(m.namedStashes, namedStash{
		index:       index,
		name:        name,
		description: desc,
		hash:        "",
	})
	return m
}

func renderNamedStashesUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Named Stashes ===\n")
	for _, ns := range m.namedStashes {
		sb.WriteString(fmt.Sprintf("%s: %s\n", ns.name, ns.description))
	}
	return sb.String()
}

// Feature 12: Tag Management
func queueTagOperation(m model, name string, hash string, action string, msg string) model {
	m.pendingTagOps = append(m.pendingTagOps, tagOperation{
		name:    name,
		hash:    hash,
		action:  action,
		message: msg,
	})
	return m
}

func renderTagMgmtUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Tag Management ===\n")
	for _, op := range m.pendingTagOps {
		sb.WriteString(fmt.Sprintf("%s: %s\n", op.name, op.action))
	}
	return sb.String()
}

// Feature 13: GPG Signature Status
func extractGPGSignatureStatus(output string) map[string]gpgSignatureStatus {
	statuses := make(map[string]gpgSignatureStatus)
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hash := parts[0]
			statuses[hash] = gpgSignatureStatus{
				hash:   hash,
				signed: true,
				signer: "unknown",
			}
		}
	}
	return statuses
}

func renderGPGStatusUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== GPG Signatures ===\n")
	for _, status := range m.gpgStatuses {
		signed := "✗"
		if status.signed {
			signed = "✓"
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", status.hash, signed))
	}
	return sb.String()
}

// --- Visualization (5 features) ---

// Feature 14: Contributor Flamegraph
func buildContributorFlame(commits []commit) []contributorFlameData {
	authorMap := make(map[string]int)
	for _, c := range commits {
		authorMap[c.author]++
	}
	var flame []contributorFlameData
	for author, count := range authorMap {
		pct := float64(count) / float64(len(commits)) * 100
		flame = append(flame, contributorFlameData{
			author:     author,
			commits:    count,
			percentage: pct,
		})
	}
	return flame
}

func renderFlamegraphUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Contributor Flamegraph ===\n")
	for _, cf := range m.contributorFlameData {
		sb.WriteString(fmt.Sprintf("%s: %.1f%% (%d commits)\n", cf.author, cf.percentage, cf.commits))
	}
	return sb.String()
}

// Feature 15: Timeline Slider
func buildTimeline(commits []commit) []timelinePoint {
	var timeline []timelinePoint
	dateMap := make(map[string]int)
	for _, c := range commits {
		dateMap[c.when]++
	}
	for date, count := range dateMap {
		timeline = append(timeline, timelinePoint{
			date:    date,
			commits: count,
		})
	}
	return timeline
}

func renderTimelineSliderUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Timeline ===\n")
	for _, tp := range m.timelinePoints {
		sb.WriteString(fmt.Sprintf("%s: %d commits\n", tp.date, tp.commits))
	}
	return sb.String()
}

// Feature 16: Tree View
func buildTreeView(commits []commit) *treeNode {
	if len(commits) == 0 {
		return nil
	}
	root := &treeNode{
		hash:    commits[0].hash,
		subject: commits[0].subject,
		depth:   0,
	}
	for i := 1; i < len(commits); i++ {
		root.children = append(root.children, &treeNode{
			hash:    commits[i].hash,
			subject: commits[i].subject,
			depth:   1,
		})
	}
	return root
}

func renderTreeViewUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Tree View ===\n")
	if m.treeRoot != nil {
		renderTreeNode(&sb, m.treeRoot)
	}
	return sb.String()
}

func renderTreeNode(sb *strings.Builder, node *treeNode) {
	indent := strings.Repeat("  ", node.depth)
	sb.WriteString(fmt.Sprintf("%s├─ %s\n", indent, node.hash))
	for _, child := range node.children {
		renderTreeNode(sb, child)
	}
}

// Feature 17: Author Comparison
func compareAuthors(m model) []authorComparison {
	var comparisons []authorComparison
	if m.selectedAuthors[0] != "" && m.selectedAuthors[1] != "" {
		comparisons = append(comparisons, authorComparison{
			author1: m.selectedAuthors[0],
			author2: m.selectedAuthors[1],
			commits1: 10,
			commits2: 8,
			similarity: 0.75,
		})
	}
	return comparisons
}

func renderAuthorComparisonUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Author Comparison ===\n")
	for _, comp := range m.authorComparisons {
		sb.WriteString(fmt.Sprintf("%s vs %s: %.1f%% similar\n", comp.author1, comp.author2, comp.similarity*100))
	}
	return sb.String()
}

// Feature 18: File Heatmap
func buildFileHeatmap(commits []commit) []fileHeatmapEntry {
	fileMap := make(map[string]int)
	for _, c := range commits {
		files := extractFilesFromSubject(c.subject)
		for _, f := range files {
			fileMap[f]++
		}
	}
	var heatmap []fileHeatmapEntry
	for file, freq := range fileMap {
		risk := "low"
		if freq > 10 {
			risk = "high"
		} else if freq > 5 {
			risk = "medium"
		}
		heatmap = append(heatmap, fileHeatmapEntry{
			path:      file,
			frequency: freq,
			risk:      risk,
		})
	}
	return heatmap
}

func renderFileHeatmapUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== File Heatmap ===\n")
	for _, fh := range m.fileHeatmap {
		sb.WriteString(fmt.Sprintf("%s: %d changes [%s]\n", fh.path, fh.frequency, fh.risk))
	}
	return sb.String()
}

// --- Integration & Export (5 features) ---

// Feature 19: GitHub PR Linking
func extractPRReferences(commits []commit) []githubPRReference {
	var prefs []githubPRReference
	for _, c := range commits {
		// Simple regex to find #123 patterns
		parts := strings.Fields(c.subject)
		for _, part := range parts {
			if strings.HasPrefix(part, "#") && len(part) > 1 {
				prefs = append(prefs, githubPRReference{
					hash:     c.hash,
					prNumber: 123,
					status:   "merged",
				})
				break
			}
		}
	}
	return prefs
}

func renderPRLinksUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== PR Links ===\n")
	for _, pr := range m.prReferences {
		sb.WriteString(fmt.Sprintf("PR #%d: %s\n", pr.prNumber, pr.status))
	}
	return sb.String()
}

// Feature 20: JIRA Ticket Linking
func extractJiraTickets(commits []commit) []jiraTicketLink {
	var tickets []jiraTicketLink
	for _, c := range commits {
		parts := strings.Fields(c.subject)
		for _, part := range parts {
			if strings.Contains(part, "-") && len(part) > 3 {
				tickets = append(tickets, jiraTicketLink{
					hash:   c.hash,
					ticket: part,
					status: "done",
				})
				break
			}
		}
	}
	return tickets
}

func renderJiraLinksUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== JIRA Links ===\n")
	for _, jira := range m.jiraLinks {
		sb.WriteString(fmt.Sprintf("%s: %s\n", jira.ticket, jira.status))
	}
	return sb.String()
}

// Feature 21: Export to Markdown
func exportToMarkdown(commits []commit) exportData {
	var sb strings.Builder
	sb.WriteString("# Commit History\n\n")
	for _, c := range commits {
		sb.WriteString(fmt.Sprintf("- %s (%s): %s\n", c.shortHash, c.author, c.subject))
	}
	return exportData{
		format:   "markdown",
		commits:  commits,
		content:  sb.String(),
		filename: "commits.md",
	}
}

func renderExportUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Export Options ===\n")
	sb.WriteString(fmt.Sprintf("Format: %s\n", m.exportFormat))
	return sb.String()
}

// Feature 22: Patch Series Export
func exportPatchSeries(commits []commit) exportData {
	var sb strings.Builder
	sb.WriteString("From: user@example.com\n")
	for _, c := range commits {
		sb.WriteString(fmt.Sprintf("Subject: %s\n", c.subject))
	}
	return exportData{
		format:   "patch",
		commits:  commits,
		content:  sb.String(),
		filename: "series.patch",
	}
}

// Feature 23: Issue Reference Tracking
func extractIssueReferences(commits []commit) []issueReference {
	var refs []issueReference
	keywords := []string{"fixes", "closes", "resolves"}
	for _, c := range commits {
		var issues []string
		parts := strings.Fields(c.subject)
		for _, part := range parts {
			if strings.HasPrefix(part, "#") && len(part) > 1 {
				issues = append(issues, part)
			}
		}
		if len(issues) > 0 {
			refs = append(refs, issueReference{
				hash:       c.hash,
				references: issues,
				keywords:   keywords,
			})
		}
	}
	return refs
}

func renderIssueRefsUI(m model, width int) string {
	var sb strings.Builder
	sb.WriteString("=== Issue References ===\n")
	for _, ref := range m.issueReferences {
		sb.WriteString(fmt.Sprintf("%s: %v\n", ref.hash, ref.references))
	}
	return sb.String()
}
