package simple_fsm

import (
	"testing"
)

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
