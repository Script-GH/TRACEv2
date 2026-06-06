package redis

import "fmt"

// Key patterns for all Redis data.
// Prefix trace:{user_id}: ensures cross-user access is structurally impossible.
//
// All keys are defined here — no magic strings anywhere else in the codebase.

const (
	keyPrefix = "trace"
)

// Graph keys

// GraphKey is the primary causal graph document for a user.
func GraphKey(userID string) string {
	return fmt.Sprintf("%s:%s:graph:current", keyPrefix, userID)
}

// GraphVersionKey holds only the current version integer — cheap to read for conflict detection.
func GraphVersionKey(userID string) string {
	return fmt.Sprintf("%s:%s:graph:version", keyPrefix, userID)
}

// GraphSnapshotKey is an immutable snapshot of a specific graph version.
// Enables rollback and diff-based explanations.
func GraphSnapshotKey(userID string, version int) string {
	return fmt.Sprintf("%s:%s:graph:v%d", keyPrefix, userID, version)
}

// User profile keys

// ProfileKey is the AES-256 encrypted user profile.
func ProfileKey(userID string) string {
	return fmt.Sprintf("%s:%s:profile", keyPrefix, userID)
}

// Workout log keys

// WorkoutLogKey is a single workout log document.
func WorkoutLogKey(userID string, logID string) string {
	return fmt.Sprintf("%s:%s:log:%s", keyPrefix, userID, logID)
}

// WorkoutLogIndexKey is a Redis sorted set: score = sessionDate as unix timestamp.
// Used for efficient "last N logs" queries.
func WorkoutLogIndexKey(userID string) string {
	return fmt.Sprintf("%s:%s:log:index", keyPrefix, userID)
}

// Workout plan keys

// WorkoutPlanKey is a single generated plan.
func WorkoutPlanKey(userID string, planID string) string {
	return fmt.Sprintf("%s:%s:plan:%s", keyPrefix, userID, planID)
}

// WorkoutPlanByDateKey maps a date to a plan ID for quick lookup.
func WorkoutPlanByDateKey(userID string, date string) string {
	return fmt.Sprintf("%s:%s:plan:date:%s", keyPrefix, userID, date)
}

// Progression history keys

// ProgressionKey is a Redis list of ProgressionEntry records for one exercise.
func ProgressionKey(userID string, exerciseID string) string {
	return fmt.Sprintf("%s:%s:progression:%s", keyPrefix, userID, exerciseID)
}

// CoT trace keys

// CoTTraceKey stores a chain-of-thought trace with a 30-day TTL.
func CoTTraceKey(traceID string) string {
	return fmt.Sprintf("%s:cot:%s", keyPrefix, traceID)
}

// Veto event keys

// VetoIndexKey is a Redis list of veto event IDs for a user.
func VetoIndexKey(userID string) string {
	return fmt.Sprintf("%s:%s:veto:index", keyPrefix, userID)
}

// VetoEventKey is a single veto event record.
func VetoEventKey(vetoID string) string {
	return fmt.Sprintf("%s:veto:%s", keyPrefix, vetoID)
}

// Session state keys

// SessionStateKey stores the current state machine state for a session.
func SessionStateKey(sessionID string) string {
	return fmt.Sprintf("%s:session:%s:state", keyPrefix, sessionID)
}

// Rate limiting keys

// RateLimitKey is the counter for API rate limiting per user.
func RateLimitKey(userID string) string {
	return fmt.Sprintf("%s:%s:ratelimit", keyPrefix, userID)
}
