package simple_fsm

import (
	"testing"
)

func TestAddSubState(t *testing.T) {
	okParent := NewState("name", nil)
	okChild := NewState("sub", nil)
	if err := okParent.addSubState(okChild, true); err != nil {
		t.Log("Adopting should succeed")
		t.FailNow()
	}
	if okParent.StartSubState != okChild || okChild.Parent != okParent {
		t.Log("Parent-child links should be set up properly")
		t.FailNow()
	}
	if len(okParent.Transitions) != 1 || okParent.Transitions[0].Guard == nil {
		t.Log("Adoption should create an unconditional parent->child transition")
		t.FailNow()
	}

	badParent := NewState("name", nil)
	badChild := NewState("sub", nil)
	badChild.Parent = okParent
	if err := badParent.addSubState(badChild, true); err == nil || err.Kind() != ErrStateIsInvalid {
		t.Log("Adoption should fail (substate already has a parent)")
		t.FailNow()
	}

	badParent = NewState("name", nil)
	badChild = NewState("sub", nil)
	badParent.StartSubState = okChild
	if err := badParent.addSubState(badChild, true); err == nil || err.Kind() != ErrStateIsInvalid {
		t.Log("Adoption should fail (parent already has start substate)")
		t.FailNow()
	}

	badParent = NewState("name", NewTransitionAlways("name->sub", "sub", nil))
	badChild = NewState("sub", nil)
	if err := badParent.addSubState(badChild, true); err == nil || err.Kind() != ErrStateIsInvalid {
		t.Log("Adoption should fail (parent already has transitions defined)")
		t.FailNow()
	}
}

func TestNewSubState(t *testing.T) {
	parent := NewState("name", nil)
	child, err := parent.newSubState("sub", nil, true)
	if err != nil {
		t.Logf("newSubState failed: %s", err.Error())
		t.FailNow()
	}

	if parent.StartSubState != child {
		t.Log("Substate should be registered as initial for the parent")
		t.FailNow()
	}
	if child.Parent != parent {
		t.Log("Parent should be registered in sub state")
		t.FailNow()
	}
}

func TestFindCommonAncestorNegative(t *testing.T) {
	g1 := NewState("one", nil)
	g2 := NewState("two", nil)

	fwd, _ := findCommonAncestor(g1, g2)
	bwd, _ := findCommonAncestor(g2, g1)
	if fwd != nil || bwd != nil {
		t.Log("States should not be related")
		t.FailNow()
	}

	g11, _ := g1.newSubState("three", nil, true)
	g111, _ := g11.newSubState("g111", nil, false)
	g1111, _ := g111.newSubState("g1111", nil, false)
	g21, _ := g2.newSubState("g21", nil, false)

	fwd, _ = findCommonAncestor(g1111, g21)
	bwd, _ = findCommonAncestor(g21, g1111)
	if fwd != nil || bwd != nil {
		t.Log("Child states from different trees should not be related")
		t.FailNow()
	}
}

func TestFindCommonAncestorPositive(t *testing.T) {
	g := NewState("root", nil)
	c1, _ := g.newSubState("c1", nil, false)

	fwd, fwdD := findCommonAncestor(c1, c1)
	if fwd != c1 {
		t.Log("The same state is it's own common ancestor")
		t.FailNow()
	}
	if fwdD != 0 {
		t.Log("State can't have a depth difference with itself")
		t.FailNow()
	}

	fwd, fwdD = findCommonAncestor(g, c1)
	bwd, bwdD := findCommonAncestor(c1, g)
	if fwd != g || bwd != g {
		t.Log("Ancestor may be one of the input states")
		t.FailNow()
	}
	if fwdD != -1 || bwdD != 1 {
		t.Log("First state should be deeper than second by 1")
		t.FailNow()
	}

	c2, _ := g.newSubState("c2", nil, false)

	fwd, fwdD = findCommonAncestor(c1, c2)
	bwd, bwdD = findCommonAncestor(c2, c1)
	if fwd != g || bwd != g {
		t.Log("Brother children should have a common ancestor")
		t.FailNow()
	}
	if fwdD != 0 || bwdD != 0 {
		t.Log("Brother children should have the same depth")
		t.FailNow()
	}

	c11, _ := c1.newSubState("c11", nil, false)
	c111, _ := c11.newSubState("c111", nil, false)

	fwd, fwdD = findCommonAncestor(c2, c111)
	bwd, bwdD = findCommonAncestor(c111, c2)
	if fwd != g || bwd != g {
		t.Log("Children should have a common ancestor")
		t.FailNow()
	}
	if fwdD != -2 || bwdD != 2 {
		t.Log("Second state should be deeper than first by 2")
		t.FailNow()
	}

	c21, _ := c2.newSubState("c21", nil, false)

	fwd, fwdD = findCommonAncestor(c21, c111)
	bwd, bwdD = findCommonAncestor(c111, c21)
	if fwd != g || bwd != g {
		t.Log("Children should have a common ancestor")
		t.FailNow()
	}
	if fwdD != -1 || bwdD != 1 {
		t.Log("Second state should be deeper than first by 1")
		t.FailNow()
	}
}

func TestStateValidate(t *testing.T) {
	parent := NewState("name", nil)
	child, err := parent.newSubState("sub", nil, true)
	if err != nil {
		t.Logf("newSubState failed: %s", err.Error())
		t.FailNow()
	}

	if err := parent.Validate(); err != nil {
		t.Logf("Parent state should be valid, error: %s", err.Error())
		t.FailNow()
	}
	if err := child.Validate(); err != nil {
		t.Logf("Child state should be valid, error: %s", err.Error())
		t.FailNow()
	}

	si := NewState("", nil)
	if err := si.Validate(); err == nil || err.Kind() != ErrStateIsInvalid {
		t.Log("State should be invalid (empty name)")
		t.FailNow()
	}

	outer := NewState("outer", nil)
	interm := NewState("interm", nil)
	inner := NewState("inner", nil)

	if e := outer.addSubState(interm, true); e != nil {
		t.Logf("Failed adding a substate, error: %s", e)
		t.FailNow()
	}
	if e := interm.addSubState(inner, true); e != nil {
		t.Logf("Failed adding a substate, error: %s", e)
		t.FailNow()
	}
	if e := inner.addSubState(outer, true); e != nil {
		t.Logf("Failed adding a substate, error: %s", e)
		t.FailNow()
	}

	if err := interm.Validate(); err == nil || err.Kind() != ErrStateIsInvalid {
		t.Logf("States should be invalid (cyclic)\n%s\n%s\n%s",
			Dump(outer),
			Dump(interm),
			Dump(inner),
		)
		t.FailNow()
	}
}
