package services

import (
	"fmt"
	"strings"

	resend "github.com/resend/resend-go/v2"

	"github.com/animalpride/partners/services/core/internal/config"
	"github.com/animalpride/partners/services/core/internal/models"
)

type LeadEmailService struct {
	config *config.Config
}

func NewLeadEmailService(cfg *config.Config) *LeadEmailService {
	return &LeadEmailService{config: cfg}
}

func (s *LeadEmailService) SendPartnerLead(app *models.PartnerApplication) error {
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

	subject := fmt.Sprintf("New Partner Application: %s", app.OrganizationName)
	line2 := ""
	if strings.TrimSpace(app.AddressLine2) != "" {
		line2 = app.AddressLine2 + "\n"
	}
	body := fmt.Sprintf(
		"A new partner application was submitted.\n\nOrganization: %s\nContact: %s\nEmail: %s\nPhone: %s\nAddress: %s\n%sCity: %s\nState: %s\nPostal Code: %s\nCountry: %s\nWebsite: %s\nMonthly Traffic: %s\nCurrent Store: %s\nGoals: %s\nNotes: %s\nSubmitted At: %s\n",
		app.OrganizationName,
		app.ContactName,
		app.Email,
		app.Phone,
		app.AddressLine1,
		line2,
		app.City,
		app.State,
		app.PostalCode,
		app.Country,
		emptyToDash(app.Website),
		emptyToDash(app.MonthlyTraffic),
		emptyToDash(app.CurrentStore),
		emptyToDash(app.Goals),
		emptyToDash(app.Notes),
		app.CreatedAt.UTC().Format("2006-01-02 15:04:05 MST"),
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
