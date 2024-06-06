# ghstat

<a href="https://snapcraft.io/ghstat"><img src="https://snapcraft.io/ghstat/badge.svg" alt="Snap Status"></a>
<a href="https://github.com/jnsgruk/ghstat/actions/workflows/release.yaml"><img src="https://github.com/jnsgruk/ghstat/actions/workflows/release.yaml/badge.svg"></a>

`ghstat` is a small utility I wrote for gathering metrics about hiring pipelines at Canonical. If
you aren't a Hiring Lead at Canonical, there is very little chance this is interesting to you 😉.

The tool uses [`go-rod`](https://pkg.go.dev/github.com/go-rod/rod) under the hood to drive a
headless browser, which can interact with both the Ubuntu One SSO, and subsequently Greenhouse.

## Installation

The easiest way to consume `ghstat` is using the [Snap](https://snapcraft.io/ghstat):

```shell
sudo snap install ghstat
```

Or you can clone, build and run like so:

```shell
git clone https://github.com/jnsgruk/ghstat
cd ghstat
go build -o ghstat main.go
./ghstat
```

## Usage

The output of `ghstat --help` can be seen below. In general, once the
[configuration](#configuration) file is in place, you'll likely just invoke `ghstat`.

On the first launch, you'll need to enter your Ubuntu One login, password and one-time password.
After that, the cookies created will be stored for future use, so you'll only need to
reauthenticate once the session expires.

```
A utility for gathering role-specific statistics from Greenhouse.

ghstat provides automation for gather statistics about a given Hiring Lead and the roles
they manage as part of Canonical's hiring process.

This tool is configured using a single file in one of the following locations:

  - ./ghstat.yaml
  - $HOME/.config/ghstat/ghstat.yaml

The configuration file should specify a top-level 'leads' list:

    leads:
    - name: Joe Bloggs
        roles:
        - 1234567
        - 8910111

By default, ghstat will try to reuse an active Greenhouse session by reading the cookies
from a previous invocation. In the case that this isn't possible, it will prompt
for Ubuntu One credentials. To streamline login, the following environment variables can be set:

  - U1_LOGIN - the username/email for Ubuntu One login
  - U1_PASSWORD - the password for Ubuntu One login

For more information, visit the homepage at: https://github.com/jnsgruk/ghstat

Usage:
  ghstat [flags]

Flags:
  -c, --config string   path to a specific config file to use
  -h, --help            help for ghstat
  -l, --leads strings   filter results to specific hiring leads from the config
  -o, --output string   choose the output format ('pretty', 'markdown' or 'json') (default "pretty")
  -v, --verbose         enable verbose logging
      --version         version for ghstat
```

## Configuration

The tool takes some simple configuration as a YAML file, which it expects to find either in the
current working directory, or in `~/.config/ghstat/ghstat.yaml`:

```yaml
# (Required): A list of Hiring Leads
leads:
  # (Required) The name or alias of the Hiring Lead in question
  - name: <hiring lead name>
    # (Required) The list of role IDs they Hiring Lead manages. These IDs can
    # be gathered by navgating to the role Dashboard, and looking at the ID in
    # the URL, for example: https://canonical.greenhouse.io/sdash/<ID here>
    roles:
      - <number>
```

An example config file can be seen below:

```yaml
leads:
  - name: Joe Bloggs
    roles:
      - 1234567
      - 8910111

  - name: A.N. Other
    roles:
      - 1213141
      - 5161718
      - 1920212
      - 2232425
```

## Development / HACKING

This project uses [goreleaser](https://goreleaser.com/) to build and release.

You can get started by just using Go, or with `goreleaser`:

```shell
# Clone the repository
git clone https://github.com/jnsgruk/releasegen
cd releasegen

# Build/run with Go
go run main.go

# Build a snapshot release with goreleaser (output in ./dist)
goreleaser build --rm-dist --snapshot
```
