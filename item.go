package main

import (
	rss "github.com/AlexanderThaller/rss-1"
)

type Item struct {
	Filter string
	data   *rss.Item
}
