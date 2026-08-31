package mailengin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendMapsRequestAndResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/developer/send" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer re_test_key" {
			t.Errorf("authorization = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != "mailengin-go/0.2.0" {
			t.Errorf("user agent = %q", got)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["template_name"] != "welcome" {
			t.Errorf("template_name = %#v", body["template_name"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"msg_1","from":"hello@example.com","to":"person@example.com","template_name":"welcome","created_at":"2026-08-18T10:00:00Z"}`))
	}))
	defer server.Close()
	client, err := New("re_test_key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Emails.Send(context.Background(), SendEmailRequest{To: "person@example.com", TemplateName: "welcome"})
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "msg_1" {
		t.Fatalf("id = %q", result.ID)
	}
}

func TestRateLimitAndValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "12")
		w.Header().Set("x-request-id", "req_1")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"Rate limit exceeded","code":"rate_limited"}`))
	}))
	defer server.Close()
	client, _ := New("re_test_key", WithBaseURL(server.URL))
	_, err := client.Emails.Send(context.Background(), SendEmailRequest{To: "person@example.com", TemplateName: "welcome"})
	apiErr, ok := err.(*Error)
	if !ok || apiErr.Status != 429 || apiErr.RetryAfter == nil || *apiErr.RetryAfter != 12 || !apiErr.IsRetryable() {
		t.Fatalf("unexpected error: %#v", err)
	}
	_, err = client.Emails.Send(context.Background(), SendEmailRequest{To: "person@example.com", HTML: "<p>Hello</p>"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestBulkMappingLimitAndCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/developer/send-bulk" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if len(body["to"].([]any)) != 2 {
			t.Fatalf("unexpected recipients: %#v", body["to"])
		}
		_, _ = w.Write([]byte(`{"success":true,"jobId":"bulk_1","queued_count":2,"template_name":"welcome","message":"Queued"}`))
	}))
	defer server.Close()
	client, _ := New("re_test_key", WithBaseURL(server.URL))
	result, err := client.Emails.SendBulk(context.Background(), SendBulkEmailRequest{
		To:           []BulkRecipient{{Email: "a@example.com"}, {Email: "b@example.com", Variables: Variables{"name": "B"}}},
		TemplateName: "welcome",
	})
	if err != nil || result.JobID != "bulk_1" {
		t.Fatalf("unexpected result: %#v, %v", result, err)
	}
	tooMany := make([]BulkRecipient, 1001)
	for index := range tooMany {
		tooMany[index] = BulkRecipient{Email: "a@example.com"}
	}
	if _, err := client.Emails.SendBulk(context.Background(), SendBulkEmailRequest{To: tooMany, TemplateName: "welcome"}); err == nil {
		t.Fatal("expected recipient limit error")
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.Emails.Send(canceled, SendEmailRequest{To: "a@example.com", TemplateName: "welcome"}); err == nil {
		t.Fatal("expected cancellation error")
	}
}
