**NOTE:** This is the @chrissnell fork of syslog.  This fork differs from @ziutek's original version in the following ways:

- It has support for RFC 5424-style syslog packets
- This version supports "extended" (non-alphanumeric) characters in the syslog tag field.  This breaks RFC spec but is useful for creating tags like "apache-access-log-prod".  These characters are specified in a string passed to Server.AddAllowedRunes().   Example:
```
	s := syslog.NewServer()

    // Allows dashes, periods, and underscores in the syslog tag field
	s.AddAllowedRunes("-._")

	s.AddHandler(newHandler())
	s.Listen(*listenAddrPtr)
```


About
-----
Using this library you can easy implement your own syslog server that:

1. Can listen on multiple UDP ports and unix domain sockets.

2. Can pass parsed syslog messages to your own handlers so your code can analyze
and respond for them.

See [documentation](http://gopkgdoc.appspot.com/pkg/github.com/ziutek/syslog)
and [example server](https://github.com/ziutek/syslog/blob/master/example_server/main.go).
