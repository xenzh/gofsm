package simple_fsm

import (
	"bytes"
	"fmt"
	"strings"
)

//
// StateContext
//

// Element of context stack, encapsulates state info and context
type StateContext struct {
	state   *StateInfo
	context Context
}

// newStateContext
// Creates state context instance
func newStateContext(state *StateInfo) StateContext {
	return StateContext{state: state, context: newContext()}
}

// ContextModifier.Put
// Adds new / modifies existing member of underlying context
func (sc *StateContext) Put(key string, value interface{}) {
	sc.context.Put(key, value)
}

// ContextModifier.PutResult
// Adds new / modifies result member of underlying context
func (sc *StateContext) PutResult(result interface{}) {
	sc.context.PutResult(result)
}

//
// ContextStack
//

// LIFO collection of contexts
// defines usual stack operations like pop/push/peek
// implements ContextAccessor - meaning it can be read exactly like
// Context (writing is intentionally excluded)
type ContextStack struct {
	stack []StateContext
}

// newContextStack
// Constructs context stack
func newContextStack() ContextStack {
	return ContextStack{}
}

// Depth
// Returns number of elements in stack context
func (st *ContextStack) Depth() int {
	return len(st.stack)
}

// Empty
// Returns true is stack context does not have a single element
func (st *ContextStack) Empty() bool {
	return st.Depth() <= 0
}

// Peek
// Returns head of the stack, without modification
func (st *ContextStack) Peek() *StateContext {
	if st.Empty() {
		return nil
	}
	return &st.stack[len(st.stack)-1]
}

// Parent
// Returns context previous to head, without modification
func (st *ContextStack) Parent() *StateContext {
	if st.Depth() < 2 {
		return nil
	}

	return &st.stack[len(st.stack)-2]
}

// Global
// Returns global (outermost) state context or nil if stack is empty
func (st *ContextStack) Global() *StateContext {
	if st.Empty() {
		return nil
	}
	return &st.stack[0]
}

// ByState
// Searches for a context given associated state name
// Will return nil in case stack is empty or there's no such state
// in active stack
func (st *ContextStack) ByState(name string) *StateContext {
	if st.Empty() {
		return nil
	}

	for idx := len(st.stack) - 1; idx >= 0; idx-- {
		if st.stack[idx].state.Name == name {
			return &st.stack[idx]
		}
	}
	return nil
}

// Pop
// Returns head of the stack, removes it from the stack
func (st *ContextStack) Pop() *StateContext {
	if st.Empty() {
		return nil
	}

	head := &st.stack[len(st.stack)-1]
	st.stack = st.stack[:len(st.stack)-1]
	return head
}

// Push
// Adds new context to the head of the stack, returns it
// Does not accept nil and duplicate states
func (st *ContextStack) Push(state *StateInfo) *StateContext {
	if state == nil {
		return nil
	}
	if !st.Empty() && st.Peek().state == state {
		return nil
	}

	st.stack = append(st.stack, newStateContext(state))
	return st.Peek()
}

// ContextAccessor.Raw
// Searches for given key in all contexts present in the stack,
// from head to tail, returns interface{}-boxed value
// Note: if there are duplicate keys in different contexts,
// one closest to the head will overshadow others
func (st *ContextStack) Raw(key string) (value interface{}, err *FsmError) {
	if st.Empty() {
		return nil, newCtxErrorKeyNotFound(key)
	}

	for idx := len(st.stack) - 1; idx >= 0; idx-- {
		if value, err = st.stack[idx].context.Raw(key); err == nil {
			return
		}
	}
	return
}

// ContextAccessor.Has
// Check whether key is present in any context in the stack
func (st *ContextStack) Has(key string) bool {
	_, err := st.Raw(key)
	return err == nil
}

// ContextAccessor.Bool
// Searches for given key in all contexts present in the stack,
// from head to tail, casts value to bool and returns it
func (st *ContextStack) Bool(key string) (value bool, err *FsmError) {
	var raw interface{}
	if raw, err = st.Raw(key); err == nil {
		var ok bool
		if value, ok = raw.(bool); !ok {
			err = newCtxErrorInvalidType(value, raw)
		}
	}
	return
}

// ContextAccessor.Int
// Searches for given key in all contexts present in the stack,
// from head to tail, casts value to int and returns it
func (st *ContextStack) Int(key string) (value int, err *FsmError) {
	var raw interface{}
	if raw, err = st.Raw(key); err == nil {
		var ok bool
		if value, ok = raw.(int); !ok {
			err = newCtxErrorInvalidType(value, raw)
		}
	}
	return
}

// ContextAccessor.Str
// Searches for given key in all contexts present in the stack,
// from head to tail, casts value to string and returns it
func (st *ContextStack) Str(key string) (value string, err *FsmError) {
	var raw interface{}
	if raw, err = st.Raw(key); err == nil {
		var ok bool
		if value, ok = raw.(string); !ok {
			err = newCtxErrorInvalidType(value, raw)
		}
	}
	return
}

// ContextModifier.Put
// Adds new / modifies a member of underlying context
func (st *ContextStack) Put(key string, value interface{}) (err *FsmError) {
	if st.Empty() {
		err = newFsmErrorRuntime("Can't put to empty context stack", st)
		return
	}
	st.Peek().Put(key, value)
	return
}

// ContextModifier.PutResult
// Adds new / modifies result member of underlying context
func (st *ContextStack) PutResult(result interface{}) (err *FsmError) {
	if st.Empty() {
		err = newFsmErrorRuntime("Can't put to empty context stack", st)
		return
	}
	st.Global().PutResult(result)
	return
}

// dump
// Print out an object in a user-friendly way, composable
func (st *ContextStack) dump(buf *bytes.Buffer, indent int) {
	indentStr := strings.Repeat("\t", indent)
	for _, elem := range st.stack {
		buf.WriteString(fmt.Sprintf("%s> state: \"%s\"\n", indentStr, elem.state.Name))
		elem.context.dump(buf, indent+1)
	}
}
