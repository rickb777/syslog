package syslog

import (
	"log"
	"net"
	"os"
	"strings"
)

type Server struct {
	conns    []net.PacketConn
	handlers []Handler
	shutdown bool
	l        Logger
}

// NewServer creates an idle server.
func NewServer() *Server {
	return &Server{l: log.New(os.Stderr, "", log.LstdFlags)}
}

// SetLogger sets logger for server errors. A running server is rather quiet and
// logs only fatal errors using [FatalLogger] interface. By default, the standard Go
// logger is used so errors are written to stderr, after which the whole
// application is halted. Using SetLogger you can change this behavior.
func (s *Server) SetLogger(l Logger) {
	s.l = l
}

// AddHandler adds h to the internal ordered list of handlers.
func (s *Server) AddHandler(h Handler) {
	s.handlers = append(s.handlers, h)
}

// Listen starts goroutine that receives syslog messages on a specified address.
// addr can be a path (for unix domain sockets) or host:port (for UDP).
func (s *Server) Listen(addr string) error {
	var c net.PacketConn
	if strings.IndexRune(addr, ':') >= 0 {
		a, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return err
		}
		c, err = net.ListenUDP("udp", a)
		if err != nil {
			return err
		}
	} else {
		a, err := net.ResolveUnixAddr("unixgram", addr)
		if err != nil {
			return err
		}
		c, err = net.ListenUnixgram("unixgram", a)
		if err != nil {
			return err
		}
	}
	s.conns = append(s.conns, c)
	go s.receiver(c)
	return nil
}

// Shutdown stops the server.
func (s *Server) Shutdown() {
	s.shutdown = true
	for _, c := range s.conns {
		err := c.Close()
		if err != nil {
			s.l.Fatalln(err)
		}
	}
	s.passToHandlers(nil)
	s.conns = nil
	s.handlers = nil
}

func isNulCrLf(r rune) bool {
	return r == 0 || r == '\r' || r == '\n'
}

func (s *Server) passToHandlers(m *Message) {
	for _, h := range s.handlers {
		m = h.Handle(m)
		if m == nil {
			break
		}
	}
}

func (s *Server) receiver(c net.PacketConn) {
	buf := make([]byte, 65536)
	for {
		n, addr, err := c.ReadFrom(buf)
		if err != nil {
			if !s.shutdown {
				s.l.Fatalln("Read error:", err)
			}
			return
		}

		bs := buf[:n]
		m, err := parseMessage(bs)
		if err != nil {
			s.l.Println(err.Error())
		} else {
			m.Source = addr
			s.passToHandlers(m)
		}
	}
}
