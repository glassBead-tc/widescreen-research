package drone

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
)

// ResearcherDrone represents a research-focused drone MCP server
type ResearcherDrone struct {
	droneID        string
	coordinatorURL string
	taskID         string
	pubsubClient   *pubsub.Client
	pubsubTopic    *pubsub.Topic
}

// NewResearcherDrone creates a new researcher drone MCP server
func NewResearcherDrone() (*ResearcherDrone, error) {
	ctx := context.Background()
	// Get configuration from environment
	droneID := os.Getenv("DRONE_ID")
	if droneID == "" {
		return nil, fmt.Errorf("DRONE_ID environment variable is required")
	}

	coordinatorURL := os.Getenv("COORDINATOR_URL")
	taskID := os.Getenv("TASK_ID")

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_PROJECT environment variable is required")
	}

	topicID := os.Getenv("PUBSUB_TOPIC")
	if topicID == "" {
		return nil, fmt.Errorf("PUBSUB_TOPIC environment variable is required")
	}

	pubsubClient, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	topic := pubsubClient.Topic(topicID)

	drone := &ResearcherDrone{
		droneID:        droneID,
		coordinatorURL: coordinatorURL,
		taskID:         taskID,
		pubsubClient:   pubsubClient,
		pubsubTopic:    topic,
	}

	return drone, nil
}

// ConductResearch performs research on a given topic
func (d *ResearcherDrone) ConductResearch(topic, timeFrame string, sources []string, maxResults int) (map[string]interface{}, error) {
	log.Printf("Drone %s conducting research on: %s", d.droneID, topic)

	// Simulate research process
	time.Sleep(2 * time.Second) // Simulate research time

	// Mock research results
	results := map[string]interface{}{
		"topic":     topic,
		"timeFrame": timeFrame,
		"findings": []map[string]interface{}{
			{
				"title":       fmt.Sprintf("Research finding 1 for %s", topic),
				"description": "Detailed analysis of the topic with supporting evidence",
				"relevance":   0.95,
				"sources":     []string{"Academic Source 1", "Historical Archive 2"},
			},
			{
				"title":       fmt.Sprintf("Research finding 2 for %s", topic),
				"description": "Additional insights and cross-references",
				"relevance":   0.87,
				"sources":     []string{"Primary Document 3", "Expert Analysis 4"},
			},
		},
		"summary":    fmt.Sprintf("Comprehensive research completed on %s within timeframe %s", topic, timeFrame),
		"confidence": 0.89,
		"droneId":    d.droneID,
		"timestamp":  time.Now(),
	}

	return results, nil
}

// AnalyzeHistoricalPeriod analyzes events in a specific historical period
func (d *ResearcherDrone) AnalyzeHistoricalPeriod(startYear, endYear int, regions, eventTypes []string) (map[string]interface{}, error) {
	log.Printf("Drone %s analyzing period %d-%d", d.droneID, startYear, endYear)

	// Simulate historical analysis
	time.Sleep(3 * time.Second)

	// Mock historical events
	events := []map[string]interface{}{
		{
			"year":        startYear + 2,
			"title":       fmt.Sprintf("Major Event in %d", startYear+2),
			"description": "Significant political development with lasting impact",
			"importance":  9.2,
			"category":    "political",
			"regions":     regions,
		},
		{
			"year":        startYear + 5,
			"title":       fmt.Sprintf("Economic Shift in %d", startYear+5),
			"description": "Important economic transformation affecting global markets",
			"importance":  8.7,
			"category":    "economic",
			"regions":     regions,
		},
		{
			"year":        startYear + 8,
			"title":       fmt.Sprintf("Social Movement in %d", startYear+8),
			"description": "Influential social movement that changed societal norms",
			"importance":  8.1,
			"category":    "social",
			"regions":     regions,
		},
	}

	analysis := map[string]interface{}{
		"period":    fmt.Sprintf("%d-%d", startYear, endYear),
		"events":    events,
		"summary":   fmt.Sprintf("Analyzed %d significant events in the period %d-%d", len(events), startYear, endYear),
		"trends":    []string{"Political instability", "Economic growth", "Social reform"},
		"droneId":   d.droneID,
		"timestamp": time.Now(),
	}

	return analysis, nil
}

// Serve starts the drone MCP server
func (d *ResearcherDrone) Serve() error {
	log.Printf("Starting Researcher Drone %s...", d.droneID)
	// For now, just keep running
	select {}
}

// Close closes the drone and cleans up resources
func (d *ResearcherDrone) Close() error {
	if d.pubsubClient != nil {
		d.pubsubClient.Close()
	}
	return nil
}

// publishResult publishes the research result to the Pub/Sub topic.
func (d *ResearcherDrone) publishResult(ctx context.Context, resultData map[string]interface{}) error {
	// We need to wrap the raw result data in the DroneResult schema
	// to be consistent with what the orchestrator expects.
	result := schemas.DroneResult{
		DroneID:        d.droneID,
		Status:         "success", // Assuming success if this method is called
		Data:           resultData,
		CompletedAt:    time.Now(),
		ProcessingTime: 0, // This can be properly calculated in the http worker
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	msg := &pubsub.Message{
		Data: jsonData,
	}

	if _, err := d.pubsubTopic.Publish(ctx, msg).Get(ctx); err != nil {
		return fmt.Errorf("failed to publish result: %w", err)
	}

	log.Printf("Drone %s published result to topic %s", d.droneID, d.pubsubTopic.String())
	return nil
}
