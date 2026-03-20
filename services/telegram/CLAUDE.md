# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Telegram bot service that manages VPN subscriptions and peer configurations. The service integrates with:
- **Telegram Bot API** for user interactions and payment processing
- **PostgreSQL** for storing client, payment, and peer connection data
- **HTTP API** (AWG service) for managing WireGuard peer configurations

## Architecture & Key Components

### Core Packages

**`cmd/main.go`** — Application entrypoint. Initializes and wires together all services:
- Telegram bot client
- HTTP client for AWG API
- PostgreSQL connection
- Main event loop

**`internal/service/`** — Business logic layer
- `service.go` — Main event loop (`Update()`) that handles Telegram updates, payments, and subscription checks
- `interface.go` — Define interfaces for `postgres`, `telegramClient`, and `httpClient` (enables testability)
- `add.go` — Add peer to AWG and save connection details
- `invoice.go` — Create and send payment invoice
- `checkSubscription.go` — Periodic check for expired subscriptions (runs every hour)
- `texts.go` — Hardcoded Russian text strings for Telegram messages

**`internal/telegram/`** — Telegram Bot API wrapper
- `telegram.go` — Bot initialization and menu UI (inline keyboards)
- `payment.go` — Payment-related handlers (pre-checkout query, successful payment)

**`internal/httpClient/`** — AWG API client
- `httpClient.go` — HTTP requests for peer management: `AddPeer()`, `DeletePeer()`, `DownloadConfFile()`

**`internal/repository/`** — Database access layer
- `postgres.go` — PostgreSQL connection setup
- `client.go` — Client queries (add, check status, test access)
- `payment.go` — Payment tracking queries
- `connection.go` — Peer connection lifecycle queries

**`internal/dto/`** — Data Transfer Objects and encoding
- `dto.go` — Request/response structs for HTTP and Telegram
- `model.go` — Database model definitions
- `errors.go` — Custom error types

**`migrations/`** — Database schema (using `golang-migrate`)
- Three migration sets: clients, payments, peers

**`logger/`** — Logging setup using `go.uber.org/zap`

### Data Flow Example: Adding a Peer

1. User clicks "получить конфиг" button
2. `service.Update()` → `service.getConfFile()`
3. Calls `httpClient.AddPeer()` → AWG service assigns hostID and returns publicKey
4. Saves to DB via `postgres.NewConnection()` and `postgres.SaveKey()`
5. Downloads config file via `httpClient.DownloadConfFile()`
6. Sends config to user via `telegram.SendFile()`

## Development Commands

### Build
```bash
go build -o telegram-service cmd/main.go
```

### Run Locally
```bash
# Terminal 1: Start dependencies
docker-compose up

# Terminal 2: Run the service
go run cmd/main.go
```

The `.env` file must contain:
- `TELEGRAM_KEY` — Telegram bot token
- `PROVIDER_TOKEN` — Stripe payment provider token
- `HTTP_URL` — AWG API endpoint (e.g., `http://localhost:7777`)
- `DB_CONN` — PostgreSQL connection string

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code (static analysis)
go vet ./...

# Check for unused imports/variables
# (No automatic tools configured. Review manually or add golangci-lint)
```

### Database Schema
Migrations run automatically via Docker Compose. To run migrations manually:
```bash
migrate -path ./migrations -database "$DB_CONN" up
```

To rollback:
```bash
migrate -path ./migrations -database "$DB_CONN" down N
```

## Testing Notes

- No existing unit tests in the codebase
- Service is tightly coupled to Telegram and database (integration test style needed)
- The `interface.go` defines mockable interfaces, but tests are not implemented yet

## Important Implementation Details

### Event Loop Logic (`service.Update()`)
The main loop processes three types of Telegram updates:
1. **PreCheckoutQuery** — Confirm payment request
2. **Message** — Handle commands (`/start`, `/menu`) and payment notifications
3. **CallbackQuery** — Handle inline button clicks with action routing

### Callback Data Encoding
Callback button data is encoded/decoded using `dto.DecodeCallbackData()` and `dto.EncodeCallbackData()`. Action strings are in Russian (e.g., "получить конфиг", "оплатить").

### Status Management
- `CheckStatus()` — Returns true if user has paid subscription
- `IsTested()` — Returns true if user used test access (24-hour trial)
- `StatusTrue()` / `StatusFalse()` — Update subscription status after payment

### Subscription Expiration
`CheckSubcription()` runs every hour, queries `ExpiredConnection()` from DB, and deletes expired peers via `httpClient.DeletePeer()`.

## Environment & Dependencies

- **Go 1.26.0** (per Dockerfile and go.mod)
- **PostgreSQL 18+** (via docker-compose)
- **golang-migrate** for schema migrations
- **go-telegram-bot-api/v5** for Telegram API
- **pgx/v5** for direct PostgreSQL driver (no ORM)
- **go.uber.org/zap** for structured logging

## Related Services

This service is part of a larger VPN infrastructure:
- **AWG service** — Manages WireGuard peer configurations; HTTP API at `HTTP_URL`
- **Parent project** — Located at `/home/user/GitProjects/vpn/`
- **Shared DTO types** — Check `../awg/internal/transport/dto/` if cross-service compatibility needed

## Claude Code Scope

When working in this repository:

- Focus ONLY on these directories:
  - ./internal
  - ./cmd
- Do NOT read or analyze files outside these directories unless I explicitly mention a file path.
- Ignore build artifacts, logs, and unrelated files in other directories.
