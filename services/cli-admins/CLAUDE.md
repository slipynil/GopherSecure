# CLAUDE.md — CLI-Admins Service

This file provides guidance for working with the `cli-admins` administrative command-line tool.

## Project Overview

**`cli-admins`** is a standalone command-line utility for system administrators to manage promotional codes in the GopherSecure VPN service. It communicates with the Telegram service's admin HTTP API to perform CRUD operations on promo codes.

**Key Use Cases:**
- Create promotional codes with customizable bonus days and usage limits
- Update code parameters on the fly
- View all active codes and their usage statistics
- Deactivate codes when promotions end

**Target Users:** VPN service operators and marketing teams

## Architecture & Components

```
cmd/main.go                 # CLI entry point and command dispatcher
                            # Parses arguments and routes to handlers
                            # Uses fmt and tabwriter for formatted output

internal/
  client/
    http.go                 # PromoClient struct for HTTP communication
                            # Methods: CreatePromo, UpdatePromo, ListPromos, DeletePromo
```

### Key Design Patterns

- **Stateless CLI**: No local state; all operations via HTTP to telegram service
- **Error Handling**: Validates input before sending requests
- **User Feedback**: Emoji indicators (✅, ❌, 📋) for command feedback
- **Table Formatting**: Uses `text/tabwriter` for aligned output

## Commands

### 1. Create Promo Code

```bash
cli-admins create <code> <bonus_days> <max_uses> <expires_at>
```

**Parameters:**
- `code` — Promo code string (e.g., "BONUS30", "SUMMER20")
- `bonus_days` — Days added to subscription on activation (integer)
- `max_uses` — Maximum activation count (0 = unlimited)
- `expires_at` — Code expiration date (RFC3339 format)

**Example:**
```bash
cli-admins create BONUS30 30 100 2026-03-29T23:59:59Z
```

**Output:**
```
✅ Промокод успешно создан:
  id: 1
  code: BONUS30
  bonus_days: 30
  max_uses: 100
```

**Validation:**
- Checks if date is valid RFC3339 format
- Checks if bonus_days and max_uses are integers
- Sends to telegram admin API for duplicate check

### 2. Update Promo Code

```bash
cli-admins update <id> <bonus_days> <max_uses> <expires_at>
```

**Parameters:**
- `id` — Promo code ID from database (integer)
- `bonus_days`, `max_uses`, `expires_at` — Same as create command

**Example:**
```bash
cli-admins update 1 60 200 2026-04-30T23:59:59Z
```

**Output:**
```
✅ Промокод #1 успешно обновлен:
  bonus_days: 60
  max_uses: 200
  expires_at: 2026-04-30T23:59:59Z
```

**Behavior:**
- Updates only specified code
- Does not reset usage counter (preserves audit trail)
- Validates RFC3339 date format

### 3. List All Promo Codes

```bash
cli-admins list
```

**Output:**
```
📋 Все промокоды:
─────────────────────────────────────────────────────────────────────────────────
ID  КОД      ДНИ  МАКС  ИСП  АКТИВНО  ИСТЕКАЕТ
─────────────────────────────────────────────────────────────────────────────────
1   BONUS30  30   100   45   true     2026-03-29
2   SUMMER20 20   0     12   true     2026-04-15
3   EXPIRED  15   50    50   false    2026-02-01
─────────────────────────────────────────────────────────────────────────────────
```

**Columns:**
- **ID** — Database promo code ID
- **КОД** — Promo code string
- **ДНИ** — Bonus days per activation
- **МАКС** — Max uses (0 = unlimited)
- **ИСП** — Current usage count
- **АКТИВНО** — Is code active and not expired?
- **ИСТЕКАЕТ** — Expiration date (formatted as YYYY-MM-DD)

**Empty Response:**
```
📭 Промокодов не найдено
```

### 4. Delete (Deactivate) Promo Code

```bash
cli-admins delete <id>
```

**Parameters:**
- `id` — Promo code ID (integer)

**Example:**
```bash
cli-admins delete 1
```

**Output:**
```
✅ Промокод #1 успешно удален:
  id: 1
  is_active: false
  deactivated_at: 2026-03-23T15:30:45Z
```

**Behavior:**
- Sets `is_active = false` (soft delete, preserves data)
- Existing user activations remain in database
- Code cannot be used for new activations
- Can be re-activated via update if needed

## Usage

### Local Development

Build from source:
```bash
cd services/cli-admins
go build -o cli-admins cmd/main.go
```

Run with default configuration (assumes telegram service on `0.0.0.0:8080`):
```bash
./cli-admins list
./cli-admins create TESTCODE 10 0 2026-12-31T23:59:59Z
```

### Custom Telegram Service Address

Override address via environment variable:
```bash
ADDRESS=api.example.com:8080 ./cli-admins list
ADDRESS=192.168.1.100:8080 ./cli-admins create PROMO1 30 100 2026-04-01T00:00:00Z
```

### Production Deployment

Build with ldflags to embed default address:
```bash
go build -ldflags "-X main.ldflagsAddr=api.production.com:8080" -o cli-admins cmd/main.go
```

Then run without ADDRESS env var:
```bash
./cli-admins list  # Uses api.production.com:8080
```

## HTTP API Integration

The CLI communicates with the Telegram service's admin API (see `services/telegram/internal/features/promocode/handler.go`):

### Request/Response Examples

**Create Promo Code:**
```
POST /admin/promo HTTP/1.1
Host: telegram-service:8080
Content-Type: application/json

{
  "code": "BONUS30",
  "bonus_days": 30,
  "max_uses": 100,
  "expires_at": "2026-03-29T23:59:59Z"
}

Response 201:
{
  "id": 1,
  "code": "BONUS30",
  "bonus_days": 30,
  "max_uses": 100,
  "used_count": 0,
  "is_active": true,
  "expires_at": "2026-03-29T23:59:59Z"
}
```

**List Promo Codes:**
```
GET /admin/promo HTTP/1.1
Host: telegram-service:8080

Response 200:
[
  {
    "id": 1,
    "code": "BONUS30",
    "bonus_days": 30,
    "max_uses": 100,
    "used_count": 45,
    "is_active": true,
    "expires_at": "2026-03-29T23:59:59Z"
  },
  ...
]
```

**Update Promo Code:**
```
PUT /admin/promo/1 HTTP/1.1
Host: telegram-service:8080
Content-Type: application/json

{
  "bonus_days": 60,
  "max_uses": 200,
  "expires_at": "2026-04-30T23:59:59Z"
}

Response 200: (same as create response)
```

**Delete Promo Code:**
```
DELETE /admin/promo/1 HTTP/1.1
Host: telegram-service:8080

Response 200:
{
  "id": 1,
  "is_active": false,
  "deactivated_at": "2026-03-23T15:30:45Z"
}
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ADDRESS` | `0.0.0.0:8080` | Telegram service admin API address |

**Format:** `ADDRESS=<host>:<port>`

Examples:
- `ADDRESS=localhost:8080` — Local development
- `ADDRESS=telegram.internal:8080` — Kubernetes internal DNS
- `ADDRESS=api.example.com:8080` — External FQDN with custom port

## Building & Running

### Prerequisites

- Go 1.26.0 or later
- Network access to Telegram service admin API

### Build

From `services/cli-admins/`:
```bash
go build -o cli-admins cmd/main.go
```

With ldflags for production:
```bash
go build \
  -ldflags "-X main.ldflagsAddr=api.production.com:8080" \
  -o cli-admins \
  cmd/main.go
```

### Run

```bash
./cli-admins list
./cli-admins create BONUS 30 100 2026-12-31T23:59:59Z
./cli-admins update 1 60 200 2026-12-31T23:59:59Z
./cli-admins delete 1
```

### Troubleshooting

**Connection Refused:**
```
❌ Ошибка: dial tcp: connection refused
```
- Check if telegram service is running
- Verify `ADDRESS` environment variable is correct
- Check firewall rules allow access to admin API port

**Invalid Arguments:**
```
❌ Использование: cli-admins create <code> <bonus_days> <max_uses> <expires_at>
```
- Verify all required arguments are provided
- Check integer arguments are numbers
- Validate date format is RFC3339 (e.g., `2026-03-29T23:59:59Z`)

**Date Format Error:**
```
❌ Неверный формат даты. Используйте RFC3339 (2026-03-29T23:59:59Z): ...
```
- Use RFC3339 format: `YYYY-MM-DDTHH:MM:SSZ`
- Example: `2026-03-29T23:59:59Z`

## Development Notes

### Code Structure

**cmd/main.go** (240 lines):
- `main()` — Entrypoint, command routing
- `handleCreate()`, `handleUpdate()`, `handleList()`, `handleDelete()` — Command handlers
- `printJSON()` — Utility to print key-value pairs
- `printUsage()` — Help text with Russian instructions

**internal/client/http.go**:
- `PromoClient` struct with `NewPromoClient(address string)` constructor
- Methods make HTTP requests to telegram admin API
- Error handling with user-friendly messages

### No Dependencies

The CLI uses only Go standard library:
- `fmt` — Output formatting
- `os` — Command-line arguments and exit codes
- `strconv` — String-to-integer conversion
- `text/tabwriter` — Aligned table output
- `time` — RFC3339 date parsing
- `net/http` — HTTP client (standard library)
- `encoding/json` — JSON marshaling/unmarshaling
- `io` — I/O utilities

### Testing

Currently no tests. When adding tests:
```bash
go test ./...
go test -v ./internal/client/...
```

### Code Style

```bash
# Format
go fmt ./...

# Vet
go vet ./...
```

## Integration with VPN Service

The CLI is one part of the promo code management system:

1. **Admin** uses CLI tool to create/update codes
2. **CLI** sends HTTP request to telegram service admin API
3. **Telegram service** stores in PostgreSQL (`promo_codes` table)
4. **Telegram Bot user** sends `/promo CODE` in chat
5. **Telegram service** validates code and applies bonus days
6. **User** gets extended subscription via promo activation

## Related Documentation

- **Parent project:** `services/../CLAUDE.md` — Project architecture
- **Telegram service:** `services/telegram/CLAUDE.md` — Admin API details
- **AWG service:** `services/awg/CLAUDE.md` — Peer management
- **Root README:** `../../README.md` — User-facing overview

## Future Enhancements

Potential improvements not yet implemented:

1. **Batch Operations** — Import/export promo codes as CSV
2. **Analytics** — Show detailed usage statistics per code
3. **Scheduling** — Automatic code creation/deactivation at specified times
4. **Rate Limiting** — Admin API protection against abuse
5. **Audit Logging** — Track who created/modified each code
6. **TOTP Auth** — Optional two-factor authentication for CLI
7. **Config File** — Use YAML/TOML instead of env vars for persistent settings
