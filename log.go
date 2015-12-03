// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"encoding/gob"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fcavani/e"
	"github.com/fcavani/tags"
	"github.com/fcavani/types"
	"github.com/go-logfmt/logfmt"
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
		"::host - ::domain - ::date - ::level - ::tags - ::file - ::msg",
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
	Timestamp time.Time  `bson:"key" log:"date"`
	Priority  Level      `log:"level"`
	Labels    *tags.Tags `log:"tags" def:"no tags"`
	Msg       string     `log:"msg"`
	Dom       string     `log:"domain"`
	E         error
	f         Formatter
	store     LogBackend
	Debug     bool
	File      string `log:"file"`
	Pkg       string `log:"pkg"`
	Func      string `log:"func"`
	Levels    map[string]*If
	DefLevel  Ruler
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
		Levels:   make(map[string]*If),
		DefLevel: True{},
	}
}

func (l *log) Logfmt(enc *logfmt.Encoder) error {
	err := enc.EncodeKeyval("date", l.Timestamp)
	if err != nil {
		return e.Forward(err)
	}
	err = enc.EncodeKeyval("level", l.Priority)
	if err != nil {
		return e.Forward(err)
	}
	err = enc.EncodeKeyval("tags", l.Labels.String())
	if err != nil {
		return e.Forward(err)
	}
	err = enc.EncodeKeyval("msg", l.Msg)
	if err != nil {
		return e.Forward(err)
	}
	err = enc.EncodeKeyval("domain", l.Dom)
	if err != nil {
		return e.Forward(err)
	}
	err = enc.EncodeKeyval("file", l.File)
	if err != nil {
		return e.Forward(err)
	}
	err = enc.EndRecord()
	if err != nil {
		return e.Forward(err)
	}
	return nil
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
		File:      l.File,
		Pkg:       l.Pkg,
		Func:      l.Func,
		Levels:    l.Levels,
		DefLevel:  l.DefLevel,
	}
}

func (l *log) error(err error) {
	if err == nil {
		return
	}
	n := l.clone()
	n.FatalLevel().Domain("logger").Tag("internal").Tag("error").Println(err)
}

func (l *log) debugInfo(level int) {
	if !l.Debug || l.File != "" {
		return
	}
	var ok bool
	var line int
	var pc uintptr
	pc, l.File, line, ok = runtime.Caller(level)
	if ok {
		s := strings.Split(l.File, "/")
		length := len(s)
		if length >= 2 {
			l.File = strings.Join(s[length-2:length], "/") + ":" + strconv.Itoa(line)
		} else {
			l.File = s[0] + ":" + strconv.Itoa(line)
		}
		f := runtime.FuncForPC(pc)
		l.Func = f.Name()
		i := strings.LastIndex(l.Func, ".")
		if i > -1 {
			l.Pkg = l.Func[:i]
		}
	}
}

func (l *log) DebugInfo() Logger {
	n := l.clone()
	n.debugInfo(3)
	return n
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

func (l *log) Tag(tags ...string) Logger {
	n := l.clone()
	n.E = e.Forward(n.Labels.MergeFromStringSlice(tags))
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

func (l *log) Sorter(r Ruler) Logger {
	n := l.clone()
	n.store.Filter(r)
	return n
}

func (l *log) SetLevel(scope string, level Level) Logger {
	if scope == "all" {
		l.DefLevel = Op(Ge, "level", level)
	} else {
		l.Levels[scope] = &If{
			Condition: Op(Pr, "pkg", scope),
			Than:      Op(Ge, "level", level),
		}
	}
	ifs := make([]*If, 0, len(l.Levels))
	for _, cond := range l.Levels {
		ifs = append(ifs, cond)
	}
	l.store.Filter(Select(ifs, l.DefLevel))
	return l
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
	n := l.clone()
	n.Msg = fmt.Sprint(v...)
	n.Timestamp = time.Now()
	n.debugInfo(2)
	n.store.Commit(n)
}

func (l *log) Printf(f string, v ...interface{}) {
	n := l.clone()
	n.Msg = fmt.Sprintf(f, v...)
	n.Timestamp = time.Now()
	n.debugInfo(2)
	n.store.Commit(n)
}

func (l *log) Println(v ...interface{}) {
	n := l.clone()
	n.Msg = fmt.Sprintln(v...)
	n.Timestamp = time.Now()
	n.debugInfo(2)
	n.store.Commit(n)
}

func (l *log) Fatal(v ...interface{}) {
	n := l.clone()
	n.Priority = FatalPrio
	n.Msg = fmt.Sprint(v...)
	n.Timestamp = time.Now()
	n.debugInfo(2)
	n.store.Commit(n)
	os.Exit(1)
}

func (l *log) Fatalf(f string, v ...interface{}) {
	n := l.clone()
	n.Priority = FatalPrio
	n.Msg = fmt.Sprintf(f, v...)
	n.Timestamp = time.Now()
	n.debugInfo(2)
	n.store.Commit(n)
	os.Exit(1)
}

func (l *log) Fatalln(v ...interface{}) {
	n := l.clone()
	n.Priority = FatalPrio
	n.Msg = fmt.Sprintln(v...)
	n.Timestamp = time.Now()
	n.debugInfo(2)
	n.store.Commit(n)
	os.Exit(1)
}

func (l *log) Panic(v ...interface{}) {
	n := l.clone()
	n.Priority = PanicPrio
	n.Msg = fmt.Sprint(v...)
	n.Timestamp = time.Now()
	n.debugInfo(2)
	n.store.Commit(n)
	panic(n.Msg)
}

func (l *log) Panicf(f string, v ...interface{}) {
	n := l.clone()
	n.Priority = PanicPrio
	n.Msg = fmt.Sprintf(f, v...)
	n.Timestamp = time.Now()
	n.debugInfo(2)
	n.store.Commit(n)
	panic(n.Msg)
}

func (l *log) Panicln(v ...interface{}) {
	n := l.clone()
	n.Priority = PanicPrio
	n.Msg = fmt.Sprintln(v...)
	n.Timestamp = time.Now()
	n.debugInfo(2)
	n.store.Commit(n)
	panic(n.Msg)
}

func (l *log) Error(v ...interface{}) {
	n := l.clone()
	n.Priority = ErrorPrio
	n.Msg = fmt.Sprint(v...)
	n.Timestamp = time.Now()
	n.debugInfo(2)
	n.store.Commit(n)
}

func (l *log) Errorf(f string, v ...interface{}) {
	n := l.clone()
	n.Priority = ErrorPrio
	n.Msg = fmt.Sprintf(f, v...)
	n.Timestamp = time.Now()
	n.debugInfo(2)
	n.store.Commit(n)
}

func (l *log) Errorln(v ...interface{}) {
	n := l.clone()
	n.Priority = ErrorPrio
	n.Msg = fmt.Sprintln(v...)
	n.Timestamp = time.Now()
	n.debugInfo(2)
	n.store.Commit(n)
}

func (l *log) GoPanic(r interface{}, stack []byte, cont bool) {
	n := l.clone()
	n.Priority = PanicPrio
	n.Timestamp = time.Now()
	switch v := r.(type) {
	case string:
		n.Msg = v + "\n"
	case fmt.Stringer:
		n.Msg = v.String() + "\n"
	default:
		n.Msg = fmt.Sprintln(r)
	}
	n.Msg += "\n" + string(stack)
	n.store.Commit(n)
	if !cont {
		os.Exit(1)
	}
}

func (l *log) EntryLevel(prio Level) Logger {
	n := l.clone()
	n.Priority = prio
	return n
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
	Log.DebugInfo().Print(vals...)
}

func Printf(str string, vals ...interface{}) {
	Log.DebugInfo().Printf(str, vals...)
}

func Println(vals ...interface{}) {
	Log.DebugInfo().Println(vals...)
}

func Fatal(vals ...interface{}) {
	Log.DebugInfo().Fatal(vals...)
}

func Fatalf(s string, vals ...interface{}) {
	Log.DebugInfo().Fatalf(s, vals...)
}

func Fatalln(vals ...interface{}) {
	Log.DebugInfo().Fatalln(vals...)
}

func Panic(vals ...interface{}) {
	Log.DebugInfo().Panic(vals...)
}

func Panicf(s string, vals ...interface{}) {
	Log.DebugInfo().Panicf(s, vals...)
}

func Panicln(vals ...interface{}) {
	Log.DebugInfo().Panicln(vals...)
}

func Error(vals ...interface{}) {
	Log.DebugInfo().Error(vals...)
}

func Errorf(s string, vals ...interface{}) {
	Log.DebugInfo().Errorf(s, vals...)
}

func Errorln(vals ...interface{}) {
	Log.DebugInfo().Errorln(vals...)
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

func Tag(tags ...string) Logger {
	return Log.Tag(tags...)
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

func Sorter(r Ruler) Logger {
	return Log.Sorter(r)
}

func SetLevel(scope string, l Level) Logger {
	return Log.SetLevel(scope, l)
}

func EntryLevel(prio Level) Logger {
	return Log.EntryLevel(prio)
}

// RecoverBufferStack amont of buffer to store the stack.
var RecoverBufferStack = 4096

// Recover from panic and log the stack. If notexit is false, call os.Exit(1),
// if not continue.
func Recover(notexit bool) {
	if r := recover(); r != nil {
		buf := make([]byte, RecoverBufferStack)
		n := runtime.Stack(buf, true)
		buf = buf[:n]
		Log.GoPanic(r, buf, notexit)
	}
}
