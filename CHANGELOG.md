# Changelog

All notable changes to this module will be documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow [Semantic Versioning](https://semver.org/).

## Unreleased

## [0.2.0] - 2026-08-31

### Changed

- Move the canonical module path to `github.com/mailengin/mailengin-go-sdk` under the MailEngin GitHub organization.
- Set the SDK user agent version to `mailengin-go/0.2.0` for the organization-path release.

### Fixed

- Apply canonical `gofmt` formatting to all Go source and test files.
- Make CI formatting failures report the affected filenames.
- Retry Go proxy registration while a new public tag propagates.

## [0.1.0] - 2026-08-31

### Added

- Initial dependency-free Go SDK with context-aware requests.
- Typed single and personalized bulk email operations.
- Template, raw HTML, variables, sender override, and reply-routing support.
- Structured API, timeout, cancellation, malformed-response, and network errors.
- Injectable HTTP transport, race-tested CI, and Go proxy registration workflow.
