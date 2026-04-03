# Payment Service

## Overview

Payment Service is a microservice responsible for processing payments and validating transaction limits. It follows Clean Architecture principles and exposes a REST API consumed by the Order Service.

---

## Architecture

### Clean Architecture Layers

```
payment/
├── cmd/payment/main.go              → Composition Root (manual DI)
├── internal/
│   ├── domain/payment.go            → Entity, constants (no external dependencies)
│   ├── usecase/payment_usecase.go   → Business logic, payment authorization
│   ├── repository/                  → PaymentRepository interface (Port)
│   │   ├── payment_repository.go
│   │   └── postgres_payment_repository.go
│   ├── transport/http/              → Gin handlers (thin delivery layer)
│   │   └── payment_handler.go
│   └── app/app.go                   → Wiring: DB + DI
└── migrations/
    └── create_payments.sql
```

### Dependency Flow

```
Handler → UseCase → Repository Interface → Postgres Implementation
```

All dependencies point inward. The domain and use case layers have zero knowledge of HTTP, JSON, or database details.

---

## Domain Model

```go
type Payment struct {
    ID            string
    OrderID       string
    TransactionID string
    Amount        int64  // Amount in cents (e.g. 15000 = $150.00)
    Status        string // "Authorized", "Declined"
}
```

---

## API Endpoints

### POST /payments
Authorizes a payment for a given order.

**Request:**
```json
{
    "order_id": "uuid",
    "amount": 15000
}
```

**Flow:**
1. Validate amount (must be > 0)
2. Check transaction limit — decline if amount > 100000
3. Generate unique `transaction_id`
4. Store payment record in DB
5. Return status and transaction ID

**Response (201) — Authorized:**
```json
{
    "status": "Authorized",
    "transaction_id": "uuid"
}
```

**Response (201) — Declined:**
```json
{
    "status": "Declined",
    "transaction_id": "uuid"
}
```

**Error responses:**
- `400 Bad Request` — invalid request body
- `500 Internal Server Error` — database error

---

### GET /payments/{order_id}
Returns payment details for a given order.

**Response (200):**
```json
{
    "ID": "uuid",
    "OrderID": "uuid",
    "TransactionID": "uuid",
    "Amount": 15000,
    "Status": "Authorized"
}
```

**Error responses:**
- `404 Not Found` — payment does not exist for this order

---

## Business Rules

1. **Financial Accuracy**: `Amount` is stored as `int64` (cents). No `float64` is used anywhere.
2. **Transaction Limit**: If `amount > 100000` (1000 units), payment is automatically `Declined`.
3. **Unique Transaction ID**: Every payment record gets a unique `transaction_id` generated with UUID.
4. **Bounded Context**: Payment Service owns its own data exclusively. It does not share database or models with Order Service.

---

## Payment Limit Logic

```
amount <= 100000  →  Status: "Authorized"
amount >  100000  →  Status: "Declined"
```

Both cases store a record in the DB and return 201. The Order Service reads the status and updates the order accordingly.

---

## Dependency Inversion

The use case layer depends on an interface, not a concrete implementation:

```go
// Port — defined in repository layer
type PaymentRepository interface {
    Create(payment *domain.Payment) error
    GetByOrderID(orderID string) (*domain.Payment, error)
}
```

The concrete PostgreSQL implementation is injected at the Composition Root in `main.go`.

---

## Database

**Database**: PostgreSQL  
**Database name**: `paymentdb`

### Schema

```sql
CREATE TABLE IF NOT EXISTS payments (
    id             VARCHAR(36) PRIMARY KEY,
    order_id       VARCHAR(36) NOT NULL,
    transaction_id VARCHAR(36) NOT NULL,
    amount         BIGINT NOT NULL,
    status         VARCHAR(20) NOT NULL
);
```

---

## Running

### Prerequisites
- Go 1.22+
- PostgreSQL with `paymentdb` database created

### Setup database
```bash
psql -U postgres -d paymentdb -f migrations/create_payments.sql
```

### Run
```bash
go run cmd/payment/main.go
```

Service starts on **port 8081**.

---

## Configuration

Database DSN is configured in `cmd/payment/main.go`:

```go
dsn := "host=localhost port=5432 user=postgres password=YOUR_PASSWORD dbname=paymentdb sslmode=disable"
```

---

## Bounded Context

Payment Service is a fully independent bounded context:

- **Own database**: `paymentdb` — not shared with Order Service
- **Own domain model**: `domain.Payment` — not imported by Order Service
- **Own business rules**: transaction limits enforced here, not in Order Service
- **Stateless from Order's perspective**: Order Service calls `POST /payments` and reads the response — it never queries the payment DB directly