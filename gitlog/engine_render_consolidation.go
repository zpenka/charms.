package gitlog

import (
	"fmt"
	"strings"
)

// Consolidated rendering functions to reduce duplication
// All render*UI functions follow this pattern for consistency

// RenderConfig holds common rendering parameters
type RenderConfig struct {
	Title       string
	Items       []string
	HasStatus   bool
	StatusMap   map[string]string
	ShowIndices bool
	MaxItems    int
}

// RenderStandardUI renders a standard feature UI with title and items
func RenderStandardUI(config RenderConfig) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== %s ===\n", config.Title))

	items := config.Items
	if config.MaxItems > 0 && len(items) > config.MaxItems {
		items = items[:config.MaxItems]
	}

	for i, item := range items {
		if config.ShowIndices {
			sb.WriteString(fmt.Sprintf("%d: ", i))
		}

		if config.HasStatus && config.StatusMap != nil {
			status := config.StatusMap[item]
			if status == "ok" || status == "done" || status == "resolved" {
				sb.WriteString("✓ ")
			} else if status == "error" || status == "failed" {
				sb.WriteString("✗ ")
			}
		}

		sb.WriteString(item)
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderAnalysisUI renders analysis data with metrics
func RenderAnalysisUI(title string, data map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== %s ===\n", title))

	for key, value := range data {
		switch v := value.(type) {
		case int:
			sb.WriteString(fmt.Sprintf("%s: %d\n", key, v))
		case float64:
			sb.WriteString(fmt.Sprintf("%s: %.2f%%\n", key, v*100))
		case string:
			sb.WriteString(fmt.Sprintf("%s: %s\n", key, v))
		case []string:
			if len(v) > 0 {
				sb.WriteString(fmt.Sprintf("%s: %v\n", key, v))
			}
		default:
			sb.WriteString(fmt.Sprintf("%s: %v\n", key, v))
		}
	}

	return sb.String()
}

// RenderDataGrid renders tabular data
func RenderDataGrid(title string, headers []string, rows [][]string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== %s ===\n", title))

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print headers
	for i, h := range headers {
		sb.WriteString(fmt.Sprintf("%-*s  ", colWidths[i], h))
	}
	sb.WriteString("\n")

	// Print separator
	for _, w := range colWidths {
		sb.WriteString(strings.Repeat("-", w) + "  ")
	}
	sb.WriteString("\n")

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			sb.WriteString(fmt.Sprintf("%-*s  ", colWidths[i], cell))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderMetricBar renders a metric with visual indicator
func RenderMetricBar(name string, value int, max int, width int) string {
	if max <= 0 {
		return fmt.Sprintf("%s: %d\n", name, value)
	}

	percent := (value * 100) / max
	filled := (value * width) / max
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return fmt.Sprintf("%s: [%s] %d/%d (%d%%)\n", name, bar, value, max, percent)
}

// RenderComparisonTable renders side-by-side comparison
func RenderComparisonTable(title string, label1, label2 string, items map[string][2]interface{}) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== %s ===\n", title))
	sb.WriteString(fmt.Sprintf("%-30s %-20s %-20s\n", "Metric", label1, label2))
	sb.WriteString(strings.Repeat("-", 70) + "\n")

	for key, values := range items {
		sb.WriteString(fmt.Sprintf("%-30s %-20v %-20v\n", key, values[0], values[1]))
	}

	return sb.String()
}

// RenderSummaryStats renders key statistics
func RenderSummaryStats(title string, stats map[string]int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== %s ===\n", title))

	for key, val := range stats {
		sb.WriteString(fmt.Sprintf("%s: %d\n", key, val))
	}

	return sb.String()
}

// RenderErrorList renders a list of errors/issues
func RenderErrorList(title string, errors []string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== %s ===\n", title))

	if len(errors) == 0 {
		sb.WriteString("✓ No issues found\n")
		return sb.String()
	}

	for i, err := range errors {
		sb.WriteString(fmt.Sprintf("%d. ✗ %s\n", i+1, err))
	}

	return sb.String()
}
