package gitlog

import (
	"strings"
	"testing"
)

// --- Consolidated Render Config Tests ---

func TestRenderConfig_SimpleList(t *testing.T) {
	config := RenderConfig{
		Title: "Test List",
		Items: []string{"item1", "item2", "item3"},
	}

	result := RenderStandardUI(config)

	if !strings.Contains(result, "Test List") {
		t.Errorf("result should contain title, got: %s", result)
	}
	if !strings.Contains(result, "item1") {
		t.Errorf("result should contain items, got: %s", result)
	}
}

func TestRenderConfig_WithStatus(t *testing.T) {
	config := RenderConfig{
		Title:     "Status List",
		Items:     []string{"feature", "bug"},
		HasStatus: true,
		StatusMap: map[string]string{
			"feature": "ok",
			"bug":     "error",
		},
	}

	result := RenderStandardUI(config)

	if !strings.Contains(result, "Status List") {
		t.Errorf("result should contain title")
	}
	if len(result) == 0 {
		t.Errorf("result should not be empty")
	}
}

func TestRenderConfig_WithIndices(t *testing.T) {
	config := RenderConfig{
		Title:       "Indexed List",
		Items:       []string{"first", "second", "third"},
		ShowIndices: true,
	}

	result := RenderStandardUI(config)

	if !strings.Contains(result, "0") || !strings.Contains(result, "1") {
		t.Errorf("result should contain indices, got: %s", result)
	}
}

func TestRenderConfig_MaxItems(t *testing.T) {
	config := RenderConfig{
		Title:    "Limited List",
		Items:    []string{"a", "b", "c", "d", "e"},
		MaxItems: 3,
	}

	result := RenderStandardUI(config)

	lines := strings.Split(result, "\n")
	itemCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" && !strings.Contains(line, "===") {
			itemCount++
		}
	}
	if itemCount > 3 {
		t.Errorf("result should be limited to max items, got %d", itemCount)
	}
}

func TestRenderAnalysisUI_WithMetrics(t *testing.T) {
	data := map[string]interface{}{
		"commits":  42,
		"files":    15,
		"changes":  3.14,
		"authors":  []string{"alice", "bob"},
	}

	result := RenderAnalysisUI("Analysis", data)

	if !strings.Contains(result, "Analysis") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, "commits") || !strings.Contains(result, "42") {
		t.Errorf("result should contain metrics")
	}
}

func TestRenderDataGrid_WithHeaders(t *testing.T) {
	headers := []string{"Hash", "Author", "Message"}
	rows := [][]string{
		{"abc123", "Alice", "Fix bug"},
		{"def456", "Bob", "Add feature"},
	}

	result := RenderDataGrid("Commits Grid", headers, rows)

	if !strings.Contains(result, "Commits Grid") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, "Hash") || !strings.Contains(result, "Author") {
		t.Errorf("result should contain headers")
	}
	if !strings.Contains(result, "abc123") || !strings.Contains(result, "Alice") {
		t.Errorf("result should contain data rows")
	}
}

func TestRenderMetricBar_Percentage(t *testing.T) {
	result := RenderMetricBar("Coverage", 75, 100, 20)

	if !strings.Contains(result, "Coverage") {
		t.Errorf("result should contain metric name")
	}
	if !strings.Contains(result, "75") || !strings.Contains(result, "100") {
		t.Errorf("result should contain values")
	}
	if !strings.Contains(result, "75%") {
		t.Errorf("result should contain percentage")
	}
}

func TestRenderSummaryStats(t *testing.T) {
	stats := map[string]int{
		"commits":  42,
		"authors":  5,
		"branches": 3,
	}

	result := RenderSummaryStats("Summary", stats)

	if !strings.Contains(result, "Summary") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, "commits") {
		t.Errorf("result should contain stats")
	}
}

func TestRenderErrorList_WithErrors(t *testing.T) {
	errors := []string{"File not found", "Permission denied"}

	result := RenderErrorList("Issues", errors)

	if !strings.Contains(result, "Issues") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, "File not found") {
		t.Errorf("result should contain error messages")
	}
	if !strings.Contains(result, "✗") {
		t.Errorf("result should contain error indicator")
	}
}

func TestRenderErrorList_NoErrors(t *testing.T) {
	result := RenderErrorList("Issues", []string{})

	if !strings.Contains(result, "No issues") {
		t.Errorf("result should indicate no issues")
	}
}

func TestRenderComparisonTable(t *testing.T) {
	items := map[string][2]interface{}{
		"files":   {10, 15},
		"commits": {50, 75},
	}

	result := RenderComparisonTable("Comparison", "Before", "After", items)

	if !strings.Contains(result, "Comparison") {
		t.Errorf("result should contain title")
	}
	if !strings.Contains(result, "Before") || !strings.Contains(result, "After") {
		t.Errorf("result should contain labels")
	}
}
