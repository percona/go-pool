/*
	Copyright (c) 2014, Percona LLC and/or its affiliates. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package pool_test

import (
	"errors"
	pool "github.com/percona/go-pool"
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

func Test(t *testing.T) { TestingT(t) }

type D struct {
	n int
}

var errFoo = errors.New("foo")

type StaticTestSuite struct {
}

var _ = Suite(&StaticTestSuite{})

func (s *StaticTestSuite) TestGetAndFree(t *C) {
	size := uint(3)
	p := pool.NewStaticPool(size, nil, nil)

	t.Check(p.Free(), Equals, size)
	t.Check(p.Size(), Equals, uint(3))

	_, err := p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, IsNil)
	t.Check(p.Free(), Equals, size-1)
	t.Check(p.Size(), Equals, uint(3))

	_, err = p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, IsNil)
	t.Check(p.Free(), Equals, size-2)
	t.Check(p.Size(), Equals, uint(3))

	_, err = p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, IsNil)
	t.Check(p.Free(), Equals, size-3)
	t.Check(p.Size(), Equals, uint(3))

	_, err = p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, Equals, pool.ErrTimeout)
	t.Check(p.Free(), Equals, uint(0))
	t.Check(p.Size(), Equals, uint(3))
}

func (s *StaticTestSuite) TestFuncsAndPut(t *C) {
	size := uint(2)
	n := 0
	newFunc := func() (interface{}, error) {
		n++
		d := &D{n: n}
		return d, nil
	}
	put := []*D{}
	putFunc := func(v interface{}) {
		d := v.(*D)
		put = append(put, d)
	}
	p := pool.NewStaticPool(size, newFunc, putFunc)

	v, _ := p.Get(time.Duration(1 * time.Millisecond))
	d := v.(*D)
	t.Check(d.n, Equals, 1)

	p.Put(d)
	t.Check(put, DeepEquals, []*D{&D{n: 1}})
}

func (s *StaticTestSuite) TestOverflow(t *C) {
	size := uint(2)
	p := pool.NewStaticPool(size, nil, nil)

	err := p.Put("one too many")
	t.Check(err, Equals, pool.ErrOverflow)
}

func (s *StaticTestSuite) TestNewFunc(t *C) {
	// newFunc should only be called once for new items.
	called := 0
	newFunc := func() (interface{}, error) {
		called++
		return 1, nil
	}
	p := pool.NewStaticPool(2, newFunc, nil)
	for i := 0; i < 5; i++ {
		d, err := p.Get(time.Duration(1 * time.Millisecond))
		t.Assert(err, IsNil)
		err = p.Put(d)
		t.Assert(err, IsNil)
	}
	t.Check(called, Equals, 2)
}

func (s *StaticTestSuite) TestNewFuncErr(t *C) {
	newFunc := func() (interface{}, error) {
		return nil, errFoo
	}
	p := pool.NewStaticPool(2, newFunc, nil)
	d, err := p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, Equals, errFoo)
	t.Check(d, IsNil)
}

// --------------------------------------------------------------------------

type DynamicTestSuite struct {
}

var _ = Suite(&DynamicTestSuite{})

func (s *DynamicTestSuite) TestGetAndFree(t *C) {
	size := uint(3)
	p := pool.NewDynamicPool(size, nil, nil)

	t.Check(p.Free(), Equals, size)
	t.Check(p.Size(), Equals, uint(0))

	_, err := p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, IsNil)
	t.Check(p.Free(), Equals, size-1)
	t.Check(p.Size(), Equals, uint(1))

	_, err = p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, IsNil)
	t.Check(p.Free(), Equals, size-2)
	t.Check(p.Size(), Equals, uint(2))

	_, err = p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, IsNil)
	t.Check(p.Free(), Equals, size-3)
	t.Check(p.Size(), Equals, uint(3))

	_, err = p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, Equals, pool.ErrTimeout)
	t.Check(p.Free(), Equals, uint(0))
	t.Check(p.Size(), Equals, uint(3))
}

func (s *DynamicTestSuite) TestFuncsAndPut(t *C) {
	size := uint(2)
	n := 0
	newFunc := func() (interface{}, error) {
		n++
		d := &D{n: n}
		return d, nil
	}
	put := []*D{}
	putFunc := func(v interface{}) {
		d := v.(*D)
		put = append(put, d)
	}
	p := pool.NewDynamicPool(size, newFunc, putFunc)

	v, _ := p.Get(time.Duration(1 * time.Millisecond))
	d := v.(*D)
	t.Check(d.n, Equals, 1)

	p.Put(d)
	t.Check(put, DeepEquals, []*D{&D{n: 1}})

	// Pool only grows on demand, so we should get the same item back.
	v, _ = p.Get(time.Duration(1 * time.Millisecond))
	d = v.(*D)
	t.Check(d.n, Equals, 1)
}

func (s *DynamicTestSuite) TestUnderflow(t *C) {
	size := uint(2)
	p := pool.NewDynamicPool(size, nil, nil)

	for i := 0; i < 2; i++ {
		d, _ := p.Get(1 * time.Millisecond)
		p.Put(d)
	}

	err := p.Put("one too many")
	t.Check(err, Equals, pool.ErrUnderflow)
}

func (s *DynamicTestSuite) TestNewFunc(t *C) {
	// newFunc should only be called once for new items.
	called := 0
	newFunc := func() (interface{}, error) {
		called++
		return 1, nil
	}
	p := pool.NewDynamicPool(2, newFunc, nil)
	for i := 0; i < 5; i++ {
		d, err := p.Get(time.Duration(1 * time.Millisecond))
		t.Assert(err, IsNil)
		err = p.Put(d)
		t.Assert(err, IsNil)
	}
	t.Check(called, Equals, 1)
}

func (s *DynamicTestSuite) TestWaitFree(t *C) {
	size := uint(2)
	p := pool.NewDynamicPool(size, nil, nil)

	// Get all items, so pool is completely used.
	for i := 0; i < 2; i++ {
		p.Get(1 * time.Millisecond)
	}

	// Try to get another item which will wait for a free item.
	var d interface{}
	var err error
	doneChan := make(chan bool, 1)
	go func() {
		d, err = p.Get(5 * time.Second)
		doneChan <- true
	}()

	t.Check(p.Free(), Equals, uint(0))

	// Wait a moment, then put an item back and the goroutine ^ should unblock
	// and be given the free item.
	time.Sleep(250 * time.Millisecond)
	p.Put(&D{n: 101})

	t.Check(p.Free(), Equals, uint(1))

	// Wait for goroutine to finish.
	select {
	case <-doneChan:
	case <-time.After(10 * time.Second):
		t.Fatal("p.Get() did not timeout")
	}

	// If gouroutine got an itme then there should not be any error.
	t.Check(err, IsNil)

	// It should have gotten the first/only free item.
	t.Check(d.(*D).n, Equals, 101)
}

func (s *DynamicTestSuite) TestNewFuncErr(t *C) {
	newFunc := func() (interface{}, error) {
		return nil, errFoo
	}
	p := pool.NewDynamicPool(2, newFunc, nil)
	d, err := p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, Equals, errFoo)
	t.Check(d, IsNil)
}
