package syslog

import (
	"github.com/rickb777/expect"
	"testing"
	"time"
)

func TestMessage_String(t *testing.T) {
	tx := time.Date(2023, 10, 26, 15, 31, 1, 0, time.UTC)

	cases := []struct {
		m         Message
		expString string
		expFormat string
	}{
		{
			m: Message{
				Time:        tx,
				Facility:    User,
				Severity:    Debug,
				Timestamp:   time.Date(2023, 10, 26, 15, 30, 0, 0, time.UTC),
				Hostname:    "myhost.example.com",
				Application: "myapp",
				ProcID:      "12345",
				MsgID:       "m1",
				Data:        `[example@32473 eventSource="system"]`,
				Content:     `This is a sample syslog message`,
			},
			expString: `<user,debug> 2023-10-26T15:30:00Z myhost.example.com myapp 12345 m1 [example@32473 eventSource="system"] This is a sample syslog message`,
			expFormat: `<15>1 2023-10-26T15:30:00Z myhost.example.com myapp 12345 m1 [example@32473 eventSource="system"] This is a sample syslog message`,
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
			expString: `<auth,crit> 2023-10-22T22:14:15Z mymachine su: 'su root' failed for lonvick on /dev/pts/8`,
			expFormat: `<34>1 2023-10-22T22:14:15Z mymachine su: 'su root' failed for lonvick on /dev/pts/8`,
		},
	}
	for i, c := range cases {
		expect.String(c.m.String()).Info(i).ToBe(t, c.expString)
		expect.String(c.m.Format()).Info(i).ToBe(t, c.expFormat)
	}
}

func TestParseMessage(t *testing.T) {
	tx := time.Date(2023, 10, 26, 15, 31, 1, 0, time.UTC)
	now = func() time.Time {
		return tx
	}

	bom := []byte{0xEF, 0xBB, 0xBF}

	cases := []struct {
		name string
		m    Message
		in   []byte
	}{
		{
			name: "RFC3164 example 1: without year",
			in:   []byte(`<34>Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8`),
			m: Message{
				Time:        tx,
				Facility:    Auth,
				Severity:    Crit,
				Version:     0,
				Timestamp:   time.Date(2023, 10, 11, 22, 14, 15, 0, time.UTC),
				Hostname:    "mymachine",
				Application: "su",
				ProcID:      "",
				MsgID:       "",
				Data:        ``,
				Content:     `: 'su root' failed for lonvick on /dev/pts/8`,
			},
		},
		{
			name: "RFC3164 example 4: with year",
			in:   []byte(`<0>1990 Oct 22 10:52:01 TZ-6 scapegoat.dmz.example.org sched[0]: That's All Folks!`),
			m: Message{
				Time:        tx,
				Facility:    Kern,
				Severity:    Emerg,
				Version:     0,
				Timestamp:   time.Date(1990, 10, 22, 10, 52, 1, 0, time.UTC),
				Hostname:    "scapegoat.dmz.example.org",
				Application: "sched",
				ProcID:      "0",
				MsgID:       "",
				Data:        ``,
				Content:     `: That's All Folks!`,
			},
		},
		{
			name: "RFC5424 example 1: with BOM but no structured data",
			in: concat([]byte(`<34>1 2003-10-11T22:14:15.003Z mymachine.example.com su - ID47 - `),
				bom, []byte("'su root' failed for lonvick on /dev/pts/8")),
			m: Message{
				Time:        tx,
				Facility:    Auth,
				Severity:    Crit,
				Version:     1,
				Timestamp:   time.Date(2003, 10, 11, 22, 14, 15, 3_000_000, time.UTC),
				Hostname:    "mymachine.example.com",
				Application: "su",
				ProcID:      "-",
				MsgID:       "ID47",
				Data:        `-`,
				Content:     `'su root' failed for lonvick on /dev/pts/8`,
			},
		},
		{
			name: "RFC5424 example 2: without structured data or BOM",
			in:   []byte(`<165>1 2003-08-24T05:14:15.000003-07:00 192.0.2.1 myproc 8710 - - %% It's time to make the donuts.`),
			m: Message{
				Time:        tx,
				Facility:    Local4,
				Severity:    Notice,
				Version:     1,
				Timestamp:   time.Date(2003, 8, 24, 5, 14, 15, 3000, time.FixedZone("", -7*60*60)),
				Hostname:    "192.0.2.1",
				Application: "myproc",
				ProcID:      "8710",
				MsgID:       `-`,
				Data:        `-`,
				Content:     `%% It's time to make the donuts.`,
			},
		},
		{
			name: "RFC5424 example 3: with BOM and structured data",
			in: concat([]byte(
				`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"] `),
				bom, []byte(`An application event log entry...`)),
			m: Message{
				Time:        tx,
				Facility:    Local4,
				Severity:    Notice,
				Version:     1,
				Timestamp:   time.Date(2003, 10, 11, 22, 14, 15, 3_000_000, time.UTC),
				Hostname:    "mymachine.example.com",
				Application: "evntslog",
				ProcID:      "-",
				MsgID:       "ID47",
				Data:        `[exampleSDID@32473 iut="3" eventSource="Application" eventID="1011"]`,
				Content:     `An application event log entry...`,
			},
		},
		{
			name: "RFC5424 example 4: with structured data only",
			in:   []byte(`<165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut="3" eventSource= "Application" eventID="1011"][examplePriority@32473 class="high"]`),
			m: Message{
				Time:        tx,
				Facility:    Local4,
				Severity:    Notice,
				Version:     1,
				Timestamp:   time.Date(2003, 10, 11, 22, 14, 15, 3_000_000, time.UTC),
				Hostname:    "mymachine.example.com",
				Application: "evntslog",
				ProcID:      "-",
				MsgID:       "ID47",
				Data:        `[exampleSDID@32473 iut="3" eventSource= "Application" eventID="1011"][examplePriority@32473 class="high"]`,
				Content:     ``,
			},
		},
	}

	for _, c := range cases {
		m, err := parseMessage(c.in)
		expect.Any(m, err).Info(c.name).ToBe(t, &c.m)
	}
}

func concat(a, b, c []byte) []byte {
	return append(a, append(b, c...)...)
}
