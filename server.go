package syslog

import (
	"net"
	"strings"
	"sync/atomic"
)

// Server handles UDP or Unix datagrams. Each received packet is parsed to obtain the syslog message.
// The message is then passed along the [Handler] chain (see [Server.AddHandler]).
//
// The handlers follow the "Chain of Responsibility" design pattern.
type Server struct {
	conns      []net.PacketConn
	queue      chan *Message
	handlers   []Handler
	acceptFunc Filter
	shutDown   atomic.Bool
}

// NewServer creates an idle server. The internal queue length can be specified and should be a
// small positive number.
func NewServer(qlen int) *Server {
	s := &Server{
		queue: make(chan *Message, qlen),
	}
	go s.passToHandlers()
	return s
}

// AddHandler adds h to the internal ordered list of handlers.
func (s *Server) AddHandler(h Handler) {
	s.handlers = append(s.handlers, h)
}

// Listen starts goroutine that receives syslog messages on a specified address.
// addr can be a path (for Unix-domain sockets) or host:port (for UDP).
// All messages are accepted.
func (s *Server) Listen(addr string) error {
	return s.ListenFilter(addr, AcceptEverything)
}

// ListenFilter starts goroutine that receives syslog messages on a specified address.
// addr can be a path (for Unix-domain sockets) or host:port (for UDP).
// Only the messages matching accept are processed.
func (s *Server) ListenFilter(addr string, accept Filter) error {
	if s.shutDown.Load() {
		panic("Server is already shut down")
	}

	var c net.PacketConn
	if strings.IndexRune(addr, ':') >= 0 {
		a, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return err
		}
		//fmt.Println("Listening on", a)
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

	go receiver(c, s.queue, accept, func() bool { return !s.shutDown.Load() })
	return nil
}

// SigHup passes a hang-up signal to all handlers. This typically is used for log rotation etc.
func (s *Server) SigHup() {
	for _, h := range s.handlers {
		if hu, ok := h.(interface{ SigHup() }); ok {
			hu.SigHup()
		}
	}
}

// Shutdown stops the server.
func (s *Server) Shutdown() {
	s.shutDown.Store(true)
	for _, c := range s.conns {
		err := c.Close()
		if err != nil {
			Logger.Fatalln(err)
		}
	}
	close(s.queue)
	s.conns = nil
	for _, h := range s.handlers {
		h.Handle(nil)
	}
	s.handlers = nil
}

func isNulCrLf(r rune) bool {
	return r == 0 || r == '\r' || r == '\n'
}

func (s *Server) passToHandlers() {
	for m := range s.queue {
		for _, h := range s.handlers {
			m = h.Handle(m)
			if m == nil {
				break
			}
		}
	}
}

func receiver(c net.PacketConn, queue chan *Message, acceptFunc Filter, running func() bool) {
	buf := make([]byte, 64*1024)
	for {
		n, addr, err := c.ReadFrom(buf)
		if err != nil {
			if running() {
				Logger.Println("Read error:", err)
			}
			return
		}

		bs := buf[:n]
		m, err := parseMessage(bs)
		if err != nil {
			Logger.Println(err.Error())
		} else if acceptFunc(m) {
			m.Source = addr
			queue <- m
		}
	}
}
