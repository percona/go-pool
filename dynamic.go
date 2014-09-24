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

package pool

import (
	"sync"
	"time"
)

type DynamicPool struct {
	size    uint
	newFunc func() interface{}
	putFunc func(interface{})
	// --
	pool chan interface{}
	out  uint
	mux  *sync.Mutex // guards out
}

func NewDynamicPool(size uint, newFunc func() interface{}, putFunc func(interface{})) *DynamicPool {
	p := &DynamicPool{
		size:    size,
		newFunc: newFunc,
		putFunc: putFunc,
		// --
		pool: make(chan interface{}, int(size)),
		mux:  &sync.Mutex{},
		out:  0,
	}
	return p
}

func (p *DynamicPool) Size() uint {
	// We could items out in the dynamic size because eventually they should be
	// put back, thus adding to the future size of the pool. In other words:
	// size is how many items have been allocated.
	p.mux.Lock()
	defer p.mux.Unlock()
	return uint(len(p.pool)) + p.out
}

func (p *DynamicPool) Free() uint {
	p.mux.Lock()
	defer p.mux.Unlock()
	return p.size - p.out
}

func (p *DynamicPool) Get(timeout time.Duration) (interface{}, error) {
	d, err := p.get(timeout)
	if err != nil {
		return nil, err
	}
	if d == nil && p.newFunc != nil {
		d = p.newFunc()
	}
	p.mux.Lock()
	defer p.mux.Unlock()
	p.out++
	return d, nil
}

func (p *DynamicPool) get(timeout time.Duration) (interface{}, error) {
	// If there's a free item, return it immediately.
	select {
	case v := <-p.pool:
		return v, nil
	default:
	}

	// There's not a free item, but if the pool has free space then make
	// a new item.
	if p.Free() > 0 {
		return nil, nil
	}

	// No item is free and the pool is at max size, so wait for an item
	// to become free.
	select {
	case v := <-p.pool:
		return v, nil
	case <-time.After(timeout):
	}
	return nil, ErrTimeout
}

func (p *DynamicPool) Put(v interface{}) error {
	if p.putFunc != nil {
		p.putFunc(v)
	}
	p.mux.Lock()
	defer p.mux.Unlock()
	if p.out == 0 {
		return ErrUnderflow
	}
	select {
	case p.pool <- v:
		p.out--
	default:
		return ErrOverflow // shouldn't happen
	}
	return nil
}
