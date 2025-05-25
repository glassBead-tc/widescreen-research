# Cloud Run MCP Drone Server Template

This template combines the best practices from Google Cloud Run microservices with MCP (Model Context Protocol) server capabilities, creating production-ready drone servers that can be dynamically spawned by a coordinator.

## Features

- 🚀 **Cloud Run Optimized**: Built on Google's Cloud Run microservice template
- 🤖 **MCP Protocol Support**: Full MCP server implementation with tools, resources, and prompts
- 📊 **Production Monitoring**: Structured logging, health checks, and metrics
- 🔄 **Multiple Transport Modes**: Support for both stdio and HTTP transports
- 🛡️ **Security**: Runs as non-root user, supports authentication
- 🎯 **Extensible**: Easy to create new drone types with specialized capabilities
- 🔍 **Research Integration**: Built-in Exa AI integration for powerful research capabilities

## Drone Types

### Generic Drone
Basic functionality for testing and simple operations:
- `echo`: Echo back messages
- `ping`: Health check
- `info`: Get drone information

### Scraper Drone
Web scraping capabilities:
- `fetch_url`: Fetch web pages
- `extract_data`: Extract data using selectors
- `parse_html`: Parse HTML content

### Processor Drone
Data transformation and processing:
- `transform_data`: Transform between formats
- `validate_data`: Validate data against schemas
- `aggregate`: Aggregate data

### Research Drone 🔍
**Powered by Exa AI** - Advanced research capabilities:
- `web_search`: Real-time web search with content extraction
- `research_papers`: Search academic papers and research content
- `company_research`: Comprehensive company information gathering
- `crawl_url`: Extract content from specific URLs
- `find_competitors`: Identify competitors for a company
- `linkedin_search`: Search LinkedIn for companies and people
- `wikipedia_search`: Search Wikipedia articles
- `github_search`: Search GitHub repositories

## Quick Start

### Local Development

```bash
# Install dependencies
npm install

# Run in development mode
npm run dev

# Run with specific drone type
DRONE_TYPE=scraper npm run dev

# Run research drone (requires EXA_API_KEY)
DRONE_TYPE=research EXA_API_KEY=your-key npm run dev
```

### Building for Production

```bash
# Build Docker image
docker build -t drone-mcp:latest .

# Run locally
docker run -p 8080:8080 \
  -e DRONE_TYPE=research \
  -e EXA_API_KEY=your-exa-key \
  -e GOOGLE_CLOUD_PROJECT=your-project \
  drone-mcp:latest
```

### Deploy to Cloud Run

```bash
# Set your project
export GOOGLE_CLOUD_PROJECT=your-project-id

# Build and push to Artifact Registry
DRONE_TYPE=research npm run build-image

# Deploy to Cloud Run
DRONE_TYPE=research npm run deploy
```

## Environment Variables

- `PORT`: HTTP port (default: 8080)
- `DRONE_TYPE`: Type of drone (generic, scraper, processor, research)
- `MCP_TRANSPORT`: Transport mode (stdio, http)
- `COORDINATOR_URL`: URL of coordinator service
- `GOOGLE_CLOUD_PROJECT`: GCP project ID
- `EXA_API_KEY`: Exa AI API key (required for research drone)

## Integration with Coordinator

The coordinator can spawn research drones using the Cloud Run API:

```go
// In your coordinator code
service := &runpb.Service{
    Template: &runpb.RevisionTemplate{
        Containers: []*runpb.Container{{
            Image: fmt.Sprintf("gcr.io/%s/drone-mcp:research", projectID),
            Env: []*runpb.EnvVar{
                {Name: "DRONE_TYPE", Value: &runpb.EnvVar_Value{Value: "research"}},
                {Name: "EXA_API_KEY", Value: &runpb.EnvVar_Value{Value: exaApiKey}},
                {Name: "COORDINATOR_URL", Value: &runpb.EnvVar_Value{Value: coordinatorURL}},
                {Name: "MCP_TRANSPORT", Value: &runpb.EnvVar_Value{Value: "http"}},
            },
        }},
    },
}
```

## Research Drone Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Coordinator   │────▶│ Research Drone  │────▶│  Exa MCP Server │
│   (Go Server)   │     │  (Node.js MCP)  │     │   (Remote API)  │
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                       │                       │
         │                       │                       └── https://mcp.exa.ai/mcp
         │                       │
         │                       └── Proxies Exa tools through MCP
         │
         └── Spawns via Cloud Run API
```

The research drone acts as a proxy to Exa's hosted MCP server, providing:
- **Scalable Research**: Each drone connects independently to Exa
- **Cost Efficiency**: Scale to zero when not in use
- **Security**: API keys isolated per drone instance
- **Reliability**: Built-in retry and error handling

## Health Checks and Monitoring

- `/health` - Basic health check
- `/ready` - Readiness probe (checks MCP server status)
- `/metrics` - Prometheus-compatible metrics
- `/` - Drone information and capabilities

## Creating New Drone Types

1. Create a new handler file in `drones/`:

```javascript
// drones/mydrone.js
export function createMyDroneHandlers() {
  return {
    tools: {
      my_tool: async (request) => {
        // Implementation
      }
    },
    resources: {},
    prompts: {}
  };
}
```

2. Add to `mcp-server.js`:

```javascript
import { createMyDroneHandlers } from './drones/mydrone.js';

// In getHandlersForType()
case 'mydrone':
  return createMyDroneHandlers();
```

## Production Best Practices

1. **Resource Limits**: Set appropriate CPU/memory limits in Cloud Run
2. **Scaling**: Configure min/max instances based on workload
3. **Security**: Use service accounts with minimal permissions
4. **Monitoring**: Enable Cloud Logging and Cloud Monitoring
5. **Cost**: Use scale-to-zero for infrequent workloads
6. **API Keys**: Store Exa API keys in Google Secret Manager

## Getting an Exa API Key

1. Visit [dashboard.exa.ai/api-keys](https://dashboard.exa.ai/api-keys)
2. Sign up for an account
3. Generate an API key
4. Add it to your environment variables or Secret Manager

## Architecture

```
┌─────────────────┐     ┌─────────────────┐
│   Coordinator   │────▶│   Cloud Run     │
│   (Go Server)   │     │      API        │
└─────────────────┘     └─────────────────┘
                                │
                                ▼
                        ┌─────────────────┐
                        │  Drone MCP      │
                        │    Server       │
                        ├─────────────────┤
                        │ Express Server  │
                        │   (Health)      │
                        ├─────────────────┤
                        │  MCP Server     │
                        │ (Tools/Resources)│
                        └─────────────────┘
                                │
                                ▼ (Research Drone Only)
                        ┌─────────────────┐
                        │  Exa MCP Server │
                        │   (Remote API)  │
                        └─────────────────┘
```

## License

Apache-2.0 