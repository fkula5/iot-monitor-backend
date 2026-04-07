# IoT Monitor Backend

A scalable backend system for managing and monitoring IoT sensors, built with Go and a microservice architecture.

## Overview

This project provides a comprehensive backend infrastructure for IoT device monitoring. Microservices communicate via gRPC and expose functionality through a REST API gateway with real-time data streaming capabilities. An event-driven alert pipeline is built on top of RabbitMQ.

### Core Services

- **Authentication Service** — User management, JWT-based authentication, and password reset via email
- **Sensor Service** — CRUD for sensors, sensor types, and sensor groups
- **Data Processing Service** — Time-series storage and retrieval with TimescaleDB; publishes readings to RabbitMQ
- **Alert Service** — Evaluates sensor readings against user-defined rules and stores triggered alerts; exposes a gRPC API
- **Alert Dispatcher** — Consumes alert events from RabbitMQ and sends email notifications via SMTP
- **API Gateway** — REST + WebSocket interface for external clients; routes to all backend services and forwards real-time alerts over WebSocket
- **Data Generation Service** — Simulates IoT sensors for testing and development

### Communication & Infrastructure

- **gRPC** for all inter-service calls
- **RabbitMQ** for event-driven messaging between the data processing service, alert service, and alert dispatcher
- **TimescaleDB** (PostgreSQL extension) for time-series sensor readings
- **Ent ORM** for type-safe schema management across all relational databases
- **WebSocket** for real-time sensor data streaming to clients

### Deployment

The system runs in Docker Compose for development. Production deployment targets a Linux VPS using systemd unit files for each microservice.

---

## Architecture

```mermaid
flowchart TD
    subgraph "Client Applications"
        WebClient["Web Client"]
        MobileClient["Mobile Client"]
    end

    subgraph "API Gateway Service"
        ChiRouter["Chi Router"]
        Middleware["Middleware\n- Logger / RequestID / Timeout\n- Recoverer / CORS / JWT Auth\n- Rate Limiting"]
        SensorRoutes["Sensor Routes"]
        SensorTypeRoutes["Sensor Type Routes"]
        SensorGroupRoutes["Sensor Group Routes"]
        AuthRoutes["Auth Routes"]
        DataRoutes["Data Routes (WS + REST)"]
        AlertRoutes["Alert Routes"]
        AlertRuleRoutes["Alert Rule Routes"]
    end

    subgraph "Auth Service"
        AuthGrpc["gRPC Server"]
        AuthSvc["Auth Service\n(JWT, bcrypt, password reset)"]
        UserStore["User Storage (Ent)"]
        AuthDB[(users DB)]
    end

    subgraph "Sensor Service"
        SensorGrpc["gRPC Server"]
        SensorSvc["Sensor / Type / Group Services"]
        SensorStore["Sensor Storage (Ent)"]
        SensorDB[(sensors DB)]
    end

    subgraph "Data Processing Service"
        DataGrpc["gRPC Server"]
        DataSvc["Data Service + Stream Manager"]
        DataStore["TimescaleDB Storage"]
        DataDB[("sensor_readings\n(TimescaleDB)")]
    end

    subgraph "Alert Service"
        AlertGrpc["gRPC Server"]
        AlertSvc["Alert / AlertRule Services"]
        AlertStore["Alert Storage (Ent)"]
        AlertDB[(alerts DB)]
    end

    subgraph "Alert Dispatcher"
        Dispatcher["Email Dispatcher\n(gomail)"]
    end

    subgraph "RabbitMQ"
        ReadingsExchange["readings_exchange (fanout)"]
        AlertsExchange["alerts_exchange (fanout)"]
    end

    subgraph "Data Generation Service"
        Generator["Sensor Data Generator"]
    end

    WebClient -->|HTTP / WS| ChiRouter
    MobileClient -->|HTTP / WS| ChiRouter
    ChiRouter --> Middleware --> AuthRoutes & SensorRoutes & SensorTypeRoutes & SensorGroupRoutes & DataRoutes & AlertRoutes & AlertRuleRoutes

    AuthRoutes -->|gRPC| AuthGrpc --> AuthSvc --> UserStore --> AuthDB
    SensorRoutes & SensorTypeRoutes & SensorGroupRoutes -->|gRPC| SensorGrpc --> SensorSvc --> SensorStore --> SensorDB
    DataRoutes -->|gRPC| DataGrpc --> DataSvc --> DataStore --> DataDB

    DataSvc -->|publish| ReadingsExchange
    ReadingsExchange --> AlertSvc
    AlertSvc --> AlertStore --> AlertDB
    AlertRoutes & AlertRuleRoutes -->|gRPC| AlertGrpc --> AlertSvc

    AlertSvc -->|publish| AlertsExchange
    AlertsExchange --> Dispatcher
    Dispatcher -->|SMTP email| Dispatcher

    AlertsExchange --> DataRoutes

    Generator -->|gRPC StoreReading| DataGrpc
    Generator -->|gRPC ListSensors| SensorGrpc
```

---

## Features

### Authentication & Authorization

- Register and login with JWT tokens
- Password hashing with bcrypt (configurable cost)
- Forgot password / reset password flow via email (SMTP)
- JWT validation middleware for all protected routes
- Configurable token expiration and lockout settings

### Sensor Management

- Full CRUD for sensors with active/inactive toggling
- Sensors are scoped to a user and associated with a sensor type
- Location and description metadata

### Sensor Type Management

- Define types with model, manufacturer, unit, and value ranges (min/max)
- Full CRUD operations

### Sensor Group Management

- Organize sensors into named groups with color and icon
- Many-to-many relationship: sensors can belong to multiple groups
- Add or remove sensors from groups dynamically
- Deleting a group does not affect its sensors

### Real-Time Data Streaming

- WebSocket endpoint for real-time sensor readings
- Subscribe to specific sensor IDs or receive all active sensor data
- Real-time alert events forwarded over the same WebSocket connection via RabbitMQ fan-out

### Alert Rules & Alerts

- Create per-sensor alert rules with condition types `GT` (greater than) or `LT` (less than) and a threshold value
- Alert service evaluates every incoming reading against enabled rules
- Triggered alerts are persisted and published to RabbitMQ
- Mark alerts as read via the API
- Paginated listing of alerts and rules

### Email Alert Dispatch

- Alert Dispatcher consumes alert events from RabbitMQ
- Fetches user email from the Auth Service via gRPC
- Sends HTML alert emails via SMTP (configurable — Mailtrap-compatible by default)

### Time-Series Data Management

- Store sensor readings with timestamp precision
- Query historical data within time ranges
- Retrieve latest N readings for a single sensor
- Batch latest readings for multiple sensors
- TimescaleDB hypertables with index on `(sensor_id, time DESC)` for efficient queries

### Data Simulation

- Generates realistic sensor values with small random variations, Gaussian noise, and midpoint drift
- Configurable generation interval (`DATA_GENERATION_INTERVAL_SECONDS`)
- Automatically discovers and processes all active sensors

### API Features

- RESTful API with consistent JSON responses
- WebSocket for real-time sensor data and alert streaming
- CORS with configurable allowed origins
- Global and per-route rate limiting (configurable via environment variables)
- Request logging, timeout, and panic recovery middleware
- Swagger/OpenAPI documentation at `/swagger/index.html`

---

## Technology Stack

| Component             | Technology                         |
| --------------------- | ---------------------------------- |
| Language              | Go 1.24+                           |
| Service communication | gRPC + Protocol Buffers            |
| Message broker        | RabbitMQ (amqp091-go)              |
| Time-series DB        | TimescaleDB (PostgreSQL extension) |
| Relational DB         | PostgreSQL                         |
| ORM                   | Ent                                |
| HTTP router           | Chi                                |
| WebSocket             | Gorilla WebSocket                  |
| Authentication        | JWT (golang-jwt/jwt)               |
| Password hashing      | bcrypt                             |
| Email                 | gomail.v2                          |
| Logging               | Uber Zap                           |
| Containerization      | Docker / Docker Compose            |
| Production            | systemd on Linux VPS               |
| CI/CD                 | GitHub Actions                     |

---

## Project Structure

```
.
├── .github/workflows          # CI/CD configuration
├── .env.example               # Environment variables template
├── cmd/
│   └── seeder/                # Database seeder (users, sensor types, sensors, alert rules, alerts)
│       └── data/              # Seed JSON files
├── internal/
│   ├── auth/                  # JWT service and password service (shared)
│   ├── database/              # Ent client initialisation helpers
│   ├── proto/                 # Generated protobuf Go code
│   │   ├── auth/
│   │   ├── sensor_service/
│   │   ├── data_service/
│   │   └── alert_service/
│   └── types/                 # Shared HTTP request/response types
├── pkg/
│   └── logger/                # Zap-based structured logger
├── proto/                     # Protobuf definition files
│   ├── auth.proto
│   ├── sensor_service.proto
│   ├── data_service.proto
│   └── alert_service.proto
├── scripts/
│   └── verify_security.sh     # CORS and rate limit verification script
├── services/
│   ├── api-gateway/           # REST + WebSocket gateway
│   │   ├── handlers/          # HTTP handlers (auth, sensor, sensortype, sensorgroup, websocket, alert, alertrule)
│   │   ├── middleware/        # JWT auth middleware
│   │   ├── routes/            # Route setup per domain
│   │   └── docs/              # Swagger docs (generated)
│   ├── auth/                  # Authentication gRPC service
│   │   ├── ent/schema/        # User entity schema
│   │   ├── handlers/          # gRPC handler
│   │   ├── services/          # Auth business logic + mailer
│   │   └── storage/           # User storage
│   ├── alert-service/         # Alert gRPC service + RabbitMQ consumer
│   │   ├── ent/schema/        # Alert and AlertRule entity schemas
│   │   ├── handlers/          # gRPC handler
│   │   ├── service/           # Alert and AlertRule services
│   │   └── storage/           # Alert and AlertRule storage
│   ├── alert-dispatcher/      # RabbitMQ consumer → SMTP email dispatcher
│   ├── data-generation-service/ # Sensor data simulator
│   │   └── services/          # Generator service
│   ├── data-processing/       # Data gRPC service + TimescaleDB + RabbitMQ publisher
│   │   ├── handlers/          # gRPC handler with stream management
│   │   └── storage/           # TimescaleDB and in-memory storage implementations
│   └── sensor-service/        # Sensor gRPC service
│       ├── ent/schema/        # Sensor, SensorType, SensorGroup schemas
│       ├── handlers/          # gRPC handler
│       ├── services/          # Business logic
│       └── storage/           # Storage implementations
├── nginx/                     # Nginx reverse proxy config template (SSL)
├── init-db.sh                 # PostgreSQL initialisation script (databases, users, TimescaleDB)
├── docker-compose.yml
├── Dockerfile                 # Multi-stage build (reusable across services via SERVICE_PATH arg)
├── Makefile
├── .golangci.yml              # Linter configuration
└── go.mod / go.sum
```

---

## Getting Started

### Prerequisites

- Go 1.24+
- Docker and Docker Compose
- `protoc` with `protoc-gen-go` and `protoc-gen-go-grpc` plugins (for proto regeneration)

### Environment Setup

```bash
cp .env.example .env
```

Fill in `.env`. Minimum required values:

```env
# API Gateway
API_GATEWAY_PORT=8080

# Auth Service
AUTH_SERVICE_GRPC_ADDR=localhost:50051
AUTH_SERVICE_GRPC_PORT=50051
AUTH_SERVICE_DB_NAME=iot_auth
AUTH_SERVICE_DB_USER=auth_user
AUTH_SERVICE_DB_PASSWORD=your-password
JWT_SECRET=your-secret-key
JWT_EXPIRATION_HOURS=24
BCRYPT_COST=12
MAX_LOGIN_ATTEMPTS=5
LOCKOUT_DURATION_MINUTES=15

# Sensor Service
SENSOR_SERVICE_GRPC_ADDR=localhost:50052
SENSOR_SERVICE_GRPC_PORT=50052
SENSOR_SERVICE_DB_NAME=iot_sensors
SENSOR_SERVICE_DB_USER=sensor_user
SENSOR_SERVICE_DB_PASSWORD=your-password

# Data Processing Service
DATA_SERVICE_GRPC_ADDR=localhost:50053
DATA_SERVICE_GRPC_PORT=50053
DATA_SERVICE_DB_NAME=iot_data
DATA_SERVICE_DB_USER=data_user
DATA_SERVICE_DB_PASSWORD=your-password

# Alert Service
ALERT_SERVICE_GRPC_ADDR=localhost:50054
ALERT_SERVICE_GRPC_PORT=50054
ALERT_SERVICE_DB_NAME=iot_alerts
ALERT_SERVICE_DB_USER=alert_user
ALERT_SERVICE_DB_PASSWORD=your-password

# Database
DB_HOST=localhost
DB_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your-password

# RabbitMQ
RABBITMQ_URL=amqp://user:pass@localhost:5672/
RABBITMQ_DEFAULT_USER=user
RABBITMQ_DEFAULT_PASS=pass

# SMTP (email alerts)
SMTP_HOST=smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USER=
SMTP_PASS=
SMTP_FROM=alerts@iot-monitor.local

# Data generation
DATA_GENERATION_INTERVAL_SECONDS=60

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:5173

# Rate limiting
RATE_LIMIT_GLOBAL_REQUESTS=100
RATE_LIMIT_GLOBAL_WINDOW=1m
RATE_LIMIT_AUTH_REQUESTS=5
RATE_LIMIT_AUTH_WINDOW=1m
```

### Running with Docker Compose

```bash
make up
# or
docker compose up --build auth-service sensor-service api-gateway data-generation-service data-processing-service
```

This also starts RabbitMQ, TimescaleDB/PostgreSQL, the alert service, and the alert dispatcher.

### Running the Database Seeder

After services are up and databases are initialised:

```bash
make seed
```

This seeds users, sensor types, sensors, alert rules, and sample alerts from `cmd/seeder/data/`.

### Building Individual Services

```bash
make build SERVICE=auth-service
make build SERVICE=sensor-service
make build SERVICE=data-processing
make build SERVICE=api-gateway
make build SERVICE=data-generation-service
make build SERVICE=alert-service
make build SERVICE=alert-dispatcher
```

Binaries are written to `bin/`.

### Regenerating Protobuf Code

```bash
make generate-proto
```

### Regenerating Ent Schemas

```bash
go generate ./services/auth/ent
go generate ./services/sensor-service/ent
go generate ./services/alert-service/ent
```

---

## API Reference

All endpoints are prefixed as shown. Protected routes require `Authorization: Bearer <token>`.

### Auth — `/auth`

| Method | Path                    | Description                  |
| ------ | ----------------------- | ---------------------------- |
| POST   | `/auth/register`        | Register a new user          |
| POST   | `/auth/login`           | Login, receive JWT           |
| GET    | `/auth/user`            | Get current user profile     |
| PUT    | `/auth/user`            | Update current user profile  |
| POST   | `/auth/forgot-password` | Request password reset email |
| POST   | `/auth/reset-password`  | Reset password with token    |

### Sensors — `/api/sensors` 🔒

| Method | Path                       | Description                         |
| ------ | -------------------------- | ----------------------------------- |
| GET    | `/api/sensors`             | List sensors for authenticated user |
| POST   | `/api/sensors`             | Create sensor                       |
| GET    | `/api/sensors/{id}`        | Get sensor by ID                    |
| PUT    | `/api/sensors/{id}`        | Update sensor                       |
| DELETE | `/api/sensors/{id}`        | Delete sensor                       |
| PUT    | `/api/sensors/{id}/active` | Toggle sensor active status         |

### Sensor Types — `/api/sensor-types` 🔒

| Method | Path                     | Description           |
| ------ | ------------------------ | --------------------- |
| GET    | `/api/sensor-types`      | List all sensor types |
| POST   | `/api/sensor-types`      | Create sensor type    |
| GET    | `/api/sensor-types/{id}` | Get sensor type       |
| PUT    | `/api/sensor-types/{id}` | Update sensor type    |
| DELETE | `/api/sensor-types/{id}` | Delete sensor type    |

### Sensor Groups — `/api/sensor-groups` 🔒

| Method | Path                              | Description                        |
| ------ | --------------------------------- | ---------------------------------- |
| GET    | `/api/sensor-groups`              | List groups for authenticated user |
| POST   | `/api/sensor-groups`              | Create group                       |
| GET    | `/api/sensor-groups/{id}`         | Get group with sensors             |
| PUT    | `/api/sensor-groups/{id}`         | Update group                       |
| DELETE | `/api/sensor-groups/{id}`         | Delete group                       |
| POST   | `/api/sensor-groups/{id}/sensors` | Add sensors to group               |
| DELETE | `/api/sensor-groups/{id}/sensors` | Remove sensors from group          |

### Data — `/api/data`

| Method | Path                                                             | Description                            |
| ------ | ---------------------------------------------------------------- | -------------------------------------- |
| GET    | `/api/data/ws/readings?sensor_ids=1,2,3`                         | WebSocket: real-time readings + alerts |
| GET    | `/api/data/readings/latest?sensor_ids=1,2,3`                     | Latest reading per sensor (batch)      |
| GET    | `/api/data/sensors/{sensor_id}/latest?limit=10`                  | Latest N readings for one sensor       |
| GET    | `/api/data/sensors/{sensor_id}/readings?start_time=…&end_time=…` | Historical readings                    |
| POST   | `/api/data/readings`                                             | Store a reading manually               |

### Alerts — `/api/alerts` 🔒

| Method | Path                          | Description                        |
| ------ | ----------------------------- | ---------------------------------- |
| GET    | `/api/alerts?page=1&limit=10` | List alerts for authenticated user |
| POST   | `/api/alerts/{id}/read`       | Mark alert as read                 |

### Alert Rules — `/api/alert-rules` 🔒

| Method | Path                               | Description                             |
| ------ | ---------------------------------- | --------------------------------------- |
| GET    | `/api/alert-rules?page=1&limit=10` | List alert rules for authenticated user |
| POST   | `/api/alert-rules`                 | Create alert rule                       |
| GET    | `/api/alert-rules/{id}`            | Get alert rule                          |
| PUT    | `/api/alert-rules/{id}`            | Update alert rule                       |
| DELETE | `/api/alert-rules/{id}`            | Delete alert rule                       |

### Other

| Method | Path                  | Description  |
| ------ | --------------------- | ------------ |
| GET    | `/health`             | Health check |
| GET    | `/swagger/index.html` | Swagger UI   |

---

## Alert Rule Request Format

```json
{
  "name": "High Temperature Alert",
  "sensor_id": 1,
  "condition_type": "GT",
  "threshold": 30.0,
  "description": "Alert when temperature exceeds 30°C"
}
```

`condition_type` is `"GT"` (greater than) or `"LT"` (less than).

---

## Event-Driven Alert Flow

```
Sensor reading stored
  → Data Processing Service publishes to readings_exchange (RabbitMQ fanout)
    → Alert Service consumes from alert_engine_queue
      → Evaluates enabled rules for that sensor_id
        → On match: saves Alert to DB, publishes to alerts_exchange
          → Alert Dispatcher consumes, fetches user email via Auth gRPC, sends SMTP email
          → API Gateway consumes, forwards alert payload over active WebSocket connections
```

---

## Architecture Decisions

**gRPC for internal communication** — type safety and performance between services.

**RabbitMQ fanout exchanges** — `readings_exchange` fans out to both the alert engine queue and any future consumers. `alerts_exchange` fans out to the dispatcher and the gateway simultaneously without either blocking the other.

**Per-service databases** — each service owns its schema and database credentials for isolation.

**TimescaleDB hypertables** — automatic time-based partitioning and a `(sensor_id, time DESC)` index make time-range and latest-reading queries efficient at scale.

**Ent ORM** — compile-time schema validation and type-safe queries across auth, sensor, and alert databases.

**Many-to-many sensor groups** — junction table managed by Ent edges; deleting a group is non-destructive to sensors.

---

## Linting

```bash
golangci-lint run
```

Configuration is in `.golangci.yml`. Auto-fix is enabled where possible. Generated files (`*.pb.go`, `ent/`) are excluded.

---

## License

[MIT](LICENSE)
