package main

import (
	"flag"
	"fmt"
	"github.com/rickb777/syslog"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type handler struct {
	// To simplify implementation of our handler we embed helper
	// syslog.BaseHandler struct.
	*syslog.BaseHandler
}

// Simple filter for named/bind messages that can be used with BaseHandler
func filter(m *syslog.Message) bool {
	return m.Data == "named" || m.Data == "bind"
}

func newHandler() *handler {
	h := handler{syslog.NewBaseHandler(5, filter, false)}
	go h.mainLoop() // BaseHandler needs some goroutine that reads from its queue
	return &h
}

// mainLoop reads from BaseHandler queue using h.Get and logs messages to stdout
func (h *handler) mainLoop() {
	for {
		m := h.Get()
		if m == nil {
			break
		}
		fmt.Println(m)
	}
	fmt.Println("Exit handler")
	h.End()
}

var (
	port = flag.Int("port", 514, "port to listen on")
	file = flag.String("file", "", "file to write messages to")
)

// Create a server with one handler and run one listen goroutine
func main() {
	flag.Parse()

	s := syslog.NewServer()
	if *file != "" {
		s.AddHandler(syslog.NewFileHandler(*file, 5, nil, false))
	} else {
		s.AddHandler(newHandler())
	}
	err := s.Listen(fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatal(err)
	}

	// Wait for terminating signal
	sc := make(chan os.Signal, 2)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT)
	<-sc

	// Shutdown the server
	fmt.Println("Shutdown the server...")
	s.Shutdown()
	fmt.Println("Server is down")
}
