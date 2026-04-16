package mail

import (
	"context"
	"fmt"
	"net/smtp"
)

type Sender interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

type SMTPSender struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewSMTPSender(host string, port int, username, password, from string) *SMTPSender {
	return &SMTPSender{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func (s *SMTPSender) SendEmail(ctx context.Context, to, subject, body string) error {
	var auth smtp.Auth
	if s.Username != "" && s.Password != "" {
		auth = smtp.PlainAuth("", s.Username, s.Password, s.Host)
	}

	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", s.From, to, subject, body))
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	return smtp.SendMail(addr, auth, s.From, []string{to}, msg)
}
