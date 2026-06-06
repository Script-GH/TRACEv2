package user

// Profile is the root aggregate for all user data.
// Stored in Redis encrypted at rest (AES-256).
// Written ONLY through MCP tools — never directly from agent code.
type Profile struct {
	UserID          string          `json:"user_id"`
	CreatedAt       int64           `json:"created_at"`
	Goals           GoalProfile     `json:"goals"`
	TrainingHistory TrainingHistory `json:"training_history"`
	Injuries        []Injury        `json:"injuries"`
	Recovery        RecoveryProfile `json:"recovery_profile"`
	Equipment       []string        `json:"equipment"`
	GraphVersion    int             `json:"graph_version"` // tracks which graph version matches this profile
	InterviewDone   bool            `json:"interview_done"`
}

// GoalProfile captures the user's fitness objectives and body composition baseline.
type GoalProfile struct {
	Primary        string  `json:"primary"`        // "visible_abs" | "weight_loss" | "strength" | "endurance"
	Secondary      string  `json:"secondary"`
	TimelineWeeks  int     `json:"timeline_weeks"`
	BodyFatNow     float32 `json:"body_fat_now"`
	BodyFatTarget  float32 `json:"body_fat_target"`
	WeightKg       float32 `json:"weight_kg"`
	WeightTarget   float32 `json:"weight_target"`
}

// TrainingHistory captures the user's fitness background.
// Used to seed the initial causal graph and calibrate adaptation speed.
type TrainingHistory struct {
	TrainingAgeYears      float32  `json:"training_age_years"`
	LastConsistentPeriod  string   `json:"last_consistent_period"` // free text, e.g. "6 months of PPL ended 4 months ago"
	ProgramsTried         []string `json:"programs_tried"`
	WhatWorked            string   `json:"what_worked"`
	WhatFailed            string   `json:"what_failed"`
	TypicalSessionMinutes int      `json:"typical_session_minutes"`
	AvailableDaysPerWeek  int      `json:"available_days_per_week"`
}
