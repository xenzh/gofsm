package simple_fsm

import (
	"bytes"
	"fmt"
	"strings"
)

// StateInfo
// Encapsulates state meta information, including entry action,
// relative hierarchy and list of outgoing transitions
type StateInfo struct {
	Name          string
	Parent        *StateInfo
	StartSubState *StateInfo
	Transitions   []Transition
}

// NewState
// Constructs state object
func NewState(name string, transitions []Transition) *StateInfo {
	return &StateInfo{name, nil, nil, transitions}
}

// addSubState
// Links given sub state with a parent
func (si *StateInfo) addSubState(sub *StateInfo, start bool) (err *FsmError) {
	if sub.Parent != nil {
		err = newFsmErrorStateIsInvalid(sub, "Sub state already has a parent")
		return
	}
	if start && si.StartSubState != nil {
		err = newFsmErrorStateIsInvalid(sub, "Parent state already has start sub state defined")
		return
	}

	sub.Parent = si
	if start {
		si.StartSubState = sub

		if len(si.Transitions) > 0 {
			err = newFsmErrorStateIsInvalid(si, `Start transition can't be used on
				parent with transitions defined, it'll lead to discarding original
				transitions`,
			)
			return
		}
		trName := fmt.Sprintf("Always %s->%s", si.Name, sub.Name)
		si.Transitions = NewTransitionAlways(trName, sub.Name, nil)
	}
	return
}

// newSubState
// Constructs child state, links it with a parent
func (si *StateInfo) newSubState(name string, transitions []Transition, start bool) (sub *StateInfo, err *FsmError) {
	sub = &StateInfo{name, nil, nil, transitions}
	err = si.addSubState(sub, start)
	return
}

// findCommonAncestor
// Searches for common parent for given 2 states
// As we should always have a common parent aka outermost/global state,
// it's enough to build a path to the root for both states
// and the point where paths start to differ is what we need
func findCommonAncestor(lhs *StateInfo, rhs *StateInfo) (ancestor *StateInfo, depth_diff int) {
	if lhs == rhs {
		return lhs, 0
	}

	make_path := func(state *StateInfo) []*StateInfo {
		path := []*StateInfo{state}
		curr := path[0].Parent
		for curr != nil {
			path = append([]*StateInfo{curr}, path...)
			curr = curr.Parent
		}
		return path
	}
	lhs_path, rhs_path := make_path(lhs), make_path(rhs)
	depth_diff = len(lhs_path) - len(rhs_path)

	for idx := range lhs_path {
		if idx >= len(rhs_path) || lhs_path[idx] != rhs_path[idx] {
			break
		}
		ancestor = lhs_path[idx]
	}
	return
}

// Validate
// Checks if given state is well-formed
func (si *StateInfo) Validate() (err *FsmError) {
	switch {
	case si.Name == "":
		err = newFsmErrorStateIsInvalid(si, "state should be named")
	case si.checkHierarchyCycled():
		err = newFsmErrorStateIsInvalid(si, "state hierarchy is cycled")
	case !si.Final():
		for idx := range si.Transitions {
			if err = si.Transitions[idx].Validate(); err != nil {
				break
			}
		}
	}
	return
}

// checkHierarchyCycled
// Check if there's a cycle in parent-child relations
// from botton to the top
func (si *StateInfo) checkHierarchyCycled() bool {
	visited := make(map[string]bool)

	for si != nil {
		if _, present := visited[si.Name]; present {
			return true
		}
		visited[si.Name] = true
		si = si.Parent
	}

	return false
}

// Final
// Checks if state if final e.g. has no outgoing transitions
func (si *StateInfo) Final() bool {
	return len(si.Transitions) <= 0
}

// dump
// Print out an object in a user-friendly way, composable
func (si *StateInfo) dump(buf *bytes.Buffer, indent int) {
	indentStr := strings.Repeat("\t", indent)

	buf.WriteString(indentStr)
	buf.WriteString("> name: \"")
	buf.WriteString(si.Name)
	buf.WriteString("\", parent: ")
	if si.Parent == nil {
		buf.WriteString("none")
	} else {
		buf.WriteString("\"")
		buf.WriteString(si.Parent.Name)
		buf.WriteString("\"")
	}
	if si.StartSubState != nil {
		buf.WriteString(", start substate: \"")
		buf.WriteString(si.StartSubState.Name)
		buf.WriteString("\"")
	}
	buf.WriteString("\n")
	buf.WriteString(indentStr)
	buf.WriteString("transitions:\n")
	trIndentStr := strings.Repeat("\t", indent+1)
	if len(si.Transitions) == 0 {
		buf.WriteString(trIndentStr)
		buf.WriteString("no transitions\n")
	} else {
		for _, tr := range si.Transitions {
			buf.WriteString(trIndentStr)
			tr.Dump(buf)
			buf.WriteString("\n")
		}
	}

}
