package syslog

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// Message is a Syslog message. See https://www.rfc-editor.org/rfc/rfc5424
// and its forerunner https://www.rfc-editor.org/rfc/rfc3164.
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

// String produces a slightly-more verbose rendering than [Message.RFC5424].
func (m *Message) String() string {
	v := max(m.Version, 1)
	return m.format("%N<%F,%S>%V %T %H %A %P %M %D %C", v)
}

// RFC5424 calls Format("%Z%V%T%H%A%P%I%D%C") with the version set to at least 1.
// This produces a rendering according to RFC5424 regardless of the message version.
func (m *Message) RFC5424() string {
	v := max(m.Version, 1)
	return m.format("<%Z>%V %T %H %A %P %M %D %C", v)
}

// RFCFormat produces RFC5424 renderings for v1 messages and a rendering quite similar
// to RFC3164 for v0 messages, although RFC3164 is not very specific.
const RFCFormat = "<%Z>%v %T %H %A %P %M %D %C"

// Format converts the message into a string representation. The format string
// can contain a sequence of place markers:
//
//   - %A = application (and process ID if version 0)
//   - %C = message content
//   - %D = structured data
//   - %F = facility
//   - %H = hostname
//   - %M = message ID
//   - %P = process ID (if version >0)
//   - %N = source network address
//   - %S = severity
//   - %T = timestamp (varies according to version)
//   - %V = version
//   - %v = version (only if >0)
//   - %Y = timestamp year (RFC3164 version 0 messages only)
//   - %Z = priority
//
// AcceptEverything else is rendered into the result.
//
// Blank fields are omitted from the result. Leading spaces are elided before each
// blank field so that the result is compact.
func (m *Message) Format(format string) string {
	return m.format(format, m.Version)
}

func (m *Message) format(format string, version int) string {
	sw := &buffer{}
	bs := []byte(format)
	space := false
	esc := false

	for _, b := range bs {
		if b == '%' {
			if esc {
				sw.WriteByte(b)
			}
			esc = !esc
		} else if esc {
			space = m.f1(sw, b, space, version)
			esc = false
		} else {
			sw.WriteByte(b)
		}
	}
	return sw.String()
}

func (m *Message) f1(sw *buffer, b byte, space bool, version int) bool {
	switch b {
	case 'A':
		if m.Application != "" {
			if version == 0 && m.Application != "" && m.ProcID != "" {
				fmt.Fprintf(sw, "%s[%s]", m.Application, m.ProcID)
			} else {
				sw.WriteString(m.Application)
			}
			space = true
		}

	case 'C':
		if m.Content != "" {
			if strings.HasPrefix(m.Content, ":") {
				sw.TrimRightFunc(func(x byte) bool {
					return x == ' '
				})
			}
			sw.WriteString(m.Content)
			space = true
		}

	case 'D':
		if m.Data != "" {
			sw.WriteString(m.Data)
			space = true
		}

	case 'F':
		sw.WriteString(m.Facility.String())

	case 'H':
		if m.Hostname != "" {
			sw.WriteString(m.Hostname)
			space = true
		}

	case 'M':
		if m.MsgID != "" {
			sw.WriteString(m.MsgID)
			space = true
		}

	case 'N':
		if m.Source != nil {
			sw.WriteString(m.Source.String())
			space = true
		}

	case 'P':
		if m.ProcID != "" {
			if version > 0 || m.Application == "" {
				sw.WriteString(m.ProcID)
				space = true
			}
		}

	case 'S':
		sw.WriteString(m.Severity.String())

	case 'T':
		if version == 0 {
			sw.TrimRightFunc(func(x byte) bool {
				return x == ' '
			})
			sw.WriteString(m.ts().Format(rfc3164LayoutNoYear))
		} else {
			sw.WriteString(m.ts().Format(time.RFC3339))
		}
		space = true

	case 'V':
		sw.WriteString(strconv.Itoa(version))
		space = true

	case 'v':
		if version > 0 {
			sw.WriteString(strconv.Itoa(version))
			space = true
		}

	case 'Y':
		if version == 0 {
			sw.WriteString(strconv.Itoa(m.ts().Year()))
			space = true
		}

	case 'Z':
		sw.WriteString(strconv.Itoa(m.Priority()))
		space = false

	case ' ':
		if space {
			sw.WriteByte(b)
			space = false
		}

	default:
		sw.WriteByte('%')
		sw.WriteByte(b)
	}
	return space
}

func (m *Message) ts() time.Time {
	if m.Timestamp.IsZero() {
		return m.Time
	}
	return m.Timestamp
}

//-------------------------------------------------------------------------------------------------

type buffer struct {
	bs []byte
}

func (b *buffer) Len() int { return len(b.bs) }

func (b *buffer) WriteByte(c byte) error {
	b.bs = append(b.bs, c)
	return nil
}

func (b *buffer) WriteString(s string) (int, error) {
	return b.Write([]byte(s))
}

func (b *buffer) Write(bs []byte) (int, error) {
	b.bs = append(b.bs, bs...)
	return len(bs), nil
}

func (b *buffer) String() string {
	return string(b.bs)
}

func (b *buffer) Last() byte {
	if len(b.bs) == 0 {
		return 0
	}
	return b.bs[len(b.bs)-1]
}

func (b *buffer) TrimRightFunc(predicate func(byte) bool) {
	for i := len(b.bs) - 1; i >= 0; i-- {
		c := b.bs[i]
		if !predicate(c) {
			b.bs = b.bs[:i+1]
			return
		}
	}
	b.bs = b.bs[:0]
}
