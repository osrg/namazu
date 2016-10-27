package signal

import (
	"time"
)

// Event signal interface (inspector->orchestrator)
type Event interface {
	// these fields are same as in Signal
	ID() string
	EntityID() string
	ArrivedTime() time.Time
	JSONMap() map[string]interface{}
	String() string

	// comparator, excluding uuid
	Equals(o Event) bool

	// if deferred, the inspector is waiting for an action from the orchestrator.
	//
	// json name: "deferred"
	Deferred() bool

	// explore policy can use this hash string as a hint for semi-deterministic replaying.
	// The hint should not contain time-dependent or random things for better determinism.
	// Note that we will not support fully deterministic replaying.
	//
	// The hint can contain any character.
	//
	// json name: "replay_hint"
	ReplayHint() string

	// default positive action. can be NopAction, but cannot be nil.
	// (NopAction is used for history storage)
	DefaultAction() (Action, error)

	// default negative action. can be nil.
	DefaultFaultAction() (Action, error)
}

// Action signal interface (orchestrator->inspector)
type Action interface {
	// these fields are same as in Signal
	ID() string
	EntityID() string
	ArrivedTime() time.Time
	JSONMap() map[string]interface{}
	String() string

	// comparator, excluding uuid
	Equals(o Action) bool

	// triggered time (only orchestrator should call this)
	TriggeredTime() time.Time

	// set triggered time (only orchestrator should call this)
	SetTriggeredTime(time.Time)

	// in fault actions, can be nil (but not always)
	Event() Event
}

const (
	ControlEnableOrchestration = iota
	ControlDisableOrchestration
)

type Control struct {
	Op int
}

type OrchestratorSideAction interface {
	// if true, the action will not be propagated to inspectors.
	//
	// this field exists mainly for NopAction(for syslog events) and some fault actions
	OrchestratorSideOnly() bool

	// execute shell command or something else
	ExecuteOnOrchestrator() error
}
