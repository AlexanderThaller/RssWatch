package rss

import (
	"strings"
	"time"

	"github.com/jinzhu/now"
	"github.com/juju/errgo"
)

func init() {
	now.TimeFormats = append(now.TimeFormats, time.ANSIC)
	now.TimeFormats = append(now.TimeFormats, time.UnixDate)
	now.TimeFormats = append(now.TimeFormats, time.RubyDate)
	now.TimeFormats = append(now.TimeFormats, time.RFC822)
	now.TimeFormats = append(now.TimeFormats, time.RFC822Z)
	now.TimeFormats = append(now.TimeFormats, time.RFC850)
	now.TimeFormats = append(now.TimeFormats, time.RFC1123)
	now.TimeFormats = append(now.TimeFormats, time.RFC1123Z)
	now.TimeFormats = append(now.TimeFormats, time.RFC3339)
	now.TimeFormats = append(now.TimeFormats, time.RFC3339Nano)
	now.TimeFormats = append(now.TimeFormats, "Mon, _2 Jan 2006 15:04:05 MST")
	now.TimeFormats = append(now.TimeFormats, "Mon, _2 Jan 2006 15:04:05 -0700")
	now.TimeFormats = append(now.TimeFormats, "Mon, _2 January 2006 15:04:05 MST")
	now.TimeFormats = append(now.TimeFormats, "Mon, _2 January 2006 15:04:05 -0700")
	now.TimeFormats = append(now.TimeFormats, "Mon, _2 Jan 06 15:04:05 -0700")
	now.TimeFormats = append(now.TimeFormats, "_2 Jan 2006 15:04:05 -0700")
}

func parseTime(source string) (time.Time, error) {
	source = strings.TrimSpace(source)
	source = strings.Replace(source, "Tues", "Tue", 1)
	source = strings.Replace(source, "Thur", "Thu", 1)
	source = strings.Replace(source, "Thus", "Thu", 1)

	t, err := now.Parse(source)
	if err != nil {
		return time.Time{}, errgo.New(err.Error())
	}

	return t, nil
}
