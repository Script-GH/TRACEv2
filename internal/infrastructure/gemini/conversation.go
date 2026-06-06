package gemini

import (
	"sync"

	"github.com/trace/trace/internal/application/ports"
)

// ConversationSession implements ports.ConversationSession.
// Holds the turn history for one multi-turn conversation (interview or log session).
// Thread-safe — multiple goroutines may read history concurrently.
type ConversationSession struct {
	mu      sync.RWMutex
	history []ports.Message
}

func NewConversationSession() *ConversationSession {
	return &ConversationSession{
		history: make([]ports.Message, 0, 32),
	}
}

func (s *ConversationSession) AddTurn(role ports.Role, content string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history = append(s.history, ports.Message{Role: role, Content: content})
}

func (s *ConversationSession) History() []ports.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ports.Message, len(s.history))
	copy(out, s.history)
	return out
}

func (s *ConversationSession) LastTurn() *ports.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.history) == 0 {
		return nil
	}
	last := s.history[len(s.history)-1]
	return &last
}

func (s *ConversationSession) TurnCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.history)
}

func (s *ConversationSession) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.history = s.history[:0]
}
