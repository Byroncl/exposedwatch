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
sudo exposedwatch --no-connections     # disable the Connections module
sudo exposedwatch --theme light        # force the palette for light terminals
```

Flags: `--proc-root`, `--docker-socket`, `--no-docker`, `--interval`, `--no-connections`, `--theme auto|dark|light` (default `auto`, which detects the terminal background). It needs an interactive terminal; otherwise it exits with code 2 (use `scan` instead). A source that fails (for example Docker unavailable) is shown as a warning line and does not stop the view.

The header shows the severity counts, an overall risk gauge (`SAFE`, `LOW`, `MEDIUM`, `HIGH` or `CRITICAL`, driven by the highest severity present), the number of ports per source and the `+N new -M closed` counters. The header also shows `conn N (in X / out Y)` when the width allows it. On the Overview the header is just the title, because the Overview shows those numbers itself.

The screen has a sidebar of modules on the left and, on the right, the tabs of the current module followed by its view:

| Module | Tabs (`1`, `2`, ...) | What it shows |
| ------ | -------------------- | ------------- |
| Overview | Summary | Overall risk gauge, severity counts, counters, the 5 worst findings, the last 5 changes and the top 3 remote addresses (inbound, only when connections are enabled) |
| Ports | Findings, All ports, Changes | `Findings`: one row per finding. `All ports`: every port, with its severity or `-` when no rule matched. `Changes`: a log of ports that opened or closed since the view started, newest first (the last 200 entries are kept) |
| Connections | Live, Top remote | `Live`: who is connected to this server right now (see below). `Top remote`: the busiest inbound remote addresses and the inbound count per local port, full size |
| Firewall | - | Not available yet (placeholder) |
| Alerts | - | Not available yet (placeholder) |
| Settings | - | Not available yet (placeholder); it only shows the current theme, refresh interval and whether connections are enabled, read-only |

The view starts on Ports, Findings. Each module remembers the tab you last used in it. Firewall, Alerts and Settings only say what they are planned to do; nothing in them works yet.

The sidebar adapts to the terminal width: full (module names with a count or flag, for example the number of findings and `NEW` on Ports, the connection count on Connections, `off` on Alerts) from 80 columns, compact (one letter per module) from 50 to 79 columns, and hidden below 50. `b` hides or shows it on top of that. Nothing is drawn outside the terminal at any size.

The first successful scan is the baseline: nothing is flagged. After each later refresh, ports that opened get a `NEW` badge for 5 refreshes and a log entry; ports that closed get a log entry with their last known severity. `+N new -M closed` counts these events since the view started. A refresh that fails, or a source that fails during a refresh, never produces `closed` events: the previous data of that source is kept. The log lives only in memory and is lost when the view exits. A detail panel shows the rule, recommendation, bind address, process or container and when this session first saw the selected port.

| Key | Action |
| --- | ------ |
| `tab` / `shift+tab` | Next / previous module |
| `]` / `[` | Next / previous tab of the current module |
| `1` to `9` | Jump to that tab of the current module (Ports: Findings / All ports / Changes; Connections: Live / Top remote) |
| `b` | Show or hide the sidebar |
| `up` / `k`, `down` / `j` | Move selection (tables) |
| `g` / `G` | Jump to top / bottom (tables) |
| `r` | Refresh now |
| `s` | Cycle sort: severity, port, process (Findings and All ports); direction, local port, remote address, process, age (Connections / Live) |
| `f` | Cycle minimum severity filter (Findings and All ports only) |
| `/` | Free-text filter (`enter` applies, `esc` cancels; on Connections / Live it matches remote address, remote port and process) |
| `l` | Show or hide loopback-to-loopback connections (Connections: Live and Top remote; hidden by default) |
| `i` | Show only inbound connections, or all of them (Connections: Live and Top remote; all shown by default) |
| `?` | Toggle help |
| `q` / `ctrl+c` | Quit |

Keys that do not apply to the current tab are ignored and their hints are not shown in the footer. When the footer is too narrow, the least important hints are dropped first.

### Connections module (Live tab)

The Live tab of the Connections module lists the established TCP connections of this host, read from `/proc/net/tcp` and `/proc/net/tcp6` on every refresh. Each row shows the direction (`IN` when the local port is one this host listens on, `OUT` otherwise), the local port (and bind address when there is room), the remote `address:port`, the state, the process (`-` when unknown) and the age: how long this session has been seeing that connection (`12s`, `4m`, `1h`, `2d`). Inbound connections that appeared since the previous refresh get a `NEW` badge for 5 refreshes; the first read is the baseline and flags nothing. Rows are ordered inbound first, then by local port and remote address.

Loopback-to-loopback connections (local services talking to each other) are hidden by default because they are noise; press `l` to show them. The title and the footer say whether they are hidden. Below the table, two panels list the busiest remote addresses and the inbound connection count per local port, computed from the same connections as the table (so they follow `l` and `i`). A failed read keeps the previous rows and shows a warning line, and it never delays the port scan, which refreshes independently. If the number of connections reaches the reader's limit, a notice shows how many were read.

Press `i` to keep only the inbound (`IN`) connections; the title and the footer show `inbound:only` or `inbound:all`. It combines with `l`, the `/` filter and every sort order. The table, the tab count and the header counter `conn N (in X / out Y)` all describe what is visible, so with `i` on the counter reads `out 0`; the two panels already list inbound traffic only. Your selection stays on the same row when that row is still visible.

Everything stays in memory: remote addresses are shown on screen only, never written to disk or logged, and they are not resolved through DNS.

#### How to read it

A connection here is an open TCP session between this machine and another one. Only `established` sessions are listed, that is, conversations that are happening right now.

| Direction | What it means | Does it matter for security? |
| --------- | ------------- | ---------------------------- |
| `IN` | Someone else connected to a port this host listens on. | Yes, this is the traffic to review: it is how other machines reach your services. |
| `OUT` | This host connected to someone else's service. | Usually not by itself. It shows what your programs talk to. |

- Many `OUT` rows are normal on a desktop: browser, chat, sync and update clients keep several connections open. Their remote addresses often belong to large cloud or CDN providers. They are shown as numbers, not names, because exposedwatch does no DNS lookup on purpose.
- Look first at `IN` rows coming from public addresses to ports you did not expect, for example a database or SSH.
- Tips: `i` shows inbound only, `s` changes the sort, `/` filters by process or address, `l` shows loopback traffic, and `sudo` is needed to see process names.

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
- Read-only: it reads `/proc` and sends a single `GET` request to the Docker socket to list containers. It does not open or change ports, and does not modify containers. With `--external`, nmap additionally opens TCP connections to the scanned targets.
- `--nmap-path` runs the given binary with the same privileges as exposedwatch, so do not let untrusted users pass arguments to `sudo exposedwatch` through sudoers rules.

## License

Apache-2.0. See [LICENSE](LICENSE).
