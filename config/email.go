package config

import "os"

// EmailConfig menyimpan konfigurasi email
type EmailConfig struct {
    Host              string
    Port              string
    Username          string
    Password          string
    Sender            string
    VerifyURL         string
    ResetPasswordURL  string
}

// LoadEmailConfig memuat konfigurasi email dari variabel lingkungan
func LoadEmailConfig() EmailConfig {
    return EmailConfig{
        Host:             os.Getenv("SMTP_HOST"),
        Port:             os.Getenv("SMTP_PORT"),
        Username:         os.Getenv("SMTP_USERNAME"),
        Password:         os.Getenv("SMTP_PASSWORD"),
        Sender:           os.Getenv("SMTP_SENDER"),
        VerifyURL:        os.Getenv("EMAIL_VERIFY_URL"),
        ResetPasswordURL: os.Getenv("EMAIL_RESET_PASSWORD_URL"),
    }
}
