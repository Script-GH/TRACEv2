package groq

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/trace/trace/config"
	"github.com/trace/trace/internal/application/ports"
)

const baseURL = "https://api.groq.com/openai/v1"

type Client struct {
	cfg  config.GroqConfig
	http *http.Client
}

func NewClient(cfg config.GroqConfig) *Client {
	return &Client{cfg: cfg, http: &http.Client{}}
}

type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatReq struct {
	Model          string      `json:"model"`
	Messages       []chatMsg   `json:"messages"`
	MaxTokens      int         `json:"max_tokens,omitempty"`
	Temperature    float32     `json:"temperature,omitempty"`
	Stream         bool        `json:"stream,omitempty"`
	ResponseFormat *respFmt    `json:"response_format,omitempty"`
}

type respFmt struct {
	Type string `json:"type"`
}

type chatResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) Chat(ctx context.Context, req ports.ChatRequest) (ports.ChatResponse, error) {
	resp, err := c.do(ctx, chatReq{
		Model:       c.cfg.Model,
		Messages:    buildMessages(req),
		MaxTokens:   c.cfg.MaxOutputTokens,
		Temperature: c.cfg.Temperature,
	})
	if err != nil {
		return ports.ChatResponse{}, err
	}
	if len(resp.Choices) == 0 {
		return ports.ChatResponse{}, fmt.Errorf("groq: empty response")
	}
	return ports.ChatResponse{
		Content:      resp.Choices[0].Message.Content,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
	}, nil
}

func (c *Client) ChatStream(ctx context.Context, req ports.ChatRequest) (<-chan ports.StreamToken, error) {
	ch := make(chan ports.StreamToken, 32)
	go func() {
		defer close(ch)
		if err := c.doStream(ctx, chatReq{
			Model:       c.cfg.Model,
			Messages:    buildMessages(req),
			MaxTokens:   c.cfg.MaxOutputTokens,
			Temperature: c.cfg.Temperature,
			Stream:      true,
		}, ch); err != nil {
			ch <- ports.StreamToken{Error: err, Done: true}
		}
	}()
	return ch, nil
}

func (c *Client) CompleteJSON(ctx context.Context, prompt string, dst any) error {
	resp, err := c.do(ctx, chatReq{
		Model: c.cfg.Model,
		Messages: []chatMsg{
			{Role: "system", Content: "You must respond with valid JSON only. No markdown, no code fences, no explanation."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:      c.cfg.MaxOutputTokens,
		Temperature:    0.3,
		ResponseFormat: &respFmt{Type: "json_object"},
	})
	if err != nil {
		return err
	}
	if len(resp.Choices) == 0 {
		return fmt.Errorf("groq: empty response")
	}
	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	if strings.HasPrefix(content, "```") {
		lines := strings.Split(content, "\n")
		if len(lines) > 2 {
			content = strings.Join(lines[1:len(lines)-1], "\n")
		}
	}
	if err := json.Unmarshal([]byte(content), dst); err != nil {
		return fmt.Errorf("groq json parse: %w (raw: %.300s)", err, content)
	}
	return nil
}

func (c *Client) do(ctx context.Context, body chatReq) (*chatResp, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("groq marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("groq request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("groq http: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("groq read: %w", err)
	}
	var result chatResp
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("groq decode: %w (body: %.200s)", err, string(data))
	}
	if res.StatusCode != http.StatusOK {
		msg := string(data)
		if result.Error != nil {
			msg = result.Error.Message
		}
		return nil, fmt.Errorf("groq %d: %s", res.StatusCode, msg)
	}
	return &result, nil
}

func (c *Client) doStream(ctx context.Context, body chatReq, ch chan<- ports.StreamToken) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("groq marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("groq request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("groq http: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(res.Body)
		return fmt.Errorf("groq stream %d: %s", res.StatusCode, string(data))
	}

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		payload := strings.TrimPrefix(line, "data: ")
		if payload == "[DONE]" {
			ch <- ports.StreamToken{Done: true}
			return nil
		}
		var chunk chatResp
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 {
			if t := chunk.Choices[0].Delta.Content; t != "" {
				ch <- ports.StreamToken{Text: t}
			}
			if chunk.Choices[0].FinishReason == "stop" {
				ch <- ports.StreamToken{Done: true}
				return nil
			}
		}
	}
	return scanner.Err()
}

func buildMessages(req ports.ChatRequest) []chatMsg {
	var msgs []chatMsg
	if req.SystemPrompt != "" {
		msgs = append(msgs, chatMsg{Role: "system", Content: req.SystemPrompt})
	}
	for _, m := range req.History {
		role := "user"
		if m.Role == ports.RoleModel {
			role = "assistant"
		}
		msgs = append(msgs, chatMsg{Role: role, Content: m.Content})
	}
	msgs = append(msgs, chatMsg{Role: "user", Content: req.UserMessage})
	return msgs
}
