package mailengin

import "fmt"

// Error describes an API, timeout, cancellation, or network failure.
type Error struct {
	Message    string
	Status     int
	Code       string
	RequestID  string
	RetryAfter *float64
	Body       any
	Cause      error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Cause }

// IsRetryable reports whether retrying may succeed. Sends are never retried automatically.
func (e *Error) IsRetryable() bool {
	return e.Code == "network_error" || e.Code == "request_timeout" || e.Status == 408 || e.Status == 429 || e.Status >= 500
}

func statusMessage(status int) string {
	return fmt.Sprintf("MailEngin API request failed with status %d.", status)
}
