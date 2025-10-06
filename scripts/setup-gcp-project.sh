#!/bin/bash
# GCP Project Setup Script for Widescreen-Research
# This script aligns your GCP project with the requirements in GCP_CONFIGURATION_REQUIREMENTS.md

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ID="widescreen-researcher"
DEFAULT_REGION="us-central1"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Widescreen-Research GCP Setup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Verify project
CURRENT_PROJECT=$(gcloud config get-value project 2>/dev/null)
if [ "$CURRENT_PROJECT" != "$PROJECT_ID" ]; then
    echo -e "${YELLOW}Setting project to $PROJECT_ID${NC}"
    gcloud config set project "$PROJECT_ID"
fi

# Set default region
echo -e "${BLUE}Setting default region to $DEFAULT_REGION${NC}"
gcloud config set compute/region "$DEFAULT_REGION"
gcloud config set run/region "$DEFAULT_REGION"

echo ""
echo -e "${GREEN}✓ Project configured: $PROJECT_ID${NC}"
echo -e "${GREEN}✓ Default region: $DEFAULT_REGION${NC}"
echo ""

# ============================================
# STEP 1: Enable Required APIs
# ============================================
echo -e "${BLUE}Step 1: Enabling Required APIs${NC}"
echo "This may take a few minutes..."
echo ""

REQUIRED_APIS=(
    "run.googleapis.com"
    "firestore.googleapis.com"
    "pubsub.googleapis.com"
    "secretmanager.googleapis.com"
    "cloudbuild.googleapis.com"
    "artifactregistry.googleapis.com"
    "logging.googleapis.com"
    "monitoring.googleapis.com"
)

for api in "${REQUIRED_APIS[@]}"; do
    echo -n "Enabling $api... "
    if gcloud services enable "$api" --quiet 2>/dev/null; then
        echo -e "${GREEN}✓${NC}"
    else
        echo -e "${YELLOW}(already enabled or error)${NC}"
    fi
done

echo ""
echo -e "${GREEN}✓ APIs enabled${NC}"
echo ""

# ============================================
# STEP 2: Create Service Accounts
# ============================================
echo -e "${BLUE}Step 2: Creating Service Accounts${NC}"
echo ""

# Coordinator Service Account
COORDINATOR_SA="coordinator-sa@${PROJECT_ID}.iam.gserviceaccount.com"
echo -n "Creating coordinator service account... "
if gcloud iam service-accounts describe "$COORDINATOR_SA" &>/dev/null; then
    echo -e "${YELLOW}(already exists)${NC}"
else
    gcloud iam service-accounts create coordinator-sa \
        --display-name="Widescreen Coordinator Service Account" \
        --description="Service account for the widescreen-research coordinator" \
        --quiet
    echo -e "${GREEN}✓${NC}"
fi

# Drone Service Account
DRONE_SA="drone-service-account@${PROJECT_ID}.iam.gserviceaccount.com"
echo -n "Creating drone service account... "
if gcloud iam service-accounts describe "$DRONE_SA" &>/dev/null; then
    echo -e "${YELLOW}(already exists)${NC}"
else
    gcloud iam service-accounts create drone-service-account \
        --display-name="Widescreen Drone Service Account" \
        --description="Service account for widescreen-research drone workers" \
        --quiet
    echo -e "${GREEN}✓${NC}"
fi

echo ""
echo -e "${GREEN}✓ Service accounts created${NC}"
echo ""

# ============================================
# STEP 3: Grant IAM Roles
# ============================================
echo -e "${BLUE}Step 3: Granting IAM Roles${NC}"
echo ""

# Coordinator roles
echo "Granting roles to coordinator service account..."
COORDINATOR_ROLES=(
    "roles/run.admin"
    "roles/iam.serviceAccountUser"
    "roles/pubsub.publisher"
    "roles/pubsub.subscriber"
    "roles/firestore.dataEditor"
    "roles/secretmanager.secretAccessor"
    "roles/logging.logWriter"
    "roles/monitoring.metricWriter"
)

for role in "${COORDINATOR_ROLES[@]}"; do
    echo -n "  - $role... "
    gcloud projects add-iam-policy-binding "$PROJECT_ID" \
        --member="serviceAccount:$COORDINATOR_SA" \
        --role="$role" \
        --condition=None \
        --quiet &>/dev/null
    echo -e "${GREEN}✓${NC}"
done

# Drone roles
echo "Granting roles to drone service account..."
DRONE_ROLES=(
    "roles/pubsub.publisher"
    "roles/pubsub.subscriber"
    "roles/firestore.dataEditor"
    "roles/secretmanager.secretAccessor"
    "roles/logging.logWriter"
    "roles/monitoring.metricWriter"
)

for role in "${DRONE_ROLES[@]}"; do
    echo -n "  - $role... "
    gcloud projects add-iam-policy-binding "$PROJECT_ID" \
        --member="serviceAccount:$DRONE_SA" \
        --role="$role" \
        --condition=None \
        --quiet &>/dev/null
    echo -e "${GREEN}✓${NC}"
done

echo ""
echo -e "${GREEN}✓ IAM roles granted${NC}"
echo ""

# ============================================
# STEP 4: Initialize Firestore
# ============================================
echo -e "${BLUE}Step 4: Initializing Firestore${NC}"
echo ""

echo -n "Checking Firestore database... "
if gcloud firestore databases describe --database="(default)" &>/dev/null; then
    echo -e "${YELLOW}(already exists)${NC}"
else
    echo ""
    echo "Creating Firestore database in Native mode..."
    gcloud firestore databases create \
        --location="$DEFAULT_REGION" \
        --type=firestore-native \
        --quiet
    echo -e "${GREEN}✓ Firestore database created${NC}"
fi

echo ""
echo -e "${GREEN}✓ Firestore initialized${NC}"
echo ""

# ============================================
# STEP 5: Create Pub/Sub Topics
# ============================================
echo -e "${BLUE}Step 5: Creating Pub/Sub Topics${NC}"
echo ""

TOPICS=(
    "drone-tasks"
    "drone-results"
)

for topic in "${TOPICS[@]}"; do
    echo -n "Creating topic: $topic... "
    if gcloud pubsub topics describe "$topic" &>/dev/null; then
        echo -e "${YELLOW}(already exists)${NC}"
    else
        gcloud pubsub topics create "$topic" --quiet
        echo -e "${GREEN}✓${NC}"
    fi
done

# Create subscriptions
echo ""
echo "Creating Pub/Sub subscriptions..."
echo -n "  - drone-results-sub... "
if gcloud pubsub subscriptions describe "drone-results-sub" &>/dev/null; then
    echo -e "${YELLOW}(already exists)${NC}"
else
    gcloud pubsub subscriptions create "drone-results-sub" \
        --topic="drone-results" \
        --ack-deadline=60 \
        --message-retention-duration=7d \
        --quiet
    echo -e "${GREEN}✓${NC}"
fi

echo ""
echo -e "${GREEN}✓ Pub/Sub topics and subscriptions created${NC}"
echo ""

# ============================================
# STEP 6: Create Artifact Registry Repository
# ============================================
echo -e "${BLUE}Step 6: Creating Artifact Registry Repository${NC}"
echo ""

REPO_NAME="spawn-mcp"
echo -n "Creating repository: $REPO_NAME... "
if gcloud artifacts repositories describe "$REPO_NAME" --location="$DEFAULT_REGION" &>/dev/null; then
    echo -e "${YELLOW}(already exists)${NC}"
else
    gcloud artifacts repositories create "$REPO_NAME" \
        --repository-format=docker \
        --location="$DEFAULT_REGION" \
        --description="Container images for widescreen-research drones" \
        --quiet
    echo -e "${GREEN}✓${NC}"
fi

echo ""
echo -e "${GREEN}✓ Artifact Registry repository created${NC}"
echo ""

# ============================================
# STEP 7: Verify Secret Manager
# ============================================
echo -e "${BLUE}Step 7: Verifying Secret Manager${NC}"
echo ""

echo -n "Checking for exa_api_key secret... "
if gcloud secrets describe "exa_api_key" &>/dev/null; then
    echo -e "${GREEN}✓ (exists)${NC}"
else
    echo -e "${RED}✗ (not found)${NC}"
    echo ""
    echo -e "${YELLOW}WARNING: exa_api_key secret not found!${NC}"
    echo "You need to create it manually with your Exa API key:"
    echo ""
    echo "  echo -n 'YOUR_EXA_API_KEY' | gcloud secrets create exa_api_key --data-file=-"
    echo ""
fi

# Grant secret access to service accounts
echo "Granting secret access to service accounts..."
for sa in "$COORDINATOR_SA" "$DRONE_SA"; do
    echo -n "  - $sa... "
    gcloud secrets add-iam-policy-binding "exa_api_key" \
        --member="serviceAccount:$sa" \
        --role="roles/secretmanager.secretAccessor" \
        --quiet &>/dev/null || true
    echo -e "${GREEN}✓${NC}"
done

echo ""
echo -e "${GREEN}✓ Secret Manager configured${NC}"
echo ""

# ============================================
# Summary
# ============================================
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Setup Complete!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}✓ APIs enabled${NC}"
echo -e "${GREEN}✓ Service accounts created${NC}"
echo -e "${GREEN}✓ IAM roles granted${NC}"
echo -e "${GREEN}✓ Firestore initialized${NC}"
echo -e "${GREEN}✓ Pub/Sub topics created${NC}"
echo -e "${GREEN}✓ Artifact Registry repository created${NC}"
echo -e "${GREEN}✓ Secret Manager configured${NC}"
echo ""
echo "Next steps:"
echo "1. Verify your .env file has the correct values"
echo "2. Build and push Docker images to Artifact Registry"
echo "3. Deploy coordinator to Cloud Run"
echo "4. Test drone spawning"
echo ""
echo "For more details, see GCP_CONFIGURATION_REQUIREMENTS.md"
echo ""

