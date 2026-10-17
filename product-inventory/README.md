# Product Inventory Management System

Go REST API + Angular frontend + PostgreSQL.

## Structure

```
backend/
  cmd/server/          entry point, wiring, startup migration
  internal/config/     env-based configuration
  internal/model/      domain types and payloads
  internal/repository/ SQL (pgx), constraint-to-error mapping
  internal/service/    validation and business rules
  internal/handler/    HTTP handlers, routes, error-to-status mapping
  internal/middleware/ CORS, request logging
  migrations/          schema + seed categories
frontend/
  src/app/core/        API service, models, error parsing
  src/app/products/    list page, add/edit form
```

## Run it

Requirements: Go 1.22+, Node 20+, Docker (or any PostgreSQL 13+).

```bash
# 1. Database
docker compose up -d

# 2. Backend (http://localhost:8080)
cd backend
cp .env.example .env        # optional; defaults match docker-compose
go mod tidy
go run ./cmd/server

# 3. Frontend (http://localhost:4200)
cd ../frontend
npm install
npm start
```

The schema and four seed categories (Electronics, Furniture, Stationery, Accessories) are applied automatically when the API starts.

Configuration (environment variables): `PORT`, `DATABASE_URL`, `ALLOWED_ORIGIN`, `LOW_STOCK_THRESHOLD` (default 10).

## API

| Method | Path | Description |
|---|---|---|
| GET | /api/categories | List categories |
| POST | /api/categories | Create category `{name}` |
| GET | /api/products?search=&category_id=&status= | List / search / filter |
| POST | /api/products | Create product |
| GET | /api/products/{id} | Get product |
| PUT | /api/products/{id} | Update product |
| DELETE | /api/products/{id} | Delete product |
| PATCH | /api/products/{id}/stock | Update stock `{mode: "set"\|"add"\|"remove", quantity}` |

Product body: `{ "name", "sku", "category_id", "price", "quantity", "status" }` where status is `active`, `inactive` or `discontinued`.

Status codes: 200 OK, 201 Created, 204 No Content (delete), 400 malformed request, 404 not found, 409 duplicate SKU/category, 422 validation failure, 500 unexpected error.

Errors look like `{ "error": "Validation failed", "fields": { "price": "Price must be greater than zero" } }`.

## Rules enforced

- Price must be greater than zero; quantity cannot be negative (checked in the service and again by database CHECK constraints).
- SKU is unique (normalised to upper case; database UNIQUE constraint, returns 409).
- Stock updates require both `mode` and `quantity`; a missing or `null` quantity is rejected with 422 (it is never treated as 0).
- Removing more stock than available is rejected with 422; stock changes are atomic in SQL.
- A product is flagged `low_stock` when quantity <= `LOW_STOCK_THRESHOLD`; the UI shows "Low stock" / "Out of stock" warnings.

## Quick API check

```bash
curl -X POST localhost:8080/api/products -H 'Content-Type: application/json' \
  -d '{"name":"USB-C Cable","sku":"usb-c-1m","category_id":1,"price":299,"quantity":25,"status":"active"}'
curl -X PATCH localhost:8080/api/products/1/stock -H 'Content-Type: application/json' \
  -d '{"mode":"remove","quantity":20}'
```

## Tests

```bash
cd backend
go test ./...
```

Covers the validation rules (price > 0, quantity >= 0, required fields, status, SKU normalisation), stock-update validation (missing/negative quantity, bad mode), the low-stock rule (`<= 10`), and the error-to-HTTP-status mapping (422, 409, 404, 500).

## Using a local PostgreSQL instead of Docker

Set `DATABASE_URL` before starting the API, for example (PowerShell):

```powershell
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5433/inventory?sslmode=disable"
go run ./cmd/server
```
