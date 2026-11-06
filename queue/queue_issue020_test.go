package queue

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"

	"golang.org/x/time/rate"
	"time"
)

// TestShutdownResolvesEnqueueAndPendingRequests avoids panic and blocked waiters.
func TestShutdownResolvesEnqueueAndPendingRequests(t *testing.T) {
	q := NewRequestQueue(1, 1)
	first := q.Enqueue(func(_ *http.Client, done chan<- RequestResult) { done <- RequestResult{} })
	select {
	case <-first:
	case <-time.After(time.Second):
		t.Fatal("first job did not start")
	}
	pending := q.Enqueue(func(_ *http.Client, done chan<- RequestResult) { done <- RequestResult{} })
	q.Shutdown()
	select {
	case result := <-pending:
		if result.Err == nil {
			t.Fatal("pending request succeeded after shutdown")
		}
	case <-time.After(time.Second):
		t.Fatal("pending request hung after shutdown")
	}
	late := q.Enqueue(func(_ *http.Client, done chan<- RequestResult) { done <- RequestResult{} })
	select {
	case result := <-late:
		if result.Err == nil {
			t.Fatal("late request accepted")
		}
	case <-time.After(time.Second):
		t.Fatal("late request hung")
	}
}

// TestQueueCancellationInterruptsRunningAndQueuedWork uses channels as barriers.
func TestQueueCancellationInterruptsRunningAndQueuedWork(t *testing.T) {
	q := NewRequestQueue(100, 100)
	started := make(chan struct{})
	exited := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	result := q.EnqueueContext(ctx, func(ctx context.Context, _ *http.Client) RequestResult {
		close(started)
		<-ctx.Done()
		close(exited)
		return RequestResult{Err: ctx.Err()}
	})
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		cancel()
		t.Fatal("queue job did not start")
	}
	cancel()
	select {
	case got := <-result:
		if !errors.Is(got.Err, context.Canceled) {
			t.Fatalf("got %v", got.Err)
		}
	case <-time.After(time.Second):
		t.Fatal("caller cancellation hung")
	}
	select {
	case <-exited:
	case <-time.After(time.Second):
		t.Fatal("job ignored cancellation")
	}
	q.Shutdown()
	q.Shutdown()
}

// TestQueueShutdownConcurrentEnqueues resolves all callers without closing a live channel.
func TestQueueShutdownConcurrentEnqueues(t *testing.T) {
	q := NewRequestQueue(100, 100)
	var wg sync.WaitGroup
	for i := 0; i < 150; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := q.EnqueueContext(context.Background(), func(ctx context.Context, _ *http.Client) RequestResult {
				<-ctx.Done()
				return RequestResult{Err: ctx.Err()}
			})
			select {
			case <-r:
			case <-time.After(2 * time.Second):
				t.Error("request hung")
			}
		}()
	}
	q.Shutdown()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("enqueue goroutines leaked")
	}
}

// TestQueuedCallerCancellationDoesNotBlockLaterWork exercises cancellation while rate limited.
func TestQueuedCallerCancellationDoesNotBlockLaterWork(t *testing.T) {
	q := NewRequestQueue(1, 1)
	defer q.Shutdown()
	first := q.EnqueueContext(context.Background(), func(context.Context, *http.Client) RequestResult { return RequestResult{} })
	select {
	case <-first:
	case <-time.After(time.Second):
		t.Fatal("initial request hung")
	}
	ctx, cancel := context.WithCancel(context.Background())
	second := q.EnqueueContext(ctx, func(context.Context, *http.Client) RequestResult {
		t.Error("canceled job executed")
		return RequestResult{}
	})
	cancel()
	select {
	case result := <-second:
		if !errors.Is(result.Err, context.Canceled) {
			t.Fatalf("got %v", result.Err)
		}
	case <-time.After(time.Second):
		t.Fatal("canceled request blocked")
	}
	q.limiter.SetLimit(rate.Inf)
	third := q.EnqueueContext(context.Background(), func(context.Context, *http.Client) RequestResult { return RequestResult{} })
	select {
	case result := <-third:
		if result.Err != nil {
			t.Fatal(result.Err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("subsequent request blocked")
	}
}
