// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"unicode"

	"github.com/fcavani/e"
	"github.com/fcavani/text"
	uni "github.com/fcavani/unicode"
)

var MinProbeName = 1
var MaxProbeName = 50

const ErrFirstChar = "first char must be letter"

func ValProbeName(name string) error {
	if len(name) < MinProbeName || len(name) > MaxProbeName {
		return e.New(text.ErrInvNumberChars)
	}
	if !uni.IsLetter(int32(name[0])) {
		return e.New(ErrFirstChar)
	}
	for _, v := range name {
		if !uni.IsLetter(v) && !unicode.IsDigit(v) {
			return e.Push(e.New(text.ErrInvCharacter), e.New("the character '%v' is invalid", string([]byte{byte(v)})))
		}
	}
	return nil
}
