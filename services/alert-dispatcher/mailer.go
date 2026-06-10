package main

import (
        "fmt"
        "time"

        "gopkg.in/gomail.v2"
)
type Mailer struct {
	dialer *gomail.Dialer
	from   string
}

func NewMailer(host string, port int, username, password, from string) *Mailer {
	return &Mailer{
		dialer: gomail.NewDialer(host, port, username, password),
		from:   from,
	}
}

func (m *Mailer) SendAlertEmail(to string, event AlertEvent) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	
	subject := "IOT Alert: " + event.Message
	title := "IOT Alert Triggered"
	if event.IsResolved {
		subject = "IOT Resolved: " + event.Message
		title = "IOT Alert Resolved"
	}
	msg.SetHeader("Subject", subject)

	msg.SetBody("text/html", fmt.Sprintf(`
		<h2>%s</h2>
		<p><strong>Message:</strong> %s</p>
		<p><strong>Sensor ID:</strong> %d</p>
		<p><strong>Value:</strong> %f</p>
		<p><strong>Time:</strong> %s</p>
		`, title, event.Message, event.SensorID, event.Value, event.Timestamp.Format(time.RFC1123)))

	return m.dialer.DialAndSend(msg)
}

