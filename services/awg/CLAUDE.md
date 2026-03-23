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
cmd/main.go                           # Entry point: loads env, inits awgctrl client, starts HTTP server
internal/
  getEnv/
    obfuscation.go                    # Parses obfuscation env vars into awgctrlgo.Obfuscation struct
  transport/
    server.go                         # Gorilla Mux router, registers routes, starts listener
    dto/
      dto.go                          # Request/response structs, DeleteResult for rollback support
    handlers/
      entity.go                       # Handler struct, interfaces for awg and repository
      add_peer.go                     # AddPeer HTTP handler: creates peer and returns keys
      delete_peer.go                  # DeletePeer HTTP handler: transactional delete with rollback
      restore_peer.go                 # RestorePeer HTTP handler: restore peer for renewals
      send_file.go                    # SendConfFile HTTP handler: returns peer config
      http_response.go                # Response formatting utilities
  repository/
    entity.go                         # Repository struct for file operations
    add_user.go                       # AddUser: persist peer to users.json
    delete_user.go                    # DeleteUser, DeleteUserEx (with rollback info), RestoreUser
    get_user.go                       # GetUser: retrieve peer by host_id
    get_file.go                       # GetFile: read peer config from disk
    load_users.go                     # LoadUsers: load all peers at startup
    restore_peer.go                   # RestorePeer: re-enable a deleted peer
    model/user.go                     # User struct definition
    repository_test.go                # Unit tests
  logger/
    logger.go                         # Structured logging to syslog
```

### Key Design Patterns

- **Dependency Injection**: Handlers and repository are injected as interfaces in `entity.go`
- **Repository Pattern**: File operations abstracted via `repository` package
- **Transactional Operations**: Delete followed by optional restore; errors at any step are loggable
- **Logging**: Centralized syslog-based logging via `logger` package

### HTTP API

| Method | Path                  | Handler        | Description                                        |
|--------|-----------------------|----------------|----------------------------------------------------|
| POST   | `/peers`              | `AddPeer`      | Creates peer, returns public & preshared keys      |
| DELETE | `/peers`              | `DeletePeer`   | Removes peer (transactional, with rollback info)   |
| GET    | `/peers/{id}/config`  | `SendConfFile` | Returns peer config file content (`.conf`)         |
| POST   | `/peers/{id}/restore` | `RestorePeer`  | Re-enable deleted peer (for subscription renewals) |

### Peer Lifecycle & Transactional Delete

**Normal Flow:**
1. `POST /peers` → creates WireGuard peer, saves to `users.json`, returns keys
2. `DELETE /peers?public_key=...` → calls `DeleteUserEx()` to get peer data, then deletes from WireGuard
3. If WireGuard delete fails → `RestoreUser()` is called to recover peer data
4. `POST /peers/{id}/restore` → re-enable peer without new key generation

**DeleteUserEx() Behavior:**
- Returns `DeleteResult` containing deleted peer data
- Allows caller (Telegram service) to decide on rollback
- Ensures rollback-safe operations across service boundaries

### Peer Config Storage

Config files are written to `/etc/amnezia/amneziawg/configs/{id}.conf` on the host.

Configs include:
- Private key (generated)
- Public key (returned to caller)
- Preshared key (returned to caller)
- Server endpoint and allowed IPs (via obfuscation params)

## Key Dependencies

- `github.com/slipynil/awgctrl-go` — AmneziaWireGuard control library (kernel interface)
- `github.com/gorilla/mux` — HTTP routing

## Safe Peer Deletion Flow

When a peer subscription expires or user cancels:

1. **Telegram Service** calls `DELETE /peers?public_key=ABC123`
2. **AWG Service Handler** calls `repository.DeleteUserEx(public_key)`
3. **Repository Layer**:
   - Reads peer from `users.json` and stores in `DeleteResult`
   - Removes from JSON file
   - Returns `DeleteResult` to handler
4. **Handler** calls WireGuard delete via `awgctrl-go`
5. **On Success**: Returns HTTP 200 with deletion timestamp
6. **On Failure**:
   - Handler calls `repository.RestoreUser(result)` to recover peer
   - WireGuard delete is retried or logged as CRITICAL error
   - Returns HTTP 500 with error details

## Go Version

The module targets **Go 1.26.0** (as declared in `go.mod`). The installed toolchain may be newer.

## Testing

Run tests:
```bash
go test ./...
go test -v -run TestName ./internal/...
```

Tests cover:
- `repository_test.go` — File I/O operations and JSON persistence
- Handler tests (if present) — HTTP routing and response formatting
