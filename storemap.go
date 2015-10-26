// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"sort"
	"sync"

	"github.com/fcavani/e"
)

const ErrKeyFound = "key found"
const ErrKeyNotFound = "key not found"
const ErrReadOnly = "read only transaction"

// func searchString(idx []string, key string) int {
// 	return sort.Search(len(idx), func(i int) bool {
// 		numKey, er1 := strconv.ParseInt(key, 10, 64)
// 		numIdx, er2 := strconv.ParseInt(idx[i], 10, 64)
// 		if er1 == nil && er2 == nil {
// 			return numIdx >= numKey
// 		}
// 		return idx[i] >= key
// 	})
// }

func searchString(idx []string, key string) int {
	return sort.SearchStrings(idx, key)
}

type Map struct {
	M   map[string]interface{}
	Idx []string
	lck sync.RWMutex
}

func NewMap(size int) (Storer, error) {
	return &Map{
		M:   make(map[string]interface{}, size),
		Idx: make([]string, 0, size),
	}, nil
}

func (m *Map) SupportTx() bool {
	return false
}

type transaction struct {
	m     map[string]interface{}
	idx   *[]string
	write bool
}

const ErrInvKey = "invalid key"

func (t *transaction) Put(key string, data interface{}) error {
	if !t.write {
		return e.New(ErrReadOnly)
	}
	if key == "" {
		return e.New(ErrInvKey)
	}
	t.m[key] = data
	idx := *t.idx
	defer func() {
		*t.idx = idx
	}()
	if len(idx) == 0 {
		idx = append(idx, key)
		return nil
	}
	i := searchString(idx, key)
	if i >= len(idx) || idx[i] != key {
		if i+1 > len(idx) {
			idx = append(idx, key)
			return nil
		}
		idx = append(idx[:i], append([]string{key}, idx[i:]...)...)
	}
	return nil
}

func (t *transaction) Get(key string) (interface{}, error) {
	data, found := t.m[key]
	if !found {
		return nil, e.New(ErrKeyNotFound)
	}
	return data, nil
}

func (t *transaction) Del(key string) error {
	if !t.write {
		return e.New(ErrReadOnly)
	}
	if _, found := t.m[key]; !found {
		return e.New(ErrKeyNotFound)
	}
	idx := *t.idx
	defer func() {
		*t.idx = idx
	}()
	i := searchString(idx, key)
	if i >= len(idx) || idx[i] != key {
		return e.New(ErrKeyNotFound)
	}
	if i+1 >= len(idx) {
		idx = idx[:i]
	} else {
		idx = append(idx[:i], idx[i+1:]...)
	}
	delete(t.m, key)
	return nil
}

type cursor struct {
	ptr   int
	m     map[string]interface{}
	idx   *[]string
	write bool
}

func (c *cursor) First() (key string, data interface{}) {
	idx := *c.idx
	if len(idx) == 0 {
		return "", nil
	}
	c.ptr = 0
	key = idx[c.ptr]
	data = c.m[key]
	return
}

func (c *cursor) Last() (key string, data interface{}) {
	idx := *c.idx
	if len(idx) == 0 {
		return "", nil
	}
	c.ptr = len(idx) - 1
	key = idx[c.ptr]
	data = c.m[key]
	return
}

func (c *cursor) Seek(wanted string) (key string, data interface{}) {
	idx := *c.idx
	if len(idx) == 0 {
		return "", nil
	}
	i := searchString(idx, wanted)
	if i >= len(idx) {
		return "", nil
	}
	c.ptr = i
	key = idx[i]
	data = c.m[key]
	return
}

func (c *cursor) Next() (key string, data interface{}) {
	idx := *c.idx
	c.ptr++
	if c.ptr >= len(idx) {
		return "", nil
	}
	key = idx[c.ptr]
	data = c.m[key]
	return
}

func (c *cursor) Prev() (key string, data interface{}) {
	idx := *c.idx
	c.ptr--
	if c.ptr < 0 || c.ptr >= len(idx) {
		return "", nil
	}
	key = idx[c.ptr]
	data = c.m[key]
	return
}

func (c *cursor) Del() error {
	if !c.write {
		return e.New(ErrReadOnly)
	}
	idx := *c.idx
	defer func() {
		*c.idx = idx
	}()
	if c.ptr >= len(idx) {
		return e.New(ErrKeyNotFound)
	}
	key := idx[c.ptr]
	if _, found := c.m[key]; !found {
		return e.New(ErrKeyNotFound)
	}
	if c.ptr+1 > len(idx) {
		idx = idx[:c.ptr]
	} else {
		idx = append(idx[:c.ptr], idx[c.ptr+1:]...)
	}
	delete(c.m, key)
	c.ptr--
	return nil
}

func (t *transaction) Cursor() Cursor {
	return &cursor{
		m:     t.m,
		idx:   t.idx,
		write: t.write,
	}
}

func (m *Map) Len() (uint, error) {
	return uint(len(m.Idx)), nil
}

func (m *Map) Tx(write bool, f func(tx Transaction) error) error {
	if write {
		m.lck.Lock()
		defer m.lck.Unlock()
	} else {
		m.lck.RLock()
		defer m.lck.RUnlock()
	}
	t := &transaction{
		m:     m.M,
		idx:   &m.Idx,
		write: write,
	}
	err := f(t)
	if err != nil {
		return err
	}
	return nil
}