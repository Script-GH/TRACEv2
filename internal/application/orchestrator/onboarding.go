package orchestrator

import (
	"context"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
)

// OnboardingOrchestrator handles the NEW_USER → AWAITING_LOG path.
// It wires the InterviewAgent and TwinAgent into state machine handlers.
type OnboardingOrchestrator struct {
	interview agents.InterviewAgent
	twin      agents.TwinAgent
	moe       agents.MoEGateAgent
	mcp       ports.MCPClient
	daily     *DailyLoopOrchestrator // reused for first workout generation
}

func NewOnboardingOrchestrator(
	interview agents.InterviewAgent,
	twin agents.TwinAgent,
	moe agents.MoEGateAgent,
	mcp ports.MCPClient,
	daily *DailyLoopOrchestrator,
) *OnboardingOrchestrator {
	return &OnboardingOrchestrator{
		interview: interview,
		twin:      twin,
		moe:       moe,
		mcp:       mcp,
		daily:     daily,
	}
}

// HandleInterviewActive processes one turn of the onboarding conversation.
// Returns EventAllFieldsExtracted when all profile fields reach ≥ 0.8 confidence.
// Returns EventInterviewTurn for every intermediate turn.
func (o *OnboardingOrchestrator) HandleInterviewActive(ctx context.Context, session *SessionContext) (Event, error) {
	panic("not implemented")
}

// HandleInterviewComplete shows the INITIAL_SUMMARY to the user.
// Returns EventUserConfirmedSummary when the user confirms (or corrects and re-confirms).
func (o *OnboardingOrchestrator) HandleInterviewComplete(ctx context.Context, session *SessionContext) (Event, error) {
	panic("not implemented")
}

// HandleGraphBuildInitial calls TwinAgent.BuildInitial and stores the result.
// Returns EventGraphBuilt on success.
func (o *OnboardingOrchestrator) HandleGraphBuildInitial(ctx context.Context, session *SessionContext) (Event, error) {
	panic("not implemented")
}

// HandleFirstWorkoutGen triggers the daily loop for the first workout.
func (o *OnboardingOrchestrator) HandleFirstWorkoutGen(ctx context.Context, session *SessionContext) (Event, error) {
	panic("not implemented")
}

// Handlers returns the StateHandler map for the onboarding path.
// Passed to Machine.NewMachine during wiring in main.go.
func (o *OnboardingOrchestrator) Handlers() map[State]StateHandler {
	return map[State]StateHandler{
		StateInterviewActive:   o.HandleInterviewActive,
		StateInterviewComplete: o.HandleInterviewComplete,
		StateGraphBuildInitial: o.HandleGraphBuildInitial,
		StateFirstWorkoutGen:   o.HandleFirstWorkoutGen,
	}
}
