package syslog

import (
	"github.com/rickb777/expect"
	"os"
	"testing"
)

func TestFilenameMangler(t *testing.T) {
	fm := newFilenameMangler("/var/log/%hostname%/%facility%/%programname%-%severity%.log")
	m := &Message{
		Hostname:    "myhost",
		Application: "myapp",
		Facility:    Daemon,
		Severity:    Warning,
	}
	expect.Any(fm.id(m)).ToBe(t, fileID{
		Hostname:    "myhost",
		Application: "myapp",
		Facility:    "daemon",
		Severity:    "warning",
	})
	expect.String(fm.name(m)).ToBe(t, "/var/log/myhost/daemon/myapp-warning.log")
	expect.String(fm.name(&Message{})).ToBe(t, "/var/log/unknown/kern/unknown-emerg.log")
}

func TestLogrotate(t *testing.T) {
	const filename = "./temp.log"
	expect.Error(os.WriteFile(filename+tmp, []byte("this is file 1\n"), 0644)).ToBeNil(t)
	defer os.Remove(filename)

	h := NewFileHandler(filename, RFCFormat)
	h.SetRotate(2)
	h.logRotate(filename)
	defer os.Remove(filename + ".1.gz")

	expect.Error(os.WriteFile(filename+tmp, []byte("this is file 2\n"), 0644)).ToBeNil(t)

	h.logRotate(filename)
	defer os.Remove(filename + ".2.gz")

	expect.Error(os.WriteFile(filename+tmp, []byte("this is file 3\n"), 0644)).ToBeNil(t)
	h.logRotate(filename)

	expect.Bool(fileExists(filename)).ToBe(t, false)
	expect.Bool(fileExists(filename+".1.gz")).ToBe(t, true)
	expect.Bool(fileExists(filename+".2.gz")).ToBe(t, true)
}
