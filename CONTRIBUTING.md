# Contributing to the MailEngin Go SDK

Thank you for improving the MailEngin SDK. Contributions should preserve Go 1.22 compatibility, standard-library-only runtime code, and behavior shared by the official SDK suite.

## Before You Start

- Use Go 1.22 or newer.
- Search existing issues and pull requests before opening a duplicate.
- Open an issue before a large public API change.
- Report security vulnerabilities privately according to [SECURITY.md](./SECURITY.md).

## Local Setup

```bash
gofmt -w .
go vet ./...
go test -race ./...
go list ./...
```

## Making Changes

1. Create a focused branch from the latest `main`.
2. Keep runtime code dependency-free unless a dependency is clearly justified.
3. Preserve exported names and module compatibility whenever possible.
4. Add tests for affected validation, JSON mapping, headers, errors, context cancellation, or transport behavior.
5. Update the README for any developer-visible behavior.
6. Add a concise entry under `Unreleased` in `CHANGELOG.md`.

Do not add automatic send retries until MailEngin supports idempotency keys.

## Pull Request Checklist

- [ ] `gofmt` produces no diff.
- [ ] `go vet ./...` and `go test -race ./...` pass.
- [ ] New behavior uses an injected `HTTPDoer` and needs no real API key.
- [ ] Exported type or option changes are documented.
- [ ] The changelog is updated.
- [ ] No API key, `.env` file, customer data, credential, or production log is included.

Maintainers may request changes to preserve cross-SDK behavior or module compatibility. By contributing, you agree that your work is released under the repository's MIT License.
