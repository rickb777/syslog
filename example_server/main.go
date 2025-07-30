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

type printHandler struct{}

func (printHandler) Handle(m *syslog.Message) *syslog.Message {
	if m != nil {
		fmt.Println(m)
	}
	return m
}

// Simple filter for 'user' messages
func filter(m *syslog.Message) bool {
	return m.Facility == syslog.User
}

var (
	port  = flag.Int("port", 514, "port to listen on")
	file  = flag.String("file", "", "file to write messages to")
	debug = flag.Bool("v", false, "verbose information")
)

// Create a server with one handler and run one listen goroutine
func main() {
	flag.Parse()

	s := syslog.NewServer(100)
	s.SetDebug(*debug)
	if *file != "" {
		s.AddHandler(syslog.NewFileHandler(*file, nil, false))
	} else {
		s.AddHandler(printHandler{})
	}

	err := s.Listen(fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatal(err)
	}

	// Wait for terminating signal
	sc := make(chan os.Signal, 2)
	signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	for v := range sc {
		switch v {
		case syscall.SIGHUP:
			s.SigHup()

		default:
			s.Shutdown()
			os.Exit(0)
		}
	}
}
