package gitlog

import (
	"strings"
)

func makeCommits(n int) []commit {
	var cs []commit
	for i := 0; i < n; i++ {
		cs = append(cs, commit{
			hash:      strings.Repeat("a", 40),
			shortHash: "abc1234",
			author:    "Test User",
			when:      "1d ago",
			subject:   "commit message",
		})
	}
	return cs
}

func makeDiffLines(n int) []diffLine {
	lines := make([]diffLine, n)
	for i := range lines {
		lines[i] = diffLine{kind: lineContext, text: "context"}
	}
	return lines
}

func makeNamedCommits() []commit {
	return []commit{
		{shortHash: "aaa1111", author: "John Doe", subject: "Fix login bug"},
		{shortHash: "bbb2222", author: "Jane Smith", subject: "Add user model"},
		{shortHash: "ccc3333", author: "John Doe", subject: "Update README"},
	}
}

func makeCommitsWithDays() []commit {
	return []commit{
		{shortHash: "aaa1111", author: "John", when: "1 day ago", subject: "Recent"},
		{shortHash: "bbb2222", author: "Jane", when: "5 days ago", subject: "Medium"},
		{shortHash: "ccc3333", author: "Bob", when: "20 days ago", subject: "Old"},
		{shortHash: "ddd4444", author: "Alice", when: "100 days ago", subject: "Very old"},
	}
}

