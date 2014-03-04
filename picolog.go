/*
picolog is a tiny levelled logging package for go. It supports syslog
log levels, subloggers, and not much else. Written because all the
existing solutions either didn't do what I needed or were too weighty.
*/
package picolog

import (
	"bufio"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"strings"
)

// Logger is a leveled logger type. It can be a sublogger of another
// logger, and have an arbitrary number of subloggers itself. 
type Logger struct {
	logLevel    LogLevel
	logger      *log.Logger
	writer      *bufio.Writer
	destStream *os.File
	prefix string
	subloggers []*Logger
	initialized bool
}

// LogLevel is a type representing the usual syslog log levels from
// LOG_DEBUG to LOG_EMERG. It does not reflect the go syslog package's
// concept of 'Priority'. 
type LogLevel syslog.Priority

const (
	LogDebug LogLevel = LogLevel(syslog.LOG_DEBUG)
	LogInfo = LogLevel(syslog.LOG_INFO)
	LogNotice = LogLevel(syslog.LOG_NOTICE)
	LogWarning = LogLevel(syslog.LOG_WARNING)
	LogErr = LogLevel(syslog.LOG_ERR)
	LogCrit = LogLevel(syslog.LOG_CRIT)
	LogAlert = LogLevel(syslog.LOG_ALERT)
	LogEmerg = LogLevel(syslog.LOG_EMERG)
)

// ParseLogLevel takes a string and returns a LogLevel according
// to the standard syslog string representation. Not case-sensitive.
func ParseLogLevel(level string) (LogLevel, error) {
	level = strings.ToLower(level)
	switch {
	case level == "emerg":
		return LogLevel(syslog.LOG_EMERG), nil
	case level == "alert":
		return LogLevel(syslog.LOG_ALERT), nil
	case level == "crit":
		return LogLevel(syslog.LOG_CRIT), nil
	case level == "err":
		return LogLevel(syslog.LOG_ERR), nil
	case level == "warning":
		return LogLevel(syslog.LOG_WARNING), nil
	case level == "notice":
		return LogLevel(syslog.LOG_NOTICE), nil
	case level == "info":
		return LogLevel(syslog.LOG_INFO), nil
	case level == "debug":
		return LogLevel(syslog.LOG_DEBUG), nil
	}
	return LogLevel(syslog.Priority(0)), fmt.Errorf("Invalid log level: %s", level)
}

// String returns the default (lowercase) string representation of a
// LogLevel.
func (l LogLevel) String() string {
	switch l {
	case LogLevel(syslog.LOG_EMERG):
		return "emerg"
	case LogLevel(syslog.LOG_ALERT):
		return "alert"
	case LogLevel(syslog.LOG_CRIT):
		return "crit"
	case LogLevel(syslog.LOG_ERR):
		return "err"
	case LogLevel(syslog.LOG_WARNING):
		return "warning"
	case LogLevel(syslog.LOG_NOTICE):
		return "notice"
	case LogLevel(syslog.LOG_INFO):
		return "info"
	case LogLevel(syslog.LOG_DEBUG):
		return "debug"
	}
	return "invalid log level"
}

// Return a new Logger. logLevel is a syslog log level,
// subpackage is used to construct the log prefix, and dest is where to
// write the log to.
func NewLogger(logLevel LogLevel, subpackage string, dest *os.File) *Logger {
	logger := new(Logger)
	logger.logLevel = logLevel
	flags := log.Ldate | log.Ltime
	// If logging at DEBUG, include file paths and line numbers
	if logLevel == LogLevel(syslog.LOG_DEBUG) {
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
	return NewLogger(LogDebug, "default", os.Stderr)
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
func (l *Logger) Printf(format string, level LogLevel, v ...interface{}) {
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
	l.Printf(format, LogDebug, v...)
}

// Errorf logs one printf-formatted message at LOG_ERR.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Printf(format, LogErr, v...)
}

// Warningf logs one printf-formatted message at LOG_WARNING.
func (l *Logger) Warningf(format string, v ...interface{}) {
	l.Printf(format, LogWarning, v...)
}

// Fatalf logs one printf-formatted message at LOG_CRIT, and then exits
// with an error code.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Printf(format, LogErr, v...)
	os.Exit(1)
}

// Infof logs one printf-formatted message at LOG_INFO.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Printf(format, LogInfo, v...)
}

// Emergf logs one printf-formatted message at LOG_EMERG.
func (l *Logger) Emergf(format string, v ...interface{}) {
	l.Printf(format, LogEmerg, v...)
}

// Alertf logs one printf-formatted message at LOG_ALERT.
func (l *Logger) Alertf(format string, v ...interface{}) {
	l.Printf(format, LogAlert, v...)
}

// Noticef logs one printf-formatted message at LOG_NOTICE.
func (l *Logger) Noticef(format string, v ...interface{}) {
	l.Printf(format, LogNotice, v...)
}
