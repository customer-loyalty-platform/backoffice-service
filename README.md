# Loyalty Platform BackOffice Service (Golang)

Single-deployable Modular Monolith backend written in Go 1.21+ using the Hexagonal Architecture pattern. It provides REST APIs for administrative operations on the Loyalty Platform Command Center.

---

## 🚀 Features

- **Hexagonal / Clean Architecture:** Clear isolation between HTTP delivery, middleware, business use cases, and database repositories.
- **Connection Pooling:** High-performance connection pools using `jackc/pgx/v5` with Master / Read-Replica splitting.
- **Dynamic Earning Rule Meta-Validator:** Strict JSON Schema validation using `santhosh-tekuri/jsonschema/v5`.
- **JWT & RBAC Security:** 60-minute scoped Admin JWT access tokens + refresh token flow with role-based permission enforcement (`SUPER_ADMIN`, `CAMPAIGN_SPEC`, `OPS_CS`, `FINANCE`).
- **Automated Audit Logging:** State-modifying requests (`POST`, `PUT`, `PATCH`, `DELETE`) are captured asynchronously in `schema_backoffice.admin_audit_logs`.
- **Maker-Checker Workflows:** Dual-control approval pipeline for sensitive financial and configuration mutations.
- **Zero Calculation Principle:** Configuration, rules, tiers, and vouchers only — point execution is strictly separated.
- **Interactive Documentation:** Comprehensive endpoint and payload reference available in [`API_DOCUMENTATION.md`](./API_DOCUMENTATION.md).

---

## 📁 Project Structure

```text
backoffice-service/
├── cmd/
│   ├── migrate/main.go          # Database migration runner
│   └── server/main.go           # Application entrypoint & graceful shutdown
├── config/                      # Viper configuration & .env loader
├── internal/
│   ├── delivery/http/           # Gin HTTP Router & endpoints
│   ├── domain/                  # Domain entities & pagination helpers
│   ├── middleware/              # Auth, RBAC, Recovery, Audit logging
│   ├── modules/                 # Modular Bounded Contexts
│   │   ├── admin_user/          # Super admin user management
│   │   ├── auth/                # Login, Refresh, Logout, Me
│   │   ├── campaign/            # Missions & Challenges
│   │   ├── earning_rule/        # Dynamic JSON Earning Rule Builder
│   │   ├── maker_checker/       # Dual-control approvals
│   │   ├── member/              # Ledger member profiles & adjustments
│   │   ├── monitoring/          # Dashboard analytics & audit logs
│   │   ├── settlement/          # Partner clearing & settlement reports
│   │   ├── tier/                # Tier definitions & perks
│   │   └── voucher_catalog/     # Voucher batches & inventory pool
│   └── pkg/                     # Shared utilities (DB, JWT, logger, validator, response)
├── migrations/                  # SQL Schema migrations
├── Dockerfile                   # Multi-stage Docker build (<25MB)
├── Makefile                     # Build & run automation
├── go.mod
└── go.sum
```

---

## ⚙️ Configuration (.env)

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | Runtime environment (`development` or `production`) |
| `APP_PORT` | `8080` | HTTP server port |
| `DB_MASTER_DSN` | `postgres://postgres:postgres@localhost:5432/loyalty_databases?sslmode=disable` | Master PostgreSQL DSN |
| `DB_REPLICA_DSN` | `postgres://postgres:postgres@localhost:5432/loyalty_databases?sslmode=disable` | Read-replica PostgreSQL DSN |
| `JWT_SECRET` | `loyalty-backoffice-jwt-secret-key-2024-secure` | Secret key for signing admin tokens |
| `JWT_ACCESS_EXPIRY` | `60m` | Access token lifespan |
| `JWT_REFRESH_EXPIRY` | `168h` | Refresh token lifespan (7 days) |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,http://localhost:5173` | Allowed frontend origins |

---

## 🛠️ Getting Started

### 1. Build the Binary
```bash
go build -o bin/backoffice-api.exe ./cmd/server/main.go
```

### 2. Run Database Migrations
When PostgreSQL is running:
```bash
go run ./cmd/migrate/main.go
```

### 3. Start the Server
```bash
./bin/backoffice-api.exe
# Or directly via Go
go run ./cmd/server/main.go
```

Server will be running at `http://localhost:8080`.

### 4. Run Unit Tests
```bash
go test ./... -v
```

---

## 🔐 Default Credentials (from Seed)

- **Username:** `superadmin`
- **Password:** `password123`
- **Role:** `SUPER_ADMIN` (`ALL_ACCESS`)

