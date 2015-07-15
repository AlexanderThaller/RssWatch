// The MIT License (MIT)
//
// Copyright (c) 2015 Alexander Thaller
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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

	"github.com/AlexanderThaller/rss"
	"github.com/AlexanderThaller/rsswatch/src/mailer"
	log "github.com/Sirupsen/logrus"
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
	mails   chan<- mailer.Mail
}

func (feed Feed) Launch(conf *Config, mails chan<- mailer.Mail) error {
	feed.config = conf
	feed.mails = mails

	if len(feed.Filters) == 0 {
		log.Debug("No filters specified will use default '.*'")
		feed.Filters = []string{`.*`}
	}

	log.Debug("Setting up filters")
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
	log.Debug("Will try to get feed")
	starttime := time.Now()
	err := feed.Get(feed.config)
	duration := time.Since(starttime)
	if err != nil {
		log.Error("can not get feed data: ", errgo.Details(err))
		return
	}
	log.Debug("Got feed")

	fetchDuration.WithLabelValues(feed.FormatTitle()).Observe(float64(duration.Nanoseconds() / 1000))

	var errcount uint
	for {
		{
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			d := 1*time.Second + time.Duration(r.Intn(9000))*time.Millisecond
			log.Debug("Sleep for ", d)
			time.Sleep(d)
		}
		refresh := feed.data.Refresh

		d := refresh.Sub(time.Now())
		log.Debug("Sleep for ", d, " (Until ", refresh, ")")
		time.Sleep(d)

		items := make(map[string]struct{})
		for item := range feed.data.ItemMap {
			items[item] = struct{}{}
		}

		log.Debug("Try to update feed")

		starttime := time.Now()
		err := feed.data.Update()
		duration := time.Since(starttime)

		if err != nil {
			log.Warning("Can not update feed: ", errgo.Details(err))
			feed.data.Refresh = time.Now().Add(1 * time.Minute)
			errcount += 1

			if errcount == 10 {
				log.Error("To much errors for this feed. Will now disable feed")
				return
			}

			continue
		}
		errcount = 0

		fetchDuration.WithLabelValues(feed.FormatTitle()).Observe(float64(duration.Nanoseconds() / 1000))

		log.Debug("Checking for new items")
		feed.Check(items)

		log.Debug("Updated feed will now try to save")
		err = feed.Save(feed.config.DataFolder)
		if err != nil {
			log.Error("Problem while saving: ", errgo.Details(err))
			return
		}
	}
}

func (feed *Feed) Send(item *rss.Item) {
	filtered := feed.Filter(item)
	for _, item := range filtered {
		message, err := feed.GenerateMessage(item)
		if err != nil {
			log.Warning("Can not generate message: ", err)
			continue
		}

		log.Debug("Sending email for filter ", item.Filter)
		feed.mails <- mailer.Mail{
			Sender:      feed.config.MailSender,
			Destination: feed.config.MailDestination,
			Message:     message.String(),
		}

		log.Debug("Sent mail")
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
	buffer := bytes.NewBufferString("")

	ftitle := feed.FormatTitle()

	ititle := strings.TrimSpace(item.data.Title)
	ititle = strings.Replace(ititle, "\n", " ", -1)

	sender := feed.config.MailSender

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
	log.Debug("Checking filter for ", item.Title)

	var out []*Item
	for filter, compiled := range feed.filters {
		log.Debug("Checking filter: ", filter)

		matches := compiled.MatchString(item.Title)
		if !matches {
			log.Debug("Item does not match")
			continue
		}
		log.Debug("Item matches filter adding to output")

		newitem := Item{
			Filter: filter,
			data:   item,
		}

		out = append(out, &newitem)
	}

	return out
}

func (feed *Feed) Check(items map[string]struct{}) {
	newitems := feed.data.Items
	for _, item := range newitems {
		_, exists := items[item.ID]

		if !exists {
			feed.Send(item)
		}
	}
}

func (feed *Feed) Get(conf *Config) error {
	if conf.SaveFeeds {
		log.Debug("Will try to restore feed")

		err := feed.Restore(conf.DataFolder)
		if err == nil {
			log.Debug("Restored feed. Will return feed")
			return nil
		}

		log.Debug("Can not restore feed")
		if !os.IsNotExist(err) {
			log.Debug("Error is not a not exists error we will return this")
			return err
		}
	}

	log.Debug("Will try to fetch feed")
	data, err := rss.Fetch(feed.Url)
	if err != nil {
		return err
	}
	log.Debug("Fetched feed")
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
	filename := feed.Filename(datafolder) + ".msgpack"

	log.Debug("Check if file exists")
	_, err := os.Stat(filename)
	if err != nil {
		return err
	}
	log.Debug("File does exist")

	log.Debug("Read from file")
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	log.Debug("Finished reading file")

	var data rss.Feed
	log.Debug("Unmarshal bytes from file")
	err = msgpack.Unmarshal(bytes, &data)
	if err != nil {
		return err
	}
	log.Debug("Finished unmarshaling")
	log.Debug("Finished restoring")
	feed.data = &data

	return nil
}

func (feed *Feed) Save(datafolder string) error {
	log.Debug("Getting filename for file")
	filename := feed.Filename(datafolder) + ".msgpack"

	log.Debug("Creating folder for file")
	err := os.MkdirAll(datafolder, 0755)
	if err != nil {
		return err
	}
	log.Debug("Created folder for file")

	log.Debug("Marshaling feed data to msgpack")
	bytes, err := msgpack.Marshal(feed.data)
	if err != nil {
		return err
	}
	log.Debug("Finished marshalling")

	log.Debug("Will now try to save the marshaled feed to the file")
	err = ioutil.WriteFile(filename, bytes, 0644)
	if err != nil {
		return err
	}
	log.Debug("Finished saving to file")

	log.Debug("Finished saving the feed")
	return nil
}

func (feed *Feed) Filename(datafolder string) string {
	saveurl := strings.Replace(feed.Url, "/", "_", -1)
	filename := filepath.Join(datafolder, saveurl)

	return filename
}
