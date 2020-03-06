package sync

import (
	"testing"
	"time"
)

func TestBoundedWaitGroupAdd(t *testing.T) {
	cap := 5
	wg := NewBoundedWaitGroup(cap)
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
			t.Logf("Loop #%d: wg.Add(1) returned, it should", i)
		case <-time.After(100 * time.Millisecond):
			if i <= cap {
				t.Errorf("Loop #%d: wg.Add(1) hangs but should not", i)
			}
		}
	}
}

func TestBoundedWaitGroupDone(t *testing.T) {
	cap := 8
	wg := NewBoundedWaitGroup(cap)
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

	for i := cap; i >= -1; i-- {
		//t.Logf("Loop #%d: wg.Done()", i)

		go doneFunc()
		time.Sleep(10 * time.Millisecond)

		select {
		case <-c:
			t.Logf("Loop #%d: wg.Done() returned, it should have", i)
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Loop #%d: wg.Done() hangs but should not", i)
		}
	}
}

// -- benchmarks ---------------------------------------------------------------

func BenchmarkBoundedWaitGroup(b *testing.B) {
	const cap = 1000
	wg := NewBoundedWaitGroup(cap)

	for i := 0; i < b.N; i++ {
		for j := 0; j < cap; j++ {
			wg.Add(1)
		}

		for j := 0; j < cap; j++ {
			wg.Done()
		}
	}
}
