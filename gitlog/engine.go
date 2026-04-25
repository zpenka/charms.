package gitlog

import "strings"

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

type model struct {
	commits    []commit
	cursor     int
	focus      panel
	diffLines  []diffLine
	diffOffset int
	repoPath   string
	width      int
	height     int
	loading    bool
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
