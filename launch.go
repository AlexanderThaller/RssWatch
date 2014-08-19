package main

import (
	"github.com/AlexanderThaller/logger"
	"github.com/AlexanderThaller/service"
)

// launch will start the appliaction based on the given configuration.
func launch(conf *Config) (err error) {
	l := logger.New(name, "launch")

	l.Trace("Watching for signals")
	service.WatchSignals()
	return nil
}
