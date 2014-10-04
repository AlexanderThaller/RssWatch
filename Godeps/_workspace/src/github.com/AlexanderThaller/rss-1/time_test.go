package rss

import (
	"log"
	"testing"
)

func Test_ParseTime(t *testing.T) {
	sources := map[string]string{
		"2014-09-30T17:22:24+00:00":           "2014-09-30 17:22:24 +0000 +0000",
		"Thu, 02 Oct 14 00:00:00 -0400":       "2014-10-02 00:00:00 -0400 -0400",
		"28 Sep 2014 22:16:00 +0100":          "2014-09-28 22:16:00 +0100 +0100",
		"Thur, 27 June 2013 04:24:36 GMT":     "2013-06-27 04:24:36 +0000 GMT",
		"Tues, 12 February 2013 04:24:36 GMT": "2013-02-12 04:24:36 +0000 GMT",
		"Thus, 3 January 2013 04:24:36 GMT":   "2013-01-03 04:24:36 +0000 GMT",
	}

	for source, expected := range sources {
		parsed, err := parseTime(source)

		output := parsed.String()
		if output != expected {
			log.Print("SOURCE: ", source)
			log.Print("ERROR: ", err)
			log.Print("GOT: '", output, "', EXPECTED: '", expected, "'")
		}
	}
}
