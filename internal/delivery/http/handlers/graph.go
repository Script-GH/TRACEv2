package handlers

import (
	"encoding/json"
	"math"
	"net/http"

	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/delivery/http/middleware"
	"github.com/trace/trace/internal/domain/graph"
)

// GraphHandler serves the causal graph data for Screen 3.
type GraphHandler struct {
	mcp ports.MCPClient
}

func NewGraphHandler(mcp ports.MCPClient) *GraphHandler {
	return &GraphHandler{mcp: mcp}
}

// GraphNodeDTO is a serialisation-friendly node for the mobile client.
type GraphNodeDTO struct {
	ID    string         `json:"id"`
	Type  string         `json:"type"`
	Label string         `json:"label"`
	X     float64        `json:"x"`
	Y     float64        `json:"y"`
	Attrs map[string]any `json:"attributes,omitempty"`
}

// GraphEdgeDTO is a serialisation-friendly edge for the mobile client.
type GraphEdgeDTO struct {
	ID         string  `json:"id"`
	FromNode   string  `json:"from_node"`
	ToNode     string  `json:"to_node"`
	Relation   string  `json:"relation"`
	Polarity   string  `json:"polarity"`
	Confidence float32 `json:"confidence"`
}

// GraphResponseDTO wraps nodes + edges with layout coordinates computed server-side.
type GraphResponseDTO struct {
	UserID      string         `json:"user_id"`
	Version     int            `json:"version"`
	Nodes       []GraphNodeDTO `json:"nodes"`
	Edges       []GraphEdgeDTO `json:"edges"`
	LastUpdated int64          `json:"last_updated"`
}

// GET /api/v1/graph
// Returns the full causal graph for the authenticated user, with layout coordinates.
func (h *GraphHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	domainFilter := r.URL.Query().Get("domain")

	g, err := h.mcp.GetCausalGraph(r.Context(), userID, domainFilter)
	if err != nil {
		writeError(w, http.StatusNotFound, "graph not yet built — complete the interview first")
		return
	}

	resp := buildGraphResponse(g)
	writeJSON(w, http.StatusOK, resp)
}

// GET /api/v1/graph/paths
// Queries for causal paths (used for "why this exercise?" drilldown).
func (h *GraphHandler) QueryPaths(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	q := ports.CausalPathQuery{
		FromNodeLabel: r.URL.Query().Get("from"),
		ToNodeLabel:   r.URL.Query().Get("to"),
		MinConfidence: 0.5,
	}

	paths, err := h.mcp.QueryCausalPaths(r.Context(), userID, q)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, paths)
}

// GET /api/v1/graph/progress — stub for Screen 4.
func (h *GraphHandler) Progress(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "coming soon"})
}

// buildGraphResponse converts a domain graph into a DTO with simple radial layout.
func buildGraphResponse(g *graph.CausalGraph) GraphResponseDTO {
	nodes := make([]GraphNodeDTO, 0, len(g.Nodes))
	i := 0
	total := len(g.Nodes)
	for _, n := range g.Nodes {
		angle := 0.0
		if total > 1 {
			angle = 2 * 3.14159 * float64(i) / float64(total)
		}
		radius := 200.0
		x := 250.0 + radius*cos(angle)
		y := 250.0 + radius*sin(angle)
		nodes = append(nodes, GraphNodeDTO{
			ID:    n.ID,
			Type:  string(n.Type),
			Label: n.Label,
			X:     x,
			Y:     y,
			Attrs: n.Attributes,
		})
		i++
	}

	edges := make([]GraphEdgeDTO, 0, len(g.Edges))
	for _, e := range g.Edges {
		edges = append(edges, GraphEdgeDTO{
			ID:         e.ID,
			FromNode:   e.FromNode,
			ToNode:     e.ToNode,
			Relation:   string(e.Relation),
			Polarity:   string(e.Polarity),
			Confidence: e.Confidence,
		})
	}

	return GraphResponseDTO{
		UserID:      g.UserID,
		Version:     g.Version,
		Nodes:       nodes,
		Edges:       edges,
		LastUpdated: g.LastUpdated,
	}
}

func cos(a float64) float64 { return math.Cos(a) }
func sin(a float64) float64 { return math.Sin(a) }

// shared helpers used across all handlers in this package.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
