package injury_test

import (
	"context"
	"testing"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/agents/injury"
	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/domain/graph"
	"github.com/trace/trace/internal/domain/user"
	"github.com/trace/trace/internal/domain/workout"
)

type mockMCP struct {
	getGraphFn       func(ctx context.Context, userID, domain string) (*graph.CausalGraph, error)
	getConstraintsFn func(ctx context.Context, userID string) ([]graph.InjuryConstraint, error)
	getLibraryFn     func(ctx context.Context, f ports.ExerciseFilter) ([]workout.Exercise, error)
	logVetoFn        func(ctx context.Context, e ports.VetoEvent) (string, error)
}

func (m *mockMCP) GetCausalGraph(ctx context.Context, userID, domain string) (*graph.CausalGraph, error) {
	if m.getGraphFn != nil {
		return m.getGraphFn(ctx, userID, domain)
	}
	return &graph.CausalGraph{}, nil
}
func (m *mockMCP) UpdateCausalGraph(ctx context.Context, userID string, g *graph.CausalGraph) (int, error) {
	return 0, nil
}
func (m *mockMCP) QueryCausalPaths(ctx context.Context, userID string, q ports.CausalPathQuery) ([]graph.CausalPath, error) {
	return nil, nil
}
func (m *mockMCP) GetWorkoutLogHistory(ctx context.Context, userID string, n int) ([]workout.Log, error) {
	return nil, nil
}
func (m *mockMCP) GetProgressionHistory(ctx context.Context, userID, exercise string) ([]workout.ProgressionEntry, error) {
	return nil, nil
}
func (m *mockMCP) GetUserProfile(ctx context.Context, userID string) (*user.Profile, error) {
	return nil, nil
}
func (m *mockMCP) GetInjuryConstraints(ctx context.Context, userID string) ([]graph.InjuryConstraint, error) {
	if m.getConstraintsFn != nil {
		return m.getConstraintsFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockMCP) GetExerciseLibrary(ctx context.Context, f ports.ExerciseFilter) ([]workout.Exercise, error) {
	if m.getLibraryFn != nil {
		return m.getLibraryFn(ctx, f)
	}
	return nil, nil
}
func (m *mockMCP) StoreWorkoutPlan(ctx context.Context, plan *workout.Plan) (string, error) { return "", nil }
func (m *mockMCP) LogVetoEvent(ctx context.Context, e ports.VetoEvent) (string, error) {
	if m.logVetoFn != nil {
		return m.logVetoFn(ctx, e)
	}
	return "veto-1", nil
}
func (m *mockMCP) FlagConflict(ctx context.Context, c ports.ConflictRecord) (string, error) { return "", nil }
func (m *mockMCP) StoreCoTTrace(ctx context.Context, t ports.CoTTrace) (string, error) { return "", nil }
func (m *mockMCP) GetCoTTrace(ctx context.Context, traceID string) (*ports.CoTTrace, error) { return nil, nil }

// highConfidenceGraphForKnee returns a graph with a 0.93-confidence negative edge
// from "high_volume_squats" → "left_knee_pain".
func highConfidenceGraphForKnee() *graph.CausalGraph {
	return &graph.CausalGraph{
		UserID:  "user-1",
		Version: 3,
		Nodes: map[string]*graph.Node{
			"n1": {ID: "n1", Label: "high_volume_squats", Type: graph.NodeTypeExercise},
			"n2": {ID: "n2", Label: "left_knee_pain", Type: graph.NodeTypeOutcome},
		},
		Edges: map[string]*graph.Edge{
			"e1": {
				ID:       "e1",
				FromNode: "n1",
				ToNode:   "n2",
				Relation: graph.RelationCaused,
				Confidence: 0.93,
				Polarity: graph.PolarityNegative,
			},
		},
	}
}

func TestInjuryAgent_VetoHighConfidenceExercise(t *testing.T) {
	mcp := &mockMCP{
		getGraphFn: func(_ context.Context, _, _ string) (*graph.CausalGraph, error) {
			return highConfidenceGraphForKnee(), nil
		},
		getConstraintsFn: func(_ context.Context, _ string) ([]graph.InjuryConstraint, error) {
			return []graph.InjuryConstraint{
				{BodyPart: "left_knee", Status: "chronic", ContraExercises: []string{"high_volume_squats"}},
			}, nil
		},
	}

	a := injury.New(mcp)
	report, err := a.Evaluate(context.Background(),
		agents.ExpertInput{UserID: "user-1", GraphVersion: 3},
		[]string{"high_volume_squats", "romanian_deadlift"},
	)
	if err != nil {
		t.Fatal(err)
	}
	// When implemented: squats must be vetoed, RDL must be approved
	_ = report
}

func TestInjuryAgent_ApprovesExerciseWithNoCausalPath(t *testing.T) {
	mcp := &mockMCP{
		getGraphFn: func(_ context.Context, _, _ string) (*graph.CausalGraph, error) {
			return &graph.CausalGraph{Nodes: map[string]*graph.Node{}, Edges: map[string]*graph.Edge{}}, nil
		},
	}
	a := injury.New(mcp)
	report, err := a.Evaluate(context.Background(),
		agents.ExpertInput{UserID: "user-1"},
		[]string{"romanian_deadlift"},
	)
	if err != nil {
		t.Fatal(err)
	}
	// When implemented: no negative graph paths → all exercises approved
	_ = report
}
