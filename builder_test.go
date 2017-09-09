package simple_fsm

import (
	"testing"
)

func TestBuildStateHierarchyPositive(t *testing.T) {
	alwaysGuard := jsonGuard{"always", "", ""}
	js := jsonStates{
		"global": jsonState{"1", "", make(map[string]jsonTransition)},
		"1":      jsonState{"", "", map[string]jsonTransition{"1-2": {"2", alwaysGuard, ""}}},
		"2":      jsonState{"", "", map[string]jsonTransition{}},
	}

	_, list, err := buildStateHierarchy(js, ActionMap{})
	if err != nil {
		t.Logf("Hierarchy building unexpectedly failed: %s", err.Error())
		t.FailNow()
	}

	for k, v := range list {
		t.Logf("%s -- %#v\n", k, v)
	}
	t.FailNow()
}

func TestBuildStateHierarchy(t *testing.T) {

}
