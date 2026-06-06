package orchestrator

import (
	"context"
	"sync"

	"github.com/trace/trace/internal/application/agents"
	"github.com/trace/trace/internal/application/ports"
	"github.com/trace/trace/internal/delivery/sse"
)

// DailyLoopOrchestrator handles the AWAITING_LOG → WORKOUT_READY cycle.
// This is the hot path — runs after every workout log.
type DailyLoopOrchestrator struct {
	logAgent    agents.LogAgent
	twin        agents.TwinAgent
	moeGate     agents.MoEGateAgent
	recovery    agents.RecoveryAgent
	injury      agents.InjuryAgent
	growth      agents.GrowthAgent
	metabolic   agents.MetabolicAgent
	progression agents.ProgressionAgent
	planner     agents.PlannerAgent
	coach       agents.CoachAgent
	mcp         ports.MCPClient
	broker      ports.MessageBroker
	sse         *sse.Broker
}

func NewDailyLoopOrchestrator(
	logAgent agents.LogAgent,
	twin agents.TwinAgent,
	moeGate agents.MoEGateAgent,
	recovery agents.RecoveryAgent,
	injury agents.InjuryAgent,
	growth agents.GrowthAgent,
	metabolic agents.MetabolicAgent,
	progression agents.ProgressionAgent,
	planner agents.PlannerAgent,
	coach agents.CoachAgent,
	mcp ports.MCPClient,
	broker ports.MessageBroker,
	sseBroker *sse.Broker,
) *DailyLoopOrchestrator {
	return &DailyLoopOrchestrator{
		logAgent:    logAgent,
		twin:        twin,
		moeGate:     moeGate,
		recovery:    recovery,
		injury:      injury,
		growth:      growth,
		metabolic:   metabolic,
		progression: progression,
		planner:     planner,
		coach:       coach,
		mcp:         mcp,
		broker:      broker,
		sse:         sseBroker,
	}
}

// HandleLogActive processes one logging conversation turn.
func (d *DailyLoopOrchestrator) HandleLogActive(ctx context.Context, session *SessionContext) (Event, error) {
	panic("not implemented")
}

// HandleGraphUpdate calls TwinAgent.Update and stores the new graph version in session.
func (d *DailyLoopOrchestrator) HandleGraphUpdate(ctx context.Context, session *SessionContext) (Event, error) {
	panic("not implemented")
}

// HandleMoEGating calls MoEGateAgent.AssignWeights and stores the weight vector.
func (d *DailyLoopOrchestrator) HandleMoEGating(ctx context.Context, session *SessionContext) (Event, error) {
	panic("not implemented")
}

// HandleAgentsRun dispatches all 5 expert agents in parallel using goroutines.
// Each agent publishes its result as an A2A REPORT message via the broker.
// InjuryAgent and RecoveryAgent may publish VETO messages.
// Returns EventVetoFired if any veto was received; otherwise EventAllAgentsDone.
func (d *DailyLoopOrchestrator) HandleAgentsRun(ctx context.Context, session *SessionContext) (Event, error) {
	var (
		wg              sync.WaitGroup
		recoveryReport  agents.RecoveryReport
		growthReport    agents.GrowthReport
		metabolicReport agents.MetabolicReport
		progressionReport agents.ProgressionReport
		injuryReport    agents.InjuryReport
		errs            []error
		mu              sync.Mutex
	)

	// Retrieve shared inputs before parallel dispatch
	weights := weightsFromSession(session)
	userID := session.UserID
	graphVer := session.GraphVersion

	input := func(w float32) agents.ExpertInput {
		return agents.ExpertInput{
			UserID:       userID,
			SessionID:    session.SessionID,
			GraphVersion: graphVer,
			Weight:       w,
		}
	}

	// --- parallel fan-out ---
	wg.Add(5)

	go func() {
		defer wg.Done()
		r, err := d.recovery.Evaluate(ctx, input(weights.Recovery))
		mu.Lock(); recoveryReport = r; if err != nil { errs = append(errs, err) }; mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		r, err := d.growth.Evaluate(ctx, input(weights.Growth))
		mu.Lock(); growthReport = r; if err != nil { errs = append(errs, err) }; mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		r, err := d.metabolic.Evaluate(ctx, input(weights.Metabolic))
		mu.Lock(); metabolicReport = r; if err != nil { errs = append(errs, err) }; mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		// ProgressionAgent needs the exercise list from GrowthAgent.
		// In the real implementation this waits on growthReport via a secondary sync.
		r, err := d.progression.Evaluate(ctx, input(weights.Progression), nil)
		mu.Lock(); progressionReport = r; if err != nil { errs = append(errs, err) }; mu.Unlock()
	}()

	go func() {
		defer wg.Done()
		// InjuryAgent needs the proposed exercise list from GrowthAgent.
		// In the real implementation this waits on growthReport via a secondary sync.
		r, err := d.injury.Evaluate(ctx, input(weights.Injury), nil)
		mu.Lock(); injuryReport = r; if err != nil { errs = append(errs, err) }; mu.Unlock()
	}()

	wg.Wait()

	// Store reports in session for the next state handler
	storeAgentReports(session, recoveryReport, injuryReport, growthReport, metabolicReport, progressionReport)

	if len(injuryReport.VetoedExercises) > 0 || recoveryReport.ForceRestDay {
		return EventVetoFired, nil
	}
	return EventAllAgentsDone, nil
}

// HandleVetoResolve processes veto messages and updates the exercise list.
func (d *DailyLoopOrchestrator) HandleVetoResolve(ctx context.Context, session *SessionContext) (Event, error) {
	panic("not implemented")
}

// HandlePlanSynth calls PlannerAgent.Synthesise with all expert reports.
func (d *DailyLoopOrchestrator) HandlePlanSynth(ctx context.Context, session *SessionContext) (Event, error) {
	panic("not implemented")
}

// HandleCoachWrite calls CoachAgent.Annotate to add causal reasoning per exercise.
// On CoachUnavailableError: plan is still delivered with empty reasoning fields.
func (d *DailyLoopOrchestrator) HandleCoachWrite(ctx context.Context, session *SessionContext) (Event, error) {
	panic("not implemented")
}

// HandleWorkoutReady pushes the completed plan to the user via SSE.
func (d *DailyLoopOrchestrator) HandleWorkoutReady(ctx context.Context, session *SessionContext) (Event, error) {
	panic("not implemented")
}

// Handlers returns the StateHandler map for the daily loop.
func (d *DailyLoopOrchestrator) Handlers() map[State]StateHandler {
	return map[State]StateHandler{
		StateLogActive:    d.HandleLogActive,
		StateGraphUpdate:  d.HandleGraphUpdate,
		StateMoEGating:    d.HandleMoEGating,
		StateAgentsRun:    d.HandleAgentsRun,
		StateVetoResolve:  d.HandleVetoResolve,
		StatePlanSynth:    d.HandlePlanSynth,
		StateCoachWrite:   d.HandleCoachWrite,
		StateWorkoutReady: d.HandleWorkoutReady,
	}
}

// --- session-scoped agent report cache (stored as typed fields on a separate struct) ---
// storeAgentReports and weightsFromSession are stubs that will manipulate
// a typed payload attached to SessionContext in the full implementation.

type agentReports struct {
	weights     agents.WeightVector
	recovery    agents.RecoveryReport
	injury      agents.InjuryReport
	growth      agents.GrowthReport
	metabolic   agents.MetabolicReport
	progression agents.ProgressionReport
}

func storeAgentReports(session *SessionContext, r agents.RecoveryReport, i agents.InjuryReport, g agents.GrowthReport, m agents.MetabolicReport, p agents.ProgressionReport) {
	panic("not implemented")
}

func weightsFromSession(session *SessionContext) agents.WeightVector {
	panic("not implemented")
}
