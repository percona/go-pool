package pool

import (
	. "gopkg.in/check.v1"
	"sync"
	"time"
)

type TestSuite struct {
}

var _ = Suite(&TestSuite{})

func (s *TestSuite) TestOverflow(t *C) {
	size := 3                // Size of the pool
	overflowSize := size + 1 // Size greater than pool size

	// We try to create more elements than pool size
	// by waiting inside of newFunc.
	// This could be done with time.After(),
	// but that's not reliable method, so we will use chan/sync.WaitGroup
	// to synchronize between gorotuines

	// wgNew ensures that each p.Get() is in state of running newFunc
	wgNew := &sync.WaitGroup{}
	// wgBlocked ensures that each p.Get() didn't left newFunc until desired
	wgBlocked := &sync.WaitGroup{}

	// Let's create our NewDynamicPool and inject our newFunc
	newFunc := func() (interface{}, error) {
		wgNew.Done()
		wgBlocked.Wait()
		return nil, nil
	}
	p := NewDynamicPool(uint(size), newFunc, nil)

	// wgDone ensures each p.Get() finished
	wgDone := &sync.WaitGroup{}

	// Let's try to get more elements from the pool than the pool size
	for i := 1; i <= overflowSize; i++ {
		wgNew.Add(1)
		wgBlocked.Add(1)
		wgDone.Add(1)
		go func() {
			_, err := p.Get(time.Duration(1 * time.Millisecond))
			t.Check(err, IsNil)
			wgDone.Done()
		}()
	}

	// Wait until each of p.Get() is running newFunc
	wgNew.Wait()
	// After reaching that point we hit race condition
	// pool p already runs more newFunc than it should

	// Let's leave newFunc in all p.Get()
	for i := 1; i <= overflowSize; i++ {
		wgBlocked.Done()
	}

	// Let's wait until all p.Get() returns result
	wgDone.Wait()

	// Time to check our expectations
	expectedFree := uint(0)    // We tried to get more elements from the pool than size allows, so free elements should be 0
	expectedSize := uint(size) // Reported size by pool should be the same as when we created it
	t.Check(p.Free(), Equals, expectedFree, Commentf("got %d; expected %d", p.Free(), expectedFree))
	t.Check(p.Size(), Equals, expectedSize, Commentf("got %d; expected %d", p.Size(), expectedSize))
	// p.Free()... got 18446744073709551615; expected 0
	// p.Size()... got 4; expected 3

	// But that's not all, right now pool p is pretty much broken,
	// You can create as many elements as you want without even invoking race conditions
	// because as you may see p.Free() returns now max size of the uint.

	// Showdown!
	// Let's clear our newFunc, it's no longer needed to invoke race conditions
	p.newFunc = nil

	// Can we create 100 more elements?
	more := 100
	for i := 0; i < more; i++ {
		_, err := p.Get(time.Duration(1 * time.Millisecond))
		t.Check(err, IsNil)
	}
	t.Check(p.Free(), Equals, expectedFree, Commentf("got %d; expected %d", p.Free(), expectedFree))
	t.Check(p.Size(), Equals, expectedSize, Commentf("got %d; expected %d", p.Size(), expectedSize))
	// Yes we can, create 100 more elements
	// p.Size() ... got 104; expected 3
}
