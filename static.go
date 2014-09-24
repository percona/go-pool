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
	"time"
)

type StaticPool struct {
	size    uint
	newFunc func() interface{}
	putFunc func(interface{})
	// --
	pool  chan interface{}
	alloc uint
}

func NewStaticPool(size uint, newFunc func() interface{}, putFunc func(interface{})) *StaticPool {
	pool := make(chan interface{}, int(size))
	for i := 0; i < int(size); i++ {
		pool <- nil
	}
	p := &StaticPool{
		size:    size,
		newFunc: newFunc,
		putFunc: putFunc,
		// --
		pool: pool,
	}
	return p
}

func (p *StaticPool) Size() uint {
	return p.size
}

func (p *StaticPool) Free() uint {
	return uint(len(p.pool))
}

func (p *StaticPool) Get(timeout time.Duration) (interface{}, error) {
	var d interface{}
	select {
	case d = <-p.pool:
		if d == nil && p.newFunc != nil {
			d = p.newFunc()
		}
		return d, nil
	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}

func (p *StaticPool) Put(v interface{}) error {
	if p.putFunc != nil {
		p.putFunc(v)
	}
	select {
	case p.pool <- v:
		return nil
	default:
		return ErrOverflow
	}
}
