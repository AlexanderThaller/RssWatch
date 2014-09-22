package main

import (
	"github.com/AlexanderThaller/config"
	"github.com/AlexanderThaller/logger"
)

type Config struct {
	LogLevel        map[logger.Logger]string
	DataFolder      string
	Feeds           []Feed
	SaveFeeds       bool
	XmppDisable     bool
	XmppDestination string
	XmppDomain      string
	XmppPassword    string
	XmppPort        uint16
	XmppSkipTLS     bool
	XmppUsername    string
	MailDisable     bool
	MailDestination string
	MailServer      string
	MailSender      string
}

func (co *Config) Default() {
	co.LogLevel = make(map[logger.Logger]string)
	co.LogLevel["."] = "Notice"

	e := Feed{
		Url:     "https://en.wikipedia.org/w/index.php?title=Special:RecentChanges&feed=atom",
		Filters: []string{".*Talk:.*"},
		Folder:  "misc",
	}

	co.Feeds = append(co.Feeds, e)

	co.DataFolder = "feeds"
	co.SaveFeeds = true
	co.XmppDisable = false
	co.XmppDestination = "admin@ejabberd"
	co.XmppDomain = "ejabberd"
	co.XmppPassword = "test"
	co.XmppPort = 5222
	co.XmppSkipTLS = true
	co.XmppUsername = "test"
	co.MailDisable = true
	co.MailDestination = "alexander@thaller.ws"
	co.MailServer = "mail.thaller.ws:25"
	co.MailSender = "rsswatch@thaller.ws"
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

// setup will prepare the environemt based on the values of the
// given configuration.
func setup(conf *Config) (err error) {
	err = logger.ImportLoggers(conf.LogLevel)
	if err != nil {
		return
	}

	return
}
