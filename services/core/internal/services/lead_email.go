package services

import (
	"fmt"
	"strings"

	resend "github.com/resend/resend-go/v2"

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
	if strings.TrimSpace(s.config.Email.ResendAPIKey) == "" {
		return fmt.Errorf("resend api key is not configured")
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

	client := resend.NewClient(s.config.Email.ResendAPIKey)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.config.Email.FromName, s.config.Email.FromEmail),
		To:      []string{recipient},
		Subject: subject,
		Text:    body,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send lead email: %w", err)
	}

	return nil
}

func emptyToDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
