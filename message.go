package syslog

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Message is a Syslog message.
type Message struct {
	Time   time.Time
	Source net.Addr
	Facility
	Severity
	Version     int       // Syslog message version
	Timestamp   time.Time // optional
	Hostname    string    // optional
	Application string    // optional
	PID         int       //
	Tag         string    // message tag as defined in RFC 3164
	Content     string    // message content as defined in RFC 3164
	Tag1        string    // alternate message tag (white rune as separator)
	Content1    string    // alternate message content (white rune as separator)
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

func (m *Message) String() string {
	const (
		timeLayout      = "2006-01-02 15:04:05"
		timestampLayout = "01-02 15:04:05"
	)
	var h []string
	h = append(h, m.Time.Format(timeLayout))

	if !m.Timestamp.IsZero() {
		h = append(h, m.Timestamp.Format(timestampLayout))
	}
	if m.Source != nil {
		h = append(h, m.Source.String())
	}
	h = append(h, fmt.Sprintf("<%s,%s>", m.Facility, m.Severity))
	if m.Hostname != "" {
		h = append(h, m.Hostname)
	}
	if m.Application != "" {
		h = append(h, m.Application)
	}
	if m.PID >= 0 {
		h = append(h, strconv.Itoa(m.PID))
	}
	h = append(h, m.Tag)
	h = append(h, m.Content)
	return strings.Join(h, " ")
}

func parseMessage(pkt []byte, addr net.Addr, isNotAlnum func(r rune) bool) *Message {
	var n int
	m := &Message{
		Source: addr,
		Time:   now(),
	}

	// Parse priority (if it exists)
	prio := 13 // default priority
	hasPrio := false

	if pkt[0] == '<' {
		n = 1 + bytes.IndexByte(pkt[1:], '>')
		if n > 1 && n < 5 {
			p, err := strconv.Atoi(string(pkt[1:n]))
			if err == nil && p >= 0 {
				hasPrio = true
				prio = p
				pkt = pkt[n+1:]
			}
		}
	}

	m.Severity = Severity(prio & 0x07)
	m.Facility = Facility(prio >> 3)

	hostnameOffset := 0
	ts := time.Now()

	// Parse header (if exists)
	if hasPrio && len(pkt) >= 26 && pkt[25] == ' ' && pkt[15] != ' ' {
		// OK, it looks like we're dealing with a RFC 5424-style packet
		ts, err := time.Parse(time.RFC3339, string(pkt[:25]))
		if err == nil && !ts.IsZero() {
			// Time parsed correctly.  This is most certainly a RFC 5424-style packet.
			// Hostname starts at pkt[26]
			hostnameOffset = 26
		}
	} else if hasPrio && len(pkt) >= 16 && pkt[15] == ' ' {
		// Looks like we're dealing with a RFC 3164-style packet
		layout := "Jan _2 15:04:05"
		ts, err := time.Parse(layout, string(pkt[:15]))
		if err == nil && !ts.IsZero() {
			// Time parsed correctly.   This is most certainly a RFC 3164-style packet.
			hostnameOffset = 16
		}
	}

	if hostnameOffset == 0 {
		log.Printf("Packet did not parse correctly:\n%v\n", string(pkt[:]))
	} else {
		n = hostnameOffset + bytes.IndexByte(pkt[hostnameOffset:], ' ')
		if n != hostnameOffset-1 {
			m.Timestamp = ts
			m.Hostname = string(pkt[hostnameOffset:n])
			pkt = pkt[n+1:]
		}
	}
	_ = hostnameOffset

	// Parse msg part
	msg := string(bytes.TrimRightFunc(pkt, isNulCrLf))
	n = strings.IndexFunc(msg, isNotAlnum)
	if n != -1 {
		m.Tag = msg[:n]
		m.Content = msg[n:]
	} else {
		m.Content = msg
	}
	msg = strings.TrimFunc(msg, unicode.IsSpace)
	n = strings.IndexFunc(msg, unicode.IsSpace)
	if n != -1 {
		m.Tag1 = msg[:n]
		m.Content1 = strings.TrimLeftFunc(msg[n+1:], unicode.IsSpace)
	} else {
		m.Content1 = msg
	}
	return m
}

var now = func() time.Time { return time.Now() }
