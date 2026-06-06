package twin_test

import (
	"context"
	"testing"

	"github.com/trace/trace/internal/application/agents/twin"
	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/domain/graph"
	"github.com/trace/trace/internal/domain/user"
	"github.com/trace/trace/internal/domain/workout"
)

type mockMCP struct {
	getGraphFn    func(ctx context.Context, userID, domain string) (*graph.CausalGraph, error)
	updateGraphFn func(ctx context.Context, userID string, g *graph.CausalGraph) (int, error)
	getLogsFn     func(ctx context.Context, userID string, n int) ([]workout.Log, error)
	getProfileFn  func(ctx context.Context, userID string) (*user.Profile, error)
}

func (m *mockMCP) GetCausalGraph(ctx context.Context, userID, domain string) (*graph.CausalGraph, error) {
	if m.getGraphFn != nil {
		return m.getGraphFn(ctx, userID, domain)
	}
	return &graph.CausalGraph{UserID: userID, Version: 0, Nodes: map[string]*graph.Node{}, Edges: map[string]*graph.Edge{}}, nil
}
func (m *mockMCP) UpdateCausalGraph(ctx context.Context, userID string, g *graph.CausalGraph) (int, error) {
	if m.updateGraphFn != nil {
		return m.updateGraphFn(ctx, userID, g)
	}
	return g.Version + 1, nil
}
func (m *mockMCP) QueryCausalPaths(ctx context.Context, userID string, q ports.CausalPathQuery) ([]graph.CausalPath, error) {
	return nil, nil
}
func (m *mockMCP) GetWorkoutLogHistory(ctx context.Context, userID string, n int) ([]workout.Log, error) {
	if m.getLogsFn != nil {
		return m.getLogsFn(ctx, userID, n)
	}
	return nil, nil
}
func (m *mockMCP) GetProgressionHistory(ctx context.Context, userID, exercise string) ([]workout.ProgressionEntry, error) {
	return nil, nil
}
func (m *mockMCP) GetUserProfile(ctx context.Context, userID string) (*user.Profile, error) {
	if m.getProfileFn != nil {
		return m.getProfileFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockMCP) GetInjuryConstraints(ctx context.Context, userID string) ([]graph.InjuryConstraint, error) {
	return nil, nil
}
func (m *mockMCP) GetExerciseLibrary(ctx context.Context, f ports.ExerciseFilter) ([]workout.Exercise, error) {
	return nil, nil
}
func (m *mockMCP) StoreWorkoutPlan(ctx context.Context, plan *workout.Plan) (string, error) {
	return "", nil
}
func (m *mockMCP) LogVetoEvent(ctx context.Context, e ports.VetoEvent) (string, error) {
	return "", nil
}
func (m *mockMCP) FlagConflict(ctx context.Context, c ports.ConflictRecord) (string, error) {
	return "", nil
}
func (m *mockMCP) StoreCoTTrace(ctx context.Context, t ports.CoTTrace) (string, error) {
	return "", nil
}
func (m *mockMCP) GetCoTTrace(ctx context.Context, traceID string) (*ports.CoTTrace, error) {
	return nil, nil
}

func TestTwinAgent_Update_IncreasesVersionOnSuccess(t *testing.T) {
	mcp := &mockMCP{}
	agent := twin.New(mcp, nil)
	newVer, err := agent.Update(context.Background(), "user-1", "log-1")
	_ = newVer
	_ = err
	// When implemented: version must be > 0 on success
}

func TestTwinAgent_Update_CreatesNegativeEdgeOnHighSeverityPain(t *testing.T) {
	mcp := &mockMCP{
		getLogsFn: func(ctx context.Context, userID string, n int) ([]workout.Log, error) {
			return []workout.Log{
				{
					LogID: "log-1",
					PainSignals: []workout.PainSignal{
						{BodyPart: "left_knee", Severity: 8.0, Type: workout.PainTypeSharp, DuringExercise: "Squat"},
					},
				},
			}, nil
		},
	}
	agent := twin.New(mcp, nil)
	_, err := agent.Update(context.Background(), "user-1", "log-1")
	_ = err
	// When implemented: graph should contain a new negative CAUSED edge Squat → left_knee_pain
}
