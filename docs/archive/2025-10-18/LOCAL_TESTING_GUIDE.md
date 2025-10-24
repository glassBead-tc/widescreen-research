# Local Testing Guide for Widescreen-Research

This guide walks you through testing the widescreen-research MCP servers locally using proper MCP testing tools.

## 📋 Prerequisites

### Required

- **Go 1.23+** - Check with `go version`
- **Node.js 20+** - For MCPJam CLI
- **GCP Project** - Already configured (see `GCP_SETUP_COMPLETE.md`)
- **Environment Variables** - Set in `.env.local`
- **GCP Authentication** - Application Default Credentials
- **MCPJam CLI** - For MCP server testing

### Optional

- **Docker** - For containerized testing
- **Anthropic/OpenAI API Keys** - For LLM-based testing

## 🔐 Authentication Setup

Before running any services, authenticate with GCP:

```bash
# Authenticate with your Google account
gcloud auth application-default login

# Verify authentication
gcloud auth list

# Verify project is set
gcloud config get-value project
# Should output: widescreen-researcher
```

## 🧪 Understanding MCP Testing

**Important**: MCP servers are NOT REST APIs. They implement the Model Context Protocol, which is:

- **Stateful** - Maintains session state across requests
- **Bidirectional** - Server and client can both initiate communication
- **Protocol-driven** - Requires proper initialization, capability negotiation, and tool discovery

**You cannot test MCP servers with simple curl commands.** You need an MCP client.

### Testing Tools

1. **MCPJam CLI** (Recommended) - Automated testing with evals
2. **MCPJam Inspector** - Interactive GUI for manual testing
3. **MCP Inspector** - Official Anthropic inspector

---

## 🎯 Option 1: Testing with MCPJam CLI (Recommended)

### Terminal 1: Start the Coordinator

```bash
# Load environment variables
source .env.local

# Run coordinator
go run ./cmd/coordinator
```

**Expected Output:**

```
Starting Spawn MCP Coordinator...
Initialized GCP client for project widescreen-researcher in region us-central1
Coordinator HTTP server listening on :8080
```

**What it does:**

- Initializes GCP clients (Cloud Run, Firestore, Pub/Sub)
- Starts HTTP server on port 8080
- Exposes endpoints: `/`, `/health`, `/drones`, `/tasks`

### Terminal 2: Start a Drone

```bash
# Load environment variables
source .env.local

# Run drone on different port
PORT=8081 go run ./cmd/drone
```

**Expected Output:**

```
Starting Drone MCP Server...
Researcher Drone HTTP listening on :8081
```

**What it does:**

- Connects to Pub/Sub topic `drone-tasks`
- Starts HTTP server on port 8081
- Exposes endpoints: `/health`, `/task`, `/mcp`

### Terminal 3: Test the Services

```bash
# Test coordinator health
curl http://localhost:8080/health
# Expected: {"status":"healthy"}

# Test coordinator info
curl http://localhost:8080/
# Expected: JSON with coordinator info

# Test drone health
curl http://localhost:8081/health
# Expected: ok

# Check active drones
curl http://localhost:8080/drones
# Expected: {"count":0,"drones":[]} (no drones registered yet)

# Test drone task execution
curl -X POST http://localhost:8081/task \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "artificial intelligence",
    "time_frame": "2024",
    "sources": ["academic", "news"],
    "max_results": 5
  }'
# Expected: JSON with research results
```

### Testing Pub/Sub Communication

```bash
# In Terminal 3, publish a task to the drone
gcloud pubsub topics publish drone-tasks \
  --message='{"task_id":"test-001","drone_id":"test-drone-001","task_type":"research","parameters":{"topic":"AI"}}'

# Check drone logs in Terminal 2 for task processing
```

---

## 🎯 Option 2: Widescreen Research MCP Server

This is the advanced orchestration server with elicitation and AI capabilities.

### Terminal 1: Start Widescreen Research MCP

```bash
# Load environment variables
source .env.local

# Run the widescreen research MCP server
go run ./cmd/widescreen-research-mcp
```

**Expected Output:**

```
Widescreen Research MCP Server starting...
Orchestrator initialized for project widescreen-researcher
MCP server ready on stdio
```

**What it does:**

- Initializes orchestrator with GCP clients
- Sets up elicitation manager
- Registers MCP tools: `widescreen-research`, `sequential-thinking`, `gcp-provision`, `analyze-findings`
- Communicates via stdio (for MCP clients like Claude Desktop)

### Testing with MCP Inspector

The MCP Inspector is a tool for testing MCP servers interactively:

```bash
# Install MCP Inspector (if not already installed)
npm install -g @modelcontextprotocol/inspector

# Run widescreen-research-mcp with inspector
npx @modelcontextprotocol/inspector go run ./cmd/widescreen-research-mcp
```

This opens a web interface where you can:

1. See available tools
2. Call tools with parameters
3. View responses in real-time

### Testing Operations

**1. Sequential Thinking:**

```json
{
  "operation": "sequential-thinking",
  "parameters": {
    "problem": "How can we optimize drone spawning costs?",
    "max_thoughts": 5
  }
}
```

**2. GCP Provisioning:**

```json
{
  "operation": "gcp-provision",
  "parameters": {
    "resource_type": "cloud_run",
    "count": 1,
    "region": "us-central1",
    "config": {
      "image": "gcr.io/widescreen-researcher/drone:latest",
      "memory": "512Mi",
      "cpu": "1000m"
    }
  }
}
```

**3. Orchestrate Research (with elicitation):**

```json
{
  "operation": "orchestrate-research",
  "query": "Research the latest developments in quantum computing"
}
```

The server will respond with elicitation questions to qualify your research needs.

---

## 🖥️ Option 3: MCP Coordinator (stdio)

This wraps the coordinator in an MCP server for desktop clients.

### Terminal 1: Start MCP Coordinator

```bash
# Load environment variables
source .env.local

# Run MCP coordinator
go run ./cmd/mcp-coordinator
```

**Expected Output:**

```
Starting Spawn MCP Coordinator...
Starting MCP coordinator server on stdio...
```

**What it does:**

- Wraps the coordinator in an MCP server
- Exposes coordinator functions as MCP tools
- Communicates via stdio

### Testing with MCP Inspector

```bash
npx @modelcontextprotocol/inspector go run ./cmd/mcp-coordinator
```

**Available Tools:**

- `spawn_drone` - Spawn a new drone on Cloud Run
- `list_drones` - List active drones
- `execute_task` - Execute a task on a drone
- `get_task_result` - Get task execution results

---

## 🧪 Integration Testing Scenarios

### Scenario 1: End-to-End Research Task

**Goal:** Test the full flow from coordinator to drone and back.

```bash
# Terminal 1: Coordinator
source .env.local && go run ./cmd/coordinator

# Terminal 2: Drone
source .env.local && PORT=8081 go run ./cmd/drone

# Terminal 3: Execute test
# 1. Send task to drone
curl -X POST http://localhost:8081/task \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "machine learning trends 2024",
    "time_frame": "2024",
    "sources": ["academic"],
    "max_results": 3
  }' | jq .

# 2. Check coordinator for results (if drone publishes to Pub/Sub)
curl http://localhost:8080/tasks?id=test-001 | jq .
```

### Scenario 2: Drone Registration Flow

```bash
# Terminal 1: Coordinator
source .env.local && go run ./cmd/coordinator

# Terminal 2: Register a drone
curl -X POST http://localhost:8080/api/drones/register \
  -H "Content-Type: application/json" \
  -d '{
    "drone_id": "test-drone-001",
    "drone_type": "researcher",
    "service_url": "http://localhost:8081",
    "region": "us-central1",
    "capabilities": ["research", "analysis"]
  }' | jq .

# Terminal 3: Verify registration
curl http://localhost:8080/drones | jq .
```

### Scenario 3: Firestore State Management

```bash
# Check Firestore for execution plans
gcloud firestore documents list --collection=execution_plans --database="(default)"

# Check drone registry
gcloud firestore documents list --collection=drone_registry --database="(default)"

# Check task results
gcloud firestore documents list --collection=task_results --database="(default)"
```

---

## 🐛 Troubleshooting

### Issue: "GOOGLE_CLOUD_PROJECT environment variable is required"

**Solution:**

```bash
# Make sure .env.local is loaded
source .env.local

# Or export manually
export GOOGLE_CLOUD_PROJECT=widescreen-researcher
```

### Issue: "Failed to create GCP client: could not find default credentials"

**Solution:**

```bash
# Re-authenticate
gcloud auth application-default login

# Verify credentials exist
ls ~/.config/gcloud/application_default_credentials.json
```

### Issue: "DRONE_ID environment variable is required"

**Solution:**

```bash
# Drone requires these variables
export DRONE_ID=test-drone-001
export PUBSUB_TOPIC=drone-tasks
export GOOGLE_CLOUD_PROJECT=widescreen-researcher

# Or use .env.local
source .env.local
```

### Issue: "Failed to create pubsub client"

**Solution:**

```bash
# Verify Pub/Sub API is enabled
gcloud services list --enabled | grep pubsub

# Verify topic exists
gcloud pubsub topics list | grep drone-tasks

# Create topic if missing
gcloud pubsub topics create drone-tasks
```

### Issue: Drone not receiving tasks

**Solution:**

```bash
# Check Pub/Sub subscription
gcloud pubsub subscriptions list

# Manually publish a test message
gcloud pubsub topics publish drone-tasks \
  --message='{"task_id":"test","drone_id":"test-drone-001"}'

# Check drone logs for message receipt
```

---

## 📊 Monitoring Local Services

### View Logs

```bash
# Coordinator logs (Terminal 1)
# Look for:
# - "Initialized GCP client"
# - "Coordinator HTTP server listening"
# - Drone registration events

# Drone logs (Terminal 2)
# Look for:
# - "Researcher Drone HTTP listening"
# - Task execution logs
# - Pub/Sub message receipt
```

### Check Resource Usage

```bash
# Monitor Go processes
ps aux | grep "go run"

# Monitor network connections
lsof -i :8080  # Coordinator
lsof -i :8081  # Drone
```

### Test Health Endpoints

```bash
# Create a simple health check script
cat > check-health.sh << 'EOF'
#!/bin/bash
echo "Coordinator: $(curl -s http://localhost:8080/health)"
echo "Drone: $(curl -s http://localhost:8081/health)"
EOF

chmod +x check-health.sh
./check-health.sh
```

---

## 🎓 Next Steps

Once local testing is successful:

1. **Build Docker Images** - See `GCP_SETUP_COMPLETE.md` step 2
2. **Deploy to Cloud Run** - See `GCP_SETUP_COMPLETE.md` step 3
3. **Test Production Deployment** - Use Cloud Run URLs instead of localhost
4. **Set Up Monitoring** - Configure Cloud Logging and Monitoring
5. **Load Testing** - Test with multiple concurrent drones

---

## 📚 Additional Resources

- **API Documentation**: See `docs/OPERATIONS.md`
- **GCP Setup**: See `GCP_SETUP_COMPLETE.md`
- **Architecture**: See `project_spec.md`
- **MCP Protocol**: See `cmd/widescreen-research-mcp/README.md`
- **Development Guide**: See `CLAUDE.md`

---

## 💡 Tips

1. **Use separate terminals** for each service to see logs clearly
2. **Enable debug logging** with `LOG_LEVEL=debug` for more details
3. **Test incrementally** - Start with health checks, then simple tasks
4. **Check GCP quotas** if you encounter rate limits
5. **Use MCP Inspector** for interactive testing of MCP servers
6. **Monitor Firestore** to see state changes in real-time

Happy testing! 🚀
