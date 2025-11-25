# IoT Monitor Backend

A scalable backend system for managing and monitoring IoT sensors, built with Go and modern microservice architecture.

## Overview

This project provides a comprehensive backend infrastructure for IoT device monitoring. It consists of several microservices that communicate via gRPC and expose functionality through a REST API gateway with real-time data streaming capabilities.

### Core Services

- **Authentication Service**: User management and JWT-based authentication
- **Sensor Service**: Core service for managing sensors and sensor types
- **Data Processing Service**: Time-series data storage and retrieval using TimescaleDB
- **API Gateway**: REST API interface for external clients with JWT authentication middleware and WebSocket support
- **Data Generation Service**: Simulates IoT sensors for testing and development

The system uses gRPC for inter-service communication, TimescaleDB (PostgreSQL extension) for time-series data persistence, WebSocket for real-time data streaming, and the Ent ORM for database schema management.

### Deployment

The system is currently deployed on a VPS using systemd for service management on Linux. Each microservice runs as a separate systemd unit, providing automatic restart on failure, centralized logging, and easy service management.

## Architecture

```mermaid
flowchart TD
    %% Client Applications
    subgraph "Client Applications"
        WebClient["Web Client"]
        MobileClient["Mobile Client"]
    end

    %% API Gateway Service
    subgraph "API Gateway Service"
        ChiRouter["Chi Router"]
        ApiRoutes["API Routes"]
        Middleware["Middleware </br>- Logger</br>- RequestID</br>- Timeout</br>- Recoverer</br>- CORS</br>- JWT Auth"]
        SensorRoutes["Sensor Routes</br>- List Sensors</br>- Get Sensor</br>- Create Sensor</br>- Update Sensor</br>- Delete Sensor</br>- Set Sensor Active"]
        SensorTypeRoutes["Sensor Type Routes</br>- List Sensor Types</br>- Get Sensor Type</br>- Create Sensor Type</br>- Update Sensor Type</br>- Delete Sensor Type"]
        AuthRoutes["Auth Routes</br>- Register</br>- Login"]
        DataRoutes["Data Routes</br>- WebSocket Stream</br>- Latest Readings</br>- Historical Data"]
    end

    %% Auth Service
    subgraph "Auth Service"
        AuthGrpcServer["gRPC Server"]
        AuthServiceImpl["Auth Service Implementation"]

        subgraph "Auth Service Layer"
            AuthService["Auth Service"]
            JWTService["JWT Service"]
            PasswordService["Password Service"]
        end

        subgraph "Auth Storage Layer"
            UserStorage["User Storage"]
            AuthEntClient["Ent ORM Client"]
        end
    end

    %% Sensor Service
    subgraph "Sensor Service"
        GrpcServer["gRPC Server"]
        SensorServiceImpl["Sensor Service Implementation"]
        SensorTypeServiceImpl["Sensor Type Service Implementation"]

        subgraph "Service Layer"
            SensorService["Sensor Service"]
            SensorTypeService["Sensor Type Service"]
        end

        subgraph "Storage Layer"
            SensorStorage["Sensor Storage"]
            SensorTypeStorage["Sensor Type Storage"]
            EntClient["Ent ORM Client"]
        end
    end

    %% Data Processing Service
    subgraph "Data Processing Service"
        DataGrpcServer["gRPC Server"]
        DataServiceImpl["Data Service Implementation"]

        subgraph "Data Service Layer"
            DataService["Data Service"]
            StreamManager["Stream Manager"]
        end

        subgraph "Data Storage Layer"
            TimescaleStorage["TimescaleDB Storage"]
        end
    end

    %% Database Server
    subgraph "PostgreSQL Server"
        AuthPostgresDB[(users DB)]
        SensorsDB[(sensors DB)]
        DataDB[("sensor_readings DB</br>(TimescaleDB)")]
    end

    %% Data Generation Service
    subgraph "Data Generation Service"
        Generator["Data Generator"]
        ValueGenerator["Value Generator Logic"]
    end

    %% Connections between components
    %% Client to API Gateway
    WebClient --> ChiRouter
    MobileClient --> ChiRouter
    WebClient -.WebSocket.-> DataRoutes
    MobileClient -.WebSocket.-> DataRoutes

    %% API Gateway internal connections
    ChiRouter --> Middleware
    Middleware --> ApiRoutes
    ApiRoutes --> SensorRoutes
    ApiRoutes --> SensorTypeRoutes
    ApiRoutes --> AuthRoutes
    ApiRoutes --> DataRoutes

    %% API Gateway to Services
    AuthRoutes -- "gRPC Client" --> AuthGrpcServer
    SensorRoutes -- "gRPC Client" --> GrpcServer
    SensorTypeRoutes -- "gRPC Client" --> GrpcServer
    DataRoutes -- "gRPC Client" --> DataGrpcServer

    %% Auth Service internal connections
    AuthGrpcServer --> AuthServiceImpl
    AuthServiceImpl --> AuthService
    AuthService --> JWTService
    AuthService --> PasswordService
    AuthService --> UserStorage
    UserStorage --> AuthEntClient
    AuthEntClient --> AuthPostgresDB

    %% Sensor Service internal connections
    GrpcServer --> SensorServiceImpl
    GrpcServer --> SensorTypeServiceImpl
    SensorServiceImpl --> SensorService
    SensorTypeServiceImpl --> SensorTypeService
    SensorService --> SensorStorage
    SensorTypeService --> SensorTypeStorage
    SensorStorage --> EntClient
    SensorTypeStorage --> EntClient
    EntClient --> SensorsDB

    %% Data Processing Service internal connections
    DataGrpcServer --> DataServiceImpl
    DataServiceImpl --> DataService
    DataService --> StreamManager
    DataService --> TimescaleStorage
    StreamManager -.Stream.-> DataGrpcServer
    TimescaleStorage --> DataDB

    %% Data Generation Service connections
    Generator --> ValueGenerator
    ValueGenerator -- "gRPC Client" --> GrpcServer
    ValueGenerator -- "gRPC Client" --> DataGrpcServer

    %% Style definitions
    classDef client fill:#e1f5fe,stroke:#01579b,color:#01579b
    classDef apiGateway fill:#e8f5e9,stroke:#2e7d32,color:#2e7d32
    classDef authService fill:#f1f8e9,stroke:#33691e,color:#33691e
    classDef sensorService fill:#fff3e0,stroke:#ff6f00,color:#ff6f00
    classDef dataService fill:#fce4ec,stroke:#c2185b,color:#c2185b
    classDef dataGeneration fill:#f3e5f5,stroke:#7b1fa2,color:#7b1fa2
    classDef database fill:#ffebee,stroke:#c62828,color:#c62828

    %% Apply styles
    class WebClient,MobileClient client
    class ChiRouter,ApiRoutes,Middleware,SensorRoutes,SensorTypeRoutes,AuthRoutes,DataRoutes apiGateway
    class AuthGrpcServer,AuthServiceImpl,AuthService,JWTService,PasswordService,UserStorage,AuthEntClient authService
    class GrpcServer,SensorServiceImpl,SensorTypeServiceImpl,SensorService,SensorTypeService,SensorStorage,SensorTypeStorage,EntClient sensorService
    class DataGrpcServer,DataServiceImpl,DataService,StreamManager,TimescaleStorage dataService
    class Generator,ValueGenerator dataGeneration
    class AuthPostgresDB,SensorsDB,DataDB database
```

## Features

### Authentication & Authorization

- User registration and login with JWT tokens
- Secure password hashing with bcrypt
- JWT-based authentication middleware for protected routes
- Configurable token expiration

### Sensor Management

- Create, read, update, and delete sensors
- Enable/disable sensors for monitoring
- Associate sensors with sensor types
- Track sensor locations and descriptions

### Sensor Type Management

- Define sensor types with specific properties (model, manufacturer, unit)
- Set value ranges (min/max) for each sensor type
- Full CRUD operations on sensor types

### Real-Time Data Streaming

- WebSocket support for real-time sensor data updates
- Subscribe to multiple sensors simultaneously
- Automatic reconnection and error handling
- Supports filtering by sensor IDs

### Time-Series Data Management

- Store sensor readings with timestamp precision
- Query historical data within time ranges
- Retrieve latest readings for single or multiple sensors
- Efficient data aggregation using TimescaleDB hypertables

### Data Simulation

- Generate realistic sensor data for testing
- Configurable data generation intervals
- Value smoothing with drift and noise simulation
- Automatic detection and processing of active sensors

### API Features

- RESTful API with consistent JSON responses
- WebSocket endpoints for real-time data streaming
- CORS support for web clients
- Request logging and timeout management
- Error recovery middleware
- Swagger/OpenAPI documentation

## Technology Stack

- **Go 1.24+**: Primary programming language
- **gRPC**: Service-to-service communication with Protocol Buffers
- **TimescaleDB**: PostgreSQL extension for time-series data
- **PostgreSQL**: Relational database for data persistence
- **Ent ORM**: Type-safe database schema management and queries
- **Chi Router**: Lightweight HTTP routing framework
- **Gorilla WebSocket**: WebSocket implementation for real-time streaming
- **JWT**: Token-based authentication
- **bcrypt**: Secure password hashing
- **Docker**: Containerization platform
- **systemd**: Service management on Linux VPS
- **GitHub Actions**: CI/CD pipeline (configured in `.github/workflows`)

## Project Structure

```
.
├── .github/workflows  # CI/CD configuration
├── .env.example       # Environment variables template
├── internal           # Shared internal packages
│   ├── auth           # JWT and password utilities
│   │   ├── jwt.go             # JWT token generation and validation
│   │   └── password.go        # Password hashing and validation
│   ├── database       # Database connection utilities
│   │   └── database.go        # Ent client initialization
│   └── proto          # Generated Protocol Buffer code
│       ├── auth                # Auth service protobuf
│       ├── sensor_service      # Sensor service protobuf
│       └── data_service        # Data processing service protobuf
├── proto              # Protocol buffer definition files
│   ├── auth.proto             # Auth service definitions
│   ├── sensor_service.proto   # Sensor service definitions
│   └── data_service.proto     # Data processing service definitions
├── services           # Microservices
│   ├── api-gateway    # REST API gateway service
│   │   ├── handlers   # HTTP request handlers
│   │   │   ├── auth.go        # Authentication endpoints
│   │   │   ├── sensor.go      # Sensor endpoints
│   │   │   ├── sensortype.go  # Sensor type endpoints
│   │   │   └── websocket.go   # WebSocket and data endpoints
│   │   ├── middleware # HTTP middleware
│   │   │   └── jwt.go         # JWT authentication middleware
│   │   ├── routes     # Route definitions
│   │   │   ├── auth.go        # Auth routes
│   │   │   ├── sensor.go      # Sensor routes
│   │   │   ├── sensortype.go  # Sensor type routes
│   │   │   └── data.go        # Data streaming routes
│   │   ├── docs       # Swagger documentation
│   │   └── main.go            # Gateway entry point
│   ├── auth           # Authentication service
│   │   ├── ent        # Ent entity definitions
│   │   │   └── schema         # Database schemas
│   │   │       └── user.go    # User entity
│   │   ├── handlers   # gRPC request handlers
│   │   │   └── grpc.go        # Auth gRPC handlers
│   │   ├── services   # Authentication business logic
│   │   │   └── auth.go        # Auth service implementation
│   │   ├── storage    # User data persistence
│   │   │   └── user.go        # User storage implementation
│   │   └── main.go            # Auth service entry point
│   ├── data-processing # Data processing service
│   │   ├── handlers   # gRPC request handlers
│   │   │   └── grpc.go        # Data service gRPC handlers
│   │   ├── storage    # Data persistence
│   │   │   ├── timescale.go   # TimescaleDB storage implementation
│   │   │   └── memory.go      # In-memory storage (legacy)
│   │   └── main.go            # Data service entry point
│   ├── data-generation-service  # Sensor data simulation
│   │   ├── services   # Data generation logic
│   │   │   └── generator_service.go  # Generator implementation
│   │   └── main.go            # Generator entry point
│   └── sensor-service # Core sensor management service
│       ├── ent        # Ent entity definitions
│       │   └── schema         # Database schemas
│       │       ├── sensor.go       # Sensor entity
│       │       └── sensortype.go   # Sensor type entity
│       ├── handlers   # gRPC request handlers
│       │   └── grpc.go        # Sensor gRPC handlers
│       ├── services   # Business logic
│       │   ├── sensor.go      # Sensor service
│       │   └── sensortype.go  # Sensor type service
│       ├── storage    # Data persistence
│       │   ├── sensor.go      # Sensor storage
│       │   └── sensortype.go  # Sensor type storage
│       └── main.go            # Sensor service entry point
├── init-db.sh         # Database initialization script
├── docker-compose.yml # Docker compose configuration
├── Dockerfile         # Multi-stage Docker build
├── Makefile           # Build and development commands
├── .golangci.yml      # Linter configuration
└── README.md          # This file
```

## Getting Started

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 12 or higher with TimescaleDB extension
- Protocol Buffer Compiler (protoc)
- protoc-gen-go and protoc-gen-go-grpc plugins

### Environment Setup

1. Copy the environment template:

   ```bash
   cp .env.example .env
   ```

2. Configure your environment variables in `.env`:

   ```env
   # API Gateway
   API_GATEWAY_PORT=3000

   # Auth Service
   AUTH_SERVICE_GRPC_ADDR=localhost:50051
   AUTH_SERVICE_GRPC_PORT=50051
   AUTH_SERVICE_DB_NAME=iot_auth
   AUTH_SERVICE_DB_USER=auth_user
   AUTH_SERVICE_DB_PASSWORD=your-password
   JWT_SECRET=your-secret-key-here
   JWT_REFRESH_SECRET=your-refresh-secret
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

   # Data Generation
   DATA_GENERATION_INTERVAL_SECONDS=60

   # Database
   DB_HOST=localhost
   DB_PORT=5432
   POSTGRES_USER=postgres
   POSTGRES_PASSWORD=your-password

   # CORS
   CORS_ALLOWED_ORIGINS=http://localhost:5173,http://localhost:3000
   ```

### Database Setup

1. Create the required databases:

   ```sql
   CREATE DATABASE iot_auth;
   CREATE DATABASE iot_sensors;
   CREATE DATABASE iot_data;
   ```

2. Enable TimescaleDB extension for the data database:

   ```sql
   \c iot_data
   CREATE EXTENSION IF NOT EXISTS timescaledb;
   ```

3. The Ent ORM will automatically create the necessary tables when each service starts.
4. TimescaleDB hypertables are created by the init-db.sh script (for Docker) or manually for local setup.

### Compilation

1. Generate protocol buffer files:

   ```bash
   make generate-proto
   ```

2. Build all services:

   ```bash
   # Build auth service
   go build -o bin/auth-service ./services/auth

   # Build sensor service
   go build -o bin/sensor-service ./services/sensor-service

   # Build data processing service
   go build -o bin/data-processing-service ./services/data-processing

   # Build API gateway
   go build -o bin/api-gateway ./services/api-gateway

   # Build data generator
   go build -o bin/data-generation-service ./services/data-generation-service
   ```

### Running with Docker Compose

For development and testing, you can use Docker Compose:

```bash
# Start all services
docker-compose up --build

# Start specific services
make up
```

### Running the System

Start the services in the following order:

1. **Auth Service** (gRPC server on port 50051):

   ```bash
   ./bin/auth-service
   ```

2. **Sensor Service** (gRPC server on port 50052):

   ```bash
   ./bin/sensor-service
   ```

3. **Data Processing Service** (gRPC server on port 50053):

   ```bash
   ./bin/data-processing-service
   ```

4. **API Gateway** (HTTP server on port 8080):

   ```bash
   ./bin/api-gateway
   ```

5. **Data Generation Service** (optional):
   ```bash
   ./bin/data-generation-service
   ```

## API Endpoints

### Authentication Endpoints

**Base URL:** `http://localhost:8080/auth`

- `POST /auth/register` - Register a new user

  ```json
  {
    "email": "user@example.com",
    "username": "username",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }
  ```

- `POST /auth/login` - Login and receive JWT token
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```

### Sensor Endpoints

**Base URL:** `http://localhost:8080/api/sensors`

**Note:** All sensor endpoints require JWT authentication via `Authorization: Bearer <token>` header

- `GET /api/sensors` - List all sensors
- `GET /api/sensors/{id}` - Get sensor by ID
- `POST /api/sensors` - Create a new sensor
  ```json
  {
    "name": "Temperature Sensor 1",
    "location": "Room 101",
    "description": "Main temperature sensor",
    "sensor_type_id": 1
  }
  ```
- `PUT /api/sensors/{id}` - Update sensor
- `DELETE /api/sensors/{id}` - Delete sensor
- `PUT /api/sensors/{id}/active` - Set sensor active/inactive status
  ```json
  {
    "active": true
  }
  ```

### Sensor Type Endpoints

**Base URL:** `http://localhost:8080/api/sensor-types`

**Note:** All sensor type endpoints require JWT authentication via `Authorization: Bearer <token>` header

- `GET /api/sensor-types` - List all sensor types
- `GET /api/sensor-types/{id}` - Get sensor type by ID
- `POST /api/sensor-types` - Create a new sensor type
  ```json
  {
    "name": "Temperature Sensor",
    "model": "DHT22",
    "manufacturer": "Adafruit",
    "description": "Digital temperature and humidity sensor",
    "unit": "°C",
    "min_value": -40.0,
    "max_value": 80.0
  }
  ```
- `PUT /api/sensor-types/{id}` - Update sensor type
- `DELETE /api/sensor-types/{id}` - Delete sensor type

### Data Endpoints

**Base URL:** `http://localhost:8080/api/data`

- `GET /api/data/ws/readings?sensor_ids=1,2,3` - WebSocket endpoint for real-time sensor data streaming
- `GET /api/data/readings/latest?sensor_ids=1,2,3` - Get latest reading for multiple sensors
- `GET /api/data/sensors/{sensor_id}/latest?limit=10` - Get latest N readings for a specific sensor
- `GET /api/data/sensors/{sensor_id}/readings?start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z` - Get historical readings

### Health Check

- `GET /health` - Check if API gateway is running

### API Documentation

- `GET /swagger/index.html` - Swagger UI for interactive API documentation

## Development

### Linting

The project uses `golangci-lint` for code quality. Configuration is in `.golangci.yml`.

Run lint checks:

```bash
golangci-lint run
```

### Code Generation

When you modify `.proto` files, regenerate the Go code:

```bash
make generate-proto
```

When you modify Ent schemas, regenerate the Ent code:

```bash
# In the auth service directory
go generate ./services/auth/ent

# In the sensor service directory
go generate ./services/sensor-service/ent
```

### Adding New Sensor Types

1. Use the sensor service gRPC API or create via API gateway:

   ```json
   {
     "name": "Temperature Sensor",
     "model": "DHT22",
     "manufacturer": "Adafruit",
     "description": "Digital temperature and humidity sensor",
     "unit": "°C",
     "min_value": -40.0,
     "max_value": 80.0
   }
   ```

2. The sensor type will be available for use when creating new sensors.

### Adding New Sensors

1. Create sensors via the API gateway:

   ```json
   {
     "name": "Living Room Temp",
     "location": "Living Room",
     "description": "Monitors living room temperature",
     "sensor_type_id": 1
   }
   ```

2. Sensors are automatically created as active and will be included in data generation.

## Architecture Decisions

### Microservice Communication

- **gRPC** is used for internal service-to-service communication for performance and type safety
- **REST/HTTP** is used for external client communication via the API gateway

### Authentication Flow

1. Users register/login through the API gateway
2. Auth service validates credentials and generates JWT tokens
3. API gateway middleware validates JWT tokens for protected routes
4. User information is extracted from JWT claims for authorization

### Database Strategy

- Each service has its own database for data isolation
- TimescaleDB hypertables are used for efficient time-series data storage
- Automatic data partitioning by time for optimal query performance
- Ent ORM provides type-safe database operations
- Schema migrations are handled automatically by Ent

### Data Generation

- Simulates realistic sensor values with:
  - Small random variations (±2%)
  - Gaussian noise (1% of range)
  - Drift toward midpoint (1%)
  - Boundary enforcement

## Testing

The data generation service can be used for testing the system:

1. Create sensor types with appropriate ranges
2. Create sensors associated with those types
3. Set sensors to active
4. The data generator will automatically start producing values every minute
5. Connect to the WebSocket endpoint to receive real-time updates
6. Use the API endpoints to query historical data

## Production Deployment

### VPS Deployment with systemd

The system is deployed on a Linux VPS using systemd for service management. This provides:

- **Automatic startup**: Services start automatically on system boot
- **Auto-restart**: Failed services are automatically restarted
- **Centralized logging**: All logs are managed by journald
- **Resource management**: CPU and memory limits can be configured
- **Easy management**: Start, stop, restart, and check status with systemctl commands

## License

[MIT](LICENSE)
