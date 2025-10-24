# GCP Setup Complete ✓

Your GCP project `widescreen-researcher` has been successfully configured for the widescreen-research coordinator-drone architecture!

## What Was Configured

### ✅ APIs Enabled
- Cloud Run API (`run.googleapis.com`)
- Cloud Firestore API (`firestore.googleapis.com`)
- Cloud Pub/Sub API (`pubsub.googleapis.com`)
- Secret Manager API (`secretmanager.googleapis.com`)
- Cloud Build API (`cloudbuild.googleapis.com`)
- Artifact Registry API (`artifactregistry.googleapis.com`)
- Cloud Logging & Monitoring APIs

### ✅ Service Accounts Created
1. **Coordinator Service Account**: `coordinator-sa@widescreen-researcher.iam.gserviceaccount.com`
   - Roles: run.admin, iam.serviceAccountUser, pubsub.publisher, pubsub.subscriber, datastore.user, secretmanager.secretAccessor, logging.logWriter, monitoring.metricWriter

2. **Drone Service Account**: `drone-service-account@widescreen-researcher.iam.gserviceaccount.com`
   - Roles: pubsub.publisher, pubsub.subscriber, datastore.user, secretmanager.secretAccessor, logging.logWriter, monitoring.metricWriter

### ✅ Firestore Database
- **Database**: `(default)`
- **Type**: FIRESTORE_NATIVE
- **Location**: us-central1
- **Status**: Active

### ✅ Pub/Sub Resources
**Topics**:
- `drone-tasks` - For distributing tasks to drones
- `drone-results` - For collecting results from drones

**Subscriptions**:
- `drone-results-sub` - Coordinator subscription for drone results

### ✅ Artifact Registry
- **Repository**: `spawn-mcp`
- **Location**: us-central1
- **Format**: Docker
- **Purpose**: Store drone container images

### ✅ Secret Manager
- **Secret**: `exa_api_key`
- **Access**: Granted to both coordinator and drone service accounts
- **Status**: Active with 1 version

## Configuration Details

### Project Settings
- **Project ID**: widescreen-researcher
- **Default Region**: us-central1
- **Default Zone**: (not set - using regional resources)

### Resource Naming Conventions
- Service accounts: `{purpose}-sa` or `{purpose}-service-account`
- Pub/Sub topics: `drone-{purpose}`
- Artifact Registry: `spawn-mcp`
- Firestore collections: `execution_plans`, `drone_registry`, `task_results`, `research_sessions`

## Next Steps

### 1. Update Your .env File
Ensure your `.env` file has these values:

```bash
# GCP Configuration
GOOGLE_CLOUD_PROJECT=widescreen-researcher
GCP_REGION=us-central1
GOOGLE_CLOUD_REGION=us-central1

# Common
LOG_LEVEL=info
PORT=8080

# Coordinator
COORDINATOR_BASE_URL=http://localhost:8080  # For local dev

# Drone (for local testing)
DRONE_ID=local-drone-001
PUBSUB_TOPIC=drone-tasks

# External APIs (loaded from Secret Manager in production)
EXA_API_KEY=your-local-exa-key-for-testing
```

### 2. Build and Push Docker Images

```bash
# Configure Docker for Artifact Registry
gcloud auth configure-docker us-central1-docker.pkg.dev

# Build coordinator image
docker build -t us-central1-docker.pkg.dev/widescreen-researcher/spawn-mcp/coordinator:latest \
  -f Dockerfile.coordinator .

# Build drone image
docker build -t us-central1-docker.pkg.dev/widescreen-researcher/spawn-mcp/drone-researcher:latest \
  -f Dockerfile.drone .

# Push images
docker push us-central1-docker.pkg.dev/widescreen-researcher/spawn-mcp/coordinator:latest
docker push us-central1-docker.pkg.dev/widescreen-researcher/spawn-mcp/drone-researcher:latest
```

### 3. Deploy Coordinator to Cloud Run

```bash
gcloud run deploy widescreen-coordinator \
  --image=us-central1-docker.pkg.dev/widescreen-researcher/spawn-mcp/coordinator:latest \
  --region=us-central1 \
  --service-account=coordinator-sa@widescreen-researcher.iam.gserviceaccount.com \
  --set-secrets=EXA_API_KEY=exa_api_key:latest \
  --set-env-vars=GOOGLE_CLOUD_PROJECT=widescreen-researcher,GCP_REGION=us-central1 \
  --allow-unauthenticated \
  --memory=2Gi \
  --cpu=2 \
  --min-instances=0 \
  --max-instances=10 \
  --timeout=900
```

### 4. Test the Setup

```bash
# Get coordinator URL
COORDINATOR_URL=$(gcloud run services describe widescreen-coordinator \
  --region=us-central1 --format='value(status.url)')

# Test health endpoint
curl $COORDINATOR_URL/health

# Test drone spawning (via MCP client)
# See examples in cmd/widescreen-research-mcp/
```

### 5. Monitor Your Deployment

```bash
# View logs
gcloud logging tail "resource.type=cloud_run_revision AND resource.labels.service_name=widescreen-coordinator"

# View metrics
gcloud monitoring dashboards list

# Check service status
gcloud run services list --region=us-central1
```

## Validation

Run the validation script anytime to check your configuration:

```bash
./scripts/validate-gcp-config.sh
```

## Cost Management

### Current Configuration Costs (Estimated)
- **Cloud Run**: ~$0 (scale-to-zero when idle)
- **Firestore**: ~$0-5/month (free tier covers light usage)
- **Pub/Sub**: ~$0-2/month (free tier covers light usage)
- **Secret Manager**: ~$0.06/month (6 secret access operations per month)
- **Artifact Registry**: ~$0.10/GB/month for storage

### Cost Optimization Tips
1. Use scale-to-zero for all Cloud Run services
2. Set appropriate min/max instances
3. Clean up old container images regularly
4. Use Pub/Sub message retention policies
5. Monitor usage with Cloud Billing reports

## Troubleshooting

### Common Issues

**Issue**: "Permission denied" errors
- **Solution**: Verify service account has correct IAM roles
- **Check**: `gcloud projects get-iam-policy widescreen-researcher`

**Issue**: Firestore write failures
- **Solution**: Ensure service account has `roles/datastore.user`
- **Check**: Run `./scripts/validate-gcp-config.sh`

**Issue**: Secret access denied
- **Solution**: Grant `roles/secretmanager.secretAccessor` to service account
- **Command**: `gcloud secrets add-iam-policy-binding exa_api_key --member=serviceAccount:SA_EMAIL --role=roles/secretmanager.secretAccessor`

**Issue**: Cloud Run deployment fails
- **Solution**: Check image exists in Artifact Registry
- **Command**: `gcloud artifacts docker images list us-central1-docker.pkg.dev/widescreen-researcher/spawn-mcp`

## Documentation References

- **GCP Configuration Requirements**: See `GCP_CONFIGURATION_REQUIREMENTS.md`
- **Operations Guide**: See `docs/OPERATIONS.md`
- **Project Specification**: See `project_spec.md`
- **Setup Script**: `scripts/setup-gcp-project.sh`
- **Validation Script**: `scripts/validate-gcp-config.sh`

## Security Considerations

### Current Security Posture
✅ Service accounts with least-privilege IAM roles
✅ Secrets stored in Secret Manager (not environment variables)
✅ Non-root container execution
✅ Firestore in Native mode with proper access controls

### Recommended Enhancements for Production
- [ ] Enable VPC Service Controls
- [ ] Implement Binary Authorization for container signing
- [ ] Set up Cloud Armor for DDoS protection
- [ ] Enable audit logging for all services
- [ ] Implement network policies and firewall rules
- [ ] Use customer-managed encryption keys (CMEK)
- [ ] Set up alerting for suspicious activity

## Support

If you encounter issues:
1. Run `./scripts/validate-gcp-config.sh` to check configuration
2. Check Cloud Logging for error messages
3. Review `GCP_CONFIGURATION_REQUIREMENTS.md` for detailed requirements
4. Verify all environment variables are set correctly

---

**Setup completed on**: 2025-10-06
**Project**: widescreen-researcher
**Region**: us-central1
**Status**: ✅ Ready for deployment

