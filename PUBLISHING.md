# Publishing the Go SDK

This guide is for MailEngin maintainers. Developers installing the module should use [README.md](./README.md).

Module: `github.com/mailengin/mailengin-go-sdk`  
Repository: `mailengin/mailengin-go-sdk`  
Discovery: [pkg.go.dev](https://pkg.go.dev/github.com/mailengin/mailengin-go-sdk)

## How Go Publication Works

Go modules are published through immutable Git tags. There is no registry upload, publishing account, API token, or package archive to push. The repository path, `module` directive, import path, and release tags are the public package identity.

## One-Time Repository Setup

1. Create the public repository `https://github.com/mailengin/mailengin-go-sdk`.
2. Enable multifactor authentication and protect `main` and release tags.
3. Create a GitHub environment named `release` with a required maintainer reviewer.
4. Confirm `go.mod` contains exactly:

```go
module github.com/mailengin/mailengin-go-sdk
```

Moving the repository or changing the module path later is a breaking change.

## Prepare a Release

Update `Version` in `client.go` and add the release notes to `CHANGELOG.md`.

```bash
test -z "$(gofmt -l .)"
go vet ./...
go test -race ./...
go list ./...
```

Inspect the files tracked by Git:

```bash
git ls-files
```

Confirm there are no API keys, `.env` files, credentials, customer data, binaries, coverage output, or editor files.

## Local Consumer Test

Before publishing, create a temporary consumer module and point it to the local SDK checkout:

```bash
mkdir mailengin-go-release-check
cd mailengin-go-release-check
go mod init example.com/mailengin-release-check
go mod edit -replace github.com/mailengin/mailengin-go-sdk=../mailengin-go-sdk
go get github.com/mailengin/mailengin-go-sdk
go list -m all
go test ./...
```

Compile a small program that calls `mailengin.New` and uses a mocked `HTTPDoer`; never use a customer API key in release checks.

## Tag and Publish

The tag must use the `vMAJOR.MINOR.PATCH` format required by Go modules:

```bash
git add client.go CHANGELOG.md
git commit -m "Release Go SDK v0.1.0"
git push origin main
git tag -a v0.1.0 -m "MailEngin Go SDK v0.1.0"
git push origin v0.1.0
```

The workflow reruns formatting, vet, and race-enabled tests after `release` approval. It then requests the tagged version from `proxy.golang.org`, which makes the module discoverable to the Go ecosystem.

Proxy propagation can lag briefly behind a new Git tag. If the final proxy request fails while all verification steps pass, wait a few minutes and rerun the workflow job or run the verification command manually. Do not move or recreate the tag.

## Verify the Public Release

Request the exact version from the public proxy:

```bash
GOPROXY=https://proxy.golang.org go list -m github.com/mailengin/mailengin-go-sdk@v0.1.0
```

Then verify from a clean module without a local `replace` directive:

```bash
mkdir mailengin-go-public-check
cd mailengin-go-public-check
go mod init example.com/mailengin-public-check
GOPROXY=https://proxy.golang.org go get github.com/mailengin/mailengin-go-sdk@v0.1.0
go list -m all
```

Confirm the version appears on pkg.go.dev with the README, license, package documentation, exported types, and source links.

## Major Versions

Starting with v2, Go requires the major version in both the module and import path:

```go
module github.com/mailengin/mailengin-go-sdk/v2
```

Consumers then import `github.com/mailengin/mailengin-go-sdk/v2`. Plan this change before publishing any v2 tag.

## Release Recovery

Never move, delete, or recreate a published Go tag. The Go checksum database records module contents and will reject changed bytes for an existing version.

For a defect, publish a corrected patch version. If credentials or sensitive data were committed, treat the incident as exposed even if the tag is deleted; rotate the secret and follow the security process.

Transfer repository ownership only when the final public URL remains compatible. If the path changes, treat it as a new module and document migration explicitly.

Official reference: [Publishing Go modules](https://go.dev/doc/modules/publishing).

See the workspace [release handbook](../PUBLISHING.md) for the shared checklist and incident process.
