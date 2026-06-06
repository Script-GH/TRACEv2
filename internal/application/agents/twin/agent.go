package twin

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/domain/graph"
)

// Agent implements agents.TwinAgent.
// It is the sole writer to the causal graph.
// LLM is used only for initial graph seeding — deterministic updates happen without it.
type Agent struct {
	mcp ports.MCPClient
	llm ports.LLMClient
}

func New(mcp ports.MCPClient, llm ports.LLMClient) *Agent {
	return &Agent{mcp: mcp, llm: llm}
}

// causalEdgeExtraction is what Gemini returns for each causal relationship.
type causalEdgeExtraction struct {
	FromLabel    string  `json:"from_label"`
	ToLabel      string  `json:"to_label"`
	Relation     string  `json:"relation"`
	Polarity     string  `json:"polarity"`
	Confidence   float32 `json:"confidence"`
	NodeTypeFrom string  `json:"node_type_from"`
	NodeTypeTo   string  `json:"node_type_to"`
}

// BuildInitial constructs the first causal graph from the user's completed profile.
// Uses Gemini to infer causal relationships from the user's self-reported history,
// then saves the resulting graph via MCPClient.
func (a *Agent) BuildInitial(ctx context.Context, userID string) (int, error) {
	profile, err := a.mcp.GetUserProfile(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("twin: get profile: %w", err)
	}

	profileJSON, _ := json.MarshalIndent(profile, "", "  ")

	var edges []causalEdgeExtraction
	prompt := buildGraphExtractionPrompt(string(profileJSON))
	if err := a.llm.CompleteJSON(ctx, prompt, &edges); err != nil {
		// Fall back to a minimal deterministic graph if Gemini fails.
		edges = deterministicEdges(profile)
	}

	now := time.Now()
	g := &graph.CausalGraph{
		UserID:    userID,
		Version:   1,
		Nodes:     make(map[string]*graph.Node),
		Edges:     make(map[string]*graph.Edge),
		CreatedAt: now.UnixNano(),
	}

	for _, ext := range edges {
		fromID := labelToID(ext.FromLabel)
		toID := labelToID(ext.ToLabel)

		if _, ok := g.Nodes[fromID]; !ok {
			g.Nodes[fromID] = &graph.Node{
				ID:          fromID,
				Type:        parseNodeType(ext.NodeTypeFrom),
				Label:       ext.FromLabel,
				Attributes:  map[string]any{},
				CreatedFrom: graph.NodeSourceInterview,
				CreatedAt:   now.UnixNano(),
			}
		}
		if _, ok := g.Nodes[toID]; !ok {
			g.Nodes[toID] = &graph.Node{
				ID:          toID,
				Type:        parseNodeType(ext.NodeTypeTo),
				Label:       ext.ToLabel,
				Attributes:  map[string]any{},
				CreatedFrom: graph.NodeSourceInterview,
				CreatedAt:   now.UnixNano(),
			}
		}

		edgeID := uuid.New().String()
		g.Edges[edgeID] = &graph.Edge{
			ID:          edgeID,
			FromNode:    fromID,
			ToNode:      toID,
			Relation:    parseRelation(ext.Relation),
			Confidence:  clamp(ext.Confidence, 0.5, 0.95),
			Polarity:    parsePolarity(ext.Polarity),
			EvidenceLog: []string{"interview"},
			CreatedAt:   now.UnixNano(),
			UpdatedAt:   now.UnixNano(),
		}
	}

	// If we got no edges, create a placeholder so the graph isn't empty.
	if len(g.Edges) == 0 {
		nodeID := labelToID(profile.Goals.Primary + " training")
		g.Nodes[nodeID] = &graph.Node{
			ID: nodeID, Type: graph.NodeTypePattern,
			Label: profile.Goals.Primary + " training",
			Attributes: map[string]any{}, CreatedFrom: graph.NodeSourceInterview,
			CreatedAt: now.UnixNano(),
		}
	}

	newVersion, err := a.mcp.UpdateCausalGraph(ctx, userID, g)
	if err != nil {
		return 0, fmt.Errorf("twin: save graph: %w", err)
	}

	// Update the profile's graph version reference.
	profile.GraphVersion = newVersion
	_ = a.mcp.SaveUserProfile(ctx, profile) // best-effort

	return newVersion, nil
}

// Update processes a new workout log against the existing graph.
// For MVP: stubs out — returns 0 without modifying the graph.
func (a *Agent) Update(ctx context.Context, userID string, logID string) (int, error) {
	g, err := a.mcp.GetCausalGraph(ctx, userID, "")
	if err != nil {
		return 0, err
	}
	return g.Version, nil
}

// ── helpers ───────────────────────────────────────────────────────────────

func buildGraphExtractionPrompt(profileJSON string) string {
	return fmt.Sprintf(`You are building a personal causal knowledge graph for a fitness user.

User Profile:
%s

Identify ALL causal relationships this user has experienced. For each one extract:
- from_label: the cause (exercise, program, behavior, nutrition pattern)
- to_label: the effect (outcome, injury, strength gain, body composition change)
- relation: "CAUSED" or "PREVENTED" or "CORRELATED"
- polarity: "POSITIVE" (beneficial) or "NEGATIVE" (harmful)
- confidence: 0.50-0.95 based on how explicitly the user stated it
- node_type_from: "EXERCISE" or "PATTERN" or "STATE" or "NUTRITION"
- node_type_to: "OUTCOME" or "STATE" or "PATTERN"

Examples:
- Programs that produced results → CAUSED POSITIVE OUTCOME
- Programs that caused injuries → CAUSED NEGATIVE OUTCOME
- Behaviors that prevented injuries → PREVENTED NEGATIVE STATE
- Recovery patterns that correlated with performance → CORRELATED POSITIVE STATE

Also add nodes for:
- Current goals (as OUTCOME nodes)
- Current injuries (as STATE nodes with NEGATIVE polarity)
- Equipment available (as PATTERN nodes)

Return ONLY a JSON array of edges. If the profile has little history, return 2-3 seed edges based on the goal.`, profileJSON)
}

// deterministicEdges returns an empty set — graph will have goal node only.
func deterministicEdges(_ interface{}) []causalEdgeExtraction {
	return nil
}

func labelToID(label string) string {
	s := strings.ToLower(label)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, "/", "_")
	return s
}

func parseNodeType(s string) graph.NodeType {
	switch strings.ToUpper(s) {
	case "EXERCISE":
		return graph.NodeTypeExercise
	case "PATTERN":
		return graph.NodeTypePattern
	case "OUTCOME":
		return graph.NodeTypeOutcome
	case "NUTRITION":
		return graph.NodeTypeNutrition
	default:
		return graph.NodeTypeState
	}
}

func parseRelation(s string) graph.Relation {
	switch strings.ToUpper(s) {
	case "CAUSED":
		return graph.RelationCaused
	case "PREVENTED":
		return graph.RelationPrevented
	case "CORRELATED":
		return graph.RelationCorrelated
	default:
		return graph.RelationUnknown
	}
}

func parsePolarity(s string) graph.Polarity {
	if strings.ToUpper(s) == "POSITIVE" {
		return graph.PolarityPositive
	}
	return graph.PolarityNegative
}

func clamp(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
