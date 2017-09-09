package simple_fsm

import (
	"fmt"
)

type jsonGuard struct {
	Type  string      `json:"type"`
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func (jg *jsonGuard) GuardFn() (guard GuardFn, err *FsmError) {
	switch jg.Type {
	case "always":
		guard = func(ctx ContextAccessor) (bool, error) { return true, nil }
	case "context":
		if len(jg.Key) == 0 || jg.Value == nil {
			err = newFsmErrorInvalid("no key/value specified")
			return
		}
		guard = func(ctx ContextAccessor) (bool, error) {
			raw, err := ctx.Raw(jg.Key)
			if err == nil {
				return raw == jg.Value, nil
			}
			return false, err
		}
	default:
		err = newFsmErrorInvalid("unknown guard type")
	}
	return
}

type jsonTransition struct {
	ToState string    `json:"to"`
	Guard   jsonGuard `json:"guard"`
	Action  string    `json:"action"`
}

func (jt *jsonTransition) Transition(name string, actions ActionMap) (tr Transition, err *FsmError) {
	if _, present := actions[jt.Action]; !present {
		cause := fmt.Sprintf("action \"%s\" was not found in the map: %v", jt.Action, actions)
		err = newFsmErrorInvalid(cause)
		return
	}
	var guard GuardFn
	if guard, err = jt.Guard.GuardFn(); err != nil {
		return
	}
	tr = NewTransition(name, jt.ToState, guard, actions[jt.Action])
	return
}

type jsonState struct {
	StartSubState string                    `json:"startsub"`
	Parent        string                    `json:"parent"`
	Transitions   map[string]jsonTransition `json:"transitions"`
}

func (js *jsonState) StateInfo(name string, parent *StateInfo, actions ActionMap) (si *StateInfo, err *FsmError) {
	var start bool
	if len(js.StartSubState) > 0 {
		if len(js.Transitions) > 0 {
			err = newFsmErrorInvalid("State w/ start sub state can't have custom transitions")
			return
		}
		si = NewState(name, NewTransitionAlways("always", js.StartSubState, nil))
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

type jsonStates map[string]jsonState
type jsonRoot map[string]jsonStates
