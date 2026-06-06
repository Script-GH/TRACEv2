package interview_test

import (
	"context"
	"testing"

	"github.com/trace/trace/internal/application/agents/interview"
	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/domain/graph"
	"github.com/trace/trace/internal/domain/user"
	"github.com/trace/trace/internal/domain/workout"
)

// mockMCPClient satisfies ports.MCPClient for testing.
// Only override the methods relevant to the test.
type mockMCPClient struct {
	saveProfileFn func(ctx context.Context, p *user.Profile) error
}

func (m *mockMCPClient) GetCausalGraph(ctx context.Context, userID, domain string) (*graph.CausalGraph, error) {
	return nil, nil
}
func (m *mockMCPClient) UpdateCausalGraph(ctx context.Context, userID string, g *graph.CausalGraph) (int, error) {
	return 0, nil
}
func (m *mockMCPClient) QueryCausalPaths(ctx context.Context, userID string, q ports.CausalPathQuery) ([]graph.CausalPath, error) {
	return nil, nil
}
func (m *mockMCPClient) GetWorkoutLogHistory(ctx context.Context, userID string, n int) ([]workout.Log, error) {
	return nil, nil
}
func (m *mockMCPClient) GetProgressionHistory(ctx context.Context, userID, exercise string) ([]workout.ProgressionEntry, error) {
	return nil, nil
}
func (m *mockMCPClient) GetUserProfile(ctx context.Context, userID string) (*user.Profile, error) {
	return nil, nil
}
func (m *mockMCPClient) GetInjuryConstraints(ctx context.Context, userID string) ([]graph.InjuryConstraint, error) {
	return nil, nil
}
func (m *mockMCPClient) GetExerciseLibrary(ctx context.Context, f ports.ExerciseFilter) ([]workout.Exercise, error) {
	return nil, nil
}
func (m *mockMCPClient) StoreWorkoutPlan(ctx context.Context, plan *workout.Plan) (string, error) {
	return "", nil
}
func (m *mockMCPClient) LogVetoEvent(ctx context.Context, e ports.VetoEvent) (string, error) {
	return "", nil
}
func (m *mockMCPClient) FlagConflict(ctx context.Context, c ports.ConflictRecord) (string, error) {
	return "", nil
}
func (m *mockMCPClient) StoreCoTTrace(ctx context.Context, t ports.CoTTrace) (string, error) {
	return "", nil
}
func (m *mockMCPClient) GetCoTTrace(ctx context.Context, traceID string) (*ports.CoTTrace, error) {
	return nil, nil
}

// mockLLMClient satisfies ports.LLMClient for testing.
type mockLLMClient struct {
	chatFn func(ctx context.Context, req ports.ChatRequest) (ports.ChatResponse, error)
}

func (m *mockLLMClient) Chat(ctx context.Context, req ports.ChatRequest) (ports.ChatResponse, error) {
	if m.chatFn != nil {
		return m.chatFn(ctx, req)
	}
	return ports.ChatResponse{}, nil
}
func (m *mockLLMClient) ChatStream(ctx context.Context, req ports.ChatRequest) (<-chan ports.StreamToken, error) {
	return nil, nil
}
func (m *mockLLMClient) CompleteJSON(ctx context.Context, prompt string, dst any) error {
	return nil
}

func TestInterviewAgent_StartSession_CreatesSession(t *testing.T) {
	agent := interview.New(&mockMCPClient{}, &mockLLMClient{})
	_, err := agent.StartSession(context.Background(), "user-1")
	// When implemented: expect no error and a valid session ID
	_ = err
}

func TestInterviewAgent_IsComplete_FalseWithNoFields(t *testing.T) {
	agent := interview.New(&mockMCPClient{}, &mockLLMClient{})
	sessionID, _ := agent.StartSession(context.Background(), "user-1")
	complete, _ := agent.IsComplete(context.Background(), sessionID)
	// When implemented: a fresh session with no extracted fields must not be complete
	_ = complete
}

func TestInterviewAgent_ProcessTurn_ExtractsGoal(t *testing.T) {
	llm := &mockLLMClient{
		chatFn: func(ctx context.Context, req ports.ChatRequest) (ports.ChatResponse, error) {
			// Simulate Gemini returning a goal extraction
			return ports.ChatResponse{Content: `{"primary_goal": "lose weight", "confidence": 0.9}`}, nil
		},
	}
	agent := interview.New(&mockMCPClient{}, llm)
	sessionID, _ := agent.StartSession(context.Background(), "user-1")
	_, err := agent.ProcessTurn(context.Background(), sessionID, "I want to lose belly fat")
	_ = err
	// When implemented: extracted fields should include primary_goal at ≥ 0.8 confidence
}
