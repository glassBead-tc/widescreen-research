package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
)

// ElicitationManager manages the elicitation process for qualifying users
type ElicitationManager struct {
	sessions map[string]*ElicitationSession
	mu       sync.RWMutex
}

// ElicitationSession represents an active elicitation session
type ElicitationSession struct {
	ID          string
	State       string
	Answers     map[string]interface{}
	StartTime   time.Time
	LastUpdated time.Time
}

// NewElicitationManager creates a new elicitation manager
func NewElicitationManager() *ElicitationManager {
	return &ElicitationManager{
		sessions: make(map[string]*ElicitationSession),
	}
}

// CreateSession creates a new elicitation session
func (em *ElicitationManager) CreateSession() string {
	em.mu.Lock()
	defer em.mu.Unlock()

	sessionID := uuid.New().String()
	em.sessions[sessionID] = &ElicitationSession{
		ID:          sessionID,
		State:       "initial",
		Answers:     make(map[string]interface{}),
		StartTime:   time.Now(),
		LastUpdated: time.Now(),
	}

	// Clean up old sessions
	go em.cleanupOldSessions()

	return sessionID
}

// GetState returns the current state of a session
func (em *ElicitationManager) GetState(sessionID string) *ElicitationSession {
	em.mu.RLock()
	defer em.mu.RUnlock()

	return em.sessions[sessionID]
}

// GetInitialQuestions returns the initial set of questions
func (em *ElicitationManager) GetInitialQuestions() []schemas.ElicitationQuestion {
	return []schemas.ElicitationQuestion{
		{
			ID:       "research_topic",
			Question: "What would you like to perform research on?",
			Type:     "text",
			Required: true,
			Metadata: map[string]interface{}{
				"placeholder": "e.g., AI safety companies, renewable energy startups, etc.",
				"multiline":   true,
			},
		},
		{
			ID:       "researcher_count",
			Question: "How many researchers do you want to provision?",
			Type:     "number",
			Required: true,
			Metadata: map[string]interface{}{
				"min":     1,
				"max":     100,
				"default": 10,
			},
		},
		{
			ID:       "research_depth",
			Question: "What level of research depth do you need?",
			Type:     "select",
			Required: true,
			Options: []schemas.ElicitationOption{
				{Value: "basic", Label: "Basic - Quick overview"},
				{Value: "standard", Label: "Standard - Comprehensive analysis"},
				{Value: "deep", Label: "Deep - Exhaustive investigation"},
			},
		},
	}
}

// ProcessAnswers processes the answers and returns next questions or completion status
func (em *ElicitationManager) ProcessAnswers(sessionID string, answers map[string]interface{}) ([]schemas.ElicitationQuestion, bool) {
	em.mu.Lock()
	defer em.mu.Unlock()

	session, exists := em.sessions[sessionID]
	if !exists {
		return nil, false
	}

	// Store answers
	for k, v := range answers {
		session.Answers[k] = v
	}
	session.LastUpdated = time.Now()

	// Determine next questions based on current state
	switch session.State {
	case "initial":
		session.State = "workflow"
		return em.getWorkflowQuestions(), false

	case "workflow":
		session.State = "advanced"
		return em.getAdvancedQuestions(session), false

	case "advanced":
		session.State = "complete"
		return nil, true

	default:
		return nil, true
	}
}

// getWorkflowQuestions returns workflow-related questions
func (em *ElicitationManager) getWorkflowQuestions() []schemas.ElicitationQuestion {
	return []schemas.ElicitationQuestion{
		{
			ID:       "workflow_templates",
			Question: "Do you have any pre-orchestrated workflows you want the researchers to use? If yes, paste them below:",
			Type:     "text",
			Required: false,
			Metadata: map[string]interface{}{
				"multiline":   true,
				"placeholder": "Paste workflow YAML or JSON here (optional)",
			},
		},
		{
			ID:       "output_format",
			Question: "What format would you like the research results in?",
			Type:     "select",
			Required: true,
			Options: []schemas.ElicitationOption{
				{Value: "structured_json", Label: "Structured JSON"},
				{Value: "markdown_report", Label: "Markdown Report"},
				{Value: "executive_summary", Label: "Executive Summary"},
				{Value: "raw_data", Label: "Raw Data"},
			},
		},
	}
}

// getAdvancedQuestions returns advanced configuration questions
func (em *ElicitationManager) getAdvancedQuestions(session *ElicitationSession) []schemas.ElicitationQuestion {
	questions := []schemas.ElicitationQuestion{
		{
			ID:       "timeout_minutes",
			Question: "Maximum time for research completion (in minutes)?",
			Type:     "number",
			Required: true,
			Metadata: map[string]interface{}{
				"min":     5,
				"max":     1440, // 24 hours
				"default": 60,
			},
		},
		{
			ID:       "priority_level",
			Question: "Research priority level?",
			Type:     "select",
			Required: true,
			Options: []schemas.ElicitationOption{
				{Value: "low", Label: "Low - Cost-optimized"},
				{Value: "normal", Label: "Normal - Balanced"},
				{Value: "high", Label: "High - Performance-optimized"},
			},
		},
	}

	// Add conditional questions based on research topic
	if topic, ok := session.Answers["research_topic"].(string); ok && topic != "" {
		questions = append(questions, schemas.ElicitationQuestion{
			ID:       "specific_sources",
			Question: fmt.Sprintf("Any specific sources or domains to focus on for '%s'?", topic),
			Type:     "text",
			Required: false,
			Metadata: map[string]interface{}{
				"placeholder": "e.g., specific websites, databases, or domains",
			},
		})
	}

	return questions
}

// GetResearchConfig builds the research configuration from session answers
func (em *ElicitationManager) GetResearchConfig(sessionID string) *schemas.ResearchConfig {
	em.mu.RLock()
	defer em.mu.RUnlock()

	session, exists := em.sessions[sessionID]
	if !exists || session.State != "complete" {
		return nil
	}

	// Build configuration from answers
	config := &schemas.ResearchConfig{
		SessionID:       sessionID,
		Topic:           em.getStringAnswer(session, "research_topic", ""),
		ResearcherCount: em.getIntAnswer(session, "researcher_count", 10),
		ResearchDepth:   em.getStringAnswer(session, "research_depth", "standard"),
		OutputFormat:    em.getStringAnswer(session, "output_format", "structured_json"),
		TimeoutMinutes:  em.getIntAnswer(session, "timeout_minutes", 60),
		PriorityLevel:   em.getStringAnswer(session, "priority_level", "normal"),
		WorkflowTemplates: em.getStringAnswer(session, "workflow_templates", ""),
		SpecificSources:  em.getStringAnswer(session, "specific_sources", ""),
		CreatedAt:       session.StartTime,
	}

	return config
}

// Helper methods

func (em *ElicitationManager) getStringAnswer(session *ElicitationSession, key string, defaultValue string) string {
	if val, ok := session.Answers[key].(string); ok {
		return val
	}
	return defaultValue
}

func (em *ElicitationManager) getIntAnswer(session *ElicitationSession, key string, defaultValue int) int {
	if val, ok := session.Answers[key].(float64); ok {
		return int(val)
	}
	if val, ok := session.Answers[key].(int); ok {
		return val
	}
	return defaultValue
}

// cleanupOldSessions removes sessions older than 1 hour
func (em *ElicitationManager) cleanupOldSessions() {
	em.mu.Lock()
	defer em.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)
	for id, session := range em.sessions {
		if session.LastUpdated.Before(cutoff) {
			delete(em.sessions, id)
		}
	}
}