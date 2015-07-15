package mailer

import (
	"bytes"
	"net/smtp"

	"github.com/juju/errgo"
)

type Mail struct {
	Sender      string
	Destination string
	Message     string
}

func (mail *Mail) Send(server string) error {
	conn, err := smtp.Dial(server)
	if err != nil {
		return errgo.Notef(err, "can not connect to server")
	}
	defer conn.Close()

	conn.Mail(mail.Sender)
	conn.Rcpt(mail.Destination)

	wc, err := conn.Data()
	if err != nil {
		return errgo.Notef(err, "can not get writer from connection")
	}
	defer wc.Close()

	buffer := bytes.NewBufferString(mail.Message)

	_, err = buffer.WriteTo(wc)
	if err != nil {
		return errgo.Notef(err, "can not write message to connection writer")
	}

	return nil
}
