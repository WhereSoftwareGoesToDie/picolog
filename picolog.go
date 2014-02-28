package picolog

import (
	"bufio"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"strings"
)

// Logger is a leveled logger type.
type Logger struct {
	logLevel    syslog.Priority
	logger      *log.Logger
	writer      *bufio.Writer
	destStream *os.File
	prefix string
	subloggers []*Logger
	initialized bool
}

// ParseLogLevel takes a string and returns a syslog.Priority according
// to the standard syslog string representation. Not case-sensitive.
func ParseLogLevel(level string) (syslog.Priority, error) {
	level = strings.ToLower(level)
	switch {
	case level == "emerg":
		return syslog.LOG_EMERG, nil
	case level == "alert":
		return syslog.LOG_ALERT, nil
	case level == "crit":
		return syslog.LOG_CRIT, nil
	case level == "err":
		return syslog.LOG_ERR, nil
	case level == "warning":
		return syslog.LOG_WARNING, nil
	case level == "notice":
		return syslog.LOG_NOTICE, nil
	case level == "info":
		return syslog.LOG_INFO, nil
	case level == "debug":
		return syslog.LOG_DEBUG, nil
	}
	return syslog.Priority(0), fmt.Errorf("Invalid log level: %s", level)
}

// Return a new Logger. logLevel is a syslog log level,
// subpackage is used to construct the log prefix, and dest is where to
// write the log to.
func NewLogger(logLevel syslog.Priority, subpackage string, dest *os.File) *Logger {
	logger := new(Logger)
	logger.logLevel = logLevel
	flags := log.Ldate | log.Ltime
	// If logging at DEBUG, include file paths and line numbers
	if logLevel == syslog.LOG_DEBUG {
		flags |= log.Lshortfile
	}
	logger.prefix = subpackage
	renderedPrefix := fmt.Sprintf("[%s] ", logger.prefix)
	logger.destStream = dest
	logger.writer = bufio.NewWriter(logger.destStream)
	logger.logger = log.New(logger.writer, renderedPrefix, flags)
	logger.initialized = true
	return logger
}

// NewDefaultLogger returns a picolog.Logger initialized with workable
// defaults (outputs to stderr, prefix "default", priority DEBUG).
// Useful as a fallback when a logger hasn't been initialized.
func NewDefaultLogger() *Logger {
	return NewLogger(syslog.LOG_DEBUG, "default", os.Stderr)
}

// initializeDefaultLogger takes a (possibly nil) *Logger and allocates
// and assigns a default logger as returned by NewDefaultLogger.
func (l *Logger) initializeDefaultLogger() {
	defaultLogger := NewDefaultLogger()
	l = defaultLogger
}

// ensureInitialized checks if the initialized flag has been set for l,
// and if not initializes a default logger.
func (l *Logger) ensureInitialized() {
	if !l.initialized {
		l.initializeDefaultLogger()
	}
}

// NewSubLogger returns a Logger writing to the same stream, with a
// prefix constructed from the provided prefix and the parent Logger's
// prefix. Subloggers can be nested.
func (l *Logger) NewSubLogger(prefix string) *Logger {
	subPrefix := fmt.Sprintf("%s][%s", l.prefix, prefix)
	sub := NewLogger(l.logLevel, subPrefix, l.destStream)
	l.subloggers = append(l.subloggers, sub)
	return sub
}

// Printf is the lowest-level output function of our Logger. Will use a
// default logger if l is not initialized.
func (l *Logger) Printf(format string, level syslog.Priority, v ...interface{}) {
	l.ensureInitialized()
	if level <= l.logLevel {
		msg := fmt.Sprintf(format, v...)
		// We use logger.Output rather than logger.Printf
		// so we can pass a custom calldepth for file
		// {path,line}-resolution purposes (the default of 2
		// is only useful when using the Logger type directly).
		l.logger.Output(3, msg)
		l.writer.Flush()
	}
}

// Debugf logs one printf-formatted message at LOG_DEBUG.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Printf(format, syslog.LOG_DEBUG, v...)
}

// Errorf logs one printf-formatted message at LOG_ERR.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Printf(format, syslog.LOG_ERR, v...)
}

// Warningf logs one printf-formatted message at LOG_WARNING.
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.Printf(format, syslog.LOG_WARNING, v...)
}

// Fatalf logs one printf-formatted message at LOG_CRIT, and then exits
// with an error code.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Printf(format, syslog.LOG_CRIT, v...)
	os.Exit(1)
}

// Infof logs one printf-formatted message at LOG_INFO.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Printf(format, syslog.LOG_INFO, v...)
}
