package user

// InjuryStatus reflects the current state of an injury.
type InjuryStatus string

const (
	InjuryStatusActive   InjuryStatus = "active"
	InjuryStatusChronic  InjuryStatus = "chronic"
	InjuryStatusResolved InjuryStatus = "resolved"
)

// Injury encodes a single injury record.
// Active and chronic injuries feed directly into InjuryAgent veto logic.
// Triggers are the movement patterns that cause flare-ups — matched against
// the exercise library in the MCP server.
type Injury struct {
	ID       string       `json:"id"`
	BodyPart string       `json:"body_part"`
	Type     string       `json:"type"`       // e.g. "patellar tendinopathy", "ACL repair"
	Status   InjuryStatus `json:"status"`
	Triggers []string     `json:"triggers"`   // e.g. ["high volume squats", "running downhill"]
	LoadCap  float32      `json:"load_cap"`   // 0–1; relative max load for this joint
	Notes    string       `json:"notes"`
}
