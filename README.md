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
```

Flags: `--proc-root`, `--docker-socket`, `--no-docker`, `--interval`. It needs an interactive terminal; otherwise it exits with code 2 (use `scan` instead). A source that fails (for example Docker unavailable) is shown as a warning line and does not stop the view.

| Key | Action |
| --- | ------ |
| `up` / `k`, `down` / `j` | Move selection |
| `g` / `G` | Jump to top / bottom |
| `r` | Refresh now |
| `s` | Cycle sort: severity, port, process |
| `f` | Cycle minimum severity filter |
| `/` | Free-text filter (`enter` applies, `esc` cancels) |
| `?` | Toggle help |
| `q` / `ctrl+c` | Quit |

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
