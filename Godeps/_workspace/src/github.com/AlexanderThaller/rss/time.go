package rss

import (
	"strings"
	"time"

	"github.com/jinzhu/now"
)

func parseTime(s string) (time.Time, error) {
	formats := []string{
		"Mon, _2 Jan 2006 15:04:05 MST",
		"Mon, _2 January 2006 15:04:05 MST",
		"Mond, _2 January 2006 15:04:05 MST",
		"Mon, _2 Jan 2006 15:04:05 -0700",
		"Mon, _2 Jan 06 15:04:05 -0700",
		"_2 Jan 2006 15:04:05 -0700",
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, format := range formats {
		now.TimeFormats = append(now.TimeFormats, format)
	}

	s = strings.TrimSpace(s)

	replaces := make(map[string]string)
	replaces["Thur"] = "Thu"
	replaces["Tues"] = "Tue"
	replaces["Thus"] = "Thu"

	for from, to := range replaces {
		s = strings.Replace(s, from, to, -1)
	}

	t, err := now.Parse(s)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}
