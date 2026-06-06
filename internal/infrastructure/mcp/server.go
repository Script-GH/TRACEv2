package mcp

import (
	"github.com/trace/trace/internal/application/ports"
)

// Server is the internal MCP tool server.
// It owns all repository references and exposes them only through
// the MCPClient interface — never as raw repositories to agent code.
//
// This is the enforcement point for the spec rule:
//   "Direct database access from agent code is a build failure."
type Server struct {
	graphs      ports.CausalGraphRepository
	users       ports.UserProfileRepository
	logs        ports.WorkoutLogRepository
	plans       ports.WorkoutPlanRepository
	exercises   ports.ExerciseLibraryRepository
	progression ports.ProgressionHistoryRepository
	vetos       ports.VetoEventRepository
	cot         ports.CoTTraceRepository
	conflicts   ports.ConflictRepository
}

func NewServer(
	graphs ports.CausalGraphRepository,
	users ports.UserProfileRepository,
	logs ports.WorkoutLogRepository,
	plans ports.WorkoutPlanRepository,
	exercises ports.ExerciseLibraryRepository,
	progression ports.ProgressionHistoryRepository,
	vetos ports.VetoEventRepository,
	cot ports.CoTTraceRepository,
	conflicts ports.ConflictRepository,
) *Server {
	return &Server{
		graphs:      graphs,
		users:       users,
		logs:        logs,
		plans:       plans,
		exercises:   exercises,
		progression: progression,
		vetos:       vetos,
		cot:         cot,
		conflicts:   conflicts,
	}
}
