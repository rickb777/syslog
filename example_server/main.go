package main

import (
	"flag"
	"fmt"
	"github.com/rickb777/syslog"
	"github.com/rickb777/syslog/internal/env"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	port     int
	file     string
	format   string
	priority string
	retain   int
	debug    bool
)

func flags() {
	portDefault, e1 := env.GetInt("PORT", 514)
	retainDefault, e2 := env.GetInt("RETAIN", -1)
	fileDefault := env.GetString("FILE", "")
	formatDefault := env.GetString("FORMAT", syslog.RFCFormat)
	priorityDefault := env.GetString("PRIORITY", "")

	flag.IntVar(&port, "port", portDefault, "UDP port to listen on.")
	flag.StringVar(&file, "file", fileDefault, "File to write messages to.")
	flag.StringVar(&format, "format", formatDefault, "Format to use for messages.")
	flag.StringVar(&priority, "priority", priorityDefault,
		"Ignore messages that are not this priority, expressed as 'facility.severity'.\n"+
			"Examples: *.* | user.* | *.notice | kern,auth.notice,warning,err - where * is a wildcard.\n"+
			"The facility is one of the following keywords: auth, authpriv, cron, daemon, kern, lpr,\n"+
			"mail, mark, news, security (same as auth), syslog, user, uucp and local0 through local7.\n"+
			"The keywords mark and security should not be used in applications. The severity is one\n"+
			"of the following keywords, in ascending order: debug, info, notice, warning, warn (same\n"+
			"as warning), err, error (same as err), crit, alert, emerg, panic (same as emerg). The\n"+
			"keywords error, warn and panic are deprecated and should not be used.")
	flag.IntVar(&retain, "retain", retainDefault,
		"Truncate logfiles and rotate this number of files when opening.\n"+
			"Negative values disable rotation.")
	flag.BoolVar(&debug, "v", false, "Verbose information")

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

	if debug {
		fmt.Printf("PORT=%d\n", port)
		fmt.Printf("FILE=%s\n", file)
		fmt.Printf("FORMAT=%s\n", format)
		fmt.Printf("RETAIN=%v\n", retain)
		fmt.Printf("PRIORITY=%v\n", priority)
	}
}

// Create a server with one handler and run one listen goroutine
func main() {
	flags()

	s := syslog.NewServer(100)
	s.SetDebug(debug)
	if file != "" {
		fh := syslog.NewFileHandler(file, format)
		fh.SetRotate(retain)
		s.AddHandler(fh)
	} else {
		s.AddHandler(printHandler{})
	}

	if priority != "" {
		s.SetFilter(parsePriorityFilter(priority))
	}

	err := s.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		syslog.Logger.Fatalln(err)
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
		syslog.Logger.Fatalf("%s: invalid priority filter\n"+
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
		syslog.Logger.Fatalln(err)
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
		syslog.Logger.Fatalln(err)
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

type printHandler struct{}

func (printHandler) Handle(m *syslog.Message) *syslog.Message {
	if m != nil {
		fmt.Println(m.Format(format))
	}
	return m
}
