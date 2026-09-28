package platforms

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type Email struct {
	dialer *gomail.Dialer
	from   string
}

func NewEmail(from, appPassword string) *Email {
	return &Email{
		dialer: gomail.NewDialer("smtp.gmail.com", 587, from, appPassword),
		from:   from,
	}
}

type EmailRecipient struct {
	*Email
	to string
}

func (e *Email) NewRecipient(to string) *EmailRecipient {
	return &EmailRecipient{
		Email: e,
		to:    to,
	}
}

func (r *EmailRecipient) Send(title string, body string, deviceID string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", r.from)
	m.SetHeader("To", r.to)
	m.SetHeader("Subject", title)
	m.SetBody("text/html", fmt.Sprintf(
		"<b>Device:</b> %s<br><pre>%s</pre>",
		deviceID, body,
	))

	return r.dialer.DialAndSend(m)
}
