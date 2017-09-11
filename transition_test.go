package simple_fsm

import (
	"testing"
)

func TestNewActionParams(t *testing.T) {
	pa := NewAction(nil)
	if pa.Fn != nil || pa.Params != nil {
		t.Log("Both function and params map are expected to be nil")
		t.FailNow()
	}

	pa.Param("key", "value")
	if pa.Params == nil || len(pa.Params) != 1 {
		t.Log("Params are expected to be created/appended with exactly 1 value")
		t.FailNow()
	}
}

func TestPackagedActionDo(t *testing.T) {
	action := NewAction(func(ctx ContextOperator) error {
		val, err := ctx.Str("getme")
		if err != nil {
			return err
		}
		if val != "42" {
			return newFsmErrorRuntime("Value is different from expected!", val)
		}
		return nil
	})

	ctx := newContextStack()
	ctx.Push(NewState("single", NewTransitionAlways("1-2", "2", action)))

	err := action.Do(&ctx)
	if err == nil {
		t.Log("Expected this action to fail")
		t.FailNow()
	}
	fsmErr, ok := err.(*FsmError)
	if !ok || fsmErr.Kind() != ErrCtxKeyNotFound {
		t.Log("Expected FsmError (key not found)")
		t.FailNow()
	}

	action.Param("getme", "not 42")

	err = action.Do(&ctx)
	if err == nil {
		t.Log("Expected this action to fail")
		t.FailNow()
	}
	fsmErr, ok = err.(*FsmError)
	if !ok || fsmErr.Kind() != ErrFsmRuntime {
		t.Log("Expected FsmError (value is different from expected)")
		t.FailNow()
	}

	action.Param("getme", "42")

	err = action.Do(&ctx)
	if err != nil {
		t.Logf("Expected this action to pass, error: %s", err)
		t.FailNow()
	}
}

func TestTransitionValidate(t *testing.T) {
	cond := func(ctx ContextAccessor) (bool, error) { return true, nil }

	tr := NewTransition("name", "to", cond, nil)
	if tr.Validate() != nil {
		t.Log("This transition should be valid")
		t.FailNow()
	}

	tr = NewTransitionAlways("name", "to", nil)[0]
	if tr.Validate() != nil {
		t.Log("This transition should be valid")
		t.FailNow()
	}

	tr = NewTransition("", "to", cond, nil)
	if tr.Validate() == nil {
		t.Log("Name should be mandatory")
		t.FailNow()
	}

	tr = NewTransition("name", "", cond, nil)
	if tr.Validate() == nil {
		t.Log("Destination state name should be mandatory")
		t.FailNow()
	}

	tr = NewTransition("name", "", nil, nil)
	if tr.Validate() == nil {
		t.Log("Condition predicate should be mandatory")
		t.FailNow()
	}
}
