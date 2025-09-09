package syslog

import (
	"errors"
	"fmt"
	"strings"
)

// ParsePriorityFilter parses a priority such as "user.info,warn,error"
func ParsePriorityFilter(pri string) (Filter, error) {
	parts := strings.Split(pri, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("%s: invalid priority filter\n"+
			"Must be like \"*.*\" | \"user.info\" | \"kern,auth.*\" etc.\n", pri)
	}

	if parts[0] == "*" && parts[1] == "*" {
		return AcceptEverything, nil
	} else if parts[0] != "*" && parts[1] != "*" {
		fs, err1 := ParseFacilities(parts[0])
		ss, err2 := ParseSeverities(parts[1])
		if err1 != nil || err2 != nil {
			return nil, errors.Join(err1, err2)
		}
		return All(fs.Filter(), ss.Filter()), nil
	} else if parts[0] == "*" && parts[1] != "*" {
		ss, err := ParseSeverities(parts[1])
		return ss.Filter(), err
	} else {
		fs, err := ParseFacilities(parts[0])
		return fs.Filter(), err
	}
}

//-------------------------------------------------------------------------------------------------

// Severity is the message severity defined in RFC5424.
type Severity byte

type Severities []Severity

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

func ParseSeverities(list string) (Severities, error) {
	words := strings.Split(list, ",")
	var ss Severities
	for _, w := range words {
		s, err := ParseSeverity(w)
		if err != nil {
			return nil, err
		}
		ss = append(ss, s)
	}
	return ss, nil
}

func (ss Severities) Filter() Filter {
	return func(m *Message) bool {
		for _, s := range ss {
			if s == m.Severity {
				return true
			}
		}
		return false
	}
}
