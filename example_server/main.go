package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rickb777/syslog"
	"github.com/rickb777/syslog/internal/env"
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
	flag.StringVar(&file, "file", fileDefault, "File to write messages to. (default stdout)")
	flag.StringVar(&format, "format", formatDefault, "Format to use for messages.")
	flag.StringVar(&priority, "priority", priorityDefault,
		"Ignore messages that are not this priority, expressed as 'facility.severity'.\n"+
			"Facility and severity are both lists, where * is a wildcard.\n"+
			"Examples: *.* | user.* | *.notice | kern,auth.notice,warning,err.\n\n"+
			"The facility is one of the following keywords:\n"+
			"auth, authpriv, cron, daemon, kern, lpr, mail, news, syslog, user, uucp\n"+
			"and local0 through local7.\n\n"+
			"The severity is one of the following keywords, in ascending order:\n"+
			"debug, info, notice, warning, err, crit, alert, emerg.\n"+
			"The keywords error (alias for err), warn (alias for warning) and panic\n"+
			"(alias for emerg) are supported but deprecated.")
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
	if debug {
		s.AddHandler(syslog.DebugHandler{})
	}
	if file != "" {
		fh := syslog.NewFileHandler(file, format)
		fh.SetRotate(retain)
		s.AddHandler(fh)
	} else {
		s.AddHandler(syslog.PrintHandler(format))
	}

	var err error
	filter := syslog.AcceptEverything
	if priority != "" {
		filter, err = syslog.ParsePriorityFilter(priority)
		if err != nil {
			syslog.Logger.Fatalln(err)
		}
	}

	err = s.ListenFilter(fmt.Sprintf(":%d", port), filter)
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
