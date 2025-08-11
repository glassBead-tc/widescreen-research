# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a Go-based MCP (Model Context Protocol) server implementation for "Widescreen Research" - a distributed research orchestration system that provisions and coordinates multiple research drones on Google Cloud Platform to conduct parallel research at scale.

## Architecture

The codebase has two main components:
1. **Go Implementation (Primary)**: The canonical control plane and workers under `cmd/`, `pkg/`
2. **Node.js Template**: Legacy/example implementation under `drone-mcp-template/` and `widescreen-research-mcp/`

Key architectural components:
- **MCP Server**: Main server handling the `widescreen-research` tool (`cmd/widescreen-research-mcp/server/`)
- **Orchestrator**: Coordinates research by managing drones and aggregating results (`cmd/widescreen-research-mcp/orchestrator/`)
- **Research Drones**: Cloud Run containers performing actual research (`cmd/drone/`, `pkg/drone/`)
- **GCP Integration**: Provisioning and managing Cloud Run, Pub/Sub, and Firestore resources

## Development Commands

### Build
```bash
# Build the main MCP server
go build -o widescreen-research ./cmd/widescreen-research-mcp

# Build coordinator
go build ./cmd/coordinator

# Build drone
go build ./cmd/drone
```

### Test
```bash
# Run all tests
go test ./...

# Run specific package tests with verbose output
go test ./cmd/widescreen-research-mcp/orchestrator -v

# Run with coverage
go test ./... -cover
```

### Run
```bash
# Run the MCP server locally
./widescreen-research

# Or directly with go run
go run ./cmd/widescreen-research-mcp
```

### Docker
```bash
# Build Docker image for widescreen-research MCP
docker build -f cmd/widescreen-research-mcp/Dockerfile -t widescreen-research-mcp .

# Build Docker image for drone
docker build -f cmd/drone/Dockerfile -t research-drone .
```

## Code Organization

- `cmd/`: Entry points for various executables
  - `widescreen-research-mcp/`: Main MCP server implementation
    - `server/`: MCP protocol server using mcp-go v0.29
    - `orchestrator/`: Research orchestration logic
    - `operations/`: Core operations (data analysis, GCP provisioning, sequential thinking)
    - `schemas/`: Type definitions for MCP protocol
  - `coordinator/`: Standalone coordinator service
  - `drone/`: Research drone worker
- `pkg/`: Shared packages
  - `coordinator/`: Campaign coordination logic
  - `drone/`: Drone worker implementation
  - `gcp/`: Google Cloud client utilities
  - `mcp/`: MCP server utilities
  - `types/`: Shared type definitions
- `fixtures/`: Test fixtures and sample data

## Current MCP Implementation

The server currently uses `github.com/mark3labs/mcp-go v0.29.0` but is being refactored to use the official MCP Go SDK (`github.com/modelcontextprotocol/go-sdk`). See `MCP-SDK-Refactor-Spec.md` for migration details.

Key MCP operations exposed:
- `widescreen-research`: Main entry point with elicitation-based qualification
- `orchestrate-research`: Coordinates distributed research execution
- `sequential-thinking`: Step-by-step reasoning for complex problems
- `gcp-provision`: Provisions GCP resources
- `analyze-findings`: Analyzes collected research data

## Environment Variables

Required for running:
- `GOOGLE_CLOUD_PROJECT`: GCP project ID
- `GOOGLE_CLOUD_REGION`: Default region (default: us-central1)
- `CLAUDE_API_KEY`: Claude API key (optional, for AI capabilities)
- `ORCHESTRATOR_URL`: Orchestrator callback URL
- `EXA_API_KEY`: Exa AI API key (for Node.js research drones)

## Testing Strategy

- Unit tests exist in `*_test.go` files alongside implementation
- Main test file: `cmd/widescreen-research-mcp/orchestrator/orchestrator_test.go`
- Tests may skip if `GOOGLE_CLOUD_PROJECT` is not set
- Mock implementations are used for external dependencies

## Important Notes

1. **Build Issues**: The code has been fixed to resolve unused variable errors in `data_analyzer.go` and updated to use the new mcp-go v0.29 API in `server.go`

2. **Git Status**: Currently on `main` branch with uncommitted changes to:
   - `cmd/widescreen-research-mcp/operations/data_analyzer.go`
   - `cmd/widescreen-research-mcp/server/server.go`

3. **MCP SDK Migration**: Active refactoring to migrate from third-party mcp-go to official SDK - check `MCP-SDK-Refactor-Spec.md` before making protocol-level changes

4. **Research Flow**: 
   - User initiates with `widescreen-research` tool
   - Elicitation phase qualifies requirements
   - Orchestrator provisions drones on GCP
   - Drones conduct parallel research
   - Results collected via Pub/Sub queue
   - Final report generated and returned