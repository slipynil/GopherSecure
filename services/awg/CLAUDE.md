# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`awg-service` is a lightweight HTTP API service for managing AmneziaWireGuard VPN peers. It is part of the larger `goFastVPN` monorepo at `github.com/slipynil/goFastVPN`. The service wraps the `awgctrl-go` library to expose peer lifecycle operations (add/delete/config) over HTTP.

## Commands

### Run

From this service directory:
```bash
go run ./cmd/main.go
```

From the monorepo root (loads `.env` automatically):
```bash
make awg-run
```
> `awg-run` requires `sudo` because AWGCTRL needs kernel-level access to manage WireGuard interfaces.

### Build

```bash
go build -o awg-service ./cmd/main.go
```

### Lint / Format

```bash
go fmt ./...
go vet ./...
```

### Tests

No tests exist yet. When added:
```bash
go test ./...
go test -v -run TestName ./internal/...
```

## Environment Variables

Copy `.env.example` to `.env`. All variables are required and the service panics on startup if any are missing.

| Variable        | Description                              |
|-----------------|------------------------------------------|
| `HTTP_ENDPOINT` | Address the service listens on           |
| `AWG_ENDPOINT`  | Address of the AWG daemon                |
| `DEVICE`        | WireGuard device name (e.g. `awg0`)      |
| `JC`            | Obfuscation: jitter count                |
| `JMIN`/`JMAX`   | Obfuscation: jitter min/max              |
| `S1`/`S2`       | Obfuscation: shift parameters            |
| `H1`–`H4`       | Obfuscation: hash parameters             |

## Architecture

```
cmd/main.go                    # Entry point: loads env, inits awgctrl client, starts HTTP server
internal/
  getEnv/obfuscation.go        # Parses obfuscation env vars into awgctrlgo.Obfuscation struct
  transport/
    server.go                  # Gorilla Mux router, registers routes, starts listener
    handlers.go                # HTTP handlers: AddPeer, DeletePeer, SendConfFile
    dto.go                     # Request/response structs, standard JSON response wrapper
```

### HTTP API

| Method | Path                  | Handler        | Description                              |
|--------|-----------------------|----------------|------------------------------------------|
| POST   | `/peers`              | `AddPeer`      | Creates peer, returns public key         |
| DELETE | `/peers`              | `DeletePeer`   | Removes peer by public key               |
| GET    | `/peers/{id}/config`  | `SendConfFile` | Returns peer config file content         |

### Peer Config Storage

Config files are written to `/etc/amnezia/amneziawg/configs/{id}.conf` on the host.

## Key Dependencies

- `github.com/slipynil/awgctrl-go` — AmneziaWireGuard control library (kernel interface)
- `github.com/gorilla/mux` — HTTP routing

## Go Version

The module targets **Go 1.25** (as declared in `go.mod`). The installed toolchain may be newer.
