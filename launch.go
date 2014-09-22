package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexanderThaller/logger"
	"github.com/AlexanderThaller/service"
	"github.com/SlyMarbo/rss"
	"github.com/juju/errgo"
	"github.com/vmihailenco/msgpack"
)

func launch(conf *Config) (err error) {
	l := logger.New(name, "launch")

	for _, d := range conf.Feeds {
		go feedLaunch(d.Url, d.Folder, d.Filters, conf)
	}

	l.Trace("Watching for signals")
	service.WatchSignals()
	return nil
}

func feedLaunch(url, folder string, filters []string, conf *Config) {
	l := logger.New(name, "feed", "Launch", url)
	l.Info("Starting")

	feed, err := feedGet(url, conf)
	if err != nil {
		l.Error("Problem when getting feed: ", errgo.Details(err))
		return
	}
	l.Debug("Got feed")
	l.Trace("Feed data: ", feed)
}

func feedGet(url string, conf *Config) (*rss.Feed, error) {
	l := logger.New(name, "feed", "Get", url)

	if conf.SaveFeeds {
		l.Debug("Will try to restore feed")

		feed, err := feedRestore(conf.DataFolder, url)
		if err == nil {
			l.Debug("Restored feed. Will return feed")
			return feed, nil
		}

		l.Debug("Can not restore feed")
		if !os.IsNotExist(err) {
			l.Debug("Error is not a not exists error we will return this")
			return nil, err
		}

		l.Trace("Error while restoring: ", err)
	}

	l.Debug("Will try to fetch feed")
	feed, err := rss.Fetch(url)
	if err != nil {
		return feed, err
	}
	l.Debug("Fetched feed")

	if conf.SaveFeeds {
		l.Debug("Will try to save feed")
		err = feedSave(feed, conf.DataFolder, url)
		if err != nil {
			return feed, err
		}
		l.Debug("Saved feed")
	}

	return feed, err
}

func feedRestore(datafolder, url string) (*rss.Feed, error) {
	filename := feedFilename(datafolder, url) + ".msgpack"
	_, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var feed rss.Feed
	err = msgpack.Unmarshal(bytes, &feed)
	if err != nil {
		return nil, err
	}

	return &feed, nil
}

func feedSave(feed *rss.Feed, datafolder, url string) error {
	filename := feedFilename(datafolder, url) + ".msgpack"

	err := os.MkdirAll(datafolder, 0755)
	if err != nil {
		return err
	}

	bytes, err := msgpack.Marshal(feed)
	err = ioutil.WriteFile(filename, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func feedFilename(datafolder, url string) string {
	saveurl := strings.Replace(url, "/", "_", -1)
	filename := filepath.Join(datafolder, saveurl)

	return filename
}
