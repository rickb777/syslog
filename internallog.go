package syslog

// Logger is an interface for package internal logging.
// This deliberately matches the API of [log.Logger].
type Logger interface {
	Print(...interface{})
	Printf(format string, v ...interface{})
	Println(...interface{})
	Fatal(...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(...interface{})
}
