package usage

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const (
	chanSize       = 8192 // buffered channel capacity
	flushInterval  = time.Second
	flushThreshold = 500 // events before early flush
	maxConcurrent  = 4   // max parallel flush goroutines
)

type bufferKey struct {
	workspaceID string
	quota       Quota
}

type event struct {
	workspaceID string
	quota       Quota
	delta       int64
}

// buffer aggregates high-frequency events in memory and batch-flushes to PostgreSQL.
// Single reader goroutine, bounded concurrent writers.
type buffer struct {
	events chan event
	done   chan struct{}
	meter  *Meter
	wg     sync.WaitGroup
	sem    chan struct{} // flush concurrency limiter
}

func newBuffer(m *Meter) *buffer {
	b := &buffer{
		events: make(chan event, chanSize),
		done:   make(chan struct{}),
		meter:  m,
		sem:    make(chan struct{}, maxConcurrent),
	}
	go b.run()
	return b
}

// send queues an event. Drops silently if the buffer is full.
func (b *buffer) send(workspaceID string, q Quota, n int64) {
	select {
	case b.events <- event{workspaceID: workspaceID, quota: q, delta: n}:
	default:
		slog.Warn("usage buffer full, dropping event",
			"workspace_id", workspaceID, "quota", q.String(), "delta", n)
	}
}

// stop signals shutdown, waits for all pending flushes to complete.
func (b *buffer) stop() {
	close(b.events)
	<-b.done
	b.wg.Wait()
}

// run is the single aggregator goroutine.
func (b *buffer) run() {
	defer close(b.done)

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	pending := make(map[bufferKey]int64)
	count := 0

	flush := func() {
		if len(pending) == 0 {
			return
		}
		snapshot := pending
		pending = make(map[bufferKey]int64)
		count = 0

		// Acquire semaphore to bound concurrent flushes.
		b.sem <- struct{}{}
		b.wg.Go(func() {
			defer func() { <-b.sem }()
			b.flushSnapshot(snapshot)
		})
	}

	for {
		select {
		case e, ok := <-b.events:
			if !ok {
				flush()
				return
			}
			pending[bufferKey{e.workspaceID, e.quota}] += e.delta
			count++
			if count >= flushThreshold {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (b *buffer) flushSnapshot(snapshot map[bufferKey]int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := b.meter.flush(ctx, snapshot); err != nil {
		slog.Error("usage buffer flush failed", "error", err, "batches", len(snapshot))
	}
}
