// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"bytes"
	golog "log"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/fcavani/e"
	"github.com/fcavani/rand"
	"github.com/fcavani/types"
	"github.com/op/go-logging"
)

func test(t *testing.T, buf *bytes.Buffer, ss ...string) {
	str, err := buf.ReadString('\n')
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	for _, s := range ss {
		if !strings.Contains(str, s) {
			t.Fatal("log didn't log", str)
		}
	}
}

func testerr(buf *bytes.Buffer, ss ...string) error {
	str, err := buf.ReadString('\n')
	if err != nil {
		return e.Forward(err)
	}
	for _, s := range ss {
		if !strings.Contains(str, s) {
			return e.New("log didn't log: %v", str)
		}
	}
	return nil
}

func testb(b *testing.B, buf *bytes.Buffer, ss ...string) {
	str, err := buf.ReadString('\n')
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	for _, s := range ss {
		if !strings.Contains(str, s) {
			b.Error("log didn't log", str)
		}
	}
}

func TestStdLog(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	l := golog.New(buf, "", golog.LstdFlags)
	multi := NewMulti(NewSendToLogger(nil), DefFormatter, NewSendToLogger(l), DefFormatter)
	Log = New(multi, false).Domain("test")

	Print("oi")
	test(t, buf, "oi")

	Printf("blá")
	test(t, buf, "blá")

	Println("ploc")
	test(t, buf, "ploc")
}

func TestLevels(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	l := golog.New(buf, "", golog.LstdFlags)
	multi := NewMulti(NewSendToLogger(nil), DefFormatter, NewSendToLogger(l), DefFormatter)
	Log = New(multi, false).Domain("test")

	Log.SetLevel(DebugPrio)

	Println("oi")
	test(t, buf, "oi")

	ProtoLevel().Println("blá")
	err := testerr(buf, "blá")
	if err != nil && !e.Contains(err, "log didn't log") {
		t.Fatal(e.Trace(e.Forward(err)))
	} else if err == nil {
		t.Fatal("nil error")
	}

}

func TestPanic(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	defer func() {
		r := recover()
		if r != nil {
			if r.(string) != "panic" {
				t.Fatal("not that panic", r.(string))
			}
			test(t, buf, "panic")
			return
		}
		t.Fatal("not panic")
	}()

	l := golog.New(buf, "", golog.LstdFlags)
	multi := NewMulti(NewSendToLogger(nil), DefFormatter, NewSendToLogger(l), DefFormatter)
	Log = New(multi, false).Domain("test")

	Panic("panic")
}

func TestPanicf(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	defer func() {
		r := recover()
		if r != nil {
			if r.(string) != "panic" {
				t.Fatal("not that panic", r.(string))
			}
			test(t, buf, "panic")
			return
		}
		t.Fatal("not panic")
	}()

	l := golog.New(buf, "", golog.LstdFlags)
	multi := NewMulti(NewSendToLogger(nil), DefFormatter, NewSendToLogger(l), DefFormatter)
	Log = New(multi, false).Domain("test")

	Panicf("%v", "panic")
}

func TestPanicln(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	defer func() {
		r := recover()
		if r != nil {
			if r.(string) != "panic\n" {
				t.Fatal("not that panic", r.(string))
			}
			test(t, buf, "panic")
			return
		}
		t.Fatal("not panic")
	}()

	l := golog.New(buf, "", golog.LstdFlags)
	multi := NewMulti(NewSendToLogger(nil), DefFormatter, NewSendToLogger(l), DefFormatter)
	Log = New(multi, false).Domain("test")

	Panicln("panic")
}

func TestProperties(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	multi := NewMulti(NewWriter(buf), DefFormatter, NewWriter(os.Stdout), DefFormatter)
	Log = New(multi, false).Domain("test").Tag("tag1")

	ProtoLevel().Println("log test")
	test(t, buf, "test", "tag1", "protocol", "log test")

	DebugLevel().Println("log test")
	test(t, buf, "test", "tag1", "debug", "log test")

	InfoLevel().Println("log test")
	test(t, buf, "test", "tag1", "info", "log test")

	WarnLevel().Println("log test")
	test(t, buf, "test", "tag1", "warning", "log test")

	ErrorLevel().Println("log test")
	test(t, buf, "test", "tag1", "error", "log test")

	FatalLevel().Println("log test")
	test(t, buf, "test", "tag1", "fatal", "log test")

	PanicLevel().Println("log test")
	test(t, buf, "test", "tag1", "panic", "log test")

	PanicLevel().Tag("tag2").Domain("domain").Println("log test")
	test(t, buf, "domain", "tag1", "tag2", "panic", "log test")
}

func TestPanicHandler(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	multi := NewMulti(NewWriter(buf), DefFormatter, NewWriter(os.Stdout), DefFormatter)
	Log = New(multi, false).Domain("test")

	defer func() {
		r := recover()
		if r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, true)
			buf = buf[:n]
			GoPanic(r, buf, true)
		}
		test(t, buf, "test", "test panic logging")
	}()

	panic("test panic logging")
}

func TestDebug(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	multi := NewMulti(NewWriter(buf), DefFormatter, NewWriter(os.Stdout), DefFormatter)
	Log = New(multi, true).Domain("test").Tag("tag1")

	Log.Println("teste debug info")
	test(t, buf, "test", "log/log_test.go:214", "teste debug info")
}

func TestMultiLine(t *testing.T) {
	formatter, _ := NewStdFormatter(
		"::",
		"::host - ::domain - ::date - ::level - ::tags - ::file\n\t::msg",
		Log,
		map[string]interface{}{
			"host": "",
		},
	)
	buf := bytes.NewBuffer([]byte{})
	multi := NewMulti(NewWriter(buf), formatter, NewWriter(os.Stdout), formatter)
	Log = New(multi, true).Domain("test").Tag("tag1")

	Log.Println("na outra linha")
	test(t, buf, "test", "test", "tag1")
	test(t, buf, "na outra linha")
}

var msg = "benchmark log test"
var l = int64(len(msg))

func BenchmarkPureGolog(b *testing.B) {
	file, err := os.OpenFile("/dev/null", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	gologger := golog.New(file, "", golog.LstdFlags)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gologger.Print(msg)
		b.SetBytes(l)
	}
}

func BenchmarkGoLogging(b *testing.B) {
	file, err := os.OpenFile("/dev/null", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	var log = logging.MustGetLogger("example")
	backend1 := logging.NewLogBackend(file, "", 0)
	logging.SetBackend(backend1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Critical(msg)
		b.SetBytes(l)
	}
}

func BenchmarkLogrus(b *testing.B) {
	file, err := os.OpenFile("/dev/null", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	logrus.SetOutput(file)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logrus.Info(msg)
		b.SetBytes(l)
	}
}

func BenchmarkGolog(b *testing.B) {
	file, err := os.OpenFile("/dev/null", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	gologger := golog.New(file, "", golog.LstdFlags)
	logger := New(
		NewSendToLogger(gologger).F(DefFormatter),
		false,
	).Domain("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Print(msg)
		b.SetBytes(l)
	}
}

func BenchmarkLogStderr(b *testing.B) {
	logger := New(
		NewWriter(os.Stderr).F(DefFormatter),
		false,
	).Domain("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Print(msg)
		b.SetBytes(l)
	}
}

func BenchmarkLogFile(b *testing.B) {
	name, err := rand.FileName("BenchmarkLogFile", ".log", 10)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	name = os.TempDir() + name
	file, err := os.Create(name)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	defer os.Remove(name)
	defer file.Close()

	logger := New(
		NewWriter(file).F(DefFormatter),
		false,
	).Domain("test")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Print(msg)
		b.SetBytes(l)
	}
}

func BenchmarkLogFileBuffer(b *testing.B) {
	name, err := rand.FileName("BenchmarkLogFile", ".log", 10)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	name = os.TempDir() + name
	file, err := os.Create(name)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	defer os.Remove(name)
	defer file.Close()

	back := NewOutBuffer(
		NewWriter(file).F(DefFormatter),
		b.N/2,
	)
	defer back.(*OutBuffer).Close()

	logger := New(back, false).Domain("test")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Print(msg)
		b.SetBytes(l)
	}
}

func BenchmarkStoreMap(b *testing.B) {
	m, _ := NewMap(b.N)
	logger := New(
		NewGeneric(m).F(DefFormatter),
		false,
	).Domain("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Print(msg)
		b.SetBytes(l)
	}
}

func BenchmarkBoltDb(b *testing.B) {
	name, err := rand.FileName("boltdb", ".db", 10)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	name = os.TempDir() + "/" + name
	gob := &Gob{
		TypeName: types.Name(&log{}),
	}
	bolt, err := NewBoltDb("test", name, 0600, nil, gob, gob)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	logger := New(
		NewGeneric(bolt).F(DefFormatter),
		false,
	).Domain("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Print(msg)
		b.SetBytes(l)
	}
}

func BenchmarkBoltDbBuffer(b *testing.B) {
	name, err := rand.FileName("boltdb", ".db", 10)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	name = os.TempDir() + "/" + name
	gob := &Gob{
		TypeName: types.Name(&log{}),
	}
	bolt, err := NewBoltDb("test", name, 0600, nil, gob, gob)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	logger := New(
		NewOutBuffer(
			NewGeneric(bolt).F(DefFormatter),
			b.N/2,
		),
		false,
	).Domain("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Print(msg)
		b.SetBytes(l)
	}
}

func BenchmarkMongoDb(b *testing.B) {
	mongodb, err := NewMongoDb("mongodb://localhost/test", "test", nil, Log, 30*time.Second)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	logger := New(
		NewGeneric(mongodb).F(DefFormatter),
		false,
	).Domain("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Print(msg)
		b.SetBytes(l)
	}
}

func BenchmarkMongoDbBuffer(b *testing.B) {
	mongodb, err := NewMongoDb("mongodb://localhost/test", "test", nil, Log, 30*time.Second)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	err = mongodb.Drop()
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	logger := New(
		NewOutBuffer(
			NewGeneric(mongodb).F(DefFormatter),
			b.N/2,
		),
		false,
	).Domain("test")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Print(msg)
		b.SetBytes(l)
	}
	b.StopTimer()
	err = mongodb.Drop()
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
}

func BenchmarkLogOuterNull(b *testing.B) {
	buf := bytes.NewBuffer([]byte{})
	//backend := NewMulti(NewWriter(buf), DefFormatter, NewWriter(os.Stdout), DefFormatter)
	backend := NewWriter(buf).F(DefFormatter)
	olog, ok := backend.(OuterLogger)
	if !ok {
		return
	}
	w := olog.OuterLog("tag")
	defer olog.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, err := w.Write([]byte(msg + "\n"))
		if err != nil {
			b.Error(e.Trace(e.Forward(err)))
		}
		if int64(n) != l+1 {
			b.Error("write failed", n, l)
		}
		// b.StopTimer()
		// time.Sleep(5 * time.Second)
		// testb(b, buf, msg)
		b.SetBytes(l)
		//b.StartTimer()
	}
}

func BenchmarkLogOuterFile(b *testing.B) {
	name, err := rand.FileName("BenchmarkLogFile", ".log", 10)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	name = os.TempDir() + "/" + name
	file, err := os.Create(name)
	if err != nil {
		b.Error(e.Trace(e.Forward(err)))
	}
	defer os.Remove(name)
	defer file.Close()

	backend := NewWriter(file).F(DefFormatter)
	olog, ok := backend.(OuterLogger)
	if !ok {
		return
	}
	w := olog.OuterLog("tag")
	defer olog.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, err := w.Write([]byte(msg + "\n"))
		if err != nil {
			b.Error(e.Trace(e.Forward(err)))
		}
		if int64(n) != l+1 {
			b.Error("write failed", n, l)
		}
		b.SetBytes(l)
	}
}
