package syslog

import "fmt"

// Handler handles syslog messages
type Handler interface {
	// Handle should return [Message] (maybe modified) for further processing by
	// other handlers, or return nil. If Handle is called with nil message it
	// should clean up and shut down before returning nil.
	Handle(*Message) *Message
}

//-------------------------------------------------------------------------------------------------

// PrintHandler is a [Handler] that prints every message to stdout in a specified format, for
// example [RFCFormat]. See [Message.Format].
//
// Simply convert the format string to a PrintHandler to use it.
type PrintHandler string

func (p PrintHandler) Handle(m *Message) *Message {
	if m != nil {
		fmt.Println(m.Format(string(p)))
	}
	return m
}

//-------------------------------------------------------------------------------------------------

// DebugHandler is a [Handler] that simply prints every message. The message is printed in its
// internal format rather than the usual syslog format. Use this for diagnostics, for example.
type DebugHandler struct{}

func (h DebugHandler) Handle(m *Message) *Message {
	if m != nil {
		fmt.Printf("%+v\n", *m)
	}
	return m
}
