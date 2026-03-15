# GopherSecure VPN Service

A production VPN infrastructure for purchasing and managing AmneziaWireGuard (AWG) VPN subscriptions through Telegram.

**Telegram Bot:** [@GopherSecureBot](https://t.me/GopherSecureBot)

## Overview

GopherSecure provides a complete VPN subscription service built on **AmneziaWireGuard** (a privacy-enhanced WireGuard fork). Users subscribe, pay via YooKassa integration with Telegram, and instantly receive VPN configurations through the Telegram bot.

### Key Features

- 🤖 **Telegram Bot Integration** — User-friendly subscription management via [@GopherSecureBot](https://t.me/GopherSecureBot)
- 💳 **YooKassa Payments** — Seamless in-Telegram payment processing with invoice handling
- 🔐 **AmneziaWireGuard Peers** — Automatic WireGuard peer creation and configuration management
- 📊 **PostgreSQL Database** — Persistent storage of clients, payments, connections, and peer configs
- ⏰ **Subscription Lifecycle** — Automated expiration checks and peer cleanup
- 🔐 **Peer Obfuscation** — Support for AWG obfuscation parameters (jitter, shift, hash)
- 📝 **Config File Delivery** — Users download `.conf` files directly from Telegram

### Technology Stack

- **Go 1.26.0** — Core services
- **PostgreSQL 18+** — Data persistence
- **AmneziaWireGuard** — VPN protocol
- **Telegram Bot API** — User interface
- **YooKassa** — Payment processing
- **Docker & Docker Compose** — Local development and deployment

## Architecture

GopherSecure is a **two-service monorepo**:

### 1. AWG Service (`services/awg/`)
Low-level WireGuard peer management via HTTP API.

**Responsibilities:**
- Interface with AmneziaWireGuard daemon
- Create/delete WireGuard peers
- Generate and serve peer configuration files
- Apply obfuscation parameters

**Endpoints:**
- `POST /peers` — Create peer
- `DELETE /peers` — Remove peer
- `GET /peers/{id}/config` — Download config file

See [`services/awg/http_api.md`](services/awg/http_api.md) for full API specification.

### 2. Telegram Service (`services/telegram/`)
User-facing bot and business logic.

**Responsibilities:**
- Handle Telegram user interactions (`/start`, `/menu`)
- Process payments via Stripe pre-checkout and success callbacks
- Call AWG service to create/delete peers for users
- Manage subscription lifecycle in PostgreSQL
- Schedule hourly expiration checks
- Send VPN configs to users

**Main Components:**
- `cmd/main.go` — Service initialization
- `internal/telegram/` — Telegram Bot API wrapper
- `internal/service/` — Business logic and event loop
- `internal/repository/` — PostgreSQL access layer
- `internal/httpClient/` — AWG API client
- `migrations/` — Database schema

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
make awg-run          # Terminal 1: AWG service (requires sudo)
make pay-run          # Terminal 2: Telegram service
make compose-up       # Terminal 3: PostgreSQL (in services/telegram)
```

**Code formatting & linting:**
```bash
cd services/awg && go fmt ./... && go vet ./...
cd services/telegram && go fmt ./... && go vet ./...
```

**Database migrations:**
```bash
cd services/telegram
migrate -path ./migrations -database "$DB_CONN" up        # Apply
migrate -path ./migrations -database "$DB_CONN" down 1    # Rollback
```

**View Telegram logs:**
```bash
cd services/telegram && make compose-logs
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

## Database Schema

The Telegram service manages three main entities:

- **clients** — User accounts (Telegram ID, subscription status)
- **payments** — Payment records (amount, date, status)
- **connections** — Peer lifecycle (hostID, public key, creation date, expiration)

See `services/telegram/migrations/` for schema definitions.

## Project Status

### Completed
- ✅ Multi-service architecture (AWG + Telegram)
- ✅ Payment integration (Stripe)
- ✅ Peer lifecycle management
- ✅ Configuration file delivery
- ✅ Subscription expiration checks

### In Development (see `todo.md`)
- 🔄 Graceful shutdown for AWG service
- 🔄 Ability to delete old configs
- 🔄 Enhanced error handling for `/start` command

### Not Yet Implemented
- Unit/integration tests
- Docker container images for production
- Multi-region deployment
- User analytics dashboard

## Documentation

- **Root architecture:** [`CLAUDE.md`](./CLAUDE.md)
- **AWG service docs:** [`services/awg/CLAUDE.md`](services/awg/CLAUDE.md)
- **AWG HTTP API:** [`services/awg/http_api.md`](services/awg/http_api.md)
- **Telegram service docs:** [`services/telegram/CLAUDE.md`](services/telegram/CLAUDE.md)
- **Open tasks:** [`todo.md`](./todo.md)

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

## Contributing

1. Create a feature branch
2. Ensure code passes `go fmt` and `go vet`
3. Test with Telegram sandbox bot if possible
4. Submit PR with description

## Support

- **Bot Issues:** [@GopherSecureBot](https://t.me/GopherSecureBot) support commands (if implemented)
- **Code Issues:** See GitHub issues
- **Documentation:** Check [`CLAUDE.md`](./CLAUDE.md) for architecture questions

---

**GopherSecure** — Fast, secure VPN subscriptions via Telegram. Built with Go and AmneziaWireGuard.
