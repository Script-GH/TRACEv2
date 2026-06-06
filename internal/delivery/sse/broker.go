package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/trace/trace/internal/domain/workout"
)

// Event types pushed to clients.
const (
	EventWorkoutReady     = "workout_ready"
	EventGraphUpdated     = "graph_updated"
	EventGenerationStatus = "generation_status" // progress indicator during workout gen
	EventCoachToken       = "coach_token"        // streaming CoT tokens from CoachAgent
)

// SSEEvent is a single server-sent event payload.
type SSEEvent struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// WorkoutReadyPayload is the SSE payload for EventWorkoutReady.
type WorkoutReadyPayload struct {
	Plan        *workout.Plan `json:"plan"`
	GraphVersion int          `json:"graph_version"`
}

// GenerationStatusPayload communicates progress during workout generation.
type GenerationStatusPayload struct {
	State   string `json:"state"`   // e.g. "GRAPH_UPDATE", "MOE_GATING"
	Message string `json:"message"` // human-readable
}

// Broker manages SSE connections per session.
// Max 2 connections per session ID — third connection closes the oldest (spec requirement).
type Broker struct {
	mu      sync.RWMutex
	clients map[string][]*client // sessionID → connected clients (max 2)
}

type client struct {
	sessionID string
	ch        chan SSEEvent
	cancel    context.CancelFunc
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[string][]*client),
	}
}

// Register adds a new SSE client for a session.
// If a third connection arrives, closes the oldest.
func (b *Broker) Register(sessionID string) *client {
	panic("not implemented")
}

// Publish sends an event to all connected clients for a session.
func (b *Broker) Publish(sessionID string, event SSEEvent) {
	panic("not implemented")
}

// PublishWorkoutReady is a typed helper for the workout_ready event.
func (b *Broker) PublishWorkoutReady(sessionID string, plan *workout.Plan, graphVersion int) {
	b.Publish(sessionID, SSEEvent{
		Type:    EventWorkoutReady,
		Payload: WorkoutReadyPayload{Plan: plan, GraphVersion: graphVersion},
	})
}

// PublishStatus sends a generation_status progress event.
func (b *Broker) PublishStatus(sessionID string, state string, message string) {
	b.Publish(sessionID, SSEEvent{
		Type:    EventGenerationStatus,
		Payload: GenerationStatusPayload{State: state, Message: message},
	})
}

// ServeHTTP handles an individual SSE connection.
// Must be called from an HTTP handler that has verified the JWT.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request, sessionID string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

	c := b.Register(sessionID)
	defer b.deregister(sessionID, c)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case event, open := <-c.ch:
			if !open {
				return
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (b *Broker) deregister(sessionID string, c *client) {
	panic("not implemented")
}
