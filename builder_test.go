package simple_fsm

import (
	"fmt"
	"testing"
)

type pC struct {
	state  string
	parent string
	ssub   string
}

func makeJsonStates(start pC, pcs ...pC) jsonStates {
	js := make(jsonStates)
	trm := make(map[string]jsonTransition)

	js[start.state] = jsonState{true, start.ssub, "", trm}

	for _, pc := range pcs {
		js[pc.state] = jsonState{false, pc.ssub, pc.parent, trm}
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

func TestBuildStateHierarchyPositive2(t *testing.T) {
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

	t.Logf("start state: %s (parent: %v)", start.Name, start.Parent)
	for k, v := range list {
		parentName := "(none)"
		if v.Parent != nil {
			parentName = v.Parent.Name
		}
		st := fmt.Sprintf("%s (parent: %s)", k, parentName)
		t.Logf("%s\n", st)
	}
	//t.FailNow()

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

}
