# Vibe Architecture

Clean Architecture microservice with full observability stack (OpenTelemetry, Prometheus, Kafka, pgx)

## Features

- ✅ Clean Architecture (Domain, Use Case, Repository, Delivery layers)
- ✅ PostgreSQL with pgx connection pool
- ✅ Kafka integration (optional, can be disabled)
- ✅ OpenTelemetry tracing (optional)
- ✅ Prometheus metrics
- ✅ Graceful shutdown with lifecycle management
- ✅ Production-ready Docker setup
- ✅ Health checks
- ✅ RESTful API for task management

## Architecture

```
cmd/
  main.go                    # Application entry point
internal/
  domain/
    models/                  # Domain entities
    repository/              # Repository interfaces
  infrastructure/
    config/                  # Configuration loader
    logger/                  # Logger implementation
    postgres/                # PostgreSQL connection
    kafka/                   # Kafka producer/consumer
    metrics/                 # Prometheus metrics
    tracing/                 # OpenTelemetry tracing
    lifecycle/               # Component lifecycle manager
  repository/                # Repository implementations
  usecase/                   # Business logic
  delivery/
    http/                    # HTTP handlers and server
config/
  config.yaml                # Application configuration
migrations/                  # Database migrations
```

## Quick Start

### Local Development

```bash
# Install dependencies
go mod download

# Run PostgreSQL (using Docker)
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=vibe_architecture \
  -p 5432:5432 \
  postgres:15-alpine

# Run migrations (manual for now)
psql -h localhost -U postgres -d vibe_architecture -f migrations/001_create_tasks_table.sql

# Run application
go run cmd/main.go
```

### Environment Variables

Key environment variables (see `config/config.yaml` for all options):

```bash
# Application
APP_NAME=vibe-architecture
APP_ENVIRONMENT=development
APP_DEBUG=true

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=vibe_architecture
DB_SSL_MODE=disable

# Kafka (optional)
KAFKA_ENABLED=false

# Tracing (optional)
TRACING_ENABLED=false

# Metrics
METRICS_ENABLED=true
METRICS_PORT=9090
```

## API Endpoints

### Health Check
```bash
GET /health
```

### Tasks
```bash
# Create task
POST /tasks
{
  "name": "My task",
  "description": "Task description",
  "priority": "high"
}

# List tasks
GET /tasks?limit=10&offset=0

# Get task by ID
GET /tasks/{id}

# Update task
PUT /tasks/{id}
{
  "name": "Updated name",
  "status": "completed"
}

# Delete task
DELETE /tasks/{id}
```

## 🚀 Deployment

### Render.com (рекомендуется)

Самый простой способ задеплоить приложение:

1. Форкните репозиторий
2. Зарегистрируйтесь на [Render.com](https://render.com)
3. Создайте новый Blueprint из репозитория
4. Render автоматически создаст PostgreSQL и Web Service

Подробная инструкция: [README_RENDER.md](./README_RENDER.md)

**Live Demo:** https://vibe-architecture.onrender.com

### Docker

```bash
# Build image
docker build -t vibe-arch .

# Run with PostgreSQL
docker run -d --name postgres -e POSTGRES_PASSWORD=postgres postgres:15-alpine
docker run -p 8080:8080 \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=postgres \
  -e KAFKA_ENABLED=false \
  --link postgres \
  vibe-arch
```

## Configuration

Application can be configured via:
1. `config/config.yaml` file (default values)
2. Environment variables (override config file)

Priority: Environment Variables > Config File

## Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## License

MIT