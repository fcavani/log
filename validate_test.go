// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"testing"

	"github.com/fcavani/e"
	"github.com/fcavani/text"
)

func TestValProbeName(t *testing.T) {
	err := ValProbeName("")
	if err != nil && !e.Equal(err, text.ErrInvNumberChars) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("nil error")
	}
	err = ValProbeName("122")
	if err != nil && !e.Equal(err, ErrFirstChar) {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("nil error")
	}
	err = ValProbeName("a_")
	if err != nil && e.Find(err, text.ErrInvCharacter) == -1 {
		t.Fatal(err)
	} else if err == nil {
		t.Fatal("nil error")
	}
	err = ValProbeName("aaaaa")
	if err != nil {
		t.Fatal(err)
	}
}
