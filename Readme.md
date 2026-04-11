# AP2 Assignment 2 – gRPC Migration & Contract-First Development

## Overview

In Assignment 2, the communication between Order Service and Payment Service was migrated from REST to gRPC. The Contract-First approach was adopted using Protocol Buffers with automated code generation via GitHub Actions.

---

## Repositories

| Repository | URL | Purpose |
|------------|-----|---------|
| Proto definitions | https://github.com/Sp206907/AP2-protos | Contains `.proto` files |
| Generated code | https://github.com/Sp206907/ap2-generated | Auto-generated `.pb.go` files |
| Main project | https://github.com/Sp206907/AP2_assignment1/tree/assignment2-grpc | Services source code |

---

## Architecture

```
┌─────────────────────────────────────┐         gRPC :50051        ┌──────────────────────────────────┐
│         ORDER SERVICE               │ ─────────────────────────► │       PAYMENT SERVICE            │
│         REST :8082 + gRPC :50052    │    ProcessPayment RPC       │       REST :8081 + gRPC :50051   │
│                                     │                             │                                  │
│  Handler → UseCase → Repository     │                             │  Handler → UseCase → Repository  │
│  GRPCPaymentClient (grpc client)    │                             │  PaymentGRPCServer (grpc server) │
│  OrderGRPCServer (streaming)        │                             │  LoggingInterceptor (bonus)      │
│               │                     │                             │               │                  │
│               ▼                     │                             │               ▼                  │
│          PostgreSQL                 │                             │          PostgreSQL               │
│           orderdb                   │                             │           paymentdb              │
└─────────────────────────────────────┘                             └──────────────────────────────────┘
         ▲ stream
         │ SubscribeToOrderUpdates
    gRPC Client
```

---

## Contract-First Flow

```
AP2-protos (repo A)          GitHub Actions           ap2-generated (repo B)
─────────────────            ──────────────           ──────────────────────
payment.proto    ──push──►   protoc + plugins  ──►   payment.pb.go
order.proto                  auto-generate           payment_grpc.pb.go
                                                     order.pb.go
                                                     order_grpc.pb.go
```

Services import generated code via:
```bash
go get github.com/Sp206907/ap2-generated@v1.0.0
```

---

## Proto Definitions

### PaymentService
```proto
service PaymentService {
  rpc ProcessPayment(PaymentRequest) returns (PaymentResponse);
}

message PaymentRequest {
  string order_id = 1;
  int64 amount = 2;
}

message PaymentResponse {
  string transaction_id = 1;
  string status = 2;
}
```

### OrderService (Streaming)
```proto
service OrderService {
  rpc SubscribeToOrderUpdates(OrderRequest) returns (stream OrderStatusUpdate);
}

message OrderRequest {
  string order_id = 1;
}

message OrderStatusUpdate {
  string order_id = 1;
  string status = 2;
  google.protobuf.Timestamp updated_at = 3;
}
```

---

## What Changed from Assignment 1

| Component | Assignment 1 | Assignment 2 |
|-----------|-------------|-------------|
| Order → Payment communication | REST HTTP | gRPC |
| Payment Service | REST only | REST + gRPC server |
| Order Service | REST only | REST + gRPC server (streaming) |
| Configuration | Hardcoded DSN | `.env` environment variables |
| Contract | Implicit REST | Explicit `.proto` contract |
| Code generation | None | GitHub Actions automated |

---

## Clean Architecture Preserved

- Domain entities and Use Cases from Assignment 1 are **unchanged**
- Only the transport layer was updated
- `GRPCPaymentClient` implements the same `PaymentClient` interface
- Use Cases have zero knowledge of gRPC

---

## Streaming

`SubscribeToOrderUpdates` is a Server-side Streaming RPC on the Order Service (port 50052).

- Client sends an `OrderRequest` with `order_id`
- Server polls the database every second
- When the order status changes in DB, the new status is pushed to the stream immediately
- Stream ends when the client disconnects

---

## Bonus — gRPC Interceptor (+10%)

A Logging Interceptor is implemented on the Payment Service. It logs every incoming gRPC request:

```
gRPC method: /payment.PaymentService/ProcessPayment | duration: 1.2ms | error: <nil>
```

---

## Configuration

### order/.env
```
REST_PORT=8082
ORDER_GRPC_PORT=50052
PAYMENT_GRPC_ADDR=localhost:50051
DB_DSN=host=localhost port=5432 user=postgres password=YOUR_PASSWORD dbname=orderdb sslmode=disable
```

### payment/.env
```
GRPC_PORT=50051
DB_DSN=host=localhost port=5432 user=postgres password=YOUR_PASSWORD dbname=paymentdb sslmode=disable
```

---

## Running

### Prerequisites
- Go 1.22+
- PostgreSQL with `orderdb` and `paymentdb` databases

### Start Payment Service
```bash
cd payment
go run cmd/payment/main.go
```
Starts REST on `:8081` and gRPC on `:50051`

### Start Order Service
```bash
cd order
go run cmd/order/main.go
```
Starts REST on `:8082` and gRPC streaming on `:50052`

---

## API Examples

### Create Order (REST → triggers gRPC to Payment)
```bash
curl -X POST http://localhost:8082/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "customer-1", "item_name": "Phone", "amount": 15000}'
```

### Get Recent Orders
```bash
curl http://localhost:8082/orders/recent?limit=5
```