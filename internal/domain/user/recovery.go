package user

// RecoveryProfile holds the user's baseline recovery indicators.
// RecoveryAgent uses these as the floor for allostatic load calculations.
type RecoveryProfile struct {
	SleepAverageHours     float32 `json:"sleep_average_hours"`
	ChronicStressLevel    float32 `json:"chronic_stress_level"` // 1–10
	NutritionPattern      string  `json:"nutrition_pattern"`    // free text
	RecoverySignalsNoticed string  `json:"recovery_signals_noticed"` // e.g. "resting HR up 8bpm = overtrained"
}
