package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type SMTPSender struct {
	from string
	pass string
	host string
	port int
}

func NewSMTPSender(from, pass, host string, port int) *SMTPSender {
	return &SMTPSender{
		from: from,
		pass: pass,
		host: host,
		port: port,
	}
}

func (s *SMTPSender) SendMessage(input SendEmailInput) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", input.To)
	m.SetHeader("Subject", input.Subject)
	m.SetBody("text/html", input.Body)

	d := gomail.NewDialer(s.host, s.port, s.from, s.pass)

	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
