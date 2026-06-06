package orchestrator

import (
	"context"
	"fmt"
	"sync"
)

// Machine is the LangGraph-equivalent state machine.
// It holds session contexts and executes the transition table.
// Thread-safe: each session is independently locked.
type Machine struct {
	mu       sync.RWMutex
	sessions map[string]*SessionContext
	handlers map[State]StateHandler
}

// StateHandler is a function executed on entry to a state.
// It receives the session context and returns the event that should
// trigger the next transition (or empty string to wait for an external event).
type StateHandler func(ctx context.Context, session *SessionContext) (Event, error)

// NewMachine creates a state machine with the given handlers wired in.
// Called from main.go with all agents and orchestration logic injected.
func NewMachine(handlers map[State]StateHandler) *Machine {
	return &Machine{
		sessions: make(map[string]*SessionContext),
		handlers: handlers,
	}
}

// StartSession registers a new session and sets initial state.
func (m *Machine) StartSession(sessionID, userID string, initial State) error {
	panic("not implemented")
}

// Send fires an event against a session. If the event is valid for the current
// state, transitions occur and the entry handler for the new state is called.
// Returns the new state.
func (m *Machine) Send(ctx context.Context, sessionID string, event Event) (State, error) {
	panic("not implemented")
}

// CurrentState returns the current state for a session.
func (m *Machine) CurrentState(sessionID string) (State, error) {
	panic("not implemented")
}

// transition looks up the next state for (from, event) in AllTransitions.
func (m *Machine) transition(from State, event Event) (State, error) {
	for _, t := range AllTransitions {
		if t.From == from && t.On == event {
			return t.To, nil
		}
	}
	return "", fmt.Errorf("no transition from %s on %s", from, event)
}
