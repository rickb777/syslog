package syslog

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// FileHandler implements [Handler] interface in the way to save messages into a
// text file (or files). It properly handles logrotate HUP signal (closes a file and tries
// to open/create new one).
type FileHandler struct {
	acceptFunc   Filter
	fm           filenameMangler
	f            map[fileID]io.StringWriter
	appendMode   int
	propagateAll bool
	l            Logger
}

// NewFileHandler handles syslog messages by writing them to a file or files.
// The filename will typically be an absolute path, i.e. beginning with '/'.
// The filename can contain either/both of these placeholders:
//
//   - "%hostname%" - is replaced by the hostname in each message
//   - "%programname%" - is replaced by the program name in each message
//
// This will result in log messages being separated into multiple files or folders
// according to the hostname and program name of each log message. Should a message
// arrive with an unknown hostname or program name, "unknown" will be substituted in
// either case.
//
// By default, I/O errors are written to [os.Stderr] using [log.Logger].
func NewFileHandler(filename string, append bool) *FileHandler {
	h := &FileHandler{
		fm:         newFilenameMangler(filename),
		f:          make(map[fileID]io.StringWriter),
		appendMode: os.O_TRUNC,
		acceptFunc: func(*Message) bool { return true },
		l:          log.New(os.Stderr, "", log.LstdFlags),
	}
	if append {
		h.appendMode = os.O_APPEND
	}
	return h
}

// SetAccept changes the function used to decide whether each message should be
// processed or discarded.
// The acceptFunc determines which messages are written; if this is nil, it
// accepts all messages.
func (h *FileHandler) SetAccept(acceptFunc Filter) {
	h.acceptFunc = acceptFunc
}

// SetPropagateAll changes whether downstream handlers see all the rejected messages.
// If propagateAll is true, downstream handles also see the accepted messages. Otherwise,
// rejected messages are silently discarded (the default).
func (h *FileHandler) SetPropagateAll(propagateAll bool) {
	h.propagateAll = propagateAll
}

// SetLogger changes the internal logger used to log I/O errors.
func (h *FileHandler) SetLogger(l Logger) {
	h.l = l
}

func (h *FileHandler) SigHup() {
	// SIGHUP probably from logrotate
	if h.f != nil {
		h.checkErr(h.closeFiles())
		h.f = nil
		// file will re-open in next call to saveMessage
	}
}

func (h *FileHandler) Handle(m *Message) *Message {
	if m == nil {
		h.checkErr(h.closeFiles())
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

func (h *FileHandler) closeFiles() error {
	if h.f == nil {
		return nil
	}
	for id, f := range h.f {
		delete(h.f, id)
		if closer, ok := f.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *FileHandler) saveMessage(m *Message) {
	id := h.fm.id(m)
	f := h.f[id]

	var err error
	if f == nil {
		filename := h.fm.name(m)

		err = os.MkdirAll(filepath.Dir(filename), 0750)
		if h.checkErr(err) {
			return
		}

		f, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|h.appendMode, 0620)
		if h.checkErr(err) {
			return
		}

		h.f[id] = f
	}

	_, err = f.WriteString(m.Format("%%") + "\n")
	h.checkErr(err)
}

func (h *FileHandler) checkErr(err error) bool {
	if err == nil {
		return false
	}
	if h.l == nil {
		log.Print(err)
	} else {
		h.l.Print(err)
	}
	return true
}

//-------------------------------------------------------------------------------------------------

type filenameMangler struct {
	HasHostname    bool
	HasApplication bool
	template       string
}

const (
	hostnamePlaceholder    = "%hostname%"
	programNamePlaceholder = "%programname%"
)

func newFilenameMangler(template string) filenameMangler {
	return filenameMangler{
		HasHostname:    strings.Index(template, hostnamePlaceholder) >= 0,
		HasApplication: strings.Index(template, programNamePlaceholder) >= 0,
		template:       template,
	}
}

type fileID struct {
	Hostname    string // absent | 1*255PRINTUSASCII
	Application string // absent | 1*48PRINTUSASCII (Application, ProcID, MsgID) is the RFC3164 Tag
}

func (fm filenameMangler) id(m *Message) fileID {
	var id fileID
	if fm.HasHostname {
		id.Hostname = m.Hostname
	}
	if fm.HasApplication {
		id.Application = m.Application
	}
	return id
}

func (fm filenameMangler) name(m *Message) string {
	name := fm.template
	if fm.HasHostname {
		name = strings.ReplaceAll(name, hostnamePlaceholder, ifBlank(m.Hostname, "unknown"))
	}
	if fm.HasApplication {
		name = strings.ReplaceAll(name, programNamePlaceholder, ifBlank(m.Application, "unknown"))
	}
	return name
}

func ifBlank(s, d string) string {
	switch s {
	case "", "-":
		return d
	}
	return s
}
