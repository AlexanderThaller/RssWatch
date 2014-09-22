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
		go launchFeed(d.Url, d.Folder, d.Filters, conf)
	}

	l.Trace("Watching for signals")
	service.WatchSignals()
	return nil
}

func launchFeed(url, folder string, filters []string, conf *Config) {
	l := logger.New(name, "launch", "Feed", url)
	l.Trace("Starting")

	feed, err := getFeed(url, conf)
	if err != nil {
		l.Error("Problem when getting feed: ", errgo.Details(err))
		return
	}
	l.Trace("Feed: ", feed)
}

func getFeed(url string, conf *Config) (*rss.Feed, error) {
	l := logger.New(name, "getFeed", url)

	if conf.SaveFeeds {
		l.Debug("Will try to restore feed")

		feed, err := restoreFeed(conf.DataFolder, url)
		if err == nil {
			l.Debug("Restored feed. Will return feed")
			return feed, nil
		}

		l.Debug("Can not restore feed")
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
		err = saveFeed(feed, conf.DataFolder, url)
		if err != nil {
			return feed, err
		}
		l.Debug("Saved feed")
	}

	return feed, err
}

func restoreFeed(datafolder, url string) (*rss.Feed, error) {
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

func saveFeed(feed *rss.Feed, datafolder, url string) error {
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
