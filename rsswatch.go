package main

import (
	"encoding/json"
	"github.com/AlexanderThaller/logger"
	"github.com/AlexanderThaller/misc"
	"github.com/SlyMarbo/rss"
	"github.com/agl/xmpp"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	name = "RssWatch"
)

func init() {
	logger.SetLevel(name+".load.feed", logger.Debug)
	logger.SetLevel(name+".write.feed", logger.Debug)
}

func main() {
	l := logger.New(name + ".main")
	l.Notice("Starting")
	defer l.Notice("Finished")

	u := []string{
		`http://www.reddit.com/.rss`,
		`http://www.rottentomatoes.com/syndication/rss/opening.xml`,
	}

	f := []string{
		`.*cracked.*`,
		`.*was.*`,
		`.*[F|f]ear.*`,
	}

	c := Config{
		DataFolder:      "feeds",
		SaveFeeds:       true,
		XmppDestination: "admin@ejabberd",
		Feeds:           u,
		Filters:         f,
	}

	i := make(chan *rss.Item, 50000)
	o := make(chan *rss.Item, 50000)

	e := watchFeeds(i, &c)
	if e != nil {
		l.Alert("Probem when parsing feed: ", e)
		os.Exit(1)
	}

	e = watchXMPP(o, &c)
	if e != nil {
		l.Alert("Problem when connecting to xmpp: ", e)
		os.Exit(1)
	}

	e = watchFilters(i, o, &c)
	if e != nil {
		l.Alert("Problem when starting filters: ", e)
		os.Exit(1)
	}

	misc.WaitForSigint()

	os.Exit(0)
}

func watchFeeds(ch chan<- *rss.Item, co *Config) (err error) {
	l := logger.New(name + ".watch.feeds")
	l.Info("Starting")
	defer l.Info("Finished")

	l.Info("Creating data folder " + co.DataFolder)
	err = os.MkdirAll(co.DataFolder, 0755)
	if err != nil {
		return
	}

	for _, d := range co.Feeds {
		u, e := url.Parse(d)
		if e != nil {
			l.Warning("Can not parse feed url ", u, ": ", e)
			continue
		}

		e = watchFeed(u, ch, co)
		if e != nil {
			l.Warning("Can not watch feed ", u, ": ", e)
			continue
		}

		l.Notice("Watching " + d)
	}

	return
}

func watchFeed(ur *url.URL, ch chan<- *rss.Item, co *Config) (err error) {
	l := logger.New(name + ".watch.feed." + ur.Host + ur.Path)
	l.Info("Starting")
	defer l.Info("Finished")

	f, err := rss.Fetch(ur.String())
	if err != nil {
		return
	}

	m, err := loadFeed(ur, co)
	if err != nil {
		return
	}

	l.Info("Run loop")
	go func() {
		for {
			l.Debug("Items: ", f.Items)

			for _, d := range f.Items {
				if _, t := m[d.Link]; t == true {
					continue
				}

				l.Debug("New item: ", strings.TrimSpace(d.Title))
				m[d.Link] = struct{}{}
				ch <- d

				e := writeFeed(m, ur, co)
				if e != nil {
					l.Warning("Can not write feed data: ", e)
				}
			}

			l.Debug("Sleep until ", f.Refresh, "(", f.Refresh.Sub(time.Now()), ")")
			sleepToTime(f.Refresh)
			time.Sleep(10 * time.Second)

			l.Debug("Updating")
			f.Update()
		}
	}()

	return
}

func loadFeed(ur *url.URL, co *Config) (itm map[string]struct{}, err error) {
	l := logger.New(name + ".load.feed." + ur.Host + ur.Path)
	l.Info("Starting")
	defer l.Info("Finished")

	itm = make(map[string]struct{})
	if !co.SaveFeeds {
		return
	}

	p := genFeedPath(ur, co)
	l.Debug("Path: ", p)

	i, err := ioutil.ReadFile(p)
	if err == nil {
		err = json.Unmarshal(i, &itm)
		return
	}
	err = nil

	return
}

func genFeedPath(ur *url.URL, co *Config) string {
	a := strings.Replace(ur.Host+ur.Path, "/", "_", -1)
	p := path.Clean(co.DataFolder + "/" + a)

	return p
}

func writeFeed(ma map[string]struct{}, ur *url.URL, co *Config) (err error) {
	l := logger.New(name + ".load.feed." + ur.Host + ur.Path)
	l.Info("Starting")
	defer l.Info("Finished")

	if !co.SaveFeeds {
		return
	}

	p := genFeedPath(ur, co)
	l.Debug("Path: ", p)

	o, err := json.Marshal(ma)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(p, o, 0644)
	if err != nil {
		return
	}

	return
}

func watchXMPP(ch <-chan *rss.Item, co *Config) (err error) {
	l := logger.New(name + ".watch.xmpp")
	l.Info("Starting")
	defer l.Info("Finished")

	// username
	u := "test"

	// domain
	d := "ejabberd"

	// password
	p := "test"

	h := d + ":5222"

	c := xmpp.Config{
		Conn:                    nil,
		InLog:                   nil,
		OutLog:                  nil,
		Log:                     nil,
		Create:                  false,
		TrustedAddress:          true,
		Archive:                 false,
		ServerCertificateSHA256: []byte(""),
		SkipTLS:                 true,
	}

	o, err := xmpp.Dial(h, u, d, p, &c)
	if err != nil {
		return
	}

	// Let server know that we are online.
	o.SignalPresence("")

	go func() {
		for {
			i := <-ch
			o.Send(co.XmppDestination, strings.TrimSpace(i.Title))
		}
	}()

	return
}

func watchFilters(in <-chan *rss.Item, ou chan<- *rss.Item, co *Config) (err error) {
	l := logger.New(name + ".watch.filters")
	l.Info("Starting")
	defer l.Info("Finished")

	var c []chan<- *rss.Item

	for _, d := range co.Filters {
		f, e := watchFilter(d, ou)

		if e != nil {
			l.Warning("Can not use filter {{ ", d, " }}: ", e)
			continue
		}

		c = append(c, f)
	}

	l.Info("Run loop")
	go func() {
		for {
			i := <-in

			l.Debug("New item: ", i.Title)
			for _, d := range c {
				d <- i
			}
		}
	}()

	return
}

func watchFilter(fi string, ou chan<- *rss.Item) (cin chan<- *rss.Item, err error) {
	l := logger.New(name + ".watch.filter.{{ " + fi + " }}")
	l.Info("Starting")
	defer l.Info("Finished")

	r, err := regexp.Compile(fi)
	if err != nil {
		return
	}
	c := make(chan *rss.Item, 50000)

	l.Info("Run loop")
	go func() {
		for {
			i := <-c

			s := r.Match([]byte(i.Title)) || r.Match([]byte(i.Content))
			if s {
				l.Debug("Matched item: ", i.Title)
				ou <- i
			}
		}
	}()

	cin = c

	return
}

func sleepToTime(ti time.Time) {
	for {
		time.Sleep(100 * time.Millisecond)
		if ti.Before(time.Now()) {
			break
		}
	}
}
