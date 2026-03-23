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
- `promo.go` — Apply promo codes and extend subscriptions (calls `RestorePeer()` on renewal)
- `invoice.go` — Create and send payment invoice
- `checkSubscription.go` — Periodic check for expired subscriptions (runs every hour)
- `texts.go` — Hardcoded Russian text strings for Telegram messages
- `add_test.go` — Unit tests for add peer logic

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
- `promocode.go` — Promo code CRUD and validation queries
- `*_test.go` — Unit tests for repository operations

**`internal/dto/`** — Data Transfer Objects and encoding
- `dto.go` — Request/response structs for HTTP and Telegram
- `model.go` — Database model definitions
- `callback.go` — Encode/decode callback button data
- `errors.go` — Custom error types

**`internal/features/promocode/`** — Promo code feature module (NEW)
- `handler.go` — Echo HTTP handlers for `/admin/promo/*` endpoints
- `dto.go` — CreatePromoRequest, UpdatePromoRequest, etc.
- `errors.go` — Promo code specific errors

**`migrations/`** — Database schema (using `golang-migrate`)
- `000001_client.up.sql` — Clients table (chat_id, username, status, is_tested)
- `000002_payment.up.sql` — Payments table (payment_id, chat_id, payload, status)
- `000003_peer.up.sql` — Peer connections (host_id, chat_id, public_key, expires_at)
- `000004_add_preshared_key_to_peer.up.sql` — Add preshared_key column
- `000005_promo_codes.up.sql` — Promo codes & activations tables (NEW)

**`logger/`** — Logging setup using `go.uber.org/zap`

### Data Flow Example: User Gets VPN Config

1. User clicks "получить конфиг" button
2. `service.Update()` → `service.getConfFile()`
3. Checks `postgres.CheckStatus()` for active subscription
4. Calls `httpClient.AddPeer()` → AWG service `POST /peers` returns public_key + preshared_key
5. Saves to DB via `postgres.NewConnection()` (stores host_id, public_key, preshared_key)
6. Downloads config file via `httpClient.DownloadConfFile()` → `GET /peers/{id}/config`
7. Sends config to user via `telegram.SendFile()`

### Data Flow Example: User Applies Promo Code

1. User sends `/promo BONUS30`
2. `service.Update()` → `service.ApplyPromoCodeFromMessage()`
3. `postgres.GetPromoCode()` validates code exists, not expired, not deactivated
4. `postgres.CanActivatePromoCode()` checks usage limits and user hasn't used it before
5. If user has no peer: creates one via `httpClient.AddPeer()`
6. If user has expired peer: calls `httpClient.RestorePeer()` on AWG to re-enable it
7. `postgres.ApplyPromoBonusDays()` extends subscription expiration
8. `postgres.ActivatePromoCode()` records activation in `promo_activations`
9. Confirms to user: "✅ Промокод применен! +30 дней"

### Database Schema

**clients** — User accounts
- `chat_id` (BIGINT, PK) — Telegram user ID
- `username` (TEXT) — Telegram username
- `status` (BOOLEAN) — Subscription active?
- `is_tested` (BOOLEAN) — Used 24-hour trial?

**payments** — Payment records
- `payment_id` (SERIAL, PK)
- `chat_id` (BIGINT, FK)
- `payload` (TEXT)
- `status` (TEXT) — 'pending', 'succeeded'

**connections** (peers) — WireGuard peer lifecycle
- `host_id` (SERIAL, PK)
- `chat_id` (BIGINT, FK)
- `public_key` (TEXT) — WireGuard public key
- `preshared_key` (TEXT) — WireGuard preshared key
- `expires_at` (TIMESTAMP) — Subscription expiration

**promo_codes** — Promotional codes (NEW)
- `id` (SERIAL, PK)
- `code` (VARCHAR, UNIQUE) — e.g., "BONUS30"
- `bonus_days` (INT) — Days to add on activation
- `max_uses` (INT) — 0 = unlimited
- `used_count` (INT) — Current usage count
- `is_active` (BOOLEAN) — Can be deactivated
- `expires_at` (TIMESTAMP) — Code expiration

**promo_activations** — Usage tracking (NEW)
- `id` (SERIAL, PK)
- `promo_id` (INT, FK)
- `chat_id` (BIGINT, FK)
- `activated_at` (TIMESTAMP)
- UNIQUE(promo_id, chat_id) — Each user can use code once

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

### Admin API

HTTP endpoints for managing promo codes (Echo framework):

**POST /admin/promo** — Create promo code
```json
{
  "code": "BONUS30",
  "bonus_days": 30,
  "max_uses": 100,
  "expires_at": "2026-03-29T23:59:59Z"
}
```

**GET /admin/promo** — List all promo codes
Returns array of promo code objects with usage stats

**PUT /admin/promo/{id}** — Update promo code
```json
{
  "bonus_days": 60,
  "max_uses": 200,
  "expires_at": "2026-04-30T23:59:59Z"
}
```

**DELETE /admin/promo/{id}** — Deactivate promo code
Sets `is_active = false` (doesn't delete, preserves audit trail)

See `internal/features/promocode/handler.go` for implementation.

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code (static analysis)
go vet ./...

# Run tests
go test ./...

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
- **golang-migrate/migrate/v4** for schema migrations
- **go-telegram-bot-api/telegram-bot-api/v5** v5.5.1 for Telegram API
- **jackc/pgx/v5** v5.8.0 for direct PostgreSQL driver (no ORM)
- **labstack/echo/v5** v5.0.4 for admin HTTP API
- **go.uber.org/zap** v1.27.1 for structured JSON logging

## Related Services

This service is part of a larger VPN infrastructure:
- **AWG service** — Manages WireGuard peer configurations; HTTP API at `HTTP_URL`
- **Parent project** — Located at `/home/user/GitProjects/vpn/`
- **Shared DTO types** — Check `../awg/internal/transport/dto/` if cross-service compatibility needed

## Integration with Other Services

**AWG Service Integration:**
- Calls `POST /peers` to create new peers
- Calls `DELETE /peers` to remove expired peers
- Calls `POST /peers/{id}/restore` to re-enable peers on promo code activation
- Calls `GET /peers/{id}/config` to download peer configs
- See `internal/httpClient/httpClient.go` for implementation

**CLI-Admins Integration:**
- CLI tool calls admin API endpoints (`/admin/promo/*`)
- Uses `ADDRESS` env var to locate telegram service
- See parent directory `../cli-admins/` for CLI implementation

## Key Concepts

**Subscription Status:**
- `status = true` — User has paid subscription (active)
- `status = false` — User doesn't have active subscription
- `is_tested = true` — User has used 24-hour trial access

**Peer Lifecycle:**
- Created: User clicks "получить конфиг" → peer created if subscription active
- Active: Peer remains in WireGuard until expiration
- Expired: `checkSubscription()` runs hourly, deletes expired peers
- Restored: Promo code application calls `RestorePeer()` to re-enable

**Promo Code States:**
- `is_active = true, expires_at > now()` — Can be applied by users
- `is_active = true, expires_at < now()` — Expired, cannot apply
- `is_active = false` — Deactivated by admin, cannot apply
- `used_count >= max_uses` (if max_uses > 0) — Limit reached, cannot apply
