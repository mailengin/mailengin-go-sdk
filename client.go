package mailengin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	Version        = "0.2.0"
	defaultBaseURL = "https://api.mailengin.app"
)

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Option func(*Client) error

type Client struct {
	apiKey  string
	baseURL string
	timeout time.Duration
	http    HTTPDoer
	Emails  *EmailsService
}

func New(apiKey string, options ...Option) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("mailengin requires a non-empty API key")
	}
	client := &Client{apiKey: strings.TrimSpace(apiKey), baseURL: defaultBaseURL, timeout: 30 * time.Second, http: http.DefaultClient}
	for _, option := range options {
		if err := option(client); err != nil {
			return nil, err
		}
	}
	client.baseURL = strings.TrimRight(client.baseURL, "/")
	client.Emails = &EmailsService{client: client}
	return client, nil
}

func WithBaseURL(value string) Option {
	return func(client *Client) error {
		if strings.TrimSpace(value) == "" {
			return errors.New("mailengin base URL cannot be empty")
		}
		client.baseURL = value
		return nil
	}
}

func WithTimeout(value time.Duration) Option {
	return func(client *Client) error {
		if value <= 0 {
			return errors.New("mailengin timeout must be positive")
		}
		client.timeout = value
		return nil
	}
}

func WithHTTPClient(value HTTPDoer) Option {
	return func(client *Client) error {
		if value == nil {
			return errors.New("mailengin HTTP client cannot be nil")
		}
		client.http = value
		return nil
	}
}

func (c *Client) post(ctx context.Context, path string, payload any, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return &Error{Message: "Unable to serialize MailEngin request.", Code: "invalid_request", Cause: err}
	}
	requestCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return &Error{Message: "Unable to create MailEngin request.", Code: "invalid_request", Cause: err}
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "mailengin-go/"+Version)
	response, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return &Error{Message: "MailEngin request was canceled.", Code: "request_aborted", Cause: err}
		}
		if requestCtx.Err() == context.DeadlineExceeded {
			return &Error{Message: "MailEngin request timed out.", Code: "request_timeout", Cause: err}
		}
		return &Error{Message: "Unable to reach the MailEngin API.", Code: "network_error", Cause: err}
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return &Error{Message: "Unable to read the MailEngin response.", Code: "network_error", Cause: err}
	}
	var parsed any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &parsed)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message, code := statusMessage(response.StatusCode), ""
		if object, ok := parsed.(map[string]any); ok {
			if value, ok := object["message"].(string); ok {
				message = value
			}
			if value, ok := object["code"].(string); ok {
				code = value
			}
		}
		var retryAfter *float64
		if value, parseErr := strconv.ParseFloat(response.Header.Get("Retry-After"), 64); parseErr == nil {
			retryAfter = &value
		}
		return &Error{Message: message, Status: response.StatusCode, Code: code, RequestID: response.Header.Get("x-request-id"), RetryAfter: retryAfter, Body: parsed}
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return &Error{Message: "MailEngin API returned invalid JSON.", Code: "invalid_response", Body: string(raw), Cause: err}
	}
	return nil
}
