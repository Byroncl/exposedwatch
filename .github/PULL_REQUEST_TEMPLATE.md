## What and why

<!-- What does this change and why? Link the ticket. -->

## Checklist

- [ ] Tests were written first and cover the change
- [ ] `gofmt -l .` prints nothing, `go vet ./...` and `go test ./... -count=1` pass
- [ ] README.md and README.es.md are updated if behavior changed
- [ ] No secrets, tokens or real IP addresses (documentation IPs only)
- [ ] If a security-sensitive package changed (`iptablessave`, `nmapscan`, `telegram`, `configfile`, `baselinefile`): adversarial tests were added
