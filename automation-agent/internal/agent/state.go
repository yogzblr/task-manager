package agent

import (
	"context"
	"fmt"
	"sync"
)

// State represents agent state
type State string

const (
	StateUnregistered State = "unregistered"
	StateRegistering  State = "registering"
	StateIdle         State = "idle"
	StateLeasing      State = "leasing"
	StateExecuting    State = "executing"
	StateUpgrading    State = "upgrading"
)

// StateMachine manages agent state transitions
type StateMachine struct {
	mu    sync.RWMutex
	state State
}

// NewStateMachine creates a new state machine
func NewStateMachine() *StateMachine {
	return &StateMachine{
		state: StateUnregistered,
	}
}

// Current returns the current state
func (sm *StateMachine) Current() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.state
}

// Transition transitions to a new state
func (sm *StateMachine) Transition(newState State) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if !sm.isValidTransition(sm.state, newState) {
		return fmt.Errorf("invalid state transition: %s -> %s", sm.state, newState)
	}
	
	sm.state = newState
	return nil
}

// isValidTransition checks if a state transition is valid
func (sm *StateMachine) isValidTransition(from, to State) bool {
	validTransitions := map[State][]State{
		StateUnregistered: {StateRegistering},
		StateRegistering:  {StateIdle, StateUnregistered},
		StateIdle:         {StateLeasing, StateUpgrading},
		StateLeasing:      {StateExecuting, StateIdle},
		StateExecuting:    {StateIdle},
		StateUpgrading:    {StateIdle, StateUnregistered},
	}
	
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	
	for _, state := range allowed {
		if state == to {
			return true
		}
	}
	
	return false
}

// Agent represents the automation agent
type Agent struct {
	ID        string
	TenantID  string
	ProjectID string
	OS        string
	Labels    map[string]string
	
	StateMachine *StateMachine
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewAgent creates a new agent
func NewAgent(id, tenantID, projectID, os string) *Agent {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Agent{
		ID:           id,
		TenantID:     tenantID,
		ProjectID:    projectID,
		OS:           os,
		Labels:       make(map[string]string),
		StateMachine: NewStateMachine(),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start starts the agent
func (a *Agent) Start() error {
	return a.StateMachine.Transition(StateRegistering)
}

// Stop stops the agent
func (a *Agent) Stop() error {
	a.cancel()
	return a.StateMachine.Transition(StateUnregistered)
}

// State returns the current state
func (a *Agent) State() State {
	return a.StateMachine.Current()
}
