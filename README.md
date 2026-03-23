# GopherSecure VPN Service

A production VPN infrastructure for purchasing and managing AmneziaWireGuard (AWG) VPN subscriptions through Telegram.

**Telegram Bot:** [@GopherSecureBot](https://t.me/GopherSecureBot)

## Overview

GopherSecure provides a complete VPN subscription service built on **AmneziaWireGuard** (a privacy-enhanced WireGuard fork). Users subscribe, pay via YooKassa integration with Telegram, and instantly receive VPN configurations through the Telegram bot.

### Key Features

- 🤖 **Telegram Bot Integration** — User-friendly subscription management via [@GopherSecureBot](https://t.me/GopherSecureBot)
- 💳 **YooKassa Payments** — Seamless in-Telegram payment processing with invoice handling
- 🔐 **AmneziaWireGuard Peers** — Automatic WireGuard peer creation and configuration management
- 📊 **PostgreSQL Database** — Persistent storage of clients, payments, connections, peer configs, and promotions
- ⏰ **Subscription Lifecycle** — Automated expiration checks and peer cleanup
- 🎁 **Promotional Codes** — Admin-managed promo codes with bonus days, usage limits, and expiration
- 🔐 **Peer Obfuscation** — Support for AWG obfuscation parameters (jitter, shift, hash)
- 📝 **Config File Delivery** — Users download `.conf` files directly from Telegram
- 🛠️ **Admin CLI Tool** — Manage promotional codes via command-line interface (`cli-admins`)

### Technology Stack

- **Go 1.26.0** — Core services
- **PostgreSQL 18+** — Data persistence
- **AmneziaWireGuard** — VPN protocol
- **Telegram Bot API** — User interface
- **YooKassa** — Payment processing
- **Docker & Docker Compose** — Local development and deployment

## Architecture

GopherSecure is a **three-service monorepo**:

### 1. AWG Service (`services/awg/`)
Low-level WireGuard peer management via HTTP API.

**Responsibilities:**
- Interface with AmneziaWireGuard daemon
- Create/delete/restore WireGuard peers (transactional operations)
- Generate and serve peer configuration files
- Apply obfuscation parameters
- Support safe rollback on failures

**Endpoints:**
- `POST /peers` — Create peer (returns public & preshared keys)
- `DELETE /peers` — Remove peer (transactional, with rollback info)
- `GET /peers/{id}/config` — Download config file
- `POST /peers/{id}/restore` — Re-enable deleted peer (for renewals)

See [`services/awg/http_api.md`](services/awg/http_api.md) for full API specification.

### 2. Telegram Service (`services/telegram/`)
User-facing bot, business logic, and admin API.

**Responsibilities:**
- Handle Telegram user interactions (`/start`, `/menu`)
- Process payments via YooKassa pre-checkout and success callbacks
- Manage promotional codes with CRUD operations
- Call AWG service to create/delete/restore peers for users
- Manage subscription lifecycle in PostgreSQL
- Expose admin HTTP API for promo code management
- Schedule hourly expiration checks
- Send VPN configs to users

**Main Components:**
- `cmd/main.go` — Service initialization
- `internal/telegram/` — Telegram Bot API wrapper
- `internal/service/` — Business logic and event loop (includes promo code application)
- `internal/repository/` — PostgreSQL access layer (including promo code queries)
- `internal/features/promocode/` — Promo code admin API handlers
- `internal/httpClient/` — AWG API client
- `internal/dto/` — Data transfer objects
- `migrations/` — Database schema (includes promo tables)

### 3. CLI-Admins Tool (`services/cli-admins/`)
Command-line interface for administrators.

**Responsibilities:**
- Create, update, list, and delete promotional codes
- Communicate with Telegram service admin API
- Provide user-friendly command-line interface with table output

**Commands:**
- `cli-admins create <code> <days> <max_uses> <expires_at>` — Create promo code
- `cli-admins update <id> <days> <max_uses> <expires_at>` — Update promo code
- `cli-admins list` — Show all codes with usage stats
- `cli-admins delete <id>` — Deactivate promo code

See `services/cli-admins/CLAUDE.md` for detailed documentation.

## Usage

### For End Users

1. Find [@GopherSecureBot](https://t.me/GopherSecureBot) on Telegram
2. Start conversation with `/start`
3. Click "Меню" (Menu) to view options:
   - **получить конфиг** — Get VPN config (add peer)
   - **оплатить** — Purchase subscription
   - **проверить статус** — Check subscription status
4. Follow prompts to complete payment via Stripe
5. Receive `.conf` file to import into VPN app

### For Developers

**Run all services locally:**
```bash
# Terminal 1: Start PostgreSQL
cd services/telegram && make compose-up

# Terminal 2: Run AWG service (requires sudo for WireGuard access)
make awg-run

# Terminal 3: Run Telegram service
make pay-run
```

**Code formatting & linting:**
```bash
cd services/awg && go fmt ./... && go vet ./...
cd services/telegram && go fmt ./... && go vet ./...
cd services/cli-admins && go fmt ./... && go vet ./...
```

**Database migrations:**
```bash
cd services/telegram
migrate -path ./migrations -database "$DB_CONN" up        # Apply all
migrate -path ./migrations -database "$DB_CONN" down 1    # Rollback last
```

**View Telegram logs:**
```bash
cd services/telegram && make compose-logs
```

**Manage promotional codes via CLI:**
```bash
cd services/cli-admins
go build -o cli-admins cmd/main.go

# List codes
./cli-admins list

# Create code: 30 bonus days, max 100 uses, expires 2026-12-31
./cli-admins create WINTER30 30 100 2026-12-31T23:59:59Z

# Update code
./cli-admins update 1 60 200 2027-01-31T23:59:59Z

# Deactivate code
./cli-admins delete 1
```

## Data Flow: User Adds Peer

```
[Telegram Bot User]
        ↓
    /start, /menu
        ↓
[Telegram Service] ← Query PostgreSQL for subscription status
        ↓
   Is status active? → No → Show payment invoice
        ↓ Yes
   Call POST /peers → [AWG Service]
        ↓                   ↓
   Save connection ← Generate WireGuard peer
   metadata to DB          Generate public key
        ↓
    Call GET /peers/{id}/config → [AWG Service]
        ↓                             ↓
        ← Return peer config          Read disk file
        ↓
[Send .conf file to Telegram user]
```

## Payment Flow

1. User clicks "оплатить" button
2. Bot sends YooKassa invoice via Telegram's built-in payments
3. User completes checkout in Telegram UI
4. YooKassa sends `pre_checkout_query` → Bot confirms
5. YooKassa sends `successful_payment` → Bot updates subscription status in DB
6. Bot extends user's subscription duration

## Promo Code Flow

1. Admin uses CLI tool: `cli-admins create BONUS30 30 100 2026-12-31T23:59:59Z`
2. CLI sends HTTP POST to telegram admin API (`/admin/promo`)
3. Telegram service stores in PostgreSQL `promo_codes` table
4. User sends `/promo BONUS30` to Telegram bot
5. Bot validates:
   - Code exists and is active
   - Expiration date hasn't passed
   - Usage limit not reached
   - User hasn't already used this code
6. If user has no peer: create new one via AWG API
7. If user had expired peer: call `RestorePeer()` on AWG to re-enable it
8. Bot adds 30 bonus days to subscription expiration
9. Bot records activation in `promo_activations` table
10. User sees "✅ Промокод применен! +30 дней"

## Database Schema

The Telegram service manages these entities:

- **clients** — User accounts (Telegram ID, username, subscription status, test access flag)
- **payments** — Payment records (amount, date, status)
- **connections** — Peer lifecycle (host ID, public/preshared keys, creation date, expiration)
- **promo_codes** — Promotional codes (code, bonus days, max uses, expiration)
- **promo_activations** — Usage tracking (user-code activation records, prevents double-usage)

See `services/telegram/migrations/` for complete schema definitions.

## Project Status

### Completed (Production Ready)
- ✅ Multi-service architecture (AWG + Telegram + CLI-Admins)
- ✅ Payment integration (YooKassa)
- ✅ Peer lifecycle management with transactional delete/restore
- ✅ Configuration file delivery with preshared key support
- ✅ Subscription expiration checks (hourly automated)
- ✅ Promotional codes system with usage limits and expiration
- ✅ Admin API for promo code CRUD operations
- ✅ CLI tool for managing promotional codes
- ✅ Safe peer rollback on operation failures

### In Development
- 🔄 Graceful shutdown handlers for all services
- 🔄 Config deletion/archival policy
- 🔄 Enhanced error handling with context wrapping

### Not Yet Implemented
- Comprehensive unit/integration test suite
- Docker container images for production
- Multi-region deployment
- Admin dashboard for analytics
- Rate limiting on admin API
- Batch promo code import/export

## Documentation

- **Root architecture:** [`CLAUDE.md`](./CLAUDE.md) — Project structure and service communication
- **AWG service docs:** [`services/awg/CLAUDE.md`](services/awg/CLAUDE.md) — Peer management and transactional operations
- **AWG HTTP API:** [`services/awg/http_api.md`](services/awg/http_api.md) — Full API specification
- **Telegram service docs:** [`services/telegram/CLAUDE.md`](services/telegram/CLAUDE.md) — Bot, payments, and promo codes
- **CLI-Admins docs:** [`services/cli-admins/CLAUDE.md`](services/cli-admins/CLAUDE.md) — Admin tool for managing codes
- **Open tasks:** [`todo.md`](./todo.md) — Future development items

## Troubleshooting

### Bot doesn't respond
- Check `TELEGRAM_KEY` is correct
- Verify Telegram service is running: `ps aux | grep "go run"`
- Check logs: `cd services/telegram && make compose-logs`

### AWG connection fails
- Verify `HTTP_URL` in `.env` matches AWG listen address
- Check AWG service is running: `ps aux | grep awg`
- Ensure sudo credentials are available for AWG

### Database connection error
- Verify `DB_CONN` string is correct
- Check PostgreSQL is running: `docker-compose ps`
- Confirm database exists: `psql -U postgres -c "\l"`

### Payment tests return errors
- Verify `PROVIDER_TOKEN` is from YooKassa test mode
- Check YooKassa webhook configuration
- Review transaction history in YooKassa Dashboard

## Admin Tools

**Managing Promotional Codes:**

Use the CLI tool in `services/cli-admins/`:

```bash
cd services/cli-admins
go build -o cli-admins cmd/main.go

# Create promo code
./cli-admins create HOLIDAY50 50 500 2026-12-31T23:59:59Z

# List all codes with usage stats
./cli-admins list

# Update code parameters
./cli-admins update 1 60 600 2027-01-31T23:59:59Z

# Deactivate code
./cli-admins delete 1
```

Configure telegram service address:
```bash
ADDRESS=api.production.com:8080 ./cli-admins list
```

See `services/cli-admins/CLAUDE.md` for complete CLI documentation.

## Contributing

1. Create a feature branch
2. Ensure code passes `go fmt` and `go vet`
3. Test with Telegram sandbox bot if possible
4. Test promo code flow end-to-end if applicable
5. Submit PR with description

## Support

- **Bot Issues:** [@GopherSecureBot](https://t.me/GopherSecureBot) support commands (if implemented)
- **Code Issues:** See GitHub issues
- **Documentation:** Check [`CLAUDE.md`](./CLAUDE.md) for architecture questions

---

**GopherSecure** — Fast, secure VPN subscriptions via Telegram. Built with Go and AmneziaWireGuard.
