package main

import (
	"flag"
	"fmt"
	"github.com/rickb777/syslog"
	"github.com/rickb777/syslog/internal/env"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type printHandler struct{}

func (printHandler) Handle(m *syslog.Message) *syslog.Message {
	if m != nil {
		fmt.Println(m)
	}
	return m
}

var (
	port     int
	file     string
	priority string
	truncate bool
	debug    bool
)

func flags() {
	portDefault, e1 := env.GetInt("PORT", 514)
	truncDefault, e2 := env.GetBool("TRUNCATE", false)
	fileDefault := env.GetString("FILE", "")

	flag.IntVar(&port, "port", portDefault, "port to listen on")
	flag.StringVar(&file, "file", fileDefault, "file to write messages to")
	flag.StringVar(&priority, "priority", env.GetString("PRIORITY", ""),
		"ignore messages that are not this priority\n"+
			"Format: *.* | user.* | *.notice | kern,auth.notice,warning,err - where * is a wildcard")
	flag.BoolVar(&truncate, "truncate", truncDefault, "truncate when opening logfiles instead of appending")
	flag.BoolVar(&truncate, "t", truncDefault, "alias for -truncate")
	flag.BoolVar(&debug, "v", false, "verbose information")

	flag.Parse()

	if e1 != nil {
		fmt.Fprintln(os.Stderr, "PORT", e1)
		flag.Usage()
		os.Exit(1)
	}
	if e2 != nil {
		fmt.Fprintln(os.Stderr, "TRUNCATE", e2)
		flag.Usage()
		os.Exit(1)
	}
}

// Create a server with one handler and run one listen goroutine
func main() {
	flags()

	s := syslog.NewServer(100)
	s.SetDebug(debug)
	if file != "" {
		s.AddHandler(syslog.NewFileHandler(file, !truncate))
	} else {
		s.AddHandler(printHandler{})
	}

	if priority != "" {
		s.SetFilter(parsePriorityFilter(priority))
	}

	err := s.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalln(err)
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

func parsePriorityFilter(pri string) syslog.Filter {
	parts := strings.Split(pri, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		log.Fatalf("%s: invalid priority filter\n"+
			"Must be like \"*.*\" | \"user.info\" | \"kern,auth.*\" etc.\n", pri)
	}

	if parts[0] == "*" && parts[1] == "*" {
		return func(m *syslog.Message) bool { return true }
	} else if parts[0] != "*" && parts[1] != "*" {
		return syslog.All(
			parseFacilityFilter(parts[0]),
			parseSeverityFilter(parts[1]),
		)
	} else if parts[0] == "*" && parts[1] != "*" {
		return parseSeverityFilter(parts[1])
	} else {
		return parseFacilityFilter(parts[0])
	}
}

func parseFacilityFilter(s string) syslog.Filter {
	words := strings.Split(s, ",")
	facs, err := syslog.ParseFacilities(words)
	if err != nil {
		log.Fatalln(err)
	}

	return func(m *syslog.Message) bool {
		for _, f := range facs {
			if f == m.Facility {
				return true
			}
		}
		return false
	}
}

func parseSeverityFilter(s string) syslog.Filter {
	words := strings.Split(s, ",")
	sevs, err := syslog.ParseSeverities(words)
	if err != nil {
		log.Fatalln(err)
	}

	return func(m *syslog.Message) bool {
		for _, s := range sevs {
			if s == m.Severity {
				return true
			}
		}
		return false
	}
}
