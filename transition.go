package simple_fsm

import (
	"bytes"
	"fmt"
)

// GuardFn
// Function serving as a transition guard for FSM states
type GuardFn func(ctx ContextAccessor) (open bool, err error)

// ActionFn
// Function describing an action done on state transition
type ActionFn func(ctx ContextOperator) error

// PackagedAction
// Encapsulates an action functor and it's input parameters
// that are put to the context right before action execution
type PackagedAction struct {
	Fn     ActionFn
	Params map[string]interface{}
}

// NewAction
// Constructs new action based on a functor
func NewAction(fn ActionFn) *PackagedAction {
	return &PackagedAction{fn, nil}
}

// Param
// Adds an input parameter to the action
// Returns action pointer, so Param() calls can be chained
func (pa *PackagedAction) Param(key string, value interface{}) *PackagedAction {
	if pa.Params == nil {
		pa.Params = make(map[string]interface{})
	}
	pa.Params[key] = value
	return pa
}

// Do
// Puts parameters to the context and executes the action
func (pa *PackagedAction) Do(ctx ContextOperator) error {
	if pa.Params != nil {
		for k, v := range pa.Params {
			if err := ctx.Put(k, v); err != nil {
				return err
			}
		}
	}
	if pa.Fn == nil {
		return newFsmErrorRuntime("PackagedAction doesn't specify a functor", pa)
	}
	return pa.Fn(ctx)
}

// Validate
// Checks if given action is well-formed and not self-contradicting
func (pa *PackagedAction) Validate() (err *FsmError) {
	if pa.Fn == nil {
		err = newFsmErrorInvalid("PackagedAction should specify a functor")
	}
	return
}

// Transition
// Describes transition to a state, guard included
type Transition struct {
	Name    string
	ToState string
	Guard   GuardFn
	Action  *PackagedAction
}

// NewTransition
// Creates new transition instance
func NewTransition(name string, to string, cond GuardFn, action *PackagedAction) Transition {
	return Transition{name, to, cond, action}
}

// NewTransitionAlways
// Creates transitions slice with single, unconditional transition
func NewTransitionAlways(name string, to string, action *PackagedAction) []Transition {
	always := func(ContextAccessor) (bool, error) { return true, nil }
	return []Transition{Transition{name, to, always, action}}
}

// Validate
// Checks if given transition is well-formed and not self-contradicting
func (tr *Transition) Validate() (err *FsmError) {
	switch {
	case tr.Name == "":
		err = newFsmErrorTransitionIsInvalid(tr, "transition should be named")
	case tr.ToState == "":
		err = newFsmErrorTransitionIsInvalid(tr, "transition should have proper destination")
	case tr.Guard == nil:
		err = newFsmErrorTransitionIsInvalid(tr, "condition has to be present")
	}

	if err == nil && tr.Action != nil {
		err = tr.Action.Validate()
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
		if tr.Action.Params == nil || len(tr.Action.Params) == 0 {
			buf.WriteString("has action (no params)")
		} else if len(tr.Action.Params) > 0 {
			buf.WriteString("has action ( ")
			for k, v := range tr.Action.Params {
				buf.WriteString(k)
				buf.WriteString("=")
				buf.WriteString(fmt.Sprintf("%v", v))
				buf.WriteString(" ")
			}
			buf.WriteString(")")
		}

	} else {
		buf.WriteString("no action")
	}

}
