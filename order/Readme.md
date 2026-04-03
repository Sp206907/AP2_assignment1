# Order Service

## Overview

Order Service is a microservice responsible for managing customer orders and their state transitions. It follows Clean Architecture principles and communicates with the Payment Service via REST to authorize payments.

---

## Architecture

### Clean Architecture Layers

```
order/
├── cmd/order/main.go           → Composition Root (manual DI)
├── internal/
│   ├── domain/order.go         → Entity, constants (no external dependencies)
│   ├── usecase/order_usecase.go→ Business logic, state transitions
│   ├── usecase/payment_client.go → PaymentClient interface (Port)
│   ├── repository/             → OrderRepository interface (Port)
│   │   ├── order_repository.go
│   │   └── postgres_order_repository.go
│   ├── transport/http/         → Gin handlers (thin delivery layer)
│   │   ├── order_handler.go
│   │   └── payment_client.go   → HTTP adapter for PaymentClient
│   └── app/app.go              → Wiring: DB + HTTP client + DI
└── migrations/
    └── create_orders.sql
```

### Dependency Flow

```
Handler → UseCase → Repository Interface → Postgres Implementation
                 → PaymentClient Interface → HTTP Implementation
```

All dependencies point inward. The domain and use case layers have zero knowledge of HTTP, JSON, or database details.

---

## Domain Model

```go
type Order struct {
    ID             string
    CustomerID     string
    ItemName       string
    Amount         int64     // Amount in cents (e.g. 15000 = $150.00)
    Status         string    // "Pending", "Paid", "Failed", "Cancelled"
    CreatedAt      time.Time
    IdempotencyKey string
}
```

### Status Transitions

```
Pending → Paid       (Payment authorized)
Pending → Failed     (Payment declined or service unavailable)
Pending → Cancelled  (Manual cancellation)
Paid    → (terminal) (Cannot be changed)
```

---

## API Endpoints

### POST /orders
Creates a new order and synchronously authorizes payment.

**Request:**
```json
{
    "customer_id": "customer-1",
    "item_name": "Phone",
    "amount": 15000
}
```

**Headers (optional):**
```
Idempotency-Key: unique-key-123
```

**Flow:**
1. Validate request (amount > 0)
2. Check idempotency key — return existing order if found
3. Create order with status `Pending` in DB
4. Call Payment Service `POST /payments` with 2-second timeout
5. Update order status to `Paid` or `Failed` based on response

**Response (201):**
```json
{
    "ID": "uuid",
    "CustomerID": "customer-1",
    "ItemName": "Phone",
    "Amount": 15000,
    "Status": "Paid",
    "CreatedAt": "2026-04-01T08:00:00Z",
    "IdempotencyKey": "unique-key-123"
}
```

**Error responses:**
- `400 Bad Request` — invalid request body or amount <= 0
- `503 Service Unavailable` — Payment Service is down or timed out

---

### GET /orders/{id}
Returns order details by ID.

**Response (200):**
```json
{
    "ID": "uuid",
    "CustomerID": "customer-1",
    "ItemName": "Phone",
    "Amount": 15000,
    "Status": "Paid",
    "CreatedAt": "2026-04-01T08:00:00Z"
}
```

**Error responses:**
- `404 Not Found` — order does not exist

---

### PATCH /orders/{id}/cancel
Cancels an order. Only `Pending` orders can be cancelled.

**Response (200):**
```json
{
    "message": "order cancelled"
}
```

**Error responses:**
- `400 Bad Request` — order is `Paid` or already `Cancelled`

---

## Business Rules

1. **Financial Accuracy**: `Amount` is stored as `int64` (cents). No `float64` is used anywhere.
2. **Amount Validation**: Amount must be greater than 0.
3. **Cancellation Rule**: Only `Pending` orders can be cancelled. `Paid` orders are immutable.
4. **Timeout**: HTTP client for Payment Service has a strict 2-second timeout.
5. **Idempotency**: If `Idempotency-Key` header is provided, duplicate requests return the existing order without creating a new one.

---

## Failure Handling

### Payment Service Unavailable

| Scenario | Behaviour |
|----------|-----------|
| Payment Service is down | Timeout trips after 2s, order marked `Failed`, returns 503 |
| Payment Service returns `Declined` | Order marked `Failed`, returns 201 with Failed status |
| Payment Service returns `Authorized` | Order marked `Paid`, returns 201 |

**Design Decision**: The order is marked as `Failed` (not left as `Pending`) when the Payment Service is unavailable. This is because an unconfirmed payment state is ambiguous — the client receives a clear signal to retry with a new request.

---

## Dependency Inversion

The use case layer depends on interfaces, not concrete implementations:

```go
// Port — defined in usecase layer
type PaymentClient interface {
    Authorize(req PaymentRequest) (*PaymentResponse, error)
}

type OrderRepository interface {
    Create(order *domain.Order) error
    GetByID(id string) (*domain.Order, error)
    UpdateStatus(id string, status string) error
    GetByIdempotencyKey(key string) (*domain.Order, error)
}
```

Concrete implementations (PostgreSQL, HTTP) are injected at the Composition Root in `main.go`.

---

## Database

**Database**: PostgreSQL  
**Database name**: `orderdb`

### Schema

```sql
CREATE TABLE IF NOT EXISTS orders (
    id              VARCHAR(36) PRIMARY KEY,
    customer_id     VARCHAR(36) NOT NULL,
    item_name       VARCHAR(255) NOT NULL,
    amount          BIGINT NOT NULL,
    status          VARCHAR(20) NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    idempotency_key VARCHAR(64)
);

CREATE UNIQUE INDEX orders_idempotency_key_idx
    ON orders (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
```

---

## Running

### Prerequisites
- Go 1.22+
- PostgreSQL with `orderdb` database created

### Setup database
```bash
psql -U postgres -d orderdb -f migrations/create_orders.sql
```

### Run
```bash
go run cmd/order/main.go
```

Service starts on **port 8082**.

---

## Configuration

Database DSN and Payment Service URL are configured in `cmd/order/main.go`:

```go
dsn := "host=localhost port=5432 user=postgres password=YOUR_PASSWORD dbname=orderdb sslmode=disable"
paymentServiceURL := "http://localhost:8081"
```