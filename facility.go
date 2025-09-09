package syslog

import (
	"fmt"
	"strings"
)

type Facility byte

type Facilities []Facility

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
	FTP
	NTP
	LogAudit
	LogAlert
	Clock
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
	"ftp",
	"ntp",
	"logaudit",
	"logalert",
	"clock",
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

func ParseFacilities(list string) (Facilities, error) {
	words := strings.Split(list, ",")
	var fs Facilities
	for _, w := range words {
		f, err := ParseFacility(w)
		if err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}
	return fs, nil
}

func (fs Facilities) Filter() Filter {
	return func(m *Message) bool {
		for _, s := range fs {
			if s == m.Facility {
				return true
			}
		}
		return false
	}
}
