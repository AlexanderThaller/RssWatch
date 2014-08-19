package logger

import (
	"io"
	"os"
	"sync"
)

const (
	defroot      = Logger(".")
	defseperator = "."
)

var (
	defout = os.Stderr
)

type logger struct {
	Format
	Logger
	Priority
	TimeFormat string
	NoColor    bool
	Output     io.Writer
}

type loggers struct {
	data  map[Logger]logger
	mutex sync.RWMutex
}

func newLoggers() loggers {
	l := new(loggers)
	l.data = make(map[Logger]logger)

	r := logger{
		Format:     Format(format),
		Priority:   DefaultPriority,
		TimeFormat: timeformat,
		Logger:     defroot,
		NoColor:    false,
		Output:     defout,
	}

	l.data[defroot] = r

	return *l
}

func (lo *loggers) SetLogger(na Logger, lg logger) {
	lo.mutex.Lock()

	lo.data[na] = lg

	lo.mutex.Unlock()
}

func (lo *loggers) GetLogger(na Logger) logger {
	lo.mutex.RLock()

	l, x := lo.data[na]

	lo.mutex.RUnlock()

	if !x {
		l = lo.GetParentLogger(na)
	}

	return l
}

func (lo *loggers) GetParentLogger(na Logger) logger {
	n := getParent(na)

	l := lo.GetLogger(n)
	l.Logger = na

	if SaveLoggerLevels {
		lo.SetLogger(na, l)
	}

	return l
}

func (lo *loggers) GetLevel(na Logger) Priority {
	l := lo.GetLogger(na)

	return l.Priority
}

func (lo *loggers) SetLevel(na Logger, pr Priority) (err error) {
	err = checkPriority(pr)
	if err != nil {
		return
	}

	l := lo.GetLogger(na)
	l.Priority = pr
	lo.SetLogger(na, l)

	return
}

func (lo *loggers) SetFormat(na Logger, fo Format) (err error) {
	//TODO: Validate Format
	l := lo.GetLogger(na)
	l.Format = fo
	lo.SetLogger(na, l)

	return
}

func (lo *loggers) SetTimeFormat(na Logger, fo string) (err error) {
	//TODO: Validate TimeFormat
	l := lo.GetLogger(na)
	l.TimeFormat = fo
	lo.SetLogger(na, l)

	return
}

func (lo *loggers) SetNoColor(na Logger, nc bool) {
	l := lo.GetLogger(na)
	l.NoColor = nc
	lo.SetLogger(na, l)

	return
}

func (lo *loggers) SetOutput(na Logger, ou io.Writer) (err error) {
	l := lo.GetLogger(na)
	l.Output = ou
	lo.SetLogger(na, l)

	return
}
