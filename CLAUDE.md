# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**widescreen-research** is a bidirectional MCP server orchestrating distributed research drones on Google Cloud Platform. The system enables horizontal research at scale through a coordinator-worker architecture where a high-powered coordinator dynamically spawns lightweight drone MCP servers on Cloud Run.

Primary language: **Go 1.24+**

## Commands

### Building and Running

```bash
# Build all binaries
make build
go build ./cmd/coordinator ./cmd/drone

# Run coordinator locally
make run-coordinator
# or with custom config:
LOG_LEVEL=debug GOOGLE_CLOUD_PROJECT=your-project-id GOOGLE_CLOUD_REGION=us-central1 go run ./cmd/coordinator

# Run drone locally
make run-drone
# or with custom config:
LOG_LEVEL=debug DRONE_TYPE=research EXA_API_KEY=dummy DRONE_ID=local-drone-1 go run ./cmd/drone

# Run widescreen-research MCP server
go run ./cmd/widescreen-research-mcp
```

### Testing

```bash
# Run all tests with race detection and coverage
make test
go test -race -cover ./...

# Run tests for specific package
go test -v ./pkg/coordinator/...

# Note: Most packages currently have [no test files]. When adding tests, follow Go testing conventions.
```

### Quality Checks

```bash
# Run all linters
make lint

# Run specific linters
make lint-go        # golangci-lint with auto-fix
make lint-docker    # hadolint for Dockerfiles

# Run security scans
make security-scan  # gosec + detect-secrets

# Pre-push validation (lint + test + security)
make pre-push
```

### Docker

```bash
# Build Docker image
make docker
docker build -t widescreen/coordinator:dev .

# Run container
docker run --rm -p 8080:8080 widescreen/coordinator:dev
```

### Git Hooks

```bash
# Install pre-commit hooks (formats code, runs linters, validates commits)
make install-hooks

# Remove hooks
make clean-hooks
```

## Architecture

### Coordinator-Drone Pattern

The system implements a distributed MCP architecture:

1. **Coordinator** (`cmd/coordinator`): High-powered server that orchestrates drone lifecycle, manages task distribution, and aggregates results. Exposes MCP tools for planning and executing distributed research.

2. **Drones** (`cmd/drone`): Lightweight workers deployed on Cloud Run. Each drone is a specialized MCP server that performs specific research tasks (web search, data scraping, API queries).

3. **Widescreen Research MCP** (`cmd/widescreen-research-mcp`): Advanced orchestration server with elicitation-based qualification, bidirectional MCP capabilities, and AI-powered report generation.

### Key Integration Points

- **GCP Services**: Cloud Run (drone hosting), Pub/Sub (result queuing), Firestore (state management), Secret Manager (API keys)
- **MCP SDKs**: Uses both `github.com/mark3labs/mcp-go` and official `github.com/modelcontextprotocol/go-sdk`
- **External APIs**: Exa AI for research capabilities (requires `EXA_API_KEY`)

### Package Structure

```
pkg/
├── coordinator/     # Coordinator server logic, MCP client, task planning
│   ├── server.go    # Main coordinator server with PlanDistributedTask, ExecuteTask
│   ├── mcp_client.go # MCP client for drone communication
│   └── campaign.go  # Campaign management for research tasks
├── drone/           # Drone worker implementation
├── gcp/             # GCP client utilities (Cloud Run, Pub/Sub, Firestore)
├── mcp/             # MCP protocol implementations
└── types/           # Shared type definitions (DroneInfo, TaskResult, ExecutionPlan)
```

### HTTP Endpoints

**Coordinator**:

- `GET /` - Service info
- `GET /health` - Health check
- `POST /api/drones/register` - Drone registration

**Drone**:

- `GET /health` - Health check
- `POST /mcp` - MCP protocol endpoint

## Development Guidelines

### Environment Variables

**Required**:

- `GOOGLE_CLOUD_PROJECT` - GCP project ID
- `EXA_API_KEY` - Exa AI API key for research drones

**Common**:

- `LOG_LEVEL` - Logging verbosity: debug, info, warn, error (default: info)
- `PORT` - HTTP server port (default: 8080)
- `GCP_REGION` - GCP region (default: us-central1)
- `DRONE_TYPE` - Drone capability type (research, scraper, etc.)
- `COORDINATOR_BASE_URL` - Coordinator callback URL for local dev

See `.env.example` for complete list.

### Commit Convention

Uses Conventional Commits:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code restructuring
- `test:` - Test additions/fixes
- `chore:` - Maintenance tasks

Enforced by pre-commit hooks.

### Code Style

- **Formatting**: Automatic via `gofmt` and `goimports` (runs on save in VS Code, via pre-commit hooks)
- **Linting**: `golangci-lint` with project-specific rules in `.golangci.yml`
- **Logging**: Structured JSON logs with fields: timestamp, level, msg, component, request_id, error
- **Error Handling**: Wrap errors with context using `fmt.Errorf("context: %w", err)`

## Testing Patterns

When writing tests for this codebase (currently minimal test coverage):

1. **Unit Tests**: Place in same package as code (`*_test.go`)
2. **Integration Tests**: May require GCP credentials; mock GCP clients when possible
3. **Table-Driven Tests**: Preferred Go pattern for testing multiple cases
4. **Test Fixtures**: Available in `fixtures/` directory

Example test structure:

```go
func TestPlanDistributedTask(t *testing.T) {
    tests := []struct {
        name        string
        taskDesc    string
        wantDrones  int
        wantErr     bool
    }{
        // test cases...
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic...
        })
    }
}
```

## Key Implementation Details

### Drone Lifecycle Management

Coordinator manages drones through:

1. **Planning**: `PlanDistributedTask()` calculates required drone count, estimates cost/time (pkg/coordinator/server.go:40)
2. **Spawning**: GCP client provisions Cloud Run services dynamically
3. **Registration**: Drones register via `/api/drones/register` endpoint
4. **Execution**: Tasks distributed via MCP protocol
5. **Collection**: Results aggregated through Pub/Sub queue
6. **Cleanup**: Automatic resource cleanup after task completion

### MCP Protocol Usage

The system uses MCP bidirectionally:

- **Coordinator as MCP Server**: Exposes tools for task planning and execution
- **Coordinator as MCP Client**: Communicates with drone MCP servers
- **Drones as MCP Servers**: Provide research capabilities to coordinator

See `pkg/coordinator/mcp_client.go` for client implementation patterns.

### State Management

- **Firestore Collections**:
  - `execution_plans` - Task execution plans
  - `drone_registry` - Active drone tracking
  - `task_results` - Aggregated results
- **In-Memory State**: `Server.activeDrones`, `Server.taskResults` (protected by sync.RWMutex)

### GCP Resource Patterns

- **Authentication**: Application Default Credentials (ADC) or service account keys
- **Container Images**: Multi-stage builds with distroless base (~2MB)
- **Scaling**: Scale-to-zero with cold start optimization (~1-2 seconds)
- **Cost Optimization**: Break-even at ~40 requests/hour; below that, use scale-to-zero

## Deployment

### Local Development

1. Copy `.env.example` to `.env` and fill in values
2. Ensure GCP authentication: `gcloud auth application-default login`
3. Enable required APIs: Cloud Run, Pub/Sub, Firestore
4. Run services with `make run-coordinator` / `make run-drone`

### Cloud Run Production

```bash
# Deploy coordinator
gcloud run deploy widescreen-coordinator \
  --source . \
  --region=us-central1 \
  --allow-unauthenticated \
  --set-env-vars=LOG_LEVEL=info \
  --set-secrets=EXA_API_KEY=exa_api_key:latest

# Drones are provisioned dynamically by coordinator
```

See `project_spec.md` for comprehensive architecture documentation and `docs/OPERATIONS.md` for operational guidance.

## Related Projects

- **clearthought-onepointfive/**: TypeScript MCP server for systematic thinking and mental models (separate Node.js project)
- **examples/node/**: Node.js MCP server examples and templates

## Resources

- Project Spec: `project_spec.md` - Canonical architecture and protocols
- Operations Guide: `docs/OPERATIONS.md` - Monitoring, troubleshooting, performance
- Contributing: `CONTRIBUTING.md` - Development workflow and standards
- Widescreen MCP README: `cmd/widescreen-research-mcp/README.md` - Advanced orchestration features
