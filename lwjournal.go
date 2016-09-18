/*
Package lwjournal provides a lightweight logging object which writes to the
systemd journal. It implements the lwlog.Logger interface.
*/
package lwjournal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"strconv"
	"sync"

	"github.com/lwithers/lwlog"
)

// Journal writes log messages to systemd's journal. The output is asynchronous,
// so you will need a short delay before exiting the program to ensure all
// messages are flushed correctly.
type Journal struct {
	// Debug may be set to true to enable writing of debug level messages.
	Debug bool

	// connection to the journal
	sock net.Conn

	// ready-formatted fields used to construct log entries
	priDebug, priInfo, priError []byte
	extraVars                   *bytes.Buffer

	// map of program counters onto CODE_FILE, CODE_LINE etc. fields ready
	// to be used to construct log entries
	codePos     map[uintptr][]byte
	codePosLock sync.RWMutex

	writes chan []byte
}

func writeField(b *bytes.Buffer, name string, message []byte) {
	var sz [8]byte
	binary.LittleEndian.PutUint64(sz[:], uint64(len(message)))

	b.WriteString(name)
	b.WriteByte('\n')
	b.Write(sz[:])
	b.Write(message)
	b.WriteByte('\n')
}

func buildField(name, message string) []byte {
	b := bytes.NewBuffer(make([]byte, 0, len(name)+len(message)+10))
	writeField(b, name, []byte(message))
	return b.Bytes()
}

// New returns a new journal. It may return an error if it is unable to connect
// to the running journal daemon.
func New() (*Journal, error) {
	sock, err := net.Dial("unixgram", "/run/systemd/journal/socket")
	if err != nil {
		return nil, err
	}

	j := &Journal{
		sock:      sock,
		priDebug:  buildField("PRIORITY", "7"),
		priInfo:   buildField("PRIORITY", "6"),
		priError:  buildField("PRIORITY", "3"),
		extraVars: bytes.NewBuffer(nil),
		codePos:   make(map[uintptr][]byte),
		writes:    make(chan []byte, 100),
	}
	go j.writer()
	return j, nil
}

// AddVariable adds a value into each log message that is written. This could be
// used if you have some sort of session or instance identifier.
func (j *Journal) AddVariable(name, value string) {
	writeField(j.extraVars, name, []byte(value))
}

// Debugf writes debug log messages. The message will only be written if j.Debug
// is true.
func (j *Journal) Debugf(fmt string, args ...interface{}) {
	if !j.Debug {
		return
	}
	j.entry(j.priDebug, fmt, args...)
}

// Infof writes a log message.
func (j *Journal) Infof(fmt string, args ...interface{}) {
	j.entry(j.priInfo, fmt, args...)
}

// Errorf writes an error log message.
func (j *Journal) Errorf(fmt string, args ...interface{}) {
	j.entry(j.priError, fmt, args...)
}

func (j *Journal) entry(preamble []byte, Fmt string, args ...interface{}) {
	codePos := j.getCodePos()

	msg := bytes.NewBuffer(make([]byte, 0, 80))
	fmt.Fprintf(msg, Fmt, args...)

	reqLen := len(preamble) + len(codePos) + j.extraVars.Len() +
		msg.Len() + 20

	buf := bytes.NewBuffer(make([]byte, 0, reqLen))
	buf.Write(preamble)
	buf.Write(codePos)
	buf.Write(j.extraVars.Bytes())
	writeField(buf, "MESSAGE", msg.Bytes())

	j.writes <- buf.Bytes()
}

func (j *Journal) writer() {
	for msg := range j.writes {
		// we can't do anything much about errors, but perhaps we should
		// detect the case that the journal daemon was shut down
		_, _ = j.sock.Write(msg)
	}
}

func (j *Journal) getCodePos() []byte {
	// walk back over the stack (at most 4 entries) until we hit the first
	// non-logging function
	pc := make([]uintptr, 4)
	runtime.Callers(3, pc) // skip getCodePos and Journal.Debugf/Infof/etc.
	frames := runtime.CallersFrames(pc)
	frame, moreFrames := frames.Next()
	for moreFrames && lwlog.IsLoggingFunction(frame.Func) {
		frame, moreFrames = frames.Next()
	}

	// use the PC as the key to look up the code location message in our
	// cache
	j.codePosLock.RLock()
	codePos, known := j.codePos[frame.PC]
	j.codePosLock.RUnlock()
	if known {
		return codePos
	}

	// not known — we'll build the relevant code position messages
	buf := bytes.NewBuffer(nil)
	writeField(buf, "CODE_FILE", []byte(frame.File))
	writeField(buf, "CODE_LINE",
		strconv.AppendInt(nil, int64(frame.Line), 10))
	writeField(buf, "CODE_FUNC", []byte(frame.Function))
	codePos = buf.Bytes()

	// update the cache (NB: it's possible that another goroutine could
	// also have passed into or through the above lock in the meantime,
	// but that's fine — it will just write the same result into the map).
	j.codePosLock.Lock()
	j.codePos[frame.PC] = codePos
	j.codePosLock.Unlock()

	return codePos
}
