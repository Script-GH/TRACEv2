package graph

// NodeType classifies what kind of thing a node represents.
type NodeType string

const (
	NodeTypeExercise  NodeType = "EXERCISE"
	NodeTypePattern   NodeType = "PATTERN"
	NodeTypeOutcome   NodeType = "OUTCOME"
	NodeTypeState     NodeType = "STATE"
	NodeTypeNutrition NodeType = "NUTRITION"
)

// NodeSource records where a node was created.
type NodeSource string

const (
	NodeSourceInterview  NodeSource = "interview"
	NodeSourceWorkoutLog NodeSource = "workout_log"
	NodeSourceInference  NodeSource = "inference"
)

// Node is a vertex in the personal causal graph.
// It represents a concrete event, state, or pattern from the user's history.
type Node struct {
	ID          string            `json:"id"`
	Type        NodeType          `json:"type"`
	Label       string            `json:"label"`        // human-readable, e.g. "High volume squats"
	Attributes  map[string]any    `json:"attributes"`
	CreatedFrom NodeSource        `json:"created_from"`
	CreatedAt   int64             `json:"created_at"` // unix nano
}
