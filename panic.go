// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"os"
)

func CommitFail(entry Entry, err error) {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "no name"
	}
	dom := entry.GetDomain()
	if dom == "" {
		dom = "no domain"
	}
	println("LOG IS IN PANIC: " + err.Error() + "\n" + hostname + " - " + dom + " - " + entry.Date().String() + " - " + entry.Level().String() + " - " + entry.Tags().String() + " - " + entry.Message())
}

func Fail(err error) {
	println("LOG IS IN PANIC: " + err.Error())
}
