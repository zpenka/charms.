package gitlog

import (
	"strings"
	"testing"
)
func TestRenderBookmarkMarker_Bookmarked(t *testing.T) {
	m := model{cursor: 0, commits: []commit{{shortHash: "abc"}}, bookmarks: []string{"abc"}}
	marker := renderBookmarkMarker(m, 0)
	if marker == "" {
		t.Error("bookmarked commit should have marker")
	}
}

func TestRenderBookmarkMarker_NotBookmarked(t *testing.T) {
	m := model{cursor: 0, commits: []commit{{shortHash: "abc"}}, bookmarks: []string{}}
	marker := renderBookmarkMarker(m, 0)
	if marker != "" {
		t.Errorf("non-bookmarked should be empty, got %q", marker)
	}
}

func TestRenderStatsBadgeInList_TruncatesLong(t *testing.T) {
	stats := commitStatistics{filesChanged: 100, insertions: 999, deletions: 888}
	badge := renderStatsBadgeInList(stats, 10)
	if len(badge) > 12 {
		t.Errorf("should truncate for width 10, got len=%d", len(badge))
	}
}

func TestRenderLineCommentMarker_HasComment(t *testing.T) {
	m := model{comments: map[int]string{5: "needs review"}}
	marker := renderLineCommentMarker(m, 5)
	if marker == "" {
		t.Error("line with comment should have marker")
	}
}

func TestRenderLineCommentMarker_NoComment(t *testing.T) {
	m := model{comments: map[int]string{}}
	marker := renderLineCommentMarker(m, 5)
	if marker != "" {
		t.Errorf("line without comment should be empty, got %q", marker)
	}
}

func TestRenderWithStats_IncludesBadges(t *testing.T) {
	m := model{showStatsBadge: true, lastStats: commitStatistics{filesChanged: 3, insertions: 10, deletions: 5}}
	output := renderCommitRowWithStats(m, 0, 50)
	if output == "" {
		t.Error("should render non-empty output")
	}
}

func TestRenderGraphView_ShowsArt(t *testing.T) {
	m := model{showGraph: true, commitGraph: []graphNode{
		{hash: "abc1234", isMerge: false},
	}}
	output := renderGraphView(m, 50)
	if output == "" {
		t.Error("should render non-empty graph")
	}
}

func TestRenderFileTimeline_Empty(t *testing.T) {
	timeline := renderFileTimeline([]commit{}, "test.go", 50)
	if timeline == "" {
		t.Error("should generate non-empty timeline")
	}
}

func TestRenderAsciiGraph_SingleCommit(t *testing.T) {
	graph := []graphNode{
		{hash: "abc", depth: 0, isMerge: false},
	}
	art := renderAsciiGraph(graph)
	if art == "" {
		t.Error("should generate non-empty ASCII art")
	}
}

func TestRenderAuthorStats_ShowsCount(t *testing.T) {
	stats := map[string]int{
		"John": 10,
		"Jane": 5,
		"Bob":  3,
	}
	output := renderAuthorStats(stats, 50)
	if output == "" {
		t.Error("should render non-empty output")
	}
	if !strings.Contains(output, "John") {
		t.Errorf("should show author, got %q", output)
	}
}

