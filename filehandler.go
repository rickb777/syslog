package syslog

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

// FileHandler implements [Handler] interface in the way to save messages into a
// text file. It properly handles logrotate HUP signal (closes a file and tries
// to open/create new one).
type FileHandler struct {
	filename     string
	acceptFunc   func(*Message) bool
	f            io.StringWriter
	propagateAll bool
	l            Logger
}

// NewFileHandler handles syslog messages by writing them to a file.
// The acceptFunc determines which messages are written; by default this is
// all messages.
//
// Downstream handlers see all the rejected messages. If propagateAll is true,
// downstream handles also see the accepted messages.
//
// By default, I/O errors are written to [os.Stderr] using [log.Logger].
func NewFileHandler(filename string, acceptFunc func(*Message) bool, propagateAll bool) *FileHandler {
	if acceptFunc == nil {
		acceptFunc = func(*Message) bool { return true }
	}
	h := &FileHandler{
		filename:     filename,
		acceptFunc:   acceptFunc,
		propagateAll: propagateAll,
		l:            log.New(os.Stderr, "", log.LstdFlags),
	}
	return h
}

// SetLogger changes the internal logger used to log I/O errors.
func (h *FileHandler) SetLogger(l Logger) {
	h.l = l
}

func (h *FileHandler) SigHup() {
	// SIGHUP probably from logrotate
	if h.f != nil {
		h.checkErr(h.closeFile())
		h.f = nil
		// file will re-open in next call to saveMessage
	}
}

func (h *FileHandler) Handle(m *Message) *Message {
	if m == nil {
		h.checkErr(h.closeFile())
	} else if h.acceptFunc(m) {
		h.saveMessage(m)
		if h.propagateAll {
			return m
		} else {
			return nil
		}
	}
	return m
}

func (h *FileHandler) closeFile() error {
	if h.f == nil {
		return nil
	}
	if closer, ok := h.f.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (h *FileHandler) saveMessage(m *Message) {
	var err error
	if h.f == nil {
		err = os.MkdirAll(filepath.Dir(h.filename), 0750)
		if h.checkErr(err) {
			return
		}

		h.f, err = os.OpenFile(h.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0620)
		if h.checkErr(err) {
			return
		}
	}

	_, err = h.f.WriteString(m.Format() + "\n")
	h.checkErr(err)
}

func (h *FileHandler) checkErr(err error) bool {
	if err == nil {
		return false
	}
	if h.l == nil {
		log.Print(h.filename, ": ", err)
	} else {
		h.l.Print(h.filename, ": ", err)
	}
	return true
}
