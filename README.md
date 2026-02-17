# Merch Store

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-Minikube-326CE5?style=flat&logo=kubernetes)](https://kubernetes.io/)
[![CI](https://img.shields.io/github/actions/workflow/status/Lexv0lk/merch-store/go.yml?branch=main&label=CI&logo=github)](https://github.com/Lexv0lk/merch-store/actions)
[![Coverage](https://img.shields.io/badge/Coverage-85.9%25-brightgreen?style=flat)](./)

A microservices-based merchandise store where employees can purchase items with coins and transfer coins to each other. Built with Go, gRPC, and PostgreSQL.

> Inspired by the [Avito Backend Internship Assignment (Winter 2025)](https://github.com/avito-tech/tech-internship/blob/main/Tech%20Internships/Backend/Backend-trainee-assignment-winter-2025/Backend-trainee-assignment-winter-2025.md), but redesigned as a microservices architecture with gRPC inter-service communication, separate databases per service, and clean architecture principles.

## Features

- **Microservices Architecture** — Three independent services communicating via gRPC
- **REST API Gateway** — HTTP interface translating requests to gRPC calls
- **JWT Authentication** — HS256 tokens with automatic user registration on first login
- **Secure Password Hashing** — Argon2id for password storage
- **Coin Economy** — Transfer coins between users with concurrent-safe transactions
- **Merchandise Shop** — 10 items available for purchase at fixed coin prices
- **Database Per Service** — Separate PostgreSQL instances for Auth and Store
- **Database Migrations** — Automatic schema management with Goose
- **Kubernetes Ready** — Full k8s manifests for deployment with Minikube (Ingress, StatefulSets, Jobs)
- **CI/CD** — Automated testing and Docker validation with GitHub Actions

## Architecture

```
┌──────────┐      ┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│  Client  │─────▶│   Gateway    │─────▶│ Auth Service │─────▶│ PostgreSQL   │
│  (HTTP)  │      │  (Gin REST)  │      │   (gRPC)     │      │  (Auth DB)   │
└──────────┘      │  :8080       │      │  :9090       │      │  :5433       │
                  │              │      └──────────────┘      └──────────────┘
                  │              │
                  │              │      ┌──────────────┐      ┌──────────────┐
                  │              │─────▶│Store Service │─────▶│ PostgreSQL   │
                  │              │      │   (gRPC)     │      │  (Store DB)  │
                  └──────────────┘      │  :9091       │      │  :5434       │
                                        └──────────────┘      └──────────────┘
```

### Key Design Decisions

- **Clean Architecture** — Each service follows Domain / Application / Infrastructure layering
- **gRPC Communication** — Type-safe, high-performance inter-service calls via Protocol Buffers
- **Database Per Service** — Auth and Store have isolated PostgreSQL databases
- **Row-Level Locking** — `SELECT ... FOR UPDATE` with `ReadCommitted` isolation for safe concurrent coin transfers and purchases
- **JWT via gRPC Metadata** — Gateway extracts Bearer tokens and forwards them through gRPC metadata

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/api/auth` | No | Authenticate (auto-registers on first login) |
| `GET` | `/api/info` | Yes | Get balance, inventory, and coin history |
| `POST` | `/api/sendCoin` | Yes | Transfer coins to another user |
| `GET` | `/api/buy/:item` | Yes | Purchase a merchandise item |

### Examples

**Authenticate:**
```bash
curl -X POST http://localhost:8080/api/auth \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "password": "secret123"}'
```
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Get User Info:**
```bash
curl http://localhost:8080/api/info \
  -H "Authorization: Bearer <token>"
```
```json
{
  "balance": 900,
  "inventory": [
    { "name": "cup", "quantity": 2 },
    { "name": "pen", "quantity": 1 }
  ],
  "coinHistory": {
    "received": [
      { "fromUsername": "bob", "amount": 100 }
    ],
    "sent": [
      { "toUsername": "charlie", "amount": 50 }
    ]
  }
}
```

**Send Coins:**
```bash
curl -X POST http://localhost:8080/api/sendCoin \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"toUser": "bob", "amount": 50}'
```

**Buy Item:**
```bash
curl http://localhost:8080/api/buy/t-shirt \
  -H "Authorization: Bearer <token>"
```

### Available Merchandise

| Item | Price (coins) |
|------|--------------|
| pen | 10 |
| socks | 10 |
| cup | 20 |
| book | 50 |
| wallet | 50 |
| t-shirt | 80 |
| powerbank | 200 |
| umbrella | 200 |
| hoody | 300 |
| pink-hoody | 500 |

## Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Language** | Go 1.25 | Core application |
| **HTTP Framework** | Gin | REST API Gateway |
| **RPC Framework** | gRPC + Protobuf | Inter-service communication |
| **Database** | PostgreSQL 16 | Persistent storage |
| **Migrations** | Goose | Schema management |
| **Auth** | JWT (HS256) + Argon2id | Authentication & password hashing |
| **Testing** | testify + testcontainers | Unit & integration tests |
| **Mocking** | golang/mock + pgxmock | Interface mocking |
| **Linting** | golangci-lint | Code quality |
| **Containerization** | Docker Compose | Deployment & local development |

## Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.25+ (for local development)

### Quick Start

1. **Clone the repository:**
```bash
git clone https://github.com/Lexv0lk/merch-store.git
cd merch-store
```

2. **Configure environment:**
```bash
cp .env.example .env
# Fill in DB_AUTH_USER, DB_AUTH_PASSWORD, DB_STORE_USER, DB_STORE_PASSWORD, JWT_SECRET
```

3. **Start all services:**
```bash
make build-app-up
```

4. **The API is now available at `http://localhost:8080`**

### Development Setup

1. **Start databases only:**
```bash
make infra-up
```

2. **Run services locally:**
```bash
go run cmd/auth/main.go
go run cmd/store/main.go
go run cmd/gateway/main.go
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `DB_AUTH_HOST` | Auth database host |
| `DB_AUTH_PORT` | Auth database port |
| `DB_AUTH_USER` | Auth database user |
| `DB_AUTH_PASSWORD` | Auth database password |
| `DB_AUTH_NAME` | Auth database name |
| `DB_STORE_HOST` | Store database host |
| `DB_STORE_PORT` | Store database port |
| `DB_STORE_USER` | Store database user |
| `DB_STORE_PASSWORD` | Store database password |
| `DB_STORE_NAME` | Store database name |
| `GRPC_AUTH_PORT` | Auth gRPC port |
| `GRPC_STORE_PORT` | Store gRPC port |
| `GRPC_AUTH_HOST` | Auth gRPC host (for gateway) |
| `GRPC_STORE_HOST` | Store gRPC host (for gateway) |
| `HTTP_PORT` | Gateway HTTP port |
| `JWT_SECRET` | Secret key for JWT signing |

## Testing

The project maintains **85.9% test coverage**. Run all tests with race detection and coverage:
```bash
make test
```

**Unit tests** cover all layers — handlers, use cases, gRPC adapters, and repositories — using interface-based mocking with golang/mock and pgxmock.

**Integration tests** use testcontainers to spin up real PostgreSQL instances and run full scenarios:
- Coin transfer flow (auth, send coins, verify balances and history)
- Merchandise purchase flow (auth, buy items, verify inventory and balance)

## Continuous Integration

GitHub Actions pipeline runs on every push and PR to `main`:

1. **Test** — Runs all tests with race detection, generates coverage report
2. **Docker Compose** — Validates configuration, builds images, starts services, and verifies health

## Project Structure

```
├── api/merch/v1/               # Protobuf definitions
│   ├── auth.proto
│   └── store.proto
├── build/package/              # Dockerfiles per service
│   ├── auth/
│   ├── gateway/
│   └── store/
├── cmd/                        # Service entry points
│   ├── auth/
│   ├── gateway/
│   └── store/
├── gen/                        # Generated code (proto stubs, mocks)
│   ├── merch/v1/
│   └── mocks/
├── internal/
│   ├── auth/                   # Auth microservice
│   │   ├── application/        #   Use cases
│   │   ├── bootstrap/          #   Init & config
│   │   ├── domain/             #   Interfaces & entities
│   │   ├── grpc/               #   gRPC handlers
│   │   └── infrastructure/     #   PostgreSQL repos
│   ├── gateway/                # API Gateway
│   │   ├── bootstrap/
│   │   ├── domain/
│   │   ├── grpc/               #   gRPC client adapters
│   │   └── infrastructure/     #   HTTP handlers & middleware
│   ├── pkg/                    # Shared utilities
│   │   ├── database/
│   │   ├── env/
│   │   ├── jwt/
│   │   └── logging/
│   └── store/                  # Store microservice
│       ├── application/
│       ├── bootstrap/
│       ├── domain/
│       ├── grpc/
│       └── infrastructure/
├── migrations/                 # Goose SQL migrations
│   ├── auth/
│   └── store/
├── tests/integration/          # Integration tests
├── docker-compose.yml
├── Makefile
└── .golangci.yml
```

## Kubernetes Deployment

The project can also be deployed to a local Kubernetes cluster using Minikube.

### K8s Structure

```
k8s/
├── namespace.yaml              # merch-store namespace
├── configmap.yaml              # Shared configuration
├── secrets.yaml                # DB credentials, JWT secret
├── auth/
│   ├── deployment.yaml         # Auth service deployment
│   └── service.yaml            # Auth ClusterIP service
├── store/
│   ├── deployment.yaml         # Store service deployment
│   └── service.yaml            # Store ClusterIP service
├── gateway/
│   ├── deployment.yaml         # Gateway deployment
│   ├── service.yaml            # Gateway ClusterIP service
│   └── ingress.yaml            # Ingress (merch-store.local)
├── postgres-auth/
│   ├── statefulset.yaml        # Auth PostgreSQL StatefulSet
│   ├── service.yaml            # Auth DB service
│   └── pvc.yaml                # Auth DB persistent volume
├── postgres-store/
│   ├── statefulset.yaml        # Store PostgreSQL StatefulSet
│   ├── service.yaml            # Store DB service
│   └── pvc.yaml                # Store DB persistent volume
└── migrations/
    ├── job-auth.yaml           # Auth DB migration Job
    └── job-store.yaml          # Store DB migration Job
```

### Running with Minikube

1. **Install and start Minikube:**
```bash
minikube start
```

2. **Enable the Ingress addon:**
```bash
minikube addons enable ingress
```

3. **Apply all manifests:**
```bash
make k8s-deploy
```

4. **Start Minikube tunnel** (requires admin/root privileges):
```bash
minikube tunnel
```

5. **Add the hostname to your `hosts` file:**
```
127.0.0.1 merch-store.local
```
- **Windows:** `C:\Windows\System32\drivers\etc\hosts`
- **Linux/macOS:** `/etc/hosts`

6. **The API is now available at `http://merch-store.local`**

## Load Testing

![Load Test Results](https://i.imgur.com/gui553G.png)

## Make Commands

| Command | Description |
|---------|-------------|
| `make build-app-up` | Build and start all services |
| `make app-up` | Start all services (without rebuild) |
| `make infra-up` | Start databases only |
| `make test` | Run tests with race detection and coverage |
