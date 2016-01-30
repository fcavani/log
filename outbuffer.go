// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

type OutBuffer struct {
	bak     LogBackend
	ch      chan Entry
	chclose chan chan struct{}
	closed  chan struct{}
	r       Ruler
}

func NewOutBuffer(bak LogBackend, size int) LogBackend {
	o := &OutBuffer{
		bak:     bak,
		ch:      make(chan Entry, size),
		chclose: make(chan chan struct{}),
		closed:  make(chan struct{}),
	}
	go func() {
		for {
			select {
			case ch := <-o.chclose:
				ch <- struct{}{}
				return
			case entry := <-o.ch:
				if entry == nil {
					o.ch = nil
					o.closed <- struct{}{}
					return
				}
				o.bak.Commit(entry)
			}
		}
	}()
	return o
}

func (o *OutBuffer) F(f Formatter) LogBackend {
	o.bak.F(f)
	return o
}

func (o *OutBuffer) GetF() Formatter {
	return o.bak.GetF()
}

func (o *OutBuffer) Filter(r Ruler) LogBackend {
	o.r = r
	return o
}

func (o *OutBuffer) Commit(entry Entry) {
	if o.r != nil && !o.r.Result(entry) {
		return
	}
	o.ch <- entry
}

func (o *OutBuffer) Close() error {
	if o.ch == nil {
		return nil
	}
	close(o.ch)
	<-o.closed
	// ch := make(chan struct{})
	// o.chclose <- ch
	// <-ch
	return nil
}
