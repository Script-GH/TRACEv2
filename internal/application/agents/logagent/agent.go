package logagent

import (
	"context"
	"sync"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/domain/workout"
)

// Agent implements agents.LogAgent.
// Mirrors InterviewAgent's session pattern but for post-workout logging.
type Agent struct {
	mcp      ports.MCPClient
	llm      ports.LLMClient
	mu       sync.RWMutex
	sessions map[string]*session
}

type session struct {
	userID      string
	sessionDate string
	conv        ports.ConversationSession
	log         workout.Log
}

func New(mcp ports.MCPClient, llm ports.LLMClient) *Agent {
	return &Agent{
		mcp:      mcp,
		llm:      llm,
		sessions: make(map[string]*session),
	}
}

func (a *Agent) StartSession(ctx context.Context, userID string, sessionDate string) (string, error) {
	panic("not implemented")
}

func (a *Agent) ProcessTurn(ctx context.Context, sessionID string, userMessage string) (agents.LogTurnOutput, error) {
	panic("not implemented")
}

func (a *Agent) GetExtractedLog(ctx context.Context, sessionID string) (*workout.Log, error) {
	panic("not implemented")
}

// Finalise saves the confirmed WorkoutLog, then triggers progressive history
// extraction to update the ProgressionEntry records for each exercise.
func (a *Agent) Finalise(ctx context.Context, sessionID string) (string, error) {
	panic("not implemented")
}
