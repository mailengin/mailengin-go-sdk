package mailengin

type Variables map[string]any

type SendEmailRequest struct {
	To                string
	TemplateName      string
	TemplateID        string
	Variables         Variables
	Subject           string
	FromEmail         string
	HTML              string
	ReplyToMailEngin  *bool
}

type SendEmailResponse struct {
	ID           string
	From         string
	To           string
	TemplateName *string
	CreatedAt    string
}

type BulkRecipient struct {
	Email     string
	Variables Variables
}

type SendBulkEmailRequest struct {
	To               []BulkRecipient
	TemplateName     string
	TemplateID       string
	Variables        Variables
	Subject          string
	FromEmail        string
	HTML             string
	ReplyToMailEngin *bool
}

type SendBulkEmailResponse struct {
	Success      bool
	JobID        string
	QueuedCount  int
	SentCount    *int
	FailedCount  *int
	TemplateName *string
	Message      string
}

// Bool returns a pointer suitable for optional boolean request fields.
func Bool(value bool) *bool { return &value }
