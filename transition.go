package simple_fsm

import (
	"bytes"
)

// GuardFn
// Function serving as a transition guard for FSM states
type GuardFn func(ctx ContextAccessor) (open bool, err error)

// ActionFn
// Function describing an action done on state transition
type ActionFn func(ctx ContextOperator) error

// Transition
// Describes transition to a state, guard included
type Transition struct {
	Name    string
	ToState string
	Guard   GuardFn
	Action  ActionFn
}

// NewTransition
// Creates new transition instance
func NewTransition(name string, to string, cond GuardFn, action ActionFn) Transition {
	return Transition{name, to, cond, action}
}

// NewTransitionAlways
// Creates transitions slice with single, unconditional transition
func NewTransitionAlways(name string, to string, action ActionFn) []Transition {
	always := func(ContextAccessor) (bool, error) { return true, nil }
	return []Transition{Transition{name, to, always, action}}
}

// Validate
// Checks if given transition is well-formed and not self-contradictory
func (tr *Transition) Validate() (err *FsmError) {
	switch {
	case tr.Name == "":
		err = newFsmErrorTransitionIsInvalid(tr, "transition should be named")
	case tr.ToState == "":
		err = newFsmErrorTransitionIsInvalid(tr, "transition should have proper destination")
	case tr.Guard == nil:
		err = newFsmErrorTransitionIsInvalid(tr, "condition has to be present")
	}
	return
}

// Dump
// Print out an object in a user-friendly way
func (tr *Transition) Dump(buf *bytes.Buffer) {
	buf.WriteString("name: \"")
	buf.WriteString(tr.Name)
	buf.WriteString("\", to: \"")
	buf.WriteString(tr.ToState)
	buf.WriteString("\", ")
	if tr.Guard != nil {
		buf.WriteString("has guard, ")
	} else {
		buf.WriteString("no guard, ")
	}
	if tr.Action != nil {
		buf.WriteString("has action")
	} else {
		buf.WriteString("no action")
	}

}
