package queue

import (
	"context"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RequestResult struct {
	Body []byte
	Err  error
}
type RequestFunc func(client *http.Client, done chan<- RequestResult)

type request struct {
	ctx    context.Context
	run    func(context.Context, *http.Client) RequestResult
	result chan RequestResult
	once   sync.Once
	stop   func() bool
}

type RequestQueue struct {
	client  *http.Client
	limiter *rate.Limiter
	queue   chan *request
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	mu      sync.Mutex
	stopped bool
	pending map[*request]struct{}
}

// NewRequestQueue creates a rate-limited queue whose pending requests can be cancelled.
func NewRequestQueue(rps int, burst int) *RequestQueue {
	ctx, cancel := context.WithCancel(context.Background())
	q := &RequestQueue{client: &http.Client{Timeout: 10 * time.Second}, limiter: rate.NewLimiter(rate.Every(time.Second/time.Duration(rps)), burst), queue: make(chan *request, 100), ctx: ctx, cancel: cancel, pending: make(map[*request]struct{})}
	q.start()
	return q
}

// finish publishes one result and removes its pending registration exactly once.
func (q *RequestQueue) finish(r *request, result RequestResult) {
	r.once.Do(func() {
		q.mu.Lock()
		delete(q.pending, r)
		stop := r.stop
		q.mu.Unlock()
		if stop != nil {
			stop()
		}
		r.result <- result
	})
}

// EnqueueContext returns one result for success, caller cancellation or shutdown.
func (q *RequestQueue) EnqueueContext(ctx context.Context, job func(context.Context, *http.Client) RequestResult) <-chan RequestResult {
	r := &request{ctx: ctx, run: job, result: make(chan RequestResult, 1)}
	q.mu.Lock()
	if q.stopped {
		q.mu.Unlock()
		r.result <- RequestResult{Err: context.Canceled}
		return r.result
	}
	q.pending[r] = struct{}{}
	r.stop = context.AfterFunc(ctx, func() { q.finish(r, RequestResult{Err: ctx.Err()}) })
	q.mu.Unlock()
	select {
	case q.queue <- r:
	case <-ctx.Done():
		q.finish(r, RequestResult{Err: ctx.Err()})
	case <-q.ctx.Done():
		q.finish(r, RequestResult{Err: context.Canceled})
	}
	return r.result
}

// Enqueue retains the legacy callback API while ensuring its waiter can be cancelled on shutdown.
func (q *RequestQueue) Enqueue(job RequestFunc) <-chan RequestResult {
	return q.EnqueueContext(context.Background(), func(ctx context.Context, client *http.Client) RequestResult {
		done := make(chan RequestResult, 1)
		go job(client, done)
		select {
		case result := <-done:
			return result
		case <-ctx.Done():
			return RequestResult{Err: ctx.Err()}
		}
	})
}

// start dispatches only live work and uses cancellation for queued and running jobs.
func (q *RequestQueue) start() {
	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		for {
			select {
			case <-q.ctx.Done():
				return
			case r := <-q.queue:
				if err := q.ctx.Err(); err != nil {
					q.finish(r, RequestResult{Err: err})
					continue
				}
				if err := r.ctx.Err(); err != nil {
					q.finish(r, RequestResult{Err: err})
					continue
				}
				jobCtx, cancel := context.WithCancel(r.ctx)
				stop := context.AfterFunc(q.ctx, cancel)
				if err := q.limiter.Wait(jobCtx); err != nil {
					stop()
					cancel()
					q.finish(r, RequestResult{Err: err})
					continue
				}
				if err := jobCtx.Err(); err != nil {
					stop()
					cancel()
					q.finish(r, RequestResult{Err: err})
					continue
				}
				go func() { defer cancel(); defer stop(); q.finish(r, r.run(jobCtx, q.client)) }()
			}
		}
	}()
}

// Shutdown refuses new work and resolves every pending waiter without closing a live channel.
func (q *RequestQueue) Shutdown() {
	q.mu.Lock()
	if q.stopped {
		q.mu.Unlock()
		q.wg.Wait()
		return
	}
	q.stopped = true
	pending := make([]*request, 0, len(q.pending))
	for r := range q.pending {
		pending = append(pending, r)
	}
	q.cancel()
	q.mu.Unlock()
	for _, r := range pending {
		q.finish(r, RequestResult{Err: context.Canceled})
	}
	q.wg.Wait()
}
