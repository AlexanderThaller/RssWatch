package main

import (
	"bytes"
	"net/smtp"
	"time"

	"github.com/AlexanderThaller/logger"
	"github.com/AlexanderThaller/service"
)

func launch(conf *Config) error {
	l := logger.New(name, "launch")

	mails, err := launchMails(conf)
	if err != nil {
		return err
	}

	for _, d := range conf.Feeds {
		err := d.Launch(conf, mails)
		if err != nil {
			return err
		}
	}

	l.Trace("Watching for signals")
	service.WatchSignals()
	return nil
}

func launchMails(conf *Config) (chan<- *bytes.Buffer, error) {
	l := logger.New(name, "launch", "Mails")
	mails := make(chan *bytes.Buffer, 50000)

	go func() {
		for {
			message := <-mails

			l.Debug("Sending email")
			err := sendMail(message, conf)
			if err != nil {
				l.Error("Problem while sending email: ", err)
				time.Sleep(2 * time.Second)

				continue
			}
		}
	}()

	return mails, nil
}

func sendMail(message *bytes.Buffer, conf *Config) error {
	conn, err := smtp.Dial(conf.MailServer)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.Mail(conf.MailSender)
	conn.Rcpt(conf.MailDestination)

	wc, err := conn.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	_, err = message.WriteTo(wc)
	if err != nil {
		return err
	}

	return nil
}
