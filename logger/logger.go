package logger

import (
	"fmt"
	"io"
	"os"
	"sync"

	stdlog "log"

	kitlog "github.com/go-kit/kit/log"
)

type LoggerBase struct {
	lock   sync.Mutex
	writer io.Writer
}

func newLoggerBase(ws ...io.Writer) *LoggerBase {
	w := io.Writer(os.Stderr)
	if len(ws) != 0 {
		w = io.MultiWriter(ws...)
	}
	return &LoggerBase{writer: w}
}

// SwapWriters allows the underlying writer used by all Loggers that
// inherit from this LoggerBase to be changed. If len(ws) == 0 then
// os.Stderr is installed as the only writer. Otherwise, exactly ws
// are installed as writers (wrapped up with an io.MultiWriter).
func (lb *LoggerBase) SwapWriters(ws ...io.Writer) {
	w := io.Writer(os.Stderr)
	if len(ws) != 0 {
		w = io.MultiWriter(ws...)
	}
	lb.lock.Lock()
	lb.writer = w
	lb.lock.Unlock()
}

func (lb *LoggerBase) Write(p []byte) (n int, err error) {
	lb.lock.Lock()
	n, err = lb.writer.Write(p)
	lb.lock.Unlock()
	return n, err
}

// Log vs via the underlying Writer, bypassing the Logger key-value
// contexts. Useful when you receive preformatted key-value pairs
// which you need to route through this LoggerBase.
func (lb *LoggerBase) LogRaw(vs ...interface{}) {
	if _, err := fmt.Fprintln(lb, vs...); err != nil {
		panic(err)
	}
}

type Logger struct {
	// Logger values include the contextual key-val pairs (i.e. via
	// With()) (hence has state)
	kit kitlog.Logger

	// All loggers inherit from the same LoggerBase
	*LoggerBase
}

// Log keyvals through the logger. Keyvals is a slice of key-value
// pairs, and are logged as key=value (logfmt format). The logger logs
// the context and the keyvals.
func (l *Logger) Log(keyvals ...interface{}) {
	if err := l.kit.Log(keyvals...); err != nil {
		panic(err)
	}
}

// With returns a new child logger with context set to the existing
// logger's context with keyvals appended. As with Log, keyvals is a
// slice of key-value pairs. The underlying logger (i.e. writers etc)
// is shared by the parent and child logger.
func (l *Logger) With(keyvals ...interface{}) *Logger {
	return &Logger{
		kit:        kitlog.With(l.kit, keyvals...),
		LoggerBase: l.LoggerBase,
	}
}

// ForSingletonPrefix returns a logging function for use with
// composer.RunAndLogCommand. The function returned inspects the
// number of arguments supplied. If only 1 argument is supplied then
// Logger.Log is called with (prefix, args[0]) as the key-value
// pairs. Otherwise, Logger.Log is called with the complete slice of
// arguments.
//
// The purpose is to cope with reading complete lines from a
// subprocess (that get logged with the prefix), but also with errors
// or other out-of-band messages (with key-val pairs) that come
// through and need to be logged.
func (l *Logger) ForSingletonPrefix(prefix string) func(...interface{}) {
	return func(keyvals ...interface{}) {
		if len(keyvals) == 1 {
			l.Log(prefix, keyvals[0])
		} else {
			l.Log(keyvals...)
		}
	}
}

// If len(ws) == 0 then os.Stderr is used. Otherwise, the logger logs
// out via ws only. Note that NewLogger should only be called once per
// process because it grabs the stdlog and redirects that via the
// new logger.
func NewLogger(ws ...io.Writer) *Logger {
	base := newLoggerBase(ws...)
	logger := kitlog.NewLogfmtLogger(base)

	l := &Logger{
		kit:        kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC),
		LoggerBase: base,
	}
	stdlog.SetOutput(kitlog.NewStdlibAdapter(l.kit))
	return l
}

// Fatal uses the logger to log the error and then calls
// os.Exit(1). Deferred functions are not called, and it is quite
// unlikely you want this function: you certainly should never use
// this function from within an element. It is deliberately not a
// method on Logger to make it less likely to be accidentally used.
func Fatal(l *Logger, err error) {
	l.Log("fatal", err)
	os.Exit(1)
}
