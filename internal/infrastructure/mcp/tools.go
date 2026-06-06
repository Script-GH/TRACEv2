package mcp

import (
	"context"
	"fmt"

	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/domain"
	"github.com/trace/trace/internal/domain/graph"
	"github.com/trace/trace/internal/domain/user"
	"github.com/trace/trace/internal/domain/workout"
)

// tools.go implements all MCPClient tool methods on Server.
// Each method validates inputs and delegates to the appropriate repository.

func (s *Server) GetCausalGraph(ctx context.Context, userID string, domainFilter string) (*graph.CausalGraph, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID required")
	}
	return s.graphs.Get(ctx, userID)
}

func (s *Server) UpdateCausalGraph(ctx context.Context, userID string, g *graph.CausalGraph) (int, error) {
	if userID == "" || g == nil {
		return 0, fmt.Errorf("userID and graph required")
	}
	g.UserID = userID
	if err := s.graphs.Save(ctx, g); err != nil {
		return 0, err
	}
	return g.Version, nil
}

func (s *Server) QueryCausalPaths(ctx context.Context, userID string, query ports.CausalPathQuery) ([]graph.CausalPath, error) {
	g, err := s.graphs.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Simple linear scan: find edges matching polarity + confidence, return single-hop paths.
	var paths []graph.CausalPath
	for _, edge := range g.Edges {
		if edge.Confidence < query.MinConfidence {
			continue
		}
		if query.Polarity != "" && edge.Polarity != query.Polarity {
			continue
		}
		fromNode := g.Nodes[edge.FromNode]
		toNode := g.Nodes[edge.ToNode]
		if fromNode == nil || toNode == nil {
			continue
		}
		if query.FromNodeLabel != "" && fromNode.Label != query.FromNodeLabel {
			continue
		}
		if query.ToNodeLabel != "" && toNode.Label != query.ToNodeLabel {
			continue
		}
		paths = append(paths, graph.CausalPath{
			Nodes:           []*graph.Node{fromNode, toNode},
			Edges:           []*graph.Edge{edge},
			TotalConfidence: edge.Confidence,
		})
	}
	return paths, nil
}

func (s *Server) GetWorkoutLogHistory(ctx context.Context, userID string, n int) ([]workout.Log, error) {
	if s.logs == nil {
		return nil, nil
	}
	return s.logs.GetHistory(ctx, userID, n)
}

func (s *Server) GetProgressionHistory(ctx context.Context, userID string, exercise string) ([]workout.ProgressionEntry, error) {
	if s.progression == nil {
		return nil, nil
	}
	return s.progression.GetHistory(ctx, userID, exercise)
}

func (s *Server) GetProgressionHistoryTool(ctx context.Context, userID string, exercise string) ([]workout.ProgressionEntry, error) {
	return s.GetProgressionHistory(ctx, userID, exercise)
}

func (s *Server) GetUserProfile(ctx context.Context, userID string) (*user.Profile, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID required")
	}
	return s.users.Get(ctx, userID)
}

func (s *Server) SaveUserProfile(ctx context.Context, profile *user.Profile) error {
	if profile == nil || profile.UserID == "" {
		return fmt.Errorf("profile and userID required")
	}
	return s.users.Save(ctx, profile)
}

func (s *Server) GetInjuryConstraints(ctx context.Context, userID string) ([]graph.InjuryConstraint, error) {
	profile, err := s.users.Get(ctx, userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, nil
		}
		return nil, err
	}
	constraints := make([]graph.InjuryConstraint, 0, len(profile.Injuries))
	for _, inj := range profile.Injuries {
		if inj.Status == user.InjuryStatusResolved {
			continue
		}
		constraints = append(constraints, graph.InjuryConstraint{
			BodyPart:        inj.BodyPart,
			Status:          string(inj.Status),
			ContraExercises: inj.Triggers,
			LoadCap:         inj.LoadCap,
			Triggers:        inj.Triggers,
		})
	}
	return constraints, nil
}

func (s *Server) GetExerciseLibrary(ctx context.Context, filter ports.ExerciseFilter) ([]workout.Exercise, error) {
	lib := mvpExerciseLibrary()
	if len(filter.Equipment) == 0 && len(filter.MuscleGroups) == 0 && len(filter.ExcludeContra) == 0 {
		return lib, nil
	}
	var out []workout.Exercise
	for _, ex := range lib {
		if len(filter.ExcludeContra) > 0 && hasOverlap(ex.Contraindications, filter.ExcludeContra) {
			continue
		}
		out = append(out, ex)
	}
	return out, nil
}

func (s *Server) StoreWorkoutPlan(ctx context.Context, plan *workout.Plan) (string, error) {
	if s.plans == nil {
		return plan.PlanID, nil
	}
	return plan.PlanID, s.plans.Save(ctx, plan)
}

func (s *Server) LogVetoEvent(ctx context.Context, event ports.VetoEvent) (string, error) {
	// Non-fatal for MVP — log and continue.
	return "veto-" + event.UserID, nil
}

func (s *Server) FlagConflict(ctx context.Context, conflict ports.ConflictRecord) (string, error) {
	return "conflict-" + conflict.UserID, nil
}

func (s *Server) StoreCoTTrace(ctx context.Context, trace ports.CoTTrace) (string, error) {
	if s.cot == nil {
		return trace.TraceID, nil
	}
	return trace.TraceID, s.cot.Save(ctx, trace)
}

func (s *Server) GetCoTTrace(ctx context.Context, traceID string) (*ports.CoTTrace, error) {
	if s.cot == nil {
		return nil, domain.ErrMCPToolNotFound
	}
	return s.cot.Get(ctx, traceID)
}

func hasOverlap(a, b []string) bool {
	set := make(map[string]struct{}, len(a))
	for _, v := range a {
		set[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := set[v]; ok {
			return true
		}
	}
	return false
}

// mvpExerciseLibrary returns a hardcoded starter exercise library for demo purposes.
func mvpExerciseLibrary() []workout.Exercise {
	return []workout.Exercise{
		{ID: "squat", Name: "Barbell Back Squat", MuscleGroups: []string{"quads", "glutes", "hamstrings"}, Equipment: []string{"barbell", "rack"}, Contraindications: []string{"knee", "lower_back"}, Substitutes: []string{"goblet_squat", "leg_press"}},
		{ID: "deadlift", Name: "Conventional Deadlift", MuscleGroups: []string{"hamstrings", "glutes", "back"}, Equipment: []string{"barbell"}, Contraindications: []string{"lower_back"}, Substitutes: []string{"rdl", "trap_bar_deadlift"}},
		{ID: "bench_press", Name: "Barbell Bench Press", MuscleGroups: []string{"chest", "triceps", "shoulders"}, Equipment: []string{"barbell", "bench"}, Contraindications: []string{"shoulder"}, Substitutes: []string{"dumbbell_press", "pushup"}},
		{ID: "overhead_press", Name: "Overhead Press", MuscleGroups: []string{"shoulders", "triceps"}, Equipment: []string{"barbell"}, Contraindications: []string{"shoulder", "neck"}, Substitutes: []string{"dumbbell_shoulder_press", "landmine_press"}},
		{ID: "pullup", Name: "Pull-up", MuscleGroups: []string{"lats", "biceps", "upper_back"}, Equipment: []string{"pullup_bar"}, Contraindications: []string{"shoulder", "elbow"}, Substitutes: []string{"lat_pulldown", "assisted_pullup"}},
		{ID: "row", Name: "Barbell Row", MuscleGroups: []string{"upper_back", "lats", "biceps"}, Equipment: []string{"barbell"}, Contraindications: []string{"lower_back"}, Substitutes: []string{"cable_row", "dumbbell_row"}},
		{ID: "rdl", Name: "Romanian Deadlift", MuscleGroups: []string{"hamstrings", "glutes"}, Equipment: []string{"barbell"}, Contraindications: []string{"lower_back"}, Substitutes: []string{"leg_curl", "nordic_curl"}},
		{ID: "dip", Name: "Dip", MuscleGroups: []string{"chest", "triceps"}, Equipment: []string{"dip_bars"}, Contraindications: []string{"shoulder", "elbow"}, Substitutes: []string{"tricep_pushdown", "close_grip_bench"}},
		{ID: "goblet_squat", Name: "Goblet Squat", MuscleGroups: []string{"quads", "glutes"}, Equipment: []string{"dumbbell", "kettlebell"}, Contraindications: []string{"knee"}, Substitutes: []string{"leg_press"}},
		{ID: "lat_pulldown", Name: "Lat Pulldown", MuscleGroups: []string{"lats", "biceps"}, Equipment: []string{"cable_machine"}, Contraindications: []string{"shoulder"}, Substitutes: []string{"pullup"}},
		{ID: "leg_press", Name: "Leg Press", MuscleGroups: []string{"quads", "glutes"}, Equipment: []string{"leg_press_machine"}, Contraindications: []string{"knee", "lower_back"}, Substitutes: []string{"goblet_squat"}},
		{ID: "hip_thrust", Name: "Hip Thrust", MuscleGroups: []string{"glutes", "hamstrings"}, Equipment: []string{"barbell", "bench"}, Contraindications: []string{"lower_back"}, Substitutes: []string{"glute_bridge"}},
		{ID: "face_pull", Name: "Face Pull", MuscleGroups: []string{"rear_delts", "upper_back"}, Equipment: []string{"cable_machine"}, Contraindications: []string{}, Substitutes: []string{"band_pull_apart"}},
		{ID: "plank", Name: "Plank", MuscleGroups: []string{"core"}, Equipment: []string{}, Contraindications: []string{"lower_back"}, Substitutes: []string{"dead_bug"}},
		{ID: "running", Name: "Running", MuscleGroups: []string{"legs", "cardiovascular"}, Equipment: []string{}, Contraindications: []string{"knee", "ankle"}, Substitutes: []string{"cycling", "rowing"}},
	}
}
