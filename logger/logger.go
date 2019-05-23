package logger

import (
	"io"
	"os"

	stdlog "log"

	kitlog "github.com/go-kit/kit/log"
)

type Logger struct {
	// To understand this a bit better, consider that implementations
	// of Logger include the keyval pairs and a reference to the
	// underlying logger...
	kit kitlog.Logger

	// ...whereas SwapLogger is just the underlying logger
	swappable *kitlog.SwapLogger
}

// Log keyvals through the logger. Keyvals is a slice of key-value
// pairs, and are logged as key=value. The logger logs the context and
// the keyvals.
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
	logger := kitlog.With(l.kit, keyvals...)
	return &Logger{
		kit:       logger,
		swappable: l.swappable,
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
// subprocess, but coping errors or other out-of-band messages that
// come through and need to be logged.
func (l *Logger) ForSingletonPrefix(prefix string) func(...interface{}) {
	return func(keyvals ...interface{}) {
		if len(keyvals) == 1 {
			l.Log(prefix, keyvals[0])
		} else {
			l.Log(keyvals...)
		}
	}
}

type wrapper struct {
	l *Logger
}

func (w wrapper) Write(p []byte) (n int, err error) {
	if err := w.l.kit.Log("msg", p); err != nil {
		return 0, err
	} else {
		return len(p), nil
	}
}

// If len(ws) == 0 then os.Stderr is used. Otherwise, the logger logs
// out via ws only. Note that NewLogger should only be called once per
// process because it grabs the stdlog and redirects that via the
// new logger.
func NewLogger(ws ...io.Writer) *Logger {
	w := io.Writer(os.Stderr)
	if len(ws) != 0 {
		w = io.MultiWriter(ws...)
	}
	logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(w))
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC)

	sl := &kitlog.SwapLogger{}
	sl.Swap(logger)
	l := &Logger{
		kit:       sl,
		swappable: sl,
	}
	stdlog.SetOutput(&wrapper{l: l})
	return l
}

// Replace the underlying writers of not only this logger, but the
// entire tree of loggers created by any calls to With from the root
// downwards. As with NewLogger, if len(ws) == 0 then os.Stderr is
// used, otherwise only ws is used.
func (l *Logger) SwapWriters(ws ...io.Writer) {
	w := io.Writer(os.Stderr)
	if len(ws) != 0 {
		w = io.MultiWriter(ws...)
	}
	logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(w))
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC)
	l.swappable.Swap(logger)
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
