// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/fcavani/e"
	"github.com/fcavani/utilitybelt/deepcopy"
)

const ErrNoSubs = "variable %v have no substitution"
const ErrNotSupported = "this format isn't made for this entry"

var TimeDateFormat = time.RFC822

func mkindex(entry Entry) (m map[string]struct {
	I   int
	Def string
}) {
	val := reflect.Indirect(reflect.ValueOf(entry))
	t := val.Type()
	m = make(map[string]struct {
		I   int
		Def string
	}, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("log")
		if tag == "" {
			continue
		}
		def := f.Tag.Get("def")

		m[tag] = struct {
			I   int
			Def string
		}{i, def}
	}
	return
}

func scapemark(in []byte) []byte {
	buf := make([]byte, len(in)*2)
	j := 0
	for i := 0; i < len(buf); i += 2 {
		buf[i] = '\\'
		buf[i+1] = in[j]
		j++
	}
	return buf
}

func strinter(val reflect.Value) (str string, err error) {
	inter := val.Interface()
	if inter == nil {
		return "", e.New("interface is nil")
	}
	switch i := inter.(type) {
	case time.Time:
		str = i.Format(TimeDateFormat)
		return
	case fmt.Stringer:
		str = i.String()
		return
	}
	return "", e.New("interface don't implement String method")
}

func stringfy(val reflect.Value) (str string) {
	var err error
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(val.Int(), 10)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str = strconv.FormatUint(val.Uint(), 10)
	case reflect.Uint8:
		str, err = strinter(val)
		if err == nil {
			return
		}
		str = strconv.FormatUint(val.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(val.Float(), 'f', 2, 10)
	case reflect.String:
		str = val.String()
	default:
		str, _ = strinter(val)
	}
	return
}

func scapeSep(in string, sep []byte) string {
	scape := scapemark(sep)
	return strings.Replace(in, string(sep), string(scape), -1)
}

// StdFormatter is a formatter for the log data. The fild in Entry
// are match with the fields in Tmpl, the tags in Entry are considered.
type StdFormatter struct {
	// Delim: every fild in Tmpl are preceded by it.
	Delim []byte
	// Tmpl is the template and are compose by deliminator fallowed by labels
	Tmpl []byte
	// E is the Entry. This fild are only used for struct analasy of E.
	E Entry
	// Map holds the replacements for labels that aren't found in E.
	Map map[string]interface{}
	// Idx are the index of the fild. Don't change.
	Idx map[string]struct {
		I   int
		Def string
	}
}

func init() {
	gob.Register(&StdFormatter{})
}

// NewStdFormatter crete a new formatter.
func NewStdFormatter(delim, tmpl string, entry Entry, values map[string]interface{}) (Formatter, error) {
	if delim == "" {
		return nil, e.New("invalid delimitator")
	}
	if tmpl == "" {
		return nil, e.New("invalid template")
	}
	if entry == nil {
		return nil, e.New("invalid entry")
	}
	if values == nil {
		return nil, e.New("invalid values")
	}
	return &StdFormatter{
		Delim: []byte(delim),
		Tmpl:  []byte(tmpl),
		E:     entry,
		Map:   values,
		Idx:   mkindex(entry),
	}, nil
}

func (s *StdFormatter) Mark(delim string) {
	s.Delim = []byte(delim)
}

func (s *StdFormatter) Template(t string) {
	s.Tmpl = []byte(t)
}

func findCut(val []byte) int {
	for i, v := range val {
		if v == ' ' || v == '\n' {
			return i
		}
	}
	return -1
}

func (s StdFormatter) Format(entry Entry) (out []byte, err error) {
	val := reflect.Indirect(reflect.ValueOf(entry))
	if val.Kind() != reflect.Struct {
		return nil, e.New("formater only accept entries that are structs ")
	}

	if val.Type() != reflect.Indirect(reflect.ValueOf(s.E)).Type() {
		return nil, e.New(ErrNotSupported)
	}

	out = make([]byte, len(s.Tmpl), len(s.Tmpl)*2)
	copy(out, s.Tmpl)
	for {
		start := bytes.Index(out, s.Delim)
		if start == -1 {
			break
		}
		var end int
		cut := out[start:]
		i := findCut(cut)
		if i > -1 {
			end = start + i
		} else {
			end = start + len(cut)
		}

		name := out[start+len(s.Delim) : end]

		var v string
		bname := string(name)
		fidx, found := s.Idx[bname]
		if found {
			fval := val.Field(fidx.I)
			v = scapeSep(stringfy(fval), s.Delim)
		} else {
			inter, found := s.Map[bname]
			if !found {
				v = ""
			} else {
				fval := reflect.Indirect(reflect.ValueOf(inter))
				v = scapeSep(stringfy(fval), s.Delim)
			}
		}
		if v == "" {
			v = fidx.Def
		}

		vb := []byte(v)

		out = append(out[:start], append(vb, out[end:]...)...)
	}

	scape := scapemark(s.Delim)
	out = bytes.Replace(out, scape, s.Delim, -1)
	out = bytes.Replace(out, []byte{' ', ' '}, []byte{' '}, -1)
	return
}

func (s *StdFormatter) Entry(entry Entry) {
	s.E = entry
}

func (s *StdFormatter) NewEntry(b LogBackend) Logger {
	return deepcopy.Iface(s.E).(Logger).SetStore(b)
}
