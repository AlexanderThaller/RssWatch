package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/smtp"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AlexanderThaller/logger"
	"github.com/AlexanderThaller/misc"
	"github.com/SlyMarbo/rss"
	"github.com/agl/xmpp"
)

const (
	name = "RssWatch"
)

var (
	configpath = flag.String("c", name+".cnf", "The path to the config file.")
)

func init() {
	logger.SetLevel(".", logger.Info)
	logger.SetLevel(name+".watch.mail", logger.Trace)

	flag.Parse()
}

func main() {
	l := logger.New(name + ".main")
	l.Notice("Starting")
	defer l.Notice("Finished")

	c, e := configure(*configpath)
	if e != nil {
		l.Alert("Problem when configuring: ", e)
		os.Exit(1)
	}

	i := make(chan *Item, 50000)
	o := make(chan *Item, 50000)

	e = watchFeeds(i, c)
	if e != nil {
		l.Alert("Probem when parsing feed: ", e)
		os.Exit(1)
	}

	if !c.XmppDisable {
		e = watchXMPP(o, c)
		if e != nil {
			l.Alert("Problem when connecting to xmpp: ", e)
			os.Exit(1)
		}
	}

	if !c.MailDisable {
		e = watchMail(o, c)
		if e != nil {
			l.Alert("Problem when watching mail: ", e)
			os.Exit(1)
		}
	}

	misc.WaitForSigint()

	os.Exit(0)
}

type Item struct {
	Feed *rss.Feed
	Item *rss.Item
}

func configure(pa string) (con *config, err error) {
	l := logger.New(name + ".configure")
	l.Info("Starting")
	defer l.Info("Finished")

	l.Info(`Loading config from path "`, pa, `"`)
	i, err := ioutil.ReadFile(pa)
	if err == nil {
		l.Debug(`Using file "`, pa, `" for config`)

		var b bytes.Buffer
		json.Compact(&b, i)

		err = json.Unmarshal(b.Bytes(), &con)
		return
	}

	c := new(config)
	c.Default()
	con = c

	o, _ := json.MarshalIndent(con, "", "  ")

	err = ioutil.WriteFile(pa, o, 0600)

	return
}

func watchFeeds(ch chan<- *Item, co *config) (err error) {
	l := logger.New(name + ".watch.feeds")
	l.Info("Starting")
	defer l.Info("Finished")

	l.Info("Creating data folder " + co.DataFolder)
	err = os.MkdirAll(co.DataFolder, 0755)
	if err != nil {
		return
	}

	for _, d := range co.Feeds {
		e := watchFeed(d, ch, co)
		if e != nil {
			l.Warning("Can not watch feed ", d, ": ", e)
			continue
		}

		l.Notice("Watching " + d.Url)
	}

	return
}

func watchFeed(feed Feed, ch chan<- *Item, co *config) (err error) {
	l := logger.New(name + ".watch.feed." + feed.Url)
	l.Info("Starting")
	defer l.Info("Finished")

	ur, err := url.Parse(feed.Url)
	if err != nil {
		return
	}

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
				ch <- &Item{
					Feed: f,
					Item: d,
				}

				e := writeFeed(m, ur, co)
				if e != nil {
					l.Warning("Can not write feed data: ", e)
				}
			}

			l.Info("Sleep until ", f.Refresh, "(", f.Refresh.Sub(time.Now()), ")")
			sleepToTime(f.Refresh)
			time.Sleep(10 * time.Second)

			l.Info("Updating")
			f.Update()
		}
	}()

	return
}

func loadFeed(ur *url.URL, co *config) (itm map[string]struct{}, err error) {
	l := logger.New(name + ".load.feed." + ur.Host + ur.Path)
	l.Debug("Starting")
	defer l.Debug("Finished")

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

func genFeedPath(ur *url.URL, co *config) string {
	a := strings.Replace(ur.Host+ur.Path, "/", "_", -1)
	p := path.Clean(co.DataFolder + "/" + a)

	return p
}

func writeFeed(ma map[string]struct{}, ur *url.URL, co *config) (err error) {
	l := logger.New(name + ".write.feed." + ur.Host + ur.Path)
	l.Debug("Starting")
	defer l.Debug("Finished")

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

func watchXMPP(ch <-chan *Item, co *config) (err error) {
	l := logger.New(name + ".watch.xmpp")
	l.Info("Starting")
	defer l.Info("Finished")

	// username
	u := co.XmppUsername

	// domain
	d := co.XmppDomain

	// password
	p := co.XmppPassword

	h := d + ":" + strconv.FormatUint(uint64(co.XmppPort), 10)

	c := xmpp.Config{
		Conn:                    nil,
		InLog:                   nil,
		OutLog:                  nil,
		Log:                     nil,
		Create:                  false,
		TrustedAddress:          true,
		Archive:                 false,
		ServerCertificateSHA256: []byte(""),
		SkipTLS:                 co.XmppSkipTLS,
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
			item := i.Item

			m := strings.TrimSpace(item.Title)
			m += " - "

			a := strings.TrimSpace(item.Content)
			if len(a) > 150 {
				m += a[0:150] + "\n..."
			} else {
				m += a
			}
			m += "\n\n"
			m += item.Link

			o.Send(co.XmppDestination, m)
			time.Sleep(100 * time.Millisecond)
		}
	}()

	return
}

func watchMail(ch <-chan *Item, co *config) (err error) {
	l := logger.New(name + ".watch.mail")
	l.Info("Starting")
	defer l.Info("Finished")

	go func() {
		for {
			i := <-ch
			feed := i.Feed
			item := i.Item

			ftitle := strings.TrimSpace(feed.Title)
			ititle := strings.TrimSpace(item.Title)

			m := "From: rsswatch@thaller.ws\n"
			m += "Subject: " + ititle + "\n"
			m += "Content-Type: text/plain; charset=utf-8\n"
			m += "List-ID: " + ftitle + "\n"

			m += ftitle + " - " + ititle + "\n"
			m += strings.TrimSpace(item.Content)

			m += "\n\n"
			m += item.Link

			conn, err := smtp.Dial(co.MailServer)
			if err != nil {
				l.Warning("Can not connect to mailserver: ", err)
				continue
			}

			conn.Mail("rsswatch@thaller.ws")
			conn.Rcpt(co.MailDestination)

			wc, err := conn.Data()
			if err != nil {
				l.Warning("Problem when sending the header: ", err)
			}

			buf := bytes.NewBufferString(m)
			_, err = buf.WriteTo(wc)
			if err != nil {
				l.Warning("Problem when sending body: ", err)
				wc.Close()
			}
			wc.Close()

			time.Sleep(100 * time.Millisecond)
		}
	}()

	return
}

func watchFilters(in <-chan *Item, ou chan<- *Item, filters []string) (err error) {
	l := logger.New(name + ".watch.filters")
	l.Info("Starting")
	defer l.Info("Finished")

	var c []chan<- *Item

	for _, d := range filters {
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

			l.Debug("New item: ", i.Item.Title)
			for _, d := range c {
				d <- i
			}
		}
	}()

	return
}

func watchFilter(fi string, ou chan<- *Item) (cin chan<- *Item, err error) {
	l := logger.New(name + ".watch.filter.{{ " + fi + " }}")
	l.Info("Starting")
	defer l.Info("Finished")

	r, err := regexp.Compile(fi)
	if err != nil {
		return
	}
	c := make(chan *Item, 50000)

	l.Info("Run loop")
	go func() {
		for {
			i := <-c

			s := r.Match([]byte(i.Item.Title)) || r.Match([]byte(i.Item.Content))
			if s {
				l.Debug("Matched item: ", i.Item.Title)
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
