package gitlog

import (
	"regexp"
	"strconv"
	"strings"
)

// truncate shortens a string to a maximum width in runes.
// If the string exceeds max length, it is truncated and "…" is appended.
// Properly handles Unicode characters by counting runes not bytes.
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

// firstWord returns the first space-delimited word of a string.
// Returns the entire string if it contains no spaces.
func firstWord(s string) string {
	if i := strings.Index(s, " "); i >= 0 {
		return s[:i]
	}
	return s
}

// parseCount parses a string as a navigation count with bounds [1, 200].
// Empty strings, invalid numbers, or out-of-range values default to 1.
// Used for repeat counts in keyboard navigation (e.g., "5j" = 5 lines down).
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

// parseGitReferences extracts GitHub issue/PR references (#123) from text.
// Returns unique references in order of appearance.
// Useful for linking commits to issues and pull requests.
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

// parseBranches extracts branch names from git branch output.
// Removes the current branch marker (*) and filters out ref pointers (HEAD -> ...).
// Returns clean branch names suitable for display and navigation.
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

// parseCurrentBranch returns the currently checked-out branch name.
// Finds the line prefixed with "* " in git branch output.
// Returns empty string if no current branch found.
func parseCurrentBranch(output string) string {
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.HasPrefix(line, "* ") {
			return strings.TrimPrefix(line, "* ")
		}
	}
	return ""
}

// parseBlameLine parses a single line from git blame --date=short output.
// Format: "hash (Author Name   YYYY-MM-DD  linenum) content"
// Returns (blameLine, true) on success, (_, false) if parsing fails.
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

// parseBlame parses all lines from git blame --date=short output.
// Silently skips any malformed lines.
// Returns a slice of blameLine with valid blame information.
func parseBlame(output string) []blameLine {
	var lines []blameLine
	for _, line := range strings.Split(output, "\n") {
		if bl, ok := parseBlameLine(line); ok {
			lines = append(lines, bl)
		}
	}
	return lines
}

// isMergeCommit detects if a commit is a merge commit from diff lines.
// Checks for "Merge:" prefix or "merge branch" text in diff metadata.
// Useful for visualizing merge commits differently in the UI.
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
