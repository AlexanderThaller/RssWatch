package main

import (
	"io/ioutil"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/AlexanderThaller/logger"
	"github.com/SlyMarbo/rss"
	"github.com/juju/errgo"
	"github.com/vmihailenco/msgpack"
)

type Feeds struct {
	OutputFolder string
	feeds        map[string]Feed
	filters      *Filters
}

func NewFeeds() *Feeds {
	feeds := new(Feeds)
	feeds.feeds = make(map[string]Feed)
	feeds.filters = NewFilters()

	return feeds
}

func (feeds *Feeds) Start(feed []Feed, output chan<- *Item) error {
	for _, d := range feed {
		if feeds.OutputFolder != "" {
			d.AutoSave = genFeedPath(feeds.OutputFolder, &d)

			err := d.Repopulate(d.AutoSave)
			if err != nil {
				return err
			}
		}

		d.filters = feeds.filters
		d.Watch(output)
	}

	return nil
}

type Feed struct {
	AutoSave   string
	Url        string
	Filters    []string
	Folder     string
	output     chan<- *Item
	feed       *rss.Feed
	items      map[string]struct{}
	itemsMutex *sync.RWMutex
	filters    *Filters
}

func (feed *Feed) Save(path string) error {
	data, err := msgpack.Marshal(feed.items)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (feed *Feed) Repopulate(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		err = feed.Save(path)
		if err != nil {
			return err
		}

		return nil
	}

	err = msgpack.Unmarshal(data, &feed.items)
	if err != nil {
		return err
	}

	return nil
}

func (feed *Feed) Watch(output chan<- *Item) error {
	l := logger.New(name, "feed", "Watch", feed.Url)
	feed.output = output

	err := feed.Fetch()
	if err != nil {
		return err
	}

	go func() {
		for {
			l.Info("Sleep until ", feed.feed.Refresh, "(", feed.feed.Refresh.Sub(time.Now()), ")")
			sleepToTime(feed.feed.Refresh)

			l.Info("Updating: ", feed.Url)
			feed.feed.Update()

			for _, item := range feed.feed.Items {
				feed.processItem(item)
			}

			if feed.AutoSave != "" {
				err = feed.Save(feed.AutoSave)
				if err != nil {
					l.Error("Can not save feed: ", err)
					break
				}
			}
		}
	}()

	return nil
}

func (feed *Feed) itemKnown(item *rss.Item) bool {
	link := strings.TrimSpace(item.Link)

	if feed.items == nil {
		feed.items = make(map[string]struct{})
	}

	if feed.itemsMutex == nil {
		feed.itemsMutex = new(sync.RWMutex)
	}

	feed.itemsMutex.RLock()
	_, exists := feed.items[link]
	feed.itemsMutex.RUnlock()
	if exists {
		return true
	}

	feed.itemsMutex.Lock()
	feed.items[link] = struct{}{}
	feed.itemsMutex.Unlock()

	return false
}

func (feed *Feed) Fetch() error {
	url, err := url.Parse(feed.Url)
	if err != nil {
		return err
	}

	feed.feed, err = rss.Fetch(url.String())
	if err != nil {
		return err
	}

	for _, item := range feed.feed.Items {
		feed.processItem(item)
	}

	if feed.AutoSave != "" {
		err = feed.Save(feed.AutoSave)
		if err != nil {
			return err
		}
	}

	return nil
}

func (feed *Feed) processItem(item *rss.Item) {
	l := logger.New(name, "feed", "processItem")
	if feed.itemKnown(item) {
		return
	}

	for _, d := range feed.Filters {
		match, err := feed.filters.Match(item, d)
		if err != nil {
			l.Warning("Problem while filtering: ", errgo.Details(err))
			continue
		}

		if !match {
			continue
		}

		feed.output <- &Item{
			Feed:   feed.feed,
			Item:   item,
			Folder: feed.Folder,
			Filter: d,
		}
	}
}

func sleepToTime(ti time.Time) {
	for {
		time.Sleep(100 * time.Millisecond)
		if ti.Before(time.Now()) {
			break
		}
	}
}

func genFeedPath(folder string, feed *Feed) string {
	feedName := strings.Replace(feed.Url, "/", "_", -1)
	return path.Join(folder, feedName)
}
