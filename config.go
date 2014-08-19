package main

type Feed struct {
	Url     string
	Filters []string
	Folder  string
}

type config struct {
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
}

func (co *config) Default() {
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
}
