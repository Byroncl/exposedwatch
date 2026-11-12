# Security Policy

exposedwatch is a beta (v0.x) tool that can run as root. Please report vulnerabilities privately and responsibly.

## Supported versions

The project is pre-1.0. Only the latest release receives security fixes.

## Reporting a vulnerability

Use GitHub's private vulnerability reporting:

<https://github.com/Byroncl/exposedwatch/security/advisories/new>

If that option is not available to you, open a minimal issue asking for a private channel. **Do not include any exploit details, proof of concept, tokens or addresses in that issue.** We will try to arrange a private way to receive the details.

Please do not report vulnerabilities in public issues, pull requests or discussions.

## What is in scope

The `exposedwatch` binary and the packages it ships, in particular:

- Execution of `iptables-save`, `ip6tables-save` and `nft`, and of the optional `nmap` scan.
- Handling of the config file and the baseline file, including the trust checks (symbolic links, ownership, permissions, special files, size limits).
- Handling of the Telegram bot token (masking, redaction from errors and logs, storage).
- The interactive terminal view (TUI).

## What to include in a report

- The exposedwatch version (`exposedwatch version`) and how it was installed.
- Operating system, architecture and whether it ran as root.
- A clear description of the impact and the steps to reproduce it.
- Any relevant configuration, with secrets, real IP addresses and host names removed.

## What to expect

This is a small project maintained on a best-effort basis. We aim to acknowledge reports and keep you informed, but we do not guarantee any response or fix time.

## Coordinated disclosure

Please give us a reasonable chance to investigate and release a fix before you disclose the issue publicly. If you would like to be credited in the release notes, say so in your report.

## Safe testing

Only scan or test systems you own or are explicitly authorized to test. Use fake tokens and documentation IP ranges (203.0.113.0/24, 198.51.100.0/24, 192.0.2.0/24) in reports.

## What the tool does and does not do

- It is read-only toward the firewall: it only runs listing commands (`iptables-save`, `ip6tables-save`, `nft list tables`). It never changes firewall rules.
- It runs with the privileges of the user who starts it, which is often root.
- The only outbound network use is opt-in: Telegram alerts (`watch`, `notify-test` and the test message in Settings) and the external scan with `nmap` (`--external`).
- Secrets such as the Telegram bot token are shown masked and are redacted from errors and logs. This is covered by tests, but it is not a formal guarantee: report any leak you find.
