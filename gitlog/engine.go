package gitlog

import (
	"regexp"
	"strconv"
	"strings"
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
	repoPath     string
	width        int
	height       int
	loading      bool
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
