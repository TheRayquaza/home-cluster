package email

import (
	"fmt"
	"log"
	"net/smtp"

	"kommande/internal/config"
)

func Send(cfg *config.Config, to, subject, body string) {
	if cfg.SMTPHost == "" || to == "" {
		return
	}
	go func() {
		addr := cfg.SMTPHost + ":" + cfg.SMTPPort
		msg := []byte(fmt.Sprintf(
			"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
			cfg.SMTPFrom, to, subject, body,
		))
		var auth smtp.Auth
		if cfg.SMTPUser != "" {
			auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
		}
		if err := smtp.SendMail(addr, auth, cfg.SMTPFrom, []string{to}, msg); err != nil {
			log.Printf("email error to %s: %v", to, err)
		}
	}()
}
