# MailEngin Go SDK

[![Go](https://img.shields.io/badge/Go-1.22%2B-00add8.svg)](https://go.dev/)
[![Dependencies](https://img.shields.io/badge/runtime_dependencies-none-16a34a.svg)](./go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-111827.svg)](./LICENSE)

The official Go SDK for sending transactional email through [MailEngin](https://mailengin.app). It uses only the Go standard library and provides context cancellation, typed request and response structs, configurable timeouts, injectable HTTP transports, and structured errors.

> [!IMPORTANT]
> This module is for server-side applications only. Never embed a MailEngin API key in browser, mobile, desktop, or other client-distributed binaries.

## Requirements

- Go 1.22 or newer
- A MailEngin API key and verified sending domain

## Installation

```bash
go get github.com/vishveshrathore/mailengin-go-sdk
```

Import the module using its default package name:

```go
import "github.com/vishveshrathore/mailengin-go-sdk"
```

## Before You Send

1. [Verify a sending domain](https://mailengin.app/dashboard/domains).
2. [Create an API key](https://mailengin.app/dashboard/api-keys) and save the full secret.
3. [Create and publish a Developer Template](https://mailengin.app/dashboard/dev-templates).
4. Copy the template API name, such as `welcome-email`.

Store the key in your runtime's secret manager or environment:

```env
MAILENGIN_API_KEY=re_your_full_secret_key
```

MailEngin displays the full key only once. A masked key cannot authenticate requests.

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/vishveshrathore/mailengin-go-sdk"
)

func main() {
	client, err := mailengin.New(os.Getenv("MAILENGIN_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	email, err := client.Emails.Send(context.Background(), mailengin.SendEmailRequest{
		To:           "user@example.com",
		FromEmail:    "hello@yourdomain.com",
		TemplateName: "welcome-email",
		Variables:    mailengin.Variables{"first_name": "Asha"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(email.ID)
}
```

The published template supplies the subject and HTML. Values in `Variables` replace matching template variables such as `{{first_name}}`.

## Send One Email

```go
email, err := client.Emails.Send(ctx, mailengin.SendEmailRequest{
	To:           "customer@example.com",
	FromEmail:    "hello@yourdomain.com",
	TemplateName: "account-verification",
	Variables: mailengin.Variables{
		"first_name":       "Asha",
		"verification_url": "https://yourapp.com/verify/token",
	},
	ReplyToMailEngin: mailengin.Bool(true),
})
if err != nil {
	return err
}

fmt.Printf("queued email %s at %s\n", email.ID, email.CreatedAt)
```

### Send request fields

| Struct field | Type | Required | Description |
| --- | --- | --- | --- |
| `To` | `string` | Yes | Recipient email address. |
| `TemplateName` | `string` | Recommended | Published template API name or exact display name. |
| `TemplateID` | `string` | No | Legacy template identifier. Prefer `TemplateName`. |
| `Variables` | `mailengin.Variables` | No | Values used to render template variables. |
| `Subject` | `string` | Raw HTML only | Template subject override, or required subject for raw HTML. |
| `FromEmail` | `string` | Recommended | Sender on a verified domain authorized for the API key. |
| `HTML` | `string` | Advanced | Raw HTML used when no template is supplied. |
| `ReplyToMailEngin` | `*bool` | No | Route replies into the MailEngin inbox. Use `mailengin.Bool`. |

Exactly one content source is required: `TemplateName`, `TemplateID`, or `HTML`. Raw HTML sends also require `Subject`.

## Send Personalized Bulk Email

Bulk requests support up to 1,000 recipients. Request-level variables apply to every recipient; recipient variables take precedence.

```go
job, err := client.Emails.SendBulk(ctx, mailengin.SendBulkEmailRequest{
	To: []mailengin.BulkRecipient{
		{
			Email:     "asha@example.com",
			Variables: mailengin.Variables{"first_name": "Asha"},
		},
		{
			Email:     "ben@example.com",
			Variables: mailengin.Variables{"first_name": "Ben"},
		},
	},
	FromEmail:    "hello@yourdomain.com",
	TemplateName: "product-update",
	Variables:    mailengin.Variables{"product_name": "MailEngin"},
})
if err != nil {
	return err
}

fmt.Printf("queued %d recipients in job %s\n", job.QueuedCount, job.JobID)
```

For recipients without individual variables, set only `Email`:

```go
job, err := client.Emails.SendBulk(ctx, mailengin.SendBulkEmailRequest{
	To: []mailengin.BulkRecipient{
		{Email: "a@example.com"},
		{Email: "b@example.com"},
	},
	TemplateName: "maintenance-notice",
})
```

A successful bulk response confirms that recipients were queued. It is not a guarantee that every message was delivered.

## Send Raw HTML

Published templates are recommended for reusable product email. For a one-off message, provide both `Subject` and `HTML`:

```go
email, err := client.Emails.Send(ctx, mailengin.SendEmailRequest{
	To:        "user@example.com",
	FromEmail: "reports@yourdomain.com",
	Subject:   "Your report is ready",
	HTML:      "<h1>Report ready</h1><p>You can download it now.</p>",
})
```

## Context and Cancellation

Every send accepts a `context.Context`. Use it to impose a shorter deadline or cancel work when an incoming request ends:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

email, err := client.Emails.Send(ctx, request)
```

The effective request deadline is the earlier of the context deadline and the client's configured timeout.

## Sender Selection

MailEngin resolves the sender in this order:

1. `FromEmail` supplied in the request.
2. Sender saved in the published Developer Template.
3. `noreply@<authorized-domain>` fallback.

The sender domain must be verified and authorized for the API key.

## Error Handling

Use `errors.As` to inspect API, timeout, cancellation, malformed-response, and network failures:

```go
import "errors"

email, err := client.Emails.Send(ctx, request)
if err != nil {
	var mailEnginError *mailengin.Error
	if errors.As(err, &mailEnginError) {
		fmt.Println(mailEnginError.Message)
		fmt.Println(mailEnginError.Status)       // HTTP status, when available
		fmt.Println(mailEnginError.Code)         // Machine-readable error code
		fmt.Println(mailEnginError.RequestID)    // Include when contacting support
		fmt.Println(mailEnginError.RetryAfter)   // Seconds supplied with HTTP 429
		fmt.Println(mailEnginError.Body)         // Parsed JSON or response data
		fmt.Println(mailEnginError.IsRetryable())
	}
	return err
}
```

`IsRetryable()` is true for network errors, timeouts, HTTP `408`, HTTP `429`, and `5xx` responses. The SDK never retries sends automatically because a retry could create a duplicate email until idempotency keys are supported.

Input validation errors are ordinary Go errors because no API request was made.

## Configuration

```go
client, err := mailengin.New(
	os.Getenv("MAILENGIN_API_KEY"),
	mailengin.WithBaseURL("https://api.mailengin.app"),
	mailengin.WithTimeout(15*time.Second),
	mailengin.WithHTTPClient(http.DefaultClient),
)
```

| Option | Default | Description |
| --- | --- | --- |
| API key | None | Full server-side MailEngin API key. |
| `WithBaseURL` | `https://api.mailengin.app` | Override for local, test, or dedicated environments. |
| `WithTimeout` | 30 seconds | Maximum duration applied to each request. |
| `WithHTTPClient` | `http.DefaultClient` | Any value implementing `Do(*http.Request)`. |

Create one client and reuse it. The client and standard HTTP transport are safe for concurrent requests.

## Testing With an Injected Transport

Implement the small `HTTPDoer` interface to test without a real API key or network request:

```go
type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) Do(request *http.Request) (*http.Response, error) {
	return fn(request)
}
```

Pass the implementation with `WithHTTPClient`. The repository tests demonstrate complete mocked success and error responses.

## Development

```bash
gofmt -w .
go vet ./...
go test -race ./...
go list ./...
```

See [CONTRIBUTING.md](./CONTRIBUTING.md) for contribution rules and [PUBLISHING.md](./PUBLISHING.md) for maintainer release instructions.

## Resources

- [MailEngin API documentation](https://mailengin.app/dashboard/docs)
- [Developer Templates](https://mailengin.app/dashboard/dev-templates)
- [API keys](https://mailengin.app/dashboard/api-keys)
- [Security policy](./SECURITY.md)
- [Changelog](./CHANGELOG.md)

## License

Released under the [MIT License](./LICENSE). Copyright 2026 MailEngin.
