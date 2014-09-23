package main

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlexanderThaller/logger"
	"github.com/SlyMarbo/rss"
	"github.com/juju/errgo"
	"github.com/vmihailenco/msgpack"
)

type Feed struct {
	Url     string
	Filters []string
	Folder  string
}

func (feed Feed) Launch(conf *Config) {
	l := logger.New(name, "Feed", "Launch", feed.Url)
	l.Info("Starting")

	l.Debug("Will try to get feed")
	data, err := feed.Get(conf)
	if err != nil {
		l.Error("Problem when getting feed: ", errgo.Details(err))
		return
	}
	l.Debug("Got feed")
	l.Trace("Feed data: ", data)

	feed.Watch(data, conf)
}

func (feed *Feed) Watch(data *rss.Feed, conf *Config) {
	l := logger.New(name, "Feed", "Watch", feed.Url)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		d := time.Duration(r.Intn(60000)) * time.Millisecond
		l.Debug("Sleep for ", d)
		time.Sleep(d)

		l.Debug("Try to update feed")
		updated, err := feed.Update(data)
		if err != nil {
			l.Warning("Can not update feed: ", errgo.Details(err))
		}

		if updated {
			l.Debug("Updated feed will now try to save")
			err = feed.Save(data, conf.DataFolder)
			if err != nil {
				l.Error("Problem while saving: ", errgo.Details(err))
				return
			}
		} else {
			l.Debug("Not updated")
		}

		d = 5 * time.Minute
		l.Debug("Sleep for ", d)
		time.Sleep(d)
	}
}

func (feed *Feed) Filter(items []*rss.Item) []*Item {
	var out []*Item
	return out
}

func (feed *Feed) Check(newitems []*rss.Item, items map[string]struct{}) []*rss.Item {
	var out []*rss.Item

	for _, d := range newitems {
		if _, exists := items[d.ID]; exists {
			out = append(out, d)
		}
	}

	return out
}

func (feed *Feed) Get(conf *Config) (*rss.Feed, error) {
	l := logger.New(name, "Feed", "Get", feed.Url)

	if conf.SaveFeeds {
		l.Debug("Will try to restore feed")

		data, err := feed.Restore(conf.DataFolder)
		if err == nil {
			l.Debug("Restored feed. Will return feed")
			return data, nil
		}

		l.Debug("Can not restore feed")
		if !os.IsNotExist(err) {
			l.Debug("Error is not a not exists error we will return this")
			return nil, err
		}

		l.Trace("Error while restoring: ", err)
	}

	l.Debug("Will try to fetch feed")
	data, err := rss.Fetch(feed.Url)
	if err != nil {
		return nil, err
	}
	l.Debug("Fetched feed")

	if conf.SaveFeeds {
		err = feed.Save(data, conf.DataFolder)
		if err != nil {
			return data, err
		}
	}

	return data, err
}

func (feed *Feed) Restore(datafolder string) (*rss.Feed, error) {
	l := logger.New(name, "Feed", "Restore", feed.Url)

	filename := feed.Filename(datafolder) + ".msgpack"
	l.Trace("Filename: ", filename)

	l.Debug("Check if file exists")
	_, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	l.Debug("File does exist")

	l.Debug("Read from file")
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	l.Debug("Finished reading file")

	var data rss.Feed
	l.Debug("Unmarshal bytes from file")
	err = msgpack.Unmarshal(bytes, &data)
	if err != nil {
		return nil, err
	}
	l.Debug("Finished unmarshaling")
	l.Debug("Finished restoring")
	l.Trace("Data: ", data)
	return &data, nil
}

func (feed *Feed) Update(data *rss.Feed) (bool, error) {
	l := logger.New(name, "Feed", "Update", feed.Url)

	l.Trace("Refresh: ", data.Refresh)
	l.Trace("After: ", data.Refresh.After(time.Now()))
	if data.Refresh.After(time.Now()) {
		l.Debug("Its not time to update yet")
		return false, nil
	}
	l.Debug("Will update feed")

	err := data.Update()
	if err != nil {
		return false, err
	}
	l.Debug("Updated feed")
	l.Trace("New refresh: ", data.Refresh)

	return true, nil
}

func (feed *Feed) Save(data *rss.Feed, datafolder string) error {
	l := logger.New(name, "Feed", "Save", feed.Url)

	l.Debug("Will try to save feed")
	filename := feed.Filename(datafolder) + ".msgpack"

	err := os.MkdirAll(datafolder, 0755)
	if err != nil {
		return err
	}

	bytes, err := msgpack.Marshal(data)
	err = ioutil.WriteFile(filename, bytes, 0644)
	if err != nil {
		return err
	}

	l.Debug("Saved feed")
	return nil
}

func (feed *Feed) Filename(datafolder string) string {
	saveurl := strings.Replace(feed.Url, "/", "_", -1)
	filename := filepath.Join(datafolder, saveurl)

	return filename
}
