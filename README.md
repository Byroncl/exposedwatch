English | [Español](README.es.md)

# exposedwatch

[![CI](https://github.com/Byroncl/exposedwatch/actions/workflows/ci.yml/badge.svg)](https://github.com/Byroncl/exposedwatch/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Byroncl/exposedwatch)](https://github.com/Byroncl/exposedwatch/releases/latest)
[![License](https://img.shields.io/github/license/Byroncl/exposedwatch)](LICENSE)

> **Status: beta (v0.x). Not yet audited by an independent human security reviewer.** See [SECURITY.md](SECURITY.md) to report a vulnerability.

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
| `exposedwatch baseline save` | Record the exposed ports as a baseline; use it with `scan --baseline` (see [Baseline mode](#baseline-mode-cron)) |
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
sudo exposedwatch --no-firewall        # disable the Firewall module (no firewall reads)
sudo exposedwatch --theme light        # force the palette for light terminals
exposedwatch --config ~/alerts.json    # config file used by Settings and Alerts (default path below)
```

Flags: `--proc-root`, `--docker-socket`, `--no-docker`, `--interval`, `--no-connections`, `--no-firewall`, `--theme auto|dark|light` (default `auto`, which detects the terminal background), `--config PATH` (the config file that Settings and Alerts read and write; same meaning as for `watch` and `notify-test`). It needs an interactive terminal; otherwise it exits with code 2 (use `scan` instead). A source that fails (for example Docker unavailable) is shown as a warning line and does not stop the view.

The header shows the severity counts, an overall risk gauge (`SAFE`, `LOW`, `MEDIUM`, `HIGH` or `CRITICAL`, driven by the highest severity present), the number of ports per source and the `+N new -M closed` counters. The header also shows `conn N (in X / out Y)` when the width allows it. On the Overview the header is just the title, because the Overview shows those numbers itself.

The screen has a sidebar of modules on the left and, on the right, the tabs of the current module followed by its view:

| Module | Tabs (`1`, `2`, ...) | What it shows |
| ------ | -------------------- | ------------- |
| Overview | Summary | Overall risk gauge, severity counts, counters, the 5 worst findings, the last 5 changes and the top 3 remote addresses (inbound, only when connections are enabled) |
| Ports | Findings, All ports, Changes | `Findings`: one row per finding. `All ports`: every port, with its severity or `-` when no rule matched. `Changes`: a log of ports that opened or closed since the view started, newest first (the last 200 entries are kept) |
| Connections | Live, Top remote | `Live`: who is connected to this server right now (see below). `Top remote`: the busiest inbound remote addresses and the inbound count per local port, full size |
| Firewall | Status, Ports | Which listening ports the host firewall lets through, as information only (see [Firewall verdicts](#firewall-verdicts)). `Status`: backend, whether the rules could be read, the notes of the engine, a count per verdict, when it was last evaluated and what each verdict means. `Ports`: one row per host or Docker port with its verdict, most worrying first |
| Alerts | Status, Preview | `Status`: whether Telegram is configured (token masked, chat id only reported as set), enabled or disabled, minimum severity, cooldown, check interval, config file and its warnings, and the result of the last test message of this session. `Preview`: the current findings at or above the saved minimum severity, as "would alert now". The TUI never sends alerts itself: `exposedwatch watch` sends them, and both `notify-test` and the `[ Send test message ]` row send a test message to Telegram (see [Settings > Telegram and the Alerts module](#settings--telegram-and-the-alerts-module)) |
| Settings | Telegram, Display | `Telegram`: an editable form for the alert configuration (see below). `Display`: the current theme, refresh interval and whether connections and the firewall are enabled, read-only |

The view starts on Ports, Findings. Each module remembers the tab you last used in it. The Overview also shows one line `Alerts: on (min high)` or `Alerts: off (min high)` (`n/a` when the config file cannot be loaded).

The sidebar adapts to the terminal width: full (module names with a count or flag, for example the number of findings and `NEW` on Ports, the connection count on Connections, the number of ports the firewall allows or restricts on Firewall (`n/a` when the rules cannot be read, `off` with `--no-firewall`), `on` when alerts are enabled and configured, `off` when they are disabled or not configured, `n/a` when the config file cannot be loaded, on Alerts) from 80 columns, compact (one letter per module) from 50 to 79 columns, and hidden below 50. `b` hides or shows it on top of that. Nothing is drawn outside the terminal at any size.

The first successful scan is the baseline: nothing is flagged. After each later refresh, ports that opened get a `NEW` badge for 5 refreshes and a log entry; ports that closed get a log entry with their last known severity. `+N new -M closed` counts these events since the view started. A refresh that fails, or a source that fails during a refresh, never produces `closed` events: the previous data of that source is kept. The log lives only in memory and is lost when the view exits. A detail panel shows the rule, recommendation, bind address, process or container and when this session first saw the selected port.

| Key | Action |
| --- | ------ |
| `tab` / `shift+tab` | Next / previous module |
| `]` / `[` | Next / previous tab of the current module |
| `1` to `9` | Jump to that tab of the current module (Ports: Findings / All ports / Changes; Connections: Live / Top remote; Firewall: Status / Ports; Alerts: Status / Preview; Settings: Telegram / Display) |
| `b` | Show or hide the sidebar |
| `up` / `k`, `down` / `j` | Move selection (tables) |
| `g` / `G` | Jump to top / bottom (tables) |
| `enter` | Settings > Telegram: toggle or cycle the selected row, start editing a value, or run `[ Save ]` / `[ Send test message ]` |
| while editing | Settings > Telegram: type or paste the value; `enter` saves the value, `esc` cancels, `backspace` deletes. No other key acts, so any character can be typed (`ctrl+c` still quits) |
| `r` | Refresh now (also re-reads the firewall rules) |
| `s` | Cycle sort: severity, port, process (Findings and All ports); direction, local port, remote address, process, age (Connections / Live); verdict, port, process (Firewall / Ports) |
| `f` | Cycle minimum severity filter (Findings and All ports only) |
| `/` | Free-text filter (`enter` applies, `esc` cancels; on Connections / Live it matches remote address, remote port and process; on Firewall / Ports it matches port, process, verdict and reason) |
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

### Firewall verdicts

The Firewall module, and a `Firewall:` line in the details of the port tables, say whether the host firewall would let a NEW inbound connection from an external source reach each listening port: `allowed`, `restricted` (only some source addresses), `blocked`, `unknown` (could not be determined), `not-exposed` (loopback only) or `docker` (published by Docker, which usually bypasses the INPUT and ufw rules, so it is never reported as blocked). The Ports tables also have a low-priority `FW` column that is the first one dropped when the terminal is narrow, and the Overview has a one-line summary. Verdicts are **informational**: they never change a severity, a finding or an exit code.

The rules are read in the background with their own timeout: after the first scan, on `r`, when the set of listening ports changes and otherwise at most every 30 seconds (`iptables-save` is not free), independently from the refresh interval. A failed read keeps the previous data and shows a warning on the Firewall views. Without root the module is simply `n/a`: every exposed port is `unknown` and the notes say why. `--no-firewall` turns the module off and the firewall is never read.

#### Limitations

- It needs root: without it the rules cannot be read and every exposed port is `unknown`.
- It reads `iptables-save` and `ip6tables-save` (found in `/usr/sbin` and `/sbin`, which are not in a normal user's `PATH`). That covers ufw, plain iptables and the iptables-nft shim. Hosts that only use native nftables are reported as `unknown`; a hand-made native `table ip filter` is not detected.
- Docker-published ports are labelled `docker`, never `blocked`.
- A verdict is an estimate for a new inbound connection from an external source. Anything the engine does not understand makes the verdict `unknown`.
- It was validated against fixtures only, not against real rule sets on real machines.

## Scripts, cron and CI: `scan`

```sh
sudo exposedwatch scan                       # table output
sudo exposedwatch scan --format json         # machine-readable output
sudo exposedwatch scan --min-severity medium # show and gate on medium and above
sudo exposedwatch scan --no-docker           # skip Docker discovery
sudo exposedwatch scan --firewall            # also show the firewall verdict of each port
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

`--firewall` is opt-in. It adds a `FIREWALL` column with the verdict word of each port and a short `Firewall` section with the backend, whether the rules could be read and the notes of the engine; with `--format json` it adds a top-level `firewall` object (`available`, `backend`, `notes` and `ports` with `port`, `protocol`, `bind_address`, `source`, `verdict` and `reason`; the Docker verdict is `docker-published` in JSON). Not being able to read the firewall (not root, no tool) is not an error: the section says so and every exposed port is `unknown`. Verdicts are informational: severities and exit codes are exactly the same with and without `--firewall`. See the [limitations](#limitations).

Exit codes (shared by all commands):

| Code | Meaning |
| ---- | ------- |
| 0 | OK, no findings at or above `--min-severity` |
| 1 | Runtime error |
| 2 | Usage error (bad flag or command) |
| 3 | Findings at or above `--min-severity` were found |
| 4 | Drift: new ports compared with a baseline (only `scan --baseline`, which never uses 3; see [Baseline mode](#baseline-mode-cron)) |

Cron or CI example: fail the job when something medium or worse is exposed.

```sh
exposedwatch scan --min-severity medium --format json > exposure.json || echo "exposure check exited with $?"
```

## Baseline mode (cron)

A baseline is a list of the ports you have accepted as exposed. `scan --baseline` compares what is exposed now with that list and reports only the difference, so a cron job is quiet until something new appears. Both commands are information only: they never open, close or change anything on the system.

```sh
sudo exposedwatch baseline save              # record the ports exposed right now
sudo exposedwatch scan --baseline /var/lib/exposedwatch/baseline.json
```

`baseline save` runs the host scan and, unless `--no-docker` is given, the Docker scan (never the external nmap scan), writes the baseline and prints one line: `Baseline saved to <path> (<N> ports)`. Flags: `--file`, `--proc-root`, `--docker-socket`, `--no-docker`, `--allow-partial`, `--force`.

`scan --baseline FILE` works with `--format table|json`, `--no-docker`, `--proc-root`, `--docker-socket`, `--firewall`, `--external` and `--min-severity`. It prints the normal report plus a `Drift` section (or, with `--format json`, a top-level `drift` object with `baseline_file`, `created_at`, `new`, `closed` and `notes`; empty lists are `[]`). Without `--baseline` the output is exactly the same as before.

Exit codes with `--baseline`:

| Code | Meaning |
| ---- | ------- |
| 0 | No drift: no new port at or above `--min-severity` (also when only closed ports differ) |
| 1 | Runtime error, including a missing, untrusted, corrupt or unreadable baseline (run `exposedwatch baseline save`), or a source (host, Docker, external) that could not be read during the scan while no new port reached `--min-severity` |
| 2 | Usage error |
| 4 | Drift: at least one NEW port reaches `--min-severity`; it wins over 1 when both apply |

Exit code 3 is **not** used in baseline mode, because accepted ports are expected to stay exposed. A new port counts at the severity of its most severe finding; a port without any finding (for example one bound to loopback) counts as `info`, so with the default `--min-severity info` every new port counts. Ports that are in the baseline but no longer exposed are listed as `CLOSED` and are informational: closed-only drift exits 0. A source that fails during the scan makes the result partial, so cron never sees a silent clean run: stderr gets an `unverified sources: <names>` line, the Drift section and the JSON `notes` say so (always, whether or not the baseline had entries for that source) and the exit code is 1, unless new ports at or above `--min-severity` were found, in which case it is 4.

Cron example (adapt the path, the schedule and the alert command to your system):

```sh
# Example only: every hour, run the comparison and react when it exits with a non-zero code.
0 * * * * root exposedwatch scan --baseline /var/lib/exposedwatch/baseline.json --no-docker >/dev/null || logger -t exposedwatch "drift or error: exit code $?"
```

Because a non-zero code can also be a runtime error (1), check the exit code, not just success, when you need to tell drift (4) from a broken baseline.

What the baseline holds and where it lives:

- The identity of a port is protocol + port + bind address + source (`host` or `docker`). A service that moves from `127.0.0.1` to `0.0.0.0` is a NEW port, because its exposure changed.
- The file is JSON with a `version`, `created_at` (UTC), `tool_version` and the `entries`, sorted. It holds no process names, container names, remote addresses or host name, so it is safe to copy around.
- Default path: `/var/lib/exposedwatch/baseline.json` as root; otherwise `$XDG_STATE_HOME/exposedwatch/baseline.json` (when it is an absolute path) or `~/.local/state/exposedwatch/baseline.json`. `--file` overrides it. The directory is created with mode 0700 and the file with 0600, written atomically (a failed save keeps the previous file).
- Loading is strict, because a tampered baseline would hide new ports. The file is refused when it is a symbolic link, changed while it was opened, is not owned by the user running the command (as root: not owned by root), or can be written by group or others; it must also be a regular file (a FIFO, device or directory is refused without blocking). As root it is also refused when its directory can be written by group or others or is not owned by root; as another user that only prints a warning (the directory must be owned by that user or by root). Only the immediate directory is checked, not its ancestors, so keep the state directory under root-owned parents. The file is read up to 4 MiB; unknown fields, a version other than 1, invalid entries and more than 20000 entries are rejected, and duplicate entries collapse. Error messages never print the file contents.
- If a source cannot be read while saving (for example Docker is not running), `baseline save` refuses to write a partial baseline, which would hide the whole source; pass `--allow-partial` to save anyway. If a source cannot be read during `scan --baseline`, its ports are unknown, not closed: its baseline entries are not reported as closed (see the exit codes above).
- `baseline save` refuses to overwrite an existing path that is not a baseline: the target must be missing, or a regular file (never a symbolic link) that loads as a trusted, valid baseline; otherwise it exits 1 with `refusing to overwrite <path>: it is not a baseline file (use --force to replace it anyway)`. `--force` is dangerous: it replaces a regular file even if its content is not a baseline, and it still never follows a symbolic link or replaces a directory.
- Bind addresses are stored in canonical form (lowercase, IPv4-mapped IPv6 written as IPv4, no IPv6 zone); `0.0.0.0` and `::` stay different entries.

`exposedwatch watch --baseline FILE` uses the same baseline in the headless alert loop: see [Drift-only alerts with a baseline](#drift-only-alerts-with-a-baseline).

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
- This is the only feature that makes an outbound connection, besides the optional nmap scan. `scan` and the interactive view send nothing anywhere, except the test message you start yourself from Settings > Telegram (it goes to Telegram too).
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

You can edit this file by hand, or from the interactive view (see [Settings > Telegram and the Alerts module](#settings--telegram-and-the-alerts-module)). If you edit it by hand as root, create it private before putting any secret in it:

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

### Settings > Telegram and the Alerts module

In the interactive view, Settings > Telegram edits the same config file that `watch` and `notify-test` read (`--config PATH` selects another one):

| Row | How it works |
| --- | ------------ |
| Enabled | `enter` toggles it |
| Bot token | `enter` starts typing (or pasting) a new token; it starts empty and every typed character is shown as a bullet. When not editing, only the masked form (for example `123456789:****`) is shown. Confirming an empty value keeps the current token: a token can be changed but not cleared from the form (remove it by editing the file) |
| Chat ID | `enter` starts editing |
| Minimum severity | `enter` cycles info, low, medium, high, critical |
| Cooldown (minutes), Check interval (seconds) | `enter` starts editing; digits only (other characters are ignored and a `digits only` hint is shown) |
| `[ Save ]` | Writes the file |
| `[ Send test message ]` | Sends one test message to Telegram with the values currently in the form, saved or not, and shows `Sending...` and then `Sent` or the failure. It needs a token and a chat id. One test runs at a time and the view keeps working while it is in flight (the whole attempt is limited to 15 seconds) |

- Move between rows with `j`/`k` or the arrows. While a value is being typed, no other key is interpreted: `q`, `tab`, `b`, `?` and digits are text. `enter` keeps the value and `esc` cancels it. Pasting works (control characters are dropped). `ctrl+c` still quits.
- `unsaved changes` is shown until you save. The limits of cooldown and interval, and every other validation, are the ones of the config file; a rejected save shows the reason and leaves the file untouched. `[ Save ]` does nothing (`No changes to save`) when nothing changed since the last load or save, and `[ Send test message ]` waits 5 seconds between two sends.
- Save writes the file with the same rules as above: a private file (`0600`), written atomically, in a directory it creates with `0700` when it is missing; saving also repairs looser permissions of an older file. A file that is a symbolic link, is not yours, or cannot be loaded is not overwritten, and the form stays read-only with the reason shown.
- Saving from the interactive view rewrites the file with the known settings only: comments and unknown keys are not preserved. If you keep extra keys, edit the file by hand. As root, saving is refused when the config directory is writable by group or others; as a normal user it saves and shows a warning.
- If the token or the chat id comes from `EXPOSEDWATCH_TELEGRAM_TOKEN` / `EXPOSEDWATCH_TELEGRAM_CHAT_ID`, its row is read-only and says `from environment (...)`. Values from the environment are never written to the file; Save keeps whatever the file already had for them.
- `exposedwatch watch` reads the file when it starts and does not reload it: restart `watch` to apply a saved change. The Alerts module shows the saved values, not unsaved edits.
- Privacy: a test message goes to Telegram's servers like any alert (see above). The token is never displayed, logged or included in an error by the view; it only travels to the config file when you save it and to Telegram when you send a test message.

### Watch

```sh
exposedwatch watch                 # loop until SIGINT or SIGTERM
exposedwatch watch --once          # one scan, then exit (for cron)
exposedwatch watch --interval 5m --min-severity critical
```

- The first scan of a run sends one summary ("N findings at or above high already present", at most 5 titles) and remembers the existing findings without alerting each one. After that, each finding that was not known alerts once. A finding that disappears and reappears within `cooldown_minutes` does not alert again. A source that fails during a scan (for example Docker) does not make its findings look closed.
- A message that cannot be delivered is retried with an exponential backoff (starting at the scan interval, doubling up to 15 minutes) for up to 24 hours, and then dropped; a dropped alert makes its finding eligible to alert again. A failed send never stops the loop or blocks newer alerts. Logs go to stderr as one timestamped line each, without tokens, URLs or remote addresses.
- A source that was unreadable during the first scan (for example Docker not started yet) is summarised once ("N existing findings from docker now visible") when it first becomes readable, instead of alerting each finding.
- State is kept in memory only. Without `--baseline`, `--once` has no state between runs, so every cron run re-sends the first-run summary of whatever is open at that moment; use a baseline to avoid that (next section).
- It needs root only for what the scan needs (process names, Docker socket).

### Drift-only alerts with a baseline

`exposedwatch watch --baseline FILE` (also with `--once`) alerts only on drift from a baseline created with `exposedwatch baseline save`:

- A finding is accepted when its port identity (protocol, port, canonical bind address and source `host` or `docker`) is in the baseline. Accepted findings never alert, neither individually nor in a summary. Process names are ignored, and the same port on a different bind address or from a different source is not accepted.
- There is no first-run summary. Every finding at or above `--min-severity` (or `alerts.min_severity`) that is not accepted alerts individually, with one extra body line: `Not in baseline (saved YYYY-MM-DD)`. Cooldown, deduplication, retries and the handling of unreadable sources work as described above. Closed ports never alert.
- The baseline must be trusted, with the same rules as `scan --baseline` (not a symlink, owned by the current user or root, not writable by group or others, regular file, safe parent directory, bounded strict parse). If it is missing, untrusted or corrupt, `watch` refuses to start with exit code 1 and a hint to run `exposedwatch baseline save`; it never falls back to running without a baseline. `--baseline ""` is a usage error (exit code 2).
- In loop mode the file is reloaded at the start of every cycle, so `exposedwatch baseline save` takes effect without restarting `watch`: findings that become accepted are forgotten silently and newly unaccepted ones alert. If a reload fails after a good load, `watch` keeps the last good baseline and logs one line (`baseline reload failed: ...; using the last good baseline`), another if the reason changes, and one line when it recovers; it does not repeat the message every cycle.
- `--once` loads the baseline once and keeps no state, so every cron run alerts for the drift present at that moment: the same drift alerts again on every run until you accept it by refreshing the baseline with `baseline save`.
- Accepting a port accepts ALL findings on it, whatever the rule or severity: a more severe finding on an already accepted port does NOT alert until the port is removed from the baseline. An empty baseline accepts nothing, so every finding at or above the minimum severity alerts.
- A reload that changes the number of entries or the saved date logs one line (`baseline reloaded: N entries (was M), saved DATE`); it adds `WIDENED` when the baseline grew by at least 50 entries or more than doubled. Unchanged reloads log nothing. A failure is logged once per distinct reason, and load warnings only when they change.
- Logs never include file contents. The path appears in the start-up line, and error messages may include the path you passed.

Example systemd unit. **It is an untested example, shown as documentation only** (this project ships no packaging for it). Create the baseline first (`exposedwatch baseline save`); without a trusted baseline file `watch --baseline` refuses to start. It runs as root. Put the secret in a file readable only by root (`chmod 0600`) and set `"enabled": true` in the config file:

```ini
# /etc/systemd/system/exposedwatch.service
[Unit]
Description=exposedwatch Telegram alerts
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/bin/exposedwatch watch --baseline /var/lib/exposedwatch/baseline.json
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

## Status and what is verified

Verified:

- Installs and runs from the published `.deb` on Debian 12 and Ubuntu 22.04 and 24.04 (`amd64`) in clean containers, and removes cleanly.
- Developed and run on Debian 13.
- The firewall module as root against real `iptables` rules (v1.8.10, nf_tables backend) on Ubuntu 24.04, inside an isolated container network: a port accepted by a rule is reported `allowed`, one accepted only from a source range `restricted`, and one dropped by the default policy `blocked`.
- Baseline mode as root: the state directory and file are created `0700` and `0600` under `/var/lib/exposedwatch`, drift gives exit code 4, and a state directory writable by others is refused.
- The code is covered by an automated test suite, and the security-sensitive parts were reviewed with automated adversarial review. There has been no independent human audit.

Not verified yet:

- The firewall module with `ufw`, the `iptables-legacy` backend, Docker-published ports and native nftables rules (only fixtures were exercised for those), and on a physical host rather than a container.
- Real Telegram delivery.
- The TUI rendering across terminals and colors.
- Cron and systemd usage.
- The `arm64` package.
- Distributions other than Debian and Ubuntu.
- Hand-made native nftables `table ip filter` rules are not detected.

## Security and responsible use

- It can run as root (for process names and firewall rules), with the privileges of the user who starts it.
- It only reads the firewall: it never changes firewall rules.
- The optional nmap scan (`--external`) must only target systems you own or are allowed to test.
- Secrets stay out of logs: the Telegram bot token is shown masked and removed from errors and logs.
- To report a vulnerability, follow [SECURITY.md](SECURITY.md). Do not open a public issue.

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
- Read-only: it reads `/proc` and sends a single `GET` request to the Docker socket to list containers. It does not open or change ports, and does not modify containers. With `--external`, nmap additionally opens TCP connections to the scanned targets. Only `watch`, `notify-test` and the test message of Settings > Telegram make an outbound connection of their own: they send alerts to Telegram (see [Alerts (Telegram)](#alerts-telegram)).
- `--nmap-path` runs the given binary with the same privileges as exposedwatch, so do not let untrusted users pass arguments to `sudo exposedwatch` through sudoers rules.

## License

Apache-2.0. See [LICENSE](LICENSE).
