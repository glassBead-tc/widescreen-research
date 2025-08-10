package operations

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/google/uuid"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
	"google.golang.org/protobuf/types/known/durationpb"
)

// GCPProvisioner handles GCP resource provisioning
type GCPProvisioner struct {
	projectID       string
	region          string
	runClient       *run.ServicesClient
	pubsubClient    *pubsub.Client
	firestoreClient *firestore.Client
}

// NewGCPProvisioner creates a new GCP provisioner
func NewGCPProvisioner() *GCPProvisioner {
	return &GCPProvisioner{
		projectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
		region:    getEnvOrDefault("GOOGLE_CLOUD_REGION", "us-central1"),
	}
}

// Execute provisions GCP resources based on parameters
func (gp *GCPProvisioner) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Initialize clients if needed
	if err := gp.initializeClients(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize GCP clients: %w", err)
	}

	// Parse request parameters
	resourceType, ok := params["resource_type"].(string)
	if !ok {
		return nil, fmt.Errorf("resource_type parameter is required")
	}

	count := 1
	if c, ok := params["count"].(float64); ok {
		count = int(c)
	}

	region := gp.region
	if r, ok := params["region"].(string); ok {
		region = r
	}

	config := make(map[string]interface{})
	if c, ok := params["config"].(map[string]interface{}); ok {
		config = c
	}

	// Create request
	request := &schemas.GCPProvisionRequest{
		ResourceType: resourceType,
		Count:        count,
		Region:       region,
		Config:       config,
	}

	// Provision based on resource type
	switch request.ResourceType {
	case "cloud_run":
		return gp.provisionCloudRun(ctx, request)
	case "pubsub":
		return gp.provisionPubSub(ctx, request)
	case "firestore":
		return gp.provisionFirestore(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", request.ResourceType)
	}
}

// provisionCloudRun provisions Cloud Run services
func (gp *GCPProvisioner) provisionCloudRun(ctx context.Context, request *schemas.GCPProvisionRequest) (*schemas.GCPProvisionResponse, error) {
	resources := make([]schemas.GCPResource, 0, request.Count)

	for i := 0; i < request.Count; i++ {
		resourceID := fmt.Sprintf("service-%s-%d", uuid.New().String()[:8], i)
		
		// Extract configuration
		image := "gcr.io/cloudrun/hello" // Default image
		if img, ok := request.Config["image"].(string); ok {
			image = img
		}

		cpu := "1000m"
		if c, ok := request.Config["cpu"].(string); ok {
			cpu = c
		}

		memory := "512Mi"
		if m, ok := request.Config["memory"].(string); ok {
			memory = m
		}

		timeout := int64(300) // 5 minutes default
		if t, ok := request.Config["timeout_seconds"].(float64); ok {
			timeout = int64(t)
		}

		// Create service configuration
		service := &runpb.Service{
			Name: resourceID,
			Template: &runpb.RevisionTemplate{
				Containers: []*runpb.Container{
					{
						Image: image,
						Resources: &runpb.ResourceRequirements{
							Limits: map[string]string{
								"cpu":    cpu,
								"memory": memory,
							},
						},
					},
				},
				MaxInstanceRequestConcurrency: 100,
				Timeout:                      &durationpb.Duration{Seconds: timeout},
			},
		}

		// Add environment variables if provided
		if envVars, ok := request.Config["env_vars"].(map[string]interface{}); ok {
			envs := make([]*runpb.EnvVar, 0, len(envVars))
			for k, v := range envVars {
				envs = append(envs, &runpb.EnvVar{
					Name:   k,
					Values: &runpb.EnvVar_Value{Value: fmt.Sprintf("%v", v)},
				})
			}
			service.Template.Containers[0].Env = envs
		}

		// Deploy the service
		operation, err := gp.runClient.CreateService(ctx, &runpb.CreateServiceRequest{
			Parent:    fmt.Sprintf("projects/%s/locations/%s", gp.projectID, request.Region),
			ServiceId: resourceID,
			Service:   service,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Cloud Run service: %w", err)
		}

		// Wait for deployment
		svc, err := operation.Wait(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for service deployment: %w", err)
		}

		resources = append(resources, schemas.GCPResource{
			ID:        resourceID,
			Type:      "cloud_run",
			URL:       svc.Uri,
			Status:    "ready",
			Region:    request.Region,
			CreatedAt: time.Now(),
		})
	}

	return &schemas.GCPProvisionResponse{
		Resources: resources,
		Status:    "success",
		Message:   fmt.Sprintf("Successfully provisioned %d Cloud Run services", len(resources)),
	}, nil
}

// provisionPubSub provisions Pub/Sub topics and subscriptions
func (gp *GCPProvisioner) provisionPubSub(ctx context.Context, request *schemas.GCPProvisionRequest) (*schemas.GCPProvisionResponse, error) {
	resources := make([]schemas.GCPResource, 0, request.Count)

	for i := 0; i < request.Count; i++ {
		topicID := fmt.Sprintf("topic-%s-%d", uuid.New().String()[:8], i)
		
		// Create topic
		topic, err := gp.pubsubClient.CreateTopic(ctx, topicID)
		if err != nil {
			return nil, fmt.Errorf("failed to create topic: %w", err)
		}

		// Create subscription if requested
		if createSub, ok := request.Config["create_subscription"].(bool); ok && createSub {
			subID := fmt.Sprintf("sub-%s", topicID)
			_, err := gp.pubsubClient.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
				Topic:             topic,
				AckDeadline:       30 * time.Second,
				RetentionDuration: 24 * time.Hour,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create subscription: %w", err)
			}
		}

		resources = append(resources, schemas.GCPResource{
			ID:        topicID,
			Type:      "pubsub_topic",
			Status:    "ready",
			Region:    "global",
			CreatedAt: time.Now(),
		})
	}

	return &schemas.GCPProvisionResponse{
		Resources: resources,
		Status:    "success",
		Message:   fmt.Sprintf("Successfully provisioned %d Pub/Sub topics", len(resources)),
	}, nil
}

// provisionFirestore provisions Firestore collections
func (gp *GCPProvisioner) provisionFirestore(ctx context.Context, request *schemas.GCPProvisionRequest) (*schemas.GCPProvisionResponse, error) {
	resources := make([]schemas.GCPResource, 0, request.Count)

	collectionPrefix := "collection"
	if prefix, ok := request.Config["collection_prefix"].(string); ok {
		collectionPrefix = prefix
	}

	for i := 0; i < request.Count; i++ {
		collectionID := fmt.Sprintf("%s-%s-%d", collectionPrefix, uuid.New().String()[:8], i)
		
		// Create initial document to establish collection
		doc := gp.firestoreClient.Collection(collectionID).Doc("_init")
		_, err := doc.Set(ctx, map[string]interface{}{
			"created_at": time.Now(),
			"type":       "initialization",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}

		resources = append(resources, schemas.GCPResource{
			ID:        collectionID,
			Type:      "firestore_collection",
			Status:    "ready",
			Region:    request.Region,
			CreatedAt: time.Now(),
		})
	}

	return &schemas.GCPProvisionResponse{
		Resources: resources,
		Status:    "success",
		Message:   fmt.Sprintf("Successfully provisioned %d Firestore collections", len(resources)),
	}, nil
}

// initializeClients initializes GCP clients if not already initialized
func (gp *GCPProvisioner) initializeClients(ctx context.Context) error {
	if gp.runClient == nil {
		client, err := run.NewServicesClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create Cloud Run client: %w", err)
		}
		gp.runClient = client
	}

	if gp.pubsubClient == nil {
		client, err := pubsub.NewClient(ctx, gp.projectID)
		if err != nil {
			return fmt.Errorf("failed to create Pub/Sub client: %w", err)
		}
		gp.pubsubClient = client
	}

	if gp.firestoreClient == nil {
		client, err := firestore.NewClient(ctx, gp.projectID)
		if err != nil {
			return fmt.Errorf("failed to create Firestore client: %w", err)
		}
		gp.firestoreClient = client
	}

	return nil
}

// GetDescription returns the operation description
func (gp *GCPProvisioner) GetDescription() string {
	return "Provisions GCP resources including Cloud Run services, Pub/Sub topics, and Firestore collections"
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}