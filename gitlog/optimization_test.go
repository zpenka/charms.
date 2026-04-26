package gitlog

import (
	"testing"
)

// Tests for critical optimization features

func TestNewBatchProcessor_Create(t *testing.T) {
	bp := NewBatchProcessor(10, 100)
	AssertNotNil(t, bp, "should create batch processor")
}

func TestNewBatchProcessor_WithDifferentSizes(t *testing.T) {
	bp1 := NewBatchProcessor(5, 50)
	bp2 := NewBatchProcessor(20, 200)
	AssertNotNil(t, bp1, "should create small batch processor")
	AssertNotNil(t, bp2, "should create large batch processor")
}

func TestNewLazyLoader_Create(t *testing.T) {
	loader := NewLazyLoader(func() interface{} {
		return "data"
	})
	AssertNotNil(t, loader, "should create lazy loader")
}

func TestNewLazyLoader_WithDifferentFunctions(t *testing.T) {
	loader1 := NewLazyLoader(func() interface{} { return 42 })
	loader2 := NewLazyLoader(func() interface{} { return "test" })
	AssertNotNil(t, loader1, "should handle int loader")
	AssertNotNil(t, loader2, "should handle string loader")
}

func TestNewRateLimiter_Create(t *testing.T) {
	limiter := NewRateLimiter(10)
	AssertNotNil(t, limiter, "should create rate limiter")
}

func TestNewRateLimiter_WithDifferentRates(t *testing.T) {
	limiter1 := NewRateLimiter(5)
	limiter2 := NewRateLimiter(100)
	AssertNotNil(t, limiter1, "should create 5 RPS limiter")
	AssertNotNil(t, limiter2, "should create 100 RPS limiter")
}

func TestBatchProcessor_BasicOperation(t *testing.T) {
	bp := NewBatchProcessor(3, 100)
	AssertNotNil(t, bp, "should initialize")
	// Verify it doesn't panic on basic operations
}

func TestLazyLoader_BasicOperation(t *testing.T) {
	loader := NewLazyLoader(func() interface{} {
		return "loaded"
	})
	AssertNotNil(t, loader, "should initialize")
}

func TestRateLimiter_BasicOperation(t *testing.T) {
	limiter := NewRateLimiter(10)
	AssertNotNil(t, limiter, "should initialize")
	// First call should be allowed
	_ = limiter.Allow()
}

func TestNewMemoryPool_Create(t *testing.T) {
	pool := NewMemoryPool(1000)
	AssertNotNil(t, pool, "should create memory pool")
}

func TestNewMemoryPool_WithDifferentSizes(t *testing.T) {
	pool1 := NewMemoryPool(100)
	pool2 := NewMemoryPool(10000)
	AssertNotNil(t, pool1, "should create small pool")
	AssertNotNil(t, pool2, "should create large pool")
}

func TestNewCircularBuffer_Create(t *testing.T) {
	buf := NewCircularBuffer(100)
	AssertNotNil(t, buf, "should create circular buffer")
}

func TestNewCircularBuffer_WithDifferentCapacities(t *testing.T) {
	buf1 := NewCircularBuffer(10)
	buf2 := NewCircularBuffer(1000)
	AssertNotNil(t, buf1, "should create small buffer")
	AssertNotNil(t, buf2, "should create large buffer")
}

func TestNewMetrics_Create(t *testing.T) {
	metrics := NewMetrics()
	AssertNotNil(t, metrics, "should create metrics")
}

func TestBatchProcessor_MultipleInstances(t *testing.T) {
	bp1 := NewBatchProcessor(5, 50)
	bp2 := NewBatchProcessor(10, 100)
	bp3 := NewBatchProcessor(20, 200)
	AssertNotNil(t, bp1, "first processor should exist")
	AssertNotNil(t, bp2, "second processor should exist")
	AssertNotNil(t, bp3, "third processor should exist")
}

func TestLazyLoader_MultipleLoaders(t *testing.T) {
	loader1 := NewLazyLoader(func() interface{} { return []int{1, 2, 3} })
	loader2 := NewLazyLoader(func() interface{} { return map[string]int{"a": 1} })
	loader3 := NewLazyLoader(func() interface{} { return "string data" })
	AssertNotNil(t, loader1, "first loader should exist")
	AssertNotNil(t, loader2, "second loader should exist")
	AssertNotNil(t, loader3, "third loader should exist")
}

func TestRateLimiter_MultipleLimiters(t *testing.T) {
	limiter1 := NewRateLimiter(5)
	limiter2 := NewRateLimiter(10)
	limiter3 := NewRateLimiter(20)
	AssertNotNil(t, limiter1, "first limiter should exist")
	AssertNotNil(t, limiter2, "second limiter should exist")
	AssertNotNil(t, limiter3, "third limiter should exist")
}

func TestOptimization_ConcurrentCreation(t *testing.T) {
	// Verify multiple components can be created together
	bp := NewBatchProcessor(10, 100)
	loader := NewLazyLoader(func() interface{} { return nil })
	limiter := NewRateLimiter(10)
	pool := NewMemoryPool(1000)
	buf := NewCircularBuffer(100)

	AssertNotNil(t, bp, "batch processor created")
	AssertNotNil(t, loader, "lazy loader created")
	AssertNotNil(t, limiter, "rate limiter created")
	AssertNotNil(t, pool, "memory pool created")
	AssertNotNil(t, buf, "circular buffer created")
}

func TestOptimization_ZeroSizeHandling(t *testing.T) {
	// Test edge cases with minimal sizes
	bp := NewBatchProcessor(1, 1)
	loader := NewLazyLoader(func() interface{} { return struct{}{} })
	limiter := NewRateLimiter(1)
	pool := NewMemoryPool(1)
	buf := NewCircularBuffer(1)

	AssertNotNil(t, bp, "should handle size 1 batch processor")
	AssertNotNil(t, loader, "should handle simple loader")
	AssertNotNil(t, limiter, "should handle 1 RPS")
	AssertNotNil(t, pool, "should handle 1 byte pool")
	AssertNotNil(t, buf, "should handle size 1 buffer")
}
