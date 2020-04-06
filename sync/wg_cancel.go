package sync

import (
	"context"
	"sync"
	"sync/atomic"
)

type CancelableWaitGroup struct {
	mu      sync.Mutex
	cond    *sync.Cond
	cap     int
	cur     int
	context context.Context
	done    int32
}

// NewCancelableWaitGroup returns a Waiter that can be canceled wia a context.
// If the context is canceled thatn Waiter.Wait() will return even if Waiter.Done()
// has not been called as much as Waiter.Add().
func NewCancelableWaitGroup(context context.Context, cap int) *CancelableWaitGroup {
	wg := &CancelableWaitGroup{
		mu:      sync.Mutex{},
		cap:     cap,
		context: context,
		done:    0,
	}

	wg.cond = sync.NewCond(&wg.mu)

	go func() {
		//nolint:gosimple
		select {
		case <-wg.context.Done():
			atomic.SwapInt32(&wg.done, 1)
			wg.cond.Broadcast()
		}
	}()

	return wg
}

// Add adds n tasks and does not return until wg.cur + n <= wg.cap.
// However, it does return if the context is canceled.
func (wg *CancelableWaitGroup) Add(n int) {
	if n > wg.cap {
		panic("libqd/sync: tryng to Add more than cap")
	}

	wg.mu.Lock()
	defer wg.mu.Unlock()

	for (wg.cur + n) > wg.cap {
		wg.cond.Wait()

		if atomic.LoadInt32(&wg.done) == 1 {
			return
		}
	}

	wg.cur += n
}

// Done decreases wg.cur
func (wg *CancelableWaitGroup) Done() {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	if wg.cur == 0 {
		panic("libqd/sync: called Done more than Add")
	}

	if atomic.LoadInt32(&wg.done) == 1 {
		return
	}

	wg.cur--

	wg.cond.Signal()
}

// Wait waits unitl wg.cur == 0 or if the context is canceled.
func (wg *CancelableWaitGroup) Wait() {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	for wg.cur > 0 {
		wg.cond.Wait()

		if atomic.LoadInt32(&wg.done) == 1 {
			return
		}
	}
}
