// Copyright 2015 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package log

import (
	"io"
	"time"

	"github.com/fcavani/tags"
	"github.com/go-logfmt/logfmt"
)

type Entry interface {
	// Date returns the time stamp of the log
	Date() time.Time
	// Level return the log level
	Level() Level
	// Message returns the formated message
	Message() string
	//Tags return the tags in a log entry
	Tags() *tags.Tags
	//Domain is the domain of the log
	Domain(d string) Logger
	//GetDomain return the current domain of the log.
	GetDomain() string
	// Error returns any erro cocurred after caling one function.
	Err() error
	// String return the formated log
	String() string
	// Bytes return the formated log in bytes
	Bytes() []byte
	// Formatter sets the formater for that entry
	Formatter(f Formatter)
	// Sorter set one filter for the backend associated with the logger.
	// This filter works after the filter set in the New statment.
	Sorter(r Ruler) Logger
	// SetLevel sets the log Level for this logger. Scope all setlevel for everything.
	// If Scope is a packege set log level only for this package.
	SetLevel(scope string, l Level) Logger
	// EntryLevel set the level for this log entry.
	EntryLevel(l Level) Logger
	// DebugInfo write into the struct debug information.
	DebugInfo() Logger
}

type TemplateSetup interface {
	// Mark change the replacement mark of the template
	Mark(mark string) Logger
	// Template replaces the template string
	Template(t string) Logger
}

type Formatter interface {
	// Format formats the template replace the marks with the contents of m.
	Format(entry Entry) (out []byte, err error)
	// Mark change the replacement mark of the template
	Mark(mark string)
	// Template replaces the template string
	Template(t string)
	// Entry mark this formater to work only its type of entry
	Entry(entry Entry)
	// NewEntry creates a new log entry
	NewEntry(b LogBackend) Logger
	// Set the time format string if empty use the default.
	SetTimeFormat(s string)
}

type LogBackend interface {
	// Commit send the log to the persistence layer.
	Commit(entry Entry)
	// F sets the formater for this backend.
	F(f Formatter) LogBackend
	// GetF returns the Formatter for this backend.
	GetF() Formatter
	//Filter change the filter associated to this backend
	Filter(r Ruler) LogBackend
	//Close stop the backend and flush all entries.
	Close() error
}

type Cursor interface {
	First() (key string, data interface{})
	Last() (key string, data interface{})
	Seek(wanted string) (key string, data interface{})
	Next() (key string, data interface{})
	Prev() (key string, data interface{})
	Del() error
}

type Transaction interface {
	Put(key string, data interface{}) error
	Get(key string) (interface{}, error)
	Del(key string) error
	Cursor() Cursor
}

type Storer interface {
	SupportTx() bool
	Tx(write bool, f func(tx Transaction) error) error
	Len() (uint, error)
	Drop() error
	Close() error
}

type Levels interface {
	// ProtoLevel set the log level to protocol
	ProtoLevel() Logger
	// DebugLevel set the log level to debug
	DebugLevel() Logger
	// InfoLevel set the log level to info
	InfoLevel() Logger
	// WarnLevel set the log level to warn
	WarnLevel() Logger
	// ErrorLevel set the log level to error
	ErrorLevel() Logger
	// FatalLevel set the log level to fatal
	FatalLevel() Logger
	// PanicLevel set the log level to panic
	PanicLevel() Logger
}

type Tagger interface {
	// Tag attach a tag
	Tag(tags ...string) Logger
}
type Storage interface {
	// Store give access to the persistence storage
	Store() LogBackend
	// SetStore allow you to set a new persistence store to the logger
	SetStore(p LogBackend) Logger
}

type StdLogger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})

	Panic(...interface{})
	Panicf(string, ...interface{})
	Panicln(...interface{})
}

type PanicStack interface {
	// GoPanic handle a panic where r is the value of recover()
	// stack is the buffer where will be the stack dump and cont is false if
	// GoPanic will call os.Exit(1).
	GoPanic(r interface{}, stack []byte, cont bool)
}

// OtherLogger provides a interface to plug via a writer another logger to this
// logger, in this case the backend that implements OtherLogger
type OuterLogger interface {
	// OtherLog creats a writer that receive log entries separeted by \n.
	OuterLog(level Level, tags ...string) io.Writer
	// Close closses the outer logger. If not closed you will have a leeked gorotine.
	Close() error
}

type Ruler interface {
	Result(entry Entry) bool
}

type Logger interface {
	Entry
	Levels
	Tagger
	TemplateSetup
	StdLogger
	Storage
	PanicStack
	Error(...interface{})
	Errorf(string, ...interface{})
	Errorln(...interface{})
}

// Logfmter encode a log entry in logfmt format.
type Logfmter interface {
	Logfmt(enc *logfmt.Encoder) error
}

// go test -bench=. -cpu=1,4 -benchmem
// go test -gcflags=-m -bench=Something
