package services

import (
	"context"
	"fmt"

	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
)

// ElicitationState represents the state of an elicitation session
type ElicitationState struct {
	SessionID   string
	CurrentStep int
	Answers     map[string]interface{}
	Complete    bool
}

// ElicitationService manages elicitation sessions
type ElicitationService struct {
	sessions map[string]*ElicitationState
}

// NewElicitationService creates a new elicitation service
func NewElicitationService() *ElicitationService {
	return &ElicitationService{
		sessions: make(map[string]*ElicitationState),
	}
}

// HandleElicitation processes elicitation requests
func (s *ElicitationService) HandleElicitation(ctx context.Context, input *schemas.WidescreenResearchInput) (interface{}, error) {
	// Check current elicitation state
	state := s.GetState(input.SessionID)
	
	if state == nil {
		// Start new elicitation
		questions := s.GetInitialQuestions()
		sessionID := s.CreateSession()
		return &schemas.ElicitationResponse{
			Type:      "elicitation",
			Questions: questions,
			SessionID: sessionID,
		}, nil
	}
	
	// Process answers and get next questions
	state.ProcessAnswers(input.ElicitationAnswers)
	
	if state.Complete {
		// Elicitation complete, prepare operation parameters
		return map[string]interface{}{
			"type":       "complete",
			"sessionId":  state.SessionID,
			"parameters": state.GetFinalParameters(),
		}, nil
	}
	
	// Get next questions
	questions := s.GetNextQuestions(state)
	return &schemas.ElicitationResponse{
		Type:      "elicitation",
		Questions: questions,
		SessionID: state.SessionID,
	}, nil
}

// GetState retrieves the state for a session
func (s *ElicitationService) GetState(sessionID string) *ElicitationState {
	return s.sessions[sessionID]
}

// CreateSession creates a new elicitation session
func (s *ElicitationService) CreateSession() string {
	sessionID := fmt.Sprintf("session_%d", len(s.sessions)+1)
	s.sessions[sessionID] = &ElicitationState{
		SessionID:   sessionID,
		CurrentStep: 0,
		Answers:     make(map[string]interface{}),
		Complete:    false,
	}
	return sessionID
}

// GetInitialQuestions returns the initial set of questions
func (s *ElicitationService) GetInitialQuestions() []schemas.ElicitationQuestion {
	return []schemas.ElicitationQuestion{
		{
			ID:       "research_scope",
			Question: "What is the scope of your research? (e.g., specific topic, company, technology)",
			Type:     "text",
		},
		{
			ID:       "depth",
			Question: "How deep should the research go?",
			Type:     "choice",
			Options: []schemas.ElicitationOption{
				{Value: "quick", Label: "Quick overview"},
				{Value: "moderate", Label: "Moderate analysis"},
				{Value: "deep", Label: "Deep investigation"},
			},
		},
		{
			ID:       "time_constraint",
			Question: "What is your time constraint?",
			Type:     "choice",
			Options: []schemas.ElicitationOption{
				{Value: "asap", Label: "ASAP (10-15 min)"},
				{Value: "standard", Label: "Standard (30 min)"},
				{Value: "thorough", Label: "Thorough (60+ min)"},
			},
		},
	}
}

// GetNextQuestions returns the next set of questions based on state
func (s *ElicitationService) GetNextQuestions(state *ElicitationState) []schemas.ElicitationQuestion {
	// This would be more sophisticated in production
	return []schemas.ElicitationQuestion{
		{
			ID:       "output_format",
			Question: "What format would you like the results in?",
			Type:     "choice",
			Options: []schemas.ElicitationOption{
				{Value: "summary", Label: "Summary report"},
				{Value: "detailed", Label: "Detailed analysis"},
				{Value: "raw", Label: "Raw data"},
			},
		},
	}
}

// ProcessAnswers processes the provided answers
func (state *ElicitationState) ProcessAnswers(answers map[string]interface{}) {
	for k, v := range answers {
		state.Answers[k] = v
	}
	state.CurrentStep++
	
	// Simple completion check - would be more sophisticated in production
	if state.CurrentStep >= 2 {
		state.Complete = true
	}
}

// GetFinalParameters returns the final operation parameters
func (state *ElicitationState) GetFinalParameters() map[string]interface{} {
	// Convert elicitation answers to operation parameters
	params := make(map[string]interface{})
	
	// Map answers to parameters
	if scope, ok := state.Answers["research_scope"]; ok {
		params["topic"] = scope
	}
	
	// Set result count based on depth
	if depth, ok := state.Answers["depth"].(string); ok {
		switch depth {
		case "Quick overview":
			params["result_count"] = 10
		case "Moderate analysis":
			params["result_count"] = 30
		case "Deep investigation":
			params["result_count"] = 50
		}
	}
	
	return params
}