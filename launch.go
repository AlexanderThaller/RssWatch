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
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
)

func launch(conf *Config, prometheusBinding string) error {
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

	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(prometheusBinding, nil)

	waitForStopSignal()
	return nil
}

func launchMails(conf *Config) (chan<- *bytes.Buffer, error) {
	mails := make(chan *bytes.Buffer, 50000)

	mailsQueue := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "rsswatch",
		Subsystem: "mails",
		Help:      "Current count of mails currently in the send queue.",
		Name:      "queue",
	})
	prometheus.MustRegister(mailsQueue)

	mailsSent := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "rsswatch",
		Subsystem: "mails",
		Help:      "The number of emails sent.",
		Name:      "sent",
	})
	prometheus.MustRegister(mailsSent)

	go func() {
		for {
			mailsQueue.Set(float64(len(mails)))
			message := <-mails

			log.Debug("Sending email")
			err := sendMail(message, conf)
			if err != nil {
				log.Error("Problem while sending email: ", err.Error())
				time.Sleep(2 * time.Second)

				continue
			}

			mailsSent.Inc()

			// Sleep to avoid overloading the server
			time.Sleep(100 * time.Millisecond)
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

func waitForStopSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	<-signals
}
