package main

import (
	"bytes"
	"net/smtp"
	"os"
	"strings"

	"github.com/AlexanderThaller/logger"
	"github.com/AlexanderThaller/service"
)

// launch will start the appliaction based on the given configuration.
func launch(conf *Config) (err error) {
	l := logger.New(name, "launch")

	feedsItems := make(chan *Item, 50000)
	feeds := NewFeeds()
	if conf.SaveFeeds {
		err = os.MkdirAll(conf.DataFolder, 0755)
		if err != nil {
			return
		}

		feeds.OutputFolder = conf.DataFolder
	}

	err = feeds.Start(conf.Feeds, feedsItems)
	if err != nil {
		return
	}

	err = launchMail(feedsItems, conf)
	if err != nil {
		return
	}

	l.Trace("Watching for signals")
	service.WatchSignals()
	return nil
}

func launchMail(items <-chan *Item, conf *Config) error {
	l := logger.New(name, "launchMail")

	go func() {
		for {
			item := <-items

			l.Debug("Sending: ", item.String())
			process(item, conf)
			l.Trace("Sent: ", item.String())
		}
	}()

	return nil
}

func process(item *Item, conf *Config) {
	l := logger.New(name, "process")

	message, err := generateMessage(item, conf.MailSender)
	if err != nil {
		l.Warning("Can not generate message: ", err)
		return
	}

	conn, err := smtp.Dial(conf.MailServer)
	if err != nil {
		l.Warning("Can not connect to mailserver: ", err)
		return
	}
	defer conn.Close()

	conn.Mail(conf.MailSender)
	conn.Rcpt(conf.MailDestination)

	wc, err := conn.Data()
	if err != nil {
		l.Warning("Problem when sending the header: ", err)
		return
	}
	defer wc.Close()

	_, err = message.WriteTo(wc)
	if err != nil {
		l.Warning("Problem when sending body: ", err)
		return
	}
}

func generateMessage(item *Item, sender string) (*bytes.Buffer, error) {
	buffer := bytes.NewBufferString("")

	ftitle := strings.TrimSpace(item.Feed.Title)
	ititle := strings.TrimSpace(item.Item.Title)

	buffer.WriteString("From: " + sender + "\n")
	buffer.WriteString("Subject: " + ititle + "\n")
	buffer.WriteString("Content-Type: text/html; charset=utf-8\n")
	buffer.WriteString("Feed: " + ftitle + "\n")
	buffer.WriteString("Folder: " + item.Folder + "\n")

	ifilter := strings.Replace(item.Filter, ".", `_`, -1)
	if ifilter != "_*" {
		buffer.WriteString("Filter: " + ifilter + "\n")
	}

	buffer.WriteString("\n\n")

	buffer.WriteString(ftitle + " - " + ititle + "\n")
	buffer.WriteString(item.Item.Content)

	buffer.WriteString("\n\n")
	buffer.WriteString(item.Item.Link)

	return buffer, nil
}
