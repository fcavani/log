// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import "github.com/fcavani/e"

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

func ParseLevel(level string) (Level, error) {
	switch level {
	case "protocol":
		return ProtoPrio, nil
	case "debug":
		return DebugPrio, nil
	case "info":
		return InfoPrio, nil
	case "warning":
		return WarnPrio, nil
	case "error":
		return ErrorPrio, nil
	case "fatal":
		return FatalPrio, nil
	case "panic":
		return PanicPrio, nil
	case "no priority":
		return NoPrio, nil
	default:
		return NoPrio, e.New("invalid priority")
	}
}
