package services

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer      *gomail.Dialer
	from        string
	frontendURL string
}

func NewMailer(host string, port int, username, password, from, frontendURL string) *Mailer {
	return &Mailer{
		dialer:      gomail.NewDialer(host, port, username, password),
		from:        from,
		frontendURL: frontendURL,
	}
}

func (m *Mailer) SendResetPasswordEmail(to, token string) error {
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", m.frontendURL, token)
	
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", "Reset Twojego hasła - IOT Monitor")
	msg.SetBody("text/html", fmt.Sprintf(`
		<h2>Odzyskiwanie hasła</h2>
		<p>Otrzymaliśmy prośbę o zresetowanie hasła do Twojego konta.</p>
		<p>Kliknij w poniższy link, aby ustawić nowe hasło (link jest ważny przez 1 godzinę):</p>
		<p><a href="%s">%s</a></p>
		<p>Jeśli to nie Ty prosiłeś o zmianę hasła, możesz zignorować tę wiadomość.</p>
		`, resetLink, resetLink))

	return m.dialer.DialAndSend(msg)
}
