package service

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexanderThaller/logger"
)

type Service interface {
	Start(chan<- Message) error
	Stop()
	Reload()
}

const (
	ChannelBufferSize = 5000
)

var (
	Messages chan<- Message
)

var (
	messages chan Message
	services map[string]Service
)

func init() {
	messages = make(chan Message, ChannelBufferSize)
	Messages = messages

	services = make(map[string]Service)
}

// Start will start a given service and return the message channel. This
// function can be run multible times and will always return the same message
// channel.
func Start(name string, service Service) (<-chan Message, error) {
	l := logger.New("service", "Start")
	l.Trace("Starting new service: ", name)

	err := service.Start(messages)
	if err != nil {
		return nil, err
	}

	services[name] = service

	return messages, nil
}

// WatchSignals waits for unix signals and handles them appropiatly.
// SIGINT/SIGTERM: Stop services and return.
// SIGHUP: Reload services.
// SIGQUIT: Dump core and return.
func WatchSignals() {
	sigc := make(chan os.Signal, 1)

	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)

	for {
		signal := <-sigc

		switch signal {
		case syscall.SIGINT:
			stopServices()
			return
		case syscall.SIGTERM:
			stopServices()
			return
		}
	}
}

func stopServices() {
	for _, d := range services {
		d.Stop()
	}
}

// Count returns the number of services.
func Count() uint {
	return uint(len(services))
}
