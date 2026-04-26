package gitlog

import (
	"strings"
	"testing"
)
func TestTruncate_Short(t *testing.T) {
	AssertEqual(t, "hello", truncate("hello", 10), "should not truncate")
}

func TestTruncate_Exact(t *testing.T) {
	AssertEqual(t, "hello", truncate("hello", 5), "exact length should not truncate")
}

func TestTruncate_Long(t *testing.T) {
	AssertEqual(t, "hello …", truncate("hello world", 7), "should truncate with ellipsis")
}

func TestTruncate_One(t *testing.T) {
	AssertEqual(t, "…", truncate("hi", 1), "max=1 should return ellipsis")
}

func TestTruncate_Zero(t *testing.T) {
	AssertEqual(t, "", truncate("hi", 0), "max=0 should return empty")
}

func TestFirstWord_WithSpace(t *testing.T) {
	AssertEqual(t, "John", firstWord("John Doe"), "should extract first word")
}

func TestFirstWord_NoSpace(t *testing.T) {
	AssertEqual(t, "John", firstWord("John"), "single word should return as-is")
}

func TestCopyAsPatch_Empty(t *testing.T) {
	patch := copyAsPatch("abc1234", []diffLine{})
	if patch == "" {
		t.Error("should generate non-empty patch")
	}
}

func TestCopyAsPatch_ContainsHash(t *testing.T) {
	lines := []diffLine{{lineAdded, "+new line"}}
	patch := copyAsPatch("abc1234", lines)
	if !strings.Contains(patch, "abc1234") {
		t.Errorf("patch should contain hash, got %q", patch)
	}
}

func TestCopyAsPatch_ContainsDiff(t *testing.T) {
	lines := []diffLine{
		{lineMeta, "diff --git a/test.go b/test.go"},
		{lineAdded, "+new line"},
	}
	patch := copyAsPatch("abc1234", lines)
	if !strings.Contains(patch, "+new line") {
		t.Errorf("patch should contain diff content, got %q", patch)
	}
}

func TestExtractReviewers_None(t *testing.T) {
	reviewers := extractReviewers("normal commit message")
	if len(reviewers) != 0 {
		t.Errorf("no reviewers should return empty, got %d", len(reviewers))
	}
}

func TestExtractReviewers_Mentioned(t *testing.T) {
	msg := "Fix: handle edge case\n\nReviewed-by: John Smith <john@example.com>"
	reviewers := extractReviewers(msg)
	if len(reviewers) != 1 {
		t.Errorf("expected 1 reviewer, got %d", len(reviewers))
	}
}

func TestExtractReviewers_Multiple(t *testing.T) {
	msg := "Feature: new API\n\nReviewed-by: Alice <alice@ex.com>\nReviewed-by: Bob <bob@ex.com>"
	reviewers := extractReviewers(msg)
	if len(reviewers) != 2 {
		t.Errorf("expected 2 reviewers, got %d", len(reviewers))
	}
}

func TestExtractCoAuthors_None(t *testing.T) {
	authors := extractCoAuthors("simple commit message")
	if len(authors) != 0 {
		t.Errorf("no co-authors should return empty, got %d", len(authors))
	}
}

func TestExtractCoAuthors_Single(t *testing.T) {
	msg := "Fix bug\n\nCo-authored-by: Jane Smith <jane@example.com>"
	authors := extractCoAuthors(msg)
	if len(authors) != 1 {
		t.Errorf("expected 1 co-author, got %d", len(authors))
	}
	if !strings.Contains(authors[0], "Jane") {
		t.Errorf("should contain name, got %q", authors[0])
	}
}

func TestExtractCoAuthors_Multiple(t *testing.T) {
	msg := "Fix bug\n\nCo-authored-by: Jane <jane@example.com>\nCo-authored-by: Bob <bob@example.com>"
	authors := extractCoAuthors(msg)
	if len(authors) != 2 {
		t.Errorf("expected 2 co-authors, got %d", len(authors))
	}
}

func TestDiffCache_StoresResult(t *testing.T) {
	cache := newDiffCache(10)
	lines := []diffLine{{lineAdded, "+test"}}
	cache.set("abc1234", lines)
	cached, ok := cache.get("abc1234")
	if !ok {
		t.Error("cache should have stored diff")
	}
	if len(cached) != 1 {
		t.Errorf("cache should return same data, got %d lines", len(cached))
	}
}

func TestDiffCache_EvictsOldest(t *testing.T) {
	cache := newDiffCache(2)
	cache.set("aaa", []diffLine{{lineAdded, "+a"}})
	cache.set("bbb", []diffLine{{lineAdded, "+b"}})
	cache.set("ccc", []diffLine{{lineAdded, "+c"}})
	_, ok := cache.get("aaa")
	if ok {
		t.Error("cache should evict oldest entry")
	}
}

func TestRegexCache_Compiles(t *testing.T) {
	cache := newRegexCache(10)
	re, err := cache.compile("test.*pattern")
	if err != nil {
		t.Errorf("should compile regex, got %v", err)
	}
	if !re.MatchString("test something pattern") {
		t.Error("compiled regex should match")
	}
}

func TestRegexCache_Reuses(t *testing.T) {
	cache := newRegexCache(10)
	re1, _ := cache.compile("test")
	re2, _ := cache.compile("test")
	if re1 != re2 {
		t.Error("cache should return same regex object")
	}
}

func TestCircularBuffer_StoresAndRetrievesItems(t *testing.T) {
	buffer := NewCircularBuffer(5)
	
	buffer.Push("commit1")
	buffer.Push("commit2")
	buffer.Push("commit3")
	
	items := buffer.GetAll()
	AssertLen(t, items, 3, "should store 3 items")
	AssertEqual(t, 3, buffer.Size(), "size should be 3")
}

func TestCircularBuffer_Wraps(t *testing.T) {
	buffer := NewCircularBuffer(3)
	
	buffer.Push("a")
	buffer.Push("b")
	buffer.Push("c")
	buffer.Push("d") // Should wrap and overwrite "a"
	
	items := buffer.GetAll()
	AssertLen(t, items, 3, "should maintain max capacity")
}

func TestCurrentFile_Empty(t *testing.T) {
	m := model{}
	if currentFile(m) != "" {
		t.Error("expected empty string with no fileItems")
	}
}

func TestCurrentFile_AtStart(t *testing.T) {
	m := model{
		fileItems:  []fileItem{{path: "a.go", diffIdx: 0}, {path: "b.go", diffIdx: 20}},
		diffOffset: 0,
	}
	if got := currentFile(m); got != "a.go" {
		t.Errorf("expected a.go, got %q", got)
	}
}

func TestCurrentFile_PastBoundary(t *testing.T) {
	m := model{
		fileItems:  []fileItem{{path: "a.go", diffIdx: 0}, {path: "b.go", diffIdx: 20}},
		diffOffset: 25,
	}
	if got := currentFile(m); got != "b.go" {
		t.Errorf("expected b.go, got %q", got)
	}
}

func TestCurrentFile_ExactlyAtBoundary(t *testing.T) {
	m := model{
		fileItems:  []fileItem{{path: "a.go", diffIdx: 0}, {path: "b.go", diffIdx: 20}},
		diffOffset: 20,
	}
	if got := currentFile(m); got != "b.go" {
		t.Errorf("expected b.go at exact boundary, got %q", got)
	}
}

func TestMetrics_TracksOperations(t *testing.T) {
	metrics := NewMetrics()

	metrics.RecordOperation("filter", true)
	metrics.RecordOperation("filter", true)
	metrics.RecordOperation("filter", false)

	rate := metrics.GetSuccessRate("filter")
	AssertTrue(t, rate > 0 && rate <= 1, "success rate should be between 0 and 1")
}

func TestMemoryPool_Reuses(t *testing.T) {
	pool := NewMemoryPool(5)
	
	// Get object from empty pool
	obj1 := pool.Get(func() interface{} {
		return &commit{hash: "new1"}
	})
	AssertNotNil(t, obj1, "should create object")
	
	// Put back to pool
	pool.Put(obj1)
	
	// Get again should reuse
	obj2 := pool.Get(func() interface{} {
		return &commit{hash: "new2"}
	})
	AssertEqual(t, obj1, obj2, "should reuse pooled object")
}

func TestMemoryPool_CreateWhenEmpty(t *testing.T) {
	pool := NewMemoryPool(1)
	
	// First object
	obj1 := pool.Get(func() interface{} {
		return &commit{hash: "first"}
	})
	AssertNotNil(t, obj1, "should create first object")
	
	// Get second without returning first
	obj2 := pool.Get(func() interface{} {
		return &commit{hash: "second"}
	})
	AssertNotNil(t, obj2, "should create new object when pool empty")
	AssertNotEqual(t, obj1, obj2, "should be different objects")
}

func TestStatCache_Memoizes(t *testing.T) {
	cache := newStatCache(10)
	lines := []diffLine{
		{lineAdded, "+test"},
		{lineRemoved, "-old"},
	}
	stats1 := cache.getOrCompute("abc", lines)
	stats2 := cache.getOrCompute("abc", lines)
	if stats1.insertions != stats2.insertions {
		t.Error("cached stats should be identical")
	}
}

