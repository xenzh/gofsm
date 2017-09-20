package simple_fsm

import (
	"bytes"
	"fmt"
)

// Structure
// Holds static finite state machive information like states and transitions
type Structure struct {
	states map[string]*StateInfo
	start  *StateInfo
}

// NewStructure
// Constructs empty Fsm structure
func NewStructure() *Structure {
	fstr := Structure{
		states: make(map[string]*StateInfo),
		start:  NewState(FsmGlobalStateName, nil),
	}
	fstr.addStateImpl(fstr.start, nil, true, false)

	return &fstr
}

// MakeStructure
// Constructs FSM structure and adds a number of (sub)states
func MakeStructure(parent *StateInfo, start *StateInfo, states ...*StateInfo) *Structure {
	fstr := NewStructure()
	fstr.AddStates(parent, start, states...)
	return fstr
}

func (fstr *Structure) Empty() bool {
	return len(fstr.states) <= FsmAutoStatesCount
}

// AddStartState
// Validates and adds a start (sub)state to the state machine
func (fstr *Structure) AddStartState(state *StateInfo, parent *StateInfo) (err *FsmError) {
	return fstr.addStateImpl(state, parent, true, true)
}

// AddStartState
// Validates and adds an intermediate (sub)state to the state machine
func (fstr *Structure) AddState(state *StateInfo, parent *StateInfo) (err *FsmError) {
	return fstr.addStateImpl(state, parent, false, true)
}

// AddStates
// Allows to add a bunch of (sub)states (including starting one) to the state machine
func (fstr *Structure) AddStates(parent *StateInfo, start *StateInfo, states ...*StateInfo) (err *FsmError) {
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
func (fstr *Structure) addStateImpl(state *StateInfo, parent *StateInfo, start bool, autoAdopt bool) (err *FsmError) {
	switch {
	case state == nil:
		return newFsmErrorStateIsInvalid(state, "state is nil")
	case fstr.start == nil:
		return newFsmErrorInvalid("global state is not defined")
	case start && parent == nil && !fstr.start.Final():
		cause := fmt.Sprintf("start state is already set to \"%s\"", fstr.start.Transitions[0].Name)
		return newFsmErrorInvalid(cause)
	case start && autoAdopt && parent != nil && len(parent.Transitions) > 0:
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

	if !autoAdopt {
		return
	}

	if parent == nil {
		err = fstr.start.addSubState(state, start)
	} else {
		if _, present := fstr.states[parent.Name]; !present {
			err = newFsmErrorInvalid("Parent state was not found (forgot to add?)")
		} else {
			err = parent.addSubState(state, start)
		}
	}

	return
}

// appendStates
// Appends an external state map to the structure.
// start is expected to defined FSM's main entry point.
// External states are assumed to have a consistent hierarchy.
// Duplicate state names are treated as an error.
func (fstr *Structure) appendStates(start *StateInfo, additional map[string]*StateInfo) *FsmError {
	if start != nil {
		if err := fstr.AddStartState(start, nil); err != nil {
			return err
		}
	}

	for k, v := range additional {
		if _, found := fstr.states[k]; found {
			return newFsmErrorStateIsInvalid(v, "Can't add a duplicate state")
		}
		if v.Parent == nil {
			fstr.start.addSubState(v, false)
		}
		fstr.states[k] = v
	}
	return nil
}

// Validate
// Checks if FSM structure is consistent:
// * no transitions to unknown
// * no dead states
func (fstr *Structure) Validate() (err *FsmError) {
	stateRefs := make(map[string]bool)
	for k, _ := range fstr.states {
		stateRefs[k] = false
	}
	stateRefs[fstr.start.Name] = true

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
			stateRefs[tr.ToState] = true
		}
	}

	var deadStates []string
	for name, referenced := range stateRefs {
		if !referenced {
			deadStates = append(deadStates, name)
		}
	}

	if len(deadStates) > 0 {
		buf := bytes.NewBufferString("there are isolated states: ")
		for idx := range deadStates {
			buf.WriteString("\"")
			buf.WriteString(deadStates[idx])
			buf.WriteString("\", ")
		}
		return newFsmErrorInvalid(buf.String())
	}

	return nil
}

func (fstr *Structure) dump(buf *bytes.Buffer, indent int) {
	if len(fstr.states) == 0 {
		buf.WriteString("\tno states\n")
	} else {
		for _, v := range fstr.states {
			v.dump(buf, indent)
		}
	}
}
