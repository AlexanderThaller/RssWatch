// Package logger provides a logging framework similar to those of python
// and haskell.
package logger

import (
	"errors"
	"io"
	"strings"
	"text/template"
	"time"
)

const (
	name = "logger"
)

// Format represents the format which will be used to print the message
// for an logger.
type Format string

// Logger represent different lognames and priorities. They can have
// parents and child loggers which will inherit the priority of the
// parent if it has none. The hirachy of loggers is represented through
// sepperation with dots ('.'). The root logger has the name '.'.
type Logger string

// Priority defines how important a log message is. Loggers will output
// messages which are above their priority level.
type Priority int

// Different priority levels ordered by their severity.
const (
	Trace Priority = iota
	Debug
	Info
	Notice
	Warning
	Error
	Critical
	Alert
	Emergency
	Disable
)

// DefaultPriority of the root logger.
const (
	DefaultPriority   Priority = Notice
	DefaultSepperator          = "."
)

var (
	format     = "[{{.Time}} {{.Priority}} {{.Logger}}] - {{.Message}}.\n"
	timeformat = time.RFC3339

	priorities     map[Priority]string
	list           loggers
	formattemplate template.Template

	// SaveLoggerLevels will make the package save loggers which are only defined
	// by their parents if it is set to true.
	SaveLoggerLevels = true
)

func init() {
	list = newLoggers()

	priorities = make(map[Priority]string)
	priorities[Trace] = "Trace"
	priorities[Debug] = "Debug"
	priorities[Info] = "Info"
	priorities[Notice] = "Notice"
	priorities[Warning] = "Warning"
	priorities[Error] = "Error"
	priorities[Critical] = "Critical"
	priorities[Alert] = "Alert"
	priorities[Emergency] = "Emergency"
	priorities[Disable] = "Disabled"
}

// ImportLoggers sets the LogLevel for the given Loggers.
func ImportLoggers(lo map[Logger]string) (err error) {
	if lo == nil {
		err = errors.New("the loglevel map is nil")
		return
	}

	for k, v := range lo {
		p, e := ParsePriority(v)
		if e != nil {
			return e
		}

		SetLevel(k, p)
	}

	return
}

// New will return a logger with the given name.
func New(na ...string) (log Logger) {
	s := strings.Join(na, DefaultSepperator)
	return Logger(s)
}

// GetLevel returns the priority level of the given logger.
func GetLevel(lo Logger) (pri Priority) {
	return list.GetLevel(lo)
}

// SetLevel sets the priority level for the given logger.
func SetLevel(lo Logger, pr Priority) (err error) {
	err = list.SetLevel(lo, pr)
	if err != nil {
		return
	}

	return
}

// SetFormat changes the message format for the given logger. Avaivable
// fields are:
//
// Time: The time when the message is printed.
//
// Logger: The name of the logger.
//
// Priority: The priority of the logger.
//
// Message: The output message.
//
// The default Format is:
//
// "[{{.Time}} {{.Logger}} {{.Priority}}] - {{.Message}}.\n"
func SetFormat(lo Logger, fo Format) error {
	return list.SetFormat(lo, fo)
}

// SetTimeFormat sets the TimeFormat which will be used in the message
// format for the specified logger
//
// The default format is: RFC3339
func SetTimeFormat(lo Logger, fo string) error {
	return list.SetTimeFormat(lo, fo)
}

// SetNoColor sets the nocolor flag for the given logger. If true no
// colors will be printed for the logger.
func SetNoColor(lo Logger, nc bool) {
	list.SetNoColor(lo, nc)
}

// SetOutput sets the output parameter of the logger to the given
// io.Writer. The default is os.Stderr.
func SetOutput(lo Logger, ou io.Writer) error {
	return list.SetOutput(lo, ou)
}

// ParsePriority tries to parse the priority by the given string.
func ParsePriority(pr string) (Priority, error) {
	for k, v := range priorities {
		if v == pr {
			return k, nil
		}
	}

	e := errors.New("can not parse priority: do not recognize " + pr)
	return DefaultPriority, e
}

// NamePriority returns the string value of the given priority.
func NamePriority(pr Priority) (pri string, err error) {
	err = checkPriority(pr)
	if err != nil {
		return
	}

	pri = priorities[pr]

	return
}

func logMessage(lo Logger, pr Priority, me ...interface{}) {
	l := list.GetLogger(lo)

	if l.Priority > pr {
		return
	}

	printMessage(l, pr, me...)
}

// Log logs a message with the given priority.
func (lo Logger) Log(pr Priority, me ...interface{}) {
	logMessage(lo, pr, me)
}

// Trace logs a message with the Trace priority.
func (lo Logger) Trace(me ...interface{}) {
	logMessage(lo, Trace, me...)
}

// Debug logs a message with the Debug priority.
func (lo Logger) Debug(me ...interface{}) {
	logMessage(lo, Debug, me...)
}

// Info logs a message with the Debug priority.
func (lo Logger) Info(me ...interface{}) {
	logMessage(lo, Info, me...)
}

// Notice logs a message with the Debug priority.
func (lo Logger) Notice(me ...interface{}) {
	logMessage(lo, Notice, me...)
}

// Warning logs a message with the Debug priority.
func (lo Logger) Warning(me ...interface{}) {
	logMessage(lo, Warning, me...)
}

// Error logs a message with the Debug priority.
func (lo Logger) Error(me ...interface{}) {
	logMessage(lo, Error, me...)
}

// Critical logs a message with the Debug priority.
func (lo Logger) Critical(me ...interface{}) {
	logMessage(lo, Critical, me...)
}

// Alert logs a message with the Debug priority.
func (lo Logger) Alert(me ...interface{}) {
	logMessage(lo, Alert, me...)
}

// Emergency logs a message with the Debug priority.
func (lo Logger) Emergency(me ...interface{}) {
	logMessage(lo, Emergency, me...)
}

// GetLevel returns the priority level of the logger.
func (lo Logger) GetLevel() Priority {
	return GetLevel(lo)
}

// SetLevel sets the priority level for the Logger.
func (lo Logger) SetLevel(pr Priority) (err error) {
	err = SetLevel(lo, pr)
	if err != nil {
		return
	}

	return
}

// SetFormat changes the message format for the Logger. Avaivable fields
// are:
//
// Time: The time when the message is printed.
//
// Logger: The name of the logger.
//
// Priority: The priority of the logger.
//
// Message: The output message.
//
// The default Format is:
//
// "[{{.Time}} {{.Logger}} {{.Priority}}] - {{.Message}}.\n"
func (lo Logger) SetFormat(fo Format) {
	SetFormat(lo, fo)
}

// SetTimeFormat sets the TimeFormat which will be used in the message
// format for the Logger
//
// The default format is: RFC3339
func (lo Logger) SetTimeFormat(fo string) error {
	return SetTimeFormat(lo, fo)
}

// SetNoColor sets the nocolor flag for the given logger. If true no
// colors will be printed for the logger.
func (lo Logger) SetNoColor(nc bool) {
	SetNoColor(lo, nc)
}

// SetOutput sets the output parameter of the logger to the given
// io.Writer. The default is os.Stderr.
func (lo Logger) SetOutput(ou io.Writer) {
	SetOutput(lo, ou)
}
