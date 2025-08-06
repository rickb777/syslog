package syslog

import (
	"log"
	"os"
)

// Logger is a pluggable handler for package internal logging.
var Logger = log.New(os.Stderr, "", log.LstdFlags)
