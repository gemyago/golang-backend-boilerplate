# OpenTelemetry Integration Plan

## Introduction/Overview

This plan details the integration of OpenTelemetry (OTel) into the golang-backend-boilerplate project to provide comprehensive observability through distributed tracing, metrics, and structured logging with trace correlation. OpenTelemetry is the industry-standard, vendor-agnostic observability framework that enables visibility into distributed systems for troubleshooting, performance optimization, and system understanding.

### Problem Statement

Currently, the application has basic logging with correlation IDs but lacks:
- Distributed tracing to track requests across components and external services
- Automatic instrumentation of HTTP handlers, database calls, and HTTP clients
- Correlation of logs with active traces (trace_id/span_id)
- Metrics collection for system and business operations
- Standardized telemetry export to observability backends

### Goal

Implement a production-ready OpenTelemetry integration that:
1. Provides automatic instrumentation for HTTP server, database, and HTTP clients
2. Enables manual span creation for critical business logic
3. Correlates all logs with active trace context
4. Exports telemetry data using OTLP protocol with proper configuration
5. Follows Go OTel best practices (API/SDK separation, resource identity, proper sampling)
6. Maintains backward compatibility and follows project architectural patterns

### Latest OpenTelemetry Go Status (2024-2025)

Based on official OpenTelemetry documentation:

**Stability Status:**
- **Traces API/SDK**: Stable
- **Metrics API/SDK**: Stable  
- **Logs API/SDK**: Beta

**Key Environment Variables (OTLP Exporter):**
Viper already supports automatic environment variable binding with `APP_` prefix and key transformations (dots/dashes to underscores). Standard OTel environment variables can be documented for external configuration, but internal config uses viper's JSON configuration with env var overrides.

- `APP_OPENTELEMETRY_ENABLED` - Enable/disable OTel (maps to `openTelemetry.enabled`)
- `APP_OPENTELEMETRY_ENDPOINT` - OTLP endpoint (maps to `openTelemetry.endpoint`)
- `APP_OPENTELEMETRY_PROTOCOL` - Protocol: `http/protobuf`, `grpc`, `http/json`
- `APP_OPENTELEMETRY_SAMPLINGRATE` - Sampling rate 0.0-1.0
- Standard OTel vars like `OTEL_EXPORTER_OTLP_HEADERS` can be used directly by exporters

**Logging Correlation:**
- `go.opentelemetry.io/contrib/bridges/otelslog` (v0.13.0+) provides native slog integration
- The `otelslog.Handler` bridges slog with OTel logging and automatically includes trace context

**Package Maintenance Status (Verified Nov 2025):**
- ✅ `opentelemetry-go`: **Highly Active** - 6.2k stars, released v1.38.0 Aug 29, 2025, daily commits, 362 contributors
- ✅ `opentelemetry-go-contrib`: **Highly Active** - 1.5k stars, released v1.38.0 Aug 29, 2025, daily commits, 265 contributors  
- ✅ `github.com/XSAM/otelsql`: **Active** - 375 stars, released v0.40.0 Sep 8, 2025, regular updates, used by 1.2k+ projects
- All packages are official OpenTelemetry projects or well-maintained community packages with active development</parameter>

## Business Logic

The integration will provide the following observability capabilities:

1. **Automatic Request Tracing**: Every HTTP request will create a root span with automatic propagation through middleware
2. **Database Operation Visibility**: All SQL queries will be traced with query details, duration, and errors
3. **External API Call Tracking**: HTTP client calls will create spans and propagate trace context to downstream services
4. **Business Logic Instrumentation**: Critical application code paths will have manual spans for detailed visibility
5. **Log Correlation**: All log entries will include `trace_id` and `span_id` when within an active trace context
6. **Error Tracking**: Errors will be properly recorded on spans with status codes
7. **Metrics Collection**: System and application metrics will be collected and exported
8. **Production-Ready Export**: Telemetry will be exported via OTLP with configurable backends

## High Level Architecture

The OpenTelemetry integration follows these architectural principles:

### Core Components

1. **OTel SDK Initialization** (`internal/infrastructure/otel/`)
   - Centralized SDK setup in infrastructure layer
   - Resource definition with service identity
   - TracerProvider, MeterProvider, and LoggerProvider configuration
   - OTLP exporter setup with environment-based configuration
   - Graceful shutdown handling

2. **Instrumentation Layer**
   - HTTP Server: Middleware wrapping all incoming requests
   - Database: SQL wrapper providing automatic query tracing
   - HTTP Client: RoundTripper middleware for outbound requests
   - Manual instrumentation helpers for application code

3. **Logging Integration** (`internal/diag/`)
   - Enhanced slog handler with OTel bridge
   - Automatic trace context injection into logs
   - Maintains existing correlation ID functionality

4. **Configuration** (`internal/config/`)
   - Environment variable-based configuration
   - Per-environment settings for sampling rates
   - Optional local development setup

### Component Interactions

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│  - Manual tracing (tracer.Start)                            │
│  - Business logic with span creation                         │
└──────────────────┬──────────────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────────────┐
│                  Incoming Adapters (API)                     │
│  - HTTP Middleware (otelhttp)                               │
│  - Automatic span creation for requests                      │
│  - Context propagation + trace_id extraction                 │
└──────────────────┬──────────────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────────────┐
│              Outgoing Adapters (Infrastructure)              │
│  - Database (otelsql)                                       │
│  - HTTP Client (otelhttp transport)                          │
│  - External service calls with context propagation           │
└──────────────────┬──────────────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────────────┐
│                   OTel SDK & Exporters                       │
│  - TracerProvider with BatchSpanProcessor                   │
│  - OTLP Exporter (viper config + env vars)                  │
│  - Sampling (ParentBased + TraceIDRatioBased)               │
│  - Graceful shutdown                                         │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   ▼
         [ OpenTelemetry Collector / Backend ]
```

### Correlation ID vs Trace ID Decision

**Current State:**
- Custom `x-correlation-id` header with UUID generation
- Stored in context via `LogAttributes.CorrelationID`
- Used for log correlation across requests

**OTel Trace Context:**
- W3C `traceparent` header format: `00-{trace-id}-{span-id}-{flags}`
- trace_id is 32 hex characters (128 bits)
- Automatically propagated across distributed systems

**Decision: Keep Both, Make Correlation ID Optional**

**Rationale:**
1. **W3C Standard**: trace_id is the industry standard and should be primary identifier
2. **Backward Compatibility**: Existing clients may send `x-correlation-id` headers
3. **External Integration**: Some systems expect custom correlation IDs separate from trace context
4. **Graceful Degradation**: When OTel is disabled, correlation ID provides request tracking

**Implementation Strategy:**
- **Primary**: Use trace_id from active span when OTel enabled
- **Fallback**: Use correlation ID when no active span or OTel disabled  
- **Logs**: Include both trace_id (from span) and correlationId (if provided) in structured logs
- **Response Headers**: Return trace_id in response for client-side correlation
- **Configuration**: Make correlation ID generation optional (enabled by default for backward compatibility)

This approach provides the best of both worlds: standard OTel trace propagation while maintaining compatibility with existing correlation ID infrastructure.
</parameter>

### API vs SDK Separation

Following OTel best practices:
- **API packages** (`go.opentelemetry.io/otel`) are imported throughout the application for creating spans and metrics
- **SDK packages** (`go.opentelemetry.io/otel/sdk`) are ONLY imported in `main` package and infrastructure setup code
- Libraries and application code remain vendor-agnostic by depending only on the API

## Detailed Architecture

### Phase 1: SDK Foundation (`internal/infrastructure/otel/`)

#### Files to Create

**`internal/infrastructure/otel/config.go`**
- Define configuration struct with fields for endpoint, protocol, sampling rate, enable flags
- Configuration loaded via viper (no direct env var parsing needed)
- Validation logic for configuration values
- Helper methods to construct exporter options from config
</parameter>

**`internal/infrastructure/otel/resource.go`**
- Create OTel Resource with semantic conventions
- Required attributes: `service.name`, `service.version`
- Optional attributes: `deployment.environment`, `host.name`
- Use `semconv` package for standardized keys

**`internal/infrastructure/otel/tracer.go`**
- Initialize TracerProvider with BatchSpanProcessor
- Configure OTLP trace exporter (gRPC or HTTP based on config)
- Set up ParentBased sampling with configurable TraceIDRatioBased
- Return provider and shutdown function

**`internal/infrastructure/otel/meter.go`**
- Initialize MeterProvider with PeriodicReader
- Configure OTLP metrics exporter
- Set up metric collection intervals
- Return provider and shutdown function

**`internal/infrastructure/otel/logger.go`**
- Initialize LoggerProvider with BatchLogRecordProcessor (Beta)
- Configure OTLP logs exporter
- Return provider and shutdown function

**`internal/infrastructure/otel/setup.go`**
- Main `SetupOTelSDK(ctx context.Context, config Config) (shutdown func(context.Context) error, err error)` function
- Initialize Resource
- Set up all providers (Tracer, Meter, Logger)
- Register providers as global (optional, but useful for contrib libraries)
- Return single shutdown function that flushes all providers
- Proper error handling and cleanup on initialization failure

**`internal/infrastructure/otel/register.go`**
- DI registration function for OTel components
- Provide TracerProvider, MeterProvider, and global Tracer instances
- Register shutdown hook with ShutdownHooks infrastructure
- Follow project's DI patterns

#### Files to Update

**`internal/config/provide.go`**
- Add OTel configuration value providers following existing patterns
- Use `provideConfigValue` for each config field with DI naming
- Example: `provideConfigValue(cfg, "openTelemetry.enabled").asBool()`

**`internal/config/default.json`**
```json
{
  "openTelemetry": {
    "enabled": false,
    "endpoint": "",
    "protocol": "http/protobuf",
    "samplingRate": 1.0,
    "enableMetrics": true,
    "enableLogs": false
  }
}
```

**`internal/config/local.json`**
```json
{
  "openTelemetry": {
    "enabled": false,
    "endpoint": "",
    "protocol": "http/protobuf",
    "samplingRate": 1.0
  }
}
```

**`internal/config/production.json`**
```json
{
  "openTelemetry": {
    "enabled": true,
    "endpoint": "http://localhost:4318",
    "protocol": "http/protobuf",
    "samplingRate": 0.1
  }
}
```

**`internal/infrastructure/register.go`**
- Call `otel.Register()` during infrastructure setup
</parameter>

### Phase 2: HTTP Server Instrumentation

#### Files to Create

**`internal/api/http/middleware/otel_tracing.go`**
- Create middleware using `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp`
- Wrap handler with `otelhttp.NewHandler`
- Configure operation naming to use route patterns (e.g., `/v1/users/{id}`)
- Integrate with existing TracingMiddleware for correlation ID
- Ensure context propagation to downstream handlers

**`internal/api/http/middleware/otel_tracing_test.go`**
- Test span creation for requests
- Test trace context extraction from headers
- Test correlation ID integration
- Test span attributes (http.method, http.route, http.status_code)

#### Files to Update

**`internal/api/http/server/server.go`**
- Add OTel middleware to middleware chain (before slog-http)
- Pass TracerProvider as dependency
- Update buildMiddlewareChain to include OTel middleware

**`internal/api/http/middleware/tracing.go`**
- Enhance to extract trace_id and span_id from active span
- Add them to LogAttributes alongside correlation ID
- Maintain backward compatibility with existing correlation ID logic

### Phase 3: Database Instrumentation

#### Files to Create

**`internal/infrastructure/otel_database.go`**
- Wrap `sql.Open` with `otelsql` instrumentation
- Use `github.com/XSAM/otelsql` library for SQL tracing
- Configure span name formatting (e.g., `db:query`, `db:exec`)
- Enable query parameter recording (configurable, disabled by default for security)
- Register database stats as metrics

**`internal/infrastructure/otel_database_test.go`**
- Test span creation for queries
- Test error recording on failed queries
- Test span attributes (db.system, db.statement, db.operation)

#### Files to Update

**`internal/infrastructure/database.go`**
- Modify `newDBProvider` to use OTel-wrapped database driver
- Ensure driver registration happens before `sql.Open`
- Add TracerProvider dependency

**`go.mod`**
- Add `github.com/XSAM/otelsql` dependency
- Add `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` dependency

### Phase 4: HTTP Client Instrumentation

#### Files to Update

**`internal/infrastructure/http/client_factory.go`**
- Add OTel transport middleware to CreateClient
- Use `otelhttp.NewTransport` to wrap base transport
- Configure transport to propagate trace context
- Apply middleware in order: Logging -> OTel -> Auth -> ErrorHandling -> BaseTransport
- Add TracerProvider dependency

**`internal/infrastructure/http/client_factory_test.go`**
- Test span creation for outbound requests
- Test trace context propagation (W3C TraceContext headers)
- Test span attributes (http.method, http.url, http.status_code)

### Phase 5: Log Correlation with slog

#### Files to Update

**`internal/diag/slog.go`**
- Replace or wrap `diagLogHandler` with `otelslog.Handler`
- Configure handler to extract trace_id and span_id from context
- Maintain correlation ID injection (now alongside trace IDs)
- Update `SetupRootLogger` to accept LoggerProvider
- Create composite handler: `diagLogHandler` wraps `otelslog.Handler` wraps base handler

**`internal/diag/testing.go`**
- Update test logger setup to support OTel handler
- Provide no-op LoggerProvider for tests

**`go.mod`**
- Add `go.opentelemetry.io/contrib/bridges/otelslog` dependency

### Phase 6: Manual Instrumentation Helpers

#### Files to Create

**`internal/infrastructure/otel/tracing.go`**
- Helper function `StartSpan(ctx context.Context, tracer trace.Tracer, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)`
- Helper function `RecordError(span trace.Span, err error, description string)` - combines RecordError + SetStatus
- Helper function `SetSpanAttributes(span trace.Span, attrs ...attribute.KeyValue)` - convenience wrapper
- Constants for common attribute keys

**`internal/infrastructure/otel/tracing_test.go`**
- Test helper functions
- Test error recording pattern
- Test attribute setting

#### Documentation

**`doc/opentelemetry-usage.md`**
- Developer guide for manual instrumentation
- Code examples for creating spans
- Best practices for error handling
- Context propagation patterns
- Examples for different scenarios (HTTP handlers, background jobs, business logic)

### Phase 7: Configuration and Bootstrap

#### Files to Update

**`cmd/server/main.go`**
- Initialize OTel SDK early in startup
- Defer shutdown function
- Handle errors during OTel setup

**`cmd/jobs/main.go`**
- Add OTel initialization for background jobs
- Ensure spans are created for job execution

**`.envrc.example`** (if exists) or create **`doc/environment-variables.md`**
- Document all OTel environment variables
- Provide examples for local development
- Provide examples for production (with collector)

</parameter>

## Key Architectural Decisions

### 1. Viper Configuration Pattern

**Decision**: All OTel configuration goes through viper's existing configuration system, not direct environment variable parsing.

**Rationale**: Maintains consistency with project patterns. Viper already handles env vars with `APP_` prefix and automatic key transformation. Allows layered configuration (default.json -> env.json -> env-user.json -> env vars).

**Impact**: Consistent configuration approach, easier testing, centralized config management.

### 2. Disabled by Default

**Decision**: OpenTelemetry is **disabled by default** in all environments except production.

**Rationale**: Zero-overhead for local development and testing. Developers opt-in when needed. Production explicitly enables observability. Prevents accidental telemetry export in development.

**Impact**: Cleaner local dev experience, explicit production configuration, no surprise overhead.

### 3. Correlation ID + Trace ID Coexistence

**Decision**: Keep correlation ID mechanism alongside OTel trace IDs. Use trace_id as primary identifier when OTel enabled, fall back to correlation ID.

**Rationale**: Backward compatibility with existing clients, graceful degradation when OTel disabled, flexibility for external integrations that require custom correlation headers.

**Impact**: Both identifiers in logs when OTel enabled, maintains existing functionality, supports gradual migration.

### 4. API-only Dependencies in Application Code

**Decision**: Application layer and incoming/outgoing adapters will only import OTel API packages, never SDK packages.

**Rationale**: This follows the OTel design principle of keeping instrumentation code vendor-agnostic. The API is stable and provides a "no-op" implementation by default. Only the infrastructure setup code imports SDK packages.

**Impact**: Better testability, no vendor lock-in, cleaner dependency graph.
</parameter>

### 2. Infrastructure Layer Ownership

**Decision**: All OTel setup, provider initialization, and exporter configuration lives in `internal/infrastructure/otel/`.

**Rationale**: Follows the project's hexagonal architecture where infrastructure components provide implementations. The setup is a technical concern, not business logic.

**Impact**: Clear separation of concerns, easier to test, follows project patterns.

### 3. Graceful Shutdown Integration

**Decision**: OTel shutdown is registered with the existing `ShutdownHooks` infrastructure.

**Rationale**: Reuses the project's existing shutdown mechanism. Ensures telemetry is flushed before application exit, preventing data loss.

**Impact**: Consistent shutdown handling across all components.

### 7. Viper-first Configuration with Env Override

**Decision**: Configuration defined in JSON files, overridable via `APP_*` environment variables through viper.

**Rationale**: Follows project's established configuration patterns. Viper provides layered config with precedence. Supports both file-based and env-based configuration seamlessly.

**Impact**: Consistent with project conventions, flexible configuration, easier testing with config files.
</parameter>

### 5. ParentBased Sampling Strategy

**Decision**: Use `ParentBased(TraceIDRatioBased(samplingRate))` instead of plain `TraceIDRatioBased`.

**Rationale**: Prevents broken traces where head and tail are sampled differently. If a parent span is sampled, all child spans are sampled. If no parent exists, sampling decision is based on trace ID.

**Impact**: Complete, coherent traces. Critical for distributed tracing correctness.

### 6. Error Recording Pattern

**Decision**: Always pair `span.RecordError(err)` with `span.SetStatus(codes.Error, "description")`.

**Rationale**: Recording an error alone doesn't mark the span as failed. Both are required for proper error visibility in tracing backends.

**Impact**: Errors are correctly surfaced in observability tools.

### 7. Log Correlation via otelslog

**Decision**: Use `go.opentelemetry.io/contrib/bridges/otelslog` to bridge slog with OTel logging.

**Rationale**: Native integration that automatically extracts trace context and converts slog records to OTel log records. Avoids manual context inspection.

**Impact**: Automatic trace correlation, simpler implementation, follows OTel patterns.

### 8. Optional OTel in Tests

**Decision**: Tests will use in-memory exporters or no-op providers, not real backends.

**Rationale**: Unit tests should be fast and not depend on external services. Integration tests can use in-memory exporters to verify span creation.

**Impact**: Fast tests, better isolation, easier CI/CD.

### 12. SQL Wrapper Approach

**Decision**: Use `github.com/XSAM/otelsql` library instead of manual instrumentation.

**Rationale**: Battle-tested library (375 stars, 1.2k+ users, actively maintained with Sep 2025 release) that handles edge cases, provides metrics, and follows OTel conventions. Well-documented with examples. Part of recommended OTel ecosystem.

**Impact**: Faster implementation, better coverage, production-proven, community-maintained with active development.
</parameter>

### 10. Collector-first Production Architecture

**Decision**: Documentation will recommend deploying with OpenTelemetry Collector as a sidecar/agent.

**Rationale**: The Collector provides buffering, retry, security (TLS), data transformation, and vendor flexibility. Direct export to backends couples the application to specific vendors.

**Impact**: Production-grade resilience, easier vendor migration, advanced sampling options.

## Uncertainties

1. **Existing Trace Context Handling**: The current `slog-http` middleware mentions `WithSpanID` and `WithTraceID`, suggesting some trace context handling exists. Need to verify if this conflicts with OTel or can be removed.
   - **Resolution**: Review `slog-http` implementation and disable its trace ID generation in favor of OTel.

2. **Database Driver Registration**: SQLite driver registration with OTel wrapper needs to be verified for compatibility with `modernc.org/sqlite`.
   - **Resolution**: Test with local database setup and verify span creation.

3. **MCP Server Instrumentation**: The project has MCP (Model Context Protocol) servers. Clarify if these need instrumentation.
   - **Resolution**: Add basic tracing to MCP handlers if they handle requests. Lower priority than HTTP.

4. **Metrics Strategy**: Determine which business and system metrics to collect beyond auto-instrumentation.
   - **Resolution**: Start with auto-instrumented metrics (HTTP request count/duration, DB query count/duration). Add custom metrics incrementally based on needs.

5. **Log Exporter Performance**: OTel logs are in Beta. Consider if log export should be optional or delayed.
   - **Resolution**: Make log export configurable. Enable for production, optionally disable for high-throughput environments.

6. **Testing Strategy for OTel**: Determine level of testing for OTel integration (unit vs integration).
   - **Resolution**: Unit tests for instrumentation components, integration tests for end-to-end span creation using in-memory exporters.

## Related Files

### New Files to Create
- `internal/infrastructure/otel/config.go`
- `internal/infrastructure/otel/resource.go`
- `internal/infrastructure/otel/tracer.go`
- `internal/infrastructure/otel/meter.go`
- `internal/infrastructure/otel/logger.go`
- `internal/infrastructure/otel/setup.go`
- `internal/infrastructure/otel/register.go`
- `internal/infrastructure/otel/tracing.go` (helpers)
- `internal/infrastructure/otel/tracing_test.go`
- `internal/infrastructure/otel_database.go`
- `internal/infrastructure/otel_database_test.go`
- `internal/api/http/middleware/otel_tracing.go`
- `internal/api/http/middleware/otel_tracing_test.go`
- `doc/opentelemetry-usage.md`
- `doc/environment-variables.md` (or update existing)

### Files to Update
- `go.mod` - Add OTel dependencies
- `internal/config/config.go` - Add OTel config struct
- `internal/config/default.json` - Add OTel defaults
- `internal/config/local.json` - Add local OTel config
- `internal/config/production.json` - Add production OTel config
- `internal/infrastructure/register.go` - Register OTel components
- `internal/infrastructure/database.go` - Add OTel wrapper
- `internal/infrastructure/http/client_factory.go` - Add OTel transport
- `internal/infrastructure/http/client_factory_test.go` - Test OTel transport
- `internal/api/http/server/server.go` - Add OTel middleware
- `internal/api/http/middleware/tracing.go` - Integrate trace IDs
- `internal/api/http/middleware/tracing_test.go` - Update tests
- `internal/diag/slog.go` - Add otelslog handler
- `internal/diag/testing.go` - Support OTel in tests
- `cmd/server/main.go` - Initialize OTel
- `cmd/server/start.go` - Add OTel dependencies
- `cmd/jobs/main.go` - Initialize OTel
- `README.md` - Document OTel feature

## Task List

This implementation will follow the TDD approach as per project conventions. Each task must leave the codebase in a buildable state with all tests passing.

### Task 1: Add OTel Dependencies

**Description**: Add all required OpenTelemetry packages to go.mod.

**Steps**:
- Run `go get go.opentelemetry.io/otel@latest`
- Run `go get go.opentelemetry.io/otel/sdk@latest`
- Run `go get go.opentelemetry.io/otel/exporters/otlp/otlptrace@latest`
- Run `go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc@latest`
- Run `go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@latest`
- Run `go get go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc@latest`
- Run `go get go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp@latest`
- Run `go get go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@latest`
- Run `go get go.opentelemetry.io/contrib/bridges/otelslog@latest`
- Run `go get github.com/XSAM/otelsql@latest`
- Run `go mod tidy`
- Verify compilation: `go build ./...`

**Success Criteria**: `make test` passes, `make lint` passes, dependencies added to go.mod

---

### Task 2: Create OTel Configuration Structure

**Description**: Define configuration structures and integrate with viper following project patterns.

**Steps**:
- Create `internal/infrastructure/otel/config.go`
- Define `Config` struct with fields:
  - `Enabled` (bool)
  - `Endpoint` (string)
  - `Protocol` (string)
  - `SamplingRate` (float64, 0.0-1.0)
  - `EnableMetrics` (bool)
  - `EnableLogs` (bool)
- Add validation method `Validate() error` on Config
- Add helper methods: `GetTracerProviderOptions()`, `GetMeterProviderOptions()` for exporter setup
- Write tests in `internal/infrastructure/otel/config_test.go`:
  - Valid configuration passes validation
  - Invalid sampling rate returns error
  - Invalid protocol returns error
  - Empty endpoint is allowed (for local dev)
  - Disabled config skips initialization
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestConfig`
  - Verify tests fail (no implementation yet)
- Implement validation logic and helper methods
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestConfig`
  - Verify tests pass
- Update `internal/config/provide.go`:
  - Add config value providers following existing pattern
  - `provideConfigValue(cfg, "openTelemetry.enabled").asBool()`
  - `provideConfigValue(cfg, "openTelemetry.endpoint").asString()`
  - `provideConfigValue(cfg, "openTelemetry.protocol").asString()`
  - `provideConfigValue(cfg, "openTelemetry.samplingRate").asFloat64()` (may need new provider method)
  - `provideConfigValue(cfg, "openTelemetry.enableMetrics").asBool()`
  - `provideConfigValue(cfg, "openTelemetry.enableLogs").asBool()`
- Update `internal/config/default.json`:
  - Add `openTelemetry` section with defaults: **enabled: false**, protocol: "http/protobuf", samplingRate: 1.0
- Update `internal/config/local.json`:
  - Add `openTelemetry` section: **enabled: false**, endpoint: "", samplingRate: 1.0
- Update `internal/config/production.json`:
  - Add `openTelemetry` section: **enabled: true**, endpoint: "http://localhost:4318", samplingRate: 0.1
- Add config documentation comment noting viper auto env var support: `APP_OPENTELEMETRY_ENABLED` etc.
</parameter>

**Success Criteria**: `make test` passes, `make lint` passes, configuration loads correctly

---

### Task 3: Implement Resource Creation

**Description**: Create OpenTelemetry Resource with service identity using semantic conventions.

**Steps**:
- Create `internal/infrastructure/otel/resource.go`
- Add stub function `func NewResource(serviceName, serviceVersion, environment string) (*resource.Resource, error)`
- Write tests in `internal/infrastructure/otel/resource_test.go`:
  - Resource contains service.name attribute
  - Resource contains service.version attribute
  - Resource contains deployment.environment attribute
  - Resource contains host.name attribute (from system)
  - Error handling for invalid inputs
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestResource`
  - Verify compilation errors are resolved (add any missing stubs)
  - Verify test failures are logical (expected values not matching)
- Implement `NewResource` function:
  - Use `resource.NewWithAttributes`
  - Use `semconv.ServiceName`, `semconv.ServiceVersion`
  - Use `semconv.DeploymentEnvironment`, `semconv.HostName`
  - Merge with default detectors (host, process, OS)
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestResource`
  - Verify all tests pass

**Success Criteria**: `make test` passes, `make lint` passes, resource correctly created with all attributes

---

### Task 4: Implement TracerProvider Setup

**Description**: Create and configure TracerProvider with OTLP exporter and proper sampling.

**Steps**:
- Create `internal/infrastructure/otel/tracer.go`
- Add stub function `func NewTracerProvider(ctx context.Context, resource *resource.Resource, config Config) (*sdktrace.TracerProvider, error)`
- Write tests in `internal/infrastructure/otel/tracer_test.go`:
  - TracerProvider is created successfully with valid config
  - gRPC exporter is used when protocol is "grpc"
  - HTTP exporter is used when protocol is "http/protobuf"
  - Sampler is ParentBased(TraceIDRatioBased(samplingRate))
  - Error when endpoint is invalid (for remote scenarios)
  - Resource is attached to provider
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestTracerProvider`
  - Resolve compilation errors
  - Verify test failures
- Implement `NewTracerProvider`:
  - Create OTLP exporter based on protocol (use `otlptracegrpc` or `otlptracehttp`)
  - Configure exporter with endpoint and headers from config
  - Handle empty endpoint (no-op exporter for local dev without collector)
  - Create `sdktrace.TracerProvider` with `BatchSpanProcessor`
  - Configure sampler: `sdktrace.ParentBased(sdktrace.TraceIDRatioBased(config.SamplingRate))`
  - Set resource on provider
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestTracerProvider`
  - Verify all tests pass

**Success Criteria**: `make test` passes, `make lint` passes, TracerProvider configured correctly

---

### Task 5: Implement MeterProvider Setup

**Description**: Create and configure MeterProvider with OTLP exporter for metrics collection.

**Steps**:
- Create `internal/infrastructure/otel/meter.go`
- Add stub function `func NewMeterProvider(ctx context.Context, resource *resource.Resource, config Config) (*sdkmetric.MeterProvider, error)`
- Write tests in `internal/infrastructure/otel/meter_test.go`:
  - MeterProvider created successfully
  - gRPC exporter used for "grpc" protocol
  - HTTP exporter used for "http/protobuf" protocol
  - PeriodicReader configured with appropriate interval
  - Resource attached
  - Disabled when EnableMetrics is false
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestMeterProvider`
  - Resolve compilation errors
  - Verify test failures
- Implement `NewMeterProvider`:
  - Return nil if `config.EnableMetrics == false`
  - Create OTLP metrics exporter based on protocol
  - Configure exporter with endpoint and headers
  - Create `sdkmetric.MeterProvider` with `PeriodicReader`
  - Set export interval (60 seconds default)
  - Set resource on provider
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestMeterProvider`
  - Verify all tests pass

**Success Criteria**: `make test` passes, `make lint` passes, MeterProvider configured correctly

---

### Task 6: Implement LoggerProvider Setup (Beta)

**Description**: Create and configure LoggerProvider with OTLP exporter for structured logs.

**Steps**:
- Create `internal/infrastructure/otel/logger.go`
- Add stub function `func NewLoggerProvider(ctx context.Context, resource *resource.Resource, config Config) (*sdklog.LoggerProvider, error)`
- Write tests in `internal/infrastructure/otel/logger_test.go`:
  - LoggerProvider created successfully when enabled
  - Returns nil when EnableLogs is false
  - Exporter configured based on protocol
  - Resource attached
  - BatchLogRecordProcessor configured
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestLoggerProvider`
  - Resolve compilation errors
  - Verify test failures
- Implement `NewLoggerProvider`:
  - Return nil if `config.EnableLogs == false`
  - Create OTLP logs exporter based on protocol (use Beta packages)
  - Configure exporter with endpoint and headers
  - Create `sdklog.LoggerProvider` with `BatchLogRecordProcessor`
  - Set resource on provider
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestLoggerProvider`
  - Verify all tests pass

**Success Criteria**: `make test` passes, `make lint` passes, LoggerProvider configured correctly

---

### Task 7: Implement Main Setup Function

**Description**: Create the main `SetupOTelSDK` function that initializes all providers and returns a shutdown function.

**Steps**:
- Create `internal/infrastructure/otel/setup.go`
- Add stub function `func SetupOTelSDK(ctx context.Context, serviceName, serviceVersion, environment string, config Config) (shutdown func(context.Context) error, err error)`
- Write tests in `internal/infrastructure/otel/setup_test.go`:
  - Setup succeeds with valid configuration
  - All providers initialized when enabled
  - Shutdown function flushes all providers
  - Shutdown function can be called multiple times safely
  - Setup returns error when resource creation fails
  - Providers are registered globally (otel.SetTracerProvider, otel.SetMeterProvider)
  - Cleanup happens on setup failure
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestSetup`
  - Resolve compilation errors
  - Verify test failures
- Implement `SetupOTelSDK`:
  - Call `NewResource` with service info
  - Call `NewTracerProvider`, `NewMeterProvider`, `NewLoggerProvider`
  - Register providers globally: `otel.SetTracerProvider(tp)`, `otel.SetMeterProvider(mp)`
  - Create composite shutdown function that calls all provider shutdowns
  - Handle errors during setup and cleanup partial initialization
  - Return shutdown function that accepts context with timeout
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestSetup`
  - Verify all tests pass

**Success Criteria**: `make test` passes, `make lint` passes, setup function works correctly

---

### Task 8: Create DI Registration for OTel

**Description**: Integrate OTel setup with the project's dependency injection system.

**Steps**:
- Create `internal/infrastructure/otel/register.go`
- Add function `func Register(container *dig.Container) error`
- Provide:
  - `*sdktrace.TracerProvider` (from global or direct)
  - `trace.Tracer` (obtained from provider with instrumentation scope)
  - `*sdkmetric.MeterProvider` (optional)
  - `metric.Meter` (optional)
  - Shutdown hook registration with `ShutdownHooks`
- Write tests in `internal/infrastructure/otel/register_test.go`:
  - Registration succeeds
  - Components can be resolved from container
  - Shutdown hook is registered
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestRegister`
  - Resolve compilation errors
  - Verify test failures
- Implement `Register`:
  - Accept config, service info from DI
  - Call `SetupOTelSDK`
  - Provide providers to container
  - Create named Tracer: `tracer := tp.Tracer("github.com/gemyago/golang-backend-boilerplate")`
  - Register shutdown function with `ShutdownHooks`
  - Handle conditional registration (skip if OTel disabled)
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestRegister`
  - Verify all tests pass
- Update `internal/infrastructure/register.go`:
  - Add call to `otel.Register(container)` after other infrastructure registrations
  - Handle error from registration

**Success Criteria**: `make test` passes, `make lint` passes, OTel registered in DI

---

### Task 9: Update Configuration Files

**Description**: Add OpenTelemetry configuration to all environment-specific config files.

**Steps**:
- Update `internal/config/default.json`:
  ```json
  "openTelemetry": {
    "enabled": false,
    "endpoint": "",
    "protocol": "http/protobuf",
    "headers": {},
    "samplingRate": 1.0,
    "enableMetrics": true,
    "enableLogs": false
  }
  ```
- Update `internal/config/local.json`:
  ```json
  "openTelemetry": {
    "enabled": true,
    "endpoint": "",
    "protocol": "http/protobuf",
    "samplingRate": 1.0,
    "enableMetrics": true,
    "enableLogs": false
  }
  ```
- Update `internal/config/production.json`:
  ```json
  "openTelemetry": {
    "enabled": true,
    "endpoint": "http://localhost:4318",
    "protocol": "http/protobuf",
    "samplingRate": 0.1,
    "enableMetrics": true,
    "enableLogs": true
  }
  ```
- Verify config loading: `go run ./cmd/server start --env local --noop`
- Check for config parsing errors in logs

**Success Criteria**: `make test` passes, `make lint` passes, all configs load successfully

---

### Task 10: HTTP Server Middleware Instrumentation

**Description**: Add OpenTelemetry tracing middleware to HTTP server to automatically trace all incoming requests.

**Steps**:
- Create `internal/api/http/middleware/otel_tracing.go`
- Add stub function `func NewOTelTracingMiddleware(tracer trace.Tracer) Middleware`
- Write tests in `internal/api/http/middleware/otel_tracing_test.go`:
  - Span created for request
  - Span name includes HTTP method and route
  - Span attributes include http.method, http.route, http.status_code
  - Trace context extracted from W3C TraceContext headers
  - Trace context propagated to handler context
  - Span marked as error when handler returns 5xx
  - Integration with existing correlation ID
- Run tests: `go test -v ./internal/api/http/middleware/ --run TestOTelTracing`
  - Resolve compilation errors
  - Verify test failures
- Implement `NewOTelTracingMiddleware`:
  - Use `otelhttp.NewHandler` to wrap next handler
  - Configure operation name to use route (may need custom formatter)
  - Use `otelhttp.WithTracerProvider` to pass tracer provider
  - Extract span from context and add to request context
  - Return middleware function
- Run tests: `go test -v ./internal/api/http/middleware/ --run TestOTelTracing`
  - Verify all tests pass
- Update `internal/api/http/server/server.go`:
  - Add `trace.Tracer` to `HTTPServerDeps`
  - Add OTel middleware to chain in `buildMiddlewareChain` (before slog-http, after existing tracing)
  - Pass tracer to middleware constructor
- Run integration test: `go test -v ./internal/api/http/server/ --run TestHTTPServer`
  - Verify no regressions

**Success Criteria**: `make test` passes, `make lint` passes, HTTP requests create spans

---

### Task 11: Enhance Correlation ID with Trace Context

**Description**: Enhance tracing middleware to extract trace_id/span_id from OTel spans while keeping correlation ID support.

**Steps**:
- Update `internal/api/http/middleware/tracing.go`:
  - After OTel middleware runs, extract span from context using `trace.SpanFromContext(ctx)`
  - Check if span is valid with `span.IsRecording()` or `span.SpanContext().IsValid()`
  - Get trace_id: `span.SpanContext().TraceID().String()`
  - Get span_id: `span.SpanContext().SpanID().String()`
  - Add them to `LogAttributes` struct
  - Keep existing correlation ID logic as fallback
- Update `internal/diag/slog.go`:
  - Add `TraceID` and `SpanID` fields to `LogAttributes` struct
  - Update `diagLogHandler.Handle` to include trace_id and span_id in log record
  - Include both trace_id (from OTel) and correlationId (from header/generated)
- Write tests in `internal/api/http/middleware/tracing_test.go`:
  - Test trace_id and span_id extracted when OTel span exists
  - Test correlation ID still works when no span
  - Test both trace_id and correlation ID present when OTel enabled
  - Test backward compatibility when OTel disabled
- Run tests: `go test -v ./internal/api/http/middleware/ --run TestTracing`
  - Verify test failures (new fields missing)
- Implement changes
- Run tests: `go test -v ./internal/api/http/middleware/ --run TestTracing`
  - Verify all tests pass
- Update documentation to explain dual-ID strategy
</parameter>

**Success Criteria**: `make test` passes, `make lint` passes, logs include trace context

---

### Task 12: Database Instrumentation

**Description**: Wrap database connections with OpenTelemetry to automatically trace all SQL queries.

**Steps**:
- Create `internal/infrastructure/otel_database.go`
- Add stub function `func WrapDatabase(driverName, dsn string, tracer trace.Tracer) (*sql.DB, error)`
- Write tests in `internal/infrastructure/otel_database_test.go`:
  - Wrapped database creates spans for queries
  - Span name includes operation type (SELECT, INSERT, etc.)
  - Span attributes include db.system, db.statement
  - Errors recorded on span when query fails
  - Span status set to error on query failure
  - Metrics recorded for query duration
- Run tests: `go test -v ./internal/infrastructure/ --run TestOTelDatabase`
  - Resolve compilation errors
  - Verify test failures
- Implement `WrapDatabase`:
  - Use `otelsql.Register` to register instrumented driver
  - Use `otelsql.Open` to open database with tracing
  - Configure options: `otelsql.WithTracerProvider`, `otelsql.WithAttributes`
  - Set db.system attribute to "sqlite"
  - Enable query recording (configurable, default disabled for security)
- Run tests: `go test -v ./internal/infrastructure/ --run TestOTelDatabase`
  - Verify all tests pass
- Update `internal/infrastructure/database.go`:
  - Add `trace.Tracer` to `DatabaseConfig` dependencies
  - Replace `sql.Open` call with `WrapDatabase`
  - Ensure foreign key pragma still executes
- Run existing database tests: `go test -v ./internal/infrastructure/ --run TestDatabase`
  - Verify no regressions

**Success Criteria**: `make test` passes, `make lint` passes, database queries create spans

---

### Task 13: HTTP Client Instrumentation

**Description**: Add OpenTelemetry tracing to HTTP client factory to trace outbound requests and propagate trace context.

**Steps**:
- Update `internal/infrastructure/http/client_factory.go`:
  - Add `trace.Tracer` to `ClientFactoryDeps`
  - Add OTel transport to middleware chain in `CreateClient`
  - Use `otelhttp.NewTransport` to wrap base transport
  - Configure transport with tracer provider
  - Place OTel transport after logging, before auth
- Write tests in `internal/infrastructure/http/client_factory_test.go`:
  - Test span created for outbound request
  - Test trace context propagated in headers (traceparent)
  - Test span attributes (http.method, http.url, http.status_code)
  - Test error recorded on non-2xx responses
  - Test integration with existing middleware
- Run tests: `go test -v ./internal/infrastructure/http/ --run TestClientFactory`
  - Verify compilation (add tracer to test setup)
  - Verify test failures
- Implement changes:
  - Add tracer field to `ClientFactory`
  - In `CreateClient`, wrap base transport with `otelhttp.NewTransport`
  - Configure transport options
  - Apply in middleware order
- Run tests: `go test -v ./internal/infrastructure/http/ --run TestClientFactory`
  - Verify all tests pass

**Success Criteria**: `make test` passes, `make lint` passes, HTTP client requests create spans and propagate context

---

### Task 14: Integrate otelslog for Log Correlation

**Description**: Replace or enhance diagLogHandler with otelslog to automatically correlate logs with traces.

**Steps**:
- Update `internal/diag/slog.go`:
  - Modify `SetupRootLogger` to accept `log.LoggerProvider` (optional)
  - Create `otelslog.Handler` wrapping base handler (JSONHandler or TextHandler)
  - Wrap `otelslog.Handler` with `diagLogHandler` to maintain correlation ID
  - Use composite handler chain: diagLogHandler -> otelslog.Handler -> base handler
- Update `internal/diag/testing.go`:
  - Provide no-op `log.LoggerProvider` for test logger
- Write tests in `internal/diag/slog_test.go` (create if not exists):
  - Test log records include trace_id when span active
  - Test log records include span_id when span active
  - Test correlation ID still included
  - Test no-op when LoggerProvider is nil
- Update existing test setup to provide nil LoggerProvider
- Run tests: `go test -v ./internal/diag/ --run TestSlog`
  - Verify test failures
- Implement changes:
  - Add otelslog integration
  - Maintain backward compatibility
  - Ensure proper handler chaining
- Run tests: `go test -v ./internal/diag/ --run TestSlog`
  - Verify all tests pass

**Success Criteria**: `make test` passes, `make lint` passes, logs automatically include trace context

---

### Task 15: Manual Instrumentation Helpers

**Description**: Create helper functions and utilities for manual span creation in application code.

**Steps**:
- Create `internal/infrastructure/otel/tracing.go`
- Add helper functions:
  - `StartSpan(ctx, tracer, name, opts) (context.Context, trace.Span)`
  - `RecordError(span, err, description)` - combines RecordError + SetStatus
  - `EndSpan(span)` - wrapper for consistent span ending
  - `SetSpanAttributes(span, ...attrs)`
- Define common attribute keys as constants
- Write tests in `internal/infrastructure/otel/tracing_test.go`:
  - Test StartSpan creates span and returns new context
  - Test RecordError sets both error and status
  - Test EndSpan can be called safely
  - Test SetSpanAttributes adds attributes correctly
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestTracingHelpers`
  - Resolve compilation errors
  - Verify test failures
- Implement helper functions
- Run tests: `go test -v ./internal/infrastructure/otel/ --run TestTracingHelpers`
  - Verify all tests pass

**Success Criteria**: `make test` passes, `make lint` passes, helper functions work correctly

---

### Task 16: Add Manual Instrumentation Example

**Description**: Add manual tracing to a sample application component to demonstrate usage patterns.

**Steps**:
- Choose a component for example (e.g., `internal/app/users_commands.go`)
- Add manual span creation in a method (e.g., `CreateUser`)
- Inject `trace.Tracer` as dependency
- Create span: `ctx, span := tracer.Start(ctx, "UsersCommands.CreateUser")`
- Add `defer span.End()`
- Add attributes to span (user email, etc. - avoid PII)
- Record errors if operation fails
- Write tests verifying span creation
- Run tests: `go test -v ./internal/app/ --run TestUsersCommands`
  - Verify no regressions
- Ensure span is created in integration tests

**Success Criteria**: `make test` passes, `make lint` passes, manual instrumentation example works

---

### Task 17: Update Main Entry Points

**Description**: Initialize OpenTelemetry SDK in main entry points (server, jobs) and ensure proper shutdown.

**Steps**:
- Update `cmd/server/main.go`:
  - Import OTel packages
  - Initialize OTel SDK early (before other setup)
  - Defer shutdown function with timeout context
  - Log OTel setup success/failure
  - Handle --noop flag appropriately
- Update `cmd/jobs/main.go`:
  - Similar OTel initialization
  - Ensure jobs create spans
- Test startup: `go run ./cmd/server start --env local --noop`
  - Verify OTel initialization logs appear
  - Verify no errors
- Test shutdown: Start server and send SIGTERM
  - Verify shutdown logs indicate OTel flushed

**Success Criteria**: `make test` passes, `make lint` passes, OTel initializes and shuts down correctly

---

### Task 18: Create Usage Documentation

**Description**: Write comprehensive documentation for developers on using OpenTelemetry in the application.

**Steps**:
- Create `doc/opentelemetry-usage.md`
- Include sections:
  - Overview of OTel in this project
  - How to enable/disable OTel (config file + env vars)
  - Configuration reference:
    - File-based: `openTelemetry.*` keys in config JSON
    - Env vars: `APP_OPENTELEMETRY_*` following viper patterns
    - Standard OTel env vars: `OTEL_EXPORTER_OTLP_HEADERS` (used by exporters directly)
  - Correlation ID + Trace ID explanation:
    - When both are present
    - Which to use when
    - Log output examples
  - Manual instrumentation patterns with code examples
  - Error handling best practices (RecordError + SetStatus)
  - Context propagation guidelines
  - Testing with OTel (disabled by default)
  - Troubleshooting common issues
  - Local development setup (disabled by default, how to enable)
  - Production deployment recommendations (Collector setup)
- Create `doc/environment-variables.md` (or update existing)
- Document viper env var mapping:
  - `APP_OPENTELEMETRY_ENABLED=true` → `openTelemetry.enabled`
  - `APP_OPENTELEMETRY_ENDPOINT=http://collector:4318` → `openTelemetry.endpoint`
  - `APP_OPENTELEMETRY_PROTOCOL=grpc` → `openTelemetry.protocol`
  - `APP_OPENTELEMETRY_SAMPLINGRATE=0.5` → `openTelemetry.samplingRate`
- Document standard OTel env vars that exporters use directly
- Provide examples for local dev, staging, and production
- Note package maintenance status and update policy
</parameter>

**Success Criteria**: Documentation is complete, accurate, and includes working examples

---

### Task 19: Update README and Project Documentation

**Description**: Update main README to document the OpenTelemetry integration feature.

**Steps**:
- Update `README.md`:
  - Add OpenTelemetry to key features list
  - Add link to detailed documentation
  - Mention observability capabilities
- Update `internal/AGENTS.md` if needed:
  - Add notes about OTel integration patterns
  - Reference usage documentation
- Verify all documentation links work
- Ensure consistency across docs

**Success Criteria**: README updated, documentation complete and consistent

---

### Task 20: Integration Testing and Validation

**Description**: Create integration tests to validate end-to-end OpenTelemetry functionality.

**Steps**:
- Create integration test file `internal/infrastructure/otel/integration_test.go`
- Test scenarios:
  - Full trace from HTTP request through database query
  - Span hierarchy is correct (parent-child relationships)
  - Trace context propagates through system
  - Logs include trace_id and span_id
  - Errors are recorded correctly on spans
  - Sampling works as configured
- Use in-memory exporter for test assertions
- Run integration tests: `go test -v ./internal/infrastructure/otel/ --run TestIntegration`
- Verify all scenarios pass
- Run full test suite: `make test`
- Verify no regressions in any package

**Success Criteria**: `make test` passes, `make lint` passes, integration tests validate full functionality

---

### Task 21: Final Validation and Cleanup

**Description**: Perform final validation, cleanup, and prepare for deployment.

**Steps**:
- Run full test suite: `make test`
  - Verify all tests pass
  - Check coverage report
- Run linter: `make lint`
  - Fix any issues
- Test all entry points with --noop flag:
  - `go run ./cmd/server start --env local --noop`
  - `go run ./cmd/jobs echo --env local --noop`
  - Verify no errors, OTel initializes
- Test with OTel disabled:
  - Set config `openTelemetry.enabled: false`
  - Verify application works normally
  - Verify no OTel logs appear
- Review all new code for:
  - Error handling completeness
  - Test coverage
  - Documentation comments
  - Adherence to project patterns
- Update AGENTS.md if new patterns introduced
- Create summary document: `doc/implementation/plan-opentelemetry-integration/summary.md`
  - Summarize what was implemented
  - Document any deviations from plan
  - Note any follow-up items

**Success Criteria**: `make test` passes (100%), `make lint` passes (0 issues), all documentation complete, summary written

---

## Summary of Implementation

This plan provides a comprehensive, production-ready OpenTelemetry integration following industry best practices and Go OTel conventions. The implementation is structured in phases to ensure:

1. **Stable Foundation**: SDK setup with proper configuration, resource identity, and lifecycle management
2. **Automatic Instrumentation**: HTTP server, database, and HTTP client instrumentation with zero code changes required in most application code
3. **Observability Excellence**: Log correlation, proper error handling, complete traces, and metrics collection
4. **Production Readiness**: Environment-based configuration, proper sampling, graceful shutdown, and collector-first architecture recommendations

The phased approach allows for incremental implementation and testing, with each task leaving the codebase in a fully functional, tested state. The integration maintains the project's architectural principles (hexagonal architecture, DI patterns, TDD approach) while adding world-class observability capabilities.