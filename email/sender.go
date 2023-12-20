package email

import (
	"bytes"
	"log"
	"text/template"
)

type SendEmailInput struct {
	To      string
	Subject string
	Body    string
}

func NewSendEmailInput(to string, subject string) *SendEmailInput {
	return &SendEmailInput{
		To:      to,
		Subject: subject,
	}
}

func (e *SendEmailInput) GenerateBody(templateFileName string, results interface{}) error {
	//TODO: Parse results from array
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		log.Fatal()
		return err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, results); err != nil {
		return err
	}

	e.Body = buf.String()

	return nil
}
