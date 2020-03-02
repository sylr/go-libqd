package sync

import (
	"context"
	"sync"
)

type CancelableWaitGroup struct {
	mu      sync.Mutex
	cond    *sync.Cond
	cap     int
	cur     int
	context context.Context
	done    bool
	doneMu  sync.RWMutex
}

// NewCancelableWaitGroup returns a Waiter that can be canceled wia a context.
// If the context is canceled thatn Waiter.Wait() will return even if Waiter.Done()
// has not been called as much as Waiter.Add().
func NewCancelableWaitGroup(context context.Context, cap int) *CancelableWaitGroup {
	wg := &CancelableWaitGroup{
		mu:      sync.Mutex{},
		cap:     cap,
		context: context,
		done:    false,
		doneMu:  sync.RWMutex{},
	}

	wg.cond = sync.NewCond(&wg.mu)

	go func() {
		select {
		case <-wg.context.Done():
			wg.doneMu.Lock()
			defer wg.doneMu.Unlock()
			wg.done = true

			wg.cond.Broadcast()
		}
	}()

	return wg
}

// Add adds n tasks and does not return until wg.cur + n <= wg.cap.
// However, it does return if the context is canceled.
func (wg *CancelableWaitGroup) Add(n int) {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	for (wg.cur + n) > wg.cap {
		wg.cond.Wait()
	}

	done := func() bool {
		wg.doneMu.RLock()
		defer wg.doneMu.RUnlock()
		return wg.done
	}()

	if done {
		return
	}

	wg.cur += n
}

// Done decreases wg.cur
func (wg *CancelableWaitGroup) Done() {
	wg.mu.Lock()
	defer wg.mu.Unlock()

	if wg.cur == 0 {
		return
	}

	done := func() bool {
		wg.doneMu.RLock()
		defer wg.doneMu.RUnlock()
		return wg.done
	}()

	if done {
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

		done := func() bool {
			wg.doneMu.RLock()
			defer wg.doneMu.RUnlock()
			return wg.done
		}()

		if done {
			return
		}
	}
}
