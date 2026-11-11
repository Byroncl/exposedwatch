# exposedwatch

A small console tool that shows what a Linux server exposes to the network.
It reads the host's listening ports (from `/proc/net`) and Docker containers' published ports, then flags risky exposure, such as a database bound to `0.0.0.0`.

```text
$ exposedwatch scan --no-docker --min-severity low
SEVERITY  PORT  PROTOCOL  BIND ADDRESS  EXPOSURE        SOURCE  PROCESS/CONTAINER  TITLE
high      5432  tcp       0.0.0.0       all-interfaces  host    -                  Database or cache reachable beyond loopback
high      5432  tcp       ::            all-interfaces  host    -                  Database or cache reachable beyond loopback

Recommendations
  - [database-exposed] Bind the datastore to 127.0.0.1 or an internal network and require authentication.

52 port(s) scanned, 2 finding(s).
```

> **Release status:** no release has been published yet. The install commands below work once the first `v*` tag is released on GitHub.

## Install on Ubuntu/Debian

Download the `.deb` for your machine architecture (`amd64` or `arm64`) from the latest release and install it:

```sh
curl -fLO "https://github.com/Byroncl/exposedwatch/releases/latest/download/exposedwatch_linux_$(dpkg --print-architecture).deb"
sudo apt install "./exposedwatch_linux_$(dpkg --print-architecture).deb"
exposedwatch version
```

Verify the download (optional):

```sh
curl -fLO https://github.com/Byroncl/exposedwatch/releases/latest/download/checksums.txt
sha256sum --ignore-missing -c checksums.txt
```

Checksums are in `checksums.txt`; the file names in it carry the release's naming, so run the check in the directory you downloaded to.

### Alternatives

Tarball:

```sh
curl -fLO "https://github.com/Byroncl/exposedwatch/releases/latest/download/exposedwatch_linux_$(dpkg --print-architecture).tar.gz"
tar -xzf "exposedwatch_linux_$(dpkg --print-architecture).tar.gz" exposedwatch
sudo install -m 0755 exposedwatch /usr/local/bin/exposedwatch
```

With Go:

```sh
go install github.com/Byroncl/exposedwatch/cmd/exposedwatch@latest
```

### Uninstall

```sh
sudo apt remove exposedwatch
```

## Usage

```sh
sudo exposedwatch scan                       # table output
sudo exposedwatch scan --format json         # machine-readable output
sudo exposedwatch scan --min-severity medium # show and gate on medium and above
sudo exposedwatch scan --no-docker           # skip Docker discovery
```

Severities: `info`, `low`, `medium`, `high`, `critical`.

Exit codes:

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

### Notes

- Run with `sudo` so process names can be resolved; without it, ports are still listed but the process column may be empty.
- Docker discovery reads the Docker socket (`/var/run/docker.sock` by default, change with `--docker-socket`). Your user needs access to it, otherwise Docker is reported as a source that could not be read.

## Status and roadmap

Available now: `exposedwatch scan` (host ports and Docker containers, risk rules, table/JSON output, exit codes).

Planned, **not yet available**:

- Interactive terminal UI (TUI)
- External scan with nmap
- Signed apt repository

## License

Apache-2.0. See [LICENSE](LICENSE).
