// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

type Level uint8

const (
	ProtoPrio Level = iota
	DebugPrio
	InfoPrio
	WarnPrio
	ErrorPrio
	FatalPrio
	PanicPrio
	NoPrio
)

func (l Level) String() string {
	switch l {
	case ProtoPrio:
		return "protocol"
	case DebugPrio:
		return "debug"
	case InfoPrio:
		return "info"
	case WarnPrio:
		return "warning"
	case ErrorPrio:
		return "error"
	case FatalPrio:
		return "fatal"
	case PanicPrio:
		return "panic"
	case NoPrio:
		return "no priority"
	default:
		panic("this isn't a priority")
	}
}
