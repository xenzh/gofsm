package simple_fsm

import (
	"encoding/json"
	"testing"
)

func TestJsonGuardUnmarshal(t *testing.T) {
	raw_json := `
    {
        "type": "context",
        "key": "hello",
        "value": "world"
    }`
	var jg jsonGuard
	if err := json.Unmarshal([]byte(raw_json), &jg); err != nil {
		t.Logf("Unmarshalling failed: %s", err.Error())
		t.FailNow()
	}
	if jg.Type != "context" || jg.Key != "hello" || jg.Value != "world" {
		t.Logf("Unmarshalled different from expected:\nexpected: %s\nactual:%v", raw_json, jg)
		t.FailNow()
	}
}

func TestJsonGuardFnAlways(t *testing.T) {
	jg := jsonGuard{"always", "", ""}
	guard, err := jg.GuardFn()
	if err != nil {
		t.Logf("Expected to succeed, error: %v", err)
		t.FailNow()
	}
	ctx := newContext()
	if ok, e := guard(&ctx); !ok || e != nil {
		t.Log("\"always\" transitions are expected to always succeed")
		t.FailNow()
	}
}

func TestJsonGuardFnContext(t *testing.T) {
	jg := jsonGuard{"context", "hello", "world"}
	guard, err := jg.GuardFn()
	if err != nil {
		t.Logf("Expected to succeed, error: %v", err)
		t.FailNow()
	}
	ctx := newContext()
	if ok, e := guard(&ctx); ok || e == nil {
		t.Logf("Expected to fail(%v)/be closed(%v)", e, ok)
		t.FailNow()
	}
	ctx.Put("hello", "someone else")
	if ok, e := guard(&ctx); ok || e != nil {
		t.Logf("Expected to pass(%v)/be closed(%v)", e, ok)
		t.FailNow()
	}
	ctx.Put("hello", "world")
	if ok, e := guard(&ctx); !ok || e != nil {
		t.Logf("Expected to pass(%v)/be opened(%v)", e, ok)
		t.FailNow()
	}
}

func TestJsonGuardFnContextIllFormed(t *testing.T) {
	jg := jsonGuard{"invalid", "", ""}
	if _, err := jg.GuardFn(); err == nil {
		t.Log("Expected to fail (unknown guard type)")
		t.FailNow()
	}

	jg = jsonGuard{"context", "", nil}
	if _, err := jg.GuardFn(); err == nil {
		t.Log("Expected to fail (no key/value specified)")
		t.FailNow()
	}
	jg = jsonGuard{"context", "key", nil}
	if _, err := jg.GuardFn(); err == nil {
		t.Log("Expected to fail (no key/value specified)")
		t.FailNow()
	}
}

func TestJsonTransitionUnmarshal(t *testing.T) {
	raw_json := `
    {
        "to": "2",
        "action": "hello",
        "guard": {
            "type": "always"
        }
    }`
	var jt jsonTransition
	if err := json.Unmarshal([]byte(raw_json), &jt); err != nil {
		t.Logf("Unmarshalling failed: %s", err.Error())
		t.FailNow()
	}
	if jt.ToState != "2" || jt.Action != "hello" || jt.Guard.Type != "always" {
		t.Logf("Unmarshalled different from expected:\nexpected: %s\nactual:%v", raw_json, jt)
		t.FailNow()
	}
}

func TestJsonTransitionFn(t *testing.T) {
	jt := jsonTransition{"2", jsonGuard{"always", "", nil}, "hello"}
	act := make(actionMap)
	if _, err := jt.Transition("1-2", act); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Log("Expected to fail (no action found)")
		t.FailNow()
	}
	act["hello"] = func(ctx ContextOperator) error { return nil }
	tr, err := jt.Transition("1-2", act)
	if err != nil {
		t.Logf("Expected to pass, error: %s", err.Error())
		t.FailNow()
	}
}
