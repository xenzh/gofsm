package simple_fsm

import (
	"testing"
)

func TestNewStructure(t *testing.T) {
	fstr := NewStructure()
	if fstr.start == nil {
		t.Log("Starting state is nil")
		t.FailNow()
	}
	if fstr.start.Name != "global" || !fstr.start.Final() {
		t.Logf("Starting state configured differently from expected: %#v", fstr.start)
		t.FailNow()
	}
	if err := fstr.Validate(); err != nil && err.Kind() != ErrStateIsInvalid {
		t.Logf("Validation failed with unexpected error: %s", err.Error())
		t.FailNow()
	}
}

func TestStructureAddStatePositive(t *testing.T) {
	fstr := NewStructure()
	first := NewState("first", NewTransitionAlways("1-2", "second", nil))
	if err := fstr.AddStartState(first, nil); err != nil {
		t.Logf("AddState failed: %s", err.Error())
		t.FailNow()
	}
	if err := fstr.Validate(); err == nil {
		t.Logf("Validation should fail because of unknown transition")
		t.FailNow()
	}

	second := NewState("second", nil)
	if err := fstr.AddState(second, nil); err != nil {
		t.Logf("AddState failed: %s", err.Error())
		t.FailNow()
	}
	if err := fstr.Validate(); err != nil {
		t.Logf("Unexpected validation error: %s", err.Error())
		t.Logf("\n%s\n", Dump(fstr))
		t.FailNow()
	}
}

func TestStructureAddStatesPositive(t *testing.T) {
	fstr := NewStructure()
	err := fstr.AddStates(
		nil,
		NewState("1", NewTransitionAlways("1-2", "2", nil)),
		NewState("2", NewTransitionAlways("2-3", "3", nil)),
		NewState("3", nil),
	)
	if err != nil {
		t.Logf("Unexpected states addition error: %s", err.Error())
		t.FailNow()
	}
	if err = fstr.Validate(); err != nil {
		t.Logf("Unexpected states addition error: %s", err.Error())
		t.FailNow()
	}
}

func TestStructureAddStateNegative(t *testing.T) {
	fstr := NewStructure()

	if err := fstr.AddState(nil, nil); err == nil || err.Kind() != ErrStateIsInvalid {
		t.Logf("Should detect an error (state is nil): %#v", err)
		t.FailNow()
	}

	fstr.AddStartState(NewState("start", NewTransitionAlways("go", "next", nil)), nil)
	if err := fstr.AddStartState(NewState("dup", nil), nil); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Logf("Should detect an error (duplicate start): %#v", err)
		t.FailNow()
	}

	fstr.AddState(NewState("next", NewTransitionAlways("last", "last", nil)), nil)
	if err := fstr.AddState(NewState("next", nil), nil); err == nil || err.Kind() != ErrStateAlreadyExists {
		t.Logf("Should detect an error (already exists): %#v", err)
		t.FailNow()
	}

	if err := fstr.AddState(NewState("last", nil), NewState("notadded", nil)); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Logf("Should detect an error (parent not added): %#v", err)
		t.FailNow()
	}

	sub1 := NewState("sub1", NewTransitionAlways("sub1-last", "last", nil))
	fstr.AddState(sub1, nil)
	if err := fstr.AddStartState(NewState("sub2", nil), sub1); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Logf("Should detect an error (parent->sub should have 1 transition): %#v", err)
		t.FailNow()
	}
}

func TestStructureValidate(t *testing.T) {
	fstr := NewStructure()
	fstr.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "2", nil)),
		NewState("2", nil),
	)
	if err := fstr.Validate(); err != nil {
		t.Log("FSM should be valid")
		t.Log(Dump(fstr))
		t.FailNow()
	}

	fstr = NewStructure()
	fstr.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "3", nil)),
		NewState("2", nil),
	)
	if err := fstr.Validate(); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Log("FSM should be invalid (unknown transition destination)")
		t.Log(Dump(fstr))
		t.FailNow()
	}

	fstr = NewStructure()
	fstr.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "3", nil)),
		NewState("2", nil),
	)
	fstr.addStateImpl(NewState("3", nil), nil, false, false)
	if err := fstr.Validate(); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Log("FSM should be invalid (transition between unrelated states)")
		t.Log(err.Error())
		t.Log(Dump(fstr))
		t.FailNow()
	}

	fstr = NewStructure()
	fstr.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "3", nil)),
		NewState("2", nil),
		NewState("3", nil),
	)
	if err := fstr.Validate(); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Log("FSM should be invalid (isolated states)")
		t.Log(Dump(fstr))
		t.FailNow()
	}
}
