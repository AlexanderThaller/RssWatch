package main

type config struct {
	DataFolder      string
	Feeds           []string
	Filters         []string
	SaveFeeds       bool
	XmppDestination string
	XmppDomain      string
	XmppPassword    string
	XmppPort        uint16
	XmppSkipTLS     bool
	XmppUsername    string
}

func (co *config) Default() {
	e := "https://en.wikipedia.org/w/index.php?title=Special:RecentChanges&feed=atom"
	co.Feeds = append(co.Feeds, e)

	i := ".*Talk:.*"
	co.Filters = append(co.Filters, i)

	co.DataFolder = "feeds"
	co.SaveFeeds = true
	co.XmppDestination = "admin@ejabberd"
	co.XmppDomain = "ejabberd"
	co.XmppPassword = "test"
	co.XmppPort = 5222
	co.XmppSkipTLS = true
	co.XmppUsername = "test"
}
