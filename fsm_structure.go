package simple_fsm

import (
	"bytes"
	"fmt"
)

// FsmStructure
// Holds static finite state machive information like states and transitions
type FsmStructure struct {
	states map[string]*StateInfo
	start  *StateInfo
}

// NewFsmStructure
// Constructs empty Fsm structure
func NewFsmStructure() *FsmStructure {
	fstr := FsmStructure{
		states: make(map[string]*StateInfo),
		start:  NewState(FsmGlobalStateName, nil),
	}
	fstr.addStateImpl(fstr.start, nil, true, false)

	return &fstr
}

// MakeFsmStructure
// Constructs FSM structure and adds a number of (sub)states
func MakeFsmStructure(parent *StateInfo, start *StateInfo, states ...*StateInfo) *FsmStructure {
	fstr := NewFsmStructure()
	fstr.AddStates(parent, start, states...)
	return fstr
}

// AddStartState
// Validates and adds a start (sub)state to the state machine
func (fstr *FsmStructure) AddStartState(state *StateInfo, parent *StateInfo) (err *FsmError) {
	return fstr.addStateImpl(state, parent, true, true)
}

// AddStartState
// Validates and adds an intermediate (sub)state to the state machine
func (fstr *FsmStructure) AddState(state *StateInfo, parent *StateInfo) (err *FsmError) {
	return fstr.addStateImpl(state, parent, false, true)
}

// AddStates
// Allows to add a bunch of (sub)states (including starting one) to the state machine
func (fstr *FsmStructure) AddStates(parent *StateInfo, start *StateInfo, states ...*StateInfo) (err *FsmError) {
	if start != nil {
		if err = fstr.AddStartState(start, parent); err != nil {
			return
		}
	}
	for _, state := range states {
		if err = fstr.AddState(state, parent); err != nil {
			break
		}
	}
	return
}

// addStateImpl
// Adds a state to the state machine, validating it beforehand
func (fstr *FsmStructure) addStateImpl(state *StateInfo, parent *StateInfo, start bool, autoAdopt bool) (err *FsmError) {
	switch {
	case state == nil:
		return newFsmErrorStateIsInvalid(state, "state is nil")
	case fstr.start == nil:
		return newFsmErrorInvalid("global state is not defined")
	case start && parent == nil && !fstr.start.Final():
		cause := fmt.Sprintf("start state is already set to \"%s\"", fstr.start.Transitions[0].Name)
		return newFsmErrorInvalid(cause)
	case start && parent != nil && len(parent.Transitions) > 0:
		cause := "parent should not have transitions (transition to start sub state is added automatically)"
		return newFsmErrorInvalid(cause)
	}

	if err = state.Validate(); err != nil {
		return
	}
	if _, present := fstr.states[state.Name]; present {
		return newFsmErrorStateAlreadyExists(state.Name)
	}

	fstr.states[state.Name] = state
	if parent == nil {
		if autoAdopt {
			err = fstr.start.addSubState(state, start)
		}
	} else {
		if _, present := fstr.states[parent.Name]; !present {
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
func (fstr *FsmStructure) Validate() (err *FsmError) {
	state_refs := make(map[string]bool)
	for k, _ := range fstr.states {
		state_refs[k] = false
	}
	state_refs[fstr.start.Name] = true

	// TODO: 1 start substate can't belong to many parents

	for _, s := range fstr.states {
		if err := s.Validate(); err != nil {
			return err
		}
		for _, tr := range s.Transitions {
			if _, present := fstr.states[tr.ToState]; !present {
				cause := fmt.Sprintf(
					"transition \"%s\" of state \"%s\" has unknown destination \"%s\"",
					tr.Name,
					s.Name,
					tr.ToState,
				)
				return newFsmErrorInvalid(cause)
			}
			if ancestor, _ := findCommonAncestor(s, fstr.states[tr.ToState]); ancestor == nil {
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

func (fstr *FsmStructure) dump(buf *bytes.Buffer, indent int) {
	if len(fstr.states) == 0 {
		buf.WriteString("\tno states\n")
	} else {
		for _, v := range fstr.states {
			v.dump(buf, indent)
		}
	}
}
