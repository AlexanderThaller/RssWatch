package logger

import (
	"fmt"
	"strings"
)

const (
	colornone   = 0
	colorred    = 31
	colorgreen  = 32
	coloryellow = 33
	colorblue   = 34

	textnormal = 0
	textbold   = 1
	textblink  = 5
)

type message struct {
	Logger
	Message  string
	Priority string
	Time     string
}

func formatPriority(pr Priority, nc bool) string {
	c, f := getPriorityFormat(pr)

	r := formatText(f, nc)
	l := formatText(c, nc)
	p := priorities[pr]

	s := r + l + p + formatReset(nc)

	return s
}

func formatReset(nc bool) string {
	s := formatText(textnormal, nc)

	return s
}

func formatText(fo int, nc bool) string {
	if nc {
		return ""
	}

	s := fmt.Sprintf("\033[%dm", fo)
	return s
}

func formatMessage(me *message, fo Format) (so string) {
	so = strings.Replace(string(fo), "{{.Time}}", me.Time, -1)
	so = strings.Replace(so, "{{.Logger}}", string(me.Logger), -1)
	so = strings.Replace(so, "{{.Priority}}", me.Priority, -1)
	so = strings.Replace(so, "{{.Message}}", me.Message, -1)

	return
}
