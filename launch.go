package main

import (
	"github.com/AlexanderThaller/logger"
	"github.com/AlexanderThaller/service"
)

func launch(conf *Config) error {
	l := logger.New(name, "launch")

	for _, d := range conf.Feeds {
		err := d.Launch(conf)
		if err != nil {
			return err
		}
	}

	l.Trace("Watching for signals")
	service.WatchSignals()
	return nil
}
