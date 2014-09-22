package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlexanderThaller/logger"
	"github.com/AlexanderThaller/service"
	"github.com/SlyMarbo/rss"
	"github.com/juju/errgo"
	"github.com/vmihailenco/msgpack"
)

func launch(conf *Config) (err error) {
	l := logger.New(name, "launch")

	for _, d := range conf.Feeds {
		go d.Launch(conf)
	}

	l.Trace("Watching for signals")
	service.WatchSignals()
	return nil
}
