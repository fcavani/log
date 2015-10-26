// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/fcavani/e"
	"github.com/fcavani/rand"
)

type store struct {
	Name  string
	Store Storer
}

var maps []*store

func addstore(t *testing.T, f interface{}, params ...interface{}) {
	vals := make([]reflect.Value, len(params))
	for i, param := range params {
		vals[i] = reflect.ValueOf(param)
	}
	cons := reflect.ValueOf(f)
	if maps == nil {
		maps = make([]*store, 0)
	}
	retvals := cons.Call(vals)
	if len(retvals) != 2 {
		t.Fatal("call to constructor fail", len(retvals))
	}
	if !retvals[1].IsValid() {
		t.Fatal("return value is invalid")
	}
	if retvals[1].Interface() != nil {
		err := retvals[1].Interface().(error)
		if err != nil {
			t.Fatal("Constructor failed:", e.Trace(e.Forward(err)))
		}
	}
	if !retvals[0].IsValid() {
		t.Fatal("return value is invalid")
	}
	if retvals[0].Interface() == nil {
		t.Fatal("constructor returned a nil storer")
	}
	s := retvals[0].Interface().(Storer)
	name := ""
	if retvals[0].Kind() == reflect.Interface {
		name = retvals[0].Elem().Type().String()
	}
	if name == "" {
		t.Fatal("Storer have no name?")
	}

	maps = append(maps, &store{
		Name:  name,
		Store: s,
	})
}

func empty(t *testing.T, s Storer) {
	err := s.Tx(true, func(tx Transaction) error {
		_, err := tx.Get("none")
		return err
	})
	if err != nil && !e.Equal(err, ErrKeyNotFound) {
		t.Fatal(e.Trace(e.Forward(err)))
	} else if err == nil {
		t.Fatal("error is nil")
	}

	err = s.Tx(true, func(tx Transaction) error {
		return tx.Del("none")
	})
	if err != nil && !e.Equal(err, ErrKeyNotFound) {
		t.Fatal(e.Trace(e.Forward(err)))
	} else if err == nil {
		t.Fatal("error is nil")
	}

	err = s.Tx(true, func(tx Transaction) error {
		return tx.Del("none")
	})
	if err != nil && !e.Equal(err, ErrKeyNotFound) {
		t.Fatal(e.Trace(e.Forward(err)))
	} else if err == nil {
		t.Fatal("error is nil")
	}

	err = s.Tx(true, func(tx Transaction) error {
		c := tx.Cursor()
		key, data := c.First()
		if key != "" || data != nil {
			return e.New("not empty")
		}
		key, data = c.Last()
		if key != "" || data != nil {
			return e.New("not empty")
		}
		key, data = c.Next()
		if key != "" || data != nil {
			return e.New("not empty")
		}
		key, data = c.Prev()
		if key != "" || data != nil {
			return e.New("not empty")
		}
		key, data = c.Seek("ing for gophers?")
		if key != "" || data != nil {
			return e.New("not empty")
		}
		return nil
	})
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}

	err = s.Tx(true, func(tx Transaction) error {
		c := tx.Cursor()
		return c.Del()
	})
	if err != nil && !e.Equal(err, ErrKeyNotFound) {
		t.Fatal(e.Trace(e.Forward(err)))
	} else if err == nil {
		t.Fatal("error is nil")
	}

	var l uint
	l, err = s.Len()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if l != 0 {
		t.Fatal("wrong length", l)
	}
}

func readonly(t *testing.T, s Storer) {
	err := s.Tx(false, func(tx Transaction) error {
		return tx.Put("none", "foo")
	})
	if err != nil && !e.Equal(err, ErrReadOnly) {
		t.Fatal(e.Trace(e.Forward(err)))
	} else if err == nil {
		t.Fatal("error is nil")
	}

	err = s.Tx(false, func(tx Transaction) error {
		return tx.Del("none")
	})
	if err != nil && !e.Equal(err, ErrReadOnly) {
		t.Fatal(e.Trace(e.Forward(err)))
	} else if err == nil {
		t.Fatal("error is nil")
	}

	err = s.Tx(false, func(tx Transaction) error {
		c := tx.Cursor()
		return c.Del()
	})
	if err != nil && !e.Equal(err, ErrReadOnly) {
		t.Fatal(e.Trace(e.Forward(err)))
	} else if err == nil {
		t.Fatal("error is nil")
	}
}

func put(t *testing.T, s Storer, num int) {
	err := s.Tx(true, func(tx Transaction) error {
		var err error
		for i := 0; i < num; i++ {
			key := strconv.Itoa(i)
			err = tx.Put(key, i)
			if err != nil {
				return err
			}
		}
		err = tx.Put("3a", 3)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	l, err := s.Len()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if l != uint(num+1) {
		t.Fatalf("wrong len %v", l)
	}
}

func get(t *testing.T, s Storer, num int) {
	err := s.Tx(false, func(tx Transaction) error {
		for i := 0; i < num; i++ {
			key := strconv.Itoa(i)
			data, err := tx.Get(key)
			if err != nil {
				return err
			}
			ii := data.(int)
			if ii != i {
				return e.New("retrieve wrong data %v %v %v", key, i, ii)
			}
		}
		data, err := tx.Get("3a")
		if err != nil {
			return err
		}
		ii := data.(int)
		if ii != 3 {
			return e.New("retrieve wrong data %v", ii)
		}
		return nil
	})
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
}

func del(t *testing.T, s Storer, num int) {
	err := s.Tx(true, func(tx Transaction) error {
		var err error
		for i := 0; i < num; i++ {
			key := strconv.Itoa(i)
			err = tx.Del(key)
			if err != nil {
				return err
			}
		}
		err = tx.Del("3a")
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	l, err := s.Len()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if l != 0 {
		t.Fatalf("something remining %v", l)
	}
}

func iter(t *testing.T, s Storer, num int) {
	err := s.Tx(true, func(tx Transaction) error {
		err := tx.Del("3a")
		if err != nil {
			return err
		}
		c := tx.Cursor()
		i := 1
		kk, _ := c.First()
		for k, _ := c.Next(); k != ""; k, _ = c.Next() {
			if k <=  kk {
				return e.New("retrieve wrong key %v", k)
			}
			kk = k
			i++
		}
		if i != num {
			t.Fatal("cursor didn't run", i)
		}
		kk, _ = c.Last()
		for k, _ := c.Last(); k != ""; k, _ = c.Prev() {
			if k > kk {
				return e.New("retrieve wrong key %v", k)
			}
			kk = k
			i--
		}
		if i != 0 {
			t.Fatal("cursor didn't run", i)
		}
		k, v := c.Seek("80")
		if k != "80" {
			return e.New("Seek failed %v", k)
		}
		data := v.(int)
		if data != 80 {
			return e.New("Seek failed %v", data)
		}
		err = c.Del()
		if err != nil {
			return e.Forward(err)
		}
		_, err = tx.Get("80")
		if err != nil && !e.Equal(err, ErrKeyNotFound) {
			return e.Forward(err)
		} else if err == nil {
			return e.New("returned nil")
		}
		k, v = c.Seek("80")
		if k != "81" {
			return e.New("Seek failed %v", k)
		}
		k, v = c.Seek("zzzzzz")
		if k != "" {
			return e.New("Seek failed %v", k)
		}
		i = 0
		for k, _ := c.First(); k != ""; k, _ = c.First() {
			err = c.Del()
			if err != nil {
				return e.Forward(err)
			}
			i++
		}
		if i != 99 {
			return e.New("cursor didn't run %v", i)
		}
		return nil
	})
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	l, err := s.Len()
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	if l != 0 {
		t.Fatalf("something remining %v", l)
	}
}

func testsort(t *testing.T, s Storer) {
	for i := 0; i < 1000; i++ {
		err := s.Tx(true, func(tx Transaction) error {
			var err error
			out := make(rand.StringPermutation, len(rand.LowerCase))
			err = rand.RandomPermutation(rand.StringPermutation(rand.LowerCase), out, "go")
			if err != nil {
				return nil
			}
			//t.Log(out)
			for i, key := range out {
				err = tx.Put(key, i)
				if err != nil {
					return err
				}
			}
			//var l uint
			// l, err = tx.Len()
			// if err != nil {
			// 	return err
			// }
			// if l != uint(len(out)) {
			// 	return e.New("wrong len %v", l)
			// }
			c := tx.Cursor()
			prev, _ := c.First()
			for k, _ := c.First(); k != ""; k, _ = c.Next() {
				if k < prev {
					t.Log("not in alphabetic sequence")
					t.Fail()
				}
				err := c.Del()
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(e.Trace(e.Forward(err)))
		}
	}
}
