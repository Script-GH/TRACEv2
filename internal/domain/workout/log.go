package workout

// PainType classifies the sensation of a reported pain signal.
type PainType string

const (
	PainTypeSharp     PainType = "SHARP"
	PainTypeAche      PainType = "ACHE"
	PainTypeTightness PainType = "TIGHTNESS"
	PainTypeFatigue   PainType = "FATIGUE"
)

// PainSignal is a single pain report from a workout log.
// Severity > 6 triggers immediate TwinAgent edge creation (spec requirement).
type PainSignal struct {
	BodyPart        string   `json:"body_part"`
	Severity        float32  `json:"severity"`         // 1–10
	Type            PainType `json:"type"`
	DuringExercise  string   `json:"during_exercise"`
}

// ExerciseSet is one exercise performed in a session.
type ExerciseSet struct {
	Exercise   string  `json:"exercise"`
	Sets       int     `json:"sets"`
	Reps       string  `json:"reps"`       // range allowed, e.g. "8-10"
	WeightKg   float32 `json:"weight_kg"`
	Difficulty float32 `json:"difficulty"` // 1–10
}

// Log is the structured output of a LogAgent conversation.
// Every field in this struct is a potential input to TwinAgent's graph update.
type Log struct {
	LogID          string        `json:"log_id"`
	UserID         string        `json:"user_id"`
	SessionDate    string        `json:"session_date"` // "2024-01-15"
	Exercises      []ExerciseSet `json:"exercises"`
	RPE            float32       `json:"rpe"`             // 1–10 overall session RPE
	EnergyLevel    float32       `json:"energy_level"`    // 1–10
	PainSignals    []PainSignal  `json:"pain_signals"`
	SleepLastNight float32       `json:"sleep_hours"`
	StressToday    float32       `json:"stress_level"`    // 1–10
	NutritionScore float32       `json:"nutrition_score"` // 1–10 rough self-report
	Notes          string        `json:"notes"`
	CreatedAt      int64         `json:"created_at"`
}
