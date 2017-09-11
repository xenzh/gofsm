package simple_fsm

import (
	"testing"
)

func TestFsmReset(t *testing.T) {
	fsm := NewFsm(MakeStructure(nil,
		NewState("1", NewTransitionAlways("1-2", "2", nil)),
		NewState("2", nil),
	))

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

	succ := NewAction(func(ctx ContextOperator) error { ctx.PutResult(true); return nil })
	fsm := NewFsm(MakeStructure(nil,
		NewState("1", NewTransitionAlways("1-2", "2", succ)),
		NewState("2", nil),
	))

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

	fail := NewAction(func(ctx ContextOperator) error { return newFsmErrorRuntime("fail", nil) })
	fsm = NewFsm(MakeStructure(nil,
		NewState("1", NewTransitionAlways("1-2", "2", fail)),
		NewState("2", nil),
	))
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

func TestFsmAdvanceWrongFlow(t *testing.T) {
	fsm := NewFsm(MakeStructure(nil,
		NewState("1", NewTransitionAlways("1-2", "3", nil)),
		NewState("2", nil),
	))
	if _, err := fsm.Advance(); err == nil {
		t.Log("FSM should be invalid (failed validation, unknown destination)")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fsm = NewFsm(MakeStructure(nil,
		NewState("1", NewTransitionAlways("1-2", "2", nil)),
		NewState("2", nil),
	))
	fsm.Run()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmWrongFlow {
		t.Log("FSM should be invalid (flow error, can't advance after completion)")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fail := NewAction(func(ctx ContextOperator) error { return newFsmErrorRuntime("fail", nil) })
	fsm = NewFsm(MakeStructure(nil,
		NewState("1", NewTransitionAlways("1-2", "2", fail)),
		NewState("2", nil),
	))
	fsm.Run()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmInFatalState {
		t.Log("FSM should be invalid (fatal error)")
		t.Log(Dump(fsm))
		t.FailNow()
	}
}

func TestFsmAdvanceGuardError(t *testing.T) {
	guard := func(ctx ContextAccessor) (bool, error) { return false, newFsmErrorRuntime("fail", nil) }
	fsm := NewFsm(MakeStructure(nil,
		NewState("1", []Transition{NewTransition("1-2", "2", guard, nil)}),
		NewState("2", nil),
	))
	fsm.Advance()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmCallbackFailed {
		t.Log("FSM should be invalid (guard callback failed)")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	guard = func(ctx ContextAccessor) (bool, error) { return false, nil }
	fsm = NewFsm(MakeStructure(nil,
		NewState("1", []Transition{NewTransition("1-2", "2", guard, nil)}),
		NewState("2", nil),
	))
	fsm.Advance()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmRuntime {
		t.Log("FSM should be invalid (guard callback failed)")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	guard = func(ctx ContextAccessor) (bool, error) { return true, nil }
	guard_list := []Transition{
		NewTransition("one", "2", guard, nil),
		NewTransition("two", "2", guard, nil),
	}
	fsm = NewFsm(MakeStructure(nil,
		NewState("1", guard_list),
		NewState("2", nil),
	))
	fsm.Advance()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmRuntime {
		t.Log("FSM should be invalid (>1 opened guards)")
		t.Log(Dump(fsm))
		t.FailNow()
	}
}

func TestFsmAdvanceTransitionError(t *testing.T) {
	fstr := NewStructure()
	s1, s2, s3 := NewState("1", nil), NewState("2", nil), NewState("3", NewTransitionAlways("3-11", "s11", nil))
	fstr.AddStartState(s1, nil)
	fstr.AddStartState(s2, s1)
	fstr.AddStartState(s3, s2)
	tr_list := []Transition{
		NewTransition("11-1", "1", func(ctx ContextAccessor) (bool, error) { return false, nil }, nil),
		NewTransition("11-3", "3", func(ctx ContextAccessor) (bool, error) { return true, nil }, nil),
	}
	fstr.AddState(NewState("s11", tr_list), nil)

	fsm := NewFsm(fstr)

	fsm.Advance()
	fsm.Advance()
	fsm.Advance()
	fsm.Advance()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmRuntime {
		t.Log("FSM should be invalid (can't go deeper that 1 state per transition)")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fail := NewAction(func(ctx ContextOperator) error { return newFsmErrorRuntime("fail", nil) })
	fsm = NewFsm(MakeStructure(nil,
		NewState("1", NewTransitionAlways("1-2", "2", fail)),
		NewState("2", nil),
	))

	fsm.Advance()
	if _, err := fsm.Advance(); err == nil || err.Kind() != ErrFsmCallbackFailed {
		t.Log("FSM should be invalid (entry callback failed)")
		t.Log(Dump(fsm))
		t.FailNow()
	}
}

func TestFsmRunFewStatesResult(t *testing.T) {
	succ_action := NewAction(func(ctx ContextOperator) error { ctx.PutResult(true); return nil })
	should_be_executed := NewAction(func(ctx ContextOperator) error { ctx.PutResult(true); return nil })
	should_not_be_executed := NewAction(func(ctx ContextOperator) error { ctx.PutResult(false); return nil })

	open_guard := func(ctx ContextAccessor) (bool, error) { return true, nil }
	closed_guard := func(ctx ContextAccessor) (bool, error) { return false, nil }

	fsm := NewFsm(MakeStructure(nil,
		NewState("1", NewTransitionAlways("1-2", "2", succ_action)),
		NewState("2", []Transition{
			NewTransition("2-2.1 - open", "2.1 - open", open_guard, should_be_executed),
			NewTransition("2-2.2 - never get here", "2.2 - never get here", closed_guard, should_not_be_executed),
		}),
		NewState("2.1 - open", nil),
		NewState("2.2 - never get here", nil),
	))

	raw, err := fsm.Run()
	if err != nil {
		t.Logf("FSM espected to pass, error: %s", err)
		t.FailNow()
	}
	if raw.(bool) == false {
		t.Log("FSM should not follow closed transitions")
		t.Log(Dump(fsm))
		t.FailNow()
	}
}

func TestFsmRunResult(t *testing.T) {
	succ := NewAction(func(ctx ContextOperator) error { ctx.PutResult(true); return nil })
	fsm := NewFsm(MakeStructure(nil,
		NewState("1", NewTransitionAlways("1-2", "2", succ)),
		NewState("2", nil),
	))
	raw, err := fsm.Run()
	if res, ok := raw.(bool); err != nil || !res || !ok {
		t.Log("FSM should complete succesfully")
		t.Log(Dump(fsm))
		t.FailNow()
	}

	fail := NewAction(func(ctx ContextOperator) error { return newFsmErrorRuntime("fail", nil) })
	fsm = NewFsm(MakeStructure(nil,
		NewState("1", NewTransitionAlways("1-2", "2", fail)),
		NewState("2", nil),
	))
	if raw, err := fsm.Run(); raw != nil || err == nil {
		t.Log("FSM should fail")
		t.Log(Dump(fsm))
		t.FailNow()
	}
}
