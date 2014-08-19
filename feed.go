package main

type Feed struct {
	Url     string
	Filters []string
	Folder  string
	input   <-chan *item
	output  <-chan *item
}
