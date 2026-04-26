package gitlog

import (
	"regexp"
	"strconv"
	"strings"
)

// filterCommits returns commits matching the search query.
// Searches subject, author, and short hash (all case-insensitive).
// Empty query returns all commits unchanged.
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

// filterCommitsByAuthor returns commits from a specific author.
// Performs exact match (case-insensitive) on the author field.
// Empty author returns all commits unchanged.
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

// filterCommitsSince returns commits from the last N days.
// Parses relative time strings like "5 days ago" and "2 weeks ago" from the when field.
// Returns all commits if days <= 0.
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

// isWithinDays checks if a relative time string represents a date within N days.
// Supports time units: days, weeks, months, years.
// Example: "5 days ago", "2 weeks ago".
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

// filterByRegex returns commits whose subject matches the regex pattern.
// Returns nil if the pattern is invalid.
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

// filterByDateRange returns commits with relative age between startDays and endDays.
// Both bounds are inclusive. Uses parseDaysAgo to extract numeric age from when field.
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

// filterByExtension returns commits whose subject contains the file extension.
// Used to filter commits by affected file types (e.g., ".go", ".js").
func filterByExtension(commits []commit, ext string) []commit {
	var filtered []commit
	for _, c := range commits {
		if strings.Contains(c.subject, ext) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// parseDaysAgo extracts the numeric age in days from a relative time string.
// Returns 0 if unable to parse. Used by filterByDateRange.
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
