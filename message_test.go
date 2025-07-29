package syslog

import (
	"github.com/rickb777/expect"
	"testing"
	"time"
)

func TestMessage_String(t *testing.T) {
	tx := time.Date(2023, 10, 26, 15, 31, 1, 0, time.UTC)
	m := &Message{
		Time:        tx,
		Facility:    User,
		Severity:    Debug,
		Timestamp:   time.Date(2023, 10, 26, 15, 30, 0, 0, time.UTC),
		Hostname:    "myhost.example.com",
		Application: "myapp",
		PID:         12345,
		Tag:         `[example@32473 eventSource="system"]`,
		Content:     `This is a sample syslog message`,
	}
	expect.String(m.String()).ToBe(t,
		`2023-10-26 15:31:01 10-26 15:30:00 <user,debug> myhost.example.com myapp 12345 [example@32473 eventSource="system"] This is a sample syslog message`)
}

//func TestParseMessage(t *testing.T) {
//	tx := time.Date(2023, 10, 26, 15, 31, 1, 0, time.UTC)
//	now = func() time.Time {
//		return tx
//	}
//
//	input := "<165>1 2023-10-26T15:30:00.000Z myhost.example.com myapp 12345 [example@32473 eventSource=\"system\"] This is a sample syslog message"
//	m := parseMessage([]byte(input), nil, isNotAlnum)
//	expect.Any(m).ToBe(t, &Message{
//		Time:        tx,
//		Facility:    User,
//		Severity:    Debug,
//		Timestamp:   time.Date(2023, 10, 26, 15, 30, 0, 0, time.UTC),
//		Hostname:    "myhost.example.com",
//		Application: "myapp",
//		PID:         12345,
//		Tag:         `[example@32473 eventSource="system"]`,
//		Content:     `This is a sample syslog message`,
//	})
//}
