# 🪑 HopSpot API

A RESTful backend API for the HopSpot bench-sharing mobile app. Discover, rate, and track your visits to park benches with your friends.

[![CI](https://github.com/CKnuchel/HopSpotAPI/actions/workflows/ci.yml/badge.svg)](https://github.com/CKnuchel/HopSpotAPI/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/CKnuchel/HopSpotAPI/branch/main/graph/badge.svg)](https://codecov.io/gh/CKnuchel/HopSpotAPI)
[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)

## 📖 Table of Contents

- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Architecture](#-architecture)
- [Getting Started](#-getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Configuration](#configuration)
- [API Documentation](#-api-documentation)
- [External Services Setup](#-external-services-setup)
  - [Firebase Cloud Messaging](#firebase-cloud-messaging-fcm)
  - [MinIO Object Storage](#minio-object-storage)
- [Development](#-development)
- [Deployment](#-deployment)
- [License](#-license)

## ✨ Features

- **Bench Management** - Create, update, delete, and browse park benches with GPS coordinates
- **Photo Upload** - Upload up to 10 photos per bench with automatic resizing (original, medium, thumbnail)
- **Visit Tracking** - Record and track your bench visits
- **Proximity Search** - Find benches within a specified radius using the Haversine formula
- **Weather Integration** - Get current weather data for any bench location
- **Push Notifications** - Receive notifications when friends add new benches
- **Invitation-based Registration** - Secure registration with invitation codes
- **Role-based Access Control** - User and Admin roles with different permissions

## 🛠 Tech Stack

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.25 |
| **Framework** | [Gin](https://gin-gonic.com/) |
| **Database** | PostgreSQL 16 |
| **ORM** | [GORM](https://gorm.io/) |
| **Object Storage** | [MinIO](https://min.io/) (S3-compatible) |
| **Push Notifications** | Firebase Cloud Messaging (FCM) |
| **Weather Data** | [Open-Meteo API](https://open-meteo.com/) (free, no API key required) |
| **Authentication** | JWT (JSON Web Tokens) |
| **API Documentation** | Swagger / OpenAPI |
| **Containerization** | Docker & Docker Compose |

## 🏗 Architecture

```
├── cmd/
│   └── server/           # Application entrypoint
├── internal/
│   ├── config/           # Configuration management
│   ├── database/         # Database connection & migrations
│   ├── domain/           # Domain models (entities)
│   ├── dto/              # Data Transfer Objects
│   │   ├── requests/     # Request DTOs
│   │   └── responses/    # Response DTOs
│   ├── handler/          # HTTP handlers (controllers)
│   ├── mapper/           # Domain <-> DTO mappers
│   ├── middleware/       # HTTP middleware (auth, logging)
│   ├── repository/       # Data access layer
│   ├── router/           # Route definitions
│   └── service/          # Business logic layer
├── pkg/
│   ├── apperror/         # Custom application errors
│   ├── notification/     # FCM client
│   ├── storage/          # MinIO client
│   ├── utils/            # Utility functions
│   └── weather/          # Weather API client
├── docs/                 # Swagger documentation
├── docker-compose.yml    # Docker Compose configuration
├── Dockerfile            # Multi-stage Docker build
└── .env.example          # Environment variables template
```

## 🚀 Getting Started

### Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Docker](https://www.docker.com/get-started) & Docker Compose
- [Firebase Project](https://console.firebase.google.com/) (for push notifications)

### Installation

1. **Clone the repository**

```bash
git clone https://github.com/CKnuchel/HopSpotAPI.git
cd HopSpotAPI
```

2. **Copy environment template**

```bash
cp .env.example .env
```

3. **Configure environment variables** (see [Configuration](#configuration))

4. **Start with Docker Compose**

```bash
docker compose up -d
```

5. **Verify the API is running**

```bash
curl http://localhost:8080/health
# Response: {"status":"healthy"}
```

6. **Get your first invitation code**

On first startup (empty database), an initial invitation code is generated and logged:

```bash
docker compose logs api | grep "invitation code"
# Look for: INF Initial invitation code generated code=XXXXXX
```

Use this code to register your first (admin) user.

### Configuration

Create a `.env` file with the following variables:

```bash
# Server
PORT=8080
LOG_LEVEL=INFO                    # INFO, DEBUG

# Database
DB_HOST=postgres                  # Use 'localhost' for local development
DB_PORT=5432
DB_USER=bench_user
DB_PASSWORD=your_secure_password
DB_NAME=bench_db

# JWT Authentication
JWT_SECRET=your_very_long_random_secret_min_32_chars
JWT_EXPIRE_SECONDS=3600           # Token validity in seconds (1 hour)
JWT_ISSUER=hopspot
JWT_AUDIENCE=hopspot_users

# MinIO Object Storage
MINIO_ENDPOINT=minio:9000         # Use 'localhost:9000' for local development
MINIO_ACCESS_KEY=minio_admin
MINIO_SECRET_KEY=your_minio_password
MINIO_USE_SSL=false
MINIO_BUCKET_NAME=hopspot-photos

# Firebase Cloud Messaging
FIREBASE_AUTH_KEY=base64_encoded_service_account_json
```

## 📚 API Documentation

### Swagger UI

When the server is running, access the interactive API documentation at:

```
http://localhost:8080/swagger/index.html
```

### API Endpoints Overview

#### Authentication (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/register` | Register with invitation code |
| `POST` | `/api/v1/auth/login` | Login and receive JWT token |

#### User Management (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/users/me` | Get current user profile |
| `PATCH` | `/api/v1/users/me` | Update profile |
| `POST` | `/api/v1/users/me/change-password` | Change password |
| `POST` | `/api/v1/auth/refresh-fcm-token` | Update FCM token |

#### Benches (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/benches` | List benches (with filters) |
| `GET` | `/api/v1/benches/:id` | Get bench details |
| `POST` | `/api/v1/benches` | Create new bench |
| `PATCH` | `/api/v1/benches/:id` | Update bench |
| `DELETE` | `/api/v1/benches/:id` | Delete bench |

**Query Parameters for `GET /api/v1/benches`:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `page` | int | Page number (default: 1) |
| `limit` | int | Items per page (default: 50, max: 100) |
| `search` | string | Search by name/description |
| `has_toilet` | bool | Filter by toilet availability |
| `has_trash_bin` | bool | Filter by trash bin availability |
| `min_rating` | int | Minimum rating (1-5) |
| `lat` | float | Latitude for distance calculation |
| `lon` | float | Longitude for distance calculation |
| `radius` | int | Search radius in meters |
| `sort_by` | string | Sort field: `name`, `rating`, `created_at`, `distance` |
| `sort_order` | string | Sort direction: `asc`, `desc` |

#### Photos (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/benches/:id/photos` | Upload photo |
| `GET` | `/api/v1/benches/:id/photos` | List bench photos |
| `DELETE` | `/api/v1/photos/:id` | Delete photo |
| `PATCH` | `/api/v1/photos/:id/main` | Set as main photo |
| `GET` | `/api/v1/photos/:id/url` | Get presigned URL |

#### Visits (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/visits` | List own visits |
| `POST` | `/api/v1/visits` | Record a visit |
| `GET` | `/api/v1/benches/:id/visits/count` | Get visit count |

#### Weather (Protected)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/weather?lat=47.37&lon=8.54` | Get current weather |

#### Admin (Admin Role Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/admin/users` | List all users |
| `PATCH` | `/api/v1/admin/users/:id` | Update user (role, active) |
| `DELETE` | `/api/v1/admin/users/:id` | Delete user |
| `GET` | `/api/v1/admin/invitation-codes` | List invitation codes |
| `POST` | `/api/v1/admin/invitation-codes` | Create invitation code |

### Authentication

All protected endpoints require a JWT token in the Authorization header:

```bash
Authorization: Bearer <your_jwt_token>
```

**Example Login Request:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "your_password"}'
```

**Example Response:**

```json
{
  "user": {
    "id": "1",
    "email": "user@example.com",
    "display_name": "John",
    "role": "user",
    "created_at": "2026-01-28T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

## 🔧 External Services Setup

### Firebase Cloud Messaging (FCM)

FCM is used to send push notifications when new benches are added.

#### 1. Create a Firebase Project

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Click **Add project**
3. Enter a project name (e.g., "HopSpot")
4. Disable Google Analytics (optional)
5. Click **Create project**

#### 2. Generate Service Account Key

1. In Firebase Console, go to **Project Settings** (gear icon)
2. Select the **Service accounts** tab
3. Click **Generate new private key**
4. Save the downloaded JSON file securely

#### 3. Encode and Configure

The API expects the service account JSON as a Base64-encoded string:

**Linux/macOS:**

```bash
# Encode the JSON file to Base64
base64 -i firebase-service-account.json | tr -d '\n'

# Or save to a file
base64 -i firebase-service-account.json | tr -d '\n' > firebase-key.txt
```

**Windows (PowerShell):**

```powershell
[Convert]::ToBase64String([IO.File]::ReadAllBytes("firebase-service-account.json"))
```

**Set the environment variable:**

```bash
FIREBASE_AUTH_KEY=eyJ0eXBlIjoic2VydmljZV9hY2NvdW50Iiw...
```

> ⚠️ **Security Note:** Never commit your Firebase credentials to version control!

### MinIO Object Storage

MinIO provides S3-compatible object storage for photos. The bucket is created automatically on startup.

#### Configuration

MinIO is included in the Docker Compose setup. Default credentials:

```bash
MINIO_ROOT_USER=minio_admin
MINIO_ROOT_PASSWORD=your_secure_password
```

#### Access MinIO Console

When running with Docker Compose:

```
http://localhost:9001
```

#### Photo Storage Structure

Photos are stored with automatic resizing:

```
hopspot-photos/
└── benches/
    └── {bench_id}/
        └── photos/
            ├── {photo_id}_original.jpg   # Max 1920x1080
            ├── {photo_id}_medium.jpg     # Max 800x600
            └── {photo_id}_thumbnail.jpg  # 200x200 (cropped)
```

## 💻 Development

### Running Locally (without Docker)

1. **Start PostgreSQL and MinIO** (via Docker):

```bash
docker compose up -d postgres minio
```

2. **Update `.env`** for local development:

```bash
DB_HOST=localhost
MINIO_ENDPOINT=localhost:9000
```

3. **Run the API**:

```bash
go run ./cmd/server
```

### Generate Swagger Documentation

```bash
# Install swag CLI (use version matching go.mod)
go install github.com/swaggo/swag/cmd/swag@v1.8.12

# Generate docs
swag init -g cmd/server/main.go -o docs
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report in browser
go tool cover -html=coverage.out

# Run specific service tests
go test -v ./internal/service/... -run TestAuthService
```

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Run with auto-fix
golangci-lint run --fix
```

### Project Commands

```bash
# Build binary
go build -o hopspot-api ./cmd/server

# Format code
go fmt ./...

# Download dependencies
go mod download

# Tidy dependencies
go mod tidy
```

## 🚢 Deployment

### Docker Compose (Recommended)

The included `docker-compose.yml` sets up:

- **API** - The HopSpot API server
- **PostgreSQL** - Database
- **MinIO** - Object storage for photos

```bash
# Start all services
docker compose up -d

# View logs
docker compose logs -f api

# Stop all services
docker compose down
```

### Production Considerations

1. **Use strong passwords** for all services
2. **Enable HTTPS** via reverse proxy (Nginx, Caddy, Cloudflare Tunnel)
3. **Configure backups** for PostgreSQL and MinIO data
4. **Set up monitoring** (Prometheus, Grafana)
5. **Use Docker secrets** or a secrets manager for sensitive values

### Health Check

```bash
curl http://localhost:8080/health
```

## 📝 Roadmap

- [x] Refresh token implementation
- [x] Redis caching for weather data
- [x] Rate limiting
- [x] Structured logging
- [x] Configuration Validation
- [x] Unit and integration tests
- [ ] E2E tests with test database

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Made with ❤️ for bench enthusiasts**