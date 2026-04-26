package gitlog

import (
	"strings"
	"testing"
)

// --- Migration Tests: Replacing old render functions with consolidated ones ---

func TestMigrate_AuthorStats(t *testing.T) {
	// Old approach: renderAuthorStats(stats map[string]int, width int) string
	// New approach: RenderSummaryStats(title string, stats map[string]int) string

	stats := map[string]int{
		"Alice": 42,
		"Bob":   28,
		"Charlie": 15,
	}

	result := RenderSummaryStats("Author Statistics", stats)

	if !strings.Contains(result, "Author Statistics") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, "Alice") || !strings.Contains(result, "42") {
		t.Errorf("result should contain author stats")
	}
}

func TestMigrate_BookmarkList(t *testing.T) {
	// Old approach: renderBookmarkList(m model, width int) string
	// New approach: RenderConfig with title and items

	items := []string{"abc123 - Fix login bug", "def456 - Add user model"}
	config := RenderConfig{
		Title:       "Bookmarks",
		Items:       items,
		ShowIndices: true,
	}

	result := RenderStandardUI(config)

	if !strings.Contains(result, "Bookmarks") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, "Fix login bug") {
		t.Errorf("result should contain bookmark items")
	}
	if !strings.Contains(result, "0:") {
		t.Errorf("result should show indices")
	}
}

func TestMigrate_StatsList(t *testing.T) {
	// Old approach: renderTimeStats(stats map[string]int, width int)
	// New approach: RenderSummaryStats

	stats := map[string]int{
		"recent":     45,
		"past_week":  32,
		"older":      18,
	}

	result := RenderSummaryStats("Time-based Statistics", stats)

	if !strings.Contains(result, "Time-based Statistics") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, "recent") {
		t.Errorf("result should contain stats categories")
	}
}

func TestMigrate_UIList(t *testing.T) {
	// Migration pattern for simple UI lists
	// Old: renderExtensionFilterUI(m model, width int) -> returns list with status
	// New: RenderConfig with HasStatus = true

	config := RenderConfig{
		Title:     "Extension Filters",
		Items:     []string{".go", ".js", ".py"},
		HasStatus: true,
		StatusMap: map[string]string{
			".go": "ok",
			".js": "ok",
			".py": "error",
		},
	}

	result := RenderStandardUI(config)

	if !strings.Contains(result, "Extension Filters") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, ".go") {
		t.Errorf("result should contain extensions")
	}
}

func TestMigrate_AnalyticsMetrics(t *testing.T) {
	// Migration for analytics data
	// Old: renderAnalyticsPanel(m model) -> map of various metrics
	// New: RenderAnalysisUI

	data := map[string]interface{}{
		"total_commits": 150,
		"file_changes":  89,
		"avg_commits":   3.5,
		"top_authors":   []string{"Alice", "Bob"},
	}

	result := RenderAnalysisUI("Analytics", data)

	if !strings.Contains(result, "Analytics") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, "total_commits") {
		t.Errorf("result should contain metric names")
	}
}

func TestMigrate_DataTable(t *testing.T) {
	// Migration for tabular displays
	// Old: Custom table rendering functions
	// New: RenderDataGrid

	headers := []string{"Function", "Count", "Complexity"}
	rows := [][]string{
		{"parseCommits", "1", "low"},
		{"filterCommits", "2", "medium"},
		{"renderUI", "73", "high"},
	}

	result := RenderDataGrid("Functions Analysis", headers, rows)

	if !strings.Contains(result, "Functions Analysis") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, "Function") {
		t.Errorf("result should contain headers")
	}
	if !strings.Contains(result, "parseCommits") {
		t.Errorf("result should contain data")
	}
}

func TestMigrate_ErrorHandling(t *testing.T) {
	// Migration for error/issue lists
	// New: RenderErrorList

	errors := []string{
		"Duplicate function: renderAuthorStats",
		"Unused function: renderOldUI",
		"Performance issue in renderLargeData",
	}

	result := RenderErrorList("Code Analysis Issues", errors)

	if !strings.Contains(result, "Code Analysis Issues") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, "Duplicate function") {
		t.Errorf("result should contain error items")
	}
}

func TestMigrate_ComparisonMetrics(t *testing.T) {
	// Migration for before/after comparisons
	// New: RenderComparisonTable

	items := map[string][2]interface{}{
		"Total Functions":        {73, 15},
		"Code Duplication":       {45, 0},
		"Average Function Lines": {32, 20},
	}

	result := RenderComparisonTable("Refactoring Impact", "Before", "After", items)

	if !strings.Contains(result, "Refactoring Impact") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, "Before") && !strings.Contains(result, "After") {
		t.Errorf("result should contain comparison labels")
	}
}

func TestMigrate_MetricsWithBars(t *testing.T) {
	// Migration for metrics with visual indicators
	// New: RenderMetricBar

	result := RenderMetricBar("Code Coverage", 85, 100, 30)

	if !strings.Contains(result, "Code Coverage") {
		t.Errorf("result should contain metric name")
	}
	if !strings.Contains(result, "85%") {
		t.Errorf("result should contain percentage")
	}
}
