// utils/email.go
package utils

import (
	"fmt"
	"net/smtp"
	"strconv"

	"github.com/mfuadfakhruzzaki/backendaurauran/config"
)

// EmailService mengatur pengiriman email
type EmailService struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
}

// NewEmailService menginisialisasi EmailService
func NewEmailService() *EmailService {
	// Convert port from string to int
	port, err := strconv.Atoi(config.AppConfig.Email.Port)
	if err != nil {
		panic(fmt.Sprintf("Invalid SMTP port: %v", err))
	}

	return &EmailService{
		Host:     config.AppConfig.Email.Host,
		Port:     port,
		Username: config.AppConfig.Email.Username,
		Password: config.AppConfig.Email.Password,
		Sender:   config.AppConfig.Email.Sender,
	}
}

// SendEmail mengirim email ke penerima dengan subject dan body tertentu
func (e *EmailService) SendEmail(to string, subject string, body string) error {
	auth := smtp.PlainAuth("", e.Username, e.Password, e.Host)
	msg := []byte("To: " + to + "\r\n" +
		"From: " + e.Sender + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		body + "\r\n")
	addr := fmt.Sprintf("%s:%d", e.Host, e.Port)
	return smtp.SendMail(addr, auth, e.Sender, []string{to}, msg)
}

// SendVerificationEmail mengirim email verifikasi
func (e *EmailService) SendVerificationEmail(to string, token string) error {
	verifyURL := fmt.Sprintf(config.AppConfig.Email.VerifyURL, token)
	subject := "Verifikasi Email Anda"
	body := `<p>Halo,</p>
             <p>Terima kasih telah mendaftar. Silakan klik link di bawah ini untuk memverifikasi email Anda:</p>
             <a href="` + verifyURL + `">Verifikasi Email</a>
             <p>Jika Anda tidak melakukan pendaftaran, silakan abaikan email ini.</p>`
	return e.SendEmail(to, subject, body)
}

// SendResetPasswordEmail mengirim email reset password
func (e *EmailService) SendResetPasswordEmail(to string, token string) error {
	resetURL := fmt.Sprintf(config.AppConfig.Email.ResetPasswordURL, token)
	subject := "Reset Password Anda"
	body := `<p>Halo,</p>
             <p>Anda telah meminta untuk mereset password Anda. Silakan klik link di bawah ini untuk melanjutkan:</p>
             <a href="` + resetURL + `">Reset Password</a>
             <p>Jika Anda tidak melakukan permintaan ini, silakan abaikan email ini.</p>`
	return e.SendEmail(to, subject, body)
}
