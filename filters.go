// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/fcavani/tags"
)

type filter struct {
	LogBackend
	r Ruler
}

// Filter creates a new filter with rules r for l backend.
func Filter(l LogBackend, r Ruler) LogBackend {
	return &filter{
		LogBackend: l,
		r:          r,
	}
}

func (f *filter) Commit(entry Entry) {
	if f.r.Result(entry) {
		f.LogBackend.Commit(entry)
	}
}

func (f *filter) F(formatter Formatter) LogBackend {
	f.LogBackend.F(formatter)
	return f
}

// Operation defines one operator for the rules.
type Operation uint8

const (
	// Equal
	Eq Operation = iota
	// Not equal
	Ne
	// Less than
	Lt
	// Greater than
	Gt
	// Less equal
	Le
	// Greater equal
	Ge
	// Not
	N
	// Exits in tags
	Ex
	// Contains
	Cnts
	// Regexp
	Re
	// Pr matches the begin of string
	Pr
)

type op struct {
	field  string
	vright reflect.Value
	op     Operation
}

func mapfieldmap(entry Entry) (m map[string]reflect.Value) {
	ve := reflect.Indirect(reflect.ValueOf(entry))
	if ve.Kind() != reflect.Struct {
		panic("logger: formater only accept entries that are structs ")
	}
	te := ve.Type()
	m = make(map[string]reflect.Value, te.NumField())
	for i := 0; i < te.NumField(); i++ {
		fte := te.Field(i)
		tag := fte.Tag.Get("log")
		if tag == "" {
			continue
		}
		vf := ve.Field(i)
		if !vf.IsValid() {
			continue
		}
		if !vf.CanSet() {
			panic("logger: the field must be exported!")
		}
		val := reflect.Indirect(ve.Field(i))
		if val.Kind() == reflect.Interface {
			val = val.Elem()
		}
		m[tag] = val
	}
	return
}

func (o op) Result(entry Entry) bool {
	mr := mapfieldmap(entry)
	vleft, found := mr[o.field]
	if !found {
		panic("logger: field name not found in entry struct")
	}
	if o.op != N && o.op != Ex && o.op != Re && o.vright.IsValid() && vleft.Type() != o.vright.Type() {
		panic("logger: type of vleft is not equal to the type of entry")
	}
	switch o.op {
	case Eq:
		switch vleft.Kind() {
		case reflect.Bool:
			return vleft.Bool() == o.vright.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return vleft.Int() == o.vright.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return vleft.Uint() == o.vright.Uint()
		case reflect.Float32, reflect.Float64:
			return vleft.Float() == o.vright.Float()
		case reflect.String:
			return vleft.String() == o.vright.String()
		case reflect.Struct:
			tl, okl := vleft.Interface().(time.Time)
			tr, okr := o.vright.Interface().(time.Time)
			if !okl || !okr {
				return reflect.DeepEqual(vleft.Interface(), o.vright.Interface())
			}
			return tl.Equal(tr)
		default:
			return reflect.DeepEqual(vleft.Interface(), o.vright.Interface())
		}
	case Ne:
		switch vleft.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return vleft.Int() != o.vright.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return vleft.Uint() != o.vright.Uint()
		case reflect.Float32, reflect.Float64:
			return vleft.Float() != o.vright.Float()
		case reflect.String:
			return vleft.String() != o.vright.String()
		case reflect.Struct:
			tl, okl := vleft.Interface().(time.Time)
			tr, okr := o.vright.Interface().(time.Time)
			if !okl || !okr {
				return !reflect.DeepEqual(vleft.Interface(), o.vright.Interface())
			}
			return !tl.Equal(tr)
		default:
			return !reflect.DeepEqual(vleft.Interface(), o.vright.Interface())
		}
	case Lt:
		switch vleft.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return vleft.Int() < o.vright.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return vleft.Uint() < o.vright.Uint()
		case reflect.Float32, reflect.Float64:
			return vleft.Float() < o.vright.Float()
		case reflect.String:
			return vleft.String() < o.vright.String()
		case reflect.Struct:
			tl, okl := vleft.Interface().(time.Time)
			tr, okr := o.vright.Interface().(time.Time)
			if !okl || !okr {
				panic("logger: struct is not time.Time")
			}
			return tl.Before(tr)
		default:
			panic("logger: field type of entry is not supported")
		}
	case Gt:
		switch vleft.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return vleft.Int() > o.vright.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return vleft.Uint() > o.vright.Uint()
		case reflect.Float32, reflect.Float64:
			return vleft.Float() > o.vright.Float()
		case reflect.String:
			return vleft.String() > o.vright.String()
		case reflect.Struct:
			tl, okl := vleft.Interface().(time.Time)
			tr, okr := o.vright.Interface().(time.Time)
			if !okl || !okr {
				panic("logger: struct is not time.Time")
			}
			return tl.After(tr)
		default:
			panic("logger: field type of entry is not supported")
		}
	case Le:
		switch vleft.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return vleft.Int() <= o.vright.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return vleft.Uint() <= o.vright.Uint()
		case reflect.Float32, reflect.Float64:
			return vleft.Float() <= o.vright.Float()
		case reflect.String:
			return vleft.String() <= o.vright.String()
		case reflect.Struct:
			tl, okl := vleft.Interface().(time.Time)
			tr, okr := o.vright.Interface().(time.Time)
			if !okl || !okr {
				panic("logger: struct is not time.Time")
			}
			return tl.Before(tr) || tl.Equal(tr)
		default:
			panic("logger: field type of entry is not supported")
		}
	case Ge:
		switch vleft.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return vleft.Int() >= o.vright.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return vleft.Uint() >= o.vright.Uint()
		case reflect.Float32, reflect.Float64:
			return vleft.Float() >= o.vright.Float()
		case reflect.String:
			return vleft.String() >= o.vright.String()
		case reflect.Struct:
			tl, okl := vleft.Interface().(time.Time)
			tr, okr := o.vright.Interface().(time.Time)
			if !okl || !okr {
				panic("logger: struct is not time.Time")
			}
			return tl.After(tr) || tl.Equal(tr)
		default:
			panic("logger: field type of entry is not supported")
		}
	case N:
		return vleft.Kind() == reflect.Bool && !vleft.Bool()
	case Ex:
		switch vleft.Kind() {
		case reflect.Slice:
			if o.vright.Kind() != reflect.String {
				panic("logger: contains only works with vleft of string type")
			}
			tag := o.vright.String()
			tagsr, okr := vleft.Interface().(tags.Tags)
			if !okr {
				panic("logger: o.vright must be of *tags.Tags")
			}
			ptr := &tagsr
			return ptr.Exist(tag)
		default:
			panic("logger: field type of entry is not supported")
		}
	case Cnts:
		switch vleft.Kind() {
		case reflect.String:
			if o.vright.Kind() != reflect.String {
				panic("logger: contains only works with vleft of string type")
			}
			return strings.Contains(vleft.String(), o.vright.String())
		default:
			panic("logger: field type of entry is not supported")
		}
	case Re:
		switch vleft.Kind() {
		case reflect.String:
			if o.vright.Kind() == reflect.String {
				re := regexp.MustCompile(o.vright.String())
				return re.MatchString(vleft.String())
			} else if o.vright.Kind() == reflect.Struct {
				re, ok := o.vright.Interface().(regexp.Regexp)
				if !ok {
					panic("logger: Re operator: struct isn't of type *regexp.Regexp")
				}
				pre := &re
				return pre.MatchString(vleft.String())
			}
			panic("logger: contains only works with vleft of string type or *regexp.Regexp")
		default:
			panic("logger: field type of entry is not supported")
		}
	case Pr:
		switch vleft.Kind() {
		case reflect.String:
			if o.vright.Kind() != reflect.String {
				panic("logger: begins only works with vleft of string type")
			}
			return strings.HasPrefix(vleft.String(), o.vright.String())
		default:
			panic("logger: field type of entry is not supported")
		}
	default:
		panic("logger: invalid operation")
	}
	panic("not here")
}

// Op is an operation in some field and with some value.
func Op(o Operation, field string, vleft ...interface{}) Ruler {
	if len(vleft) > 1 {
		panic("Op accept only zero or one val")
	}
	var val reflect.Value
	if len(vleft) == 1 {
		val = reflect.Indirect(reflect.ValueOf(vleft[0]))
		if val.IsValid() && val.Kind() == reflect.Interface {
			val = val.Elem()
		}
	}
	return &op{
		field:  field,
		vright: val,
		op:     o,
	}
}

type and struct {
	rulers []Ruler
}

func (a and) Result(entry Entry) (r bool) {
	r = true
	for _, ruler := range a.rulers {
		r = r && ruler.Result(entry)
	}
	return
}

// And operator between two rules.
func And(v ...Ruler) Ruler {
	return &and{
		rulers: v,
	}
}

type or struct {
	rulers []Ruler
}

func (o or) Result(entry Entry) (r bool) {
	r = false
	for _, ruler := range o.rulers {
		r = r || ruler.Result(entry)
	}
	return
}

// Or operator between two rules.
func Or(v ...Ruler) Ruler {
	return &or{
		rulers: v,
	}
}

type not struct {
	Ruler
}

func (n not) Result(entry Entry) bool {
	return !n.Ruler.Result(entry)
}

// Not operator for one rule.
func Not(r Ruler) Ruler {
	return &not{
		Ruler: r,
	}
}

type apply struct {
	condition Ruler
	rule      Ruler
}

func (a apply) Result(entry Entry) bool {
	if a.condition.Result(entry) {
		return a.rule.Result(entry)
	}
	return true
}

// ApplyRuleIf test if condition is true than apply rule. If condition is false
// do nothing, return true.
func ApplyRuleIf(condition, rule Ruler) Ruler {
	return &apply{
		condition: condition,
		rule:      rule,
	}
}

type applyelse struct {
	condition Ruler
	rule      Ruler
	el        Ruler
}

func (a applyelse) Result(entry Entry) bool {
	if a.condition.Result(entry) {
		return a.rule.Result(entry)
	}
	return a.el.Result(entry)
}

// ApplyRuleIfElse test if condition is true than apply rule. If condition is false
// run else rule.
func ApplyRuleIfElse(condition, rule, el Ruler) Ruler {
	return &applyelse{
		condition: condition,
		rule:      rule,
		el:        el,
	}
}

// True ruler return alway true
type True struct{}

func (t True) Result(entry Entry) bool {
	return true
}

// False ruler return always false
type False struct{}

func (f False) Result(entry Entry) bool {
	return false
}

type If struct {
	Condition Ruler
	Than      Ruler
}

type sel struct {
	Ifs     []*If
	Default Ruler
}

func (s *sel) Result(entry Entry) bool {
	for _, cond := range s.Ifs {
		if cond.Condition.Result(entry) {
			return cond.Than.Result(entry)
		}
	}
	return s.Default.Result(entry)
}

func Select(ifs []*If, def Ruler) Ruler {
	return &sel{
		Ifs:     ifs,
		Default: def,
	}
}
