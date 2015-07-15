// The MIT License (MIT)
//
// Copyright (c) 2015 Alexander Thaller
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import "github.com/AlexanderThaller/config"

type Config struct {
	DataFolder      string
	MailDestination string
	MailDisable     bool
	MailSender      string
	MailServer      string
	SaveFeeds       bool
	Profile         bool
	ProfileBind     string

	Feeds []Feed
}

func (co *Config) Default() {
	e1 := Feed{
		Url:    "https://en.wikipedia.org/w/index.php?title=RSS&feed=atom&action=history",
		Folder: "misc.wikipedia",
	}

	e2 := Feed{
		Url:     "https://en.wikipedia.org/w/index.php?title=Special:RecentChanges&feed=atom",
		Filters: []string{".*", ".*Talk:.*"},
		Folder:  "misc.wikipedia",
	}

	co.Feeds = append(co.Feeds, e1)
	co.Feeds = append(co.Feeds, e2)

	co.DataFolder = "feeds"
	co.SaveFeeds = true
	co.MailDisable = false
	co.MailDestination = "myemail@example.com"
	co.MailServer = "mail.example.com:25"
	co.MailSender = "rsswatch@example.com"
	co.Profile = false
	co.ProfileBind = "localhost:6060"
}

func (co *Config) Format() config.Format {
	return config.FormatTOML
}

// configure will parse the config file and return a new Config.
func configure(path string) (conf *Config, err error) {
	c := new(Config)
	err = config.Configure(path, c)
	if err != nil {
		return
	}

	conf = c
	return
}
