// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"testing"
	"os"
	
	"github.com/fcavani/rand"
	"github.com/fcavani/e"
	"github.com/boltdb/bolt"
	"github.com/fcavani/types"
)

func TestBoltDb(t *testing.T) {
	name, err := rand.FileName("boltdb", ".db", 10)
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	name = os.TempDir() + "/" + name
	gob := &Gob{
		TypeName: types.Name(int(0)),
	}
	addstore(t, NewBoltDb, "test", name, os.FileMode(0600), &bolt.Options{}, gob, gob)
}

