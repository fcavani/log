// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"fmt"
	"os"
	"time"
	"runtime"
	"strings"
	"strconv"
	"encoding/gob"

	"github.com/fcavani/e"
	"github.com/fcavani/tags"
	"github.com/fcavani/types"
)

var Log Logger

var DefFormatter Formatter

func init() {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "no name"
	}
	DefFormatter, _ = NewStdFormatter(
		"::",
		"::host - ::domain - ::date - ::level - ::tags - ::file ::msg",
		&log{Labels: &tags.Tags{}},
		map[string]interface{}{
			"host": hostname,
		},
	)
	Log = New(
		NewWriter(os.Stdout).F(DefFormatter),
		false,
	)
}

type log struct {
	Timestamp time.Time  `bson:"key",log:"date"`
	Priority  Level      `log:"level"`
	Labels    *tags.Tags `log:"tags",def:"no tags"`
	Msg       string     `log:"msg"`
	Dom       string     `log:"domain"`
	E       error
	f         Formatter
	store     LogBackend
	Debug	  bool
	File	  string	`log:"file"`
}

func init() {
	gob.Register(&log{})
	types.Insert(&log{})
}

func New(b LogBackend, debug bool) *log {
	return &log{
		Priority: NoPrio,
		Labels:   &tags.Tags{},
		store:    b,
		Debug:    debug,
	}
}

func (l *log) clone() *log {
	return &log{
		Timestamp: l.Timestamp,
		Priority:  l.Priority,
		Labels:    l.Labels.Copy(),
		Msg:       l.Msg,
		Dom:       l.Dom,
		E:         e.Copy(l.E),
		store:     l.store,
		Debug:     l.Debug,
	}
}

func (l *log) error(err error) {
	if err == nil {
		return
	}
	n := l.clone()
	n.FatalLevel().Domain("logger").Tag("internal").Tag("error").Println(err)
}

func (l *log) debugInfo() {
	if !l.Debug {
		return
	}
	var ok bool
	var line int
	_, l.File, line, ok = runtime.Caller(2)
	if ok {
		s := strings.Split(l.File, "/")
		length := len(s)
		if length >= 2 {
			l.File = strings.Join(s[length-2:length], "/") + ":" + strconv.Itoa(line)
		} else {
			l.File = s[0] + ":" + strconv.Itoa(line)
		}
	}
}

func (l *log) Err() error {
	return l.E
}

func (l *log) Date() time.Time {
	return l.Timestamp
}

func (l *log) Level() Level {
	return l.Priority
}

func (l *log) Message() string {
	return l.Msg
}

func (l *log) Tag(tag string) Logger {
	n := l.clone()
	n.E = e.Forward(n.Labels.Add(tag))
	l.error(n.E)
	return n
}

func (l *log) Mark(mark string) Logger {
	f := l.store.GetF()
	if f != nil {
		f.Mark(mark)
	}
	return l
}

func (l *log) Template(t string) Logger {
	f := l.store.GetF()
	if f != nil {
		f.Template(t)
	}
	return l
}

func (l *log) Tags() *tags.Tags {
	return l.Labels
}

func (l *log) Domain(d string) Logger {
	n := l.clone()
	n.Dom = d
	return n
}

func (l *log) GetDomain() string {
	return l.Dom
}

func (l *log) Store() LogBackend {
	return l.store
}

func (l *log) SetStore(b LogBackend) Logger {
	n := l.clone()
	n.store = b
	return n
}

func (l *log) Bytes() []byte {
	buf, err := l.f.Format(l)
	if err != nil {
		return []byte("Can't format the log entry: " + err.Error())
	}
	return buf
}

func (l *log) String() string {
	buf, err := l.f.Format(l)
	if err != nil {
		return "Can't format the log entry: " + err.Error()
	}
	return string(buf)
}

func (l *log) Formatter(f Formatter) {
	l.f = f
}

func (l *log) Print(v ...interface{}) {
	l.Msg = fmt.Sprint(v...)
	l.Timestamp = time.Now()
	l.debugInfo()
	l.store.Commit(l)
}

func (l *log) Printf(f string, v ...interface{}) {
	l.Msg = fmt.Sprintf(f, v...)
	l.Timestamp = time.Now()
	l.debugInfo()
	l.store.Commit(l)
}

func (l *log) Println(v ...interface{}) {
	l.Msg = fmt.Sprintln(v...)
	l.Timestamp = time.Now()
	l.debugInfo()
	l.store.Commit(l)
}

func (l *log) Fatal(v ...interface{}) {
	l.Priority = FatalPrio
	l.Msg = fmt.Sprint(v...)
	l.Timestamp = time.Now()
	l.debugInfo()
	l.store.Commit(l)
	os.Exit(1)
}

func (l *log) Fatalf(f string, v ...interface{}) {
	l.Priority = FatalPrio
	l.Msg = fmt.Sprintf(f, v...)
	l.Timestamp = time.Now()
	l.debugInfo()
	l.store.Commit(l)
	os.Exit(1)
}

func (l *log) Fatalln(v ...interface{}) {
	l.Priority = FatalPrio
	l.Msg = fmt.Sprintln(v...)
	l.Timestamp = time.Now()
	l.debugInfo()
	l.store.Commit(l)
	os.Exit(1)
}

func (l *log) Panic(v ...interface{}) {
	l.Priority = PanicPrio
	l.Msg = fmt.Sprint(v...)
	l.Timestamp = time.Now()
	l.debugInfo()
	l.store.Commit(l)
	panic(l.Msg)
}

func (l *log) Panicf(f string, v ...interface{}) {
	l.Priority = PanicPrio
	l.Msg = fmt.Sprintf(f, v...)
	l.Timestamp = time.Now()
	l.debugInfo()
	l.store.Commit(l)
	panic(l.Msg)
}

func (l *log) Panicln(v ...interface{}) {
	l.Priority = PanicPrio
	l.Msg = fmt.Sprintln(v...)
	l.Timestamp = time.Now()
	l.debugInfo()
	l.store.Commit(l)
	panic(l.Msg)
}

func (l *log) Error(v ...interface{}) {
	l.Priority = ErrorPrio
	l.Msg = fmt.Sprint(v...)
	l.Timestamp = time.Now()
	l.debugInfo()
	l.store.Commit(l)
}

func (l *log) Errorf(f string, v ...interface{}) {
	l.Priority = ErrorPrio
	l.Msg = fmt.Sprintf(f, v...)
	l.Timestamp = time.Now()
	l.debugInfo()
	l.store.Commit(l)
}

func (l *log) Errorln(v ...interface{}) {
	l.Priority = ErrorPrio
	l.Msg = fmt.Sprintln(v...)
	l.Timestamp = time.Now()
	l.debugInfo()
	l.store.Commit(l)
}

func (l *log) GoPanic(r interface{}, stack []byte, cont bool) {
	l.Priority = PanicPrio
	l.Timestamp = time.Now()
	switch v := r.(type) {
	case string:
		l.Msg = v + "\n"
	case fmt.Stringer:
		l.Msg = v.String() + "\n"
	default:
		l.Msg = fmt.Sprintln(r)
	}
	l.Msg += "\n" + string(stack)
	l.store.Commit(l)
	if !cont {
		os.Exit(1)
	}
}

func (l *log) ProtoLevel() Logger {
	n := l.clone()
	n.Priority = ProtoPrio
	return n
}

func (l *log) DebugLevel() Logger {
	n := l.clone()
	n.Priority = DebugPrio
	return n
}

func (l *log) InfoLevel() Logger {
	n := l.clone()
	n.Priority = InfoPrio
	return n
}

func (l *log) WarnLevel() Logger {
	n := l.clone()
	n.Priority = WarnPrio
	return n
}

func (l *log) ErrorLevel() Logger {
	n := l.clone()
	n.Priority = ErrorPrio
	return n
}

func (l *log) FatalLevel() Logger {
	n := l.clone()
	n.Priority = FatalPrio
	return n
}

func (l *log) PanicLevel() Logger {
	n := l.clone()
	n.Priority = PanicPrio
	return n
}

func Mark(mark string) Logger {
	return Log.Mark(mark)
}

func Template(t string) Logger {
	return Log.Template(t)
}

func Domain(d string) Logger {
	return Log.Domain(d)
}

func Print(vals ...interface{}) {
	Log.Print(vals...)
}

func Printf(str string, vals ...interface{}) {
	Log.Printf(str, vals...)
}

func Println(vals ...interface{}) {
	Log.Println(vals...)
}

func Fatal(vals ...interface{}) {
	Log.Fatal(vals...)
}

func Fatalf(s string, vals ...interface{}) {
	Log.Fatalf(s, vals...)
}

func Fatalln(vals ...interface{}) {
	Log.Fatalln(vals...)
}

func Panic(vals ...interface{}) {
	Log.Panic(vals...)
}

func Panicf(s string, vals ...interface{}) {
	Log.Panicf(s, vals...)
}

func Panicln(vals ...interface{}) {
	Log.Panicln(vals...)
}

func Error(vals ...interface{}) {
	Log.Error(vals...)
}

func Errorf(s string, vals ...interface{}) {
	Log.Errorf(s, vals...)
}

func Errorln(vals ...interface{}) {
	Log.Errorln(vals...)
}

func ProtoLevel() Logger {
	return Log.ProtoLevel()
}

func DebugLevel() Logger {
	return Log.DebugLevel()
}

func InfoLevel() Logger {
	return Log.InfoLevel()
}

func WarnLevel() Logger {
	return Log.WarnLevel()
}

func ErrorLevel() Logger {
	return Log.ErrorLevel()
}

func FatalLevel() Logger {
	return Log.FatalLevel()
}

func PanicLevel() Logger {
	return Log.PanicLevel()
}

func Tag(tag string) Logger {
	return Log.Tag(tag)
}

func GoPanic(r interface{}, stack []byte, cont bool) {
	Log.GoPanic(r, stack, cont)
}

func SetStore(b LogBackend) Logger {
	n := Log.(*log).clone()
	n.store = b
	return n
}

func Store() LogBackend {
	return Log.Store()
}

func GetLogger() Logger {
	n := Log.(*log).clone()
	return n
}
