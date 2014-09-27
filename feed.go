package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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
	filters map[string]*regexp.Regexp
}

func (feed Feed) Launch(conf *Config) error {
	l := logger.New(name, "Feed", "Launch", feed.Url)
	l.Info("Starting")

	l.Debug("Setting up filters")
	feed.filters = make(map[string]*regexp.Regexp)
	for _, filter := range feed.Filters {
		compiled, err := regexp.Compile(filter)
		if err != nil {
			return err
		}

		feed.filters[filter] = compiled
	}

	l.Debug("Will try to get feed")
	data, err := feed.Get(conf)
	if err != nil {
		return err
	}
	l.Debug("Got feed")
	l.Trace("Feed data: ", data)

	go feed.Watch(data, conf)

	return nil
}

func (feed *Feed) Watch(data *rss.Feed, conf *Config) {
	l := logger.New(name, "Feed", "Watch", feed.Url)

	for {
		d := data.Refresh.Sub(time.Now())
		l.Debug("Sleep for ", d, " (Until ", data.Refresh, ")")
		time.Sleep(d)

		items := make(map[string]struct{})
		for item := range data.ItemMap {
			items[item] = struct{}{}
		}

		l.Trace("Items length: ", len(items))
		l.Debug("Try to update feed")
		updated, err := data.Update()
		if err != nil {
			l.Warning("Can not update feed: ", errgo.Details(err))
		}
		l.Trace("Data ItemMap length: ", len(data.ItemMap))

		if !updated {
			l.Debug("Not updated")
			continue
		}

		l.Debug("Checking for new items")
		feed.Check(data.Items, items)

		l.Debug("Updated feed will now try to save")
		err = feed.Save(data, conf.DataFolder)
		if err != nil {
			l.Error("Problem while saving: ", errgo.Details(err))
			return
		}
	}
}

func (feed *Feed) Send(item *rss.Item) {
	l := logger.New(name, "Feed", "Send", feed.Url)
	l.Trace("Sending item: ", item)

	filtered := feed.Filter(item)
	for _, item := range filtered {
		l.Debug("Sending filtered item: ", item.Data.ID)
		l.Trace("Sending filtered item: ", item)
	}
}

func (feed *Feed) Filter(item *rss.Item) []*Item {
	l := logger.New(name, "Feed", "Filter", feed.Url, item.ID)
	l.Trace("Item: ", item)
	l.Debug("Checking filter for ", item.Title)

	var out []*Item
	for filter, compiled := range feed.filters {
		l.Debug("Checking filter: ", filter)

		matches := compiled.MatchString(item.Title)
		if !matches {
			l.Debug("Item does not match")
			continue
		}
		l.Debug("Item matches filter adding to output")

		newitem := Item{
			Filter: filter,
			Data:   item,
		}

		l.Trace("Newitem: ", newitem)
		out = append(out, &newitem)
	}

	l.Trace(out)
	return out
}

func (feed *Feed) Check(newitems []*rss.Item, items map[string]struct{}) {
	l := logger.New(name, "Feed", "Check", feed.Url)

	l.Trace("Items: ", items)
	for _, item := range newitems {
		l.Trace("Item id: ", item.ID)
		_, exists := items[item.ID]

		l.Trace("Exists: ", exists)
		if !exists {
			l.Trace("New item: ", item)
			go feed.Send(item)
		}
	}
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

	for _, item := range data.Items {
		go feed.Send(item)
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

func (feed *Feed) Save(data *rss.Feed, datafolder string) error {
	l := logger.New(name, "Feed", "Save", feed.Url)

	l.Debug("Getting filename for file")
	filename := feed.Filename(datafolder) + ".msgpack"
	l.Trace("Filename: ", filename)

	l.Debug("Creating folder for file")
	err := os.MkdirAll(datafolder, 0755)
	if err != nil {
		return err
	}
	l.Debug("Created folder for file")

	l.Debug("Marshaling feed data to msgpack")
	bytes, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}
	l.Debug("Finished marshalling")

	l.Debug("Will now try to save the marshaled feed to the file")
	err = ioutil.WriteFile(filename, bytes, 0644)
	if err != nil {
		return err
	}
	l.Debug("Finished saving to file")

	l.Debug("Finished saving the feed")
	return nil
}

func (feed *Feed) Filename(datafolder string) string {
	saveurl := strings.Replace(feed.Url, "/", "_", -1)
	filename := filepath.Join(datafolder, saveurl)

	return filename
}
