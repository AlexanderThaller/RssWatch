package main

import (
	"github.com/AlexanderThaller/config"
	"github.com/AlexanderThaller/logger"
)

type Config struct {
	DataFolder      string
	Feeds           []Feed
	LogLevel        map[string]string
	MailDestination string
	MailDisable     bool
	MailSender      string
	MailServer      string
	SaveFeeds       bool
	XmppDestination string
	XmppDisable     bool
	XmppDomain      string
	XmppPassword    string
	XmppPort        uint16
	XmppSkipTLS     bool
	XmppUsername    string
}

func (co *Config) Default() {
	co.LogLevel = make(map[string]string)
	co.LogLevel["."] = "Notice"

	e := Feed{
		Url:     "https://en.wikipedia.org/w/index.php?title=Special:RecentChanges&feed=atom",
		Filters: []string{".*Talk:.*"},
		Folder:  "misc",
	}

	co.Feeds = append(co.Feeds, e)

	co.DataFolder = "feeds"
	co.SaveFeeds = true
	co.XmppDisable = true
	co.XmppDestination = "admin@ejabberd"
	co.XmppDomain = "ejabberd"
	co.XmppPassword = "test"
	co.XmppPort = 5222
	co.XmppSkipTLS = true
	co.XmppUsername = "test"
	co.MailDisable = false
	co.MailDestination = "alexander@thaller.ws"
	co.MailServer = "mail.thaller.ws:25"
	co.MailSender = "rsswatch@thaller.ws"
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

// setup will prepare the environemt based on the values of the
// given configuration.
func setup(conf *Config) (err error) {
	err = logger.ImportLoggers(conf.LogLevel)
	if err != nil {
		return
	}

	return
}
