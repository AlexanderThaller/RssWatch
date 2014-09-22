package main

import "time"

type Item struct {
	Feed    string
	Filter  string
	Title   string
	Content string
	Link    string
	Date    time.Time
}

func (item *Item) Send(conf *Config) {
	return
}
