// Custom logger for gotrace
package log

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/xg0n/routineid"
)

// counter to mark each call so that entry and exit points can be correlated
var (
	counter    uint64
	L          *log.Logger
	setupOnce  sync.Once
	formatSize int
	Enable     atomic.Bool
	sigCh      chan os.Signal
	pid        int
	ppid       int
)

// Setup our logger
// return  a value so this van be executed in a toplevel var statement
func Setup(output, prefix string, size int, enableByDefault bool, toggleSignalNum int) int {
	setupOnce.Do(func() {
		setupLogFile(output, prefix, size)
		setupSignal(enableByDefault, toggleSignalNum)
		setupGlobalVar()
	})
	return 0
}

func setupLogFile(output, prefix string, size int) {
	var out io.Writer
	switch output {
	case "stdout":
		out = os.Stdout
	case "stderr":
		out = os.Stderr
	default:
		file, err := os.Create(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open log file \"%s\", err: %s\n", output, err)
			break
		}
		out = file
	}
	L = log.New(out, prefix, log.Lmicroseconds)
	formatSize = size
}

func setupSignal(enableByDefault bool, toggleSignalNum int) {
	Enable.Store(enableByDefault)
	sigCh = make(chan os.Signal)
	toggleSignal := syscall.Signal(toggleSignalNum)
	signal.Notify(sigCh, toggleSignal)
	go signalHandler(toggleSignal)
}

func setupGlobalVar() {
	pid = os.Getpid()
	ppid = os.Getppid()
}

func signalHandler(signal syscall.Signal) {
	for {
		select {
		case sig := <-sigCh:
			if sig == signal {
				curentVal := Enable.Load()
				L.Printf("Toggle log write from %t to %t\n", curentVal, !curentVal)
				Enable.Store(!curentVal)
			}
		}
	}
}

func min(a int, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}

/* Return process id and parent process id */
func GetPidPpid() (int, int) {
	return pid, ppid
}

/* Return current and parent goroutine id in string format */
func GetRoutineIds() (currentId uint64, parrentId uint64) {
	return routineid.GetRoutineIds()
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

func Now() time.Time {
	return time.Now()
}

func Since(t time.Time) time.Duration {
	return time.Since(t)
}
