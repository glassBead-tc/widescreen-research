#!/bin/bash
# GCP Configuration Validation Script
# Validates that all required GCP resources are properly configured

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ID="widescreen-researcher"
REGION="us-central1"
ERRORS=0
WARNINGS=0

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}GCP Configuration Validation${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Helper functions
check_pass() {
    echo -e "${GREEN}✓${NC} $1"
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    ((ERRORS++))
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARNINGS++))
}

# Validate project
echo -e "${BLUE}Checking Project Configuration...${NC}"
CURRENT_PROJECT=$(gcloud config get-value project 2>/dev/null)
if [ "$CURRENT_PROJECT" = "$PROJECT_ID" ]; then
    check_pass "Project: $PROJECT_ID"
else
    check_fail "Project mismatch: expected $PROJECT_ID, got $CURRENT_PROJECT"
fi
echo ""

# Validate APIs
echo -e "${BLUE}Checking Required APIs...${NC}"
REQUIRED_APIS=(
    "run.googleapis.com:Cloud Run"
    "firestore.googleapis.com:Firestore"
    "pubsub.googleapis.com:Pub/Sub"
    "secretmanager.googleapis.com:Secret Manager"
    "cloudbuild.googleapis.com:Cloud Build"
    "artifactregistry.googleapis.com:Artifact Registry"
)

for api_info in "${REQUIRED_APIS[@]}"; do
    IFS=':' read -r api name <<< "$api_info"
    if gcloud services list --enabled --filter="config.name:$api" --format="value(config.name)" 2>/dev/null | grep -q "$api"; then
        check_pass "$name API enabled"
    else
        check_fail "$name API not enabled"
    fi
done
echo ""

# Validate Service Accounts
echo -e "${BLUE}Checking Service Accounts...${NC}"
if gcloud iam service-accounts describe "coordinator-sa@${PROJECT_ID}.iam.gserviceaccount.com" &>/dev/null; then
    check_pass "Coordinator service account exists"
else
    check_fail "Coordinator service account missing"
fi

if gcloud iam service-accounts describe "drone-service-account@${PROJECT_ID}.iam.gserviceaccount.com" &>/dev/null; then
    check_pass "Drone service account exists"
else
    check_fail "Drone service account missing"
fi
echo ""

# Validate IAM Roles
echo -e "${BLUE}Checking IAM Roles...${NC}"
COORDINATOR_SA="coordinator-sa@${PROJECT_ID}.iam.gserviceaccount.com"
COORDINATOR_ROLES=("roles/run.admin" "roles/iam.serviceAccountUser" "roles/pubsub.publisher" "roles/datastore.user")

for role in "${COORDINATOR_ROLES[@]}"; do
    if gcloud projects get-iam-policy "$PROJECT_ID" --flatten="bindings[].members" \
        --filter="bindings.role:$role AND bindings.members:serviceAccount:$COORDINATOR_SA" \
        --format="value(bindings.role)" 2>/dev/null | grep -q "$role"; then
        check_pass "Coordinator has $role"
    else
        check_fail "Coordinator missing $role"
    fi
done
echo ""

# Validate Firestore
echo -e "${BLUE}Checking Firestore...${NC}"
if gcloud firestore databases describe --database="(default)" &>/dev/null; then
    DB_TYPE=$(gcloud firestore databases describe --database="(default)" --format="value(type)" 2>/dev/null)
    if [ "$DB_TYPE" = "FIRESTORE_NATIVE" ]; then
        check_pass "Firestore database exists (Native mode)"
    else
        check_warn "Firestore database exists but not in Native mode: $DB_TYPE"
    fi
else
    check_fail "Firestore database not found"
fi
echo ""

# Validate Pub/Sub
echo -e "${BLUE}Checking Pub/Sub Topics...${NC}"
for topic in "drone-tasks" "drone-results"; do
    if gcloud pubsub topics describe "$topic" &>/dev/null; then
        check_pass "Topic: $topic"
    else
        check_fail "Topic missing: $topic"
    fi
done

echo ""
echo -e "${BLUE}Checking Pub/Sub Subscriptions...${NC}"
if gcloud pubsub subscriptions describe "drone-results-sub" &>/dev/null; then
    check_pass "Subscription: drone-results-sub"
else
    check_fail "Subscription missing: drone-results-sub"
fi
echo ""

# Validate Artifact Registry
echo -e "${BLUE}Checking Artifact Registry...${NC}"
if gcloud artifacts repositories describe "spawn-mcp" --location="$REGION" &>/dev/null; then
    check_pass "Repository: spawn-mcp"
else
    check_fail "Repository missing: spawn-mcp"
fi
echo ""

# Validate Secrets
echo -e "${BLUE}Checking Secret Manager...${NC}"
if gcloud secrets describe "exa_api_key" &>/dev/null; then
    check_pass "Secret: exa_api_key"
    
    # Check if secret has a version
    VERSION_COUNT=$(gcloud secrets versions list "exa_api_key" --format="value(name)" 2>/dev/null | wc -l)
    if [ "$VERSION_COUNT" -gt 0 ]; then
        check_pass "Secret has $VERSION_COUNT version(s)"
    else
        check_warn "Secret exists but has no versions"
    fi
else
    check_fail "Secret missing: exa_api_key"
fi
echo ""

# Test Firestore access
echo -e "${BLUE}Testing Firestore Access...${NC}"
TEST_COLLECTION="validation_test"
TEST_DOC="test_$(date +%s)"
if gcloud firestore documents create --database="(default)" \
    --collection="$TEST_COLLECTION" \
    --document-id="$TEST_DOC" \
    --data='{"test": true, "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"}' &>/dev/null; then
    check_pass "Firestore write access"
    
    # Clean up test document
    gcloud firestore documents delete --database="(default)" \
        "$TEST_COLLECTION/$TEST_DOC" --quiet &>/dev/null || true
else
    check_warn "Firestore write access test failed (may need authentication)"
fi
echo ""

# Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Validation Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

if [ $ERRORS -eq 0 ] && [ $WARNINGS -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed!${NC}"
    echo ""
    echo "Your GCP project is properly configured for widescreen-research."
    exit 0
elif [ $ERRORS -eq 0 ]; then
    echo -e "${YELLOW}⚠ Validation completed with $WARNINGS warning(s)${NC}"
    echo ""
    echo "Your GCP project is mostly configured, but review the warnings above."
    exit 0
else
    echo -e "${RED}✗ Validation failed with $ERRORS error(s) and $WARNINGS warning(s)${NC}"
    echo ""
    echo "Please fix the errors above before deploying."
    echo "Run ./scripts/setup-gcp-project.sh to fix missing resources."
    exit 1
fi

