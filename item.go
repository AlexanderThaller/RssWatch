package main

import (
	"strings"

	"github.com/SlyMarbo/rss"
)

type Item struct {
	Feed   *rss.Feed
	Item   *rss.Item
	Folder string
	Filter string
}

func (item *Item) String() string {
	s := strings.TrimSpace(item.Feed.Title) + " - "
	s += strings.TrimSpace(item.Item.Title)
	return s
}
