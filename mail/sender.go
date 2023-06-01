package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

type iEmailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

type GmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

func NewGmailSender(name string, fromEmailAdd string, fromEmailPass string) iEmailSender {
	return &GmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAdd,
		fromEmailPassword: fromEmailPass,
	}
}

func (sender *GmailSender) SendEmail(subject string, content string, to []string, cc []string, bcc []string, attachFiles []string) error {
	// 1. create email object
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	// 2. attach files
	for _, f := range attachFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	// 3. authenticate with smtp
	smtpAuth := smtp.PlainAuth("", sender.fromEmailAddress, sender.fromEmailPassword, smtpAuthAddress)
	// 4. send the email
	return e.Send(smtpServerAddress, smtpAuth)
}
