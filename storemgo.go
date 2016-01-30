// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/fcavani/e"
	"github.com/fcavani/types"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func getKey(i interface{}) string {
	val := reflect.ValueOf(i)
	val = reflect.Indirect(val)
	t := val.Type()
	switch t.Kind() {
	case reflect.String:
		return val.String()
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tags := f.Tag.Get("bson")
			if tags == "" {
				continue
			}
			if tags != "key" {
				continue
			}
			vf := val.Field(i)
			if vf.Kind() != reflect.String {
				continue
			}
			return vf.String()
		}
		return ""
	default:
		return ""
	}
}

const ErrDataNotComp = "data type is not compatible, use a struct"

func chkData(i interface{}) error {
	val := reflect.Indirect(reflect.ValueOf(i))
	if val.Kind() != reflect.Struct {
		return e.New(ErrDataNotComp)
	}
	return nil
}

type MongoDb struct {
	session    *mgo.Session
	db         *mgo.Database
	dbname     string
	c          *mgo.Collection
	collection string
	tentry     reflect.Type
}

func NewMongoDb(rawurl, collection string, safe *mgo.Safe, entry Entry, timeout time.Duration) (Storer, error) {
	var err error
	m := &MongoDb{}
	err = chkData(entry)
	if err != nil {
		return nil, e.Forward(err)
	}
	m.tentry = reflect.TypeOf(entry)
	if m.tentry.Kind() != reflect.Ptr {
		m.tentry = reflect.PtrTo(m.tentry)
	}
	m.session, err = mgo.DialWithTimeout(rawurl, timeout)
	if err != nil {
		return nil, e.New(err)
	}
	parsed, err := url.Parse(rawurl)
	if err != nil {
		return nil, e.New(err)
	}
	if safe == nil {
		safe = &mgo.Safe{}
	}
	m.session.SetSafe(safe)
	m.dbname = strings.Trim(parsed.Path, "/")
	m.db = m.session.DB(m.dbname)
	m.collection = collection
	m.c = m.db.C(collection)
	return m, nil
}

func (m *MongoDb) SupportTx() bool {
	return false
}

type txMongoDb struct {
	c        *mgo.Collection
	tentry   reflect.Type
	writeble bool
}

func (t *txMongoDb) Put(key string, data interface{}) error {
	if !t.writeble {
		return e.New(ErrReadOnly)
	}
	n, err := t.c.Find(bson.M{"key": key}).Count()
	if err != nil {
		return e.New(err)
	}
	if n >= 1 {
		err := t.c.Update(bson.M{"key": key}, data)
		if e.Contains(err, "not found") {
			return e.New(ErrKeyNotFound)
		} else if err != nil {
			return e.New(err)
		}
		return nil
	}
	err = t.c.Insert(data)
	if err != nil {
		return e.New(err)
	}
	return nil
}

func (t *txMongoDb) Get(key string) (interface{}, error) {
	val := types.Make(t.tentry)
	inter := val.Interface()
	err := t.c.Find(bson.M{"key": key}).One(inter)
	if _, ok := err.(*mgo.QueryError); ok {
		return nil, e.New(ErrKeyNotFound)
	} else if e.Contains(err, "not found") {
		return nil, e.New(ErrKeyNotFound)
	} else if err != nil {
		return nil, e.New(err)
	}
	return inter, nil
}

func (t *txMongoDb) Del(key string) error {
	if !t.writeble {
		return e.New(ErrReadOnly)
	}
	err := t.c.Remove(bson.M{"key": key})
	if e.Equal(err, mgo.ErrNotFound) {
		return e.New(ErrKeyNotFound)
	} else if err != nil {
		return e.New(err)
	}
	return nil
}

type cursorMongoDb struct {
	iter     *mgo.Iter
	c        *mgo.Collection
	tentry   reflect.Type
	writeble bool
	dir      dir
}

type dir uint8

const (
	LeftToRight dir = iota
	RightToLeft
)

func (c *cursorMongoDb) First() (key string, data interface{}) {
	c.iter = c.c.Find(nil).Sort("key").Iter()
	c.dir = LeftToRight

	val := types.Make(c.tentry)
	inter := val.Interface()

	c.iter.Next(inter)

	key = getKey(inter)
	if key == "" {
		return "", nil
	}
	return key, inter
}

func (c *cursorMongoDb) Last() (key string, data interface{}) {
	c.iter = c.c.Find(nil).Sort("-key").Iter()
	c.dir = RightToLeft

	val := types.Make(c.tentry)
	inter := val.Interface()

	c.iter.Next(inter)

	key = getKey(inter)
	if key == "" {
		return "", nil
	}
	return key, inter
}

func (c *cursorMongoDb) Seek(wanted string) (key string, data interface{}) {
	c.iter = c.c.Find(bson.M{"key": bson.M{"$gte": wanted}}).Sort("key").Iter()
	c.dir = LeftToRight

	val := types.Make(c.tentry)
	inter := val.Interface()

	c.iter.Next(inter)

	key = getKey(inter)
	if key == "" {
		return "", nil
	}
	return key, inter
}

func (c *cursorMongoDb) Next() (key string, data interface{}) {
	if c.iter == nil {
		return "", nil
	}
	if c.dir != LeftToRight {
		return "", nil
	}
	val := types.Make(c.tentry)
	inter := val.Interface()
	c.iter.Next(inter)
	key = getKey(inter)
	if key == "" {
		return "", nil
	}
	return key, inter
}

func (c *cursorMongoDb) Prev() (key string, data interface{}) {
	if c.iter == nil {
		return "", nil
	}
	if c.dir != RightToLeft {
		return "", nil
	}
	val := types.Make(c.tentry)
	inter := val.Interface()
	c.iter.Next(inter)
	key = getKey(inter)
	if key == "" {
		return "", nil
	}
	return key, inter
}

func (c *cursorMongoDb) Del() (err error) {
	if !c.writeble {
		return e.New(ErrReadOnly)
	}
	panic("not implemented")
}

func (t *txMongoDb) Cursor() Cursor {
	c := &cursorMongoDb{
		c:        t.c,
		tentry:   t.tentry,
		writeble: t.writeble,
	}
	return c
}

func (m *MongoDb) Len() (l uint, err error) {
	n, err := m.c.Count()
	if err != nil {
		return 0, e.New(err)
	}
	return uint(n), nil
}

func (m *MongoDb) Tx(write bool, f func(tx Transaction) error) error {
	tx := &txMongoDb{
		c:        m.c,
		tentry:   m.tentry,
		writeble: write,
	}
	err := f(tx)
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

// Drop clears the database
func (m *MongoDb) Drop() error {
	err := m.db.DropDatabase()
	if err != nil {
		return e.New(err)
	}
	m.db = m.session.DB(m.dbname)
	m.c = m.db.C(m.collection)
	return nil
}

func (m *MongoDb) Close() error {
	m.session.Close()
	return nil
}
