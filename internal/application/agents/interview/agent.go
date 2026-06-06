package interview

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/domain/user"
)

// Agent implements agents.InterviewAgent.
// Dependencies: MCPClient (to save profile), LLMClient (for conversation).
// All session state is held in memory — sessions are short-lived (8–15 min).
type Agent struct {
	mcp      ports.MCPClient
	llm      ports.LLMClient
	mu       sync.RWMutex
	sessions map[string]*session
}

// session holds in-flight interview state for one user.
type session struct {
	userID    string
	history   []ports.Message
	extracted extractionResult
	mu        sync.Mutex
}

// extractionResult mirrors the JSON schema Gemini fills in each turn.
type extractionResult struct {
	PrimaryGoal           fieldVal[string]          `json:"primary_goal"`
	GoalTimelineWeeks     fieldVal[int]             `json:"goal_timeline_weeks"`
	TrainingAgeYears      fieldVal[float32]         `json:"training_age_years"`
	ProgramsTried         fieldVal[[]string]        `json:"programs_tried"`
	WhatWorked            fieldVal[string]          `json:"what_worked"`
	WhatFailed            fieldVal[string]          `json:"what_failed"`
	TypicalSessionMinutes fieldVal[int]             `json:"typical_session_minutes"`
	AvailableDaysPerWeek  fieldVal[int]             `json:"available_days_per_week"`
	EquipmentAccess       fieldVal[[]string]        `json:"equipment_access"`
	Injuries              fieldVal[[]injuryExt]     `json:"injuries"`
	SleepAverageHours     fieldVal[float32]         `json:"sleep_average_hours"`
	StressLevel           fieldVal[float32]         `json:"stress_level"`
	WeightKg              fieldVal[float32]         `json:"weight_kg"`
	BodyFatNow            fieldVal[float32]         `json:"body_fat_now"`
}

type fieldVal[T any] struct {
	Value      T       `json:"value"`
	Confidence float32 `json:"confidence"`
}

type injuryExt struct {
	BodyPart string   `json:"body_part"`
	Type     string   `json:"type"`
	Status   string   `json:"status"`
	Triggers []string `json:"triggers"`
}

func New(mcp ports.MCPClient, llm ports.LLMClient) *Agent {
	return &Agent{
		mcp:      mcp,
		llm:      llm,
		sessions: make(map[string]*session),
	}
}

// StartSession creates a new interview session and returns its ID.
func (a *Agent) StartSession(ctx context.Context, userID string) (string, error) {
	sessionID := uuid.New().String()
	sess := &session{userID: userID}
	a.mu.Lock()
	a.sessions[sessionID] = sess
	a.mu.Unlock()
	return sessionID, nil
}

// combinedResponse is what the single combined LLM call returns.
// Merging reply + extraction into one call halves API usage and avoids rate limits.
type combinedResponse struct {
	Reply      string          `json:"reply"`
	Extraction extractionResult `json:"extraction"`
}

// ProcessTurn handles one conversation turn using a single LLM call that
// simultaneously generates the conversational reply and extracts structured fields.
func (a *Agent) ProcessTurn(ctx context.Context, sessionID string, userMessage string) (agents.InterviewTurnOutput, error) {
	a.mu.RLock()
	sess := a.sessions[sessionID]
	a.mu.RUnlock()
	if sess == nil {
		return agents.InterviewTurnOutput{}, fmt.Errorf("session %s not found", sessionID)
	}

	sess.mu.Lock()
	defer sess.mu.Unlock()

	sess.history = append(sess.history, ports.Message{Role: ports.RoleUser, Content: userMessage})
	historyBefore := make([]ports.Message, len(sess.history)-1)
	copy(historyBefore, sess.history[:len(sess.history)-1])

	var combined combinedResponse
	prompt := buildCombinedPrompt(userMessage, historyBefore)
	if err := a.llm.CompleteJSON(ctx, prompt, &combined); err != nil {
		return agents.InterviewTurnOutput{}, fmt.Errorf("interview reply: %w", err)
	}

	reply := combined.Reply
	if reply == "" {
		reply = "Could you tell me a bit more about that?"
	}

	// Honour the [COMPLETE] signal the agent embeds when it has enough data.
	agentSignalsComplete := strings.Contains(reply, "[COMPLETE]")
	reply = strings.ReplaceAll(reply, "[COMPLETE]", "")
	reply = strings.TrimSpace(reply)

	sess.history = append(sess.history, ports.Message{Role: ports.RoleModel, Content: reply})
	sess.extracted = combined.Extraction

	return agents.InterviewTurnOutput{
		AgentReply:      reply,
		ExtractedFields: sess.fieldSummary(),
		IsComplete:      sess.isComplete() || agentSignalsComplete,
	}, nil
}

// GetExtractedProfile returns the current best-effort profile derived from the conversation.
func (a *Agent) GetExtractedProfile(ctx context.Context, sessionID string) (*user.Profile, error) {
	a.mu.RLock()
	sess := a.sessions[sessionID]
	a.mu.RUnlock()
	if sess == nil {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}
	sess.mu.Lock()
	defer sess.mu.Unlock()
	p := sess.toProfile()
	return &p, nil
}

// IsComplete returns true when all required fields have confidence >= 0.75.
func (a *Agent) IsComplete(ctx context.Context, sessionID string) (bool, error) {
	a.mu.RLock()
	sess := a.sessions[sessionID]
	a.mu.RUnlock()
	if sess == nil {
		return false, fmt.Errorf("session %s not found", sessionID)
	}
	sess.mu.Lock()
	defer sess.mu.Unlock()
	return sess.isComplete(), nil
}

// Finalise saves the profile and marks InterviewDone = true.
// Removes the session from memory after saving.
func (a *Agent) Finalise(ctx context.Context, sessionID string) error {
	a.mu.RLock()
	sess := a.sessions[sessionID]
	a.mu.RUnlock()
	if sess == nil {
		return fmt.Errorf("session %s not found", sessionID)
	}

	sess.mu.Lock()
	profile := sess.toProfile()
	profile.InterviewDone = true
	profile.CreatedAt = time.Now().Unix()
	sess.mu.Unlock()

	if err := a.mcp.SaveUserProfile(ctx, &profile); err != nil {
		return fmt.Errorf("finalise save profile: %w", err)
	}

	a.mu.Lock()
	delete(a.sessions, sessionID)
	a.mu.Unlock()
	return nil
}

// ── session helpers ────────────────────────────────────────────────────────

func (s *session) isComplete() bool {
	return s.extracted.PrimaryGoal.Confidence >= 0.65 &&
		s.extracted.AvailableDaysPerWeek.Confidence >= 0.60 &&
		s.extracted.TrainingAgeYears.Confidence >= 0.55 &&
		s.extracted.TypicalSessionMinutes.Confidence >= 0.55
}

func (s *session) fieldSummary() []agents.InterviewField {
	return []agents.InterviewField{
		{FieldName: "primary_goal", Value: s.extracted.PrimaryGoal.Value, Confidence: s.extracted.PrimaryGoal.Confidence},
		{FieldName: "goal_timeline_weeks", Value: s.extracted.GoalTimelineWeeks.Value, Confidence: s.extracted.GoalTimelineWeeks.Confidence},
		{FieldName: "training_age_years", Value: s.extracted.TrainingAgeYears.Value, Confidence: s.extracted.TrainingAgeYears.Confidence},
		{FieldName: "programs_tried", Value: s.extracted.ProgramsTried.Value, Confidence: s.extracted.ProgramsTried.Confidence},
		{FieldName: "what_worked", Value: s.extracted.WhatWorked.Value, Confidence: s.extracted.WhatWorked.Confidence},
		{FieldName: "what_failed", Value: s.extracted.WhatFailed.Value, Confidence: s.extracted.WhatFailed.Confidence},
		{FieldName: "typical_session_minutes", Value: s.extracted.TypicalSessionMinutes.Value, Confidence: s.extracted.TypicalSessionMinutes.Confidence},
		{FieldName: "available_days_per_week", Value: s.extracted.AvailableDaysPerWeek.Value, Confidence: s.extracted.AvailableDaysPerWeek.Confidence},
		{FieldName: "equipment_access", Value: s.extracted.EquipmentAccess.Value, Confidence: s.extracted.EquipmentAccess.Confidence},
		{FieldName: "injuries", Value: s.extracted.Injuries.Value, Confidence: s.extracted.Injuries.Confidence},
		{FieldName: "sleep_average_hours", Value: s.extracted.SleepAverageHours.Value, Confidence: s.extracted.SleepAverageHours.Confidence},
		{FieldName: "stress_level", Value: s.extracted.StressLevel.Value, Confidence: s.extracted.StressLevel.Confidence},
	}
}

func (s *session) toProfile() user.Profile {
	p := user.Profile{
		UserID: s.userID,
		Goals: user.GoalProfile{
			Primary:       s.extracted.PrimaryGoal.Value,
			TimelineWeeks: s.extracted.GoalTimelineWeeks.Value,
			WeightKg:      s.extracted.WeightKg.Value,
			BodyFatNow:    s.extracted.BodyFatNow.Value,
		},
		TrainingHistory: user.TrainingHistory{
			TrainingAgeYears:      s.extracted.TrainingAgeYears.Value,
			ProgramsTried:         s.extracted.ProgramsTried.Value,
			WhatWorked:            s.extracted.WhatWorked.Value,
			WhatFailed:            s.extracted.WhatFailed.Value,
			TypicalSessionMinutes: s.extracted.TypicalSessionMinutes.Value,
			AvailableDaysPerWeek:  s.extracted.AvailableDaysPerWeek.Value,
		},
		Equipment: s.extracted.EquipmentAccess.Value,
		Recovery: user.RecoveryProfile{
			SleepAverageHours:  s.extracted.SleepAverageHours.Value,
			ChronicStressLevel: s.extracted.StressLevel.Value,
		},
	}
	for _, inj := range s.extracted.Injuries.Value {
		status := user.InjuryStatusActive
		switch inj.Status {
		case "chronic":
			status = user.InjuryStatusChronic
		case "resolved":
			status = user.InjuryStatusResolved
		}
		p.Injuries = append(p.Injuries, user.Injury{
			ID:       uuid.New().String(),
			BodyPart: inj.BodyPart,
			Type:     inj.Type,
			Status:   status,
			Triggers: inj.Triggers,
			LoadCap:  0.7,
		})
	}
	return p
}

// ── prompts ───────────────────────────────────────────────────────────────

// buildCombinedPrompt produces a single prompt that asks Gemini to both reply
// conversationally AND extract structured fields — one API call per turn.
func buildCombinedPrompt(userMessage string, history []ports.Message) string {
	var sb strings.Builder
	sb.WriteString(`You are TRACE, an AI fitness coach doing an onboarding interview. Your job is to learn the user's fitness history and goals through natural conversation.

Rules:
- Ask ONE question at a time. Follow up on vague answers.
- Be warm and direct, like a knowledgeable coach.
- Gather: primary goal + timeline, training history (programs tried, what worked/failed), session constraints (days/week, duration, equipment), injuries (body part + triggers), recovery (sleep, stress).
- Once you have enough info across ALL areas, end your reply with exactly: [COMPLETE]

Conversation so far:
`)
	for _, msg := range history {
		if msg.Role == ports.RoleUser {
			sb.WriteString("User: ")
		} else {
			sb.WriteString("TRACE: ")
		}
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}
	sb.WriteString("User: ")
	sb.WriteString(userMessage)
	sb.WriteString(`

Now respond with ONLY valid JSON (no markdown, no extra text):
{
  "reply": "<your warm conversational response or follow-up question>",
  "extraction": {
    "primary_goal": {"value": "muscle_gain", "confidence": 0.0},
    "goal_timeline_weeks": {"value": 0, "confidence": 0.0},
    "training_age_years": {"value": 0.0, "confidence": 0.0},
    "programs_tried": {"value": [], "confidence": 0.0},
    "what_worked": {"value": "", "confidence": 0.0},
    "what_failed": {"value": "", "confidence": 0.0},
    "typical_session_minutes": {"value": 60, "confidence": 0.0},
    "available_days_per_week": {"value": 0, "confidence": 0.0},
    "equipment_access": {"value": [], "confidence": 0.0},
    "injuries": {"value": [], "confidence": 0.0},
    "sleep_average_hours": {"value": 0.0, "confidence": 0.0},
    "stress_level": {"value": 5.0, "confidence": 0.0},
    "weight_kg": {"value": 0.0, "confidence": 0.0},
    "body_fat_now": {"value": 0.0, "confidence": 0.0}
  }
}

Confidence: 0.0=not mentioned, 0.3=vaguely implied, 0.5=unclear, 0.8=clearly stated, 1.0=confirmed.
For primary_goal use: muscle_gain, fat_loss, strength, endurance, health, aesthetics
For injuries: {"body_part":"left knee","type":"tendinopathy","status":"active|chronic|resolved","triggers":["squats"]}`)
	return sb.String()
}
