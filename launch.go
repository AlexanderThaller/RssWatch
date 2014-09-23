package main

import (
	"github.com/AlexanderThaller/logger"
	"github.com/AlexanderThaller/service"
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
