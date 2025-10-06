# Widescreen-Research Quick Start

Fast reference for getting started with local testing.

## ⚡ 30-Second Setup

```bash
# 1. Authenticate
gcloud auth application-default login

# 2. Load environment
source .env.local

# 3. Start services (in separate terminals)
# Terminal 1:
go run ./cmd/coordinator

# Terminal 2:
PORT=8081 go run ./cmd/drone

# Terminal 3:
curl http://localhost:8080/health
curl http://localhost:8081/health
```

## 🎯 Three Testing Options

### Option 1: Basic Coordinator + Drone (HTTP)
```bash
# Terminal 1
source .env.local && go run ./cmd/coordinator

# Terminal 2
source .env.local && PORT=8081 go run ./cmd/drone

# Terminal 3 - Test
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8080/drones
```

### Option 2: Widescreen Research MCP (Advanced)
```bash
# Terminal 1
source .env.local && go run ./cmd/widescreen-research-mcp

# Terminal 2 - Test with MCP Inspector
npx @modelcontextprotocol/inspector go run ./cmd/widescreen-research-mcp
```

### Option 3: MCP Coordinator (stdio)
```bash
# Terminal 1
source .env.local && go run ./cmd/mcp-coordinator

# Terminal 2 - Test with MCP Inspector
npx @modelcontextprotocol/inspector go run ./cmd/mcp-coordinator
```

## 🔍 Quick Tests

### Health Checks
```bash
curl http://localhost:8080/health  # Coordinator
curl http://localhost:8081/health  # Drone
```

### Coordinator Info
```bash
curl http://localhost:8080/ | jq .
```

### List Drones
```bash
curl http://localhost:8080/drones | jq .
```

### Execute Research Task
```bash
curl -X POST http://localhost:8081/task \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "AI trends 2024",
    "time_frame": "2024",
    "sources": ["academic"],
    "max_results": 5
  }' | jq .
```

### Register Drone
```bash
curl -X POST http://localhost:8080/api/drones/register \
  -H "Content-Type: application/json" \
  -d '{
    "drone_id": "test-drone-001",
    "drone_type": "researcher",
    "service_url": "http://localhost:8081",
    "region": "us-central1",
    "capabilities": ["research"]
  }' | jq .
```

## 🐛 Common Issues

### "GOOGLE_CLOUD_PROJECT required"
```bash
source .env.local
# or
export GOOGLE_CLOUD_PROJECT=widescreen-researcher
```

### "Could not find default credentials"
```bash
gcloud auth application-default login
```

### "DRONE_ID required"
```bash
source .env.local
# or
export DRONE_ID=test-drone-001
export PUBSUB_TOPIC=drone-tasks
```

### Port already in use
```bash
# Use different port
PORT=8082 go run ./cmd/drone
```

## 📊 Monitoring

### View Logs
```bash
# Coordinator and drone output in their terminals
# Look for:
# - "Initialized GCP client"
# - "HTTP server listening"
# - Task execution logs
```

### Check GCP Resources
```bash
# Firestore
gcloud firestore documents list --collection=execution_plans --database="(default)"

# Pub/Sub
gcloud pubsub topics list | grep drone

# Cloud Run (if deployed)
gcloud run services list --region=us-central1
```

## 🚀 Next Steps

1. ✅ Local testing working? → Build Docker images
2. ✅ Docker images built? → Deploy to Cloud Run
3. ✅ Deployed to Cloud Run? → Test production endpoints
4. ✅ Production working? → Set up monitoring & alerts

## 📚 Full Documentation

- **Detailed Testing**: `docs/LOCAL_TESTING_GUIDE.md`
- **GCP Setup**: `GCP_SETUP_COMPLETE.md`
- **Operations**: `docs/OPERATIONS.md`
- **Architecture**: `project_spec.md`

## 💡 Pro Tips

- Use `LOG_LEVEL=debug` for verbose output
- Test health endpoints first
- Use MCP Inspector for interactive testing
- Monitor Firestore for state changes
- Check Pub/Sub for message flow

---

**Ready to test?** Start with Option 1 (Basic Coordinator + Drone) and work your way up!

