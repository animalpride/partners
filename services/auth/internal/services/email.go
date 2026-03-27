package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"strings"

	resend "github.com/resend/resend-go/v2"

	"github.com/animalpride/animalpride-core/services/denops-auth/internal/config"
)

type EmailService struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

// InvitationEmailData contains data for invitation email template
type InvitationEmailData struct {
	RecipientName   string
	InviterName     string
	ApplicationName string
	CompanyName     string
	InviteLink      string
	SupportEmail    string
	RoleName        string
}

type PasswordResetEmailData struct {
	ApplicationName string
	CompanyName     string
	ResetLink       string
	SupportEmail    string
}

type PasswordChangedEmailData struct {
	ApplicationName string
	CompanyName     string
	SupportEmail    string
}

// SendInvitationEmail sends an invitation email to a new user.
// roleName is used to tailor the email body (e.g. "admin", "partner").
func (s *EmailService) SendInvitationEmail(to, recipientName, roleName, invitationToken string) error {
	inviteLink := s.buildInviteLink(invitationToken)

	// Email template data
	data := InvitationEmailData{
		RecipientName:   recipientName,
		InviterName:     "Animal Pride Partners Team",
		ApplicationName: "Animal Pride Partners",
		CompanyName:     "Animal Pride",
		InviteLink:      inviteLink,
		SupportEmail:    s.config.Email.FromEmail,
		RoleName:        roleName,
	}

	// HTML email template
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
		<title>You're Invited to Join {{.ApplicationName}}</title>
    <style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #1f2f3d; background-color: #f2f6f9; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.panel { border: 1px solid #d3dee8; border-radius: 12px; overflow: hidden; background-color: #e9eff4; }
		.header { background-color: #00698f; color: #ffffff; padding: 20px; text-align: center; }
		.content { padding: 30px 20px; background-color: #f2f6f9; }
        .button { 
            display: inline-block; 
			background-color: #00698f; 
			color: #ffffff; 
            padding: 12px 30px; 
            text-decoration: none; 
            border-radius: 5px; 
            margin: 20px 0;
			font-weight: 600;
			letter-spacing: 0.2px;
        }
		.footer { text-align: center; padding: 20px; font-size: 12px; color: #53677a; }
    </style>
</head>
<body>
	<div class="container">
		<div class="panel">
			<div class="header">
				<h1>Welcome to {{.ApplicationName}}</h1>
			</div>
			<div class="content">
				<h2>Hello{{if .RecipientName}} {{.RecipientName}}{{end}}!</h2>
				{{if eq .RoleName "admin"}}
				<p>You have been invited to join <strong>{{.ApplicationName}}</strong> as an <strong>Administrator</strong>.</p>
				<p>As an Administrator, you will have full access to manage content, users, and settings for the {{.CompanyName}} partners platform.</p>
				{{else if eq .RoleName "partner"}}
				<p>You have been invited to join <strong>{{.ApplicationName}}</strong> as a <strong>Partner</strong>.</p>
				<p>Your partner account will give you access to the {{.CompanyName}} partner portal and resources.</p>
				{{else}}
				<p>You have been invited to join <strong>{{.ApplicationName}}</strong>, the {{.CompanyName}} operations platform.</p>
				<p>The {{.InviterName}} has created an invitation for you.</p>
				{{end}}
				<p>To complete your registration, click the button below:</p>
				<div style="text-align: center;">
					<a href="{{.InviteLink}}" class="button" style="color: #ffffff !important; text-decoration: none;">Complete Registration</a>
				</div>
				<p><strong>What's next?</strong></p>
				<ul>
					<li>Open the invitation link above</li>
					<li>Set your name and password</li>
					<li>Start using {{.ApplicationName}} workflows</li>
				</ul>
				<p>If you have any questions, feel free to reach out to our support team at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a>.</p>
				<p>Welcome aboard!</p>
				<p>Best regards,<br>The {{.CompanyName}} Team</p>
			</div>
			<div class="footer">
				<p>This invitation link will expire in 48 hours for security reasons.</p>
				<p>If you didn't expect this invitation, you can safely ignore this email.</p>
			</div>
		</div>
	</div>
</body>
</html>`

	// Parse and execute template
	tmpl, err := template.New("invitation").Parse(htmlTemplate)
	if err != nil {
		log.Printf("SendInvitationEmail: parse template failed: %v", err)
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		log.Printf("SendInvitationEmail: execute template failed: %v", err)
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	subject := fmt.Sprintf("You're invited to join %s", data.ApplicationName)
	return s.sendHTML(to, subject, body.String())
}

func (s *EmailService) buildInviteLink(token string) string {
	base := strings.TrimSpace(s.config.Email.Links.InviteBaseURL)
	if base == "" {
		base = "http://localhost:3000/accept-invitation"
	}
	parsed, err := url.Parse(base)
	if err != nil {
		log.Printf("buildInviteLink: parse failed: %v", err)
		return base + "?token=" + url.QueryEscape(token)
	}
	query := parsed.Query()
	query.Set("token", token)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func (s *EmailService) buildResetLink(token string) string {
	base := strings.TrimSpace(s.config.Email.Links.ResetBaseURL)
	if base == "" {
		base = "http://localhost:3000/reset-password"
	}
	parsed, err := url.Parse(base)
	if err != nil {
		log.Printf("buildResetLink: parse failed: %v", err)
		return base + "?token=" + url.QueryEscape(token)
	}
	query := parsed.Query()
	query.Set("token", token)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

// SendPasswordResetEmail sends a password reset email with a single-use link.
func (s *EmailService) SendPasswordResetEmail(to, resetToken string) error {
	resetLink := s.buildResetLink(resetToken)

	data := PasswordResetEmailData{
		ApplicationName: "Animal Pride Partners",
		CompanyName:     "Animal Pride",
		ResetLink:       resetLink,
		SupportEmail:    s.config.Email.FromEmail,
	}

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
		<title>Reset your {{.ApplicationName}} password</title>
    <style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #1f2f3d; background-color: #f2f6f9; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.panel { border: 1px solid #d3dee8; border-radius: 12px; overflow: hidden; background-color: #e9eff4; }
		.header { background-color: #00698f; color: #ffffff; padding: 20px; text-align: center; }
		.content { padding: 30px 20px; background-color: #f2f6f9; }
        .button {
            display: inline-block;
			background-color: #00698f;
			color: #ffffff;
            padding: 12px 30px;
            text-decoration: none;
            border-radius: 5px;
            margin: 20px 0;
			font-weight: 600;
			letter-spacing: 0.2px;
        }
		.footer { text-align: center; padding: 20px; font-size: 12px; color: #53677a; }
    </style>
</head>
<body>
	<div class="container">
		<div class="panel">
			<div class="header">
				<h1>{{.ApplicationName}} Password Reset</h1>
			</div>
			<div class="content">
				<p>We received a request to reset your <strong>{{.ApplicationName}}</strong> password.</p>
				<p>Click the button below to choose a new password:</p>
				<div style="text-align: center;">
					<a href="{{.ResetLink}}" class="button" style="color: #ffffff !important; text-decoration: none;">Reset password</a>
				</div>
				<p>This link expires in 15 minutes and can only be used once.</p>
				<p>If you did not request a password reset, you can safely ignore this email.</p>
				<p>If you have questions, contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a>.</p>
				<p>Thanks,<br>The {{.CompanyName}} Team</p>
			</div>
			<div class="footer">
				<p>For your security, never share reset links with anyone.</p>
			</div>
		</div>
	</div>
</body>
</html>`

	tmpl, err := template.New("password_reset").Parse(htmlTemplate)
	if err != nil {
		log.Printf("SendPasswordResetEmail: parse template failed: %v", err)
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		log.Printf("SendPasswordResetEmail: execute template failed: %v", err)
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	subject := fmt.Sprintf("Reset your %s password", data.ApplicationName)
	return s.sendHTML(to, subject, body.String())
}

// SendPasswordChangedEmail notifies the user of a successful password change.
func (s *EmailService) SendPasswordChangedEmail(to string) error {
	data := PasswordChangedEmailData{
		ApplicationName: "Animal Pride Partners",
		CompanyName:     "Animal Pride",
		SupportEmail:    s.config.Email.FromEmail,
	}

	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
		<title>Your {{.ApplicationName}} password was changed</title>
    <style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #1f2f3d; background-color: #f2f6f9; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.panel { border: 1px solid #d3dee8; border-radius: 12px; overflow: hidden; background-color: #e9eff4; }
		.header { background-color: #00698f; color: #ffffff; padding: 20px; text-align: center; }
		.content { padding: 30px 20px; background-color: #f2f6f9; }
		.footer { text-align: center; padding: 20px; font-size: 12px; color: #53677a; }
    </style>
</head>
<body>
	<div class="container">
		<div class="panel">
			<div class="header">
				<h1>Password Updated</h1>
			</div>
			<div class="content">
				<p>Your <strong>{{.ApplicationName}}</strong> password was changed successfully.</p>
				<p>If you did not make this change, please reset your password immediately or contact us at <a href="mailto:{{.SupportEmail}}">{{.SupportEmail}}</a>.</p>
				<p>Thanks,<br>The {{.CompanyName}} Team</p>
			</div>
			<div class="footer">
				<p>This is an automated notification. No reply is necessary.</p>
			</div>
		</div>
	</div>
</body>
</html>`

	tmpl, err := template.New("password_changed").Parse(htmlTemplate)
	if err != nil {
		log.Printf("SendPasswordChangedEmail: parse template failed: %v", err)
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		log.Printf("SendPasswordChangedEmail: execute template failed: %v", err)
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	subject := fmt.Sprintf("Your %s password was changed", data.ApplicationName)
	return s.sendHTML(to, subject, body.String())
}

func (s *EmailService) sendHTML(to, subject, htmlBody string) error {
	client := resend.NewClient(s.config.Email.ResendAPIKey)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.config.Email.FromName, s.config.Email.FromEmail),
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		log.Printf("sendHTML: resend failed: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
