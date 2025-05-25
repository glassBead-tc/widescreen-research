package gcp

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Client wraps all GCP service clients
type Client struct {
	ProjectID       string
	Region          string
	RunClient       *run.ServicesClient
	FirestoreClient *firestore.Client
	PubSubClient    *pubsub.Client
}

// NewClient creates a new GCP client with all necessary services
func NewClient(ctx context.Context, projectID, region string, opts ...option.ClientOption) (*Client, error) {
	// Initialize Cloud Run client
	runClient, err := run.NewServicesClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run client: %w", err)
	}

	// Initialize Firestore client
	firestoreClient, err := firestore.NewClient(ctx, projectID, opts...)
	if err != nil {
		runClient.Close()
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}

	// Initialize Pub/Sub client
	pubsubClient, err := pubsub.NewClient(ctx, projectID, opts...)
	if err != nil {
		runClient.Close()
		firestoreClient.Close()
		return nil, fmt.Errorf("failed to create Pub/Sub client: %w", err)
	}

	return &Client{
		ProjectID:       projectID,
		Region:          region,
		RunClient:       runClient,
		FirestoreClient: firestoreClient,
		PubSubClient:    pubsubClient,
	}, nil
}

// Close closes all GCP clients
func (c *Client) Close() error {
	var errs []error

	if err := c.RunClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close Cloud Run client: %w", err))
	}

	if err := c.FirestoreClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close Firestore client: %w", err))
	}

	if err := c.PubSubClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close Pub/Sub client: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing clients: %v", errs)
	}

	return nil
}

// CreateCloudRunService creates a new Cloud Run service for a drone
func (c *Client) CreateCloudRunService(ctx context.Context, serviceName, imageURI string, env map[string]string) (*runpb.Service, error) {
	log.Printf("Creating Cloud Run service: %s with image: %s", serviceName, imageURI)

	// Convert env map to EnvVar slice with correct structure
	var envVars []*runpb.EnvVar
	for key, value := range env {
		envVars = append(envVars, &runpb.EnvVar{
			Name: key,
			Values: &runpb.EnvVar_Value{
				Value: value,
			},
		})
	}

	// Build the service request with proper API structure
	req := &runpb.CreateServiceRequest{
		Parent:    fmt.Sprintf("projects/%s/locations/%s", c.ProjectID, c.Region),
		ServiceId: serviceName,
		Service: &runpb.Service{
			Template: &runpb.RevisionTemplate{
				Containers: []*runpb.Container{
					{
						Image: imageURI,
						Env:   envVars,
						Resources: &runpb.ResourceRequirements{
							Limits: map[string]string{
								"memory": "512Mi",
								"cpu":    "1000m",
							},
						},
						Ports: []*runpb.ContainerPort{
							{
								Name:          "http1",
								ContainerPort: 8080,
							},
						},
					},
				},
				Scaling: &runpb.RevisionScaling{
					MinInstanceCount: 0,
					MaxInstanceCount: 10,
				},
				ServiceAccount: fmt.Sprintf("drone-service-account@%s.iam.gserviceaccount.com", c.ProjectID),
				Timeout:        durationpb.New(5 * time.Minute), // 5 minute timeout
			},
			// Configure IAM policy for service-to-service authentication
			// This allows the coordinator to invoke the drone service
			Ingress: runpb.IngressTraffic_INGRESS_TRAFFIC_ALL, // Allow all traffic for now
		},
	}

	// Create the service (returns long-running operation)
	op, err := c.RunClient.CreateService(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run service: %w", err)
	}

	log.Printf("Service creation initiated, waiting for completion...")

	// Wait for operation to complete
	service, err := op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for service creation: %w", err)
	}

	log.Printf("Successfully created Cloud Run service: %s at %s", service.Name, service.Uri)
	return service, nil
}

// DeleteCloudRunService deletes a Cloud Run service
func (c *Client) DeleteCloudRunService(ctx context.Context, serviceName string) error {
	req := &runpb.DeleteServiceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/services/%s", c.ProjectID, c.Region, serviceName),
	}

	op, err := c.RunClient.DeleteService(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete Cloud Run service: %w", err)
	}

	// Wait for deletion to complete
	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for service deletion: %w", err)
	}

	log.Printf("Deleted Cloud Run service: %s", serviceName)
	return nil
}

// GetServiceURL retrieves the URL for a Cloud Run service
func (c *Client) GetServiceURL(ctx context.Context, serviceName string) (string, error) {
	req := &runpb.GetServiceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/services/%s", c.ProjectID, c.Region, serviceName),
	}

	service, err := c.RunClient.GetService(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get service: %w", err)
	}

	if service.Uri == "" {
		return "", fmt.Errorf("service URL not available yet")
	}

	return service.Uri, nil
}

// UpdateServiceTraffic updates traffic allocation for a Cloud Run service
func (c *Client) UpdateServiceTraffic(ctx context.Context, serviceName string, trafficPercent int32) error {
	log.Printf("Updating traffic for service %s to %d%%", serviceName, trafficPercent)

	// Get the current service
	getReq := &runpb.GetServiceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/services/%s", c.ProjectID, c.Region, serviceName),
	}

	service, err := c.RunClient.GetService(ctx, getReq)
	if err != nil {
		return fmt.Errorf("failed to get service for traffic update: %w", err)
	}

	// Update traffic allocation
	service.Traffic = []*runpb.TrafficTarget{
		{
			Type:     runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST,
			Percent:  trafficPercent,
			Revision: "", // Empty for latest revision
		},
	}

	// Update the service
	updateReq := &runpb.UpdateServiceRequest{
		Service: service,
	}

	op, err := c.RunClient.UpdateService(ctx, updateReq)
	if err != nil {
		return fmt.Errorf("failed to update service traffic: %w", err)
	}

	// Wait for update to complete
	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for traffic update: %w", err)
	}

	log.Printf("Successfully updated traffic for service %s to %d%%", serviceName, trafficPercent)
	return nil
}

// StoreDocument stores a document in Firestore
func (c *Client) StoreDocument(ctx context.Context, collection, docID string, data interface{}) error {
	_, err := c.FirestoreClient.Collection(collection).Doc(docID).Set(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to store document: %w", err)
	}
	return nil
}

// GetDocument retrieves a document from Firestore
func (c *Client) GetDocument(ctx context.Context, collection, docID string, dest interface{}) error {
	doc, err := c.FirestoreClient.Collection(collection).Doc(docID).Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	if err := doc.DataTo(dest); err != nil {
		return fmt.Errorf("failed to unmarshal document: %w", err)
	}

	return nil
}

// PublishMessage publishes a message to a Pub/Sub topic
func (c *Client) PublishMessage(ctx context.Context, topicName string, data []byte, attributes map[string]string) error {
	topic := c.PubSubClient.Topic(topicName)

	// Check if topic exists, create if it doesn't
	exists, err := topic.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check topic existence: %w", err)
	}

	if !exists {
		_, err = c.PubSubClient.CreateTopic(ctx, topicName)
		if err != nil {
			return fmt.Errorf("failed to create topic: %w", err)
		}
	}

	msg := &pubsub.Message{
		Data:       data,
		Attributes: attributes,
	}

	result := topic.Publish(ctx, msg)

	// Wait for publish to complete
	_, err = result.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// SubscribeToTopic subscribes to a Pub/Sub topic with a callback
func (c *Client) SubscribeToTopic(ctx context.Context, subscriptionName string, callback func(ctx context.Context, msg *pubsub.Message)) error {
	sub := c.PubSubClient.Subscription(subscriptionName)

	// Check if subscription exists, create if it doesn't
	exists, err := sub.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check subscription existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("subscription %s does not exist", subscriptionName)
	}

	// Configure subscription settings
	sub.ReceiveSettings.MaxOutstandingMessages = 100

	// Start receiving messages
	err = sub.Receive(ctx, callback)
	if err != nil {
		return fmt.Errorf("failed to receive messages: %w", err)
	}

	return nil
}

// WaitForServiceReady waits for a Cloud Run service to be ready
func (c *Client) WaitForServiceReady(ctx context.Context, serviceName string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for service to be ready")
		case <-ticker.C:
			req := &runpb.GetServiceRequest{
				Name: fmt.Sprintf("projects/%s/locations/%s/services/%s", c.ProjectID, c.Region, serviceName),
			}

			service, err := c.RunClient.GetService(ctx, req)
			if err != nil {
				log.Printf("Error checking service status: %v", err)
				continue
			}

			// Check if service is ready
			for _, condition := range service.Conditions {
				if condition.Type == "Ready" && condition.State == runpb.Condition_CONDITION_SUCCEEDED {
					return nil
				}
			}
		}
	}
}
