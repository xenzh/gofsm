package simple_fsm

import (
	"bytes"
	"fmt"
	"strings"
)

// ContextAccessor
// Interface describing context accessors
type ContextAccessor interface {
	Has(string) bool
	Raw(string) (interface{}, *FsmError)
	Bool(string) (bool, *FsmError)
	Int(string) (int, *FsmError)
	Float(string) (float64, *FsmError)
	Str(string) (string, *FsmError)
}

// ContextModifier
// Interface describing context modifiers
type ContextModifier interface {
	Put(string, interface{}) *FsmError
	PutParent(string, interface{}) *FsmError
	PutResult(interface{}) *FsmError
}

// ContextOperator
// Interface describing full Context functionality
type ContextOperator interface {
	ContextAccessor
	ContextModifier
}

// Context
// Provides associative storage for objects of any type
// Implements both ContextAccessor and ContextModifier
// Provides means for adding objects to context and
// getting them out, either as interface{} or bool/int/string
type Context struct {
	members map[string]interface{}
}

// newContext
// Consturcts and initializes context instance
func newContext() Context {
	return Context{members: make(map[string]interface{})}
}

// ContextModifier.Put
// Adds new / modifies a member of the context
func (ctx *Context) Put(key string, value interface{}) *FsmError {
	ctx.members[key] = value
	return nil
}

// ContextModifier.PutResult
// Adds new / modifies result member of the context
func (ctx *Context) PutResult(result interface{}) *FsmError {
	ctx.Put(FsmResultCtxMemberName, result)
	return nil
}

// ContextAccessor.Has
// Check whether key is present in the stack
func (ctx *Context) Has(key string) bool {
	_, present := ctx.members[key]
	return present
}

// ContextAccessor.Raw
// Searches for given key in the context,
// Returns interface{}-boxed value
func (ctx *Context) Raw(key string) (value interface{}, err *FsmError) {
	var present bool
	value, present = ctx.members[key]
	if !present {
		err = newCtxErrorKeyNotFound(key)
	}
	return
}

// ContextAccessor.Bool
// Searches for given key in the context,
// Casts value to bool and returns it
func (ctx *Context) Bool(key string) (value bool, err *FsmError) {
	var raw interface{}
	if raw, err = ctx.Raw(key); err == nil {
		var ok bool
		if value, ok = raw.(bool); !ok {
			err = newCtxErrorInvalidType(value, raw)
		}
	}
	return
}

// ContextAccessor.Int
// Searches for given key in the context,
// Casts value to int and returns it
func (ctx *Context) Int(key string) (value int, err *FsmError) {
	var raw interface{}
	if raw, err = ctx.Raw(key); err == nil {
		var ok bool
		if value, ok = raw.(int); !ok {
			err = newCtxErrorInvalidType(value, raw)
		}
	}
	return
}

// ContextAccessor.Float
// Searches for given key in the context,
// Casts value to float64 and returns it
func (ctx *Context) Float(key string) (value float64, err *FsmError) {
	var raw interface{}
	if raw, err = ctx.Raw(key); err == nil {
		var ok bool
		if value, ok = raw.(float64); !ok {
			err = newCtxErrorInvalidType(value, raw)
		}
	}
	return
}

// ContextAccessor.Str
// Searches for given key in the context,
// Casts value to str and returns it
func (ctx *Context) Str(key string) (value string, err *FsmError) {
	var raw interface{}
	if raw, err = ctx.Raw(key); err == nil {
		var ok bool
		if value, ok = raw.(string); !ok {
			err = newCtxErrorInvalidType(value, raw)
		}
	}
	return
}

// dump
// Print out an object in a user-friendly way, composable
func (ctx *Context) dump(buf *bytes.Buffer, indent int) {
	indentStr := strings.Repeat("\t", indent)

	if len(ctx.members) == 0 {
		buf.WriteString(indentStr)
		buf.WriteString("(empty)")
	} else {
		for k, v := range ctx.members {
			buf.WriteString(fmt.Sprintf("%s%s: %v\n", indentStr, k, v))
		}
	}
	buf.WriteString("\n")
}
