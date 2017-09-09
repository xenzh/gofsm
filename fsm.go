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
	structure *Structure
	stack     ContextStack
	history   History
	fatal     *FsmError
}

// NewFsm
// Constructs new state machine, initializes auto states and structures
func NewFsm(structure *Structure) *Fsm {
	fsm := Fsm{
		structure: structure,
		stack:     newContextStack(),
		history:   make([]HistoryItem, 0, FsmDefaultHistoryCapacity),
		fatal:     nil,
	}
	fsm.initStackAutoStates()

	return &fsm
}

// Reset
// Resets FSM to state, ready for execution (initial)
// Progress/results from previous run is discarded
func (fsm *Fsm) Reset() {
	fsm.stack = newContextStack()
	fsm.initStackAutoStates()
	fsm.history = make([]HistoryItem, 0, FsmDefaultHistoryCapacity)
	fsm.fatal = nil
}

// initStackAutoStates
// Populates stack with automatic stuff (default global state, as of now)
func (fsm *Fsm) initStackAutoStates() {
	if !fsm.stack.Empty() {
		return
	}
	fsm.stack.Push(fsm.structure.start)
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

// Advance
// Event that makes state machine to transition to the next state
func (fsm *Fsm) Advance() (step HistoryItem, err *FsmError) {
	current := fsm.stack.Peek()
	currentName := current.state.Name

	// Process current FSM status
	switch {
	case fsm.Idle():
		if err = fsm.structure.Validate(); err != nil {
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
	var openedTransitionCount int
	for idx := range current.state.Transitions {
		currTransition := &current.state.Transitions[idx]
		open, e := currTransition.Guard(&fsm.stack)
		if e != nil {
			err = newFsmErrorCallbackFailed("guard", e)
			fsm.goFatal(err)
			return
		}
		if open {
			transition = currTransition
			openedTransitionCount++
		}
	}

	// * if there are some but no one fits, error
	// * if there are some and several fits, error
	var next *StateInfo
	switch openedTransitionCount {
	case 0:
		err = newFsmErrorRuntime("all transitions are closed", current)
	case 1:
		next = fsm.structure.states[transition.ToState]
	default:
		err = newFsmErrorRuntime("more than 1 transitions are opened", current)
	}
	if next == nil {
		fsm.goFatal(err)
		return
	}

	// pop the stack until common parent is found for current and next states
	var depthDiff int
	ancestor, depthDiff := findCommonAncestor(current.state, next)
	if ancestor == nil {
		cause := fmt.Sprintf("\"%s\" and \"%s\" don't have a common parent", currentName, next.Name)
		err = newFsmErrorRuntime(cause, fsm.structure.states)
		fsm.goFatal(err)
		return
	}
	switch {
	case depthDiff == -1:
		// meaning we go from parent state to substate,
		// no need to pop anything from the stack
	case depthDiff < -1:
		err = newFsmErrorRuntime("Trying to go deeper than 1 state at a time", current.state)
		fsm.goFatal(err)
		return
	case depthDiff >= 0:
		for idx := 0; idx < depthDiff+1; idx++ {
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
	fsm.structure.dump(buf, 1)

	buf.WriteString("> status:\n")
	buf.WriteString(fmt.Sprintf("\t%s: %v\n", "idle", fsm.Idle()))
	buf.WriteString(fmt.Sprintf("\t%s: %v\n", "running", fsm.Running()))
	buf.WriteString(fmt.Sprintf("\t%s: %v\n", "completed", fsm.Completed()))
	buf.WriteString(fmt.Sprintf("\t%s: %v\n", "fatal", fsm.Fatal()))

	buf.WriteString("> history:\n")
	fsm.history.dump(buf, 1)

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
