# Contributing to exposedwatch

Thanks for helping. exposedwatch is a beta project; small, focused pull requests are the easiest to review.

Found a vulnerability? Do not open an issue: follow [SECURITY.md](SECURITY.md).

## Setup

Use the Go version declared in `go.mod`.

```sh
go build ./cmd/exposedwatch
```

## Checks every PR must pass

```sh
gofmt -l .              # must print nothing
go vet ./...
go test ./... -count=1
```

Also run `go test -race ./...` (or the affected packages) when you touch concurrent code.

## Branch model

- Long-lived branches: `main` (production), `staging` (pre-production), `development` (integration).
- Work branches: `FEATURE|BUGFIX|HOTFIX/<TICKET>-<n>/<kebab-description>`, for example `FEATURE/EXW-026/prerelease-pack`.
- Open pull requests against `development`. Never push directly to `main`, `staging` or `development`.
- Promotion goes `development` -> `staging` -> `main` with merge commits.

## Commits

Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`).

## Architecture map

The project is hexagonal. Dependencies point inward.

| Package | Role |
| ------- | ---- |
| `internal/domain` | Core types and rules, no I/O |
| `internal/ports` | Interfaces the application needs |
| `internal/app` | Use cases and services |
| `internal/adapters/*` | Implementations: `procnet`, `dockerapi`, `iptablessave`, `nmapscan`, `telegram`, `configfile`, `baselinefile` |
| `internal/report` | Table and JSON output |
| `internal/risk` | Risk rules and severities |
| `internal/tui` | Interactive view; it must not import adapters |

## Code quality rules

- Descriptive identifiers; no one or two character names except `ctx` and `err`.
- Named constants live in `constants*.go` inside each package. No central constants or utils package.
- Use named types instead of bare primitives where a concept deserves a name.
- One responsibility per file.
- Test first: write the failing test, then the code.
- Keep pull requests small.

## Tests and privacy

- Use fake tokens only.
- Use documentation IPs only: `203.0.113.0/24`, `198.51.100.0/24`, `192.0.2.0/24`. Never real addresses, host names, chat ids or secrets, in tests, fixtures, issues or pull requests.

## Security-sensitive packages

Take extra care in `iptablessave`, `nmapscan`, `telegram`, `configfile` and `baselinefile`. Changes there need adversarial tests: hostile input, symbolic links, FIFOs and special files, size and count bounds, and token redaction.

## License

By contributing you agree that your contributions are licensed under the Apache License 2.0, the same as the project (see [LICENSE](LICENSE)).
