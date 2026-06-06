package graph

// CausalGraph is the central data structure of TRACE.
// It is the personal causal model of a single user's body, built from their
// reported history and updated after every workout log.
//
// Stored in Redis as a versioned JSON document.
// Updates are append-only — version increments on every write.
// All decisions made by the MoE agents are traceable to edges in this graph.
type CausalGraph struct {
	UserID       string           `json:"user_id"`
	Version      int              `json:"version"`       // monotonic, incremented on each TwinAgent update
	Nodes        map[string]*Node `json:"nodes"`         // keyed by Node.ID
	Edges        map[string]*Edge `json:"edges"`         // keyed by Edge.ID
	CreatedAt    int64            `json:"created_at"`
	LastUpdated  int64            `json:"last_updated"`
}

// CausalPath is a resolved sequence of nodes + edges from a graph query.
// Used by InjuryAgent and GrowthAgent to find chains leading to outcomes.
type CausalPath struct {
	Nodes      []*Node  `json:"nodes"`
	Edges      []*Edge  `json:"edges"`
	TotalConfidence float32 `json:"total_confidence"` // product of edge confidences along path
}

// InjuryConstraint is a pre-computed constraint derived from the graph.
// Materialized by the MCP server so InjuryAgent doesn't need to traverse the graph each time.
type InjuryConstraint struct {
	BodyPart       string   `json:"body_part"`
	Status         string   `json:"status"`       // "active" | "chronic" | "resolved"
	ContraExercises []string `json:"contra_exercises"` // exercises with negative causal edges ≥ threshold
	LoadCap        float32  `json:"load_cap"`     // 0–1 relative load ceiling
	Triggers       []string `json:"triggers"`
}
