package main

type Config struct {
	DataFolder      string
	SaveFeeds       bool
	XmppDestination string
	Feeds           []string
	Filters         []string
}
