package logger

import (
	"fmt"
	"time"
)

func printMessage(lo logger, pr Priority, me ...interface{}) {
	m := new(message)
	m.Time = time.Now().Format(string(lo.TimeFormat))
	m.Logger = lo.Logger
	m.Priority = formatPriority(pr, lo.NoColor)
	m.Message = fmt.Sprint(me...)

	s := formatMessage(m, lo.Format)

	fmt.Fprint(lo.Output, s)
}
