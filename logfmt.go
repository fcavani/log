// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"io"
	"reflect"

	"github.com/go-logfmt/logfmt"
)

type Logfmt struct {
	enc *logfmt.Encoder
	r   Ruler
}

func NewLogfmt(w io.Writer) *Logfmt {
	enc := logfmt.NewEncoder(w)
	return &Logfmt{
		enc: enc,
	}
}

func (l *Logfmt) F(f Formatter) LogBackend {
	return l
}

func (l *Logfmt) GetF() Formatter {
	return nil
}

func (l *Logfmt) Filter(r Ruler) LogBackend {
	l.r = r
	return l
}

func (l *Logfmt) Commit(entry Entry) {
	if l.r != nil && !l.r.Result(entry) {
		return
	}
	if lfmt, ok := entry.(Logfmter); ok {
		err := lfmt.Logfmt(l.enc)
		if err != nil {
			Fail(err)
		}
		return
	}
	val := reflect.Indirect(reflect.ValueOf(entry))
	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		tag := ft.Tag.Get("log")
		if tag == "" {
			continue
		}
		vf := val.Field(i)
		if !vf.IsValid() {
			continue
		}
		if !vf.CanSet() {
			continue
		}
		err := l.enc.EncodeKeyval(tag, vf.Interface())
		if err != nil {
			Fail(err)
			break
		}
	}
	err := l.enc.EndRecord()
	if err != nil {
		Fail(err)
	}
}
