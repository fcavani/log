// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"bytes"
	"testing"

	"github.com/fcavani/e"
)

func TestLogfmt(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	Log = New(NewLogfmt(buf), false).Domain("test").Tag("tag1")
	Log.Println("test logfmt")
	str, err := buf.ReadString('\n')
	if err != nil {
		t.Fatal(e.Trace(e.Forward(err)))
	}
	t.Log(str)
}
