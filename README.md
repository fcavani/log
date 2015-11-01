#Log

[![Build Status](https://travis-ci.org/fcavani/log.svg?branch=master)](https://travis-ci.org/fcavani/log) [![GoDoc](https://godoc.org/github.com/fcavani/log?status.svg)](https://godoc.org/github.com/fcavani/log)

Log package is much like the go log package. But it have a fill more tricks, like backends and filters.

#Normal use

First import: `import "github.com/fcavani/log"`
Than use one of the free functions in the [documentation](https://godoc.org/github.com/fcavani/log).

```
// Normal log
log.Println("log this")
// Log error
log.Error("Error:", err)
// Log fatal, fallowed by a os.Exit(1) call
log.Fatal("Can't handle this error:", err)
// Panic, calls panic after log. Maybe because the backend the log may get lost.
log.Panic("panic!")
```

Modificators:

```
// Associate a tag with the log.
log.Tag("tag").Println("test")
// More tags
log.Tag("tag1", "tag2").Println("test")
// Determine the level
log.ProtoLevel().Println("some dirty protocol thing")
```

Setting the default level:

```
log.SetLevel(log.WarnPrio)
```

#Change the log format

You can change the format of the log entry. In NewStdFormatter we have
the separator, the template, with fields named after the struct that implements
[Entry interface](https://godoc.org/github.com/fcavani/log#Entry), a sample of that struct,
a map with values that appears in template string but not appears in the struct,
like host in the example below.

```
form, _ := log.NewStdFormatter(
  "::",
  "::host - ::domain - ::date - ::level - ::tags - ::file ::msg",
  log.Log,
  map[string]interface{}{
    "host": hostname,
  }
)
```

To use this format: `log.Log.Formatter(form)`

#Considerations about speed

Below is the table with the go benchmark for some loggers packages
([PureGolog](https://golang.org/pkg/log/), [GoLogging](https://github.com/op/go-logging), [Logrus](https://github.com/Sirupsen/logrus))
and for some cases with different backends.
All three loggers above was set to log to /dev/null, and use the most simple
configuration that I can find, thus they will have good performance, with is
the case for the go log, it is the most simple and the base line for the others.
In some cases this logger can win Logrus, the slower logger between the
other two, but I doubt that StoreMap, that stores
all logged data into memory, will have practical use by anyone.

```
BenchmarkPureGolog-4    	 1000000	      1465 ns/op	  12.28 MB/s
BenchmarkGoLogging-4    	 1000000	      2235 ns/op	   8.05 MB/s
BenchmarkLogrus-4       	  300000	      5167 ns/op	   3.48 MB/s
BenchmarkGolog-4        	  200000	      6483 ns/op	   2.78 MB/s
BenchmarkLogStderr-4    	  200000	      5641 ns/op	   3.19 MB/s
BenchmarkLogFile-4      	  200000	      8276 ns/op	   2.17 MB/s
BenchmarkLogFileBuffer-4	  300000	      4838 ns/op	   3.72 MB/s
BenchmarkStoreMap-4     	  500000	      3749 ns/op	   4.80 MB/s
BenchmarkBoltDb-4       	    3000	    552564 ns/op	   0.03 MB/s
BenchmarkBoltDbBuffer-4 	    5000	    270518 ns/op	   0.07 MB/s
BenchmarkMongoDb-4      	    2000	   1052129 ns/op	   0.02 MB/s
BenchmarkMongoDbBuffer-4	     100	  11864291 ns/op	   0.00 MB/s
BenchmarkLogOuterNull-4	    200000	      8006 ns/op	   2.25 MB/s
BenchmarkLogOuterFile-4	    100000	     15351 ns/op	   1.17 MB/s
```

LogFileBuffer is interesting if you can setup a buffer larger enough to
accommodate all income data without saturate the buffer. In this tests the
buffer was set to half the size of b.N, the number of runs in the second
column. Because this the buffer is relevant until it is full in the half
of the test than the backend operate normally, without buffer, to the end
of the test. Thus we have in the third column one number that shows the
buffer working but shows too the backend working without buffer.

Using buffers in log application create a latency between the time the
event occurs and the time that someone or some log analyser will see the
event, for that reason it is not suitable for real time application. But for small
applications, that latency not will compromise the operation, even the
file log without buffer will work.

For the database (BoltDb and MongoDb), I don't have information if this numbers
correspond to the reality. In the case of MongoDb with buffers the number of
runs is too small, the buffer size become a problem and make the backend
work slower.

The last two rows are for cases where you want to plug one logger to this one.
I use this to redirect the [yamux](https://github.com/hashicorp/yamux)
logger to my logger.

This logger can offer much customizations, you can make a new entry, a new formatter
and backends. You can mix all backends with all filters and make a great custom
logger easily, but its slow. :) It's the first version too.

#Backends

The backend is responsable to put the log entry in some place. The backends avalible
are:

* `NewSendToLogger(logger *golog.Logger) LogBackend` - Only to demonstrate,
send all log entries to go logger.
* `NewWriter(w io.Writer) LogBackend` - Log to a writer. It can be a file or
anything.
* `NewGeneric(s Storer) LogBackend` - Log to anything that implements a [Storer
interface](https://godoc.org/github.com/fcavani/log#Storer).
* `NewSyslog(w *syslog.Writer) LogBackend` - Log to syslog.
* `NewMulti(vals ...interface{}) LogBackend` - Log the data to multiples backends.
  The syntax is: first the backend followed by the formattter, than another
  backend and follows like this.
* `NewOutBuffer(bak LogBackend, size int) LogBackend` - NewOutBuffer creates a
buffer between the bak backend and the commit of a new log entry. It can improve
the latency of commit but delays the final store, with can't cause log miss
if the application shutdown improperly (without close the buffer) or panic
before the buffer become empty.

Anything that implements the [LogBackend interface](https://godoc.org/github.com/fcavani/log#LogBackend)
can be used to store the log entry.

## Example 1

```
// Write to stdout with DefFormatter formatter.
Log = log.New(
  log.NewWriter(os.Stdout).F(log.DefFormatter),
  false,
)
```

## Example 2
```
// Write to stdout and to MongoDb
mongodb, _ = log.NewMongoDb(...)
Log = log.New(
  log.NewMulti(
    log.NewWriter(os.Stdout),
    log.DefFormatter,
    log.NewGeneric(mongodb),
    log.DefFormatter,
  ),
  false,
)
```  

#Filters

With filter you can chose what you will see in each backend.
In the example below if the field msg not (log.Not)
contains (log.Cnts) "not log" the message will be logged. The principal function
is the `log.Op(o Operation, field string, vleft ...interface{}) Ruler`.
In this function you can use [others operations like Cnts](https://godoc.org/github.com/fcavani/log#Operation)
and you can modify the result of it with the function `log.Not` or you can
combine the result with others Op functions with `log.And` and `log.Or`.
Like the backends anything that implements [Ruler interface](https://godoc.org/github.com/fcavani/log#Ruler)
can be used to filter one log entry.

## Example

```
Log = log.New(
  log.Filter(
    log.NewWriter(buf),
    log.Not(log.Op(log.Cnts, "msg", "not log")),
  ).F(log.DefFormatter),
  false,
)
```

#Storer

Stores with `NewGeneric(s Storer)` can put the logs entries in any place for
future analysis. All stores must respect the [Storer interface](https://godoc.org/github.com/fcavani/log#Storer),
that is a simple CRUD like interface with the DB inspired in the BoltDb API.
This is the stores available:

* [MongoDB](https://godoc.org/github.com/fcavani/log#MongoDb)
* [BoltDB](https://godoc.org/github.com/fcavani/log#BoltDb)
* [Map](https://godoc.org/github.com/fcavani/log#Map): That is a storer that uses
go map to store log entries.
