# syslog

[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/rickb777/syslog)
[![Go Report Card](https://goreportcard.com/badge/github.com/rickb777/syslog)](https://goreportcard.com/report/github.com/rickb777/syslog)
[![Issues](https://img.shields.io/github/issues/rickb777/syslog.svg)](https://github.com/rickb777/syslog/issues)

**NOTE:** This is from the @chrissnell fork of syslog.  This fork differs from @ziutek's original version in the following ways:

- It has support for both RFC 5424 and RFC 3164 syslog packets; note that RFC 3164 packets are not clearly defined.

```
	s := syslog.NewServer()
	s.AddHandler(newHandler())
	s.Listen(*listenAddrPtr)
```


About
-----
Using this library you can easy implement your own syslog server that:

1. Can listen on specified UDP ports and Unix domain sockets.

2. Can pass parsed Syslog messages to your own handlers so your code can analyze and respond to them.

See the [example server](https://github.com/rickb777/syslog/blob/master/example_server/main.go).
