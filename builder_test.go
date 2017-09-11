package simple_fsm

import (
	"testing"
)

type pC struct {
	state  string
	parent string
	ssub   string
}

func makeJsonStates(start pC, pcs ...pC) JsonStates {
	js := make(JsonStates)
	trm := make(map[string]JsonTransition)

	js[start.state] = JsonState{true, start.ssub, "", trm}

	for _, pc := range pcs {
		js[pc.state] = JsonState{false, pc.ssub, pc.parent, trm}
	}
	return js
}

func checkState(t *testing.T, st *StateInfo, parent string) {
	root := st.Parent == nil && parent == ""
	sub := !root && st.Parent.Name == parent
	if !root && !sub {
		t.Logf("State %s has parent (%v) different from expected (%s)",
			st.Name, st.Parent, parent)
		t.FailNow()
	}
}

func TestBuildStateHierarchyPositive1(t *testing.T) {
	js := makeJsonStates(
		pC{"0", "", ""},
		pC{"1", "", ""},
		pC{"2", "", ""},
	)

	start, list, err := buildStateHierarchy(js, ActionMap{})
	if err != nil {
		t.Logf("Hierarchy building unexpectedly failed: %s", err.Error())
		t.FailNow()
	}

	if start == nil || start.Name != "0" {
		t.Log("Start state is different from expected")
		t.FailNow()
	}

	checkState(t, start, "")
	checkState(t, list["1"], "")
	checkState(t, list["2"], "")
}

func TestBuildStateHierarchyPositive2(t *testing.T) {
	js := makeJsonStates(
		pC{"0", "", "1"},
		pC{"1", "0", ""},
		pC{"2", "0", ""},
	)

	start, list, err := buildStateHierarchy(js, ActionMap{})
	if err != nil {
		t.Logf("Hierarchy building unexpectedly failed: %s", err.Error())
		t.FailNow()
	}

	if start == nil || start.Name != "0" {
		t.Log("Start state is different from expected")
		t.FailNow()
	}

	checkState(t, start, "")
	checkState(t, list["1"], "0")
	checkState(t, list["2"], "0")
}

func TestBuildStateHierarchyPositive3(t *testing.T) {
	js := makeJsonStates(
		pC{"0", "", "1"},
		pC{"31", "2", ""},
		pC{"21", "1", ""},
		pC{"3", "2", ""},
		pC{"2", "1", "3"},
		pC{"1", "0", "2"},
		pC{"22", "1", ""},
	)

	start, list, err := buildStateHierarchy(js, ActionMap{})
	if err != nil {
		t.Logf("Hierarchy building unexpectedly failed: %s", err.Error())
		t.FailNow()
	}

	if start == nil || start.Name != "0" {
		t.Log("Start state is different from expected")
		t.FailNow()
	}

	checkState(t, start, "")
	checkState(t, list["1"], "0")
	checkState(t, list["2"], "1")
	checkState(t, list["21"], "1")
	checkState(t, list["22"], "1")
	checkState(t, list["3"], "2")
	checkState(t, list["31"], "2")
}

func TestBuildStateHierarchyCycled(t *testing.T) {
	js := makeJsonStates(
		pC{"0", "", "1"},
		pC{"1", "0", "2"},
		pC{"2", "1", "3"},
		pC{"3", "2", "0"},
	)

	_, _, err := buildStateHierarchy(js, ActionMap{})
	if err == nil || err.Kind() != ErrFsmLoading {
		t.Log("Hierarchy building is expected to fail (state hierarchy cycled)")
		t.FailNow()
	}
}

func TestBuildStateSeveralEntryPoints(t *testing.T) {
	js := makeJsonStates(
		pC{"0", "", "1"},
		pC{"1", "0", "2"},
		pC{"2", "1", "3"},
		pC{"3", "2", ""},
	)
	js["4"] = JsonState{true, "", "2", nil}

	_, _, err := buildStateHierarchy(js, ActionMap{})
	if err == nil || err.Kind() != ErrFsmLoading {
		t.Log("Hierarchy building is expected to fail (several entry points)")
		t.FailNow()
	}
}

func TestBuilderFromJsonFile(t *testing.T) {
	actions := ActionMap{
		"setnext": func(ctx ContextOperator) error {
			ctx.Put("next", 14)
			return nil
		},
		"setresult13": func(ctx ContextOperator) error {
			ctx.PutResult(13)
			return nil
		},
		"setresult42": func(ctx ContextOperator) error {
			ctx.PutResult(42)
			return nil
		},
	}
	fstr, berr := NewBuilder(actions).FromJsonFile("./fsm-sample.json").Structure()
	if berr != nil {
		t.Logf("Structure construction failed, %s", berr.Error())
		t.FailNow()
	}

	fsm := NewFsm(fstr)
	res, rerr := fsm.Run()
	if rerr != nil {
		t.Logf("Loaded FSM execution failed: %s", rerr.Error())
		t.Logf("State machine dump:\n%s", Dump(fsm))
		t.FailNow()
	}
	if val, ok := res.(int); !ok || val != 42 {
		t.Logf("FSM result (%v) is different from extected (%v)", val, 42)
		t.FailNow()
	}
}
