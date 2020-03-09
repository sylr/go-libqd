package sync

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestCancelableWaitGroupAdd(t *testing.T) {
	cap := 5
	ctx := context.Background()
	wg := NewCancelableWaitGroup(ctx, cap)
	c := make(chan int, cap*3)

	addFunc := func(i int) {
		wg.Add(i)
		c <- 1
	}

	for i := 1; i <= cap+1; i++ {
		//t.Logf("Loop #%d: wg.Add(1)", i)

		go addFunc(1)
		time.Sleep(10 * time.Millisecond)

		select {
		case <-c:
			t.Logf("Loop #%d: wg.Add(1) returned, it should have", i)
		case <-time.After(100 * time.Millisecond):
			if i <= cap {
				t.Errorf("Loop #%d: wg.Add(1) hangs but should not", i)
			}
		}
	}
}

func TestCancelableWaitGroupDone(t *testing.T) {
	cap := 8
	context := context.Background()
	wg := NewCancelableWaitGroup(context, cap)
	c := make(chan int, cap*3)

	addFunc := func(i int) {
		wg.Add(i)
		c <- 1
	}

	doneFunc := func() {
		wg.Done()
		c <- 1
	}

	for i := 1; i <= cap+3; i++ {
		//t.Logf("Loop #%d: wg.Add(1)", i)

		go addFunc(1)
		time.Sleep(10 * time.Millisecond)

		select {
		case <-c:
			t.Logf("Loop #%d: wg.Add(1) returned, it should have", i)
		case <-time.After(100 * time.Millisecond):
			if i <= cap {
				t.Errorf("Loop #%d: wg.Add(1) hangs but should not", i)
			} else {
				t.Logf("Loop #%d: wg.Add(1) hangs, it should", i)
			}
		}
	}

	for i := wg.cur; i > 0; i-- {
		//t.Logf("Reverse Loop #%d: wg.Done()", i)

		go doneFunc()
		time.Sleep(10 * time.Millisecond)

		select {
		case <-c:
			t.Logf("Reverse Loop #%d: wg.Done() returned, it should have", i)
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Reverse Loop #%d: wg.Done() hangs but should not", i)
		}
	}
}

func TestCancelableWaitGroupWait(t *testing.T) {
	cap := 8
	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)

	wg := NewCancelableWaitGroup(ctx, cap)
	c := make(chan int, cap*3)
	r := make(chan int, cap*3)

	addFunc := func(i int) {
		wg.Add(i)
		c <- 1
	}

	doneFunc := func() {
		wg.Done()
		c <- 1
	}

	waitFunc := func() {
		wg.Wait()
		r <- 1
	}

	// Test wg.Add(cap)
	go addFunc(cap)
	time.Sleep(10 * time.Microsecond)

	select {
	case <-c:
		t.Logf("wg.Add(%d) returned, it should have", cap)
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("wg.Add(%d) hangs but should not", cap)
	}

	// Test wg.Wait() that should not return
	go waitFunc()
	time.Sleep(10 * time.Microsecond)

	select {
	case <-r:
		t.Errorf("wg.Wait() returned but should not have")
	case <-time.After(100 * time.Millisecond):
		t.Logf("wg.Wait() hangs, it should")
	}

	// Test wg.Done()
	for i := 0; i < cap; i++ {
		//t.Logf("Loop #%d: wg.Done() wg.cur=%d", i, wg.cur)
		doneFunc()
	}

	select {
	case <-r:
		t.Logf("wg.Wait() returned, it should have")
	case <-time.After(100 * time.Millisecond):
		t.Errorf("wg.Wait() hangs but it should not")
	}

	// Test wg.Wait() that should return after cancel
	for i := 0; i < cap; i++ {
		//t.Logf("Loop #%d: wg.Add(1) wg.cur=%d", i, wg.cur)
		addFunc(1)
	}

	go waitFunc()
	time.Sleep(10 * time.Microsecond)

	t.Logf("Calling ctx.cancelFunc()")
	cancelFunc()

	select {
	case <-r:
		t.Logf("wg.Wait() returned, it should have")
	case <-time.After(1000 * time.Millisecond):
		t.Errorf("wg.Wait() hangs but it should not")
	}
}

// -- benchmarks ---------------------------------------------------------------

func BenchmarkCancelableWaitGroup(b *testing.B) {
	const cap = 1000
	ctx := context.Background()
	wg := NewCancelableWaitGroup(ctx, cap)

	for i := 0; i < b.N; i++ {
		for j := 0; j < cap; j++ {
			wg.Add(1)
		}

		for j := 0; j < cap; j++ {
			wg.Done()
		}
	}
}

func BenchmarkCancelableWaitGroupConcurent(b *testing.B) {
	// We test cap/2 iterations of Add() and Done() but because of concurrency
	// we need the wg to be twice the number of iterations to not fall into the
	// case where Done() is called more times than Add() and throws a panic.
	const cap = 2000
	ctx := context.Background()
	wg := NewCancelableWaitGroup(ctx, cap)
	w := sync.WaitGroup{}
	b.StopTimer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.Add(2)
		wg.Add(cap / 2)

		b.StartTimer()

		go func() {
			for j := 0; j < cap/2; j++ {
				wg.Add(1)
			}
			w.Done()
		}()

		go func() {
			for j := 0; j < cap/2; j++ {
				wg.Done()
			}
			w.Done()
		}()

		w.Wait()

		b.StopTimer()

		wg.Add(-cap / 2)
		wg.Wait()
	}
}
