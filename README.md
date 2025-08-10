# Levity - OCPP Charge Point Management System

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

Levity is a modern, scalable Open Charge Point Protocol (OCPP) management system built in Go. It provides a robust backend for managing electric vehicle charging stations, handling OCPP communications, and offering comprehensive monitoring and management capabilities.

## 🚀 Features

- **OCPP Protocol Support**: Full OCPP 1.6 and 2.0.1 compliance
- **Real-time Communication**: WebSocket-based communication with charge points
- **Database Management**: SQLite database with automatic migrations
- **RESTful API**: HTTP API for external integrations
- **Monitoring & Metrics**: Built-in monitoring and health checks
- **Plugin System**: Extensible architecture for custom functionality
- **Configuration Management**: Flexible configuration via YAML and environment variables
- **Logging**: Structured logging with multiple output formats
- **Docker Support**: Containerized deployment ready

## 🏗️ Architecture

```
levity/
├── cmd/levity/          # Main application entry point
├── core/                # Core business logic
├── server/              # HTTP and WebSocket servers
├── db/                  # Database layer and migrations
├── config/              # Configuration management
├── monitoring/          # Metrics and health checks
├── plugins/             # Plugin system
└── sql/                 # Database migrations and schemas
```

## 📋 Prerequisites

- Go 1.21 or higher
- SQLite3
- Make (for build automation)

## 🛠️ Installation

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/keeth/levity.git
   cd levity
   ```

2. **Setup development environment**
   ```bash
   make setup
   ```

3. **Create configuration**
   ```bash
   make config
   ```

4. **Build and run**
   ```bash
   make run
   ```

### Manual Setup

1. **Install dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

2. **Build the application**
   ```bash
   go build -o levity ./cmd/levity
   ```

3. **Run the application**
   ```bash
   ./levity
   ```

## 🔧 Configuration

The application can be configured through:

- **Configuration file**: `config/config.yaml`
- **Environment variables**: All config values can be overridden via environment variables
- **Command line flags**: Coming soon

### Configuration Options

| Section | Key | Default | Description |
|---------|-----|---------|-------------|
| `server` | `address` | `:8080` | HTTP server address |
| `server` | `read_timeout` | `30s` | Request read timeout |
| `server` | `write_timeout` | `30s` | Response write timeout |
| `database` | `path` | `./levity.db` | SQLite database path |
| `database` | `max_open_conns` | `25` | Maximum database connections |
| `ocpp` | `heartbeat_interval` | `60s` | OCPP heartbeat frequency |
| `log` | `level` | `info` | Logging level (debug, info, warn, error) |
| `monitoring` | `enabled` | `true` | Enable monitoring endpoints |

## 🚀 Usage

### Development

```bash
# Start development mode with hot reload
make dev

# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make format

# Lint code
make lint
```

### Production

```bash
# Build for production
make build

# Run production build
make run-build

# Run with Docker
make docker-build
make docker-run
```

### Database Management

```bash
# Run migrations
make migrate

# Reset database (WARNING: deletes all data)
make db-reset
```

## 📊 API Endpoints

### Health Check
- `GET /health` - Application health status
- `GET /ready` - Readiness probe endpoint

### OCPP Endpoints
- `POST /ocpp/chargepoint/{id}/boot` - Charge point boot notification
- `POST /ocpp/chargepoint/{id}/heartbeat` - Heartbeat endpoint
- `POST /ocpp/chargepoint/{id}/status` - Status update endpoint

### Management API
- `GET /api/v1/chargepoints` - List all charge points
- `GET /api/v1/chargepoints/{id}` - Get charge point details
- `GET /api/v1/transactions` - List transactions
- `GET /api/v1/metrics` - Application metrics

## 🔌 Plugin System

Levity supports a plugin architecture for extending functionality:

```go
package main

import "github.com/keeth/levity/plugins"

type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Initialize() error {
    // Plugin initialization logic
    return nil
}

func (p *MyPlugin) Shutdown() error {
    // Plugin cleanup logic
    return nil
}
```

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test package
go test ./core/...

# Run tests with verbose output
go test -v ./...
```

## 📦 Docker

### Build Image
```bash
make docker-build
```

### Run Container
```bash
make docker-run
```

### Custom Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o levity ./cmd/levity

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/levity .
EXPOSE 8080
CMD ["./levity"]
```

## 🔍 Monitoring

Levity includes built-in monitoring capabilities:

- **Health checks**: `/health` and `/ready` endpoints
- **Metrics**: Prometheus-compatible metrics at `/metrics`
- **Logging**: Structured JSON logging
- **Database stats**: Connection pool and query statistics

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: [Wiki](https://github.com/keeth/levity/wiki)
- **Issues**: [GitHub Issues](https://github.com/keeth/levity/issues)
- **Discussions**: [GitHub Discussions](https://github.com/keeth/levity/discussions)

## 🙏 Acknowledgments

- [OCPP Specification](https://www.openchargealliance.org/protocols/ocpp-16/)
- [Go Community](https://golang.org/community/)
- [SQLite](https://www.sqlite.org/)
- [Gin Web Framework](https://github.com/gin-gonic/gin)

---

**Built with ❤️ by the Levity Team**
