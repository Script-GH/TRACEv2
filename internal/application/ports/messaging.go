package ports

import "context"

// Intent classifies the purpose of an A2A message.
type Intent string

const (
	IntentReport    Intent = "REPORT"    // expert agent → PlannerAgent
	IntentVeto      Intent = "VETO"      // InjuryAgent/RecoveryAgent → PlannerAgent
	IntentRequest   Intent = "REQUEST"   // any agent → MCP server
	IntentACK       Intent = "ACK"       // acknowledgement
	IntentOverride  Intent = "OVERRIDE"  // PlannerAgent → agent (re-evaluate)
	IntentHeartbeat Intent = "HEARTBEAT" // all agents → orchestrator
)

// A2AMessage is the typed envelope for all inter-agent communication.
// Signature is HMAC-SHA256 of the payload — PlannerAgent verifies before synthesis.
type A2AMessage struct {
	MsgID        string  `json:"msg_id"`
	SessionID    string  `json:"session_id"`
	FromAgent    string  `json:"from_agent"`
	ToAgent      string  `json:"to_agent"` // agent name, "planner", or "broadcast"
	Intent       Intent  `json:"intent"`
	Payload      any     `json:"payload"`      // typed per agent — see agent packages
	Confidence   float32 `json:"confidence"`
	CoTTraceID   string  `json:"cot_trace_id"` // nullable
	TimestampNS  int64   `json:"timestamp_ns"`
	GraphVersion int     `json:"graph_version"`
	Signature    string  `json:"signature"` // HMAC-SHA256 of payload bytes
}

// MessageBroker routes A2A messages between agents within a session.
// The infrastructure/messaging package implements this with buffered channels.
type MessageBroker interface {
	// Publish sends a message. Non-blocking; returns error if broker is full.
	Publish(ctx context.Context, msg A2AMessage) error

	// Subscribe returns a channel that receives messages addressed to agentName.
	// The channel is closed when ctx is cancelled.
	Subscribe(ctx context.Context, sessionID string, agentName string) (<-chan A2AMessage, error)

	// Drain collects all messages for a session into a slice.
	// Used by PlannerAgent after all experts have reported.
	Drain(ctx context.Context, sessionID string) ([]A2AMessage, error)

	// Close tears down all subscriptions for a session.
	Close(sessionID string) error
}
