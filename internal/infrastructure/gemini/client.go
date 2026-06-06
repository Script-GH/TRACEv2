package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	genai "github.com/google/generative-ai-go/genai"
	"github.com/trace/trace/config"
	"github.com/trace/trace/internal/application/ports"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client implements ports.LLMClient using the Google Generative AI SDK.
// Stores the raw *genai.Client so each call gets a fresh model instance —
// this makes setting SystemInstruction per-call safe under concurrent use.
type Client struct {
	gc  *genai.Client
	cfg config.GeminiConfig
}

func NewClient(ctx context.Context, cfg config.GeminiConfig) (*Client, error) {
	c, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}
	return &Client{gc: c, cfg: cfg}, nil
}

func (c *Client) newModel(systemPrompt string) *genai.GenerativeModel {
	model := c.gc.GenerativeModel(c.cfg.Model)
	model.SetMaxOutputTokens(c.cfg.MaxOutputTokens)
	model.SetTemperature(c.cfg.Temperature)
	if systemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(systemPrompt)},
		}
	}
	return model
}

func (c *Client) Chat(ctx context.Context, req ports.ChatRequest) (ports.ChatResponse, error) {
	model := c.newModel(req.SystemPrompt)
	cs := model.StartChat()
	cs.History = buildHistory(req.History)

	var resp *genai.GenerateContentResponse
	var err error
	for attempt := 0; attempt < 4; attempt++ {
		resp, err = cs.SendMessage(ctx, genai.Text(req.UserMessage))
		if err == nil {
			break
		}
		if !isRateLimit(err) || attempt == 3 {
			return ports.ChatResponse{}, fmt.Errorf("gemini chat: %w", err)
		}
		wait := time.Duration(1<<uint(attempt+1)) * time.Second // 2s, 4s, 8s
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ports.ChatResponse{}, ctx.Err()
		}
		// reset the chat session for retry
		cs = model.StartChat()
		cs.History = buildHistory(req.History)
	}

	var sb strings.Builder
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if t, ok := part.(genai.Text); ok {
					sb.WriteString(string(t))
				}
			}
		}
	}

	var inputTokens, outputTokens int
	if resp.UsageMetadata != nil {
		inputTokens = int(resp.UsageMetadata.PromptTokenCount)
		outputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	}

	return ports.ChatResponse{
		Content:      sb.String(),
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	}, nil
}

func (c *Client) ChatStream(ctx context.Context, req ports.ChatRequest) (<-chan ports.StreamToken, error) {
	ch := make(chan ports.StreamToken, 32)
	model := c.newModel(req.SystemPrompt)
	cs := model.StartChat()
	cs.History = buildHistory(req.History)

	go func() {
		defer close(ch)
		iter := cs.SendMessageStream(ctx, genai.Text(req.UserMessage))
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				ch <- ports.StreamToken{Done: true}
				return
			}
			if err != nil {
				ch <- ports.StreamToken{Error: err, Done: true}
				return
			}
			for _, cand := range resp.Candidates {
				if cand.Content != nil {
					for _, part := range cand.Content.Parts {
						if t, ok := part.(genai.Text); ok && string(t) != "" {
							ch <- ports.StreamToken{Text: string(t)}
						}
					}
				}
			}
		}
	}()

	return ch, nil
}

// CompleteJSON sends a structured prompt and unmarshals the JSON response into dst.
// Strips markdown code fences if the model wraps its output despite instructions.
func (c *Client) CompleteJSON(ctx context.Context, prompt string, dst any) error {
	resp, err := c.Chat(ctx, ports.ChatRequest{
		SystemPrompt: "You must respond with valid JSON only. No markdown, no code fences, no explanation.",
		UserMessage:  prompt,
	})
	if err != nil {
		return err
	}
	content := strings.TrimSpace(resp.Content)
	if strings.HasPrefix(content, "```") {
		lines := strings.Split(content, "\n")
		if len(lines) > 2 {
			content = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	if err := json.Unmarshal([]byte(content), dst); err != nil {
		return fmt.Errorf("gemini json parse: %w (raw: %.300s)", err, content)
	}
	return nil
}

func isRateLimit(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "429") || strings.Contains(s, "quota") || strings.Contains(s, "RESOURCE_EXHAUSTED")
}

func buildHistory(history []ports.Message) []*genai.Content {
	contents := make([]*genai.Content, 0, len(history))
	for _, msg := range history {
		role := "user"
		if msg.Role == ports.RoleModel {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []genai.Part{genai.Text(msg.Content)},
		})
	}
	return contents
}
