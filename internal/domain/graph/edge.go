package graph

// Relation describes the nature of causality between two nodes.
type Relation string

const (
	RelationCaused      Relation = "CAUSED"
	RelationPrevented   Relation = "PREVENTED"
	RelationCorrelated  Relation = "CORRELATED"
	RelationUnknown     Relation = "UNKNOWN"
)

// Polarity records whether the causal relationship is beneficial or harmful.
type Polarity string

const (
	PolarityPositive Polarity = "POSITIVE"
	PolarityNegative Polarity = "NEGATIVE"
)

// Edge is a directed causal relationship between two nodes.
// Confidence grows with each WorkoutLog that provides evidence.
// EvidenceLog is the audit trail — every log ID that supports this edge.
type Edge struct {
	ID          string   `json:"id"`
	FromNode    string   `json:"from_node"`
	ToNode      string   `json:"to_node"`
	Relation    Relation `json:"relation"`
	Confidence  float32  `json:"confidence"`    // 0.0–1.0
	Polarity    Polarity `json:"polarity"`
	EvidenceLog []string `json:"evidence_log"` // workout log IDs
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
}
