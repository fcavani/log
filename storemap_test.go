// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"testing"
)

const numregs = 100

func TestMap(t *testing.T) {
	addstore(t, NewMap, numregs)
}

func TestStores(t *testing.T) {
	for _, store := range maps {
		t.Log("Testing store:", store.Name)
		empty(t, store.Store)
		readonly(t, store.Store)
		put(t, store.Store, numregs)
		get(t, store.Store, numregs)
		del(t, store.Store, numregs)
		put(t, store.Store, numregs)
		iter(t, store.Store, numregs)
		testsort(t, store.Store)
	}
}
