# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**`vpn`** is a monorepo for **goFastVPN** / **GopherSecure** — a VPN infrastructure built around AmneziaWireGuard (AWG). It consists of two independent Go services that work together:

1. **`awg`** — Low-level peer management via HTTP API
2. **`telegram`** — User-facing Telegram bot (@GopherSecureBot) for subscriptions and peer provisioning via YooKassa payments

Both services use PostgreSQL for persistence and communicate via HTTP. The project is currently in active development with recent work on API improvements, payment integration, and subscription logic.

## Repository Structure

```
/vpn/
├── Makefile                 # Root-level commands for running services locally
├── services/
│   ├── awg/                 # AmneziaWireGuard service (HTTP peer management)
│   │   ├── CLAUDE.md        # Detailed AWG service docs
│   │   ├── http_api.md      # HTTP API specification
│   │   ├── cmd/main.go      # Entrypoint
│   │   ├── internal/        # Service code (handlers, repository, logging)
│   │   └── {go.mod,go.sum}  # Dependencies
│   └── telegram/            # Telegram Bot + payment service
│       ├── CLAUDE.md        # Detailed Telegram service docs
│       ├── cmd/main.go      # Entrypoint
│       ├── internal/        # Service code (telegram, repository, http client)
│       ├── migrations/      # Database schema (golang-migrate)
│       ├── docker-compose.yml # PostgreSQL for local development
│       └── {go.mod,go.sum}  # Dependencies
└── todo.md                  # Open development tasks
```

## Quick Start: Running Services Locally

### Prerequisites

- **Go 1.26.0** (or later)
- **Docker & Docker Compose** (for PostgreSQL)
- `.env` files must exist in each service directory (copy `.env.example`)

### AWG Service

```bash
# From root, requires sudo (kernel access for WireGuard):
make awg-run

# Or from services/awg/:
export $(cat .env | xargs)
cd services/awg
sudo go run cmd/main.go
```

**Environment variables** (see `services/awg/.env.example`):
- `HTTP_ENDPOINT` — Listen address
- `AWG_ENDPOINT` — AWG daemon socket
- `DEVICE` — WireGuard interface name
- `JC`, `JMIN`, `JMAX`, `S1`, `S2`, `H1`-`H4` — Obfuscation parameters

### Telegram Service

```bash
# Terminal 1: Start PostgreSQL
cd services/telegram
make compose-up
# or: docker-compose up -d

# Terminal 2: Run the service
go run cmd/main.go

# See logs:
make compose-logs
# or: docker-compose logs -f
```

**Environment variables** (see `services/telegram/.env.example`):
- `TELEGRAM_KEY` — Telegram bot token
- `PROVIDER_TOKEN` — YooKassa payment provider token
- `HTTP_URL` — AWG service endpoint (e.g., `http://localhost:7777`)
- `DB_CONN` — PostgreSQL connection string

## Root-Level Commands

```bash
# Build & Run
make awg-run               # Run AWG service (requires sudo)
make pay-run              # Run Telegram service
make compose-up           # Start PostgreSQL for telegram service
make compose-down         # Stop PostgreSQL
make compose-logs         # View PostgreSQL logs

# Common Development Tasks
cd services/awg && go fmt ./... && go vet ./...
cd services/telegram && go fmt ./... && go vet ./...
```

## Service Communication

**Data Flow: Add Peer**

1. User clicks "получить конфиг" in Telegram bot
2. **Telegram service** calls AWG HTTP API: `POST /peers`
3. **AWG service** generates WireGuard peer via `awgctrl-go`
4. **Telegram service** saves connection metadata to PostgreSQL
5. **Telegram service** retrieves config via `GET /peers/{id}/config` and sends to user

**API Contract**

- AWG exposes `/peers` for CRUD operations on VPN peers (see `services/awg/http_api.md`)
- Telegram calls AWG and PostgreSQL to manage user subscriptions and peer lifecycle
- Both services are stateless; state lives in PostgreSQL and WireGuard configs

## Development Notes

### Testing

- **AWG** — No tests yet
- **Telegram** — No tests yet
- Service coupling is tight (concrete Telegram + HTTP client); see `interface.go` for mockable interfaces

### Database

Telegram service uses **golang-migrate** for schema versioning:
```bash
cd services/telegram
migrate -path ./migrations -database "$DB_CONN" up      # Apply migrations
migrate -path ./migrations -database "$DB_CONN" down 1  # Rollback 1 migration
```

Migrations are auto-run via Docker Compose startup.

### Logging

- **AWG** — Structured syslog via `internal/logger/`
- **Telegram** — Structured JSON logging via `go.uber.org/zap` in `internal/logger/`

### Code Style

- Format: `go fmt ./...`
- Vet: `go vet ./...`
- No linter configured (consider adding `golangci-lint` if needed)

## Key Dependencies

- `github.com/slipynil/awgctrl-go` — WireGuard control (AWG only)
- `github.com/go-telegram-bot-api/v5` — Telegram Bot API (Telegram only)
- `pgx/v5` — PostgreSQL driver (direct, no ORM)
- `golang-migrate/migrate/v4` — Schema versioning
- `go.uber.org/zap` — Structured logging
- `github.com/gorilla/mux` — HTTP routing (AWG only)

## Open Tasks

See `todo.md` for active development items:
1. Graceful shutdown for AWG service
2. Ability to delete old configs
3. Error handling ("wrap error") for Telegram commands

## Per-Service Deep Dives

For detailed architecture, design patterns, and implementation specifics:
- **AWG**: Read `services/awg/CLAUDE.md` and `services/awg/http_api.md`
- **Telegram**: Read `services/telegram/CLAUDE.md`

Both files contain handler signatures, database schema notes, and current coupling points.
