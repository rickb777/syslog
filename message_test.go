package syslog

import (
	"github.com/rickb777/expect"
	"testing"
	"time"
)

func TestMessage_String(t *testing.T) {
	tx := time.Date(2023, 10, 26, 15, 31, 1, 0, time.UTC)

	cases := []struct {
		m          Message
		expString  string
		expRFC5424 string
		expF1      string
		expF2      string
	}{
		{
			m: Message{
				Time:        tx,
				Facility:    User,
				Severity:    Debug,
				Version:     1,
				Timestamp:   time.Date(2023, 10, 26, 15, 30, 0, 0, time.UTC),
				Hostname:    "myhost.example.com",
				Application: "myapp",
				ProcID:      "12345",
				MsgID:       "m1",
				Data:        `[example@32473 eventSource="system"]`,
				Content:     `This is a sample syslog message`,
			},
			expString:  `<user,debug>1 2023-10-26T15:30:00Z myhost.example.com myapp 12345 m1 [example@32473 eventSource="system"] This is a sample syslog message`,
			expRFC5424: `<15>1 2023-10-26T15:30:00Z myhost.example.com myapp 12345 m1 [example@32473 eventSource="system"] This is a sample syslog message`,
			expF1:      `<15>1 2023-10-26T15:30:00Z myhost.example.com myapp 12345 m1 [example@32473 eventSource="system"] This is a sample syslog message`,
			expF2:      `<15>,1,2023-10-26T15:30:00Z,myhost.example.com,myapp,12345,m1,[example@32473 eventSource="system"],This is a sample syslog message`,
		},
		{
			m: Message{
				Time:        tx,
				Facility:    Auth,
				Severity:    Crit,
				Version:     0,
				Timestamp:   time.Date(2023, 10, 22, 22, 14, 15, 0, time.UTC),
				Hostname:    "mymachine",
				Application: "su",
				ProcID:      "",
				MsgID:       "",
				Data:        ``,
				Content:     `: 'su root' failed for lonvick on /dev/pts/8`,
			},
			expString:  `<auth,crit>1 2023-10-22T22:14:15Z mymachine su: 'su root' failed for lonvick on /dev/pts/8`,
			expRFC5424: `<34>1 2023-10-22T22:14:15Z mymachine su: 'su root' failed for lonvick on /dev/pts/8`,
			expF1:      `<34>Oct 22 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8`,
			expF2:      `<34>,0,Oct 22 22:14:15,mymachine,su,,,,: 'su root' failed for lonvick on /dev/pts/8`,
		},
	}
	for i, c := range cases {
		expect.String(c.m.String()).Info(i).ToBe(t, c.expString)
		expect.String(c.m.RFC5424()).Info(i).ToBe(t, c.expRFC5424)
		expect.String(c.m.Format(RFC3164Format)).Info(i).ToBe(t, c.expF1)
		expect.String(c.m.Format("<%Z>,%V,%T,%H,%A,%P,%M,%D,%C")).Info(i).ToBe(t, c.expF2)
	}
}

func TestMessage_Format(t *testing.T) {
	tx := time.Date(2023, 10, 26, 15, 31, 1, 0, time.UTC)

	m := Message{
		Time:        tx,
		Facility:    User,
		Severity:    Debug,
		Version:     0,
		Timestamp:   time.Date(2023, 10, 26, 15, 30, 0, 0, time.UTC),
		Hostname:    "myhost.example.com",
		Application: "myapp",
		ProcID:      "12345",
		MsgID:       "m1",
		Data:        `[example@32473 eventSource="system"]`,
		Content:     `This is a sample syslog message`,
	}
	cases := []struct {
		f      string
		v0, v1 string
	}{
		{f: "<%Z>", v0: "<15>", v1: "<15>"},
		{f: "%V", v0: "0", v1: "1"},
		{f: "%v", v0: "", v1: "1"},
		{f: "%Y", v0: "2023", v1: ""},
		{f: "%T", v0: "Oct 26 15:30:00", v1: "2023-10-26T15:30:00Z"},
		{f: "%H", v0: "myhost.example.com", v1: "myhost.example.com"},
		{f: "%A", v0: "myapp[12345]", v1: "myapp"},
		{f: "%P", v0: "", v1: "12345"},
		{f: "%M", v0: "m1", v1: "m1"},
		{f: "%D", v0: "[example@32473 eventSource=\"system\"]", v1: "[example@32473 eventSource=\"system\"]"},
		{f: "%C", v0: "This is a sample syslog message", v1: "This is a sample syslog message"},
		{f: "%F", v0: "user", v1: "user"},
		{f: "%S", v0: "debug", v1: "debug"},
		{f: "%%", v0: "%", v1: "%"},
	}
	for _, c := range cases {
		expect.String(m.Format(c.f)).Info(c.f).ToBe(t, c.v0)
	}
	m.Version = 1
	for _, c := range cases {
		expect.String(m.Format(c.f)).Info(c.f).ToBe(t, c.v1)
	}
}
