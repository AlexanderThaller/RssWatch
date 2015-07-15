package main

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/AlexanderThaller/logger"
	"github.com/AlexanderThaller/rss"
	"github.com/juju/errgo"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

var (
	fetchDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "rsswatch",
		Subsystem: "feed",
		Help:      "The duration it took to update the feed",
		Name:      "fetch_duration_milliseconds",
	},
		[]string{
			"feed",
		})
)

func init() {
	prometheus.MustRegister(fetchDuration)
}

type Feed struct {
	Url     string
	Filters []string
	Folder  string
	filters map[string]*regexp.Regexp
	data    *rss.Feed
	config  *Config
	mails   chan<- *bytes.Buffer
}

func (feed Feed) Launch(conf *Config, mails chan<- *bytes.Buffer) error {
	l := logger.New(name, "Feed", "Launch", feed.Url)
	l.Info("Starting")

	feed.config = conf
	feed.mails = mails

	if len(feed.Filters) == 0 {
		l.Debug("No filters specified will use default '.*'")
		feed.Filters = []string{`.*`}
	}

	l.Debug("Setting up filters")
	l.Trace("Filters: ", feed.Filters)
	feed.filters = make(map[string]*regexp.Regexp)
	for _, filter := range feed.Filters {
		compiled, err := regexp.Compile(filter)
		if err != nil {
			return err
		}

		feed.filters[filter] = compiled
	}

	go feed.Watch()

	return nil
}

func (feed Feed) Watch() {
	l := logger.New(name, "Feed", "Watch", feed.Url)

	l.Debug("Will try to get feed")
	starttime := time.Now()
	err := feed.Get(feed.config)
	duration := time.Since(starttime)
	if err != nil {
		l.Error("can not get feed data: ", errgo.Details(err))
		return
	}
	l.Debug("Got feed")

	fetchDuration.WithLabelValues(feed.FormatTitle()).Observe(float64(duration.Nanoseconds() / 1000))

	var errcount uint
	for {
		{
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			d := 1*time.Second + time.Duration(r.Intn(9000))*time.Millisecond
			l.Debug("Sleep for ", d)
			time.Sleep(d)
		}
		refresh := feed.data.Refresh

		d := refresh.Sub(time.Now())
		l.Debug("Sleep for ", d, " (Until ", refresh, ")")
		time.Sleep(d)

		items := make(map[string]struct{})
		for item := range feed.data.ItemMap {
			items[item] = struct{}{}
		}

		l.Trace("Items length: ", len(items))
		l.Debug("Try to update feed")

		starttime := time.Now()
		err := feed.data.Update()
		duration := time.Since(starttime)

		if err != nil {
			l.Warning("Can not update feed: ", errgo.Details(err))
			feed.data.Refresh = time.Now().Add(1 * time.Minute)
			errcount += 1

			if errcount == 10 {
				l.Error("To much errors for this feed. Will now disable feed")
				return
			}

			continue
		}
		errcount = 0

		fetchDuration.WithLabelValues(feed.FormatTitle()).Observe(float64(duration.Nanoseconds() / 1000))

		l.Debug("Checking for new items")
		feed.Check(items)

		l.Debug("Updated feed will now try to save")
		err = feed.Save(feed.config.DataFolder)
		if err != nil {
			l.Error("Problem while saving: ", errgo.Details(err))
			return
		}
	}
}

func (feed *Feed) Send(item *rss.Item) {
	l := logger.New(name, "Feed", "Send", feed.Url, item.ID)
	l.Trace("Sending item: ", item)

	filtered := feed.Filter(item)
	for _, item := range filtered {
		message, err := feed.GenerateMessage(item)
		if err != nil {
			l.Warning("Can not generate message: ", err)
			continue
		}
		l.Trace("Message: ", message.String())

		l.Debug("Sending email for filter ", item.Filter)
		feed.mails <- message
		l.Debug("Sent mail")
	}
}

func (feed *Feed) FormatTitle() string {
	ftitle := strings.TrimSpace(feed.data.Title)
	ftitle = strings.Replace(ftitle, ".", "_", -1)
	ftitle = strings.Replace(ftitle, "/", "_", -1)
	ftitle = strings.TrimSpace(ftitle)
	ftitle = strings.Replace(ftitle, "\n", " ", -1)

	return ftitle
}

func (feed *Feed) GenerateMessage(item *Item) (*bytes.Buffer, error) {
	l := logger.New(name, "Feed", "Generate", "Message", item.data.ID)

	buffer := bytes.NewBufferString("")

	ftitle := feed.FormatTitle()

	ititle := strings.TrimSpace(item.data.Title)
	ititle = strings.Replace(ititle, "\n", " ", -1)

	sender := feed.config.MailSender

	l.Trace("From: ", sender)
	l.Trace("Subject: ", ititle)

	buffer.WriteString("From: " + sender + "\n")
	buffer.WriteString("Subject: " + ititle + "\n")
	buffer.WriteString("Content-Type: text/html; charset=utf-8\n")
	buffer.WriteString("Feed: " + ftitle + "\n")
	buffer.WriteString("Folder: " + feed.Folder + "\n")

	ifilter := strings.Replace(item.Filter, ".", `_`, -1)
	if ifilter != "_*" {
		buffer.WriteString("Filter: " + ifilter + "\n")
	}

	buffer.WriteString("\n\n")

	buffer.WriteString(ftitle + " - " + ititle + "<br>\n")
	buffer.WriteString(item.data.Content)

	buffer.WriteString("<br>\n")
	buffer.WriteString(`<a href="` + item.data.Link + `">Link</a>`)

	return buffer, nil
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
			data:   item,
		}

		l.Trace("Newitem: ", newitem)
		out = append(out, &newitem)
	}

	l.Trace(out)
	return out
}

func (feed *Feed) Check(items map[string]struct{}) {
	l := logger.New(name, "Feed", "Check", feed.Url)

	newitems := feed.data.Items
	l.Trace("Items: ", items)
	for _, item := range newitems {
		l.Trace("Item id: ", item.ID)
		_, exists := items[item.ID]

		l.Trace("Exists: ", exists)
		if !exists {
			l.Trace("New item: ", item)
			feed.Send(item)
		}
	}
}

func (feed *Feed) Get(conf *Config) error {
	l := logger.New(name, "Feed", "Get", feed.Url)

	if conf.SaveFeeds {
		l.Debug("Will try to restore feed")

		err := feed.Restore(conf.DataFolder)
		if err == nil {
			l.Debug("Restored feed. Will return feed")
			return nil
		}

		l.Debug("Can not restore feed")
		if !os.IsNotExist(err) {
			l.Debug("Error is not a not exists error we will return this")
			return err
		}

		l.Trace("Error while restoring: ", err)
	}

	l.Debug("Will try to fetch feed")
	data, err := rss.Fetch(feed.Url)
	if err != nil {
		return err
	}
	l.Debug("Fetched feed")
	feed.data = data

	for _, item := range data.Items {
		feed.Send(item)
	}

	if conf.SaveFeeds {
		err = feed.Save(conf.DataFolder)
		if err != nil {
			return err
		}
	}

	return err
}

func (feed *Feed) Restore(datafolder string) error {
	l := logger.New(name, "Feed", "Restore", feed.Url)

	filename := feed.Filename(datafolder) + ".msgpack"
	l.Trace("Filename: ", filename)

	l.Debug("Check if file exists")
	_, err := os.Stat(filename)
	if err != nil {
		return err
	}
	l.Debug("File does exist")

	l.Debug("Read from file")
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	l.Debug("Finished reading file")

	var data rss.Feed
	l.Debug("Unmarshal bytes from file")
	err = msgpack.Unmarshal(bytes, &data)
	if err != nil {
		return err
	}
	l.Debug("Finished unmarshaling")
	l.Debug("Finished restoring")
	l.Trace("Data: ", data)
	feed.data = &data

	return nil
}

func (feed *Feed) Save(datafolder string) error {
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
	bytes, err := msgpack.Marshal(feed.data)
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
