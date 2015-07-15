// The MIT License (MIT)
//
// Copyright (c) 2015 Alexander Thaller
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"flag"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/juju/errgo"
)

const (
	name                     = "RssWatch"
	DefaultChannelBufferSize = 50000
)

var (
	buildVersion string
	buildTime    string
)

func main() {
	var (
		configPath        = flag.String("config.path", "config.toml", "the path to the config file.")
		logLevel          = flag.String("log.level", "info", "the loglevel of the application.")
		prometheusBinding = flag.String("prometheus.binding", ":9132", "the address and port to bind the prometheus metrics to.")
	)
	flag.Parse()

	// Set loglevel
	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal(errgo.Notef(err, "can not parse loglevel"))
	}
	log.SetLevel(level)

	// Load configuration
	configuration, err := configure(*configPath)
	if err != nil {
		log.Fatal("Can not configure ", errgo.Details(err))
	}
	log.Debug("Configuration: ", fmt.Sprintf("%+v", configuration))

	// Launch
	err = launch(configuration, *prometheusBinding)
	if err != nil {
		log.Fatal("Problem while launching: ", errgo.Details(err))
	}
}
