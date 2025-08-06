package syslog

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// FileHandler implements [Handler] interface such that messages are written into a
// text file (or files). It properly handles logrotate HUP signal (closes a file and tries
// to open/create new one). Alternatively, it can be configured to perform log file rotation.
type FileHandler struct {
	acceptFunc   Filter
	fm           filenameMangler
	f            map[fileID]io.StringWriter
	format       string
	retain       int // built-in log rotation when in O_TRUNC mode
	appendMode   int
	propagateAll bool
}

// NewFileHandler handles syslog messages by writing them to a file or files.
// The filename will typically be an absolute path, i.e. beginning with '/'.
// The filename can contain either/both of these placeholders:
//
//   - "%hostname%" - is replaced by the hostname in each message
//   - "%programname%" - is replaced by the program name in each message
//   - "%facility%" - is replaced by the facility name in each message
//   - "%severity%" - is replaced by the severity name in each message
//
// This will result in log messages being separated into multiple files or folders
// according to the hostname and program name of each log message. Should a message
// arrive with an unknown hostname or program name, "unknown" will be substituted in
// either case.
//
// By default, I/O errors are written to [os.Stderr] using [syslog.Logger].
func NewFileHandler(filename, format string) *FileHandler {
	h := &FileHandler{
		fm:         newFilenameMangler(filename),
		f:          make(map[fileID]io.StringWriter),
		format:     format,
		appendMode: os.O_APPEND,
		acceptFunc: func(*Message) bool { return true },
	}
	return h
}

// SetRotate configures the FileHandler to rotate pre-existing files before new ones
// are opened. The number of pre-existing files to be retained is specified. Each
// retained file is gzipped and follows the number sequence "file.log.1.gz",
// "file.log.2.gz" etc.
//
// Log rotation can be triggered via [FileHandler.SigHup].
//
// Use a negative value to disable log rotation; thereafter, any existing logfile
// will be appended, not truncated. In this case, logfile rotation can be handled
// by 'logrotate' in Linux instead.
func (h *FileHandler) SetRotate(retain int) {
	if retain < 0 {
		h.appendMode = os.O_APPEND
		h.retain = 0
	} else {
		h.appendMode = os.O_TRUNC
		h.retain = retain
	}
}

// SetFilter changes the function used to decide whether each message should be
// processed or discarded. The acceptFunc determines which messages are written;
// if this is nil, it accepts all messages.
func (h *FileHandler) SetFilter(acceptFunc Filter) {
	h.acceptFunc = acceptFunc
}

// SetPropagateAll changes whether downstream handlers see all the rejected messages.
// If propagateAll is true, downstream handles also see the accepted messages. Otherwise,
// rejected messages are silently discarded (the default).
func (h *FileHandler) SetPropagateAll(propagateAll bool) {
	h.propagateAll = propagateAll
}

// SigHup closes any open files. If log rotation is enabled, it will occur as needed when
// log files are re-opened. If 'logrotate' is being used, rotation will happen externally.
func (h *FileHandler) SigHup() {
	if h.f != nil {
		checkErr(h.closeFiles())
		h.f = nil
		// file will re-open in next call to saveMessage
	}
}

func (h *FileHandler) Handle(m *Message) *Message {
	if m == nil {
		checkErr(h.closeFiles())
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalln(err)
	}
	return err == nil
}

func (h *FileHandler) saveMessage(m *Message) {
	id := h.fm.id(m)
	f := h.f[id]

	var err error
	if f == nil {
		filename := h.fm.name(m)

		f, err = h.openFile(filename)
		if checkErr(err) {
			return
		}

		h.f[id] = f
	}

	checkErr2(f.WriteString(m.Format(h.format)))
	checkErr2(f.WriteString("\n"))
}

const tmp = ".tmp"

func (h *FileHandler) openFile(filename string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(filename), 0750); err != nil {
		return nil, err
	}

	if h.appendMode == os.O_TRUNC {
		if fileExists(filename) {
			// rename so we can use a goroutine
			if !checkErr(os.Rename(filename, filename+tmp), "mv", filename, filename+tmp) {
				go h.logRotate(filename)
			}
		}
	}

	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|h.appendMode, 0620)
}

func (h *FileHandler) logRotate(filename string) {
	var old, older string
	older = fmt.Sprintf("%s.%d.gz", filename, h.retain)
	if fileExists(older) {
		checkErr(os.Remove(older), "rm", older)
	}

	for i := h.retain - 1; i > 0; i-- {
		old = fmt.Sprintf("%s.%d.gz", filename, i)
		if fileExists(old) {
			checkErr(os.Rename(old, older), "mv", old, older)
		}
		older = old
	}

	tmpFile := filename + tmp
	in, err := os.Open(tmpFile)
	if err != nil {
		Logger.Println("open", tmpFile, err)
		return
	}

	old = fmt.Sprintf("%s.1.gz", filename)
	o, err := os.OpenFile(old, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0620)
	if err != nil {
		Logger.Println("create", old, err)
		return
	}

	gz, err := gzip.NewWriterLevel(o, 5) // level 5 is both quite good and quite fast
	if err != nil {
		Logger.Println("gzip", old, err)
		return
	}

	if checkErr2(io.Copy(gz, in)) {
		return
	}
	if checkErr(gz.Close(), "gz-close", old) {
		return
	}
	if checkErr(o.Close(), "close", old) {
		return
	}
	checkErr(os.Remove(tmpFile), "rm", tmpFile)
}

func checkErr2(_ any, err error) bool {
	return checkErr(err)
}

func checkErr(err error, info ...string) bool {
	if err != nil {
		description := strings.Join(info, " ")
		Logger.Println(description, err)
		return true
	}
	return false
}

//-------------------------------------------------------------------------------------------------

type filenameMangler struct {
	HasHostname    bool
	HasApplication bool
	HasFacility    bool
	HasSeverity    bool
	template       string
}

const (
	hostnamePlaceholder    = "%hostname%"
	programNamePlaceholder = "%programname%"
	facilityPlaceholder    = "%facility%"
	severityPlaceholder    = "%severity%"
)

func newFilenameMangler(template string) filenameMangler {
	return filenameMangler{
		HasHostname:    strings.Index(template, hostnamePlaceholder) >= 0,
		HasApplication: strings.Index(template, programNamePlaceholder) >= 0,
		HasFacility:    strings.Index(template, facilityPlaceholder) >= 0,
		HasSeverity:    strings.Index(template, severityPlaceholder) >= 0,
		template:       template,
	}
}

type fileID struct {
	Hostname    string
	Application string
	Facility    string
	Severity    string
}

func (fm filenameMangler) id(m *Message) fileID {
	var id fileID
	if fm.HasHostname {
		id.Hostname = m.Hostname
	}
	if fm.HasApplication {
		id.Application = m.Application
	}
	if fm.HasFacility {
		id.Facility = m.Facility.String()
	}
	if fm.HasSeverity {
		id.Severity = m.Severity.String()
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
	if fm.HasFacility {
		name = strings.ReplaceAll(name, facilityPlaceholder, ifBlank(m.Facility.String(), "unknown"))
	}
	if fm.HasSeverity {
		name = strings.ReplaceAll(name, severityPlaceholder, ifBlank(m.Severity.String(), "unknown"))
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
