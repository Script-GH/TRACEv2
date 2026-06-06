package workout

// Exercise is a record from the exercise library — the catalogue of possible movements.
// Not a planned exercise — just the definition.
// Returned by the get_exercise_library MCP tool, filtered by equipment and restrictions.
type Exercise struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	MuscleGroups    []string `json:"muscle_groups"`
	Equipment       []string `json:"equipment"`
	Contraindications []string `json:"contraindications"` // body parts this is hard on
	Substitutes     []string `json:"substitutes"`       // IDs of safe alternatives
}

// ProgressionEntry is a single data point in per-exercise weight/rep history.
// Used by ProgressionAgent to compute overload targets.
type ProgressionEntry struct {
	LogID     string  `json:"log_id"`
	Date      string  `json:"date"`
	Sets      int     `json:"sets"`
	Reps      string  `json:"reps"`
	WeightKg  float32 `json:"weight_kg"`
	RPE       float32 `json:"rpe"`
}
