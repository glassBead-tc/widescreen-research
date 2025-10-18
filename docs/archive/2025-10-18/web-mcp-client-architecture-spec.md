# Web-Based MCP Client Architecture Specification

**Version:** 1.0
**Date:** September 29, 2025
**Status:** Draft
**Target Platform:** Google Cloud Platform

## Executive Summary

This specification describes the architecture for implementing a browser-based MCP client that communicates with our existing Go MCP server infrastructure. The browser becomes a full MCP host—similar to Claude Desktop or Cursor—enabling universal access to Widescreen Research capabilities without desktop application installation.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [System Components](#system-components)
3. [GCP Infrastructure](#gcp-infrastructure)
4. [Communication Flow](#communication-flow)
5. [Implementation Details](#implementation-details)
6. [Security Architecture](#security-architecture)
7. [State of Codebase After Implementation](#state-of-codebase-after-implementation)
8. [Migration Path](#migration-path)
9. [Performance Considerations](#performance-considerations)
10. [Monitoring and Observability](#monitoring-and-observability)

## Architecture Overview

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    Google Cloud Platform                      │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Cloud CDN / Cloud Storage (Frontend Assets)           │  │
│  │  - Next.js static export                               │  │
│  │  - @mcp-ui/client SDK                                  │  │
│  │  - MCP protocol implementation                         │  │
│  └───────────────────────┬────────────────────────────────┘  │
│                          │                                    │
│                          │ HTTPS/WSS                          │
│                          │ (Internal)                         │
│                          ▼                                    │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Cloud Run (MCP HTTP Server)                           │  │
│  │  - HTTP/SSE transport layer                            │  │
│  │  - Authentication & authorization                      │  │
│  │  - Session management                                  │  │
│  │  - Tool registry & discovery                           │  │
│  │  - UI resource serving                                 │  │
│  └───────────────────────┬────────────────────────────────┘  │
│                          │                                    │
│                          │ gRPC / Internal                    │
│                          │                                    │
│                          ▼                                    │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Cloud Run (Campaign Coordinator)                      │  │
│  │  - Campaign orchestration                              │  │
│  │  - Drone lifecycle management                          │  │
│  │  - Results aggregation                                 │  │
│  └───────────────────────┬────────────────────────────────┘  │
│                          │                                    │
│                          │ Pub/Sub                            │
│                          │                                    │
│                          ▼                                    │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Cloud Run Jobs (Research Drones)                      │  │
│  │  - Parallel task execution                             │  │
│  │  - LLM API calls                                       │  │
│  │  - Results processing                                  │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                               │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Supporting Services                                   │  │
│  │  - Firestore: Session state, campaigns, results       │  │
│  │  - Cloud Storage: Datasets, artifacts, exports        │  │
│  │  - Pub/Sub: Event streaming, progress updates         │  │
│  │  - Secret Manager: API keys, credentials              │  │
│  │  - Cloud IAM: Service-to-service auth                 │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

### User Experience Flow

```
User Browser ──HTTPS──> Cloud CDN ──serves──> React App
     │
     │ (App loads MCP client SDK)
     │
     ▼
MCP Client Init ──WSS/HTTPS──> Cloud Run (MCP Server)
     │
     │ listTools()
     │ ────────────────────────────>
     │                              Discover available tools
     │ <────────────────────────────
     │ { tools: [...] }
     │
     │ callTool('start_campaign')
     │ ────────────────────────────>
     │                              Validate & orchestrate
     │                              ──gRPC──> Coordinator
     │                                       ──Pub/Sub──> Drones
     │
     │ <────────────────────────────
     │ { content: [{ type: 'resource', uri: 'ui://campaign/...' }] }
     │
     │ readResource('ui://campaign/...')
     │ ────────────────────────────>
     │                              Fetch UI component
     │ <────────────────────────────
     │ { contents: [{ mimeType: 'text/html', text: '...' }] }
     │
     │ (Render interactive dashboard)
     │
     │ SSE: progress updates ──────>
     │                              Campaign events
     │ <────────────────────────────
     │ { progress: 67%, findings: [...] }
```

## System Components

### 1. Frontend (Browser MCP Client)

**Technology Stack:**

- **Framework:** Next.js 14+ (App Router)
- **MCP SDK:** `@modelcontextprotocol/sdk` (TypeScript client)
- **UI SDK:** `@mcp-ui/client`
- **State Management:** React Query + Zustand
- **Styling:** Tailwind CSS
- **Deployment:** Cloud Storage + Cloud CDN

**Key Responsibilities:**

- MCP protocol client implementation
- Tool discovery and invocation
- UI resource rendering
- Real-time progress updates via SSE
- User session management
- Error handling and retry logic

**Directory Structure (New):**

```
web-client/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── campaigns/
│   │   │   ├── [id]/
│   │   │   │   └── page.tsx   # Campaign detail view
│   │   │   └── page.tsx       # Campaigns list
│   │   ├── layout.tsx
│   │   └── page.tsx           # Home
│   ├── components/
│   │   ├── mcp/
│   │   │   ├── MCPProvider.tsx        # MCP context
│   │   │   ├── ToolInvoker.tsx        # Tool calling UI
│   │   │   ├── ResourceRenderer.tsx   # UI resource renderer
│   │   │   └── ProgressStream.tsx     # SSE handler
│   │   ├── campaigns/
│   │   │   ├── CampaignCard.tsx
│   │   │   ├── CampaignDashboard.tsx
│   │   │   └── CreateCampaignForm.tsx
│   │   └── ui/                # Shadcn components
│   ├── lib/
│   │   ├── mcp/
│   │   │   ├── client.ts      # MCP client wrapper
│   │   │   ├── transport.ts   # HTTP/SSE transport
│   │   │   └── types.ts       # MCP types
│   │   └── api/
│   │       └── auth.ts        # Auth helpers
│   └── hooks/
│       ├── useMCP.ts          # MCP client hook
│       ├── useCampaign.ts     # Campaign operations
│       └── useTools.ts        # Tool discovery
├── public/
├── package.json
├── next.config.js
├── tailwind.config.js
└── tsconfig.json
```

### 2. MCP HTTP Server (New)

**Location:** `cmd/mcp-http-server/`

**Technology Stack:**

- **Language:** Go 1.23+
- **Framework:** Chi router
- **MCP SDK:** Custom implementation (adapting existing `pkg/mcp/server.go`)
- **Transport:** HTTP/SSE
- **Deployment:** Cloud Run

**Key Responsibilities:**

- HTTP/SSE transport for MCP protocol
- Authentication and authorization
- Session management (Firestore)
- Tool registry and discovery
- Resource serving (UI components)
- Request routing to Coordinator
- Rate limiting and quota management

**Directory Structure (New):**

```
cmd/mcp-http-server/
├── main.go
├── handlers/
│   ├── mcp.go              # MCP protocol handlers
│   ├── auth.go             # Authentication middleware
│   ├── session.go          # Session management
│   └── sse.go              # Server-Sent Events
├── transport/
│   ├── http.go             # HTTP transport implementation
│   └── sse.go              # SSE stream management
└── config/
    └── config.go           # Server configuration

pkg/web/
├── auth/
│   ├── middleware.go       # Auth middleware
│   ├── session.go          # Session handling
│   └── tokens.go           # JWT/token management
├── mcp_http/
│   ├── server.go           # HTTP MCP server
│   ├── transport.go        # Transport layer
│   └── handlers.go         # Request handlers
└── ui/
    ├── resources.go        # UI resource management
    └── templates/          # Dashboard templates
        ├── campaign.html
        ├── findings.html
        └── status.html
```

### 3. Campaign Coordinator (Enhanced)

**Location:** `cmd/coordinator/` (existing, enhanced)

**Enhancements:**

- Add gRPC server interface for MCP HTTP server
- Implement UI resource generation
- Add SSE event streaming
- Enhanced progress tracking

**New Files:**

```
cmd/coordinator/
├── grpc_server.go          # gRPC interface for HTTP server
└── ui_generator.go         # Generate UI resources

pkg/coordinator/
├── ui_resources.go         # UI resource definitions
└── events.go               # Event streaming
```

### 4. Research Drones (Existing)

**Location:** `cmd/drone/` (minimal changes)

**Changes:**

- Enhanced progress reporting (Pub/Sub)
- Structured output for UI rendering

## GCP Infrastructure

### Resource Topology

```
project: widescreen-research
├── Cloud Run Services
│   ├── mcp-http-server
│   │   ├── CPU: 1
│   │   ├── Memory: 512Mi
│   │   ├── Min instances: 1
│   │   ├── Max instances: 10
│   │   └── Ingress: Internal + Cloud Load Balancing
│   ├── coordinator
│   │   ├── CPU: 2
│   │   ├── Memory: 1Gi
│   │   ├── Min instances: 1
│   │   ├── Max instances: 5
│   │   └── Ingress: Internal only
│   └── drone (template)
│       ├── CPU: 1
│       ├── Memory: 512Mi
│       └── Execution: Jobs
├── Cloud Storage Buckets
│   ├── widescreen-web-client (CDN-enabled)
│   │   └── Frontend static assets
│   ├── widescreen-datasets
│   │   └── Input JSONL files
│   └── widescreen-results
│       └── Campaign outputs
├── Firestore Database
│   ├── sessions/           # MCP sessions
│   ├── campaigns/          # Campaign metadata
│   ├── findings/           # Research results
│   └── users/              # User profiles
├── Pub/Sub Topics
│   ├── drone-tasks         # Task distribution
│   ├── drone-results       # Result collection
│   └── campaign-events     # Progress updates
├── Secret Manager
│   ├── anthropic-api-key
│   ├── openai-api-key
│   └── jwt-signing-key
├── Cloud Load Balancing
│   ├── HTTPS Load Balancer
│   │   ├── Frontend: Cloud Storage backend
│   │   └── API: Cloud Run backend
│   └── SSL Certificate (managed)
└── Cloud IAM
    ├── Service Accounts
    │   ├── mcp-http-server@
    │   ├── coordinator@
    │   └── drone@
    └── IAM Bindings
```

### Network Architecture

```
Internet
    │
    │ HTTPS (443)
    ▼
Cloud Load Balancer
    │
    ├─> /                    ──> Cloud Storage (Frontend)
    │   /*.js, /*.css, etc.
    │
    └─> /api/*               ──> Cloud Run (mcp-http-server)
        ├─> /api/mcp         MCP protocol endpoint
        ├─> /api/sse         SSE event stream
        └─> /api/resources   UI resources

Internal (VPC)
    │
    mcp-http-server ──gRPC──> coordinator
                                    │
                                    │ Pub/Sub
                                    ▼
                              [ Drone Jobs ]
```

### IAM Configuration

```yaml
# Service Account: mcp-http-server@
roles:
  - roles/run.invoker              # Invoke coordinator
  - roles/datastore.user           # Firestore access
  - roles/pubsub.publisher         # Publish events
  - roles/storage.objectViewer     # Read datasets
  - roles/secretmanager.secretAccessor

# Service Account: coordinator@
roles:
  - roles/run.admin                # Manage drone jobs
  - roles/datastore.user           # Firestore access
  - roles/pubsub.subscriber        # Subscribe to results
  - roles/pubsub.publisher         # Publish tasks
  - roles/storage.objectCreator    # Write results

# Service Account: drone@
roles:
  - roles/datastore.user           # Write findings
  - roles/pubsub.publisher         # Publish results
  - roles/storage.objectViewer     # Read datasets
  - roles/secretmanager.secretAccessor  # LLM API keys
```

## Communication Flow

### 1. Session Establishment

```go
// pkg/web/auth/session.go
type Session struct {
    ID        string    `firestore:"id"`
    UserID    string    `firestore:"user_id"`
    CreatedAt time.Time `firestore:"created_at"`
    ExpiresAt time.Time `firestore:"expires_at"`
    MCPState  MCPState  `firestore:"mcp_state"`
}

type MCPState struct {
    ConnectedServers []string          `firestore:"connected_servers"`
    AvailableTools   []mcp.Tool        `firestore:"available_tools"`
    ActiveCampaigns  []string          `firestore:"active_campaigns"`
    Metadata         map[string]string `firestore:"metadata"`
}

func (s *SessionManager) Create(ctx context.Context, userID string) (*Session, error) {
    session := &Session{
        ID:        uuid.New().String(),
        UserID:    userID,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour),
        MCPState:  MCPState{ConnectedServers: []string{"widescreen-coordinator"}},
    }

    _, err := s.firestore.Collection("sessions").Doc(session.ID).Set(ctx, session)
    return session, err
}
```

### 2. Tool Discovery

```typescript
// web-client/src/lib/mcp/client.ts
export class MCPClient {
  async listTools(): Promise<ListToolsResponse> {
    const response = await this.transport.request({
      jsonrpc: '2.0',
      id: this.generateId(),
      method: 'tools/list',
      params: {}
    });

    return response.result;
  }
}
```

```go
// cmd/mcp-http-server/handlers/mcp.go
func (h *MCPHandler) handleToolsList(w http.ResponseWriter, r *http.Request) {
    session := r.Context().Value("session").(*Session)

    tools := []mcp.Tool{
        {
            Name:        "start_campaign",
            Description: "Start a new research campaign",
            InputSchema: jsonschema.Object{
                Properties: map[string]jsonschema.Schema{
                    "dataset":     jsonschema.String{},
                    "parallelism": jsonschema.Integer{},
                    "llm_config":  jsonschema.Object{},
                },
                Required: []string{"dataset"},
            },
        },
        {
            Name:        "get_campaign_status",
            Description: "Get status of a running campaign",
            InputSchema: jsonschema.Object{
                Properties: map[string]jsonschema.Schema{
                    "campaign_id": jsonschema.String{},
                },
                Required: []string{"campaign_id"},
            },
        },
        // ... more tools
    }

    // Filter tools based on user permissions
    filtered := h.filterToolsByPermissions(tools, session.UserID)

    json.NewEncoder(w).Encode(map[string]interface{}{
        "tools": filtered,
    })
}
```

### 3. Tool Invocation

```typescript
// web-client/src/hooks/useCampaign.ts
export function useCampaign() {
  const { client } = useMCP();

  async function startCampaign(config: CampaignConfig) {
    const result = await client.callTool({
      name: 'start_campaign',
      arguments: {
        dataset: config.datasetUri,
        parallelism: config.parallelism,
        llm_config: config.llmConfig
      }
    });

    // Extract UI resource from response
    const uiResource = result.content.find(
      c => c.type === 'resource' && c.uri.startsWith('ui://campaign/')
    );

    // Extract campaign ID
    const campaignId = result.content.find(
      c => c.type === 'text'
    )?.text.match(/Campaign ID: (\S+)/)?.[1];

    return { campaignId, uiResource };
  }

  return { startCampaign };
}
```

```go
// cmd/mcp-http-server/handlers/mcp.go
func (h *MCPHandler) handleCallTool(w http.ResponseWriter, r *http.Request) {
    session := r.Context().Value("session").(*Session)

    var req mcp.CallToolRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Route to appropriate handler
    switch req.Params.Name {
    case "start_campaign":
        h.handleStartCampaign(w, r, session, req.Params.Arguments)
    case "get_campaign_status":
        h.handleGetCampaignStatus(w, r, session, req.Params.Arguments)
    // ... more tools
    default:
        http.Error(w, "unknown tool", http.StatusNotFound)
    }
}

func (h *MCPHandler) handleStartCampaign(
    w http.ResponseWriter,
    r *http.Request,
    session *Session,
    args map[string]interface{},
) {
    // Call coordinator via gRPC
    campaignID, err := h.coordinator.StartCampaign(r.Context(), &pb.StartCampaignRequest{
        UserId:      session.UserID,
        Dataset:     args["dataset"].(string),
        Parallelism: int32(args["parallelism"].(float64)),
    })

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Generate UI resource URI
    uiResourceURI := fmt.Sprintf("ui://campaign/%s", campaignID)

    // Return MCP response
    response := mcp.CallToolResponse{
        Content: []mcp.Content{
            {
                Type: "text",
                Text: fmt.Sprintf("Campaign %s started successfully", campaignID),
            },
            {
                Type: "resource",
                Resource: &mcp.ResourceReference{
                    URI:      uiResourceURI,
                    MimeType: "text/html",
                },
            },
        },
    }

    json.NewEncoder(w).Encode(response)
}
```

### 4. UI Resource Rendering

```typescript
// web-client/src/components/mcp/ResourceRenderer.tsx
import { UIResourceRenderer } from '@mcp-ui/client';

export function MCPResourceRenderer({ resourceUri }: { resourceUri: string }) {
  const { client } = useMCP();
  const [resource, setResource] = useState<Resource | null>(null);

  useEffect(() => {
    client.readResource({ uri: resourceUri }).then(setResource);
  }, [resourceUri]);

  if (!resource) return <LoadingSpinner />;

  return (
    <UIResourceRenderer
      resource={resource}
      onUIAction={async (action) => {
        if (action.type === 'tool') {
          // User interaction triggered tool call
          await client.callTool({
            name: action.payload.toolName,
            arguments: action.payload.params
          });
        }
      }}
    />
  );
}
```

```go
// pkg/web/ui/resources.go
func (g *UIGenerator) GenerateCampaignDashboard(
    ctx context.Context,
    campaignID string,
) (string, error) {
    // Fetch campaign data
    campaign, err := g.firestore.Collection("campaigns").Doc(campaignID).Get(ctx)
    if err != nil {
        return "", err
    }

    // Generate HTML with embedded actions
    tmpl := template.Must(template.ParseFiles("templates/campaign.html"))

    var buf bytes.Buffer
    err = tmpl.Execute(&buf, map[string]interface{}{
        "CampaignID": campaignID,
        "Status":     campaign.Data()["status"],
        "Progress":   campaign.Data()["progress"],
        "Findings":   campaign.Data()["findings_count"],
        "Actions": []UIAction{
            {
                Type:  "tool",
                Label: "Pause Campaign",
                Tool:  "pause_campaign",
                Params: map[string]interface{}{
                    "campaign_id": campaignID,
                },
            },
            {
                Type:  "tool",
                Label: "Export Results",
                Tool:  "export_results",
                Params: map[string]interface{}{
                    "campaign_id": campaignID,
                    "format":      "jsonl",
                },
            },
        },
    })

    return buf.String(), err
}
```

### 5. Real-Time Updates (SSE)

```typescript
// web-client/src/components/mcp/ProgressStream.tsx
export function CampaignProgressStream({ campaignId }: { campaignId: string }) {
  const [progress, setProgress] = useState<Progress | null>(null);

  useEffect(() => {
    const eventSource = new EventSource(
      `/api/sse/campaigns/${campaignId}`
    );

    eventSource.onmessage = (event) => {
      const update = JSON.parse(event.data);
      setProgress(update);
    };

    return () => eventSource.close();
  }, [campaignId]);

  if (!progress) return null;

  return (
    <div className="progress-bar">
      <div
        className="progress-fill"
        style={{ width: `${progress.percentage}%` }}
      />
      <span>{progress.completed} / {progress.total} tasks</span>
    </div>
  );
}
```

```go
// cmd/mcp-http-server/handlers/sse.go
func (h *SSEHandler) StreamCampaignProgress(w http.ResponseWriter, r *http.Request) {
    campaignID := chi.URLParam(r, "id")

    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "SSE not supported", http.StatusInternalServerError)
        return
    }

    // Subscribe to Pub/Sub campaign events
    sub := h.pubsub.Subscription(fmt.Sprintf("campaign-%s-events", campaignID))

    ctx := r.Context()
    err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
        // Parse event
        var event CampaignEvent
        json.Unmarshal(msg.Data, &event)

        // Send to client
        fmt.Fprintf(w, "data: %s\n\n", string(msg.Data))
        flusher.Flush()

        msg.Ack()
    })

    if err != nil {
        log.Printf("SSE error: %v", err)
    }
}
```

## Implementation Details

### Phase 1: MCP HTTP Server Foundation

**Duration:** 2 weeks

**Deliverables:**

1. HTTP transport layer for MCP protocol
2. Session management with Firestore
3. Basic authentication middleware
4. Tool registry and discovery
5. Integration with existing coordinator

**Files Created:**

```
cmd/mcp-http-server/
  main.go
  handlers/mcp.go
  handlers/auth.go
  transport/http.go

pkg/web/
  mcp_http/server.go
  auth/middleware.go
  auth/session.go
```

**Testing:**

```bash
# Manual testing with curl
curl -X POST http://localhost:8080/api/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

### Phase 2: Frontend Client Implementation

**Duration:** 3 weeks

**Deliverables:**

1. Next.js application structure
2. MCP client SDK integration
3. Tool invocation UI
4. Campaign dashboard
5. Resource renderer

**Files Created:**

```
web-client/
  src/app/...
  src/components/...
  src/lib/mcp/...
  src/hooks/...
```

**Testing:**

```bash
cd web-client
npm run dev
# Manual testing in browser
```

### Phase 3: UI Resources & SSE

**Duration:** 2 weeks

**Deliverables:**

1. UI resource generation
2. HTML templates for dashboards
3. SSE event streaming
4. Real-time progress updates

**Files Created:**

```
pkg/web/ui/
  resources.go
  templates/campaign.html
  templates/findings.html

cmd/mcp-http-server/handlers/sse.go
```

### Phase 4: GCP Deployment

**Duration:** 1 week

**Deliverables:**

1. Cloud Run services
2. Cloud Storage + CDN setup
3. Load balancer configuration
4. IAM and networking
5. Monitoring setup

**Files Created:**

```
deploy/
  terraform/
    main.tf
    cloud-run.tf
    storage.tf
    iam.tf
  cloudbuild.yaml
```

### Phase 5: Testing & Optimization

**Duration:** 2 weeks

**Deliverables:**

1. Integration tests
2. Load testing
3. Performance optimization
4. Documentation

## Security Architecture

### Authentication Flow

```
Browser ──1. Login──> Identity Platform
                            │
                            │ 2. ID Token
                            ▼
Browser ──3. Token──> MCP HTTP Server
                            │
                            │ 4. Verify token
                            │ 5. Create session
                            ▼
                      Firestore (sessions)
```

### Implementation

```go
// pkg/web/auth/middleware.go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token
        token := extractBearerToken(r)
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Verify with Identity Platform
        idToken, err := firebaseAuth.VerifyIDToken(r.Context(), token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Load or create session
        session, err := sessionManager.GetOrCreate(r.Context(), idToken.UID)
        if err != nil {
            http.Error(w, "Session error", http.StatusInternalServerError)
            return
        }

        // Add to context
        ctx := context.WithValue(r.Context(), "session", session)
        ctx = context.WithValue(ctx, "user_id", idToken.UID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Authorization Rules

```go
// pkg/web/auth/authz.go
type Permission string

const (
    PermissionStartCampaign Permission = "campaign:start"
    PermissionViewCampaign  Permission = "campaign:view"
    PermissionPauseCampaign Permission = "campaign:pause"
    PermissionExportResults Permission = "results:export"
)

func (a *Authorizer) CheckPermission(
    ctx context.Context,
    userID string,
    permission Permission,
    resourceID string,
) error {
    // Check if user owns resource
    if permission == PermissionViewCampaign ||
       permission == PermissionPauseCampaign {
        campaign, err := a.getCampaign(ctx, resourceID)
        if err != nil {
            return err
        }
        if campaign.UserID != userID {
            return ErrForbidden
        }
    }

    // Check quota limits
    if permission == PermissionStartCampaign {
        count, err := a.getActiveCampaignCount(ctx, userID)
        if err != nil {
            return err
        }
        if count >= maxConcurrentCampaigns {
            return ErrQuotaExceeded
        }
    }

    return nil
}
```

### Rate Limiting

```go
// pkg/web/ratelimit/limiter.go
type RateLimiter struct {
    store  *redis.Client
    limits map[string]Limit
}

type Limit struct {
    Requests int
    Window   time.Duration
}

func (rl *RateLimiter) Allow(ctx context.Context, userID, operation string) bool {
    limit := rl.limits[operation]
    key := fmt.Sprintf("ratelimit:%s:%s", userID, operation)

    // Token bucket algorithm
    val, err := rl.store.Get(ctx, key).Int()
    if err == redis.Nil {
        // First request
        rl.store.Set(ctx, key, limit.Requests-1, limit.Window)
        return true
    }

    if val <= 0 {
        return false
    }

    rl.store.Decr(ctx, key)
    return true
}

// Default limits
var DefaultLimits = map[string]Limit{
    "start_campaign":       {Requests: 10, Window: time.Hour},
    "get_campaign_status":  {Requests: 100, Window: time.Minute},
    "export_results":       {Requests: 5, Window: time.Hour},
}
```

## State of Codebase After Implementation

### New Directory Structure

```
/Users/b.c.nims/widescreen-research/
├── cmd/
│   ├── mcp-http-server/               # NEW: Web MCP server
│   │   ├── main.go
│   │   ├── handlers/
│   │   │   ├── mcp.go
│   │   │   ├── auth.go
│   │   │   ├── session.go
│   │   │   └── sse.go
│   │   ├── transport/
│   │   │   ├── http.go
│   │   │   └── sse.go
│   │   └── config/
│   │       └── config.go
│   ├── coordinator/                   # ENHANCED
│   │   ├── main.go
│   │   ├── grpc_server.go            # NEW: gRPC interface
│   │   └── ui_generator.go           # NEW: UI resources
│   └── drone/                         # MINIMAL CHANGES
│       └── main.go
├── pkg/
│   ├── web/                           # NEW: Web-specific packages
│   │   ├── mcp_http/
│   │   │   ├── server.go
│   │   │   ├── transport.go
│   │   │   └── handlers.go
│   │   ├── auth/
│   │   │   ├── middleware.go
│   │   │   ├── session.go
│   │   │   ├── authz.go
│   │   │   └── tokens.go
│   │   ├── ui/
│   │   │   ├── resources.go
│   │   │   └── templates/
│   │   │       ├── campaign.html
│   │   │       ├── findings.html
│   │   │       └── status.html
│   │   └── ratelimit/
│   │       └── limiter.go
│   ├── coordinator/                   # ENHANCED
│   │   ├── server.go
│   │   ├── ui_resources.go           # NEW
│   │   └── events.go                 # NEW
│   └── mcp/                           # EXISTING
│       └── server.go
├── web-client/                        # NEW: Frontend application
│   ├── src/
│   │   ├── app/
│   │   │   ├── campaigns/
│   │   │   │   ├── [id]/
│   │   │   │   │   └── page.tsx
│   │   │   │   └── page.tsx
│   │   │   ├── layout.tsx
│   │   │   └── page.tsx
│   │   ├── components/
│   │   │   ├── mcp/
│   │   │   │   ├── MCPProvider.tsx
│   │   │   │   ├── ToolInvoker.tsx
│   │   │   │   ├── ResourceRenderer.tsx
│   │   │   │   └── ProgressStream.tsx
│   │   │   ├── campaigns/
│   │   │   │   ├── CampaignCard.tsx
│   │   │   │   ├── CampaignDashboard.tsx
│   │   │   │   └── CreateCampaignForm.tsx
│   │   │   └── ui/
│   │   ├── lib/
│   │   │   ├── mcp/
│   │   │   │   ├── client.ts
│   │   │   │   ├── transport.ts
│   │   │   │   └── types.ts
│   │   │   └── api/
│   │   │       └── auth.ts
│   │   └── hooks/
│   │       ├── useMCP.ts
│   │       ├── useCampaign.ts
│   │       └── useTools.ts
│   ├── public/
│   ├── package.json
│   ├── next.config.js
│   ├── tailwind.config.js
│   └── tsconfig.json
├── deploy/                            # NEW: Deployment configs
│   ├── terraform/
│   │   ├── main.tf
│   │   ├── cloud-run.tf
│   │   ├── storage.tf
│   │   ├── firestore.tf
│   │   ├── pubsub.tf
│   │   ├── iam.tf
│   │   └── variables.tf
│   ├── cloudbuild/
│   │   ├── mcp-http-server.yaml
│   │   ├── coordinator.yaml
│   │   ├── drone.yaml
│   │   └── web-client.yaml
│   └── kubernetes/                    # Alternative to Cloud Run
│       └── ...
└── docs/
    ├── web-mcp-client-architecture-spec.md  # THIS DOCUMENT
    └── web-client-deployment-guide.md       # NEW: Deployment guide
```

### Modified Files

```
pkg/coordinator/server.go              # Add gRPC server
pkg/coordinator/campaign.go            # Enhanced events
cmd/drone/main.go                      # Enhanced reporting
go.mod                                 # New dependencies
Makefile                               # New build targets
```

### New Dependencies

**Go (`go.mod`):**

```go
require (
    // Existing...

    // New for web server
    github.com/go-chi/chi/v5 v5.0.0
    github.com/go-chi/cors v1.2.1
    google.golang.org/grpc v1.60.0
    google.golang.org/protobuf v1.31.0

    // Firebase Auth
    firebase.google.com/go/v4 v4.13.0

    // Rate limiting (optional, can use Firestore)
    github.com/redis/go-redis/v9 v9.3.0
)
```

**TypeScript (`web-client/package.json`):**

```json
{
  "dependencies": {
    "next": "^14.0.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "@modelcontextprotocol/sdk": "^1.0.0",
    "@mcp-ui/client": "^0.1.0",
    "@tanstack/react-query": "^5.0.0",
    "zustand": "^4.4.0",
    "tailwindcss": "^3.4.0",
    "lucide-react": "^0.300.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "@types/react": "^18.2.0",
    "typescript": "^5.3.0",
    "eslint": "^8.0.0",
    "eslint-config-next": "^14.0.0"
  }
}
```

### Build Artifacts

```
# Go binaries
cmd/mcp-http-server/mcp-http-server
cmd/coordinator/coordinator
cmd/drone/drone

# Docker images (pushed to Artifact Registry)
gcr.io/widescreen-research/mcp-http-server:latest
gcr.io/widescreen-research/coordinator:latest
gcr.io/widescreen-research/drone:latest

# Frontend build
web-client/.next/
web-client/out/        # Static export for Cloud Storage
```

### Configuration Files

**`deploy/terraform/main.tf`:**

```hcl
terraform {
  required_version = ">= 1.6"

  backend "gcs" {
    bucket = "widescreen-terraform-state"
    prefix = "prod"
  }

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

variable "project_id" {
  default = "widescreen-research"
}

variable "region" {
  default = "us-central1"
}
```

**`deploy/terraform/cloud-run.tf`:**

```hcl
resource "google_cloud_run_v2_service" "mcp_http_server" {
  name     = "mcp-http-server"
  location = var.region

  template {
    containers {
      image = "gcr.io/${var.project_id}/mcp-http-server:latest"

      env {
        name  = "PROJECT_ID"
        value = var.project_id
      }

      env {
        name  = "COORDINATOR_URL"
        value = google_cloud_run_v2_service.coordinator.uri
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }
    }

    scaling {
      min_instance_count = 1
      max_instance_count = 10
    }
  }

  traffic {
    percent = 100
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
  }
}

resource "google_cloud_run_v2_service" "coordinator" {
  name     = "coordinator"
  location = var.region

  template {
    containers {
      image = "gcr.io/${var.project_id}/coordinator:latest"

      resources {
        limits = {
          cpu    = "2"
          memory = "1Gi"
        }
      }
    }

    scaling {
      min_instance_count = 1
      max_instance_count = 5
    }
  }
}

# IAM binding for internal communication
resource "google_cloud_run_service_iam_member" "coordinator_invoker" {
  service  = google_cloud_run_v2_service.coordinator.name
  location = google_cloud_run_v2_service.coordinator.location
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.mcp_http_server.email}"
}
```

**`deploy/cloudbuild/mcp-http-server.yaml`:**

```yaml
steps:
  # Build
  - name: 'gcr.io/cloud-builders/docker'
    args:
      - 'build'
      - '-t'
      - 'gcr.io/$PROJECT_ID/mcp-http-server:$COMMIT_SHA'
      - '-t'
      - 'gcr.io/$PROJECT_ID/mcp-http-server:latest'
      - '-f'
      - 'cmd/mcp-http-server/Dockerfile'
      - '.'

  # Push
  - name: 'gcr.io/cloud-builders/docker'
    args:
      - 'push'
      - '--all-tags'
      - 'gcr.io/$PROJECT_ID/mcp-http-server'

  # Deploy to Cloud Run
  - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
    entrypoint: 'gcloud'
    args:
      - 'run'
      - 'deploy'
      - 'mcp-http-server'
      - '--image=gcr.io/$PROJECT_ID/mcp-http-server:$COMMIT_SHA'
      - '--region=us-central1'
      - '--platform=managed'

images:
  - 'gcr.io/$PROJECT_ID/mcp-http-server:$COMMIT_SHA'
  - 'gcr.io/$PROJECT_ID/mcp-http-server:latest'
```

**`web-client/next.config.js`:**

```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',  // For Docker
  // OR
  // output: 'export',   // For Cloud Storage static hosting

  env: {
    NEXT_PUBLIC_MCP_API_URL: process.env.MCP_API_URL || 'http://localhost:8080/api',
  },

  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: process.env.MCP_API_URL + '/:path*',
      },
    ];
  },
};

module.exports = nextConfig;
```

## Migration Path

### Phase 1: Foundation (Weeks 1-2)

**Goals:**

- MCP HTTP server running locally
- Basic tool discovery working
- Session management implemented

**Steps:**

1. Create `cmd/mcp-http-server/` directory structure
2. Implement HTTP transport layer
3. Create session management with in-memory store (Firestore later)
4. Implement tool registry
5. Write integration tests

**Validation:**

```bash
# Start server
cd cmd/mcp-http-server
go run main.go

# Test with curl
curl -X POST http://localhost:8080/api/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list",
    "params": {}
  }'

# Expected response
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "tools": [
      {"name": "start_campaign", ...},
      {"name": "get_campaign_status", ...}
    ]
  }
}
```

### Phase 2: Frontend Client (Weeks 3-5)

**Goals:**

- Next.js app scaffold
- MCP client working
- Basic campaign UI

**Steps:**

1. Initialize Next.js project
2. Install MCP SDK
3. Create MCP client wrapper
4. Build campaign list view
5. Build campaign detail view
6. Implement tool invocation

**Validation:**

```bash
cd web-client
npm run dev

# Open http://localhost:3000
# Should see:
# - List of tools from server
# - Button to start campaign
# - Campaign list (empty initially)
```

### Phase 3: UI Resources (Weeks 6-7)

**Goals:**

- Dynamic dashboards working
- Real-time updates via SSE

**Steps:**

1. Create UI resource templates
2. Implement resource generation in coordinator
3. Add SSE endpoint
4. Connect frontend to SSE
5. Test end-to-end flow

**Validation:**

```bash
# Start campaign from UI
# Should see:
# - Dashboard appears
# - Progress bar updates in real-time
# - Findings appear as they're generated
```

### Phase 4: GCP Deployment (Week 8)

**Goals:**

- All services running on GCP
- Frontend on Cloud Storage + CDN

**Steps:**

1. Write Terraform configs
2. Create service accounts
3. Deploy Cloud Run services
4. Set up Cloud Storage + CDN
5. Configure load balancer
6. Test production deployment

**Validation:**

```bash
# Apply Terraform
cd deploy/terraform
terraform init
terraform plan
terraform apply

# Test production URL
curl https://widescreen-research.com/api/health

# Open browser
open https://widescreen-research.com
```

### Phase 5: Testing & Polish (Weeks 9-10)

**Goals:**

- Comprehensive tests
- Performance optimization
- Documentation

**Steps:**

1. Write integration tests
2. Load testing
3. Performance profiling
4. Security audit
5. Documentation

## Performance Considerations

### Latency Budget

```
User clicks "Start Campaign"
    │
    ├─ Browser → Cloud CDN: ~10ms
    │
    ├─ Cloud CDN → Cloud Run (MCP Server): ~20ms
    │
    ├─ MCP Server → Coordinator (gRPC): ~10ms
    │
    ├─ Coordinator → Firestore: ~20ms
    │
    ├─ Coordinator → Pub/Sub: ~30ms
    │
    └─ Total: ~90ms ✅

Target: < 200ms for tool invocation
```

### Caching Strategy

```go
// pkg/web/cache/cache.go
type CacheLayer struct {
    local  *ristretto.Cache  // In-memory, 100MB
    redis  *redis.Client     // Distributed, 1GB
}

func (c *CacheLayer) Get(ctx context.Context, key string) (interface{}, error) {
    // Try local first
    if val, found := c.local.Get(key); found {
        return val, nil
    }

    // Try Redis
    val, err := c.redis.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, ErrNotFound
    }

    // Store in local cache
    c.local.Set(key, val, 1)

    return val, nil
}

// Cache strategy
var CacheTTLs = map[string]time.Duration{
    "tools":         5 * time.Minute,   // Tool list rarely changes
    "campaign:meta": 10 * time.Second,  // Campaign metadata
    "ui:template":   1 * time.Hour,     // UI templates
}
```

### Connection Pooling

```go
// pkg/web/mcp_http/pool.go
var (
    coordinatorPool *grpc.ClientConn
    firestorePool   *firestore.Client
    pubsubPool      *pubsub.Client
)

func InitPools(ctx context.Context) error {
    // gRPC connection pool
    coordinatorPool, err := grpc.Dial(
        coordinatorAddr,
        grpc.WithDefaultCallOptions(
            grpc.MaxCallRecvMsgSize(10 * 1024 * 1024),
            grpc.MaxCallSendMsgSize(10 * 1024 * 1024),
        ),
        grpc.WithKeepaliveParams(keepalive.ClientParameters{
            Time:                10 * time.Second,
            Timeout:             5 * time.Second,
            PermitWithoutStream: true,
        }),
    )

    // Firestore with connection pooling
    firestorePool, err = firestore.NewClient(ctx, projectID)

    // Pub/Sub with settings
    pubsubPool, err = pubsub.NewClient(ctx, projectID,
        option.WithGRPCConnectionPool(10),
    )

    return nil
}
```

### Frontend Optimization

```typescript
// web-client/src/lib/mcp/client.ts
export class MCPClient {
  private toolsCache: Map<string, Tool[]> = new Map();
  private cacheExpiry: Map<string, number> = new Map();

  async listTools(): Promise<Tool[]> {
    const now = Date.now();
    const cached = this.toolsCache.get('tools');
    const expiry = this.cacheExpiry.get('tools');

    // Return cached if fresh
    if (cached && expiry && expiry > now) {
      return cached;
    }

    // Fetch fresh
    const tools = await this.transport.request({
      method: 'tools/list',
    });

    // Cache for 5 minutes
    this.toolsCache.set('tools', tools);
    this.cacheExpiry.set('tools', now + 5 * 60 * 1000);

    return tools;
  }

  // Debounced status polling
  private statusPoller = debounce(async (campaignId: string) => {
    const status = await this.callTool({
      name: 'get_campaign_status',
      arguments: { campaign_id: campaignId },
    });
    return status;
  }, 1000);
}
```

## Monitoring and Observability

### Metrics

```go
// pkg/web/metrics/metrics.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/metric"
)

var (
    meter = otel.Meter("widescreen.web")

    requestCounter = meter.Int64Counter(
        "http.requests",
        metric.WithDescription("Number of HTTP requests"),
    )

    requestDuration = meter.Float64Histogram(
        "http.request.duration",
        metric.WithDescription("HTTP request duration"),
        metric.WithUnit("ms"),
    )

    toolInvocations = meter.Int64Counter(
        "mcp.tool.invocations",
        metric.WithDescription("MCP tool invocations"),
    )

    sseConnections = meter.Int64UpDownCounter(
        "sse.connections",
        metric.WithDescription("Active SSE connections"),
    )
)

func RecordRequest(ctx context.Context, method, path string, duration time.Duration) {
    requestCounter.Add(ctx, 1,
        metric.WithAttributes(
            attribute.String("method", method),
            attribute.String("path", path),
        ),
    )

    requestDuration.Record(ctx, float64(duration.Milliseconds()),
        metric.WithAttributes(
            attribute.String("method", method),
            attribute.String("path", path),
        ),
    )
}
```

### Logging

```go
// pkg/web/logging/logger.go
import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func InitLogger() {
    config := zap.Config{
        Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
        Development: false,
        Encoding:    "json",
        EncoderConfig: zapcore.EncoderConfig{
            TimeKey:        "timestamp",
            LevelKey:       "severity",
            NameKey:        "logger",
            CallerKey:      "caller",
            MessageKey:     "message",
            StacktraceKey:  "stacktrace",
            LineEnding:     zapcore.DefaultLineEnding,
            EncodeLevel:    zapcore.LowercaseLevelEncoder,
            EncodeTime:     zapcore.ISO8601TimeEncoder,
            EncodeDuration: zapcore.SecondsDurationEncoder,
            EncodeCaller:   zapcore.ShortCallerEncoder,
        },
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
    }

    logger, _ = config.Build()
}

func LogToolInvocation(ctx context.Context, tool string, duration time.Duration, err error) {
    fields := []zap.Field{
        zap.String("tool", tool),
        zap.Duration("duration", duration),
    }

    if userID := ctx.Value("user_id"); userID != nil {
        fields = append(fields, zap.String("user_id", userID.(string)))
    }

    if err != nil {
        fields = append(fields, zap.Error(err))
        logger.Error("tool invocation failed", fields...)
    } else {
        logger.Info("tool invocation succeeded", fields...)
    }
}
```

### Tracing

```go
// pkg/web/tracing/tracer.go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/cloudtrace"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer(projectID string) error {
    exporter, err := cloudtrace.New(cloudtrace.WithProjectID(projectID))
    if err != nil {
        return err
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )

    otel.SetTracerProvider(tp)
    return nil
}

func TraceToolCall(ctx context.Context, toolName string, fn func(context.Context) error) error {
    tracer := otel.Tracer("widescreen.web")
    ctx, span := tracer.Start(ctx, "tool."+toolName)
    defer span.End()

    return fn(ctx)
}
```

### Dashboards

**Cloud Monitoring Dashboard Config:**

```json
{
  "displayName": "Widescreen Web MCP",
  "mosaicLayout": {
    "columns": 12,
    "tiles": [
      {
        "width": 6,
        "height": 4,
        "widget": {
          "title": "HTTP Request Rate",
          "xyChart": {
            "dataSets": [{
              "timeSeriesQuery": {
                "timeSeriesFilter": {
                  "filter": "resource.type=\"cloud_run_revision\" resource.labels.service_name=\"mcp-http-server\" metric.type=\"run.googleapis.com/request_count\""
                }
              }
            }]
          }
        }
      },
      {
        "width": 6,
        "height": 4,
        "widget": {
          "title": "Tool Invocation Latency",
          "xyChart": {
            "dataSets": [{
              "timeSeriesQuery": {
                "timeSeriesFilter": {
                  "filter": "metric.type=\"custom.googleapis.com/mcp/tool/duration\""
                }
              }
            }]
          }
        }
      },
      {
        "width": 6,
        "height": 4,
        "widget": {
          "title": "Active SSE Connections",
          "xyChart": {
            "dataSets": [{
              "timeSeriesQuery": {
                "timeSeriesFilter": {
                  "filter": "metric.type=\"custom.googleapis.com/sse/connections\""
                }
              }
            }]
          }
        }
      },
      {
        "width": 6,
        "height": 4,
        "widget": {
          "title": "Error Rate",
          "xyChart": {
            "dataSets": [{
              "timeSeriesQuery": {
                "timeSeriesFilter": {
                  "filter": "resource.type=\"cloud_run_revision\" metric.type=\"run.googleapis.com/request_count\" metric.labels.response_code_class=\"5xx\""
                }
              }
            }]
          }
        }
      }
    ]
  }
}
```

## Appendices

### A. API Reference

**MCP HTTP Endpoints:**

```
POST /api/mcp
  - MCP protocol endpoint (JSON-RPC 2.0)
  - Methods: initialize, tools/list, tools/call, resources/list, resources/read

GET /api/sse/campaigns/{id}
  - Server-Sent Events stream for campaign progress

POST /api/auth/login
  - Exchange Firebase token for session

POST /api/auth/logout
  - Invalidate session

GET /api/health
  - Health check endpoint
```

### B. Environment Variables

**MCP HTTP Server:**

```bash
PROJECT_ID=widescreen-research
COORDINATOR_URL=https://coordinator-xyz.run.app
FIRESTORE_DATABASE=(default)
JWT_SIGNING_KEY=secret://jwt-signing-key
CORS_ORIGINS=https://widescreen-research.com
LOG_LEVEL=info
```

**Web Client:**

```bash
NEXT_PUBLIC_MCP_API_URL=https://api.widescreen-research.com
NEXT_PUBLIC_FIREBASE_CONFIG='{"apiKey":"...","authDomain":"..."}'
```

### C. Deployment Commands

**Build and Deploy:**

```bash
# Build all services
make build-all

# Deploy to GCP
cd deploy/terraform
terraform apply

# Deploy web client
cd web-client
npm run build
gsutil -m rsync -r out/ gs://widescreen-web-client/

# Deploy Cloud Run services
gcloud builds submit --config deploy/cloudbuild/mcp-http-server.yaml
gcloud builds submit --config deploy/cloudbuild/coordinator.yaml
```

### D. Testing Strategy

**Unit Tests:**

```bash
# Go tests
go test ./pkg/web/...

# TypeScript tests
cd web-client
npm test
```

**Integration Tests:**

```bash
# End-to-end test
cd test/integration
go test -v
```

**Load Tests:**

```bash
# Using k6
k6 run test/load/campaign-creation.js
```

### E. Security Checklist

- [ ] HTTPS everywhere (TLS 1.3)
- [ ] Firebase Authentication configured
- [ ] IAM service accounts with least privilege
- [ ] Rate limiting enabled
- [ ] CORS configured correctly
- [ ] Input validation on all endpoints
- [ ] SQL injection prevention (N/A - Firestore)
- [ ] XSS prevention in UI rendering
- [ ] CSRF tokens for state-changing operations
- [ ] Security headers (CSP, HSTS, etc.)
- [ ] Regular dependency updates
- [ ] Secrets in Secret Manager
- [ ] Audit logging enabled

### F. Cost Estimates

**Monthly Costs (Estimated):**

| Service | Usage | Cost |
|---------|-------|------|
| Cloud Run (MCP Server) | 1M requests, 100GB-hour | $20 |
| Cloud Run (Coordinator) | 0.5M requests, 200GB-hour | $25 |
| Cloud Run (Drones) | 10K job executions, 50GB-hour | $15 |
| Cloud Storage | 10GB storage, 100GB egress | $3 |
| Cloud CDN | 1TB egress | $85 |
| Firestore | 10M reads, 1M writes | $6 |
| Pub/Sub | 100GB messages | $4 |
| Load Balancer | 1 rule, 100GB data | $20 |
| **Total** | | **~$178/month** |

*Costs scale with usage. Monitor and set budget alerts.*

---

## Conclusion

This specification provides a comprehensive blueprint for transforming the Widescreen Research system into a web-accessible MCP client architecture. The implementation follows a phased approach, ensuring each component is tested and validated before proceeding.

The result is a modern, scalable system where users can access powerful research capabilities through any web browser, with the backend remaining a clean, protocol-compliant MCP server that can also be accessed through other MCP clients like Claude Desktop or Cursor.

**Next Steps:**

1. Review and approve this specification
2. Set up GCP project and initial infrastructure
3. Begin Phase 1 implementation
4. Regular progress reviews and adjustments

**Success Criteria:**

- User can start a campaign from web browser
- Real-time dashboard updates work
- Sub-200ms tool invocation latency
- 99.9% uptime
- Positive user feedback

---

*Document Version: 1.0*
*Last Updated: September 29, 2025*
*Authors: Widescreen Research Team*
