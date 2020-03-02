package sync

import (
	"context"
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

	for i := cap; i >= -3; i-- {
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
