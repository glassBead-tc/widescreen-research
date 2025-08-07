package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
)

// ResearchQueue manages the queue for collecting research results
type ResearchQueue struct {
	sessionID     string
	subscription  *pubsub.Subscription
	results       []schemas.DroneResult
	mu            sync.Mutex
	resultChan    chan schemas.DroneResult
	errorChan     chan error
}

// NewResearchQueue creates a new research queue
func NewResearchQueue(sessionID string) *ResearchQueue {
	return &ResearchQueue{
		sessionID:  sessionID,
		results:    make([]schemas.DroneResult, 0),
		resultChan: make(chan schemas.DroneResult, 100),
		errorChan:  make(chan error, 10),
	}
}

// Subscribe subscribes to the results topic
func (q *ResearchQueue) Subscribe(ctx context.Context, client *pubsub.Client) error {
	topicName := fmt.Sprintf("research-results-%s", q.sessionID)
	topic := client.Topic(topicName)

	// Create topic if it doesn't exist
	exists, err := topic.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check topic existence: %w", err)
	}
	if !exists {
		topic, err = client.CreateTopic(ctx, topicName)
		if err != nil {
			return fmt.Errorf("failed to create topic: %w", err)
		}
	}

	// Create subscription
	subscriptionName := fmt.Sprintf("research-results-sub-%s", q.sessionID)
	q.subscription = client.Subscription(subscriptionName)

	exists, err = q.subscription.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check subscription existence: %w", err)
	}
	if !exists {
		q.subscription, err = client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic:                 topic,
			AckDeadline:           30 * time.Second,
			RetentionDuration:     24 * time.Hour,
			ExpirationPolicy:      25 * time.Hour,
			EnableMessageOrdering: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	// Start receiving messages
	go q.receiveMessages(ctx)

	return nil
}

// receiveMessages receives messages from the subscription
func (q *ResearchQueue) receiveMessages(ctx context.Context) {
	err := q.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		// Parse the message
		var result schemas.DroneResult
		if err := json.Unmarshal(msg.Data, &result); err != nil {
			q.errorChan <- fmt.Errorf("failed to unmarshal result: %w", err)
			msg.Nack()
			return
		}

		// Add to results
		q.mu.Lock()
		q.results = append(q.results, result)
		q.mu.Unlock()

		// Send to channel
		select {
		case q.resultChan <- result:
		default:
			// Channel full, log warning
		}

		// Acknowledge the message
		msg.Ack()
	})

	if err != nil {
		q.errorChan <- fmt.Errorf("subscription receive error: %w", err)
	}
}

// GetResults returns all collected results
func (q *ResearchQueue) GetResults() []schemas.DroneResult {
	q.mu.Lock()
	defer q.mu.Unlock()

	results := make([]schemas.DroneResult, len(q.results))
	copy(results, q.results)
	return results
}

// GetResultCount returns the number of results collected
func (q *ResearchQueue) GetResultCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.results)
}

// ResultChannel returns the channel for receiving results
func (q *ResearchQueue) ResultChannel() <-chan schemas.DroneResult {
	return q.resultChan
}

// ErrorChannel returns the channel for receiving errors
func (q *ResearchQueue) ErrorChannel() <-chan error {
	return q.errorChan
}

// Close closes the queue and cleans up resources
func (q *ResearchQueue) Close() {
	close(q.resultChan)
	close(q.errorChan)
}