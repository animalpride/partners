package services

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/animalpride/animalpride-core/services/core/internal/config"
	"github.com/animalpride/animalpride-core/services/core/internal/models"
)

type LeadEmailService struct {
	config *config.Config
}

func NewLeadEmailService(cfg *config.Config) *LeadEmailService {
	return &LeadEmailService{config: cfg}
}

func (s *LeadEmailService) SendPartnerLead(lead *models.LeadSubmission) error {
	recipient := strings.TrimSpace(s.config.Email.PartnerLeadsTo)
	if recipient == "" {
		return fmt.Errorf("partner lead recipient is not configured")
	}
	if strings.TrimSpace(s.config.Email.SMTPHost) == "" || s.config.Email.SMTPPort <= 0 {
		return fmt.Errorf("smtp host or port is not configured")
	}
	if strings.TrimSpace(s.config.Email.FromEmail) == "" {
		return fmt.Errorf("from email is not configured")
	}

	subject := fmt.Sprintf("New Partner Lead: %s", lead.OrganizationName)
	body := fmt.Sprintf(
		"A new partner lead was submitted.\n\nOrganization: %s\nContact: %s\nEmail: %s\nPhone: %s\nWebsite: %s\nMonthly Traffic: %s\nCurrent Store: %s\nGoals: %s\nNotes: %s\nSubmitted At: %s\n",
		lead.OrganizationName,
		lead.ContactName,
		lead.Email,
		emptyToDash(lead.Phone),
		emptyToDash(lead.Website),
		emptyToDash(lead.MonthlyTraffic),
		emptyToDash(lead.CurrentStore),
		emptyToDash(lead.Goals),
		emptyToDash(lead.Notes),
		lead.CreatedAt.UTC().Format("2006-01-02 15:04:05 MST"),
	)

	headers := map[string]string{
		"From":         fmt.Sprintf("%s <%s>", s.config.Email.FromName, s.config.Email.FromEmail),
		"To":           recipient,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}

	message := ""
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + body

	smtpAddr := fmt.Sprintf("%s:%d", s.config.Email.SMTPHost, s.config.Email.SMTPPort)
	client, err := smtp.Dial(smtpAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to smtp server: %w", err)
	}
	defer client.Close()

	if err := client.Hello("partners.animalpride.com"); err != nil {
		return fmt.Errorf("failed to say hello to smtp server: %w", err)
	}

	if s.config.Email.SMTPTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("smtp server does not support STARTTLS")
		}
		tlsConfig := &tls.Config{ServerName: s.config.Email.SMTPHost}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start tls: %w", err)
		}
	}

	if s.config.Email.SMTPAuth {
		auth := smtp.PlainAuth("", s.config.Email.SMTPUser, s.config.Email.SMTPPassword, s.config.Email.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate with smtp server: %w", err)
		}
	}

	if err := client.Mail(s.config.Email.FromEmail); err != nil {
		return fmt.Errorf("failed to set from address: %w", err)
	}
	if err := client.Rcpt(recipient); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	dataWriter, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to begin message body: %w", err)
	}
	if _, err := dataWriter.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write message body: %w", err)
	}
	if err := dataWriter.Close(); err != nil {
		return fmt.Errorf("failed to finalize message body: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("failed to close smtp session: %w", err)
	}

	return nil
}

func emptyToDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
