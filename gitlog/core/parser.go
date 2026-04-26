package gitlog

import (
	"strings"
)

// parseCommits parses git log output into commit structures.
// The input format uses null bytes to separate fields:
// hash\x00shortHash\x00author\x00when\x00subject\n per line.
// Returns nil if input is empty or contains no valid commits.
func parseCommits(input string) []commit {
	if input == "" {
		return nil
	}
	if strings.TrimSpace(input) == "" {
		return nil
	}

	var commits []commit
	for _, line := range strings.Split(input, "\n") {
		if line == "" {
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

	if len(commits) == 0 {
		return nil
	}
	return commits
}

// parseDiff parses a unified diff format into individual classified lines.
// Each line is categorized (added, removed, context, hunk, meta) based on its prefix.
// Returns a slice of diffLine with appropriate lineKind classifications.
func parseDiff(diff string) []diffLine {
	var lines []diffLine
	for _, line := range strings.Split(diff, "\n") {
		if len(line) == 0 {
			continue
		}

		kind := lineContext
		text := line

		switch line[0] {
		case '+':
			kind = lineAdded
			text = line[1:]
		case '-':
			kind = lineRemoved
			text = line[1:]
		case '@':
			kind = lineHunk
		case 'd', 'i', 'n': // diff, index, new file
			kind = lineMeta
		}

		lines = append(lines, diffLine{kind: kind, text: text})
	}
	return lines
}

// parseFileItems extracts file items from commits based on commit subjects.
// Assumes commit subject contains the file path.
// Returns fileItems indexed by commit position for diff navigation.
func parseFileItems(commits []commit) []fileItem {
	var items []fileItem
	for i, c := range commits {
		items = append(items, fileItem{
			path:    c.subject,
			diffIdx: i,
		})
	}
	return items
}
