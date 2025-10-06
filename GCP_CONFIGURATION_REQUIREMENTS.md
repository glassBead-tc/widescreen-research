# GCP Configuration Requirements for Widescreen-Research

This document outlines the complete GCP project configuration required for the widescreen-research coordinator-drone architecture.

## Overview

The project uses a coordinator-worker pattern where:
- **Coordinator**: High-powered server that orchestrates drone lifecycle and task distribution
- **Drones**: Lightweight workers deployed on Cloud Run that perform research tasks
- **Communication**: Pub/Sub for async messaging, Firestore for state management

## Required GCP APIs

The following APIs must be enabled in your GCP project:

### Core Services
1. **Cloud Run API** (`run.googleapis.com`)
   - Purpose: Deploy and manage coordinator and drone services
   - Used by: Coordinator to spawn drones, both services for hosting

2. **Cloud Firestore API** (`firestore.googleapis.com`)
   - Purpose: Distributed state management and coordination
   - Collections used:
     - `execution_plans` - Task execution plans
     - `drone_registry` - Active drone tracking
     - `task_results` - Aggregated research results
     - `research_sessions` - Session metadata for widescreen-research-mcp

3. **Cloud Pub/Sub API** (`pubsub.googleapis.com`)
   - Purpose: Asynchronous message passing between coordinator and drones
   - Topics used:
     - `drone-tasks` - Task distribution to drones
     - `drone-results` - Result collection from drones
     - `research-results-{session_id}` - Session-specific result topics

4. **Secret Manager API** (`secretmanager.googleapis.com`)
   - Purpose: Secure storage of API keys and credentials
   - Secrets required:
     - `exa_api_key` - Exa AI API key for research capabilities

### Supporting Services
5. **Cloud Build API** (`cloudbuild.googleapis.com`)
   - Purpose: Building container images for deployment
   - Optional but recommended for CI/CD

6. **Artifact Registry API** (`artifactregistry.googleapis.com`)
   - Purpose: Storing Docker images
   - Repository: `gcr.io/{PROJECT_ID}/spawn-mcp/`
   - Images:
     - `drone-worker:latest`
     - `drone-analyzer:latest`
     - `drone-processor:latest`
     - `drone-researcher:latest`
     - `drone-synthesizer:latest`

7. **Cloud Logging API** (`logging.googleapis.com`)
   - Purpose: Centralized logging (enabled by default)

8. **Cloud Monitoring API** (`monitoring.googleapis.com`)
   - Purpose: Metrics and alerting (enabled by default)

## Service Accounts

### 1. Coordinator Service Account
**Name**: `coordinator-sa@{PROJECT_ID}.iam.gserviceaccount.com`

**Purpose**: Run the coordinator service with permissions to spawn drones and manage resources

**Required IAM Roles**:
- `roles/run.admin` - Create and manage Cloud Run services (drones)
- `roles/iam.serviceAccountUser` - Impersonate drone service account
- `roles/pubsub.publisher` - Publish tasks to drones
- `roles/pubsub.subscriber` - Subscribe to drone results
- `roles/firestore.dataEditor` - Read/write Firestore state
- `roles/secretmanager.secretAccessor` - Access API keys from Secret Manager
- `roles/logging.logWriter` - Write logs
- `roles/monitoring.metricWriter` - Write metrics

### 2. Drone Service Account
**Name**: `drone-service-account@{PROJECT_ID}.iam.gserviceaccount.com`

**Purpose**: Run drone services with minimal permissions

**Required IAM Roles**:
- `roles/pubsub.publisher` - Publish results back to coordinator
- `roles/pubsub.subscriber` - Subscribe to task queue
- `roles/firestore.dataEditor` - Write task results to Firestore
- `roles/secretmanager.secretAccessor` - Access Exa API key
- `roles/logging.logWriter` - Write logs
- `roles/monitoring.metricWriter` - Write metrics

**Note**: Drone service account is referenced in `pkg/gcp/client.go:124`

## Firestore Configuration

### Database Setup
- **Mode**: Native mode (not Datastore mode)
- **Location**: Same as `GCP_REGION` (default: `us-central1`)
- **Database ID**: `(default)` or custom

### Collections Schema

#### `execution_plans`
```
{
  "plan_id": string,
  "task_description": string,
  "drone_count": number,
  "estimated_cost": number,
  "estimated_time_minutes": number,
  "created_at": timestamp,
  "status": string  // "pending", "running", "completed", "failed"
}
```

#### `drone_registry`
```
{
  "drone_id": string,
  "drone_type": string,  // "worker", "analyzer", "processor", "researcher", "synthesizer"
  "service_url": string,
  "region": string,
  "status": string,  // "spawning", "ready", "busy", "error"
  "last_heartbeat": timestamp,
  "current_task_id": string,
  "capabilities": array<string>
}
```

#### `task_results`
```
{
  "task_id": string,
  "drone_id": string,
  "result_data": map,
  "completed_at": timestamp,
  "execution_time_ms": number
}
```

#### `research_sessions` (widescreen-research-mcp)
```
{
  "session_id": string,
  "topic": string,
  "researcher_count": number,
  "status": string,
  "created_at": timestamp,
  "completed_at": timestamp
}
```

## Pub/Sub Configuration

### Topics

#### `drone-tasks`
- **Purpose**: Coordinator publishes tasks for drones to consume
- **Subscriptions**: Created dynamically by drones
- **Message Schema**:
  ```json
  {
    "task_id": "string",
    "drone_id": "string",
    "task_type": "string",
    "parameters": {}
  }
  ```

#### `drone-results`
- **Purpose**: Drones publish results back to coordinator
- **Subscriptions**: `drone-results-sub` (coordinator)
- **Message Schema**:
  ```json
  {
    "task_id": "string",
    "drone_id": "string",
    "status": "success|error",
    "result": {},
    "error": "string"
  }
  ```

#### `research-results-{session_id}` (dynamic)
- **Purpose**: Session-specific result collection
- **Lifecycle**: Created per research session, cleaned up after completion
- **Retention**: 7 days

### Subscription Configuration
- **Acknowledgement Deadline**: 60 seconds
- **Message Retention**: 7 days
- **Retry Policy**: Exponential backoff, max 10 retries

## Secret Manager Secrets

### `exa_api_key`
- **Purpose**: Exa AI API key for research drones
- **Access**: Granted to both coordinator and drone service accounts
- **Rotation**: Manual (recommended: quarterly)
- **Versions**: Keep latest 3 versions

### Future Secrets (Planned)
- `anthropic_api_key` - For AI-powered report generation
- `openai_api_key` - Alternative LLM provider

## Cloud Run Configuration

### Coordinator Service
- **Name**: `widescreen-coordinator`
- **Region**: `us-central1` (or as specified in `GCP_REGION`)
- **CPU**: 2 vCPU
- **Memory**: 2Gi
- **Min Instances**: 0 (scale to zero)
- **Max Instances**: 10
- **Timeout**: 15 minutes
- **Concurrency**: 80
- **Service Account**: `coordinator-sa@{PROJECT_ID}.iam.gserviceaccount.com`
- **Ingress**: All traffic (or Internal + Cloud Load Balancing for production)
- **Environment Variables**:
  - `GOOGLE_CLOUD_PROJECT`
  - `GCP_REGION`
  - `LOG_LEVEL`
  - `PORT=8080`

### Drone Services (Dynamically Created)
- **Name Pattern**: `drone-{type}-{uuid}`
- **Region**: Same as coordinator or specified per-task
- **CPU**: 1 vCPU
- **Memory**: 512Mi
- **Min Instances**: 0
- **Max Instances**: 10
- **Timeout**: 5 minutes
- **Service Account**: `drone-service-account@{PROJECT_ID}.iam.gserviceaccount.com`
- **Ingress**: All traffic
- **Environment Variables**:
  - `DRONE_ID` (set by coordinator)
  - `DRONE_TYPE` (set by coordinator)
  - `GOOGLE_CLOUD_PROJECT`
  - `PUBSUB_TOPIC`
  - `COORDINATOR_URL`
  - `SESSION_ID` (for widescreen-research-mcp)

## Artifact Registry Repository

### Repository Details
- **Format**: Docker
- **Location**: Same as `GCP_REGION`
- **Name**: `spawn-mcp`
- **Full Path**: `{GCP_REGION}-docker.pkg.dev/{PROJECT_ID}/spawn-mcp`

### Alternative: Container Registry (Legacy)
- **Path**: `gcr.io/{PROJECT_ID}/spawn-mcp/`
- **Note**: Container Registry is being deprecated; migrate to Artifact Registry

## Network Configuration

### Current (Development)
- **Ingress**: All traffic allowed
- **Egress**: All traffic allowed
- **VPC**: Default network

### Recommended (Production)
- **VPC**: Custom VPC with private subnets
- **Serverless VPC Access**: Connector for Cloud Run to VPC
- **Cloud NAT**: For outbound internet access from private services
- **VPC Service Controls**: Perimeter around sensitive APIs

## IAM Bindings Summary

```bash
# Coordinator service account
gcloud projects add-iam-policy-binding {PROJECT_ID} \
  --member="serviceAccount:coordinator-sa@{PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/run.admin"

# Drone service account  
gcloud projects add-iam-policy-binding {PROJECT_ID} \
  --member="serviceAccount:drone-service-account@{PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/pubsub.publisher"
```

## Cost Estimates

### Monthly Costs (Moderate Usage)
- **Cloud Run**: ~$50-100 (coordinator + drones)
- **Firestore**: ~$10-20 (reads/writes/storage)
- **Pub/Sub**: ~$5-10 (message volume)
- **Secret Manager**: ~$1 (secret access)
- **Logging/Monitoring**: ~$10-20
- **Total**: ~$75-150/month

### Cost Optimization
- Scale-to-zero for all services
- Firestore query optimization
- Pub/Sub message batching
- Log retention policies (30 days)

## Validation Checklist

- [ ] All required APIs enabled
- [ ] Service accounts created with correct roles
- [ ] Firestore database initialized in Native mode
- [ ] Pub/Sub topics created
- [ ] Secret Manager secrets configured
- [ ] Artifact Registry repository created
- [ ] Cloud Run services deployable
- [ ] IAM bindings verified
- [ ] Network connectivity tested
- [ ] Logging and monitoring configured

