package syslog

import (
	"bytes"
	"fmt"
	"github.com/rickb777/iso8601/v2"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Message is a Syslog message.
type Message struct {
	Time   time.Time // locally determined
	Source net.Addr  // from network socket
	//--- Header ---
	Facility
	Severity
	Version     int       // Syslog message version
	Timestamp   time.Time // absent | RFC3339
	Hostname    string    // absent | 1*255PRINTUSASCII
	Application string    // absent | 1*48PRINTUSASCII (Application, ProcID, MsgID) is the RFC3164 Tag
	ProcID      string    // absent | 1*128PRINTUSASCII
	MsgID       string    // absent | 1*32PRINTUSASCII
	Data        string    // structured data as defined in RFC 5424 like `[id item="value"]
	Content     string    // message content
}

func (m *Message) Priority() int {
	return int(m.Facility)<<3 | int(m.Severity)
}

// NetSrc only network part of Source as string (IP for UDP or Name for UDS)
func (m *Message) NetSrc() string {
	switch a := m.Source.(type) {
	case *net.UDPAddr:
		return a.IP.String()
	case *net.UnixAddr:
		return a.Name
	case *net.TCPAddr:
		return a.IP.String()
	}
	// Unknown type
	return m.Source.String()
}

func (m *Message) format(pri string, version int) []string {
	var h []string
	t := m.Timestamp
	if t.IsZero() {
		t = m.Time
	}

	if version == 0 {
		h = append(h, fmt.Sprintf("%s%s", pri, t.Format(rfc3164LayoutNoYear)))
	} else {
		h = append(h, fmt.Sprintf("%s%d %s", pri, version, t.Format(time.RFC3339)))
	}

	if m.Hostname != "" {
		h = append(h, m.Hostname)
	}

	if version == 0 && m.Application != "" && m.ProcID != "" {
		h = append(h, fmt.Sprintf("%s[%s]", m.Application, m.ProcID))
	} else {
		if m.Application != "" {
			h = append(h, m.Application)
		}
		if m.ProcID != "" {
			h = append(h, m.ProcID)
		}
	}

	if m.MsgID != "" {
		h = append(h, m.MsgID)
	}

	if m.Data != "" {
		h = append(h, m.Data)
	}

	return h
}

func (m *Message) Format() string {
	var h []string
	pri := fmt.Sprintf("<%d>", m.Priority())
	h = append(h, m.format(pri, m.Version)...)
	return strings.Join(h, " ") + leadingSpaceIfNotColon(m.Content)
}

func (m *Message) String() string {
	var h []string
	if m.Source != nil {
		h = append(h, m.Source.String())
	}
	pri := fmt.Sprintf("<%s,%s>", m.Facility, m.Severity)
	h = append(h, m.format(pri, 1)...)
	return strings.Join(h, " ") + leadingSpaceIfNotColon(m.Content)
}

func leadingSpaceIfNotColon(s string) string {
	if strings.HasPrefix(s, ":") {
		return s
	}
	return " " + s
}

//-------------------------------------------------------------------------------------------------

func parseMessage(pkt []byte) (*Message, error) {
	var n int
	ts := now()
	m := Message{
		Time:      ts,
		Timestamp: ts,
	}

	bs := bytes.TrimRightFunc(pkt, isNulCrLf)

	bom := findBOM(bs)
	if bom >= 0 { // Byte Order Mark was found
		m.Content = string(bs[bom+3:])
		bs = bs[:bom]
	}

	s := string(bs)

	//---------- Parse priority (if it exists)
	prio := 13 // default priority

	// we treat PRI as optional although RFC3164 and RFC5424 require it to be present
	if s[0] == '<' {
		n = 1 + strings.IndexByte(s[1:], '>')
		if n > 1 && n < 5 {
			p, err := strconv.Atoi(s[1:n])
			if err != nil {
				return nil, fmt.Errorf("%s: message has invalid priority (%s)",
					s[1:n], cropString(s, 50))
			}
			prio = p
			s = s[n+1:]
		}
	}

	m.Severity = Severity(prio & 0x07)
	m.Facility = Facility(prio >> 3)

	if strings.HasPrefix(s, "1 ") {
		m.Version = 1
		s = s[2:]
		return parseRFC5424Message(&m, s, bom >= 0)
	}

	return parseRFC3164Message(&m, s)
}

//-------------------------------------------------------------------------------------------------

const (
	rfc3164LayoutNoYear   = "Jan _2 15:04:05"
	rfc3164LayoutWithYear = "2006 Jan _2 15:04:05"
)

func parseRFC3164Message(m *Message, s string) (*Message, error) {
	s = strings.TrimLeftFunc(s, unicode.IsSpace)

	if len(s) > 15 && s[15] == ' ' {
		// date without year
		ts, err := time.Parse(rfc3164LayoutNoYear, s[:15])
		if err == nil {
			if ts.Year() == 0 {
				// There will be unavoidable race errors at the very end of
				// December 31st / start of January 1st.
				ts = ts.AddDate(m.Time.Year(), 0, 0)
			}
			m.Timestamp = ts
			s = s[15:]
		}
	} else if len(s) > 20 && s[20] == ' ' {
		// date with year
		ts, err := time.Parse(rfc3164LayoutWithYear, s[:20])
		if err == nil {
			m.Timestamp = ts
			s = s[20:]
		}
	}

	s = strings.TrimLeftFunc(s, unicode.IsSpace)

	if strings.HasPrefix(s, "TZ") {
		sp := nextSpace(s)
		if 0 < sp && sp <= len(s) {
			tz, err := strconv.Atoi(s[2:sp])
			if err == nil && -12 <= tz && tz <= 12 {
				m.Timestamp = m.Timestamp.In(time.FixedZone(s[:sp], tz*3600))
			}
			s = s[sp+1:]
		}
	}

	colon := indexRune(s, ':')
	if colon < 0 {
		m.Content = s
		return m, nil
	}

	m.Content = s[colon:]
	s = s[:colon]

	words := strings.Split(s, " ")
	if len(words) > 0 {
		m.Hostname = words[0]
	}

	if len(words) > 1 {
		last := words[len(words)-1]
		if strings.HasSuffix(last, "]") {
			l := strings.IndexByte(last, '[')
			if l > 0 {
				m.Application = last[:l]
				m.ProcID = last[l+1 : len(last)-1]
			} else {
				m.Application = last
			}
		} else {
			m.Application = last
		}
	}
	return m, nil
}

//-------------------------------------------------------------------------------------------------

func parseRFC5424Message(m *Message, s string, hasBOM bool) (*Message, error) {
	if strings.HasPrefix(s, "- ") {
		s = s[2:] // no time field
	} else {
		sp := strings.IndexByte(s, ' ')
		if sp >= 0 {
			ts, err := iso8601.ParseString(s[:sp])
			if err == nil {
				m.Timestamp = ts
				s = s[sp+1:]
			}
		}
	}

	s = nextField(s, &m.Hostname)
	s = nextField(s, &m.Application)
	s = nextField(s, &m.ProcID)
	s = nextField(s, &m.MsgID)

	if strings.HasPrefix(s, "- ") {
		m.Data = "-"
		s = s[2:]
	} else if strings.HasPrefix(s, "[") {
		r := indexRune(s, ']')
		for r >= 0 {
			if r == len(s)-1 {
				m.Data = s
				s = s[r+1:]
				break
			} else if 0 < r && r < len(s) {
				l := indexRune(s[r:], '[')
				if l > 0 {
					r2 := indexRune(s[r+l:], ']')
					if r2 >= 0 {
						r += r2 + l
					}
				} else {
					r++
					m.Data = s[:r]
					s = s[r+1:]
					r = indexRune(s, ']')
				}
			}
		}
	}

	if !hasBOM { // no Byte Order Mark
		if strings.HasPrefix(s, " ") {
			s = s[1:]
		}
		m.Content = s
	}

	return m, nil
}

//-------------------------------------------------------------------------------------------------

func nextField(s string, field *string) string {
	if strings.HasPrefix(s, "- ") { // NILVALUE
		*field = "-"
		s = s[2:]
	} else {
		sp := nextSpace(s)
		if 0 < sp && sp <= len(s) && s[0] != '[' {
			*field = s[:sp]
			s = s[sp+1:]
		}
	}
	return s
}

func nextSpace(s string) int {
	for i, r := range s {
		if r == ' ' {
			return i
		} else if r < 32 || r > 126 {
			return 0 // anything outside PRINTASCII %d33-126
		}
	}

	return len(s) // not found
}

// indexRune finds the next c in s, skipping any characters escaped with '\'.
func indexRune(s string, c rune) int {
	esc := false

	for i, r := range s {
		switch r {
		case '\\':
			esc = !esc
		case c:
			if !esc {
				return i
			}
			esc = false
		default:
			esc = false
		}
	}

	return -1 // not found
}

func cropString(s string, crop int) string {
	if len(s) > crop {
		return s[:crop] + "..."
	}
	return s
}

// findBOM finds the byte order mark, if present.
func findBOM(bs []byte) int {
	// We don't care about the possibility of 0xEF occurring more than once because the
	// header part is always only 7-bit ASCII, so any subsequent 0xEF will be after the BOM.
	bom := bytes.IndexByte(bs, 0xEF)
	if 0 <= bom && bs[bom+1] == 0xBB && bs[bom+2] == 0xBF {
		return bom
	}
	return -1
}

var now = func() time.Time { return time.Now() }
