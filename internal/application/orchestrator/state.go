package orchestrator

// State represents a node in the LangGraph state machine.
type State string

const (
	// Onboarding path
	StateNewUser          State = "NEW_USER"
	StateInterviewActive  State = "INTERVIEW_ACTIVE"
	StateInterviewComplete State = "INTERVIEW_COMPLETE"
	StateGraphBuildInitial State = "GRAPH_BUILD_INITIAL"
	StateFirstWorkoutGen  State = "FIRST_WORKOUT_GEN"

	// Daily loop path
	StateAwaitingLog  State = "AWAITING_LOG"
	StateLogActive    State = "LOG_ACTIVE"
	StateLogComplete  State = "LOG_COMPLETE"
	StateGraphUpdate  State = "GRAPH_UPDATE"
	StateMoEGating    State = "MOE_GATING"
	StateAgentsRun    State = "AGENTS_RUN"
	StateVetoResolve  State = "VETO_RESOLVE"
	StatePlanSynth    State = "PLAN_SYNTH"
	StateCoachWrite   State = "COACH_WRITE"
	StateWorkoutReady State = "WORKOUT_READY"
)

// Event drives transitions in the state machine.
type Event string

const (
	// Onboarding events
	EventUserCreated       Event = "USER_CREATED"
	EventInterviewTurn     Event = "INTERVIEW_TURN"          // a conversation turn processed
	EventAllFieldsExtracted Event = "ALL_FIELDS_EXTRACTED"   // all fields ≥ 0.8 confidence
	EventUserConfirmedSummary Event = "USER_CONFIRMED_SUMMARY"
	EventGraphBuilt        Event = "GRAPH_BUILT"
	EventWorkoutGenTriggered Event = "WORKOUT_GEN_TRIGGERED"

	// Daily loop events
	EventUserOpenedLog   Event = "USER_OPENED_LOG"
	EventLogTurn         Event = "LOG_TURN"
	EventLogConfirmed    Event = "LOG_CONFIRMED"
	EventGraphUpdated    Event = "GRAPH_UPDATED"
	EventWeightsAssigned Event = "WEIGHTS_ASSIGNED"
	EventAllAgentsDone   Event = "ALL_AGENTS_DONE"
	EventVetoFired       Event = "VETO_FIRED"
	EventVetoResolved    Event = "VETO_RESOLVED"
	EventPlanSynthesised Event = "PLAN_SYNTHESISED"
	EventReasoningAdded  Event = "REASONING_ADDED"
	EventWorkoutPushed   Event = "WORKOUT_PUSHED"
)

// Transition maps a (State, Event) pair to the next State.
type Transition struct {
	From  State
	On    Event
	To    State
}

// AllTransitions is the complete state machine transition table.
var AllTransitions = []Transition{
	// Onboarding path
	{StateNewUser, EventUserCreated, StateInterviewActive},
	{StateInterviewActive, EventInterviewTurn, StateInterviewActive},       // loop until complete
	{StateInterviewActive, EventAllFieldsExtracted, StateInterviewComplete},
	{StateInterviewComplete, EventUserConfirmedSummary, StateGraphBuildInitial},
	{StateGraphBuildInitial, EventGraphBuilt, StateFirstWorkoutGen},
	{StateFirstWorkoutGen, EventWorkoutGenTriggered, StateAwaitingLog},

	// Daily loop path
	{StateAwaitingLog, EventUserOpenedLog, StateLogActive},
	{StateLogActive, EventLogTurn, StateLogActive},         // loop until complete
	{StateLogActive, EventLogConfirmed, StateLogComplete},
	{StateLogComplete, EventLogConfirmed, StateGraphUpdate},
	{StateGraphUpdate, EventGraphUpdated, StateMoEGating},
	{StateMoEGating, EventWeightsAssigned, StateAgentsRun},
	{StateAgentsRun, EventAllAgentsDone, StatePlanSynth},
	{StateAgentsRun, EventVetoFired, StateVetoResolve},
	{StateVetoResolve, EventVetoResolved, StatePlanSynth},
	{StatePlanSynth, EventPlanSynthesised, StateCoachWrite},
	{StateCoachWrite, EventReasoningAdded, StateWorkoutReady},
	{StateWorkoutReady, EventWorkoutPushed, StateAwaitingLog}, // cycle
}

// SessionContext is the in-flight state for one user session.
type SessionContext struct {
	SessionID    string
	UserID       string
	CurrentState State
	GraphVersion int
	LogID        string  // populated after LOG_COMPLETE
	PlanID       string  // populated after PLAN_SYNTH
}
