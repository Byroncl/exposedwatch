English | [Español](README.es.md)

# exposedwatch

[![CI](https://github.com/Byroncl/exposedwatch/actions/workflows/ci.yml/badge.svg)](https://github.com/Byroncl/exposedwatch/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Byroncl/exposedwatch)](https://github.com/Byroncl/exposedwatch/releases/latest)
[![License](https://img.shields.io/github/license/Byroncl/exposedwatch)](LICENSE)

See what your Linux server exposes to the network, live in your terminal, and get flagged when something risky is open, such as a database bound to `0.0.0.0`.

It reads the host's listening ports (from `/proc/net`) and Docker containers' published ports.

## Quick start

On Ubuntu/Debian (`amd64` or `arm64`):

```sh
curl -fLO "https://github.com/Byroncl/exposedwatch/releases/latest/download/exposedwatch_linux_$(dpkg --print-architecture).deb"
sudo apt install "./exposedwatch_linux_$(dpkg --print-architecture).deb"
sudo exposedwatch
```

`sudo exposedwatch` opens a full-screen, auto-refreshing view (like `htop`). Press `?` for help and `q` to quit.

| Command | What it does |
| ------- | ------------ |
| `exposedwatch` | Interactive full-screen view (same as `exposedwatch tui`) |
| `exposedwatch scan` | One-shot report for scripts and cron (table or JSON) |
| `exposedwatch version` | Print the version (also `--version`) |
| `exposedwatch help` | Show help (also `-h`, `--help`) |

Flags placed first, such as `exposedwatch --interval 5s`, go to the interactive view.

## What it detects

Rules are matched by port. A rule fires when the port is exposed at least as widely as the minimum exposure shown.

| Service | Port(s) | Minimum exposure | Severity |
| ------- | ------- | ---------------- | -------- |
| SSH | 22 | `all-interfaces` | medium |
| Telnet | 23 | `all-interfaces` | critical |
| Database or cache: MySQL, PostgreSQL, Redis, MongoDB, Elasticsearch, Memcached | 3306, 5432, 6379, 27017, 9200, 11211 | `private` | high |
| Docker API (unauthenticated) | 2375 | `all-interfaces` | critical |
| RDP | 3389 | `all-interfaces` | high |
| SMB / NetBIOS | 445, 139 | `all-interfaces` | high |
| Any other port | any | `all-interfaces` | info |

Exposure levels, from narrowest to widest: `loopback`, `private`, `all-interfaces`, `public` (bound to a specific public IP address). Each finding includes a recommendation.

Severities: `info`, `low`, `medium`, `high`, `critical`.

## Interactive view

```sh
sudo exposedwatch                      # refreshes every 2s
sudo exposedwatch --interval 5s        # custom refresh interval (minimum 1s)
sudo exposedwatch --no-docker          # skip Docker discovery
sudo exposedwatch --no-connections     # disable the Connections tab
sudo exposedwatch --theme light        # force the palette for light terminals
```

Flags: `--proc-root`, `--docker-socket`, `--no-docker`, `--interval`, `--no-connections`, `--theme auto|dark|light` (default `auto`, which detects the terminal background). It needs an interactive terminal; otherwise it exits with code 2 (use `scan` instead). A source that fails (for example Docker unavailable) is shown as a warning line and does not stop the view.

The header shows the severity counts, an overall risk gauge (`SAFE`, `LOW`, `MEDIUM`, `HIGH` or `CRITICAL`, driven by the highest severity present), the number of ports per source and the `+N new -M closed` counters. The header also shows `conn N (in X / out Y)` when the width allows it. Four tabs share the screen:

- **Findings** (`1`): one row per finding.
- **All ports** (`2`): every port, with its severity or `-` when no rule matched.
- **Changes** (`3`): a log of ports that opened or closed since the view started, newest first (the last 200 entries are kept).
- **Connections** (`4`): who is connected to this server right now (see below).

The first successful scan is the baseline: nothing is flagged. After each later refresh, ports that opened get a `NEW` badge for 5 refreshes and a log entry; ports that closed get a log entry with their last known severity. `+N new -M closed` counts these events since the view started. A refresh that fails, or a source that fails during a refresh, never produces `closed` events: the previous data of that source is kept. The log lives only in memory and is lost when the view exits. A detail panel shows the rule, recommendation, bind address, process or container and when this session first saw the selected port.

| Key | Action |
| --- | ------ |
| `tab` / `shift+tab` | Next / previous tab |
| `1` / `2` / `3` / `4` | Jump to Findings / All ports / Changes / Connections |
| `up` / `k`, `down` / `j` | Move selection |
| `g` / `G` | Jump to top / bottom |
| `r` | Refresh now |
| `s` | Cycle sort: severity, port, process (Findings and All ports); direction, local port, remote address, process, age (Connections) |
| `f` | Cycle minimum severity filter (Findings and All ports; ignored on Connections) |
| `/` | Free-text filter (`enter` applies, `esc` cancels; on Connections it matches remote address, remote port and process) |
| `l` | Show or hide loopback-to-loopback connections (Connections; hidden by default) |
| `?` | Toggle help |
| `q` / `ctrl+c` | Quit |

### Connections tab

The Connections tab lists the established TCP connections of this host, read from `/proc/net/tcp` and `/proc/net/tcp6` on every refresh. Each row shows the direction (`IN` when the local port is one this host listens on, `OUT` otherwise), the local port (and bind address when there is room), the remote `address:port`, the state, the process (`-` when unknown) and the age: how long this session has been seeing that connection (`12s`, `4m`, `1h`, `2d`). Inbound connections that appeared since the previous refresh get a `NEW` badge for 5 refreshes; the first read is the baseline and flags nothing. Rows are ordered inbound first, then by local port and remote address.

Loopback-to-loopback connections (local services talking to each other) are hidden by default because they are noise; press `l` to show them. The title and the footer say whether they are hidden. Below the table, two panels list the busiest remote addresses and the inbound connection count per local port, computed from the same connections as the table (so they follow `l`). A failed read keeps the previous rows and shows a warning line, and it never delays the port scan, which refreshes independently. If the number of connections reaches the reader's limit, a notice shows how many were read.

Everything stays in memory: remote addresses are shown on screen only, never written to disk or logged, and they are not resolved through DNS.

Limitations:

- Connections are sampled on every refresh, so very short-lived connections between two refreshes can be missed.
- Only the host's own sockets are read. Connections to Docker-published ports that the kernel handles through NAT may not appear.
- Process names need `sudo` (root), as for the process column of the port tabs.
- Only established TCP connections are listed; UDP is not.

## Scripts, cron and CI: `scan`

```sh
sudo exposedwatch scan                       # table output
sudo exposedwatch scan --format json         # machine-readable output
sudo exposedwatch scan --min-severity medium # show and gate on medium and above
sudo exposedwatch scan --no-docker           # skip Docker discovery
```

Example output:

```text
$ exposedwatch scan --no-docker --min-severity high
SEVERITY  PORT  PROTOCOL  BIND ADDRESS  EXPOSURE        SOURCE  PROCESS/CONTAINER  TITLE
high      5432  tcp       0.0.0.0       all-interfaces  host    -                  Database or cache reachable beyond loopback
high      5432  tcp       ::            all-interfaces  host    -                  Database or cache reachable beyond loopback

Recommendations
  - [database-exposed] Bind the datastore to 127.0.0.1 or an internal network and require authentication.

54 port(s) scanned, 2 finding(s).
```

Exit codes (shared by all commands):

| Code | Meaning |
| ---- | ------- |
| 0 | OK, no findings at or above `--min-severity` |
| 1 | Runtime error |
| 2 | Usage error (bad flag or command) |
| 3 | Findings at or above `--min-severity` were found |

Cron or CI example: fail the job when something medium or worse is exposed.

```sh
exposedwatch scan --min-severity medium --format json > exposure.json || echo "exposure check exited with $?"
```

## External scan (nmap)

`--external` also scans this machine from the network side with [nmap](https://nmap.org) and reports what it finds with source `external`, merged with the host and Docker results and evaluated by the same risk rules. It is off by default.

```sh
sudo apt install nmap
sudo exposedwatch scan --external
```

- Default targets: this host's own non-loopback, non-link-local IPv4 addresses. Nothing else is scanned unless you ask.
- Other targets (`--external-target`; comma separated or repeatable) must be IPv4 addresses, IPv4 CIDRs (`/24` or narrower) or hostnames. IPv6 targets are not supported yet. They require `--confirm-ownership`. **Scanning systems you do not own or have no permission to scan may be illegal.**
- TCP connect scan only (`-sT`); UDP is not scanned. Use `--external-ports 22,80,443,8000-8100` to choose ports (default: nmap's top 1000) and `--external-timeout` to bound the run (default 2m).
- nmap must be installed (the `.deb` only recommends it). If it is missing, the external source is reported as an error and the host and Docker results still print.

## Alerts (Telegram)

`exposedwatch watch` runs headless (no terminal needed), scans on an interval and sends a Telegram message when a **new** risky port appears. It is off until you configure it.

**Privacy: read this first.**

- Alerts are sent to Telegram's servers (`api.telegram.org`), a third party. Each message contains this host's name, the port and protocol, the severity, the process name, the rule id and a short recommendation.
- Alerts never contain remote IP addresses: they describe what this machine exposes, not who connects to it.
- This is the only feature that makes an outbound connection, besides the optional nmap scan. `scan` and the interactive view send nothing anywhere.
- The bot token is a secret. The tool shows it only masked (for example `123456789:****`) and removes it from the errors and logs it produces, and it never writes the token anywhere itself. Keep the file you put it in at mode `0600`, or better, keep the token out of the disk in an environment variable.
- Standard proxy environment variables (`HTTPS_PROXY`, `NO_PROXY`) are honored. HTTPS goes through the proxy with `CONNECT`, so the proxy sees the destination host (`api.telegram.org`) but not the request path, which contains the bot token.

### Set up

1. Create a bot: open a chat with [@BotFather](https://t.me/BotFather) in Telegram, send `/newbot` and follow the questions. BotFather gives you the bot token.
2. Find your chat id: send any message to your new bot (for a group, add the bot to the group and write there). Then open `https://api.telegram.org/bot<TOKEN>/getUpdates` with your token and look for `"chat":{"id":...}` in the answer. Group ids are negative numbers.
3. Put the values in the config file or in the environment, then check them:

```sh
exposedwatch notify-test
```

`notify-test` sends one test message and prints either a success line or the failure. It shows the masked token and whether it came from the file or the environment, never the token itself. Exit codes: 0 ok, 1 runtime error, 2 usage error.

### Configuration

The config file is JSON. Its default location is `/etc/exposedwatch/config.json` when you run as root, and `$XDG_CONFIG_HOME/exposedwatch/config.json` (or `~/.config/exposedwatch/config.json`) otherwise. `--config PATH` overrides it.

No command writes this file yet: you create and edit it by hand. As root, create it private before putting any secret in it:

```sh
sudo install -d -m 700 /etc/exposedwatch
sudo install -m 600 /dev/null /etc/exposedwatch/config.json
sudo $EDITOR /etc/exposedwatch/config.json
```

```json
{
  "version": 1,
  "telegram": { "enabled": true, "bot_token": "<token>", "chat_id": "<chat id>" },
  "alerts": { "min_severity": "high", "cooldown_minutes": 60, "interval_seconds": 60 }
}
```

- Use mode `0600` for the file. A looser mode only produces a warning. A symbolic link, or a file not owned by the user running the tool, is refused. As root, a file whose directory can be written by group or others is refused too; other users only get a warning for that.
- `EXPOSEDWATCH_TELEGRAM_TOKEN` and `EXPOSEDWATCH_TELEGRAM_CHAT_ID` override the file, so the token does not have to be on disk. Alerts are sent only when `telegram.enabled` is `true` and both values are set from either place.
- `min_severity` is `info`, `low`, `medium`, `high` or `critical`. `interval_seconds` must be at least 10.

### Watch

```sh
exposedwatch watch                 # loop until SIGINT or SIGTERM
exposedwatch watch --once          # one scan, then exit (for cron)
exposedwatch watch --interval 5m --min-severity critical
```

- The first scan of a run sends one summary ("N findings at or above high already present", at most 5 titles) and remembers the existing findings without alerting each one. After that, each finding that was not known alerts once. A finding that disappears and reappears within `cooldown_minutes` does not alert again. A source that fails during a scan (for example Docker) does not make its findings look closed.
- A message that cannot be delivered is retried with an exponential backoff (starting at the scan interval, doubling up to 15 minutes) for up to 24 hours, and then dropped; a dropped alert makes its finding eligible to alert again. A failed send never stops the loop or blocks newer alerts. Logs go to stderr as one timestamped line each, without tokens, URLs or remote addresses.
- A source that was unreadable during the first scan (for example Docker not started yet) is summarised once ("N existing findings from docker now visible") when it first becomes readable, instead of alerting each finding.
- State is kept in memory only. `--once` has no state between runs, so every cron run re-sends the first-run summary of whatever is open at that moment.
- It needs root only for what the scan needs (process names, Docker socket).

Example systemd unit. **It is an untested example, shown as documentation only** (this project ships no packaging for it). It runs as root. Put the secret in a file readable only by root (`chmod 0600`) and set `"enabled": true` in the config file:

```ini
# /etc/systemd/system/exposedwatch.service
[Unit]
Description=exposedwatch Telegram alerts
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/bin/exposedwatch watch
EnvironmentFile=/etc/exposedwatch/telegram.env
Restart=on-failure
NoNewPrivileges=yes
PrivateTmp=yes
ProtectHome=yes

[Install]
WantedBy=multi-user.target
```

Only conservative hardening is suggested. Stricter sandboxing options may break access to `/proc` (process names), the Docker socket or iptables, so add them only after testing.

```sh
# /etc/exposedwatch/telegram.env  (mode 0600, owner root)
EXPOSEDWATCH_TELEGRAM_TOKEN=<token>
EXPOSEDWATCH_TELEGRAM_CHAT_ID=<chat id>
```

## Install

### Ubuntu/Debian (`.deb`, recommended)

See [Quick start](#quick-start). Verify the download (optional):

```sh
curl -fLO https://github.com/Byroncl/exposedwatch/releases/latest/download/checksums.txt
sha256sum --ignore-missing -c checksums.txt
```

Checksums are in `checksums.txt`; run the check in the directory you downloaded to.

### Tarball

```sh
curl -fLO "https://github.com/Byroncl/exposedwatch/releases/latest/download/exposedwatch_linux_$(dpkg --print-architecture).tar.gz"
tar -xzf "exposedwatch_linux_$(dpkg --print-architecture).tar.gz" exposedwatch
sudo install -m 0755 exposedwatch /usr/local/bin/exposedwatch
```

### With Go

```sh
go install github.com/Byroncl/exposedwatch/cmd/exposedwatch@latest
```

### Uninstall

```sh
sudo apt remove exposedwatch
```

## Notes

- Run with `sudo` so process names can be resolved; without it, ports are still listed but the process column may be empty.
- Docker discovery reads the Docker socket (`/var/run/docker.sock` by default, change with `--docker-socket`). Your user needs access to it, otherwise Docker is reported as a source that could not be read.
- Read-only: it reads `/proc` and sends a single `GET` request to the Docker socket to list containers. It does not open or change ports, and does not modify containers. With `--external`, nmap additionally opens TCP connections to the scanned targets. Only `watch` and `notify-test` make an outbound connection of their own: they send alerts to Telegram (see [Alerts (Telegram)](#alerts-telegram)).
- `--nmap-path` runs the given binary with the same privileges as exposedwatch, so do not let untrusted users pass arguments to `sudo exposedwatch` through sudoers rules.

## License

Apache-2.0. See [LICENSE](LICENSE).
