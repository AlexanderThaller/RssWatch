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

	log "github.com/Sirupsen/logrus"
	"github.com/juju/errgo"
	"github.com/spf13/viper"
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
		logLevel = flag.String("log.level", "info", "the loglevel of the application.")
		//prometheusBinding = flag.String("prometheus.binding", ":9132", "the address and port to bind the prometheus metrics to.")
	)
	flag.Parse()

	// Set loglevel
	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal(errgo.Notef(err, "can not parse loglevel"))
	}
	log.SetLevel(level)

	// Load configuration
	err = configure()
	if err != nil {
		log.Fatal("can not configure ", err)
	}
	log.Debug("config: ", viper.AllSettings())

	log.Debug("feeds: ", viper.GetStringMap("feeds"))

	// Launch
	/*err = launch(*prometheusBinding)
	if err != nil {
		log.Fatal("Problem while launching: ", errgo.Details(err))
	}*/
}

// configure will set default values for the configuration
func configure() error {
	viper.SetConfigName("config")

	viper.AddConfigPath("$HOME/.rsswatch")
	viper.AddConfigPath("/etc/rsswatch/")
	viper.AddConfigPath("/usr/local/etc/rsswatch/")

	viper.SetDefault("Feeds.DataFolder", "feeds")
	viper.SetDefault("Feeds.Save", true)
	viper.SetDefault("Mail.Destination", "myemail@example.com")
	viper.SetDefault("Mail.Enable", true)
	viper.SetDefault("Mail.Sender", "rsswatch@example.com")
	viper.SetDefault("Mail.Server", "mail.example.com:25")

	err := viper.ReadInConfig()
	if err != nil {
		return errgo.Notef(err, "can not read in config file")
	}

	return nil
}
