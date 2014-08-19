package main

import "regexp"

type Filter struct {
	filter *regexp.Regexp
	Filter string
	input  <-chan *item
	output <-chan *item
}
