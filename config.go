package main

import (
	"github.com/AlexanderThaller/config"
	"github.com/AlexanderThaller/logger"
)

type Config struct {
	DataFolder      string
	MailDestination string
	MailDisable     bool
	MailSender      string
	MailServer      string
	SaveFeeds       bool
	Profile         bool
	ProfileBind     string

	LogLevel map[string]string
	Feeds    []Feed
}

func (co *Config) Default() {
	co.LogLevel = make(map[string]string)
	co.LogLevel["."] = "Info"

	e := Feed{
		Url:     "https://en.wikipedia.org/w/index.php?title=Special:RecentChanges&feed=atom",
		Filters: []string{".*", ".*Talk:.*"},
		Folder:  "misc.wikipedia",
	}

	co.Feeds = append(co.Feeds, e)

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

// setup will prepare the environemt based on the values of the
// given configuration.
func setup(conf *Config) (err error) {
	err = logger.ImportLoggers(conf.LogLevel)
	if err != nil {
		return
	}

	return
}
