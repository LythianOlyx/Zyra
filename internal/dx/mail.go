package dx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/smtp"
	"sync"
	"time"
)

// Email represents an email message to be sent.
type Email struct {
	To          []string          `json:"to"`
	From        string            `json:"from"`
	Subject     string            `json:"subject"`
	Body        string            `json:"body"`
	HTML        string            `json:"html"`
	Template    string            `json:"template"`
	Data        map[string]any    `json:"data"`
	Headers     map[string]string `json:"headers"`
	Attachments []Attachment      `json:"attachments,omitempty"`
}

// Attachment represents an email file attachment.
type Attachment struct {
	Filename    string `json:"filename"`
	Content     []byte `json:"-"`
	ContentType string `json:"contentType"`
}

// Mailer defines the contract for sending emails.
type Mailer interface {
	Send(ctx context.Context, email Email) error
}

// MailManager orchestrates dev and production email adapters.
type MailManager struct {
	mu            sync.RWMutex
	mode          string // "dev", "smtp", "resend"
	devMailer     *DevMailer
	smtpMailer    *SMTPMailer
	resendMailer  *ResendMailer
	templateStore map[string]*template.Template
}

// DevMailer logs and stores emails locally for development testing.
type DevMailer struct {
	mu     sync.RWMutex
	Emails []Email
}

func NewDevMailer() *DevMailer {
	return &DevMailer{
		Emails: make([]Email, 0),
	}
}

func (d *DevMailer) Send(ctx context.Context, email Email) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if email.HTML == "" && email.Template != "" && email.Data != nil {
		// Simple fallback template rendering
		email.HTML = fmt.Sprintf("<!-- Rendered Template: %s -->\n<div>%v</div>", email.Template, email.Data)
	}

	d.Emails = append(d.Emails, email)
	fmt.Printf("[zyra.Mail DevLog %s] To: %v | Subject: %s\n", time.Now().Format(time.RFC3339), email.To, email.Subject)
	return nil
}

func (d *DevMailer) LastSent() (Email, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if len(d.Emails) == 0 {
		return Email{}, false
	}
	return d.Emails[len(d.Emails)-1], true
}

func (d *DevMailer) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Emails = make([]Email, 0)
}

// SMTPMailer delivers email using standard SMTP.
type SMTPMailer struct {
	Host     string
	Port     int
	Username string
	Password string
}

func (s *SMTPMailer) Send(ctx context.Context, email Email) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	body := email.Body
	if email.HTML != "" {
		body = email.HTML
	}

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		email.To[0], email.Subject, body))

	return smtp.SendMail(addr, auth, email.From, email.To, msg)
}

// ResendMailer delivers email using Resend HTTP API.
type ResendMailer struct {
	APIKey     string
	HTTPClient *http.Client
}

func (r *ResendMailer) Send(ctx context.Context, email Email) error {
	if r.APIKey == "" {
		return fmt.Errorf("resend api key is missing")
	}

	client := r.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	payload := map[string]any{
		"from":    email.From,
		"to":      email.To,
		"subject": email.Subject,
	}
	if email.HTML != "" {
		payload["html"] = email.HTML
	} else {
		payload["text"] = email.Body
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend api returned error status: %d", resp.StatusCode)
	}

	return nil
}

// NewMailManager creates a MailManager defaulting to dev mode.
func NewMailManager() *MailManager {
	dev := NewDevMailer()
	return &MailManager{
		mode:          "dev",
		devMailer:     dev,
		templateStore: make(map[string]*template.Template),
	}
}

func (m *MailManager) SetMode(mode string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mode = mode
}

func (m *MailManager) ConfigureSMTP(host string, port int, user, pass string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.smtpMailer = &SMTPMailer{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
	}
}

func (m *MailManager) ConfigureResend(apiKey string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resendMailer = &ResendMailer{
		APIKey: apiKey,
	}
}

func (m *MailManager) RegisterTemplate(name string, tmplContent string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	parsed, err := template.New(name).Parse(tmplContent)
	if err != nil {
		return err
	}
	m.templateStore[name] = parsed
	return nil
}

func (m *MailManager) Send(ctx context.Context, email Email) error {
	m.mu.RLock()
	mode := m.mode
	dev := m.devMailer
	smtpM := m.smtpMailer
	resendM := m.resendMailer
	tmplStore := m.templateStore
	m.mu.RUnlock()

	if email.Template != "" && email.HTML == "" {
		if tmpl, ok := tmplStore[email.Template]; ok {
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, email.Data); err == nil {
				email.HTML = buf.String()
			}
		}
	}

	switch mode {
	case "smtp":
		if smtpM != nil {
			return smtpM.Send(ctx, email)
		}
		return fmt.Errorf("smtp mailer not configured")
	case "resend":
		if resendM != nil {
			return resendM.Send(ctx, email)
		}
		return fmt.Errorf("resend mailer not configured")
	default:
		return dev.Send(ctx, email)
	}
}

func (m *MailManager) DevLogs() []Email {
	if m.devMailer != nil {
		m.devMailer.mu.RLock()
		defer m.devMailer.mu.RUnlock()
		cp := make([]Email, len(m.devMailer.Emails))
		copy(cp, m.devMailer.Emails)
		return cp
	}
	return nil
}
