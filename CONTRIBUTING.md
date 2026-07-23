# Contributing to proxmox-cli

Thanks for your interest in contributing. This document covers how to set up
a development environment, the checks a change must pass, and how releases
work.

## Development setup

Requirements:

- Go 1.26 or newer
- A Proxmox VE server for manual testing (optional; the test suite runs
  entirely against mocks)

```bash
git clone https://github.com/Adz-ai/proxmox-cli.git
cd proxmox-cli
make build          # builds to build/proxmox-cli
```

Point the CLI at a scratch config while developing so your real
configuration is untouched:

```bash
export PROXMOX_CLI_CONFIG=/tmp/proxmox-cli-dev.json
./build/proxmox-cli init
```

## Project layout

- `cmd/` - cobra commands, one package per command group (`vm`, `lxc`,
  `nodes`, `backup`, `auth`, `images`), plus `cmd/utility` for shared
  helpers (client construction, config, task waiting, confirmation
  prompts).
- `internal/interfaces/` - interfaces over the go-proxmox client. All
  commands talk to Proxmox through these so they can be tested with mocks.
- `internal/tui/` - the k9s-style terminal UI (bubbletea).
- `test/mocks/` - generated gomock implementations. Regenerate with
  `make generate` after changing the interfaces; CI fails on drift.
- `test/bdd/` - godog feature tests.

## Before opening a pull request

CI runs the following; run them locally first:

```bash
gofmt -l .                                     # must print nothing
go mod tidy -diff
go generate ./internal/interfaces              # then: git diff test/mocks/
go vet ./...
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run ./...
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
go test -race ./...
```

Guidelines:

- Follow the existing patterns: commands are built by `newXxxCmd()`
  constructors, use `RunE`, and reach Proxmox only through the interfaces
  in `internal/interfaces`.
- New functionality needs unit tests using the gomock mocks. Destructive
  commands must prompt for confirmation and accept `--yes`.
- Use conventional commit messages (`feat:`, `fix:`, `docs:`, `ci:`,
  `chore:`); the release changelog is generated from them.
- Keep output plain and professional: no emojis, tables aligned, errors
  wrapped with context.

## Releases

Releases are cut by pushing a `v*` tag. The release workflow re-runs the
full verification gate, then GoReleaser builds archives for
linux/darwin/windows, deb/rpm/apk packages, a Homebrew cask, and publishes
a GitHub release with a generated changelog.

## Reporting issues

Open a GitHub issue with the CLI version (`proxmox-cli --version`), your
Proxmox VE version, the command you ran, and the full error output.
