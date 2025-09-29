# widescreen-research

[![Go CI](https://github.com/glassBead-tc/widescreen-research/actions/workflows/ci-go.yaml/badge.svg)](https://github.com/glassBead-tc/widescreen-research/actions/workflows/ci-go.yaml)

Bidirectional MCP server orchestrating an army of research drones on Cloud Run in GCP.

## Repo status
- Primary implementation: Go (coordinator and drone)
- Examples: Node (under examples/node)

## Clone
```bash
git clone https://github.com/glassBead-tc/widescreen-research.git
cd widescreen-research
```

## Quickstart (Go)
```bash
go version  # require 1.23+
go mod download

# terminal 1
LOG_LEVEL=debug go run ./cmd/coordinator

# terminal 2
DRONE_TYPE=research EXA_API_KEY=your-local-only-key LOG_LEVEL=debug go run ./cmd/drone
```

Open http://localhost:8080 and GET /health.

## Docker (Go)
```bash
docker build -t widescreen/coordinator:dev .
docker run --rm -p 8080:8080 widescreen/coordinator:dev
```

## Deploy to Cloud Run (Go)
Preferred: inject secrets from Secret Manager. Example:

```bash
gcloud run deploy widescreen-coordinator \
  --source . \
  --region=us-central1 \
  --allow-unauthenticated \
  --set-env-vars=LOG_LEVEL=info \
  --set-secrets=EXA_API_KEY=exa_api_key:latest
```

## Node examples (optional)
The Node projects live under ./examples/node/ and are provided as examples only.
- Requires Node 18+ LTS
- See each example's README for usage

## Architecture and design
See [project_spec.md](./project_spec.md) for the canonical architecture, roles, and protocols.

## Directory layout
- `cmd/coordinator`: Go coordinator service (binary)
- `cmd/drone`: Go drone worker (binary)
- `pkg/...`: Reusable Go packages
- `examples/node/...`: Node example MCP servers and templates

## Operations
- Endpoints: GET / (info), GET /health (readiness)
- Logging: structured logs; control via LOG_LEVEL
- Recommended: enable OpenTelemetry to Cloud Monitoring (future work)

## Contributing
See CONTRIBUTING.md (coming soon). Please run go fmt/go vet and tests locally before submitting PRs.
