// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"testing"
	"time"

	"gopkg.in/mgo.v2"
)

const numregs = 100

func TestMap(t *testing.T) {
	addstore(t, NewMap, numregs)
}

func TestMgo(t *testing.T) {
	t.SkipNow()
	addstore(t, NewMongoDb, "mongodb://localhost/test", "test", &mgo.Safe{}, &TestStruct{}, 30*time.Second)
}

func TestStores(t *testing.T) {
	for _, store := range maps {
		t.Log("Testing store:", store.Name)
		dropeverthing(t, store.Store)
		empty(t, store.Store)
		readonly(t, store.Store)
		put(t, store.Store, numregs)
		length(t, store.Store, numregs+1)
		get(t, store.Store, numregs)
		del(t, store.Store, numregs)
		put(t, store.Store, numregs)
		iter(t, store.Store, numregs)
		testsort(t, store.Store)
		testiter2(t, store.Store)
		dropeverthing(t, store.Store)
	}
}
