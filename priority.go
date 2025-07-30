package syslog

import (
	"fmt"
)

type Facility byte

const (
	Kern Facility = iota
	User
	Mail
	Daemon
	Auth
	Syslog
	Lpr
	News
	Uucp
	Cron
	Authpriv
	System0
	System1
	System2
	System3
	System4
	Local0
	Local1
	Local2
	Local3
	Local4
	Local5
	Local6
	Local7
)

var facToStr = [...]string{
	"kern",
	"user",
	"mail",
	"daemon",
	"auth",
	"syslog",
	"lpr",
	"news",
	"uucp",
	"cron",
	"authpriv",
	"system0",
	"system1",
	"system2",
	"system3",
	"system4",
	"local0",
	"local1",
	"local2",
	"local3",
	"local4",
	"local5",
	"local6",
	"local7",
}

func (f Facility) String() string {
	if f > Local7 {
		return "unknown"
	}
	return facToStr[f]
}

func ParseFacility(s string) (Facility, error) {
	for i, c := range facToStr {
		if c == s {
			return Facility(i), nil
		}
	}
	return 0, fmt.Errorf("%s: unknown facility", s)
}

func ParseFacilities(words []string) ([]Facility, error) {
	var facs []Facility
	for _, w := range words {
		fac, err := ParseFacility(w)
		if err != nil {
			return nil, err
		}
		facs = append(facs, fac)
	}
	return facs, nil
}

// Severity is the message severity defined in RFC5424.
type Severity byte

const (
	Emerg Severity = iota
	Alert
	Crit
	Err
	Warning
	Notice
	Info
	Debug
)

var sevToStr = [...]string{
	"emerg",
	"alert",
	"crit",
	"err",
	"warning",
	"notice",
	"info",
	"debug",
}

func (s Severity) String() string {
	if s > Debug {
		return "unknown"
	}
	return sevToStr[s]
}

func ParseSeverity(s string) (Severity, error) {
	for i, c := range sevToStr {
		if c == s {
			return Severity(i), nil
		}
	}
	switch s {
	case "warn":
		return Warning, nil
	case "error":
		return Err, nil
	}
	return 0, fmt.Errorf("%s: unknown severity", s)
}

func ParseSeverities(words []string) ([]Severity, error) {
	var sevs []Severity
	for _, w := range words {
		sev, err := ParseSeverity(w)
		if err != nil {
			return nil, err
		}
		sevs = append(sevs, sev)
	}
	return sevs, nil
}
