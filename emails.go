package mailengin

import (
	"context"
	"errors"
	"strings"
)

type EmailsService struct { client *Client }

func (service *EmailsService) Send(ctx context.Context, request SendEmailRequest) (*SendEmailResponse, error) {
	if strings.TrimSpace(request.To) == "" { return nil, errors.New("mailengin.Emails.Send requires To") }
	if err := requireContent(request.TemplateName, request.TemplateID, request.HTML, request.Subject); err != nil { return nil, err }
	payload := map[string]any{"to": request.To}
	put(payload, "template_name", request.TemplateName)
	put(payload, "template_id", request.TemplateID)
	if request.Variables != nil { payload["variables"] = request.Variables }
	put(payload, "subject", request.Subject)
	put(payload, "from_email", request.FromEmail)
	put(payload, "html", request.HTML)
	if request.ReplyToMailEngin != nil { payload["reply_to_mailengin"] = *request.ReplyToMailEngin }
	var wire struct {
		ID string `json:"id"`
		From string `json:"from"`
		To string `json:"to"`
		TemplateName *string `json:"template_name"`
		CreatedAt string `json:"created_at"`
	}
	if err := service.client.post(ctx, "/api/developer/send", payload, &wire); err != nil { return nil, err }
	if wire.ID == "" || wire.From == "" || wire.To == "" || wire.CreatedAt == "" {
		return nil, &Error{Message: "MailEngin API returned an invalid response.", Code: "invalid_response", Body: wire}
	}
	return &SendEmailResponse{ID: wire.ID, From: wire.From, To: wire.To, TemplateName: wire.TemplateName, CreatedAt: wire.CreatedAt}, nil
}

func (service *EmailsService) SendBulk(ctx context.Context, request SendBulkEmailRequest) (*SendBulkEmailResponse, error) {
	if len(request.To) == 0 { return nil, errors.New("mailengin.Emails.SendBulk requires recipients") }
	if len(request.To) > 1000 { return nil, errors.New("mailengin.Emails.SendBulk accepts up to 1000 recipients") }
	if err := requireContent(request.TemplateName, request.TemplateID, request.HTML, request.Subject); err != nil { return nil, err }
	recipients := make([]any, 0, len(request.To))
	for _, recipient := range request.To {
		if strings.TrimSpace(recipient.Email) == "" { return nil, errors.New("every bulk recipient requires an email address") }
		if recipient.Variables == nil { recipients = append(recipients, recipient.Email) } else { recipients = append(recipients, map[string]any{"email": recipient.Email, "variables": recipient.Variables}) }
	}
	payload := map[string]any{"to": recipients}
	put(payload, "template_name", request.TemplateName)
	put(payload, "template_id", request.TemplateID)
	if request.Variables != nil { payload["variables"] = request.Variables }
	put(payload, "subject", request.Subject)
	put(payload, "from_email", request.FromEmail)
	put(payload, "html", request.HTML)
	if request.ReplyToMailEngin != nil { payload["reply_to_mailengin"] = *request.ReplyToMailEngin }
	var wire struct {
		Success bool `json:"success"`
		JobID string `json:"jobId"`
		QueuedCount int `json:"queued_count"`
		SentCount *int `json:"sent_count"`
		FailedCount *int `json:"failed_count"`
		TemplateName *string `json:"template_name"`
		Message string `json:"message"`
	}
	if err := service.client.post(ctx, "/api/developer/send-bulk", payload, &wire); err != nil { return nil, err }
	if wire.JobID == "" || wire.Message == "" {
		return nil, &Error{Message: "MailEngin API returned an invalid response.", Code: "invalid_response", Body: wire}
	}
	return &SendBulkEmailResponse{Success: wire.Success, JobID: wire.JobID, QueuedCount: wire.QueuedCount, SentCount: wire.SentCount, FailedCount: wire.FailedCount, TemplateName: wire.TemplateName, Message: wire.Message}, nil
}

func requireContent(templateName, templateID, html, subject string) error {
	hasTemplate := strings.TrimSpace(templateName) != "" || strings.TrimSpace(templateID) != ""
	if !hasTemplate && strings.TrimSpace(html) == "" { return errors.New("provide TemplateName, TemplateID, or HTML") }
	if !hasTemplate && strings.TrimSpace(subject) == "" { return errors.New("raw HTML sends require Subject") }
	return nil
}

func put(values map[string]any, key, value string) { if value != "" { values[key] = value } }
