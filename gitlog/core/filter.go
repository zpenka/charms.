package gitlog

import (
	"regexp"
	"strconv"
	"strings"
)

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

// filterCommitsByAuthor returns commits matching the exact author name (case-insensitive).
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

// filterByRegex filters commits by regex pattern matching against the subject.
func filterByRegex(commits []commit, pattern string) []commit {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}

	var result []commit
	for _, c := range commits {
		if re.MatchString(c.subject) {
			result = append(result, c)
		}
	}
	return result
}

// filterByDateRange returns commits within the specified day range (inclusive).
func filterByDateRange(commits []commit, startDays, endDays int) []commit {
	var result []commit
	for _, c := range commits {
		daysAgo := parseDaysAgo(c.when)
		if daysAgo >= startDays && daysAgo <= endDays {
			result = append(result, c)
		}
	}
	return result
}

// filterByExtension returns commits whose subject contains the given extension.
func filterByExtension(commits []commit, ext string) []commit {
	var filtered []commit
	for _, c := range commits {
		if strings.Contains(c.subject, ext) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// parseDaysAgo extracts the number of days from a "when" string.
func parseDaysAgo(when string) int {
	parts := strings.Fields(when)
	if len(parts) < 2 {
		return 0
	}
	days, _ := strconv.Atoi(parts[0])
	return days
}

// groupCommits groups commits by the specified grouping mode.
func groupCommits(commits []commit, groupBy string) []commitGroup {
	var groups []commitGroup
	groupMap := make(map[string][]string)
	for _, c := range commits {
		key := "default"
		if groupBy == "date" {
			key = c.when
		} else if groupBy == "author" {
			key = c.author
		} else if groupBy == "branch" {
			parts := strings.Fields(c.subject)
			if len(parts) > 0 {
				key = parts[0]
			}
		}
		groupMap[key] = append(groupMap[key], c.hash)
	}

	for k, v := range groupMap {
		groups = append(groups, commitGroup{
			name:        k,
			commitHashes: v,
		})
	}
	return groups
}
