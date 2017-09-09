package simple_fsm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// ActionMap
// Predefined set of functions that can be used by Builder to
// automatically define state entry actions
type ActionMap map[string]ActionFn

// Builder
// A tool for creating/loading FSMs
// Now only supports loading FSM structure from json file/stream/objects
type Builder struct {
	actions *ActionMap
	fstr    *Structure
	err     *FsmError
}

// NewBuilder
// Constructs new builder
func NewBuilder(actions *ActionMap) *Builder {
	return &Builder{actions, NewStructure(), nil}
}

// Structure
// Returns constructed state machine structure or an construction fail error
func (bld *Builder) Structure() (fstr *Structure, err *FsmError) {
	if bld.fstr == nil || bld.fstr.Empty() {
		err = newFsmErrorLoading("FSM structure is not created yet")
		return
	}
	if bld.err != nil {
		err = bld.err
		return
	}
	if err = bld.fstr.Validate(); err == nil {
		fstr = bld.fstr
	}
	return
}

// FromJsonFile
// Constructs state machine structure from json file
func (bld *Builder) FromJsonFile(path string) *Builder {
	if bld.err != nil || bld.fstr.Empty() {
		return bld
	}

	var (
		rawJson []byte
		err     error
	)
	if rawJson, err = ioutil.ReadFile(path); err != nil {
		cause := fmt.Sprintf("I/O error occured: %s", err.Error())
		bld.err = newFsmErrorLoading(cause)
		return bld
	}
	return bld.FromRawJson(rawJson)
}

// FromRawJson
// Constructs state machine structure from json byte slice
func (bld *Builder) FromRawJson(rawJson []byte) *Builder {
	if bld.err != nil || bld.fstr.Empty() {
		return bld
	}

	root := make(jsonRoot)
	if err := json.Unmarshal(rawJson, &root); err != nil {
		cause := fmt.Sprintf("Unmarshalling error occured: %s", err.Error())
		bld.err = newFsmErrorLoading(cause)
		return bld
	}

	return bld.FromJsonType(root)
}

// FromJsonType
// Constructs state machine structure from unmarshalled json data structure
func (bld *Builder) FromJsonType(root jsonRoot) *Builder {
	return bld
}
