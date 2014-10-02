package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/AlexanderThaller/logger"
	"github.com/AlexanderThaller/service"
	"github.com/juju/errgo"
)

const (
	name                     = "RssWatch"
	DefaultChannelBufferSize = 5000
)

var (
	buildVersion string
	buildTime    string

	flagConfigPath = flag.String("config", name+".cnf.toml",
		"The path to the config file.")

	configuration *Config
)

func init() {
	flag.Parse()
	l := logger.New(name, "init")

	// Load configuration
	var err error
	configuration, err = configure(*flagConfigPath)
	if err != nil {
		l.Alert("Can not configure ", errgo.Details(err))
		os.Exit(1)
	}

	// Setup environment
	err = setup(configuration)
	if err != nil {
		l.Alert("Can not setup environment ", errgo.Details(err))
		os.Exit(1)
	}

	// Start profiling
	if configuration.Profile {
		profilebind := configuration.ProfileBind

		l.Info("Starting profiling on ", profilebind)
		go func() { l.Notice(http.ListenAndServe(profilebind, nil)) }()
	}
}

func main() {
	l := logger.New(name, "main")
	l.Notice("Starting")
	l.Info("Version: ", buildVersion)
	l.Info("Buildtime: ", buildTime)
	defer l.Notice("Finished")

	l.Debug("Configuration: ", fmt.Sprintf("%+v", configuration))

	watch()

	// Launch
	err := launch(configuration)
	if err != nil {
		l.Alert("Problem while launching: ", errgo.Details(err))
		os.Exit(1)
	}
}

func watch() {
	service.WatchInt(name+".runtime.goroutines", runtime.NumGoroutine)
	service.WatchUint(name+".service.count", service.Count)
	service.WatchRuntimeMemory(name + ".runtime.memory")
}
