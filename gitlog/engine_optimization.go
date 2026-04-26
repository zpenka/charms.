package gitlog

// Performance optimization patterns and utilities

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	Hits       int
	Misses     int
	Evictions  int
	Size       int
	MaxSize    int
}

// GetHitRate returns cache hit percentage
func (c CacheMetrics) GetHitRate() float64 {
	total := c.Hits + c.Misses
	if total == 0 {
		return 0
	}
	return float64(c.Hits) / float64(total)
}

// ShouldEvict returns true if cache should evict oldest entry
func (c CacheMetrics) ShouldEvict() bool {
	return c.Size >= c.MaxSize
}

// LazyLoader provides lazy initialization pattern
type LazyLoader struct {
	loaded   bool
	data     interface{}
	loadFunc func() interface{}
}

// NewLazyLoader creates a lazy loader with initialization function
func NewLazyLoader(fn func() interface{}) *LazyLoader {
	return &LazyLoader{
		loaded:   false,
		data:     nil,
		loadFunc: fn,
	}
}

// Load initializes data if not already loaded
func (l *LazyLoader) Load() interface{} {
	if !l.loaded && l.loadFunc != nil {
		l.data = l.loadFunc()
		l.loaded = true
	}
	return l.data
}

// IsLoaded checks if data has been initialized
func (l *LazyLoader) IsLoaded() bool {
	return l.loaded
}

// Reset clears the loaded data
func (l *LazyLoader) Reset() {
	l.loaded = false
	l.data = nil
}

// MemoryPool provides object pooling for allocation reduction
type MemoryPool struct {
	items chan interface{}
	cap   int
}

// NewMemoryPool creates a new object pool
func NewMemoryPool(size int) *MemoryPool {
	return &MemoryPool{
		items: make(chan interface{}, size),
		cap:   size,
	}
}

// Get retrieves an object from pool or creates new
func (mp *MemoryPool) Get(fn func() interface{}) interface{} {
	select {
	case item := <-mp.items:
		return item
	default:
		return fn()
	}
}

// Put returns object to pool if space available
func (mp *MemoryPool) Put(item interface{}) {
	select {
	case mp.items <- item:
	default:
		// Pool full, discard
	}
}

// BatchProcessor handles bulk operations efficiently
type BatchProcessor struct {
	batchSize int
	timeout   int
	items     []interface{}
}

// NewBatchProcessor creates batch processor
func NewBatchProcessor(size int, timeout int) *BatchProcessor {
	return &BatchProcessor{
		batchSize: size,
		timeout:   timeout,
		items:     make([]interface{}, 0, size),
	}
}

// Add adds item to batch
func (bp *BatchProcessor) Add(item interface{}) bool {
	bp.items = append(bp.items, item)
	return len(bp.items) >= bp.batchSize
}

// IsFull checks if batch is ready
func (bp *BatchProcessor) IsFull() bool {
	return len(bp.items) >= bp.batchSize
}

// Get returns items and resets
func (bp *BatchProcessor) Get() []interface{} {
	items := bp.items
	bp.items = make([]interface{}, 0, bp.batchSize)
	return items
}

// CircularBuffer provides fixed-size ring buffer
type CircularBuffer struct {
	data  []interface{}
	head  int
	tail  int
	count int
}

// NewCircularBuffer creates circular buffer
func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		data: make([]interface{}, size),
	}
}

// Push adds item, overwrites oldest if full
func (cb *CircularBuffer) Push(item interface{}) {
	cb.data[cb.tail] = item
	cb.tail = (cb.tail + 1) % len(cb.data)

	if cb.count < len(cb.data) {
		cb.count++
	} else {
		cb.head = (cb.head + 1) % len(cb.data)
	}
}

// GetAll returns all items in order
func (cb *CircularBuffer) GetAll() []interface{} {
	result := make([]interface{}, cb.count)
	for i := 0; i < cb.count; i++ {
		result[i] = cb.data[(cb.head+i)%len(cb.data)]
	}
	return result
}

// Size returns number of items
func (cb *CircularBuffer) Size() int {
	return cb.count
}

// RateLimiter provides rate limiting
type RateLimiter struct {
	tokens    int
	maxTokens int
	lastTime  int64
	interval  int64
}

// NewRateLimiter creates rate limiter
func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{
		maxTokens: rps,
		tokens:    rps,
		interval:  1000 / int64(rps),
	}
}

// Allow checks if operation is allowed
func (rl *RateLimiter) Allow() bool {
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}
	return false
}

// Metrics provides general performance metrics tracking operations by name
type Metrics struct {
	Operations map[string]int    // operation name -> count
	Successes  map[string]int    // operation name -> success count
	Failures   map[string]int    // operation name -> failure count
	TotalTime  map[string]int64  // operation name -> total duration
}

// NewMetrics creates a new metrics tracker
func NewMetrics() *Metrics {
	return &Metrics{
		Operations: make(map[string]int),
		Successes:  make(map[string]int),
		Failures:   make(map[string]int),
		TotalTime:  make(map[string]int64),
	}
}

// RecordOperation records an operation result
func (m *Metrics) RecordOperation(name string, success bool) {
	m.Operations[name]++
	if success {
		m.Successes[name]++
	} else {
		m.Failures[name]++
	}
}

// GetSuccessRate returns success rate for an operation (0.0 to 1.0)
func (m *Metrics) GetSuccessRate(name string) float64 {
	total := m.Operations[name]
	if total == 0 {
		return 0
	}
	return float64(m.Successes[name]) / float64(total)
}
