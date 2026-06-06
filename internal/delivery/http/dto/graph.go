package dto

import "github.com/trace/trace/internal/domain/graph"

// Graph DTOs — used by Screen 3 (causal graph visual) and Screen 4 (progress).

type GetGraphResponse struct {
	Graph        *graph.CausalGraph `json:"graph"`
	ConfidenceScore float32         `json:"confidence_score"` // average confidence across all edges
	UpdatedEdges []string           `json:"updated_edges"`    // edge IDs updated in last 24h
}

type GetGraphPathsRequest struct {
	UserID        string  `json:"user_id"`
	ToNodePattern string  `json:"to_node_pattern"` // e.g. "knee_pain"
	MinConfidence float32 `json:"min_confidence"`
}

type GetGraphPathsResponse struct {
	Paths []GraphPathView `json:"paths"`
}

// GraphPathView is the UI-friendly representation of a causal path.
type GraphPathView struct {
	Nodes      []NodeView `json:"nodes"`
	Edges      []EdgeView `json:"edges"`
	TotalConfidence float32 `json:"total_confidence"`
	EvidenceCount   int    `json:"evidence_count"` // total unique workout logs supporting this path
}

type NodeView struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

type EdgeView struct {
	ID         string  `json:"id"`
	FromLabel  string  `json:"from_label"`
	ToLabel    string  `json:"to_label"`
	Relation   string  `json:"relation"`
	Confidence float32 `json:"confidence"`
	Polarity   string  `json:"polarity"`
}

// ProgressSummaryResponse is for Screen 4.
type ProgressSummaryResponse struct {
	GoalOnTrack      bool             `json:"goal_on_track"`
	PredictionAccuracy float32        `json:"prediction_accuracy"` // hit/miss rate last 7 workouts
	KeyInsights      []string         `json:"key_insights"`        // plain-English top 3 patterns
	TrendFlags       []TrendFlag      `json:"trend_flags"`
}

type TrendFlag struct {
	Message  string `json:"message"`  // e.g. "Knee pain in 3 of last 5 sessions"
	Severity string `json:"severity"` // "info" | "warning" | "critical"
}
