# syslog

[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/rickb777/syslog)
[![Go Report Card](https://goreportcard.com/badge/github.com/rickb777/syslog)](https://goreportcard.com/report/github.com/rickb777/syslog)
[![Issues](https://img.shields.io/github/issues/rickb777/syslog.svg)](https://github.com/rickb777/syslog/issues)

Using this library you can easy implement your own Syslog server that:

1. Can listen on specified UDP ports and Unix domain sockets.
2. Can listen on multiple ports/sockets simultaneously.
3. Can be easily configured to accept or ignore various Syslog messages.
4. Can pass parsed Syslog messages to your own handlers so your code can analyze and respond to them.
5. Each of your handlers can accept or ignore the messages it cares about.

See the [example server](https://github.com/rickb777/syslog/blob/master/example_server/main.go).

```go
	s := syslog.NewServer(10)
	s.AddHandler(myHandler())
	s.Listen(":1514") // receives syslog packets on UDP port 1514
```

## Earlier Work

This is a fork of the [@chrissnell fork](https://github.com/chrissnell/syslog) of syslog.  This fork differs from [@ziutek's original version](https://github.com/ziutek/syslog) in the following ways:

- It has support for both RFC-5424 and older RFC-3164 syslog packets. Although RFC-3164 packets are not well-defined, they can be handled satisfactorily nonetheless.
- It has a more flexible example server.


## About

