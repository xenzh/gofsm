package simple_fsm

import (
	"fmt"
)

type JsonGuard struct {
	Type  string      `json:"type"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func (jg *JsonGuard) GuardFn() (guard GuardFn, err *FsmError) {
	switch jg.Type {
	case "always", "":
		// empty value means no guard specified => unconditional transition implication
		guard = func(ctx ContextAccessor) (bool, error) { return true, nil }
	case "context":
		if len(jg.Key) == 0 || jg.Value == nil {
			err = newFsmErrorInvalid("No key/value specified")
			return
		}
		// this extra closure is required to evaluate jg.Key and jg.Value values as parameters
		// in order to avoid all guards closures referencing the same key/value objects
		// from last transition object of the state
		guard = func(key string, value interface{}) GuardFn {
			return func(ctx ContextAccessor) (bool, error) {
				var open bool
				raw, e := ctx.Raw(key)
				if e == nil {
					// See https://blog.golang.org/json-and-go for default unmarshal types
					switch v := value.(type) {
					case bool, string, nil:
						open = v == raw
					case float64:
						var fl float64
						fl, e = castToFloat64(raw)
						open = (e == nil && v == fl)
					default:
						e = newFsmErrorInvalid("Internal error: unknown unmarshalled type")
					}
				}
				if e == nil {
					return open, nil
				} else {
					return open, e
				}
			}
		}(jg.Key, jg.Value)
	default:
		err = newFsmErrorInvalid("unknown guard type")
	}
	return
}

type JsonTransition struct {
	ToState string    `json:"to"`
	Guard   JsonGuard `json:"guard"`
	Action  string    `json:"action"`
}

func (jt *JsonTransition) Transition(name string, actions ActionMap) (tr Transition, err *FsmError) {
	var action ActionFn
	if len(jt.Action) > 0 {
		if act, present := actions[jt.Action]; !present {
			cause := fmt.Sprintf("action \"%s\" was not found in the map: %v", jt.Action, actions)
			err = newFsmErrorInvalid(cause)
			return
		} else {
			action = act
		}
	}

	var guard GuardFn
	if guard, err = jt.Guard.GuardFn(); err != nil {
		return
	}

	tr = NewTransition(name, jt.ToState, guard, action)
	return
}

type JsonState struct {
	Start         bool                      `json:"start"`
	StartSubState string                    `json:"startsub"`
	Parent        string                    `json:"parent"`
	Transitions   map[string]JsonTransition `json:"transitions"`
}

func (js JsonState) StateInfo(name string, parent *StateInfo, actions ActionMap) (si *StateInfo, err *FsmError) {
	var start bool
	if len(js.StartSubState) > 0 {
		if len(js.Transitions) > 0 {
			err = newFsmErrorInvalid("State w/ start sub state can't have custom transitions")
			return
		}
		trName := fmt.Sprintf("Always %s->%s", name, js.StartSubState)
		si = NewState(name, NewTransitionAlways(trName, js.StartSubState, nil))
		start = true
	} else {
		trs := make([]Transition, 0, len(js.Transitions))
		for trName, jtr := range js.Transitions {
			var tr Transition
			if tr, err = jtr.Transition(trName, actions); err != nil {
				return
			}
			trs = append(trs, tr)
		}
		si = NewState(name, trs)
	}

	if len(js.Parent) > 0 {
		if parent == nil {
			err = newFsmErrorInvalid("Json defined a parent, but parent object is empty")
			return
		}
		if parent.Name != js.Parent {
			cause := fmt.Sprintf("Parent (%s) is different from expected (%s)", parent.Name, js.Parent)
			err = newFsmErrorInvalid(cause)
			return
		}

		parent.addSubState(si, start)
	}

	return
}

type JsonStates map[string]JsonState
type JsonRoot map[string]JsonStates
