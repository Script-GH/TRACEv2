package domain

import "errors"

var (
	// Graph errors
	ErrGraphNotFound      = errors.New("causal graph not found for user")
	ErrGraphVersionConflict = errors.New("graph version conflict — stale write")
	ErrInvalidGraphStructure = errors.New("graph structure failed validation")

	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrProfileIncomplete  = errors.New("user profile fields below confidence threshold")
	ErrInterviewIncomplete = errors.New("interview not yet complete")

	// Workout errors
	ErrWorkoutLogNotFound = errors.New("workout log not found")
	ErrWorkoutPlanNotFound = errors.New("workout plan not found")
	ErrNoLogsAvailable    = errors.New("no workout logs available for this user")

	// Agent errors
	ErrAgentTimeout       = errors.New("agent timed out")
	ErrVetoFired          = errors.New("veto fired — exercise list modified")
	ErrUnresolvableConflict = errors.New("agent outputs are irreconcilable")
	ErrAgentUnhealthy     = errors.New("agent missed heartbeats — marked unhealthy")

	// MCP errors
	ErrMCPToolNotFound    = errors.New("mcp tool not found")
	ErrMCPToolTimeout     = errors.New("mcp tool call timed out")

	// Auth errors
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")
)
