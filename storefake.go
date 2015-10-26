// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"github.com/fcavani/e"
)

type StoreFake struct{}

func (s StoreFake) SupportTx() bool {
	return false
}

func (s StoreFake) Tx(write bool, f func(tx Transaction) error) error {
	return e.New("store not implemented")
}
