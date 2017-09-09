package simple_fsm

import (
	"fmt"
	"reflect"
)

// FsmErrorKind
// Enum-like type describing internal FSM error causes
type FsmErrorKind int

const (
	ErrCtxKeyNotFound FsmErrorKind = iota
	ErrCtxInvalidType
	ErrStateAlreadyExists
	ErrStateIsInvalid
	ErrFsmLoading
	ErrFsmWrongFlow
	ErrFsmIsInvalid
	ErrFsmRuntime
	ErrFsmCallbackFailed
	ErrFsmInFatalState
)

// FsmError
// Type containing information about internal FSM error
type FsmError struct {
	kind        FsmErrorKind
	description string
}

// Kind
// Returns error cause
func (e *FsmError) Kind() FsmErrorKind {
	return e.kind
}

// Error
// Implementation fo standard error interface
// Returns a string with combined error description
func (e *FsmError) Error() string {
	switch e.kind {
	case ErrCtxKeyNotFound:
		return fmt.Sprintf("No such key in the context: \"%s\"", e.description)
	case ErrCtxInvalidType:
		return fmt.Sprintf("Value has a type different from requested, %s", e.description)
	case ErrStateAlreadyExists:
		return fmt.Sprintf("State with the name \"%s\" is already added", e.description)
	case ErrStateIsInvalid:
		return fmt.Sprintf("This state is not valid: %s", e.description)
	case ErrFsmLoading:
		return fmt.Sprintf("Structure loading error: %s", e.description)
	case ErrFsmWrongFlow:
		return fmt.Sprintf("You're using it wrong: %s", e.description)
	case ErrFsmIsInvalid:
		return fmt.Sprintf("FSM internal structure is invalid: %s", e.description)
	case ErrFsmRuntime:
		return fmt.Sprintf("FSM encountered runtime error: %s", e.description)
	case ErrFsmCallbackFailed:
		return fmt.Sprintf("User-defined callback returned an error: %s", e.description)
	case ErrFsmInFatalState:
		return fmt.Sprintf("FSM stopped due to fatal error: %s", e.description)
	default:
		return "Unknown error"
	}
}

// newCtxErrorKeyNotFound
// Constructs "key was not found in the context" error
func newCtxErrorKeyNotFound(key string) *FsmError {
	return &FsmError{kind: ErrCtxKeyNotFound, description: key}
}

// newCtxErrorInvalidType
// Constructs "value type is different from requested" error
func newCtxErrorInvalidType(requested interface{}, actual interface{}) *FsmError {
	types := fmt.Sprintf("requested: %s, actual: %s", reflect.TypeOf(requested), reflect.TypeOf(actual))
	return &FsmError{kind: ErrCtxInvalidType, description: types}
}

// newFsmErrorStateAlreadyExists
// Constructs "state is already present in FSM" error
func newFsmErrorStateAlreadyExists(name string) *FsmError {
	return &FsmError{
		kind:        ErrStateAlreadyExists,
		description: name,
	}
}

// newFsmErrorStateIsInvalid
// Constructs "State object is invalid and cannot be used in FSM" error
func newFsmErrorStateIsInvalid(state *StateInfo, cause string) *FsmError {
	return &FsmError{
		kind:        ErrStateIsInvalid,
		description: fmt.Sprintf("cause: \"%s\"; object dump: %#v", cause, state),
	}
}

func newFsmErrorLoading(cause string) *FsmError {
	return &FsmError{
		kind:        ErrFsmLoading,
		description: cause,
	}
}

// newFsmErrorStateIsInvalid
// Constructs "State object is invalid and cannot be used in FSM" error
func newFsmErrorTransitionIsInvalid(transition *Transition, cause string) *FsmError {
	return &FsmError{
		kind:        ErrStateIsInvalid,
		description: fmt.Sprintf("cause: \"%s\"; object dump: %#v", cause, transition),
	}
}

// newFsmErrorWrongFlow
// Constructs "this sction can't be executed in this FSM state" error
func newFsmErrorWrongFlow(action string, fsm_state string) *FsmError {
	return &FsmError{
		kind:        ErrFsmWrongFlow,
		description: fmt.Sprintf("Can't %s while FSM is %s", action, fsm_state),
	}
}

// newFsmErrorInvalid
// Constructs "FSM internal structure is invalid/not conststent" error
func newFsmErrorInvalid(cause string) *FsmError {
	return &FsmError{
		kind:        ErrFsmIsInvalid,
		description: cause,
	}
}

// newFsmErrorRuntime
// Constructs "FSM encountered runtime problem" error
func newFsmErrorRuntime(cause string, obj interface{}) *FsmError {
	return &FsmError{
		kind:        ErrFsmRuntime,
		description: fmt.Sprintf("Cause: %s, object: %#v", cause, obj),
	}
}

// newFsmErrorCallbackFailed
// Constructs "FSM callback (action or condition) failed" error
func newFsmErrorCallbackFailed(who string, e error) *FsmError {
	return &FsmError{
		kind:        ErrFsmCallbackFailed,
		description: fmt.Sprintf("%s, \"%s\"", who, e.Error()),
	}
}

// newFsmErrorInFatalState
// Constructs "FSM is stopped due to fatal error" error
func newFsmErrorInFatalState(cause *FsmError, stackDump string, history History) *FsmError {
	return &FsmError{
		kind: ErrFsmInFatalState,
		description: fmt.Sprintf(
			"\ncause:\n%s\n\nstack:\n\n%s\nhistory:\n%s",
			cause.Error(),
			stackDump,
			history.Dump(),
		),
	}
}
