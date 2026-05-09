# SPE - Payment Services POC

Payment Services POC:
- Auth service on port 9090
- Payment service on port 9091
- PostgreSQL, Redis, and Kafka

## Project Structure

```text
SPE/
|- auth-services/
|- payment-services/
|- database/
|  |- migrations/
|  |  |- auth/
|  |  |- payment/
|  |- seed/
|- postman/
|- docker-compose.yml
```

## Prerequisites

- Docker Desktop
- Go 1.25+

## How To Run

### 1. Start infrastructure only

Use this for first run.

```bash
docker compose up -d --build postgres redis kafka
```

### 2. Create database if not exists

```bash
docker compose exec -T postgres psql -U postgres -d postgres -tc "SELECT 1 FROM pg_database WHERE datname='spe'" | grep -q 1 || docker compose exec -T postgres psql -U postgres -d postgres -c "CREATE DATABASE spe;"
```

### 3. Run migrations

```bash
# Auth schema
docker compose exec -T postgres psql -U postgres -d spe -f /dev/stdin < database/migrations/auth/001_init_auth.sql

# Payment schema
docker compose exec -T postgres psql -U postgres -d spe -f /dev/stdin < database/migrations/payment/001_init_payment.sql
docker compose exec -T postgres psql -U postgres -d spe -f /dev/stdin < database/migrations/payment/002_optimize_indexes_payment.sql
```

### 4. Run seed

```bash
docker compose exec -T postgres psql -U postgres -d spe -f /dev/stdin < database/seed/auth_seed.sql
docker compose exec -T postgres psql -U postgres -d spe -f /dev/stdin < database/seed/payment_seed.sql
```

### 5. Start app services

```bash
docker compose up -d auth-service payment-service
```

### 6. Health checks

```bash
curl http://localhost:9090/ping
curl http://localhost:9091/ping
```

## Alternative: Start Everything at Once

If your database volume is already initialized and schema already exists:

```bash
docker compose up -d --build
```

## Environment Variables

### auth-services/.env

| Variable | Default | Description |
|---|---|---|
| AUTH_PORT | 9090 | Auth service port |
| DATABASE_URL | postgres://postgres:postgres@localhost:5432/spe?sslmode=disable | PostgreSQL connection string |
| REDIS_ADDR | localhost:6379 | Redis address |
| JWT_SECRET | super-secret-jwt-key | JWT signing secret |
| JWT_TTL_MINUTES | 15 | Token lifetime in minutes |
| RATE_LIMIT_MAX_ATTEMPTS | 5 | Max requests per window |
| RATE_LIMIT_WINDOW_MINUTES | 10 | Rate limit window in minutes |
| AUTH_LOGIN_ENABLED | false | Feature Flag for Enable or disable login endpoint |

### payment-services/.env

| Variable | Default | Description |
|---|---|---|
| PAYMENT_PORT | 8081 | Payment service port |
| DATABASE_URL | postgres://postgres:postgres@localhost:5432/spe?sslmode=disable | PostgreSQL connection string |
| REDIS_ADDR | localhost:6379 | Redis address |
| JWT_SECRET | super-secret-jwt-key | Must match auth service JWT secret |
| KAFKA_BROKERS | localhost:9092 | Kafka broker |
| KAFKA_TOPIC | payment.transaction-events | Kafka topic |
| CHECK_STATUS_CACHE_TTL_SECONDS | 30 | Cache TTL for check-status |

Important: `JWT_SECRET` must be the same in auth and payment services.

## API Endpoints

### Auth Service (http://localhost:9090)

#### Issue Token

```http
POST /api/v1/auth/token
Content-Type: application/json
```

Request:

```json
{
  "merchant_id": "008800223497",
  "merchant_secret": "yourSecret"
}
```

Success response:

```json
{
  "code": "00",
  "message": "success",
  "access_token": "<jwt>",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### Payment Service (http://localhost:9091)

Required headers:

```text
Authorization: Bearer <access_token>
X-Signature: <HMAC-SHA512 signature>
```

#### Transaction Notification

```http
POST /api/v1/transaction-notification
```

Signature format:

```text
HMAC-SHA512("{request_id}:{rrn}:{merchant_id}", signing_key)
```

Body:

```json
{
  "request_id": "XwVjF5zfuHhrDZuw",
  "customer_pan": "9360001110000000019",
  "amount": 10000.00,
  "transaction_datetime": "2021-02-25T13:36:13Z",
  "rrn": "123456789012",
  "bill_number": "12345678901234567890",
  "customer_name": "John Doe",
  "merchant_id": "008800223497",
  "merchant_name": "Sukses Makmur Bendungan Hilir",
  "merchant_city": "Jakarta Pusat",
  "currency_code": "360",
  "payment_status": "00",
  "payment_description": "Payment Success"
}
```

#### Check Status

```http
POST /api/v1/check-status
```

Signature format:

```text
HMAC-SHA512("{bill_number}", signing_key)
```

Body:

```json
{
  "request_id": "XwVjF5zfuHhrDZuw",
  "bill_number": "12345678901234567890"
}
```
## Error Codes

### Auth Service

| Code | HTTP | Description |
|---|---|---|
| 00 | 200 | Success |
| 01 | 400 | Invalid request payload |
| 02 | 401 | Invalid merchant credentials |
| 03 | 503 | Auth service temporarily disabled |

### Payment Service

| Code | HTTP | Description |
|---|---|---|
| 00 | 200 | Success |
| 01 | 400 | Invalid request payload |
| 02 | 401 | Invalid or expired bearer token |
| 03 | 401 | Merchant token mismatch |
| 04 | 401 | Merchant not registered |
| 05 | 401 | Invalid signature |
| 06 | 409 | Bill number already used |
| 14 | 404 | Transaction not found |
| 99 | 500 | Internal server error |
