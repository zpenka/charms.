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
	repoPath        string
	width           int
	height          int
	loading         bool
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
