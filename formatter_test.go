// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/fcavani/e"
	"github.com/fcavani/tags"
)

type entryTest struct {
	Tag  string `log:"tag"`
	Tag2 string `log:"tag2"`
	Tag3 string `log:"tag3"`
}

func (et *entryTest) Date() time.Time {
	return time.Time{}
}

func (et *entryTest) Level() Level {
	return FatalPrio
}

func (et *entryTest) Message() string {
	return ""
}

func (et *entryTest) Tags() *tags.Tags {
	return nil
}

func (et *entryTest) Domain(d string) Logger {
	return nil
}

func (et *entryTest) GetDomain() string {
	return ""
}

func (et *entryTest) Err() error {
	return nil
}

func (et *entryTest) String() string {
	return ""
}

func (et *entryTest) Bytes() []byte {
	return nil
}

func (et *entryTest) Formatter(f Formatter) {

}

func (et *entryTest) SetLevel(scope string, level Level) Logger {
	return nil
}

func (et *entryTest) Sorter(r Ruler) Logger {
	return nil
}

func (et *entryTest) EntryLevel(l Level) Logger {
	return nil
}

func (et *entryTest) DebugInfo() Logger {
	return nil
}

type entryTest2 struct {
	Tag string `log:"tag"`
}

func (et *entryTest2) Date() time.Time {
	return time.Time{}
}

func (et *entryTest2) Level() Level {
	return FatalPrio
}

func (et *entryTest2) Message() string {
	return ""
}

func (et *entryTest2) Tags() *tags.Tags {
	return nil
}

func (et *entryTest2) Domain(d string) Logger {
	return nil
}

func (et *entryTest2) GetDomain() string {
	return ""
}

func (et *entryTest2) Err() error {
	return nil
}

func (et *entryTest2) String() string {
	return ""
}

func (et *entryTest2) Bytes() []byte {
	return nil
}

func (et *entryTest2) Formatter(f Formatter) {

}

func (et *entryTest2) SetLevel(scope string, level Level) Logger {
	return nil
}

func (et *entryTest2) Sorter(r Ruler) Logger {
	return nil
}

func (et *entryTest2) EntryLevel(l Level) Logger {
	return nil
}

func (et *entryTest2) DebugInfo() Logger {
	return nil
}

type teststruct struct {
	raw    string
	entry  *entryTest
	result string
}

var tests = []teststruct{
	{"::tag", &entryTest{Tag: "flor"}, "flor"},
	{"::tag de outubro", &entryTest{Tag: "flor"}, "flor de outubro"},
	{"colhi a ::tag", &entryTest{Tag: "flor"}, "colhi a flor"},
	{"hoje a ::tag irei", &entryTest{Tag: "noite"}, "hoje a noite irei"},
	{"::tag ::tag2 ::tag3", &entryTest{Tag: "1", Tag2: "2", Tag3: "3"}, "1 2 3"},
}

func TestFormat(t *testing.T) {
	f, _ := NewStdFormatter(
		"::",
		"d",
		&entryTest{},
		map[string]interface{}{},
		"",
	)
	f.Entry(&entryTest{})
	for i, test := range tests {
		f.(*StdFormatter).Tmpl = []byte(test.raw)
		out, err := f.Format(test.entry)
		if err != nil {
			t.Fatal(e.Trace(e.Forward(err)))
		}
		if string(out) != test.result {
			t.Fatal("not the same", i, string(out), test.result)
		}
		t.Log(string(out))
	}
}

func TestFormatInvalid(t *testing.T) {
	f, _ := NewStdFormatter(
		"::",
		"d",
		&entryTest{},
		map[string]interface{}{},
		"",
	)
	f.Entry(&entryTest{})
	_, err := f.Format(&entryTest2{})
	if err != nil && !e.Equal(err, ErrNotSupported) {
		t.Fatal(e.Trace(e.Forward(err)))
	} else if err == nil {
		t.Fatal("nil")
	}
}

func TestNewEntry(t *testing.T) {
	f, _ := NewStdFormatter(
		"::",
		"::host - ::domain - ::date - ::level - ::tags - ::msg",
		&log{Labels: &tags.Tags{}},
		map[string]interface{}{"host": "no host"},
		"",
	)
	buf := bytes.NewBuffer([]byte{})
	multi := NewMulti(NewWriter(buf), f, NewWriter(os.Stdout), f)
	logger := New(multi, false).Domain("test")
	f.Entry(logger)

	f.NewEntry(logger.Store()).Println("log test")
	test(t, buf, "test", "log test")
}
