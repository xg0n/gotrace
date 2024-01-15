// Custom logger for gotrace
package log

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

// counter to mark each call so that entry and exit points can be correlated
var (
	counter    uint64
	L          *log.Logger
	setupOnce  sync.Once
	formatSize int
)

// Setup our logger
// return  a value so this van be executed in a toplevel var statement
func Setup(output, prefix string, size int) int {
	setupOnce.Do(func() {
		setup(output, prefix, size)
	})
	return 0
}

func setup(output, prefix string, size int) {
	var out io.Writer
	switch output {
	case "stdout":
		out = os.Stdout
	default:
		out = os.Stderr
	}

	L = log.New(out, prefix, log.Lmicroseconds)
	formatSize = size
}

// Make things a little more readable. Format as strings with %q when we can,
// strip down empty slices, and don't print the internals from buffers.
func formatter(i interface{}, size int) (s string) {
	// don't show the internal state of buffers
	switch v := i.(type) {
	case *bufio.Reader:
		s = "&bufio.Reader{}"
	case *bufio.Writer:
		s = "&bufio.Writer{}"
	case *bytes.Buffer:
		s = fmt.Sprintf("&bytes.Buffer{%q}", v.String())
	case *bytes.Reader:
		buf := make([]byte, min(size, v.Len()))
		_, err := io.ReadFull(v, buf)
		if err != nil {
			s = "&bytes.Reader{unknown}"
		} else {
			s = string(buf)
		}
	case *strings.Reader:
		buf := make([]byte, min(size, v.Len()))
		_, err := io.ReadFull(v, buf)
		if err != nil {
			s = "&strings.Reader{unknown}"
		} else {
			s = string(buf)
		}
	case []byte:
		// bytes slices are often empty, so trim them down
		b := bytes.TrimLeft(v, "\x00")
		if len(b) == 0 {
			s = "[]byte{0...}"
		} else if utf8.Valid(v) {
			s = fmt.Sprintf("[]byte{%q}", v)
		} else {
			s = fmt.Sprintf("%#v", v)
		}
	case string:
		s = fmt.Sprintf("%q", v)
	}

	if s == "" {
		s = fmt.Sprintf("%#v", i)
	}

	if len(s) > size {
		last := s[len(s)-1]
		s = s[:size] + "..." + string(last)
	}

	return s
}

// Format N number of arguments for logging, and limit the length of each formatted arg.
func Format(args ...interface{}) string {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = formatter(arg, formatSize)
	}
	return strings.Join(parts, ", ")
}

func ID() uint64 {
	return atomic.AddUint64(&counter, 1)
}

func Now() time.Time {
	return time.Now()
}

func Since(t time.Time) time.Duration {
	return time.Since(t)
}
