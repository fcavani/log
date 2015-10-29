// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"log/syslog"
)

// Syslog sends all messages to syslog.
type Syslog struct {
	w *syslog.Writer
}

func NewSyslog(w *syslog.Writer) LogBackend {
	return &Syslog{
		w: w,
	}
}

// F: syslog don't need a formatter.
func (s *Syslog) F(f Formatter) LogBackend {
	return s
}

// GetF always return nil, syslog don't need a formatter.
func (s *Syslog) GetF() Formatter {
	return nil
}

func (s *Syslog) Commit(entry Entry) {
	switch entry.Level() {
	case ProtoPrio:
		err := s.w.Debug(entry.Message())
		if err != nil {
			CommitFail(entry, err)
		}
	case DebugPrio:
		err := s.w.Debug(entry.Message())
		if err != nil {
			CommitFail(entry, err)
		}
	case InfoPrio:
		err := s.w.Info(entry.Message())
		if err != nil {
			CommitFail(entry, err)
		}
	case WarnPrio:
		err := s.w.Warning(entry.Message())
		if err != nil {
			CommitFail(entry, err)
		}
	case ErrorPrio:
		err := s.w.Err(entry.Message())
		if err != nil {
			CommitFail(entry, err)
		}
	case FatalPrio:
		err := s.w.Crit(entry.Message())
		if err != nil {
			CommitFail(entry, err)
		}
	case PanicPrio:
		err := s.w.Emerg(entry.Message())
		if err != nil {
			CommitFail(entry, err)
		}
	case NoPrio:
		err := s.w.Notice(entry.Message())
		if err != nil {
			CommitFail(entry, err)
		}
	default:
		err := s.w.Notice(entry.Message())
		if err != nil {
			CommitFail(entry, err)
		}
	}
}
