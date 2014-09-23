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
	pool "github.com/percona/go-pool"
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct {
}

var _ = Suite(&TestSuite{})

type D struct {
	n int
}

func (s *TestSuite) TestGetAndFree(t *C) {
	size := uint(3)
	p := pool.NewPool(size, nil, nil)

	t.Check(p.Free(), Equals, size)

	_, err := p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, IsNil)
	t.Check(p.Free(), Equals, size-1)

	_, err = p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, IsNil)
	t.Check(p.Free(), Equals, size-2)

	_, err = p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, IsNil)
	t.Check(p.Free(), Equals, size-3)

	_, err = p.Get(time.Duration(1 * time.Millisecond))
	t.Check(err, Equals, pool.ErrTimeout)
	t.Check(p.Free(), Equals, uint(0))
}

func (s *TestSuite) TestFuncsAndPut(t *C) {
	size := uint(2)
	n := 0
	newFunc := func() interface{} {
		n++
		d := &D{n: n}
		return d
	}
	put := []*D{}
	putFunc := func(v interface{}) {
		d := v.(*D)
		put = append(put, d)
	}
	p := pool.NewPool(size, newFunc, putFunc)

	v, _ := p.Get(time.Duration(1 * time.Millisecond))
	d := v.(*D)
	t.Check(d.n, Equals, 1)

	p.Put(d)
	t.Check(put, DeepEquals, []*D{&D{n: 1}})
}

func (s *TestSuite) TestOverflow(t *C) {
	size := uint(2)
	p := pool.NewPool(size, nil, nil)

	err := p.Put("one too many")
	t.Check(err, Equals, pool.ErrOverflow)
}

func (s *TestSuite) TestNewFunc(t *C) {
	// newFunc should only be called once for new items.
	called := 0
	newFunc := func() interface{} {
		called++
		return 1
	}
	p := pool.NewPool(2, newFunc, nil)
	for i := 0; i < 5; i++ {
		d, err := p.Get(time.Duration(1 * time.Millisecond))
		t.Assert(err, IsNil)
		err = p.Put(d)
		t.Assert(err, IsNil)
	}
	t.Check(called, Equals, 2)
}
