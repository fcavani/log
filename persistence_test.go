// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"bytes"
	golog "log"
	"reflect"
	"strings"
	"testing"
	"os"

	"github.com/fcavani/e"
)

type testPers struct {
	name    string
	backend LogBackend
	result  string
	store   Storer
}

var buf *bytes.Buffer
var goLogger *golog.Logger
var mapstore Storer
var testsPers []testPers

func init() {
	buf = bytes.NewBuffer([]byte{})
	goLogger = golog.New(buf, "", golog.LstdFlags)
	mapstore, _ = NewMap(0)

	testsPers = []testPers{
		{"SendToLogger", &SendToLogger{Logger: goLogger}, "test log", nil},
		{"Writer", &Writer{w: buf}, "another test log", nil},
		{"Generic", &Generic{s: mapstore}, "test testing tested", mapstore},
	}
}

func chkresult(t *testing.T, s Storer, result string) {
	err := s.Tx(false, func(tx Transaction) error {
		c := tx.Cursor()
		for k, d := c.First(); k != ""; k, d = c.Next() {
			entry := d.(Entry)
			if strings.Contains(entry.Message(), result) {
				return nil
			}
		}
		return e.New("log message not found")
	})
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
}

func testLogBackend(t *testing.T, pers testPers) {
	pers.backend.F(DefFormatter)
	f := pers.backend.GetF()
	if !reflect.DeepEqual(DefFormatter, f) {
		t.Fatal("miracle! not equal")
	}
	
	back := NewMulti(pers.backend, DefFormatter, NewWriter(os.Stdout), DefFormatter)
	DefFormatter.NewEntry(back).Print(pers.result)
	//DefFormatter.NewEntry(pers.backend).Println(pers.result)

	if pers.store == nil {
		test(t, buf, pers.result)
	} else {
		chkresult(t, pers.store, pers.result)
	}

	olog, ok := pers.backend.(OtherLogger)
	if !ok {
		return
	}
	w := olog.OtherLog("tag")
	defer olog.Close()
	str := "test outer logger\n"
	n, err := w.Write([]byte(str))
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if n != len(str) {
		t.Fatal("write failed", n)
	}
	err = olog.Close()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	if pers.store == nil {
		test(t, buf, str)
	} else {
		chkresult(t, pers.store, str)
	}

}

func TestBackends(t *testing.T) {
	for i, p := range testsPers {
		t.Log(i, p.name)
		testLogBackend(t, p)
	}
}
