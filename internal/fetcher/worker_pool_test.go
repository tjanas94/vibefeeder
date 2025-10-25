package fetcher

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tjanas94/vibefeeder/internal/shared/database"
)

func TestNewWorkerPool(t *testing.T) {
	tests := []struct {
		name        string
		workerCount int
		logger      *slog.Logger
		description string
	}{
		{
			name:        "creates pool with custom logger",
			workerCount: 5,
			logger:      slog.Default(),
			description: "Should create WorkerPool with provided logger",
		},
		{
			name:        "uses default logger when nil",
			workerCount: 10,
			logger:      nil,
			description: "Should use slog.Default() when logger is nil",
		},
		{
			name:        "creates pool with single worker",
			workerCount: 1,
			logger:      slog.Default(),
			description: "Should create WorkerPool with single worker",
		},
		{
			name:        "creates pool with many workers",
			workerCount: 100,
			logger:      slog.Default(),
			description: "Should create WorkerPool with many workers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := NewWorkerPool(tt.workerCount, tt.logger)

			assert.NotNil(t, wp, tt.description)
			assert.NotNil(t, wp.logger, "logger should never be nil")
			assert.Equal(t, tt.workerCount, wp.workerCount)
			assert.Equal(t, tt.workerCount, cap(wp.semaphore))
		})
	}
}

func TestWorkerPool_ProcessFeeds_BasicOperation(t *testing.T) {
	wp := NewWorkerPool(3, slog.Default())
	ctx := context.Background()

	feeds := []database.PublicFeedsSelect{
		{Id: "feed-1", Url: "https://example.com/feed1"},
		{Id: "feed-2", Url: "https://example.com/feed2"},
		{Id: "feed-3", Url: "https://example.com/feed3"},
	}

	var processed atomic.Int32
	var mu sync.Mutex
	processedIDs := make([]string, 0)

	processFn := func(feed database.PublicFeedsSelect) {
		processed.Add(1)
		mu.Lock()
		processedIDs = append(processedIDs, feed.Id)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond) // Simulate some work
	}

	wp.ProcessFeeds(ctx, feeds, processFn)

	assert.Equal(t, int32(3), processed.Load(), "Should process all feeds")
	assert.Len(t, processedIDs, 3)
	assert.Contains(t, processedIDs, "feed-1")
	assert.Contains(t, processedIDs, "feed-2")
	assert.Contains(t, processedIDs, "feed-3")
}

func TestWorkerPool_ProcessFeeds_EmptyList(t *testing.T) {
	wp := NewWorkerPool(3, slog.Default())
	ctx := context.Background()

	var processed atomic.Int32
	processFn := func(feed database.PublicFeedsSelect) {
		processed.Add(1)
	}

	wp.ProcessFeeds(ctx, []database.PublicFeedsSelect{}, processFn)

	assert.Equal(t, int32(0), processed.Load(), "Should process zero feeds")
}

func TestWorkerPool_ProcessFeeds_NilList(t *testing.T) {
	wp := NewWorkerPool(3, slog.Default())
	ctx := context.Background()

	var processed atomic.Int32
	processFn := func(feed database.PublicFeedsSelect) {
		processed.Add(1)
	}

	// Should not panic with nil list
	assert.NotPanics(t, func() {
		wp.ProcessFeeds(ctx, nil, processFn)
	})

	assert.Equal(t, int32(0), processed.Load(), "Should process zero feeds")
}

func TestWorkerPool_ProcessFeeds_WorkerLimit(t *testing.T) {
	workerCount := 2
	wp := NewWorkerPool(workerCount, slog.Default())
	ctx := context.Background()

	feeds := []database.PublicFeedsSelect{
		{Id: "feed-1"},
		{Id: "feed-2"},
		{Id: "feed-3"},
		{Id: "feed-4"},
		{Id: "feed-5"},
	}

	var activeWorkers atomic.Int32
	var maxActiveWorkers atomic.Int32

	processFn := func(feed database.PublicFeedsSelect) {
		active := activeWorkers.Add(1)

		// Track maximum concurrent workers
		for {
			max := maxActiveWorkers.Load()
			if active <= max || maxActiveWorkers.CompareAndSwap(max, active) {
				break
			}
		}

		time.Sleep(50 * time.Millisecond) // Simulate work
		activeWorkers.Add(-1)
	}

	wp.ProcessFeeds(ctx, feeds, processFn)

	assert.LessOrEqual(t, int(maxActiveWorkers.Load()), workerCount,
		"Should never exceed worker limit")
}

func TestWorkerPool_ProcessFeeds_ContextCancellation(t *testing.T) {
	wp := NewWorkerPool(2, slog.Default())
	ctx, cancel := context.WithCancel(context.Background())

	feeds := make([]database.PublicFeedsSelect, 20)
	for i := 0; i < 20; i++ {
		feeds[i] = database.PublicFeedsSelect{Id: "feed-" + string(rune('0'+i))}
	}

	var processed atomic.Int32
	var started atomic.Int32

	processFn := func(feed database.PublicFeedsSelect) {
		started.Add(1)
		time.Sleep(100 * time.Millisecond) // Simulate slow work
		processed.Add(1)
	}

	// Cancel context after short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	wp.ProcessFeeds(ctx, feeds, processFn)
	elapsed := time.Since(start)

	// Should terminate quickly after cancellation
	assert.Less(t, elapsed, 500*time.Millisecond,
		"Should terminate quickly after context cancellation")

	// Should process fewer feeds than total due to cancellation
	assert.Less(t, int(processed.Load()), len(feeds),
		"Should process fewer feeds due to cancellation")

	// Started count should be at least processed count
	assert.GreaterOrEqual(t, int(started.Load()), int(processed.Load()),
		"Started count should be >= processed count")
}

func TestWorkerPool_ProcessFeeds_ContextCancellationWaitsForRunningWorkers(t *testing.T) {
	wp := NewWorkerPool(2, slog.Default())
	ctx, cancel := context.WithCancel(context.Background())

	feeds := make([]database.PublicFeedsSelect, 10)
	for i := 0; i < 10; i++ {
		feeds[i] = database.PublicFeedsSelect{Id: "feed-" + string(rune('0'+i))}
	}

	var completed atomic.Int32
	var started atomic.Int32

	processFn := func(feed database.PublicFeedsSelect) {
		started.Add(1)
		time.Sleep(150 * time.Millisecond) // Simulate work
		completed.Add(1)
	}

	// Cancel context immediately
	cancel()

	wp.ProcessFeeds(ctx, feeds, processFn)

	// All started workers should have completed
	assert.Equal(t, started.Load(), completed.Load(),
		"All started workers should complete even after cancellation")
}

func TestWorkerPool_ProcessFeeds_PanicRecovery(t *testing.T) {
	wp := NewWorkerPool(3, slog.Default())
	ctx := context.Background()

	feeds := []database.PublicFeedsSelect{
		{Id: "feed-1"},
		{Id: "feed-2-panic"},
		{Id: "feed-3"},
		{Id: "feed-4-panic"},
		{Id: "feed-5"},
	}

	var processed atomic.Int32

	processFn := func(feed database.PublicFeedsSelect) {
		if feed.Id == "feed-2-panic" || feed.Id == "feed-4-panic" {
			panic("simulated panic in worker")
		}
		processed.Add(1)
	}

	// Should not panic despite worker panics
	assert.NotPanics(t, func() {
		wp.ProcessFeeds(ctx, feeds, processFn)
	})

	// Should have processed the non-panicking feeds
	assert.Equal(t, int32(3), processed.Load(),
		"Should process non-panicking feeds successfully")
}

func TestWorkerPool_ProcessFeeds_MultiplePanics(t *testing.T) {
	wp := NewWorkerPool(5, slog.Default())
	ctx := context.Background()

	feeds := make([]database.PublicFeedsSelect, 10)
	for i := 0; i < 10; i++ {
		feeds[i] = database.PublicFeedsSelect{Id: "feed-" + string(rune('0'+i))}
	}

	processFn := func(feed database.PublicFeedsSelect) {
		// Every feed panics
		panic("panic in all workers")
	}

	// Should handle multiple panics without crashing
	assert.NotPanics(t, func() {
		wp.ProcessFeeds(ctx, feeds, processFn)
	})
}

func TestWorkerPool_ProcessFeeds_Concurrency(t *testing.T) {
	workerCount := 5
	wp := NewWorkerPool(workerCount, slog.Default())
	ctx := context.Background()

	feedCount := 50
	feeds := make([]database.PublicFeedsSelect, feedCount)
	for i := 0; i < feedCount; i++ {
		feeds[i] = database.PublicFeedsSelect{Id: "feed-" + string(rune('0'+i%10)) + string(rune('0'+i/10))}
	}

	var processed atomic.Int32
	processDelay := 20 * time.Millisecond

	processFn := func(feed database.PublicFeedsSelect) {
		time.Sleep(processDelay)
		processed.Add(1)
	}

	start := time.Now()
	wp.ProcessFeeds(ctx, feeds, processFn)
	elapsed := time.Since(start)

	// All feeds should be processed
	assert.Equal(t, int32(feedCount), processed.Load())

	// With concurrent processing, should be faster than sequential
	sequentialTime := time.Duration(feedCount) * processDelay
	assert.Less(t, elapsed, sequentialTime,
		"Concurrent processing should be faster than sequential")

	// Should utilize multiple workers (rough estimation)
	expectedMinTime := time.Duration(feedCount/workerCount) * processDelay
	assert.GreaterOrEqual(t, elapsed, expectedMinTime-100*time.Millisecond,
		"Should take at least the theoretical minimum time")
}

func TestWorkerPool_ProcessFeeds_OrderIndependence(t *testing.T) {
	wp := NewWorkerPool(3, slog.Default())
	ctx := context.Background()

	feeds := []database.PublicFeedsSelect{
		{Id: "feed-1"},
		{Id: "feed-2"},
		{Id: "feed-3"},
		{Id: "feed-4"},
		{Id: "feed-5"},
	}

	var mu sync.Mutex
	processOrder := make([]string, 0)

	processFn := func(feed database.PublicFeedsSelect) {
		// Random delay to increase chance of different ordering
		delay := time.Duration(len(feed.Id)) * 10 * time.Millisecond
		time.Sleep(delay)

		mu.Lock()
		processOrder = append(processOrder, feed.Id)
		mu.Unlock()
	}

	wp.ProcessFeeds(ctx, feeds, processFn)

	// All feeds should be processed
	assert.Len(t, processOrder, 5)

	// Order may differ from input order due to concurrency
	// Just verify all IDs are present
	for _, feed := range feeds {
		assert.Contains(t, processOrder, feed.Id)
	}
}

func TestWorkerPool_ProcessFeeds_SemaphoreReleaseOnPanic(t *testing.T) {
	workerCount := 2
	wp := NewWorkerPool(workerCount, slog.Default())
	ctx := context.Background()

	// First batch - all panic
	firstBatch := []database.PublicFeedsSelect{
		{Id: "panic-1"},
		{Id: "panic-2"},
	}

	processFn := func(feed database.PublicFeedsSelect) {
		time.Sleep(10 * time.Millisecond)
		panic("panic to test semaphore release")
	}

	wp.ProcessFeeds(ctx, firstBatch, processFn)

	// Semaphore should be fully released (all slots available)
	assert.Equal(t, 0, len(wp.semaphore), "Semaphore should be empty after processing")

	// Second batch - should process normally
	secondBatch := []database.PublicFeedsSelect{
		{Id: "normal-1"},
		{Id: "normal-2"},
	}

	var processed atomic.Int32
	normalProcessFn := func(feed database.PublicFeedsSelect) {
		processed.Add(1)
	}

	wp.ProcessFeeds(ctx, secondBatch, normalProcessFn)

	assert.Equal(t, int32(2), processed.Load(),
		"Should process second batch normally after panics")
}

func TestWorkerPool_ProcessFeeds_SingleWorker(t *testing.T) {
	wp := NewWorkerPool(1, slog.Default())
	ctx := context.Background()

	feeds := []database.PublicFeedsSelect{
		{Id: "feed-1"},
		{Id: "feed-2"},
		{Id: "feed-3"},
	}

	var mu sync.Mutex
	processOrder := make([]string, 0)

	processFn := func(feed database.PublicFeedsSelect) {
		mu.Lock()
		processOrder = append(processOrder, feed.Id)
		mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}

	wp.ProcessFeeds(ctx, feeds, processFn)

	// All feeds should be processed
	assert.Len(t, processOrder, 3)

	// With single worker, should process in order (mostly)
	// Note: Due to goroutine scheduling, order might still vary slightly
	// but we can at least verify all were processed
	assert.Contains(t, processOrder, "feed-1")
	assert.Contains(t, processOrder, "feed-2")
	assert.Contains(t, processOrder, "feed-3")
}

func TestWorkerPool_ProcessFeeds_ManyWorkers(t *testing.T) {
	workerCount := 100
	wp := NewWorkerPool(workerCount, slog.Default())
	ctx := context.Background()

	feedCount := 200
	feeds := make([]database.PublicFeedsSelect, feedCount)
	for i := 0; i < feedCount; i++ {
		feeds[i] = database.PublicFeedsSelect{Id: "feed"}
	}

	var processed atomic.Int32

	processFn := func(feed database.PublicFeedsSelect) {
		time.Sleep(10 * time.Millisecond)
		processed.Add(1)
	}

	start := time.Now()
	wp.ProcessFeeds(ctx, feeds, processFn)
	elapsed := time.Since(start)

	assert.Equal(t, int32(feedCount), processed.Load())

	// Should process very quickly with many workers
	assert.Less(t, elapsed, 1*time.Second)
}

func TestWorkerPool_ProcessFeeds_ComplexFeedData(t *testing.T) {
	wp := NewWorkerPool(3, slog.Default())
	ctx := context.Background()

	feeds := []database.PublicFeedsSelect{
		{
			Id:  "feed-1",
			Url: "https://example.com/feed1",
		},
		{
			Id:  "feed-2",
			Url: "https://example.com/feed2",
		},
	}

	var mu sync.Mutex
	processedFeeds := make([]database.PublicFeedsSelect, 0)

	processFn := func(feed database.PublicFeedsSelect) {
		mu.Lock()
		processedFeeds = append(processedFeeds, feed)
		mu.Unlock()
	}

	wp.ProcessFeeds(ctx, feeds, processFn)

	require.Len(t, processedFeeds, 2)

	// Verify feed data is passed correctly
	var foundFeed1, foundFeed2 bool
	for _, feed := range processedFeeds {
		if feed.Id == "feed-1" && feed.Url == "https://example.com/feed1" {
			foundFeed1 = true
		}
		if feed.Id == "feed-2" && feed.Url == "https://example.com/feed2" {
			foundFeed2 = true
		}
	}

	assert.True(t, foundFeed1, "feed-1 should be processed with correct data")
	assert.True(t, foundFeed2, "feed-2 should be processed with correct data")
}

func BenchmarkWorkerPool_ProcessFeeds_SmallPool(b *testing.B) {
	wp := NewWorkerPool(5, slog.Default())
	ctx := context.Background()

	feeds := make([]database.PublicFeedsSelect, 100)
	for i := 0; i < 100; i++ {
		feeds[i] = database.PublicFeedsSelect{Id: "feed"}
	}

	processFn := func(feed database.PublicFeedsSelect) {
		// Minimal work
		_ = feed.Id
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wp.ProcessFeeds(ctx, feeds, processFn)
	}
}

func BenchmarkWorkerPool_ProcessFeeds_LargePool(b *testing.B) {
	wp := NewWorkerPool(50, slog.Default())
	ctx := context.Background()

	feeds := make([]database.PublicFeedsSelect, 100)
	for i := 0; i < 100; i++ {
		feeds[i] = database.PublicFeedsSelect{Id: "feed"}
	}

	processFn := func(feed database.PublicFeedsSelect) {
		// Minimal work
		_ = feed.Id
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wp.ProcessFeeds(ctx, feeds, processFn)
	}
}
