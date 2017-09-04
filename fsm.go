// Simple finite state machine implementation
// Supports nested states, state entry and transition actions,
// uses multi-level contexts for nested states.
// Normal operation flow:
// * TBD
package simple_fsm

import (
	"bytes"
	"fmt"
)

const (
	FsmGlobalStateName        = "global"
	FsmResultCtxMemberName    = "result"
	FsmDefaultHistoryCapacity = 10
	FsmAutoStatesCount        = 1
)

// Fsm
// Describes finite state machine
// Contains state meta info, machine entry point and associated data (contexts)
type Fsm struct {
	states  map[string]*StateInfo
	start   *StateInfo
	stack   ContextStack
	history History
	fatal   *FsmError
}

// NewFsm
// Constructs new state machine, initializes auto states and structures
func NewFsm() *Fsm {
	fsm := Fsm{
		states:  make(map[string]*StateInfo),
		start:   NewState(FsmGlobalStateName, nil),
		stack:   newContextStack(),
		history: make([]HistoryItem, 0, FsmDefaultHistoryCapacity),
		fatal:   nil,
	}
	fsm.addStateImpl(fsm.start, nil, true, false)
	fsm.initAutoStates()

	return &fsm
}

// Reset
// Resets FSM to state, ready for execution (initial)
// Progress/results from previous run is discarded
func (fsm *Fsm) Reset() {
	fsm.stack = newContextStack()
	fsm.initAutoStates()
	fsm.history = make([]HistoryItem, 0, FsmDefaultHistoryCapacity)
	fsm.fatal = nil
}

// initAutoStates
// Populates stack with automatic stuff (default global state, as of now)
func (fsm *Fsm) initAutoStates() {
	if !fsm.stack.Empty() {
		return
	}
	fsm.stack.Push(fsm.start)
}

// Fatal
// Check if fatal error is occured and FSM went into fatal state
// Note: Fatal() implies Completed()
func (fsm *Fsm) Fatal() bool {
	return fsm.fatal != nil
}

// Running
// Check is FSM execution is in progress
func (fsm *Fsm) Running() bool {
	return fsm.stack.Depth() > FsmAutoStatesCount &&
		!fsm.Fatal() &&
		!fsm.Completed()
}

// Completed
// Checks if FSM execution is done and there's a result to grab
func (fsm *Fsm) Completed() bool {
	return !fsm.Fatal() &&
		fsm.stack.Depth() > FsmAutoStatesCount &&
		fsm.stack.Peek().state.Final()
}

// Idle
// Check if FSM is not running, completed nor stopped due to fatal error
func (fsm *Fsm) Idle() bool {
	return !fsm.Running() &&
		!fsm.Completed() &&
		!fsm.Fatal()
}

// fatalError
// Returns fatal error object in case FSM is in fatal state
func (fsm *Fsm) fatalError() *FsmError {
	if !fsm.Fatal() {
		return nil
	}
	return fsm.fatal
}

// Result
// Returns final FSM execution result in case it's completed
func (fsm *Fsm) Result() (value interface{}, err *FsmError) {
	if !fsm.Completed() {
		err = newFsmErrorWrongFlow("get result", "not completed")
		return
	}
	if fsm.Fatal() {
		err = fsm.fatalError()
		return
	}

	value, err = fsm.stack.Global().context.Raw(FsmResultCtxMemberName)
	return
}

func (fsm *Fsm) History() History {
	return fsm.history
}

// AddStartState
// Validates and adds a start (sub)state to the state machine
func (fsm *Fsm) AddStartState(state *StateInfo, parent *StateInfo) (err *FsmError) {
	return fsm.addStateImpl(state, parent, true, true)
}

// AddStartState
// Validates and adds an intermediate (sub)state to the state machine
func (fsm *Fsm) AddState(state *StateInfo, parent *StateInfo) (err *FsmError) {
	return fsm.addStateImpl(state, parent, false, true)
}

// AddStates
// Allows to add a bunch of (sub)states (including starting one) to the state machine
func (fsm *Fsm) AddStates(parent *StateInfo, start *StateInfo, states ...*StateInfo) (err *FsmError) {
	if start != nil {
		if err = fsm.AddStartState(start, parent); err != nil {
			return
		}
	}
	for _, state := range states {
		if err = fsm.AddState(state, parent); err != nil {
			break
		}
	}
	return
}

// addStateImpl
// Adds a state to the state machine, validating it beforehand
func (fsm *Fsm) addStateImpl(state *StateInfo, parent *StateInfo, start bool, autoAdopt bool) (err *FsmError) {
	switch {
	case !fsm.Idle():
		return newFsmErrorWrongFlow("add state", "running/completed/fatal")
	case state == nil:
		return newFsmErrorStateIsInvalid(state, "state is nil")
	case fsm.start == nil:
		return newFsmErrorInvalid("global state is not defined")
	case start && parent == nil && !fsm.start.Final():
		cause := fmt.Sprintf("start state is already set to \"%s\"", fsm.start.Transitions[0].Name)
		return newFsmErrorInvalid(cause)
	case start && parent != nil && len(parent.Transitions) > 0:
		cause := "parent should not have transitions (transition to start sub state is added automatically)"
		return newFsmErrorInvalid(cause)
	}

	if err = state.Validate(); err != nil {
		return
	}
	if _, present := fsm.states[state.Name]; present {
		return newFsmErrorStateAlreadyExists(state.Name)
	}

	fsm.states[state.Name] = state
	if parent == nil {
		if autoAdopt {
			err = fsm.start.addSubState(state, start)
		}
	} else {
		if _, present := fsm.states[parent.Name]; !present {
			err = newFsmErrorInvalid("Parent state was not found (forgot to add?)")
		} else {
			err = parent.addSubState(state, start)
		}
	}
	if err != nil {
		return
	}

	if start && autoAdopt {
		state.Parent.Transitions = NewTransitionAlways("start", state.Name, nil)
	}
	return
}

// Validate
// Checks if FSM structure is consistent:
// * no transitions to unknown
// * no dead states
func (fsm *Fsm) Validate() (err *FsmError) {
	state_refs := make(map[string]bool)
	for k, _ := range fsm.states {
		state_refs[k] = false
	}
	state_refs[fsm.start.Name] = true

	// TODO: 1 start substate can't belong to many parents

	for _, s := range fsm.states {
		if err := s.Validate(); err != nil {
			return err
		}
		for _, tr := range s.Transitions {
			if _, present := fsm.states[tr.ToState]; !present {
				cause := fmt.Sprintf(
					"transition \"%s\" of state \"%s\" has unknown destination \"%s\"",
					tr.Name,
					s.Name,
					tr.ToState,
				)
				return newFsmErrorInvalid(cause)
			}
			if ancestor, _ := findCommonAncestor(s, fsm.states[tr.ToState]); ancestor == nil {
				cause := fmt.Sprintf("\"%s\" and \"%s\" don't have a common parent", s.Name, tr.ToState)
				return newFsmErrorInvalid(cause)
			}
			state_refs[tr.ToState] = true
		}
	}

	var dead_states []string
	for name, referenced := range state_refs {
		if !referenced {
			dead_states = append(dead_states, name)
		}
	}
	if len(dead_states) > 0 {
		buf := bytes.NewBufferString("there are isolated states: ")
		for idx := range dead_states {
			buf.WriteString("\"")
			buf.WriteString(dead_states[idx])
			buf.WriteString("\", ")
		}
		return newFsmErrorInvalid(buf.String())
	}

	return nil
}

// Advance
// Event that makes state machine to transition to the next state
func (fsm *Fsm) Advance() (step HistoryItem, err *FsmError) {
	current := fsm.stack.Peek()
	currentName := current.state.Name

	// Process current FSM status
	switch {
	case fsm.Idle():
		if err = fsm.Validate(); err != nil {
			fsm.goFatal(err)
			return
		}
	case fsm.Completed():
		err = newFsmErrorWrongFlow("advance", "completed")
		return
	case fsm.Fatal():
		err = fsm.fatalError()
		return
	}

	// find target state by checking opened transitions
	var transition *Transition
	var opened_transitions int
	for idx := range current.state.Transitions {
		transition = &current.state.Transitions[idx]
		open, e := transition.Guard(&fsm.stack)
		if e != nil {
			err = newFsmErrorCallbackFailed("guard", e)
			fsm.goFatal(err)
			return
		}
		if open {
			opened_transitions++
		}
	}

	// * if there are some but no one fits, error
	// * if there are some and several fits, error
	var next *StateInfo
	switch opened_transitions {
	case 0:
		err = newFsmErrorRuntime("all transitions are closed", current)
	case 1:
		next = fsm.states[transition.ToState]
	default:
		err = newFsmErrorRuntime("more than 1 transitions are opened", current)
	}
	if next == nil {
		fsm.goFatal(err)
		return
	}

	// pop the stack until common parent is found for current and next states
	var depth_diff int
	ancestor, depth_diff := findCommonAncestor(current.state, next)
	if ancestor == nil {
		cause := fmt.Sprintf("\"%s\" and \"%s\" don't have a common parent", currentName, next.Name)
		err = newFsmErrorRuntime(cause, fsm.states)
		fsm.goFatal(err)
		return
	}
	switch {
	case depth_diff == -1:
		// meaning we go from parent state to substate,
		// no need to pop anything from the stack
	case depth_diff < -1:
		err = newFsmErrorRuntime("Trying to go deeper than 1 state at a time", current.state)
		fsm.goFatal(err)
		return
	case depth_diff >= 0:
		for idx := 0; idx < depth_diff+1; idx++ {
			if fsm.stack.Depth() <= FsmAutoStatesCount {
				break
			}
			fsm.stack.Pop()
		}
	}

	// Prepare new stack, log and execute transition action
	if fsm.stack.Push(next) == nil {
		err = newFsmErrorRuntime("pushing new state to the stack failed", next)
		fsm.goFatal(err)
		return
	}

	step = HistoryItem{
		currentName,
		next.Name,
		transition.Name,
	}
	fsm.history = append(fsm.history, step)

	if transition.Action != nil {
		if e := transition.Action(&fsm.stack); e != nil {
			err = newFsmErrorCallbackFailed("entry action", e)
			fsm.goFatal(err)
		}
	}

	// TODO: error detection: infinite transition loop
	return
}

// Run
// Executes whole FSM until it's completed or failed
func (fsm *Fsm) Run() (res interface{}, err *FsmError) {
	for !fsm.Completed() && !fsm.Fatal() && err == nil {
		_, err = fsm.Advance()
	}
	if fsm.Completed() {
		res, err = fsm.Result()
	}
	return
}

func (fsm *Fsm) goFatal(cause *FsmError) {
	if fsm.Fatal() {
		return
	}

	fsm.fatal = newFsmErrorInFatalState(cause,
		Dump(&fsm.stack),
		fsm.history,
	)
}

// Dump
// Print out an object in a user-friendly way
func (fsm *Fsm) dump(buf *bytes.Buffer, indent int) {
	buf.WriteString("finite state machine dump\n")

	buf.WriteString("> states:\n")
	if len(fsm.states) == 0 {
		buf.WriteString("\tno states\n")
	} else {
		for _, v := range fsm.states {
			v.dump(buf, 1)
		}
	}

	buf.WriteString("> status:\n")
	buf.WriteString(fmt.Sprintf("\t%s: %v\n", "idle", fsm.Idle()))
	buf.WriteString(fmt.Sprintf("\t%s: %v\n", "running", fsm.Running()))
	buf.WriteString(fmt.Sprintf("\t%s: %v\n", "completed", fsm.Completed()))
	buf.WriteString(fmt.Sprintf("\t%s: %v\n", "fatal", fsm.Fatal()))

	buf.WriteString("> history:\n")
	fsm.history.dumpImpl(buf, 1)

	buf.WriteString("> context stack:\n")
	fsm.stack.dump(buf, 1)

	buf.WriteString("> status info\n")
	buf.WriteString("\t")
	switch {
	case fsm.Running():
		buf.WriteString(fmt.Sprintf("FSM is running, %d transitions made\n", len(fsm.history)))
	case fsm.Fatal():
		buf.WriteString(fmt.Sprintf("FSM is fatal: %s\n", fsm.fatal))
	case fsm.Completed():
		res, err := fsm.Result()
		buf.WriteString(fmt.Sprintf("FSM is completed, result is: %v, error is: %s\n", res, err))
	}
}
