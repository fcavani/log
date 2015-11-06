// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"regexp"
	"testing"
	"time"

	"github.com/fcavani/e"
	"github.com/fcavani/tags"
)

type testEntry struct {
	Str      string  `log:"str"`
	Float    float32 `log:"float"`
	Integer  int     `log:"int"`
	Uinteger uint    `log:"uint"`
	Boolean  bool    `log:"bool"`
	Invalid  struct {
		A int
		B float32
	} `log:"invalid"`
	Time   time.Time  `log:"time"`
	Labels *tags.Tags `log:"tags"`
}

func (t *testEntry) Date() time.Time {
	return time.Now()
}

func (t *testEntry) Level() Level {
	return FatalPrio
}

func (t *testEntry) Message() string {
	return ""
}

func (t *testEntry) Tags() *tags.Tags {
	return nil
}

func (t *testEntry) Domain(d string) Logger {
	return nil
}

func (t *testEntry) GetDomain() string {
	return ""
}

func (t *testEntry) Err() error {
	return nil
}

func (t *testEntry) String() string {
	return ""
}

func (t *testEntry) Bytes() []byte {
	return nil
}

func (t *testEntry) Formatter(f Formatter) {
	return
}

func (t *testEntry) SetLevel(scope string, level Level) Logger {
	return nil
}

func (t *testEntry) Sorter(r Ruler) Logger {
	return nil
}

func (t *testEntry) EntryLevel(l Level) Logger {
	return nil
}

func TestInvalidOperation(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			if r.(string) != "logger: invalid operation" {
				t.Fatal(r)
			}
		}
	}()
	ruler := Op(Operation(69), "bool")
	ruler.Result(&testEntry{})
}

func TestInvalidType(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			if r.(string) != "logger: struct is not time.Time" {
				t.Fatal(r)
			}
		}
	}()
	ruler := Op(Lt, "invalid", struct {
		A int
		B float32
	}{2, 3})
	ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{2, 3},
	})
}

func TestStructEqual(t *testing.T) {
	ruler := Op(Eq, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r := ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{2, 3},
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Eq, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{4, 2},
	})
	if r {
		t.Fatal("result is invalid")
	}

}

func TestStructNotEqual(t *testing.T) {
	ruler := Op(Ne, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r := ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{2, 3},
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ne, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{4, 2},
	})
	if !r {
		t.Fatal("result is invalid")
	}
}

func TestTags(t *testing.T) {
	ruler := Op(Ex, "tags", "teste")
	tags, _ := tags.NewTags("tag1, tag2, teste")
	r := ruler.Result(&testEntry{
		Labels: tags,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ex, "tags", "sbravists")
	r = ruler.Result(&testEntry{
		Labels: tags,
	})
	if r {
		t.Fatal("result is invalid")
	}
}

func TestCnts(t *testing.T) {
	ruler := Op(Cnts, "str", "teste")
	r := ruler.Result(&testEntry{
		Str: "isto é apenas um teste",
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Cnts, "str", "meleca")
	r = ruler.Result(&testEntry{
		Str: "isto é apenas um teste",
	})
	if r {
		t.Fatal("result is invalid")
	}
}

func TestPr(t *testing.T) {
	ruler := Op(Pr, "str", "isto é")
	r := ruler.Result(&testEntry{
		Str: "isto é apenas um teste",
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Pr, "str", "meleca")
	r = ruler.Result(&testEntry{
		Str: "isto é apenas um teste",
	})
	if r {
		t.Fatal("result is invalid")
	}
}

func TestRe(t *testing.T) {
	ruler := Op(Re, "str", `([a-zA-Z0-9]+)([.-_][a-zA-Z0-9]+)*@([a-zA-Z0-9]+)([.-_][a-zA-Z0-9]+)*`)
	r := ruler.Result(&testEntry{
		Str: "name@isp.com",
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Re, "str", `([a-zA-Z0-9]+)([.-_][a-zA-Z0-9]+)*@([a-zA-Z0-9]+)([.-_][a-zA-Z0-9]+)*`)
	r = ruler.Result(&testEntry{
		Str: "isto é apenas um teste",
	})
	if r {
		t.Fatal("result is invalid")
	}

	re := regexp.MustCompile(`([a-zA-Z0-9]+)([.-_][a-zA-Z0-9]+)*@([a-zA-Z0-9]+)([.-_][a-zA-Z0-9]+)*`)
	ruler = Op(Re, "str", re)
	r = ruler.Result(&testEntry{
		Str: "name@isp.com",
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Re, "str", re)
	r = ruler.Result(&testEntry{
		Str: "isto é apenas um teste",
	})
	if r {
		t.Fatal("result is invalid")
	}
}

func TestEq(t *testing.T) {
	ruler := Op(Eq, "bool", true)
	r := ruler.Result(&testEntry{
		Boolean: true,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Eq, "bool", true)
	r = ruler.Result(&testEntry{
		Boolean: false,
	})
	if r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Eq, "int", 1)
	r = ruler.Result(&testEntry{
		Integer: 1,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Eq, "int", 1)
	r = ruler.Result(&testEntry{
		Integer: 2,
	})
	if r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Eq, "uint", uint(1))
	r = ruler.Result(&testEntry{
		Uinteger: 1,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Eq, "uint", uint(1))
	r = ruler.Result(&testEntry{
		Uinteger: 2,
	})
	if r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Eq, "float", float32(1))
	r = ruler.Result(&testEntry{
		Float: 1,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Eq, "float", float32(1))
	r = ruler.Result(&testEntry{
		Float: 2,
	})
	if r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Eq, "str", "teste")
	r = ruler.Result(&testEntry{
		Str: "teste",
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Eq, "str", "teste")
	r = ruler.Result(&testEntry{
		Str: "no teste",
	})
	if r {
		t.Fatal("result is invalid")
	}

	now := time.Now()
	ruler = Op(Eq, "time", now)
	r = ruler.Result(&testEntry{
		Time: now,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Eq, "time", now)
	r = ruler.Result(&testEntry{
		Time: time.Now(),
	})
	if r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Eq, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{2, 3},
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Eq, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{4, 2},
	})
	if r {
		t.Fatal("result is invalid")
	}

}

func TestNe(t *testing.T) {
	ruler := Op(Ne, "bool", true)
	r := ruler.Result(&testEntry{
		Boolean: true,
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ne, "bool", true)
	r = ruler.Result(&testEntry{
		Boolean: false,
	})
	if !r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Ne, "int", 1)
	r = ruler.Result(&testEntry{
		Integer: 1,
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ne, "int", 1)
	r = ruler.Result(&testEntry{
		Integer: 2,
	})
	if !r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Ne, "uint", uint(1))
	r = ruler.Result(&testEntry{
		Uinteger: 1,
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ne, "uint", uint(1))
	r = ruler.Result(&testEntry{
		Uinteger: 2,
	})
	if !r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Ne, "float", float32(1))
	r = ruler.Result(&testEntry{
		Float: 1,
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ne, "float", float32(1))
	r = ruler.Result(&testEntry{
		Float: 2,
	})
	if !r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Ne, "str", "teste")
	r = ruler.Result(&testEntry{
		Str: "teste",
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ne, "str", "teste")
	r = ruler.Result(&testEntry{
		Str: "no teste",
	})
	if !r {
		t.Fatal("result is invalid")
	}

	now := time.Now()
	ruler = Op(Ne, "time", now)
	r = ruler.Result(&testEntry{
		Time: now,
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ne, "time", now)
	r = ruler.Result(&testEntry{
		Time: time.Now(),
	})
	if !r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Ne, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{2, 3},
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ne, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{4, 2},
	})
	if !r {
		t.Fatal("result is invalid")
	}

}

func TestLt(t *testing.T) {
	ruler := Op(Lt, "int", 1)
	r := ruler.Result(&testEntry{
		Integer: 2,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Lt, "int", 1)
	r = ruler.Result(&testEntry{
		Integer: 0,
	})
	if r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Lt, "uint", uint(1))
	r = ruler.Result(&testEntry{
		Uinteger: 2,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Lt, "uint", uint(1))
	r = ruler.Result(&testEntry{
		Uinteger: 0,
	})
	if r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Lt, "float", float32(1))
	r = ruler.Result(&testEntry{
		Float: 2,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Lt, "float", float32(1))
	r = ruler.Result(&testEntry{
		Float: 0,
	})
	if r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Lt, "str", "a")
	r = ruler.Result(&testEntry{
		Str: "b",
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Lt, "str", "b")
	r = ruler.Result(&testEntry{
		Str: "a",
	})
	if r {
		t.Fatal("result is invalid")
	}

	now := time.Now()
	ruler = Op(Lt, "time", now)
	r = ruler.Result(&testEntry{
		Time: time.Now(),
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Lt, "time", time.Now())
	r = ruler.Result(&testEntry{
		Time: now,
	})
	if r {
		t.Fatal("result is invalid")
	}

	defer func() {
		r := recover()
		if r != nil && r.(string) != "logger: struct is not time.Time" {
			t.Fatal(r)
		}
	}()

	ruler = Op(Lt, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{2, 3},
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Lt, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{4, 2},
	})
	if r {
		t.Fatal("result is invalid")
	}

}

func TestGt(t *testing.T) {
	ruler := Op(Gt, "int", 1)
	r := ruler.Result(&testEntry{
		Integer: 2,
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Gt, "int", 1)
	r = ruler.Result(&testEntry{
		Integer: 0,
	})
	if !r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Gt, "uint", uint(1))
	r = ruler.Result(&testEntry{
		Uinteger: 2,
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Gt, "uint", uint(1))
	r = ruler.Result(&testEntry{
		Uinteger: 0,
	})
	if !r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Gt, "float", float32(1))
	r = ruler.Result(&testEntry{
		Float: 2,
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Gt, "float", float32(1))
	r = ruler.Result(&testEntry{
		Float: 0,
	})
	if !r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Gt, "str", "a")
	r = ruler.Result(&testEntry{
		Str: "b",
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Gt, "str", "b")
	r = ruler.Result(&testEntry{
		Str: "a",
	})
	if !r {
		t.Fatal("result is invalid")
	}

	now := time.Now()
	ruler = Op(Gt, "time", now)
	r = ruler.Result(&testEntry{
		Time: time.Now(),
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Gt, "time", time.Now())
	r = ruler.Result(&testEntry{
		Time: now,
	})
	if !r {
		t.Fatal("result is invalid")
	}

	defer func() {
		r := recover()
		if r != nil && r.(string) != "logger: struct is not time.Time" {
			t.Fatal(r)
		}
	}()

	ruler = Op(Gt, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{2, 3},
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Gt, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{4, 2},
	})
	if !r {
		t.Fatal("result is invalid")
	}
}

func TestLe(t *testing.T) {
	ruler := Op(Le, "int", 1)
	r := ruler.Result(&testEntry{
		Integer: 2,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Le, "int", 1)
	r = ruler.Result(&testEntry{
		Integer: 0,
	})
	if r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Le, "uint", uint(1))
	r = ruler.Result(&testEntry{
		Uinteger: 2,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Le, "uint", uint(1))
	r = ruler.Result(&testEntry{
		Uinteger: 0,
	})
	if r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Le, "float", float32(1))
	r = ruler.Result(&testEntry{
		Float: 2,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Le, "float", float32(1))
	r = ruler.Result(&testEntry{
		Float: 0,
	})
	if r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Le, "str", "a")
	r = ruler.Result(&testEntry{
		Str: "b",
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Le, "str", "b")
	r = ruler.Result(&testEntry{
		Str: "a",
	})
	if r {
		t.Fatal("result is invalid")
	}

	now := time.Now()
	ruler = Op(Le, "time", now)
	r = ruler.Result(&testEntry{
		Time: time.Now(),
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Le, "time", time.Now())
	r = ruler.Result(&testEntry{
		Time: now,
	})
	if r {
		t.Fatal("result is invalid")
	}

	defer func() {
		r := recover()
		if r != nil && r.(string) != "logger: struct is not time.Time" {
			t.Fatal(r)
		}
	}()

	ruler = Op(Le, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{2, 3},
	})
	if !r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Le, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{4, 2},
	})
	if r {
		t.Fatal("result is invalid")
	}
}

func TestGe(t *testing.T) {
	ruler := Op(Ge, "int", 1)
	r := ruler.Result(&testEntry{
		Integer: 2,
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ge, "int", 1)
	r = ruler.Result(&testEntry{
		Integer: 0,
	})
	if !r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Ge, "uint", uint(1))
	r = ruler.Result(&testEntry{
		Uinteger: 2,
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ge, "uint", uint(1))
	r = ruler.Result(&testEntry{
		Uinteger: 0,
	})
	if !r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Ge, "float", float32(1))
	r = ruler.Result(&testEntry{
		Float: 2,
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ge, "float", float32(1))
	r = ruler.Result(&testEntry{
		Float: 0,
	})
	if !r {
		t.Fatal("result is invalid")
	}

	ruler = Op(Ge, "str", "a")
	r = ruler.Result(&testEntry{
		Str: "b",
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ge, "str", "b")
	r = ruler.Result(&testEntry{
		Str: "a",
	})
	if !r {
		t.Fatal("result is invalid")
	}

	now := time.Now()
	ruler = Op(Ge, "time", now)
	r = ruler.Result(&testEntry{
		Time: time.Now(),
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ge, "time", time.Now())
	r = ruler.Result(&testEntry{
		Time: now,
	})
	if !r {
		t.Fatal("result is invalid")
	}

	defer func() {
		r := recover()
		if r != nil && r.(string) != "logger: struct is not time.Time" {
			t.Fatal(r)
		}
	}()

	ruler = Op(Ge, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{2, 3},
	})
	if r {
		t.Fatal("result is invalid")
	}
	ruler = Op(Ge, "invalid", struct {
		A int
		B float32
	}{2, 3})
	r = ruler.Result(&testEntry{
		Invalid: struct {
			A int
			B float32
		}{4, 2},
	})
	if !r {
		t.Fatal("result is invalid")
	}
}

func TestN(t *testing.T) {
	ruler := Op(N, "bool")
	r := ruler.Result(&testEntry{
		Boolean: false,
	})
	if !r {
		t.Fatal("result is invalid")
	}
}

func TestAnd(t *testing.T) {
	ruler := And(Op(Eq, "bool", true), Op(Cnts, "str", "teste"))
	r := ruler.Result(&testEntry{
		Str:     "isto é apena um teste",
		Boolean: true,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	r = ruler.Result(&testEntry{
		Str:     "isto é apena um bo",
		Boolean: true,
	})
	if r {
		t.Fatal("result is invalid")
	}
	r = ruler.Result(&testEntry{
		Str:     "isto é apena um teste",
		Boolean: false,
	})
	if r {
		t.Fatal("result is invalid")
	}
	r = ruler.Result(&testEntry{
		Str:     "isto é apena um bo",
		Boolean: false,
	})
	if r {
		t.Fatal("result is invalid")
	}
}

func TestOr(t *testing.T) {
	ruler := Or(Op(Eq, "bool", true), Op(Cnts, "str", "teste"))
	r := ruler.Result(&testEntry{
		Str:     "isto é apena um teste",
		Boolean: true,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	r = ruler.Result(&testEntry{
		Str:     "isto é apena um bo",
		Boolean: true,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	r = ruler.Result(&testEntry{
		Str:     "isto é apena um teste",
		Boolean: false,
	})
	if !r {
		t.Fatal("result is invalid")
	}
	r = ruler.Result(&testEntry{
		Str:     "isto é apena um bo",
		Boolean: false,
	})
	if r {
		t.Fatal("result is invalid")
	}
}

func TestNot(t *testing.T) {
	ruler := Not(Op(Eq, "bool", true))
	r := ruler.Result(&testEntry{
		Boolean: true,
	})
	if r {
		t.Fatal("result is invalid")
	}
	r = ruler.Result(&testEntry{
		Boolean: false,
	})
	if !r {
		t.Fatal("result is invalid")
	}
}

func TestFilter(t *testing.T) {
	Log = New(
		Filter(
			NewWriter(buf),
			Not(Op(Cnts, "msg", "not log")),
		).F(DefFormatter),
		false,
	).Domain("test")

	Println("log")
	test(t, buf, "log")

	Println("not log")
	err := testerr(buf, "not log")
	if err != nil && !e.Contains(err, "EOF") {
		t.Fatal(e.Trace(e.Forward(err)))
	} else if err == nil {
		t.Fatal("nil")
	}

	Println("log2")
	test(t, buf, "log2")

}
