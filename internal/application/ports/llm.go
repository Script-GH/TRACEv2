package ports

import "context"

// LLMClient is the abstraction over Gemini.
// Every agent that needs language model calls injects this interface.
// The infrastructure/gemini package implements it.
// Tests can inject a deterministic mock without any API calls.
type LLMClient interface {
	// Chat sends a conversational turn and returns the full response text.
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)

	// ChatStream returns a channel that receives tokens as they stream.
	// The channel closes when the response is complete or ctx is cancelled.
	ChatStream(ctx context.Context, req ChatRequest) (<-chan StreamToken, error)

	// CompleteJSON sends a structured completion request and unmarshals the
	// response into dst. dst must be a pointer to a struct.
	// Used by MoEGateAgent (weight vector), TwinAgent (graph inference).
	CompleteJSON(ctx context.Context, prompt string, dst any) error
}

// Message is a single turn in a conversation history.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Role identifies the speaker in a conversation.
type Role string

const (
	RoleUser  Role = "user"
	RoleModel Role = "model"
)

// ChatRequest encapsulates everything needed for a single LLM call.
type ChatRequest struct {
	SystemPrompt string    `json:"system_prompt"`
	History      []Message `json:"history"`
	UserMessage  string    `json:"user_message"`
}

// ChatResponse holds the model's reply and token usage.
type ChatResponse struct {
	Content    string `json:"content"`
	InputTokens  int  `json:"input_tokens"`
	OutputTokens int  `json:"output_tokens"`
}

// StreamToken is a single streamed token from the model.
type StreamToken struct {
	Text  string
	Done  bool
	Error error
}

// ConversationSession manages turn history for a multi-turn conversation.
// InterviewAgent and LogAgent use this to maintain context across turns.
type ConversationSession interface {
	AddTurn(role Role, content string)
	History() []Message
	LastTurn() *Message
	TurnCount() int
	Reset()
}
