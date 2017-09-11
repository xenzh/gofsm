package simple_fsm

import (
	"encoding/json"
	"testing"
)

func TestJsonGuardUnmarshal(t *testing.T) {
	rawJson := `
    {
        "type": "context",
        "key": "hello",
        "value": "world"
    }`
	var jg JsonGuard
	if err := json.Unmarshal([]byte(rawJson), &jg); err != nil {
		t.Logf("Unmarshalling failed: %s", err.Error())
		t.FailNow()
	}
	if jg.Type != "context" || jg.Key != "hello" || jg.Value != "world" {
		t.Logf("Unmarshalled different from expected:\nexpected: %s\nactual:%v", rawJson, jg)
		t.FailNow()
	}
}

func TestJsonGuardFnAlwaysExplicit(t *testing.T) {
	jg := JsonGuard{"always", "", ""}
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

func TestJsonGuardFnAlwaysImplicit(t *testing.T) {
	jg := JsonGuard{"", "", ""}
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
	jg := JsonGuard{"context", "hello", "world"}
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
	jg := JsonGuard{"invalid", "", ""}
	if _, err := jg.GuardFn(); err == nil {
		t.Log("Expected to fail (unknown guard type)")
		t.FailNow()
	}

	jg = JsonGuard{"context", "", nil}
	if _, err := jg.GuardFn(); err == nil {
		t.Log("Expected to fail (no key/value specified)")
		t.FailNow()
	}
	jg = JsonGuard{"context", "key", nil}
	if _, err := jg.GuardFn(); err == nil {
		t.Log("Expected to fail (no key/value specified)")
		t.FailNow()
	}
}

func TestJsonTransitionUnmarshal(t *testing.T) {
	rawJson := `
    {
        "to": "2",
        "action": {
        	"name": "hello",
        	"params": {
        		"what": "world"
        	}
        },
        "guard": {
            "type": "always"
        }
    }`
	var jt JsonTransition
	if err := json.Unmarshal([]byte(rawJson), &jt); err != nil {
		t.Logf("Unmarshalling failed: %s", err.Error())
		t.FailNow()
	}

	wh, found := jt.Action.Params["what"]
	if jt.ToState != "2" || jt.Guard.Type != "always" ||
		jt.Action.Name != "hello" || len(jt.Action.Params) != 1 || !found || wh != "world" {
		t.Logf("Unmarshalled different from expected:\nexpected: %s\nactual:%v", rawJson, jt)
		t.FailNow()
	}

}

func TestJsonTransitionFn(t *testing.T) {
	jt := JsonTransition{"2", JsonGuard{"always", "", nil}, JsonAction{"hello", nil}}
	act := make(ActionMap)
	if _, err := jt.Transition("1-2", act); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Log("Expected to fail (no action found)")
		t.FailNow()
	}
	act["hello"] = func(ctx ContextOperator) error { return nil }
	_, err := jt.Transition("1-2", act)
	if err != nil {
		t.Logf("Expected to pass, error: %s", err.Error())
		t.FailNow()
	}
}

func TestJsonStateUnmarshalValid1(t *testing.T) {
	rawJson := `
	{
		"start": true,
		"startsub": "11"
	}`

	var js JsonState
	if err := json.Unmarshal([]byte(rawJson), &js); err != nil {
		t.Logf("Unmarshalling failed: %s", err.Error())
		t.FailNow()
	}
	if js.StartSubState != "11" {
		t.Logf("Unmarshalled different from expected:\nexpected: %s\nactual:%v", rawJson, js)
		t.FailNow()
	}
}

func TestJsonStateUnmarshalValid2(t *testing.T) {
	rawJson := `
	{
		"parent": "1",
		"transitions": {
			"11-12": {
				"to": "12",
				"guard": {
					"type": "always"
				},
				"action": {
					"name": "setnext"
				}
			}
		}
	}`

	var js JsonState
	if err := json.Unmarshal([]byte(rawJson), &js); err != nil {
		t.Logf("Unmarshalling failed: %s", err.Error())
		t.FailNow()
	}
	if js.Parent != "1" {
		t.Logf("Unmarshalled different from expected:\nexpected: %s\nactual:%v", rawJson, js)
		t.FailNow()
	}
}

func TestJsonStartStateInfoValid(t *testing.T) {
	rawJson := `
	{
		"parent": "1",
		"startsub": "111"
	}`

	var js JsonState
	json.Unmarshal([]byte(rawJson), &js)

	act := make(map[string]ActionFn)
	parent := NewState("1", nil)

	si, err := js.StateInfo("11", parent, act)
	if err != nil {
		t.Logf("Constructing state info failed: %s", err.Error())
		t.FailNow()
	}
	if si.Name != "11" || si.Parent.Name != "1" || si.Transitions[0].ToState != "111" {
		t.Logf("StateInfo object is different from expected")
		t.FailNow()
	}
}

func TestJsonSubStateInfoValid(t *testing.T) {
	rawJson := `
	{
		"parent": "1",
		"transitions": {
			"11-12": {
				"to": "12",
				"guard": {
					"type": "always"
				},
				"action": "setnext"
			}
		}
	}`

	var js JsonState
	json.Unmarshal([]byte(rawJson), &js)

	act := ActionMap{"setnext": func(ctx ContextOperator) error { return nil }}
	parent := NewState("1", nil)

	si, err := js.StateInfo("state", parent, act)
	if err != nil {
		t.Logf("Constructing state info failed: %s", err.Error())
		t.FailNow()
	}

	if si.Parent == nil || si.Parent.Name != "1" {
		t.Logf("Parent is different from expected: %v", si.Parent)
		t.FailNow()
	}
	if len(si.Transitions) != 1 || si.Transitions[0].Name != "11-12" {
		t.Logf("Transitions are different from expected: %v", si.Transitions)
		t.FailNow()
	}
}

func TestJsonStateInfoInvalidParameters(t *testing.T) {
	rawJson := `
	{
		"parent": "1",
		"transitions": {
			"11-12": {
				"to": "12",
				"guard": {
					"type": "always"
				},
				"action": "setnext"
			}
		}
	}`

	var js JsonState
	json.Unmarshal([]byte(rawJson), &js)

	act := ActionMap{"setnext": func(ctx ContextOperator) error { return nil }}
	if _, err := js.StateInfo("11", nil, act); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Logf("StateInfo() should fail (parent defined but not passed): %s", err.Error())
		t.FailNow()
	}

	wrongParent := NewState("not1", nil)
	if _, err := js.StateInfo("11", wrongParent, act); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Logf("StateInfo() should fail (wrong parent passed): %s", err.Error())
		t.FailNow()
	}
}

func TestJsonStateInfoIllFormed(t *testing.T) {
	rawJson := `
	{
		"startsub": "1",
		"transitions": {
			"11-12": {
				"to": "12",
				"guard": {
					"type": "always"
				},
				"action": "setnext"
			}
		}
	}`

	var js JsonState
	json.Unmarshal([]byte(rawJson), &js)

	act := ActionMap{"setnext": func(ctx ContextOperator) error { return nil }}
	if _, err := js.StateInfo("11", nil, act); err == nil || err.Kind() != ErrFsmIsInvalid {
		t.Logf("StateInfo() should fail (state w/ start sub can't have costom transitions): %s",
			err.Error())
		t.FailNow()
	}
}
