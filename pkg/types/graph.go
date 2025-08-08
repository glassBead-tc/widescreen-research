package types

// EntityType enumerates core entity types for founder research.
type EntityType string

const (
	EntityPerson     EntityType = "Person"
	EntityCompany    EntityType = "Company"
	EntityInvestor   EntityType = "Investor"
	EntityYCBatch    EntityType = "YC_Batch"
	EntitySchool     EntityType = "School"
	EntityEmployer   EntityType = "PastEmployer"
	EntityHandle     EntityType = "Handle"
)

// EdgeType enumerates relation types captured in mem0 graph.
type EdgeType string

const (
	EdgeCoFounded       EdgeType = "co_founded"
	EdgeCoFoundedWith   EdgeType = "co_founded_with"
	EdgeInvestedIn      EdgeType = "invested_in"
	EdgeInvestedBy      EdgeType = "invested_by"
	EdgeAdvisorTo       EdgeType = "advisor_to"
	EdgeWorkedAt        EdgeType = "worked_at"
	EdgeStudiedAt       EdgeType = "studied_at"
	EdgeBatchMember     EdgeType = "batch_member"
	EdgeSameInvestorAs  EdgeType = "same_investor_as"
	EdgeSameCofounderAs EdgeType = "same_cofounder_as"
)

// Entity represents a node to be persisted in mem0.
type Entity struct {
	ID   string     `json:"id"`
	Type EntityType `json:"type"`
	Name string     `json:"name"`
	Props map[string]any `json:"props,omitempty"`
}

// Triple represents a graph edge (subject-predicate-object).
type Triple struct {
	SubjectID string   `json:"subject_id"`
	Predicate EdgeType `json:"predicate"`
	ObjectID  string   `json:"object_id"`
	Citations []string `json:"citations,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

// MemoryRecord captures narrative plus structured graph for a subject.
type MemoryRecord struct {
	SubjectID string   `json:"subject_id"`
	Summary   string   `json:"summary"`
	Citations []string `json:"citations"`
	Entities  []Entity `json:"entities"`
	Triples   []Triple `json:"triples"`
}