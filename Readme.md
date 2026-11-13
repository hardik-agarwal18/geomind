# GeoMind

**An AI-powered proximity and location-intelligence platform.**

GeoMind is being developed as a modular location-intelligence platform. Its backend is written in Go, using the standard library HTTP server rather than a third-party web framework.

The initial development focuses on building a dependable Go application foundation: HTTP handling, middleware, configuration, structured logging, request cancellation, and graceful shutdown.

> **Current status:** Go foundation and HTTP server implementation. Geospatial services, databases, AI orchestration, and frontend functionality are planned for subsequent development phases.

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Technology Stack](#technology-stack)
3. [Current Features](#current-features)
4. [Architecture](#architecture)
5. [Project Structure](#project-structure)
6. [HTTP API](#http-api)
7. [Middleware](#middleware)
8. [Configuration](#configuration)
9. [Logging and Observability](#logging-and-observability)
10. [Context and Cancellation](#context-and-cancellation)
11. [Server Lifecycle](#server-lifecycle)
12. [Getting Started](#getting-started)
13. [Development Commands](#development-commands)
14. [Testing](#testing)
15. [Design Principles](#design-principles)
16. [Planned Features](#planned-features)

---

## Project Overview

GeoMind aims to provide intelligent proximity search and location-based insights.

The intended platform will eventually combine:

- Geospatial queries and location intelligence.
- Real-time location processing.
- Natural-language location search.
- AI-powered recommendations.
- Demand prediction and geographic analytics.
- Interactive map-based exploration.

The current implementation establishes the backend foundation required to build these capabilities reliably.

### Current development goals

- Learn Go through a substantial software engineering project.
- Understand HTTP server internals and request lifecycles.
- Build reusable middleware without an HTTP framework.
- Implement reliable application configuration.
- Establish structured logging and request correlation.
- Handle cancellation, timeouts, and graceful shutdown.
- Maintain a testable, modular application architecture.

---

## Technology Stack

### Currently implemented

| Technology       | Purpose                                                     |
| ---------------- | ----------------------------------------------------------- |
| Go 1.22+         | Backend programming language; method-aware ServeMux routing |
| net/http         | HTTP server, router, handlers, and middleware integration   |
| context          | Request cancellation, deadlines, and lifecycle propagation  |
| log/slog         | Structured application logging                              |
| encoding/json    | JSON serialization and response handling                    |
| crypto/rand      | Cryptographically secure random request ID generation       |
| sync / WaitGroup | Goroutine lifecycle coordination patterns                   |
| httptest         | HTTP handler and server testing                             |
| Git              | Version control                                             |
| GNU Make         | Development command automation                              |

Development log formatting can use the lightweight `tint` slog handler for colored terminal output.

### Planned technology stack

| Technology            | Intended responsibility                    |
| --------------------- | ------------------------------------------ |
| PostgreSQL            | Durable application data                   |
| PostGIS               | Geospatial storage and spatial queries     |
| Redis GEO             | Fast location indexing and caching         |
| NATS JetStream        | Asynchronous event processing              |
| Next.js               | Web application                            |
| MapLibre GL JS        | Interactive maps                           |
| pgvector              | Vector similarity search where appropriate |
| Python / scikit-learn | Demand prediction baseline                 |
| Prometheus            | Metrics collection                         |
| Grafana               | Monitoring and dashboards                  |
| OpenTelemetry         | Distributed tracing and telemetry          |
| Docker                | Containerization                           |
| AWS                   | Cloud infrastructure                       |

These are planned integrations, not part of the current running application.

---

## Current Features

### 1. Go application foundation

- Go module initialization.
- Executable entry point under `cmd/api`.
- Application-private packages under `internal/`.
- Standard-library-based HTTP server.
- Separate responsibilities for application startup, server configuration, routing, and endpoint logic.

### 2. HTTP server

The application uses `http.Server` and `http.ServeMux`.

Implemented capabilities:

- HTTP routing using Go's standard library.
- Method-aware route patterns.
- Wildcard route matching and path parameter extraction.
- Explicit HTTP server configuration.
- Request and response handling.
- JSON response serialization.
- Consistent JSON error responses.
- HTTP method validation.
- HTTP 404 and 405 handling.

### 3. Health and readiness endpoints

Two separate endpoints are available:

- `/healthz` — reports application liveness.
- `/readyz` — reports initial application readiness.

The readiness endpoint currently does not claim to validate external dependencies.

### 4. Middleware pipeline

Reusable middleware is implemented using:

```go
type Middleware func(http.Handler) http.Handler
```

Currently implemented:

- Request ID generation.
- Request ID propagation through context.
- Request ID response headers.
- HTTP access logging.
- Response status capture.
- Response byte counting.
- Panic recovery.
- ResponseWriter wrapping and unwrapping.
- Middleware composition.

### 5. Configuration

Typed, environment-based configuration is implemented.

Capabilities:

- Centralized configuration loading.
- Environment variable parsing.
- Defaults for optional settings.
- Duration parsing.
- Log level validation.
- HTTP address validation.
- Maximum header size validation.
- Fail-fast behavior for invalid configuration.

### 6. Structured logging

The application uses Go's `log/slog`.

Logging capabilities:

- Log levels: DEBUG, INFO, WARN, ERROR.
- Structured attributes.
- Application service and environment metadata.
- Request ID correlation.
- HTTP method, path, status, response size, and duration.
- Panic diagnostics without exposing panic values in HTTP responses.
- JSON logging for machine-readable output.
- Optional colored development logging through `tint`.

Request and response bodies, authorization headers, and other sensitive information are not deliberately included in access logs.

### 7. Context and timeouts

Implemented concepts and infrastructure include:

- Request-scoped contexts.
- Application lifecycle context.
- Operation-level deadline patterns.
- HTTP read-header timeout.
- HTTP read timeout.
- HTTP write timeout.
- HTTP idle timeout.
- Maximum request header size.
- Request cancellation testing.

### 8. Graceful shutdown

The HTTP application supports shutdown through operating-system signals.

Implemented capabilities:

- SIGINT handling.
- SIGTERM handling on supported operating systems.
- `signal.NotifyContext`.
- Buffered server error channel.
- Graceful HTTP server shutdown.
- Configurable shutdown deadline.
- Forced connection closure when graceful shutdown fails.

Database and background worker shutdown coordination will be added when those components are implemented.

---

## Architecture

### Current request lifecycle

```text
                 HTTP Client
                      |
                      v
                 HTTP Server
                      |
                      v
                 Request ID
                      |
                      v
                 Access Log
                      |
                      v
                   Recovery
                      |
                      v
                  ServeMux
                      |
           +----------+----------+
           |                     |
           v                     v
       /healthz                /readyz
           |                     |
           v                     v
     Health Handler        Ready Handler
           |                     |
           +----------+----------+
                      |
                      v
                HTTP Response
```

### Current startup dependency graph

```text
Environment Variables
         |
         v
Configuration Loader
         |
         v
Configuration Validation
         |
         v
Structured Logger
         |
         v
HTTP Routes and Middleware
         |
         v
HTTP Server Construction
         |
         v
Application Startup
         |
         v
Signal Monitoring
         |
         v
Graceful Shutdown
```

Invalid configuration prevents the HTTP server from starting.

### Middleware execution order

The middleware chain is composed as:

```go
middleware.Chain(
    mux,
    middleware.RequestID(logger),
    middleware.AccessLog(logger),
    middleware.Recovery(logger),
)
```

Request execution order:

1. Request ID.
2. Access logging.
3. Panic recovery.
4. HTTP router.
5. Endpoint handler.

Response processing unwinds in reverse order.

---

## Project Structure

```text
geomind/
│
├── cmd/
│   └── api/
│       └── main.go
│
├── internal/
│   │
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   │
│   ├── health/
│   │   ├── handler.go
│   │   └── handler_test.go
│   │
│   ├── httpresponse/
│   │   └── response.go
│   │
│   ├── logging/
│   │   └── logger.go
│   │
│   ├── middleware/
│   │   ├── access_log.go
│   │   ├── chain.go
│   │   ├── middleware_test.go
│   │   ├── recovery.go
│   │   ├── request_id.go
│   │   └── response_writer.go
│   │
│   └── server/
│       ├── http_server.go
│       ├── http_server_test.go
│       ├── lifecycle_test.go
│       ├── server.go
│       └── server_test.go
│
├── bin/
│   └── geomind-api.exe
│
├── .env
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

The `bin/` directory contains generated build artifacts and should be excluded from version control.

Empty or not-yet-created package directories are intentionally omitted from this tree.

---

## HTTP API

The current API exposes two health endpoints.

### GET /healthz

Checks whether the application is alive.

Example:

```bash
curl -i http://localhost:8080/healthz
```

Response:

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
X-Request-ID: <request-id>
```

```json
{
  "status": "ok",
  "service": "geomind-api"
}
```

### GET /readyz

Reports the application's initial readiness.

```bash
curl -i http://localhost:8080/readyz
```

Response:

```json
{
  "status": "ready",
  "service": "geomind-api"
}
```

### Unsupported HTTP methods

```bash
curl -i -X POST http://localhost:8080/healthz
```

Expected response:

```http
HTTP/1.1 405 Method Not Allowed
Allow: GET, HEAD
Content-Type: application/json; charset=utf-8
```

```json
{
  "error": {
    "code": "method_not_allowed",
    "message": "Only GET and HEAD are supported"
  }
}
```

### Unknown endpoints

```bash
curl -i http://localhost:8080/unknown
```

Expected status:

```http
HTTP/1.1 404 Not Found
```

Response:

```json
{
  "error": {
    "code": "not_found",
    "message": "Resource not found"
  }
}
```

---

## Middleware

### Request ID middleware

Generates a UUID version 4 request identifier for each incoming request.

The ID is:

- Stored in the request context.
- Added to the `X-Request-ID` response header.
- Available to downstream handlers and middleware.

This makes it possible to correlate log entries associated with a single request.

### Access logging middleware

Records:

| Field      | Description                    |
| ---------- | ------------------------------ |
| request_id | Request correlation identifier |
| method     | HTTP method                    |
| path       | URL path                       |
| status     | HTTP response status           |
| bytes      | Response body bytes written    |
| duration   | Request processing duration    |

It does not deliberately log request or response bodies.

### Recovery middleware

Recovers from panics occurring in the handler execution chain.

If the response has not been committed, it returns:

```http
HTTP/1.1 500 Internal Server Error
```

```json
{
  "error": {
    "code": "internal_error",
    "message": "Internal server error"
  }
}
```

If the response has already been committed, it cannot replace the original status or headers.

### Response writer wrapper

The custom `statusWriter` tracks:

- Final HTTP status.
- Whether the response is committed.
- Response body byte count.

It implements `Unwrap()` to allow `http.ResponseController` to discover supported capabilities of the underlying response writer.

---

## Configuration

GeoMind reads its configuration from environment variables.

An example configuration is provided in `.env.example`.

| Variable              | Default       | Description                   |
| --------------------- | ------------- | ----------------------------- |
| `APP_ENV`             | `development` | Application environment       |
| `HTTP_ADDR`           | `:8080`       | HTTP listening address        |
| `LOG_LEVEL`           | `info`        | Minimum logging severity      |
| `SHUTDOWN_TIMEOUT`    | `10s`         | Graceful shutdown deadline    |
| `READ_HEADER_TIMEOUT` | `5s`          | Header read timeout           |
| `READ_TIMEOUT`        | `15s`         | Complete request read timeout |
| `WRITE_TIMEOUT`       | `30s`         | Response write timeout        |
| `IDLE_TIMEOUT`        | `60s`         | Idle connection timeout       |
| `MAX_HEADER_BYTES`    | `1048576`     | Maximum request header size   |

Supported environments:

- `development`
- `test`
- `production`

Supported log levels:

- `debug`
- `info`
- `warn`
- `error`

Durations use Go duration syntax, such as `5s`, `250ms`, and `2m`.

Configuration is validated at startup. Invalid values cause the application to fail instead of silently substituting defaults.

### Local environment setup

Git Bash:

```bash
cp .env.example .env

set -a
source .env
set +a
```

Go does not automatically load `.env` files. The commands above export the variables into the shell environment.

Never commit `.env` or files containing actual secrets.

---

## Logging and Observability

The application uses Go's `log/slog` API.

The logger is initialized once and shared with components that need it.

Example log fields:

```json
{
  "level": "INFO",
  "msg": "http_request",
  "service": "geomind-api",
  "environment": "development",
  "request_id": "abc123",
  "method": "GET",
  "path": "/healthz",
  "status": 200
}
```

Structured fields make application logs easier to filter and aggregate.

The logging configuration can use:

- Colored, readable text output in development through `tint`.
- JSON output for production.
- Captured or discarded output in tests.

Future observability work will introduce metrics, tracing, and dashboards.

---

## Context and Cancellation

GeoMind distinguishes between different context lifetimes.

| Context                   | Lifetime                | Intended usage                                               |
| ------------------------- | ----------------------- | ------------------------------------------------------------ |
| Request context           | Individual HTTP request | Database and network operations associated with that request |
| Application context       | Application lifetime    | Long-running workers and application-wide cancellation       |
| Operation timeout context | One bounded operation   | Database query or external API deadline                      |
| Shutdown context          | Shutdown window         | Limits graceful termination time                             |

Example of a bounded request operation:

```go
queryCtx, cancel := context.WithTimeout(
    r.Context(),
    2*time.Second,
)
defer cancel()
```

Future database and network integrations should propagate contexts instead of allowing operations to run indefinitely.

---

## Server Lifecycle

The application uses `http.Server` with explicit timeouts.

| Setting           | Current default |
| ----------------- | --------------: |
| ReadHeaderTimeout |       5 seconds |
| ReadTimeout       |      15 seconds |
| WriteTimeout      |      30 seconds |
| IdleTimeout       |      60 seconds |
| MaxHeaderBytes    |           1 MiB |

Graceful shutdown:

1. Receive SIGINT or SIGTERM.
2. Stop accepting new HTTP connections.
3. Close idle connections.
4. Allow active HTTP requests to finish.
5. Respect the configured shutdown deadline.
6. Force-close connections if graceful shutdown fails.

Background-worker coordination and dependency shutdown ordering will be added when the corresponding components are introduced.

---

## Getting Started

### Prerequisites

- Go 1.22 or later.
- Git.
- Git Bash on Windows, or a Unix-compatible shell.
- GNU Make is optional.

### Clone the repository

```bash
git clone <repository-url>
cd geomind
```

### Configure the environment

```bash
cp .env.example .env

set -a
source .env
set +a
```

### Install Go dependencies

```bash
go mod download
```

### Run GeoMind

```bash
go run ./cmd/api
```

Default address:

```text
http://localhost:8080
```

### Build the executable

For Windows:

```bash
mkdir -p bin
go build -o bin/geomind-api.exe ./cmd/api
```

For Linux or macOS:

```bash
mkdir -p bin
go build -o bin/geomind-api ./cmd/api
```

---

## Development Commands

### Format source code

```bash
gofmt -w cmd internal
```

Or:

```bash
make fmt
```

### Run tests

```bash
go test ./...
```

### Run tests without using cached results

```bash
go test -count=1 ./...
```

### Run race detector

```bash
go test -race ./...
```

### Run static analysis

```bash
go vet ./...
```

### Build API executable

```bash
go build -o bin/geomind-api.exe ./cmd/api
```

### Full Makefile check

```bash
make check
```

The Makefile provides:

| Target       | Purpose                                     |
| ------------ | ------------------------------------------- |
| `make fmt`   | Format source code                          |
| `make vet`   | Run Go static analysis                      |
| `make test`  | Run all package tests                       |
| `make run`   | Run the API                                 |
| `make build` | Build the executable                        |
| `make check` | Run formatting, vetting, testing, and build |

---

## Testing

GeoMind uses Go's standard testing tools.

The existing test suite covers:

### Configuration tests

- Default values.
- Custom environment values.
- Invalid environment names.
- Invalid log levels.
- Invalid HTTP addresses and ports.
- Invalid timeout values.
- Invalid header-size settings.

### HTTP tests

- Health endpoint response.
- Readiness endpoint response.
- Response status codes.
- JSON decoding.
- Content-Type.
- Unsupported methods.
- Unknown routes.
- Malformed HTTP request handling.

### Middleware tests

- Middleware execution order.
- Request ID propagation.
- Access log status capture.
- Implicit and explicit HTTP status codes.
- Panic recovery.
- Sensitive panic value exclusion.
- Response headers surviving middleware wrapping.
- ResponseController flushing through wrappers.

### Server lifecycle tests

- Request context cancellation after client disconnect.
- Graceful server shutdown.
- Server goroutine termination.

Run everything:

```bash
go test -count=1 -race ./...
go vet ./...
go build ./cmd/api
```

---

## Design Principles

GeoMind follows these engineering principles:

**1. Standard library first**

Use Go's standard library wherever it provides the required functionality. Avoid introducing dependencies without a clear benefit.

**2. Explicit dependencies**

Configuration, loggers, handlers, and services should be passed explicitly rather than fetched from global state.

**3. Fail fast**

Invalid configuration should prevent application startup.

**4. Separation of responsibilities**

Application startup, routing, middleware, handlers, configuration, and logging should remain independently understandable.

**5. Context propagation**

Request-scoped and long-running work should respect cancellation and deadlines.

**6. Bounded operations**

HTTP operations, database queries, external requests, and shutdown processes should have appropriate time limits.

**7. Safe logging**

Avoid logging credentials, authorization headers, sensitive request bodies, and personal information.

**8. Testable components**

Use small components and standard-library testing tools to verify behavior without requiring unnecessary infrastructure.

**9. Resource lifecycle management**

A component should stop using a resource before that resource is closed.

**10. No premature microservices**

The initial backend uses a modular monolith with independently runnable workers when required. Services will be split only when there's a meaningful operational or architectural reason.

---

## Planned Features

The following capabilities are part of GeoMind's wider roadmap but are not implemented in the current HTTP foundation.

### Database and geospatial foundation

- PostgreSQL integration.
- PostGIS extension.
- Spatial data modeling.
- Spatial indexes.
- Distance and proximity queries.
- Geographic filtering.

### Proximity engine

- Nearby place search.
- Radius-based queries.
- Geographic bounding-box queries.
- Efficient spatial APIs.
- Location indexing and caching.

### Redis GEO

- Real-time location indexing.
- Location lookup.
- Cache management.
- Cache invalidation strategies.

### Event processing

- NATS JetStream.
- Location update events.
- Asynchronous workers.
- Event replay.
- Background job lifecycle management.

### Frontend and mapping

- Next.js application.
- MapLibre GL JS.
- Interactive maps.
- Location search interface.
- Geographic visualization.

### AI and machine learning

- Natural-language location search.
- Structured AI outputs.
- Deterministic search execution.
- Location recommendations.
- Semantic search.
- Demand prediction baseline.

### Observability and deployment

- Prometheus metrics.
- Grafana dashboards.
- OpenTelemetry tracing.
- Docker deployment.
- AWS infrastructure.
- CI/CD pipelines.
- Performance and reliability testing.

---

## Project Status

GeoMind is under active development.

The initial backend foundation establishes the conventions and infrastructure for the remaining project phases.

The current focus is learning and implementing robust engineering fundamentals before introducing additional infrastructure and application complexity.

**Build deliberately. Understand the systems you create. Measure before optimizing.**
