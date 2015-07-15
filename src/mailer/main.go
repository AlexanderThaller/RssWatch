package mailer

import (
	"bytes"
	"io/ioutil"
	"net/smtp"
	"os"
	"sync"
	"time"

	"github.com/Unknwon/log"
	"github.com/juju/errgo"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

const (
	DefaultQueueFilePath   = "queue.msgpack"
	DefaultBufferSize      = 50000
	DefaultSleepDurationMs = 100
)

type Mailer struct {
	server       string
	activeSender *sync.WaitGroup
	stopping     bool
	queue        chan Mail

	QueueFilePath string
	SleepDuration uint64
}

func New(server string) *Mailer {
	mailer := new(Mailer)
	mailer.server = server
	mailer.QueueFilePath = DefaultQueueFilePath
	mailer.SleepDuration = DefaultSleepDurationMs

	mailer.activeSender = new(sync.WaitGroup)
	mailer.queue = make(chan Mail, DefaultBufferSize)

	return mailer
}

func (mailer *Mailer) Launch() (chan<- Mail, error) {
	err := mailer.LoadQueue(mailer.QueueFilePath)
	if err != nil {
		return nil, errgo.Notef(err, "can not restore queue")
	}

	go mailer.startLoop(mailer.queue)

	return mailer.queue, nil
}

func (mailer *Mailer) startLoop(queue <-chan Mail) {
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

	for {
		if mailer.stopping {
			break
		}

		mailsQueue.Set(float64(len(queue)))
		message := <-queue

		log.Debug("Sending email")
		mailer.activeSender.Add(1)
		err := sendMail(message, mailer.server)
		if err != nil {
			log.Error("Problem while sending email: ", err.Error())
			time.Sleep(2 * time.Second)
			mailer.activeSender.Done()

			continue
		}

		mailsSent.Inc()
		mailer.activeSender.Done()

		// Sleep to avoid overloading the server
		time.Sleep(time.Duration(mailer.SleepDuration) * time.Millisecond)
	}
}

func (mailer *Mailer) Stop() error {
	mailer.stopping = true
	mailer.activeSender.Wait()

	err := mailer.SaveQueue(mailer.QueueFilePath)
	if err != nil {
		return errgo.Notef(err, "can not save queue to file")
	}

	return nil
}

func (mailer *Mailer) SaveQueue(path string) error {
	var data []Mail

	for {
		if len(mailer.queue) == 0 {
			break
		}

		message := <-mailer.queue
		data = append(data, message)
	}
	log.Debug("Queue: ", data)

	marshal, err := msgpack.Marshal(data)
	if err != nil {
		return errgo.Notef(err, "can not marshal data to msgpack format")
	}

	err = ioutil.WriteFile(path, marshal, 0644)
	if err != nil {
		return errgo.Notef(err, "can not write data to file")
	}

	return nil
}

func (mailer *Mailer) LoadQueue(path string) error {
	// If there is no queue file we can return early
	_, err := os.Stat(path)
	if err != nil {
		return nil
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var data []Mail
	err = msgpack.Unmarshal(file, &data)
	if err != nil {
		return err
	}

	for _, message := range data {
		mailer.queue <- message
	}

	return nil
}

func sendMail(message Mail, server string) error {
	conn, err := smtp.Dial(server)
	if err != nil {
		return errgo.Notef(err, "can not connect to server")
	}
	defer conn.Close()

	conn.Mail(message.Sender)
	conn.Rcpt(message.Destination)

	wc, err := conn.Data()
	if err != nil {
		return errgo.Notef(err, "can not get writer from connection")
	}
	defer wc.Close()

	buffer := bytes.NewBufferString(message.Message)

	_, err = buffer.WriteTo(wc)
	if err != nil {
		return errgo.Notef(err, "can not write message to connection writer")
	}

	return nil
}
