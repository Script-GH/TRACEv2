package messaging

import (
	"context"
	"sync"

	"github.com/trace/trace/internal/application/ports"
)

// Broker implements ports.MessageBroker using in-process buffered channels.
// One broker instance per server process — sessions are isolated by sessionID prefix.
// For multi-instance scaling, swap this implementation with a Redis pub/sub broker.
type Broker struct {
	mu          sync.RWMutex
	subscribers map[string]map[string]chan ports.A2AMessage // sessionID → agentName → channel
	bufSize     int
}

func NewBroker(bufSize int) *Broker {
	return &Broker{
		subscribers: make(map[string]map[string]chan ports.A2AMessage),
		bufSize:     bufSize,
	}
}

// Publish routes a message to the channel for msg.ToAgent within msg.SessionID.
// If ToAgent == "broadcast", delivers to all subscribers in the session.
// Non-blocking: drops the message and returns an error if the channel is full.
func (b *Broker) Publish(ctx context.Context, msg ports.A2AMessage) error {
	panic("not implemented")
}

// Subscribe creates a channel for agentName within sessionID.
// The channel is buffered at b.bufSize. Closing ctx closes the channel.
func (b *Broker) Subscribe(ctx context.Context, sessionID string, agentName string) (<-chan ports.A2AMessage, error) {
	panic("not implemented")
}

// Drain collects all pending messages for a session (all subscribers).
// Used by PlannerAgent after all experts have reported.
func (b *Broker) Drain(ctx context.Context, sessionID string) ([]ports.A2AMessage, error) {
	panic("not implemented")
}

// Close tears down all subscriptions and channels for a session.
func (b *Broker) Close(sessionID string) error {
	panic("not implemented")
}
