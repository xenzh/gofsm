package simple_fsm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// ActionMap
// Predefined set of actions that can be used by Builder to
// automatically define FSMs
type actionMap map[string]ActionFn

// Builder
// A tool for creating/loading FSMs
// * Allows to manually construct FSM (much like )
type Builder struct {
	actions *actionMap
	fsm     *Fsm
	err     *FsmError
}

// NewBuilder
// Constructs new builder
func NewBuilder(actions *actionMap) *Builder {
	return &Builder{actions, NewFsm(), nil}
}

// Fsm
// Returns constructed state machine or an error, if it's invalid
func (bld *Builder) Fsm() (fsm *Fsm, err *FsmError) {
	if bld.fsm == nil {
		err = newFsmErrorInvalid("FSM not created")
		return
	}
	if bld.err != nil {
		err = bld.err
		return
	}
	if err = bld.fsm.Validate(); err == nil {
		fsm = bld.fsm
	}
	return
}

func (bld *Builder) FromJsonFile(path string) (out *Builder, err error) {
	var rawJson []byte
	if rawJson, err = ioutil.ReadFile(path); err != nil {
		return
	}
	return bld.FromJson(rawJson)
}

func (bld *Builder) FromJson(rawJson []byte) (out *Builder, err error) {
	a := make(jsonRoot)
	err = json.Unmarshal(rawJson, &a)
	cause := fmt.Sprintf("unmarshalled json:\n%#v\n", a)

	return bld, newFsmErrorInvalid(cause)
}
