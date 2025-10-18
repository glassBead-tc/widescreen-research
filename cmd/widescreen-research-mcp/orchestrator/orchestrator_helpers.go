// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	runpb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/schemas"
)

// Helper methods for orchestrator

// loadTemplates loads research templates
func (o *Orchestrator) loadTemplates() {
	// Load default templates
	o.templates["company-research"] = &ResearchTemplate{
		ID:          "company-research",
		Name:        "Company Research Template",
		Description: "Template for researching companies and organizations",
		Workflow: map[string]interface{}{
			"steps": []string{
				"company_overview",
				"financial_data",
				"competitor_analysis",
				"market_position",
			},
		},
	}

	o.templates["academic-research"] = &ResearchTemplate{
		ID:          "academic-research",
		Name:        "Academic Research Template",
		Description: "Template for academic and scientific research",
		Workflow: map[string]interface{}{
			"steps": []string{
				"literature_review",
				"methodology_analysis",
				"data_collection",
				"peer_review",
			},
		},
	}
}

// createPubSubTopics creates required Pub/Sub topics
func (o *Orchestrator) createPubSubTopics(ctx context.Context) error {
	// Create main orchestrator topics
	topics := []string{
		"research-commands",
		"research-status",
		"research-metrics",
	}

	for _, topicName := range topics {
		topic := o.pubsubClient.Topic(topicName)
		exists, err := topic.Exists(ctx)
		if err != nil {
			return fmt.Errorf("failed to check topic %s: %w", topicName, err)
		}

		if !exists {
			_, err = o.pubsubClient.CreateTopic(ctx, topicName)
			if err != nil {
				return fmt.Errorf("failed to create topic %s: %w", topicName, err)
			}
			log.Printf("Created Pub/Sub topic: %s", topicName)
		}
	}

	return nil
}

// monitorSession monitors the health of a research session
func (o *Orchestrator) monitorSession(ctx context.Context, session *ResearchSession) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			o.mu.RLock()
			drones := make([]*DroneInfo, 0, len(session.Drones))
			for _, drone := range session.Drones {
				drones = append(drones, drone)
			}
			o.mu.RUnlock()

			// Check drone health
			for _, drone := range drones {
				if err := o.checkDroneHealth(ctx, drone); err != nil {
					log.Printf("Drone %s health check failed: %v", drone.ID, err)
					drone.Status = "unhealthy"
				}
			}

			// Check for session timeout
			if time.Since(session.StartTime) > time.Duration(session.Config.TimeoutMinutes)*time.Minute {
				log.Printf("Session %s timed out", session.Config.SessionID)
				session.Status = "timeout"
				return
			}
		}
	}
}

// checkDroneHealth checks the health of a drone
func (o *Orchestrator) checkDroneHealth(ctx context.Context, drone *DroneInfo) error {
	// Make HTTP health check request
	healthURL := fmt.Sprintf("%s/health", drone.ServiceURL)

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	drone.LastCheckin = time.Now()
	return nil
}

// sendInstructionsToDrone sends research instructions to a drone
func (o *Orchestrator) sendInstructionsToDrone(ctx context.Context, drone *DroneInfo, task map[string]interface{}) error {
	// Create command message
	command := map[string]interface{}{
		"type":         "research_command",
		"instructions": task,
		"timestamp":    time.Now(),
	}

	// Send via HTTP POST to drone
	instructURL := fmt.Sprintf("%s/instructions", drone.ServiceURL)

	jsonData, err := json.Marshal(command)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", instructURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send instructions, status: %d", resp.StatusCode)
	}

	return nil
}

// collectResults collects results from the research queue
func (o *Orchestrator) collectResults(ctx context.Context, session *ResearchSession) {
	// Subscribe to results queue
	if err := session.Queue.Subscribe(ctx, o.pubsubClient); err != nil {
		log.Printf("Failed to subscribe to results queue: %v", err)
		return
	}

	// Process results as they arrive
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-session.Queue.ResultChannel():
			o.mu.Lock()
			session.Results = append(session.Results, result)
			if drone, ok := session.Drones[result.DroneID]; ok {
				drone.Status = result.Status
			}
			o.mu.Unlock()

			log.Printf("Collected result from drone %s with status %s", result.DroneID, result.Status)

			// Update progress file
			if err := o.updateProgressFile(session); err != nil {
				log.Printf("Warning: failed to update progress file for session %s: %v", session.Config.SessionID, err)
			}

		case err := <-session.Queue.ErrorChannel():
			log.Printf("Queue error: %v", err)
		}
	}
}

// calculateMetrics calculates final metrics for the research session
func (o *Orchestrator) calculateMetrics(session *ResearchSession) schemas.ResearchMetrics {
	metrics := schemas.ResearchMetrics{
		DronesProvisioned:   len(session.Drones),
		DronesCompleted:     0,
		DronesFailed:        0,
		TotalDuration:       time.Since(session.StartTime),
		DataPointsCollected: 0,
		CostEstimate:        0.0,
	}

	// Calculate from results
	for _, result := range session.Results {
		if result.Status == "completed" {
			metrics.DronesCompleted++
			metrics.DataPointsCollected += len(result.Data)
		} else {
			metrics.DronesFailed++
		}
	}

	// Estimate costs based on Cloud Run pricing
	cpuHours := float64(metrics.DronesProvisioned) * metrics.TotalDuration.Hours()
	metrics.CostEstimate = cpuHours * 0.0000024 * 1000 // Approximate cost per vCPU-ms

	return metrics
}

// updateProgressFile writes the current session progress to a markdown file.
func (o *Orchestrator) updateProgressFile(session *ResearchSession) error {
	// Ensure the reports directory exists.
	if err := os.MkdirAll("reports", 0755); err != nil {
		return fmt.Errorf("failed to create reports directory: %w", err)
	}

	filePath := fmt.Sprintf("reports/progress_%s.md", session.Config.SessionID)

	var content strings.Builder
	content.WriteString(fmt.Sprintf("# Research Progress: %s\n\n", session.Config.Topic))
	content.WriteString(fmt.Sprintf("**Session ID:** `%s`\n", session.Config.SessionID))
	content.WriteString(fmt.Sprintf("**Overall Status:** `%s`\n", session.Status))
	content.WriteString(fmt.Sprintf("**Last Updated:** %s\n\n", time.Now().Format(time.RFC1123)))

	content.WriteString("## Drone Status\n\n")
	content.WriteString("| Drone ID | Status |\n")
	content.WriteString("|---|---|\n")

	o.mu.RLock()
	for id, drone := range session.Drones {
		content.WriteString(fmt.Sprintf("| %s | %s |\n", id, drone.Status))
	}
	o.mu.RUnlock()

	// Add results summary
	content.WriteString(fmt.Sprintf("\n**Results Collected:** %d / %d\n", len(session.Results), len(session.Drones)))

	return os.WriteFile(filePath, []byte(content.String()), 0644)
}

// cleanupSession cleans up resources after a research session
func (o *Orchestrator) cleanupSession(ctx context.Context, session *ResearchSession) {
	log.Printf("Cleaning up session %s", session.Config.SessionID)

	// Delete Cloud Run services
	for _, drone := range session.Drones {
		if err := o.deleteDroneService(ctx, drone.ID); err != nil {
			log.Printf("Failed to delete drone service %s: %v", drone.ID, err)
		}
	}

	// Delete Pub/Sub resources
	topicName := fmt.Sprintf("research-results-%s", session.Config.SessionID)
	topic := o.pubsubClient.Topic(topicName)
	if err := topic.Delete(ctx); err != nil {
		log.Printf("Failed to delete topic %s: %v", topicName, err)
	}

	// Close queue
	session.Queue.Close()

	// Remove from active sessions
	o.mu.Lock()
	delete(o.activeSessions, session.Config.SessionID)
	o.mu.Unlock()
}

// deleteDroneService deletes a drone Cloud Run service
func (o *Orchestrator) deleteDroneService(ctx context.Context, droneID string) error {
	req := &runpb.DeleteServiceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/services/%s", o.projectID, o.region, droneID),
	}

	operation, err := o.runClient.DeleteService(ctx, req)
	if err != nil {
		return err
	}

	// Wait for deletion to complete
	_, err = operation.Wait(ctx)
	return err
}
