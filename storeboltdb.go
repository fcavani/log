// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"bytes"
	"encoding/gob"
	"os"
	"time"
	
	"github.com/boltdb/bolt"
	"github.com/fcavani/e"
	"github.com/fcavani/buffactory"
	"github.com/fcavani/types"
)

var bufmaker *buffactory.BufferFactory

func init() {
	bufmaker = &buffactory.BufferFactory{
		NumBuffersPerSize: 100,
		MinBuffers: 10,
		MaxBuffers: 1000,
		MinBufferSize: 256,
		MaxBufferSize: 1024,
		Reposition: 10 *time.Second,
	}
	err := bufmaker.StartBufferFactory()
	if err != nil {
		Fatal("StartBufferFactory failed:", err)
	}
}

type Encoder interface {
	Encode(i interface{}) ([]byte, error)
}

type Decoder interface {
	Decode(buf []byte) (interface{}, error)
}

type Gob struct {
	TypeName string
}

func (g *Gob) Encode(i interface{}) ([]byte, error) {
	buf := bufmaker.RequestBuffer(1024)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(i)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (g *Gob) Decode(b []byte) (interface{}, error) {
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	val := types.Make(types.Type(g.TypeName))
	err := dec.DecodeValue(val)
	if err != nil {
		return nil, e.New(err)
	}
	return val.Interface(), nil
}

type BoltDb struct {
	db *bolt.DB
	bucket string
	enc Encoder
	dec Decoder
	path string
	mode os.FileMode
	options *bolt.Options
}

func NewBoltDb(bucket, path string, mode os.FileMode, options *bolt.Options, enc Encoder, dec Decoder) (Storer, error) {
	var err error
	b := new(BoltDb)
	b.bucket = bucket
	b.path = path
	b.mode = mode
	b.options = options
	b.enc = enc
	b.dec = dec
	b.db, err = bolt.Open(path, mode, options)
	if err != nil {
		return nil, e.New(err)
	}
	return b, nil
}

func (b *BoltDb) SupportTx() bool {
	return true
}

type txBoltDb struct {
	b *bolt.Bucket
	enc Encoder
	dec Decoder
	bufs [][]byte
}

func (t *txBoltDb) Put(key string, data interface{}) error {
	buf, err := t.enc.Encode(data)
	if err != nil {
		return e.Forward(err)
	}
	defer func() {
		t.bufs = append(t.bufs, buf)
	}()
	err = t.b.Put([]byte(key), buf)
	if e.Contains(err, "tx not writable") {
		return e.New(ErrReadOnly)
	} else if err != nil {
		return e.New(err)
	}
	return nil
}

func (t *txBoltDb) Get(key string) (interface{}, error) {
	buf := t.b.Get([]byte(key))
	if buf == nil {
		return nil, e.New(ErrKeyNotFound)
	}
	data, err := t.dec.Decode(buf)
	if err != nil {
		return nil, e.Forward(err)
	}
	return data, nil
}

func (t *txBoltDb) Del(key string) error {
	k := []byte(key)
	if !t.b.Writable() {
		return e.New(ErrReadOnly)
	}
	buf := t.b.Get(k)
	if buf == nil {
		return e.New(ErrKeyNotFound)
	}
	err := t.b.Delete(k)
	if err != nil {
		return e.New(err)
	}
	return nil
}

type cursorBoltDb struct {
	c *bolt.Cursor
	b *bolt.Bucket
	dec Decoder
}

func (c *cursorBoltDb) First() (key string, data interface{}) {
	var err error
	k, v := c.c.First()
	if k == nil {
		return "", nil
	}
	key = string(k)
	data, err = c.dec.Decode(v)
	if err != nil {
		return "", nil
	}
	return
}

func (c *cursorBoltDb) Last() (key string, data interface{}) {
	var err error
	k, v := c.c.Last()
	if k == nil {
		return "", nil
	}
	key = string(k)
	data, err = c.dec.Decode(v)
	if err != nil {
		return "", nil
	}
	return
}

func (c *cursorBoltDb) Seek(wanted string) (key string, data interface{}) {
	var err error
	k, v := c.c.Seek([]byte(wanted))
	if k == nil {
		return "", nil
	}
	key = string(k)
	data, err = c.dec.Decode(v)
	if err != nil {
		return "", nil
	}
	return
}

func (c *cursorBoltDb) Next() (key string, data interface{}) {
	var err error
	k, v := c.c.Next()
	if k == nil {
		return "", nil
	}
	key = string(k)
	data, err = c.dec.Decode(v)
	if err != nil {
		return "", nil
	}
	return
}

func (c *cursorBoltDb) Prev() (key string, data interface{}) {
	var err error
	k, v := c.c.Prev()
	if k == nil {
		return "", nil
	}
	key = string(k)
	data, err = c.dec.Decode(v)
	if err != nil {
		return "", nil
	}
	return
}

func (c *cursorBoltDb) Del() (err error) {
	defer func() {
		if r := recover(); r != nil {
			er, ok := r.(error)
			if !ok {
				return
			}
			if e.Contains(er, "runtime error: index out of range") {
				err = e.New(ErrKeyNotFound)
				return
			}
		}
	}()
	if !c.b.Writable() {
		return e.New(ErrReadOnly)
	}
	err = c.c.Delete()
	if err != nil {
		return e.New(err)
	}
	return
}

func (t *txBoltDb) Cursor() Cursor {
	c := new(cursorBoltDb)
	c.dec = t.dec
	c.c = t.b.Cursor()
	c.b = t.b
	return c
}

func (db *BoltDb) Len() (l uint, err error) {
	err = db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.bucket))
		if b == nil {
			return e.New("error when checking len")
		}
		l = uint(b.Stats().KeyN)
		return nil
	})
	err = e.Forward(err)
	return
}

func (db *BoltDb) Tx(write bool, f func(tx Transaction) error) error {
	trans := new(txBoltDb)
	trans.enc = db.enc
	trans.dec = db.dec
	trans.bufs = make([][]byte, 100)
	tx, err := db.db.Begin(write)
	if err != nil {
		return e.New(err)
	}
	if write {
		trans.b, err = tx.CreateBucketIfNotExists([]byte(db.bucket))
		if err != nil {
			return e.New(err)
		}
	} else {
		trans.b = tx.Bucket([]byte(db.bucket))
		if trans.b == nil {
			return e.New("error creating transaction")
		}
	}
	err = f(trans)
	defer func() {
		for _, buf := range trans.bufs {
			bufmaker.Return(buf)
		}
		trans.bufs = trans.bufs[:0]
	}()
	if err != nil {
		er := tx.Rollback()
		if er != nil {
			return e.Push(err, er)
		}
		return e.New(err)
	}
	if write {
		err = tx.Commit()
		if err != nil {
			return e.New(err)
		}
	} else {
		err = tx.Rollback()
		if err != nil {
			return e.New(err)
		}
	}
	return nil
}

func (db *BoltDb) Drop() error {
	err := db.db.Close()
	if err != nil {
		return e.Forward(err)
	}
	db.db, err = bolt.Open(db.path, db.mode, db.options)
	if err != nil {
		return e.New(err)
	}
	return nil
}
