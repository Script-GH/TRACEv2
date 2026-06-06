package coach

import (
	"context"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/domain/workout"
)

// Agent implements agents.CoachAgent.
// It uses Gemini to generate the plain-language CausalReason for each exercise,
// grounded in the actual causal graph paths used to select that exercise.
//
// If Gemini times out or fails:
//   - Returns the plan with empty CausalReason fields
//   - Returns a non-fatal CoachUnavailableError
//   - Caller renders "Reasoning unavailable — tap to retry" in the UI
type Agent struct {
	mcp ports.MCPClient
	llm ports.LLMClient
}

func New(mcp ports.MCPClient, llm ports.LLMClient) *Agent {
	return &Agent{mcp: mcp, llm: llm}
}

// CoachUnavailableError is returned when Gemini is unreachable.
// It is non-fatal — the plan is still delivered without reasoning text.
type CoachUnavailableError struct {
	Cause error
}

func (e *CoachUnavailableError) Error() string {
	return "coach agent unavailable: " + e.Cause.Error()
}

// Annotate iterates over each PlannedExercise in the plan, queries the causal
// graph for the path that justified the exercise, and sends that path to Gemini
// for plain-English narration. Stores the CoT trace via MCPClient.StoreCoTTrace.
func (a *Agent) Annotate(ctx context.Context, userID string, plan *workout.Plan) (*workout.Plan, error) {
	panic("not implemented")
}

// annotateOne generates the CausalReason for a single exercise.
// Separated for per-exercise retry logic.
func (a *Agent) annotateOne(ctx context.Context, userID string, ex *workout.PlannedExercise, graphVersion int, cot *agents.ExpertInput) error {
	panic("not implemented")
}
