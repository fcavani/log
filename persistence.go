// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"bytes"
	"io"
	golog "log"
	"os"
	"sync"
	"time"

	"github.com/fcavani/e"
)

var DateFormat = time.RFC3339

// SendToLogger simple log that forward all logged messages to go log.
type SendToLogger struct {
	f Formatter
	*golog.Logger
}

func (s *SendToLogger) F(f Formatter) LogBackend {
	s.f = f
	return s
}

func (s *SendToLogger) GetF() Formatter {
	return s.f
}

func (s *SendToLogger) Commit(entry Entry) {
	var err error
	defer func() {
		if err != nil {
			CommitFail(entry, err)
		}
	}()
	if s.f == nil {
		err = e.New("formater not set")
		return
	}
	entry.Formatter(s.f)
	s.Println(entry.String())
}

// NewSendToLogger creates a logger from a go log.
func NewSendToLogger(logger *golog.Logger) LogBackend {
	if logger == nil {
		return &SendToLogger{
			Logger: golog.New(os.Stderr, "", golog.LstdFlags),
		}
	}
	return &SendToLogger{
		Logger: logger,
	}
}

type other struct {
	ch  chan []byte
	buf []byte
}

func (o *other) Write(p []byte) (n int, err error) {
	o.buf = append(o.buf, p...)
	i := bytes.Index(o.buf, []byte{'\n'})
	if i == -1 {
		return len(p), nil
	}
	o.ch <- o.buf[:i]
	if i+1 < len(o.buf) {
		o.buf = o.buf[i+1:]
		return len(p), nil
	}
	o.buf = o.buf[:0]
	return len(p), nil
}

// MultiLog copy the log entry to multiples backends.
type MultiLog struct {
	mp []LogBackend
	chclose chan chan struct{}
	o       *other
}

//NewMulti creates a MultiLog
func NewMulti(vals ...interface{}) LogBackend {
	if len(vals)%2 != 0 {
		Fail(e.New("parameters must be in pair of LogBackend and Formatter"))
		return nil
	}
	l := len(vals) / 2
	mp := make([]LogBackend, 0, l)
	for i := 0; i < len(vals); i += 2 {
		bak, ok := vals[i].(LogBackend)
		if !ok {
			Fail(e.New("not a LogBackend"))
			return nil
		}
		f, ok := vals[i+1].(Formatter)
		if !ok {
			Fail(e.New("not a Formatter"))
			return nil
		}
		bak.F(f)
		mp = append(mp, bak)
	}
	return &MultiLog{
		mp: mp,
	}
}

func (mp *MultiLog) F(f Formatter) LogBackend {
	return mp
}

func (mp *MultiLog) GetF() Formatter {
	return nil
}

func (mp *MultiLog) Commit(entry Entry) {
	for _, p := range mp.mp {
		p.Commit(entry)
	}
}

// OtherLog is like others OtherLogs but the nem entry is
// created from the first BackLog in the list.
func (mp *MultiLog) OtherLog(tag string) io.Writer {
	if mp.o != nil {
		return mp.o
	}
	mp.chclose = make(chan chan struct{})
	ch := make(chan []byte)
	if len(mp.mp) < 2 {
		return nil
	}
	f := mp.mp[0].GetF()
	logger := f.NewEntry(mp).Tag("outer")
	go func() {
		for {
			select {
			case buf := <-ch:
				logger.Tag(tag).Println(string(buf))
			case ch := <-mp.chclose:
				ch <- struct{}{}
				return
			}
		}
	}()
	mp.o = &other{
		ch:  ch,
		buf: make([]byte, 0),
	}
	return mp.o
}

func (mp *MultiLog) Close() error {
	if mp.o == nil {
		return e.New("already close")
	}
	ch := make(chan struct{})
	mp.chclose <- ch
	<-ch
	mp.o = nil
	return nil
}


// Writer log to an io.Writer
type Writer struct {
	f       Formatter
	w       io.Writer
	chclose chan chan struct{}
	o       *other
	lck     sync.Mutex
}

// NewWriter creates a backend that log to w.
func NewWriter(w io.Writer) LogBackend {
	return &Writer{
		w: w,
	}
}

func (w *Writer) F(f Formatter) LogBackend {
	w.f = f
	return w
}

func (w *Writer) GetF() Formatter {
	return w.f
}

func (w *Writer) Writer(writter io.Writer) {
	w.lck.Lock()
	defer w.lck.Unlock()
	w.w = writter
}

func (w *Writer) Commit(entry Entry) {
	w.lck.Lock()
	defer w.lck.Unlock()
	var err error
	defer func() {
		if err != nil {
			CommitFail(entry, err)
		}
	}()
	if w.f == nil {
		err = e.New("formater not set")
		return
	}
	entry.Formatter(w.f)
	buf := entry.Bytes()
	l := len(buf)
	if buf[l-1] != '\n' {
		if cap(buf) < l + 1 {
			tmp := make([]byte, l + 1)
			copy(tmp, buf)
			buf = tmp
		}
		buf = buf[0:l+1]
		buf[len(buf)-1] = '\n'
	}
	var n int
	for len(buf) > 0 {
		n, err = w.w.Write(buf)
		if err != nil {
			return
		}
		buf = buf[n:]
	}
}

func (w *Writer) OtherLog(tag string) io.Writer {
	if w.o != nil {
		return w.o
	}
	w.chclose = make(chan chan struct{})
	ch := make(chan []byte)
	logger := w.f.NewEntry(w).Tag("outer")
	go func() {
		for {
			select {
			case buf := <-ch:
				logger.Tag(tag).Println(string(buf))
			case ch := <-w.chclose:
				ch <- struct{}{}
				return
			}
		}
	}()
	w.o = &other{
		ch:  ch,
		buf: make([]byte, 0),
	}
	return w.o
}

func (w *Writer) Close() error {
	if w.o == nil {
		return e.New("already close")
	}
	ch := make(chan struct{})
	w.chclose <- ch
	<-ch
	w.o = nil
	return nil
}

type Generic struct {
	f       Formatter
	s       Storer
	chclose chan chan struct{}
	o       *other
	lck     sync.Mutex
}

func NewGeneric(s Storer) LogBackend {
	return &Generic{
		s: s,
	}
}

func (g *Generic) F(f Formatter) LogBackend {
	g.f = f
	return g
}

func (g *Generic) GetF() Formatter {
	return g.f
}

func (g *Generic) Commit(entry Entry) {
	g.lck.Lock()
	defer g.lck.Unlock()
	var err error
	defer func() {
		if err != nil {
			CommitFail(entry, err)
		}
	}()
	if g.f == nil {
		err = e.New("formater not set")
		return
	}
	err = g.s.Tx(true, func(tx Transaction) error {
		entry.Formatter(g.f)
		err := tx.Put(entry.Date().Format(time.RFC3339Nano), entry)
		if err != nil {
			return e.Forward(err)
		}
		return nil
	})
	if err != nil {
		err = e.Forward(err)
		return
	}
}

func (g *Generic) OtherLog(tag string) io.Writer {
	if g.o != nil {
		return g.o
	}
	g.chclose = make(chan chan struct{})
	ch := make(chan []byte)
	logger := g.f.NewEntry(g).Tag("outer")
	go func() {
		for {
			select {
			case buf := <-ch:
				logger.Tag(tag).Println(string(buf))
			case ch := <-g.chclose:
				ch <- struct{}{}
				return
			}
		}
	}()
	g.o = &other{
		ch:  ch,
		buf: make([]byte, 0),
	}
	return g.o
}

func (g *Generic) Close() error {
	if g.o == nil {
		return e.New("already close")
	}
	ch := make(chan struct{})
	g.chclose <- ch
	<-ch
	g.o = nil
	return nil
}
