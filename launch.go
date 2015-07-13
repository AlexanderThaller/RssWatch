package main

import (
	"bytes"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlexanderThaller/logger"
	"github.com/prometheus/client_golang/prometheus"
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

	http.Handle("/metrics", prometheus.Handler())
	go http.ListenAndServe(":9132", nil)

	l.Trace("Watching for signals")
	waitForStopSignal()
	return nil
}

func launchMails(conf *Config) (chan<- *bytes.Buffer, error) {
	l := logger.New(name, "launch", "Mails")
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

			l.Debug("Sending email")
			l.Trace("Message:\n", message)
			err := sendMail(message, conf)
			if err != nil {
				l.Error("Problem while sending email: ", err)
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
