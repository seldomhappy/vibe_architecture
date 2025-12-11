# Vibe Architecture

Production-ready Clean Architecture microservice in Go with full observability stack.

## 🏗️ Architecture

```
Clean Architecture Layers:
┌─────────────────────────────────────────┐
│         Delivery Layer (HTTP)           │
│  - REST API handlers                    │
│  - Middleware (tracing, metrics, logs)  │
│  - Request/Response DTOs                │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│         Use Case Layer                  │
│  - Business logic                       │
│  - Orchestration                        │
│  - Transaction management               │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│         Repository Layer                │
│  - Data access                          │
│  - Query building                       │
│  - Transaction support                  │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│      Infrastructure Layer               │
│  - PostgreSQL (pgx)                     │
│  - Kafka Producer/Consumer              │
│  - OpenTelemetry                        │
│  - Prometheus                           │
└─────────────────────────────────────────┘
```

## 🚀 Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Native `net/http`
- **Database**: PostgreSQL 15 with [pgx](https://github.com/jackc/pgx)
- **Message Queue**: Apache Kafka with [sarama](https://github.com/IBM/sarama)
- **Tracing**: OpenTelemetry + Jaeger
- **Metrics**: Prometheus + Grafana
- **Configuration**: [cleanenv](https://github.com/ilyakaznacheev/cleanenv)
- **Migrations**: [tern](https://github.com/jackc/tern)
- **Container**: Docker + Docker Compose

## 📦 Features

✅ **Clean Architecture** - Separation of concerns with clear dependency rules  
✅ **Full Observability** - Logs, Metrics, and Distributed Tracing  
✅ **Event-Driven** - Kafka integration for domain events  
✅ **High Performance** - pgx connection pooling  
✅ **Structured Config** - YAML + Environment variables  
✅ **Graceful Shutdown** - Proper lifecycle management  
✅ **Production Ready** - Health checks, error handling, timeouts  

## 🎯 Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make (optional, but recommended)

### 1. Clone the repository

```bash
git clone https://github.com/seldomhappy/vibe_architecture.git
cd vibe_architecture
```

### 2. Copy environment file

```bash
cp .env.example .env
```

### 3. Start infrastructure

```bash
make docker-up
```

This starts:
- PostgreSQL on `:5432`
- Kafka on `:9092`
- Kafka UI on `:8090`
- Jaeger on `:16686`
- Prometheus on `:9091`
- Grafana on `:3000`

### 4. Run migrations

```bash
make migrate
```

### 5. Start the application

```bash
make run
```

The API will be available at `http://localhost:8080`

## 📚 API Documentation

### Health Check

```bash
curl http://localhost:8080/health
```

### Create Task

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Implement feature X",
    "description": "Add new authentication feature",
    "priority": "high",
    "created_by": 1
  }'
```

### Get Task

```bash
curl http://localhost:8080/tasks/1
```

### List Tasks

```bash
# All tasks
curl http://localhost:8080/tasks

# Filter by status
curl "http://localhost:8080/tasks?status=pending"

# Filter by priority
curl "http://localhost:8080/tasks?priority=high"

# Pagination
curl "http://localhost:8080/tasks?limit=10&offset=0"
```

### Update Task

```bash
curl -X PUT http://localhost:8080/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated task name",
    "status": "in_progress",
    "priority": "medium"
  }'
```

### Assign Task

```bash
curl -X POST http://localhost:8080/tasks/1/assign \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 42
  }'
```

### Complete Task

```bash
curl -X POST http://localhost:8080/tasks/1/complete
```

### Delete Task

```bash
curl -X DELETE http://localhost:8080/tasks/1
```

## 🔍 Observability

### Logs

The application uses structured logging with request ID and trace ID:

```
[vibe-architecture] [INFO] [req-123][trace:abc...def] Creating task: Implement feature X
```

### Metrics (Prometheus)

View metrics at: `http://localhost:9090/metrics`

Available metrics:
- **HTTP**: `http_requests_total`, `http_request_duration_seconds`, `http_requests_in_flight`
- **Business**: `tasks_created_total`, `tasks_completed_total`, `tasks_by_status`
- **Database**: `db_connections_open`, `db_query_duration_seconds`
- **System**: `app_info`, `app_uptime_seconds`

Prometheus UI: `http://localhost:9091`

### Tracing (Jaeger)

View distributed traces at: `http://localhost:16686`

Every request creates a trace with spans across:
- HTTP handler
- Use case
- Repository
- Database queries

### Kafka Events

Monitor Kafka topics with Kafka UI: `http://localhost:8090`

Events published:
- `task.created` - When a task is created
- `task.updated` - When a task is updated
- `task.completed` - When a task is completed
- `task.deleted` - When a task is deleted

### Grafana Dashboards

Access Grafana at: `http://localhost:3000`
- Username: `admin`
- Password: `admin`

## ⚙️ Configuration

Configuration is managed via YAML files and environment variables.

### Configuration Files

- `config/config.yaml` - Development configuration
- `config/config.production.yaml` - Production configuration
- `.env` - Environment-specific overrides

### Environment Variables

Key environment variables (see `.env.example` for full list):

```bash
APP_ENVIRONMENT=development
APP_DEBUG=true
SERVER_PORT=8080
DB_HOST=localhost
KAFKA_BROKERS=localhost:9092
TRACING_ENABLED=true
METRICS_ENABLED=true
```

## 🛠️ Development

### Available Make Commands

```bash
make help          # Show all available commands
make run           # Run the application
make build         # Build binary
make test          # Run tests
make lint          # Run linter
make docker-up     # Start infrastructure
make docker-down   # Stop infrastructure
make migrate       # Run migrations
make clean         # Clean build artifacts
make dev           # Start dev environment (docker + migrate + run)
```

### Project Structure

```
.
├── cmd/                    # Application entrypoints
│   └── main.go            # Main application
├── config/                 # Configuration files
│   ├── config.go          # Config struct
│   ├── config.yaml        # Development config
│   └── config.production.yaml
├── internal/
│   ├── domain/            # Domain layer (entities, events, errors)
│   ├── usecase/           # Use case layer (business logic)
│   ├── repository/        # Repository layer (data access)
│   ├── infrastructure/    # Infrastructure (DB, Kafka)
│   ├── delivery/          # Delivery layer (HTTP)
│   └── pkg/               # Shared packages
├── logger/                 # Logger implementation
├── docker-compose.yml     # Infrastructure setup
├── Makefile               # Build commands
└── README.md              # This file
```

### Adding New Features

1. **Define domain entities** in `internal/domain/`
2. **Create use case interface** in `internal/usecase/`
3. **Implement repository** in `internal/repository/`
4. **Add HTTP handlers** in `internal/delivery/http/`
5. **Wire dependencies** in `cmd/main.go`

## 🚀 Production Deployment

### Build for Production

```bash
go build -o bin/app -ldflags="-s -w" cmd/main.go
```

### Docker Build

```bash
docker build -t vibe-architecture:latest .
```

### Environment Variables

Ensure these are set in production:

```bash
APP_ENVIRONMENT=production
APP_DEBUG=false
DB_SSL_MODE=require
TRACING_SAMPLING_RATE=0.1
```

### Health Checks

Configure health check endpoint for orchestrators:

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 3s
  retries: 3
```

## 🐛 Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# View PostgreSQL logs
docker-compose logs postgres

# Test connection
psql -h localhost -U postgres -d vibe_architecture
```

### Kafka Connection Issues

```bash
# Check if Kafka is running
docker-compose ps kafka

# View Kafka logs
docker-compose logs kafka

# List topics
docker exec -it vibe-kafka kafka-topics --list --bootstrap-server localhost:9092
```

### Application Logs

```bash
# Enable debug logging
export LOG_LEVEL=debug
make run
```

## 📝 License

MIT License - see LICENSE file for details

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📧 Contact

For questions or feedback, please open an issue on GitHub.