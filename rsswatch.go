package main

import (
	"github.com/AlexanderThaller/logger"
	"github.com/AlexanderThaller/misc"
	"github.com/SlyMarbo/rss"
	"github.com/agl/xmpp"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	name = "RssWatch"
)

func init() {
	logger.SetLevel(name+".watch.filter", logger.Debug)
}

func main() {
	l := logger.New(name + ".main")
	l.Notice("Starting")
	defer l.Notice("Finished")

	x := "admin@ejabberd"

	u := []string{
		`http://www.reddit.com/.rss`,
		`http://www.rottentomatoes.com/syndication/rss/opening.xml`,
	}

	f := []string{
		`.*cracked.*`,
		`.*was.*`,
		`.*[F|f]ear.*`,
	}

	i := make(chan *rss.Item, 50000)
	o := make(chan *rss.Item, 50000)

	e := watchFeeds(u, i)
	if e != nil {
		l.Alert("Probem when parsing feed: ", e)
		os.Exit(1)
	}

	e = watchXMPP(x, o)
	if e != nil {
		l.Alert("Problem when connecting to xmpp: ", e)
		os.Exit(1)
	}

	e = watchFilters(f, i, o)
	if e != nil {
		l.Alert("Problem when starting filters: ", e)
		os.Exit(1)
	}

	misc.WaitForSigint()

	os.Exit(0)
}

func watchFeeds(fe []string, ch chan<- *rss.Item) (err error) {
	l := logger.New(name + ".watch.feeds")
	l.Info("Starting")
	defer l.Info("Finished")

	for _, d := range fe {
		u, e := url.Parse(d)
		if e != nil {
			l.Warning("Can not parse feed url ", u, ": ", e)
			continue
		}

		e = watchFeed(u, ch)
		if e != nil {
			l.Warning("Can not watch feed ", u, ": ", e)
			continue
		}

		l.Notice("Watching " + d)
	}

	return
}

func watchFeed(ur *url.URL, ch chan<- *rss.Item) (err error) {
	l := logger.New(name + ".watch.feed." + ur.Host + ur.Path)
	l.Info("Starting")
	defer l.Info("Finished")

	f, err := rss.Fetch(ur.String())
	if err != nil {
		return
	}

	m := make(map[string]struct{})

	l.Info("Run loop")
	go func() {
		for {
			l.Debug("Items: ", f.Items)

			for _, d := range f.Items {
				if _, t := m[d.Link]; t == false {
					l.Debug("New item: ", strings.TrimSpace(d.Title))
					m[d.Link] = struct{}{}
					ch <- d
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

func watchXMPP(id string, ch <-chan *rss.Item) (err error) {
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

	go func() {
		for {
			i := <-ch
			o.Send(id, strings.TrimSpace(i.Title))
		}
	}()

	return
}

func watchFilters(fi []string, in <-chan *rss.Item, ou chan<- *rss.Item) (err error) {
	l := logger.New(name + ".watch.filters")
	l.Info("Starting")
	defer l.Info("Finished")

	var c []chan<- *rss.Item

	for _, d := range fi {
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

			if r.Match([]byte(i.Title)) {
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
