package simple_fsm

import (
	"testing"
)

func TestNewFsm(t *testing.T) {
	fsm := NewFsm()
	if fsm.start == nil {
		t.Log("Starting state is nil")
		t.FailNow()
	}
	if fsm.start.Name != "global" || !fsm.start.Final() {
		t.Logf("Starting state configured differently from expected: %#v", fsm.start)
		t.FailNow()
	}
	if err := fsm.Validate(); err != nil && err.Kind() != ErrStateIsInvalid {
		t.Logf("Validation failed with unexpected error: %s", err.Error())
		t.FailNow()
	}
}

func TestFsmReset(t *testing.T) {
	fsm := NewFsm()
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "2", nil)),
		NewState("2", nil),
	)

	fsm.Advance()
	fsm.Reset()

	if !fsm.Idle() || len(fsm.History()) > 0 {
		t.Logf("FSM was not properly reset")
		t.FailNow()
	}
}

func TestFsmFlow(t *testing.T) {
	statuses := []func(*Fsm) bool{
		func(fsm *Fsm) bool { return fsm.Idle() },
		func(fsm *Fsm) bool { return fsm.Running() },
		func(fsm *Fsm) bool { return fsm.Completed() },
		func(fsm *Fsm) bool { return fsm.Fatal() },
	}
	check := func(fsm *Fsm, sts []func(*Fsm) bool, expected []bool) bool {
		for i, flag := range expected {
			if flag != sts[i](fsm) {
				return false
			}
		}
		return true
	}

	fsm := NewFsm()
	succ := func(ctx ContextOperator) error { ctx.PutResult(true); return nil }
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "2", succ)),
		NewState("2", nil),
	)

	if !check(fsm, statuses, []bool{true, false, false, false}) {
		t.Log("FSM should be in Idle state")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm.Advance()
	if !check(fsm, statuses, []bool{false, true, false, false}) {
		t.Log("FSM should be in Running state")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm.Advance()
	if !check(fsm, statuses, []bool{false, false, true, false}) {
		t.Log("FSM should be in Completed state")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm.Advance()
	if !check(fsm, statuses, []bool{false, false, true, false}) {
		t.Log("FSM should still be in Completed state")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm = NewFsm()
	fail := func(ctx ContextOperator) error { return newFsmErrorRuntime("fail", nil) }
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "2", fail)),
		NewState("2", nil),
	)
	fsm.Advance()
	fsm.Advance()

	if !check(fsm, statuses, []bool{false, false, false, true}) {
		t.Log("FSM should be in Fatal state")
		t.Log(Dump(fsm))
		t.FailNow()
	}
	if !check(fsm, statuses, []bool{false, false, false, true}) {
		t.Log("FSM should still be in Fatal state")
		t.Log(Dump(fsm))
		t.FailNow()
	}
}

func TestFsmAddStatePositive(t *testing.T) {
	fsm := NewFsm()
	first := NewState("first", NewTransitionAlways("1-2", "second", nil))
	if err := fsm.AddStartState(first, nil); err != nil {
		t.Logf("AddState failed: %s", err.Error())
		t.FailNow()
	}
	if err := fsm.Validate(); err == nil {
		t.Logf("Validation should fail because of unknown transition")
		t.FailNow()
	}

	second := NewState("second", nil)
	if err := fsm.AddState(second, nil); err != nil {
		t.Logf("AddState failed: %s", err.Error())
		t.FailNow()
	}
	if err := fsm.Validate(); err != nil {
		t.Logf("Unexpected validation error: %s", err.Error())
		t.Logf("\n%s\n", Dump(fsm))
		t.FailNow()
	}
}

func TestFsmAddStatesPositive(t *testing.T) {
	fsm := NewFsm()
	err := fsm.AddStates(
		nil,
		NewState("1", NewTransitionAlways("1-2", "2", nil)),
		NewState("2", NewTransitionAlways("2-3", "3", nil)),
		NewState("3", nil),
	)
	if err != nil {
		t.Logf("Unexpected states addition error: %s", err.Error())
		t.FailNow()
	}
	if err = fsm.Validate(); err != nil {
		t.Logf("Unexpected states addition error: %s", err.Error())
		t.FailNow()
	}
}

func TestFsmAddStateNegative(t *testing.T) {
	fsm := NewFsm()

	if err := fsm.AddState(nil, nil); err == nil || err.Kind() != ErrStateIsInvalid {
		t.Logf("Should detect an error (state is nil): %#v", err)
		t.FailNow()
	}

	fsm.AddStartState(NewState("start", NewTransitionAlways("go", "next", nil)), nil)
	if err := fsm.AddStartState(NewState("dup", nil), nil); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Logf("Should detect an error (duplicate start): %#v", err)
		t.FailNow()
	}

	fsm.AddState(NewState("next", NewTransitionAlways("last", "last", nil)), nil)
	if err := fsm.AddState(NewState("next", nil), nil); err == nil || err.Kind() != ErrStateAlreadyExists {
		t.Logf("Should detect an error (already exists): %#v", err)
		t.FailNow()
	}

	if err := fsm.AddState(NewState("last", nil), NewState("notadded", nil)); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Logf("Should detect an error (parent not added): %#v", err)
		t.FailNow()
	}

	fsm.AddState(NewState("last", nil), nil)
	fsm.Advance()
	if err := fsm.AddState(NewState("run", nil), nil); err == nil || err.Kind() != ErrFsmWrongFlow {
		t.Logf("Should detect an error (wrong flow): %#v", err)
		t.FailNow()
	}

	fsm.Reset()
	sub1 := NewState("sub1", NewTransitionAlways("sub1-last", "last", nil))
	fsm.AddState(sub1, nil)
	if err := fsm.AddStartState(NewState("sub2", nil), sub1); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Logf("Should detect an error (parent->sub should have 1 transition): %#v", err)
		t.FailNow()
	}
}

func TestFsmValidate(t *testing.T) {
	fsm := NewFsm()
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "2", nil)),
		NewState("2", nil),
	)
	if err := fsm.Validate(); err != nil {
		t.Log("FSM should be valid")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm = NewFsm()
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "3", nil)),
		NewState("2", nil),
	)
	if err := fsm.Validate(); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Log("FSM should be invalid (unknown transition destination)")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm = NewFsm()
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "3", nil)),
		NewState("2", nil),
	)
	fsm.addStateImpl(NewState("3", nil), nil, false, false)
	if err := fsm.Validate(); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Log("FSM should be invalid (transition between unrelated states)")
		t.Log(err.Error())
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm = NewFsm()
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "3", nil)),
		NewState("2", nil),
		NewState("3", nil),
	)
	if err := fsm.Validate(); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Log("FSM should be invalid (isolated states)")
		t.Log(Dump(fsm))
		t.FailNow()
	}
}

func TestFsmAdvanceFlowError(t *testing.T) {
	fsm := NewFsm()
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "3", nil)),
		NewState("2", nil),
	)
	if _, err := fsm.Advance(); err == nil {
		t.Log("FSM should be invalid (failed validation, unknown destination)")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm = NewFsm()
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "2", nil)),
		NewState("2", nil),
	)
	fsm.Run()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmWrongFlow {
		t.Log("FSM should be invalid (flow error, can't advance after completion)")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm = NewFsm()
	fail := func(ctx ContextOperator) error { return newFsmErrorRuntime("fail", nil) }
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "2", fail)),
		NewState("2", nil),
	)
	fsm.Run()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmInFatalState {
		t.Log("FSM should be invalid (fatal error)")
		t.Log(Dump(fsm))
		t.FailNow()
	}
}

func TestFsmAdvanceGuardError(t *testing.T) {
	fsm := NewFsm()
	guard := func(ctx ContextAccessor) (bool, error) { return false, newFsmErrorRuntime("fail", nil) }
	fsm.AddStates(nil,
		NewState("1", []Transition{NewTransition("1-2", "2", guard, nil)}),
		NewState("2", nil),
	)
	fsm.Advance()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmCallbackFailed {
		t.Log("FSM should be invalid (guard callback failed)")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm = NewFsm()
	guard = func(ctx ContextAccessor) (bool, error) { return false, nil }
	fsm.AddStates(nil,
		NewState("1", []Transition{NewTransition("1-2", "2", guard, nil)}),
		NewState("2", nil),
	)
	fsm.Advance()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmRuntime {
		t.Log("FSM should be invalid (guard callback failed)")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm = NewFsm()
	guard = func(ctx ContextAccessor) (bool, error) { return true, nil }
	guard_list := []Transition{
		NewTransition("one", "2", guard, nil),
		NewTransition("two", "2", guard, nil),
	}
	fsm.AddStates(nil,
		NewState("1", guard_list),
		NewState("2", nil),
	)
	fsm.Advance()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmRuntime {
		t.Log("FSM should be invalid (>1 opened guards)")
		t.Log(Dump(fsm))
		t.FailNow()
	}
}

func TestFsmAdvanceTransitionError(t *testing.T) {
	fsm := NewFsm()
	s1, s2, s3 := NewState("1", nil), NewState("2", nil), NewState("3", NewTransitionAlways("3-11", "s11", nil))
	fsm.AddStartState(s1, nil)
	fsm.AddStartState(s2, s1)
	fsm.AddStartState(s3, s2)
	tr_list := []Transition{
		NewTransition("11-1", "1", func(ctx ContextAccessor) (bool, error) { return false, nil }, nil),
		NewTransition("11-3", "3", func(ctx ContextAccessor) (bool, error) { return true, nil }, nil),
	}
	fsm.AddState(NewState("s11", tr_list), nil)

	fsm.Advance()
	fsm.Advance()
	fsm.Advance()
	fsm.Advance()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmRuntime {
		t.Log("FSM should be invalid (can't go deeper that 1 state per transition)")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm = NewFsm()
	fail := func(ctx ContextOperator) error { return newFsmErrorRuntime("fail", nil) }
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "2", fail)),
		NewState("2", nil),
	)
	fsm.Advance()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmCallbackFailed {
		t.Log("FSM should be invalid (entry callback failed)")
		t.Log(Dump(fsm))
		t.FailNow()
	}
}

func TestFsmRunResult(t *testing.T) {
	fsm := NewFsm()
	succ := func(ctx ContextOperator) error { ctx.PutResult(true); return nil }
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "2", succ)),
		NewState("2", nil),
	)
	raw, err := fsm.Run()
	if res, ok := raw.(bool); err != nil || !res || !ok {
		t.Log("FSM should complete succesfully")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm = NewFsm()
	fail := func(ctx ContextOperator) error { return newFsmErrorRuntime("fail", nil) }
	fsm.AddStates(nil,
		NewState("1", NewTransitionAlways("1-2", "2", fail)),
		NewState("2", nil),
	)
	if raw, err := fsm.Run(); raw != nil || err == nil {
		t.Log("FSM should fail")
		t.Log(Dump(fsm))
		t.FailNow()
	}
}
