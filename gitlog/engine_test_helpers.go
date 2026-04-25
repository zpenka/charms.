package gitlog

import "testing"

// Common test helpers to reduce duplication across 370+ tests

// TestFixture provides reusable test data
type TestFixture struct {
	Commits        []commit
	CommitCount    int
	AuthorCount    int
	SampleFiles    []string
	SampleAuthors  []string
}

// NewTestFixture creates standard test fixture
func NewTestFixture() *TestFixture {
	return &TestFixture{
		Commits: []commit{
			{hash: "aaa1111111111111111111111111111111111", shortHash: "aaa1111", author: "Alice", subject: "Add feature", when: "1 day ago"},
			{hash: "bbb2222222222222222222222222222222222", shortHash: "bbb2222", author: "Bob", subject: "Fix bug", when: "2 days ago"},
			{hash: "ccc3333333333333333333333333333333333", shortHash: "ccc3333", author: "Charlie", subject: "Refactor code", when: "3 days ago"},
			{hash: "ddd4444444444444444444444444444444444", shortHash: "ddd4444", author: "Alice", subject: "Update docs", when: "4 days ago"},
			{hash: "eee5555555555555555555555555555555555", shortHash: "eee5555", author: "Eve", subject: "Add tests", when: "5 days ago"},
		},
		CommitCount:   5,
		AuthorCount:   4,
		SampleFiles:   []string{"main.go", "utils.go", "config.yaml", "README.md"},
		SampleAuthors: []string{"Alice", "Bob", "Charlie", "Eve"},
	}
}

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotEqual checks if two values are not equal
func AssertNotEqual(t *testing.T, a, b interface{}, message string) {
	if a == b {
		t.Errorf("%s: values should not be equal", message)
	}
}

// AssertTrue checks if value is true
func AssertTrue(t *testing.T, condition bool, message string) {
	if !condition {
		t.Errorf("%s: expected true", message)
	}
}

// AssertFalse checks if value is false
func AssertFalse(t *testing.T, condition bool, message string) {
	if condition {
		t.Errorf("%s: expected false", message)
	}
}

// AssertNotNil checks if value is not nil
func AssertNotNil(t *testing.T, value interface{}, message string) {
	if value == nil {
		t.Errorf("%s: expected non-nil value", message)
	}
}

// AssertNil checks if value is nil
func AssertNil(t *testing.T, value interface{}, message string) {
	if value != nil {
		t.Errorf("%s: expected nil value", message)
	}
}

// AssertLen checks if slice/array length matches
func AssertLen(t *testing.T, items interface{}, expectedLen int, message string) {
	switch v := items.(type) {
	case []string:
		if len(v) != expectedLen {
			t.Errorf("%s: expected length %d, got %d", message, expectedLen, len(v))
		}
	case []commit:
		if len(v) != expectedLen {
			t.Errorf("%s: expected length %d, got %d", message, expectedLen, len(v))
		}
	default:
		t.Errorf("%s: unsupported type for length check", message)
	}
}

// AssertStringContains checks if string contains substring
func AssertStringContains(t *testing.T, str, substr, message string) {
	found := false
	for i := 0; i < len(str)-len(substr)+1; i++ {
		if str[i:i+len(substr)] == substr {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("%s: string '%s' does not contain '%s'", message, str, substr)
	}
}

// AssertIntRange checks if value is within range
func AssertIntRange(t *testing.T, value, min, max int, message string) {
	if value < min || value > max {
		t.Errorf("%s: expected value between %d and %d, got %d", message, min, max, value)
	}
}

// AssertFloatRange checks if float is within range
func AssertFloatRange(t *testing.T, value, min, max float64, message string) {
	if value < min || value > max {
		t.Errorf("%s: expected value between %f and %f, got %f", message, min, max, value)
	}
}

// AssertMapContains checks if map has key
func AssertMapContains(t *testing.T, m map[string]interface{}, key, message string) {
	if _, exists := m[key]; !exists {
		t.Errorf("%s: map does not contain key '%s'", message, key)
	}
}

// AssertSliceContains checks if slice contains value
func AssertSliceContains(t *testing.T, items []string, item, message string) {
	found := false
	for _, s := range items {
		if s == item {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("%s: slice does not contain '%s'", message, item)
	}
}

// TestCategory represents a logical grouping of tests
type TestCategory struct {
	Name  string
	Tests int
}

// Categories organized by feature area
var TestCategories = map[string]*TestCategory{
	"Core":                 &TestCategory{"Core Parsing & Navigation", 0},
	"Search":               &TestCategory{"Search & Filtering", 0},
	"Visualization":        &TestCategory{"Visualization & Graph", 0},
	"Stash":                &TestCategory{"Stash & Reflog", 0},
	"Analytics":            &TestCategory{"Analytics & Stats", 0},
	"BisectRecovery":       &TestCategory{"Bisect & Recovery", 0},
	"CodeQuality":          &TestCategory{"Code Quality Analysis", 0},
	"Analysis":             &TestCategory{"Commit Analysis", 0},
	"Workflows":            &TestCategory{"Advanced Workflows", 0},
	"AIInsights":           &TestCategory{"AI-Powered Insights", 0},
	"Compliance":           &TestCategory{"Compliance & Security", 0},
	"Release":              &TestCategory{"Release & Versioning", 0},
	"Performance":          &TestCategory{"Performance", 0},
}
