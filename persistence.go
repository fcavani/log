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
	r Ruler
}

func (s *SendToLogger) F(f Formatter) LogBackend {
	s.f = f
	return s
}

func (s *SendToLogger) GetF() Formatter {
	return s.f
}

func (s *SendToLogger) Filter(r Ruler) LogBackend {
	s.r = r
	return s
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
	if s.r != nil && !s.r.Result(entry) {
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

type outer struct {
	ch  chan []byte
	buf []byte
}

func (o *outer) Write(p []byte) (n int, err error) {
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
	mp      []LogBackend
	chclose chan chan struct{}
	r       Ruler
	chouter chan []byte
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

func (mp *MultiLog) Filter(r Ruler) LogBackend {
	mp.r = r
	return mp
}

func (mp *MultiLog) Commit(entry Entry) {
	if mp.r != nil && !mp.r.Result(entry) {
		return
	}
	for _, p := range mp.mp {
		p.Commit(entry)
	}
}

// outerLog is like outers outerLogs but the nem entry is
// created from the first BackLog in the list.
func (mp *MultiLog) OuterLog(tag string, level Level) io.Writer {
	if mp.chouter != nil {
		return &outer{
			ch:  mp.chouter,
			buf: make([]byte, 0),
		}
	}
	mp.chclose = make(chan chan struct{})
	mp.chouter = make(chan []byte)
	if len(mp.mp) < 2 {
		return nil
	}
	f := mp.mp[0].GetF()
	logger := f.NewEntry(mp).Tag("outer").EntryLevel(level)
	go func() {
		for {
			select {
			case buf := <-mp.chouter:
				logger.Tag(tag).Println(string(buf))
			case ch := <-mp.chclose:
				ch <- struct{}{}
				return
			}
		}
	}()
	return &outer{
		ch:  mp.chouter,
		buf: make([]byte, 0),
	}
}

func (mp *MultiLog) Close() error {
	if mp.chclose == nil {
		return e.New("already close")
	}
	ch := make(chan struct{})
	mp.chclose <- ch
	<-ch
	mp.chclose = nil
	return nil
}

// Writer log to an io.Writer
type Writer struct {
	f       Formatter
	w       io.Writer
	chclose chan chan struct{}
	chouter chan []byte
	lck     sync.Mutex
	r       Ruler
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

func (w *Writer) Filter(r Ruler) LogBackend {
	w.r = r
	return w
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
	if w.r != nil && !w.r.Result(entry) {
		return
	}
	if w.f == nil {
		err = e.New("formater not set")
		return
	}
	entry.Formatter(w.f)
	buf := entry.Bytes()
	l := len(buf)
	if buf[l-1] != '\n' {
		if cap(buf) < l+1 {
			tmp := make([]byte, l+1)
			copy(tmp, buf)
			buf = tmp
		}
		buf = buf[0 : l+1]
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

func (w *Writer) OuterLog(tag string, level Level) io.Writer {
	if w.chouter != nil {
		return &outer{
			ch:  w.chouter,
			buf: make([]byte, 0),
		}
	}
	w.chclose = make(chan chan struct{})
	w.chouter = make(chan []byte)
	logger := w.f.NewEntry(w).Tag("outer").EntryLevel(level)
	go func() {
		for {
			select {
			case buf := <-w.chouter:
				logger.Tag(tag).Println(string(buf))
			case ch := <-w.chclose:
				ch <- struct{}{}
				return
			}
		}
	}()
	return &outer{
		ch:  w.chouter,
		buf: make([]byte, 0),
	}
}

func (w *Writer) Close() error {
	if w.chclose == nil {
		return e.New("already close")
	}
	ch := make(chan struct{})
	w.chclose <- ch
	<-ch
	w.chclose = nil
	return nil
}

type Generic struct {
	f       Formatter
	s       Storer
	chclose chan chan struct{}
	chouter chan []byte
	r       Ruler
}

func NewGeneric(s Storer) LogBackend {
	g := &Generic{
		s: s,
	}
	return g
}

func (g *Generic) F(f Formatter) LogBackend {
	g.f = f
	return g
}

func (g *Generic) GetF() Formatter {
	return g.f
}

func (g *Generic) Filter(r Ruler) LogBackend {
	g.r = r
	return g
}

func (g *Generic) Commit(entry Entry) {
	var err error
	defer func() {
		if err != nil {
			CommitFail(entry, err)
		}
	}()
	if g.r != nil && !g.r.Result(entry) {
		return
	}
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

func (g *Generic) OuterLog(tag string, level Level) io.Writer {
	if g.chouter != nil {
		return &outer{
			ch:  g.chouter,
			buf: make([]byte, 0),
		}
	}
	g.chclose = make(chan chan struct{})
	g.chouter = make(chan []byte)
	logger := g.f.NewEntry(g).Tag("outer").EntryLevel(level)
	go func() {
		for {
			select {
			case buf := <-g.chouter:
				logger.Tag(tag).Println(string(buf))
			case ch := <-g.chclose:
				ch <- struct{}{}
				return
			}
		}
	}()
	return &outer{
		ch:  g.chouter,
		buf: make([]byte, 0),
	}
}

func (g *Generic) Close() error {
	if g.chclose == nil {
		return e.New("already close")
	}
	ch := make(chan struct{})
	g.chclose <- ch
	<-ch
	g.chclose = nil
	return nil
}
