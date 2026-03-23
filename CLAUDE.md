# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**`vpn`** is a monorepo for **goFastVPN** / **GopherSecure** — a VPN infrastructure built around AmneziaWireGuard (AWG). It consists of three independent Go services that work together:

1. **`awg`** — Low-level peer management via HTTP API with transactional delete/restore operations
2. **`telegram`** — User-facing Telegram bot (@GopherSecureBot) for subscriptions, peer provisioning via YooKassa payments, and admin API for promo code management
3. **`cli-admins`** — Command-line tool for administrators to manage promotional codes and system parameters

Services use PostgreSQL for persistence and communicate via HTTP. The project is currently in active development with recent work on promo codes, transactional peer operations, and payment integration.

## Repository Structure

```
/vpn/
├── Makefile                 # Root-level commands for running services locally
├── CLAUDE.md                # This file — project architecture guidance
├── README.md                # User-facing project description
├── services/
│   ├── awg/                 # AmneziaWireGuard service (HTTP peer management)
│   │   ├── CLAUDE.md        # Detailed AWG service docs
│   │   ├── http_api.md      # HTTP API specification
│   │   ├── cmd/main.go      # Entrypoint
│   │   ├── internal/
│   │   │   ├── repository/  # Peer persistence (add, delete, restore, load)
│   │   │   ├── transport/   # HTTP handlers and DTOs
│   │   │   ├── getEnv/      # Environment variable parsing
│   │   │   └── logger/      # Structured syslog logging
│   │   └── {go.mod,go.sum}  # Dependencies
│   ├── telegram/            # Telegram Bot + payment service + admin API
│   │   ├── CLAUDE.md        # Detailed Telegram service docs
│   │   ├── cmd/main.go      # Entrypoint
│   │   ├── internal/
│   │   │   ├── service/     # Business logic (add peer, payments, subscriptions, promos)
│   │   │   ├── telegram/    # Telegram Bot API wrapper
│   │   │   ├── httpClient/  # AWG API client
│   │   │   ├── repository/  # PostgreSQL access layer
│   │   │   ├── features/    # Feature modules (promocode CRUD, etc.)
│   │   │   ├── dto/         # Data transfer objects
│   │   │   └── logger/      # Structured JSON logging
│   │   ├── migrations/      # Database schema (golang-migrate)
│   │   ├── docker-compose.yml # PostgreSQL for local development
│   │   └── {go.mod,go.sum}  # Dependencies
│   └── cli-admins/          # Admin CLI for managing promo codes
│       ├── CLAUDE.md        # Detailed CLI docs
│       ├── cmd/main.go      # CLI commands (create, update, list, delete)
│       ├── internal/
│       │   └── client/      # HTTP client for telegram admin API
│       └── {go.mod,go.sum}  # Dependencies (stdlib only)
└── .gitignore               # Git exclusions
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
- `HTTP_ENDPOINT` — Listen address (e.g., `0.0.0.0:7777`)
- `AWG_ENDPOINT` — AWG daemon socket
- `DEVICE` — WireGuard interface name (e.g., `awg0`)
- `JC`, `JMIN`, `JMAX`, `S1`, `S2`, `H1`-`H4` — Obfuscation parameters

**Key Features:**
- Transactional peer delete/restore (safe rollback on errors)
- Automatic peer loading on startup via `LoadUsers()`
- Structured syslog logging
- Config file generation with preshared key support

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
- `ADMIN_ADDRESS` — Admin API listen address (default: `0.0.0.0:8080`)

**Key Features:**
- Telegram user interface for peer management
- YooKassa payment integration
- Promo code system with usage limits and expiration
- Admin HTTP API for managing promo codes
- Subscription lifecycle with automatic expiration checks
- Structured JSON logging via zap

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
cd services/cli-admins && go fmt ./... && go vet ./...
```

## CLI-Admins Tool

Run from `services/cli-admins/`:

```bash
# Build
go build -o cli-admins cmd/main.go

# List all promo codes
./cli-admins list

# Create a promo code
./cli-admins create BONUS30 30 100 2026-03-29T23:59:59Z

# Update promo code
./cli-admins update 1 60 200 2026-04-30T23:59:59Z

# Delete (deactivate) promo code
./cli-admins delete 1
```

**Environment:**
- `ADDRESS` — Telegram admin API endpoint (default: `0.0.0.0:8080`)

See `services/cli-admins/CLAUDE.md` for detailed documentation.

## Service Communication

**Data Flow: Add Peer (User Gets VPN Config)**

1. User clicks "получить конфиг" in Telegram bot
2. **Telegram service** checks subscription status in PostgreSQL
3. **Telegram service** calls AWG HTTP API: `POST /peers`
4. **AWG service** generates WireGuard peer via `awgctrl-go` (includes preshared key)
5. **Telegram service** saves connection metadata to PostgreSQL (host_id, public_key, preshared_key)
6. **Telegram service** retrieves config via `GET /peers/{id}/config` and sends to user

**Data Flow: Apply Promo Code (Extend Subscription)**

1. User sends `/promo BONUS30` to Telegram bot
2. **Telegram service** validates promo code in PostgreSQL
3. If promo is valid and not expired:
   - If user has existing peer: call `RestorePeer()` on AWG to re-enable it
   - Add bonus days to subscription expiration date
   - Record activation in `promo_activations` table
4. Confirm promo applied to user

**API Contracts**

- **AWG API** — `/peers` for CRUD on VPN peers (see `services/awg/http_api.md`)
  - Transactional delete with safe rollback via `RestoreUser()`
- **Telegram Service** — User-facing Telegram bot interface
- **Admin API** — `/admin/promo/*` endpoints for promo code CRUD (Echo framework)
  - CLI-Admins calls these endpoints
- All services are stateless; state lives in PostgreSQL and WireGuard configs

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

**AWG Service:**
- `github.com/slipynil/awgctrl-go` v1.2.0 — WireGuard control library
- `github.com/gorilla/mux` v1.8.1 — HTTP routing

**Telegram Service:**
- `github.com/go-telegram-bot-api/telegram-bot-api/v5` v5.5.1 — Telegram Bot API
- `github.com/jackc/pgx/v5` v5.8.0 — PostgreSQL driver (direct access, no ORM)
- `github.com/labstack/echo/v5` v5.0.4 — HTTP framework (admin API)
- `go.uber.org/zap` v1.27.1 — Structured JSON logging
- `github.com/golang-migrate/migrate/v4` v4.x — Database schema versioning

**CLI-Admins:**
- Standard library only (no external dependencies)

## Recent Improvements

### Completed in Latest Releases

**Transactional Peer Management (AWG Service)**
- Peer delete now uses safe two-phase operation (JSON first, then WireGuard)
- `DeleteUserEx()` returns peer data for rollback if subsequent operations fail
- `RestoreUser()` recovers peers when needed (e.g., subscription renewal via promo code)
- See `services/awg/internal/repository/delete_user.go` and `restore_peer.go`

**Promo Code System (Telegram Service)**
- Full CRUD operations for promotional codes
- Usage limits and expiration date support
- Automatic peer restoration on promo code application
- Admin HTTP API for code management
- See `services/telegram/internal/features/promocode/` and `internal/service/promo.go`

**Admin CLI Tool (CLI-Admins Service)**
- Standalone command-line tool for administrators
- Manage promo codes: create, update, list, delete
- Communicates with Telegram admin API
- Can be deployed with ldflags for custom configuration

## Future Development

Potential improvements (not yet scheduled):
1. Graceful shutdown handlers for all services
2. Automated config deletion/archival policy
3. Enhanced error messages with context wrapping
4. Unit tests for promo code flow
5. Rate limiting on admin API endpoints

## Per-Service Deep Dives

For detailed architecture, design patterns, and implementation specifics:
- **AWG**: Read `services/awg/CLAUDE.md` and `services/awg/http_api.md`
- **Telegram**: Read `services/telegram/CLAUDE.md`

Both files contain handler signatures, database schema notes, and current coupling points.
