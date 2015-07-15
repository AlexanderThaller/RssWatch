package main

import (
	"github.com/AlexanderThaller/rss"
)

type Item struct {
	Filter string
	data   *rss.Item
}
