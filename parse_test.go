package syslog

import (
	"github.com/rickb777/expect"
	"testing"
	"time"
)

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

		//------------------------------ RFC 3164 ------------------------------
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
			name: "RFC3164 example 2: with date",
			in:   []byte(`<13>Feb  5 17:32:18 10.0.0.99 Use the BFG!`),
			m: Message{
				Time:        tx,
				Facility:    User,
				Severity:    Notice,
				Version:     0,
				Timestamp:   time.Date(2023, 2, 5, 17, 32, 18, 0, time.UTC),
				Hostname:    "", // no requirement to recognise the IP address as a hostname
				Application: "",
				ProcID:      "",
				MsgID:       "",
				Data:        ``,
				Content:     `10.0.0.99 Use the BFG!`, // contains the IP address
			},
		},
		{
			name: "RFC3164 example 3: malformed message",
			in: []byte(`<165>Aug 24 05:34:00 CST 1987 mymachine myproc[10]: %% It's time to make the do-nuts.  %%  Ingredients: Mix=OK, Jelly=OK ` +
				`# Devices: Mixer=OK, Jelly_Injector=OK, Frier=OK # Transport: Conveyer1=OK, Conveyer2=OK # %%`),
			m: Message{
				Time:        tx,
				Facility:    Local4,
				Severity:    Notice,
				Version:     0,
				Timestamp:   time.Date(2023, 8, 24, 5, 34, 0, 0, time.UTC),
				Hostname:    "CST", // because time zone is not expected in RFC3164
				Application: "myproc",
				ProcID:      "10",
				MsgID:       "",
				Data:        ``,
				Content: `: %% It's time to make the do-nuts.  %%  Ingredients: Mix=OK, Jelly=OK ` +
					`# Devices: Mixer=OK, Jelly_Injector=OK, Frier=OK # Transport: Conveyer1=OK, Conveyer2=OK # %%`,
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

		//------------------------------ RFC 5424 ------------------------------
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
			name: "RFC5424 like example 2: without time or structured data or BOM",
			in:   []byte(`<165>1 - 192.0.2.1 myproc 8710 - - %% It's time to make the donuts.`),
			m: Message{
				Time:        tx,
				Facility:    Local4,
				Severity:    Notice,
				Version:     1,
				Timestamp:   tx,
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
