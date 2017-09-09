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
	actions ActionMap
	fstr    *Structure
	err     *FsmError
}

// NewBuilder
// Constructs new builder
func NewBuilder(actions ActionMap) *Builder {
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
// Constructs state machine structure from json file (see json format below)
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
// Constructs state machine structure from json byte slice.
// Json format (see fsm-sample.json):
// {
//     "states": {                             -- required key name for state list
//         "global": {                         -- required key name for FSM entry point
//             "startsub": "1"                 -- required field, FSM entry state
//         },
//         "0": {                              -- no parent key implies global parent (topmost state)
//             "startsub": "1"                 -- parent state can't have custom transitions, only start substate
//         },
//         "1": {
//             "parent": "0",                  -- empty parent means top level state
//             "transitions": {
//                 "1-2": {
//                     "to": "2",              -- transition target, should be in the same level of hierarchy
//                     "guard": {              -- no guard key implies unconditional transition
//                         "type": "always"    -- can be either unconditional or conditional (see below)
//                     },
//                     "action": "setnext"     -- action key should be present in builder's ActionMap
//                 }
//             }
//         },
//         "2": {
//             "parent": "0",
//             "transitions": {
//                 "2-3": {
//                     "to": "3",
//                     "guard": {
//                         "type": "context",  -- conditional guard, checks named value from the context against a value
//                         "key": "next",      -- context key to check
//                         "value": "42"       -- expected key value for the guard to open
//                     }
//                 }
//             }
//         },
//         "3": {
//             "parent": "0"                   -- no transitions means final state (FSM will be considered completed)
//         }
//     }
// }
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
	if bld.err != nil || bld.fstr.Empty() {
		return bld
	}

	jsStates, found := root["states"]
	if !found {
		bld.err = newFsmErrorLoading("Json is ill-formed: no top-level \"states\" object found")
		return bld
	}

	start, list, err := buildStateHierarchy(jsStates, bld.actions)
	if err != nil {
		bld.err = err
		return bld
	}

	bld.err = bld.fstr.appendStates(start, list)

	return bld
}

// depMarkers, depGraph, depStates
// Internal data structures for calculating state dependency order
type depMarker struct {
	visited  bool
	visiting bool
}
type depMarkers []depMarker
type depGraph [][]bool
type depStates map[string]*StateInfo

// buildStateHierarchy
// States are build so that you have to have a parent to be able to add a substate to the structure.
// Json doesn't constrain states in any way so they could be in any order.
// So input json states need to be traversed from topmost parents to downmost children to make a proper structure.
// This method... TBD
func buildStateHierarchy(states jsonStates, actions ActionMap) (start *StateInfo, list depStates, err *FsmError) {
	count := len(states)

	// map state indexes to names
	names := make([]string, count)
	indexes := make(map[string]int)
	var idx int
	for k, _ := range states {
		names[idx] = k
		indexes[k] = idx
		idx++
	}

	// build dependency graph
	// graph[i][j] == true means i depends on j (i is a child of j)
	graph := make(depGraph, count)
	for i, _ := range graph {
		graph[i] = make([]bool, count)
	}
	for name, state := range states {
		if len(state.Parent) == 0 {
			continue
		}

		i := indexes[name]
		j := indexes[state.Parent]
		graph[i][j] = true
	}

	// satisfy dependencies of every state
	list = make(depStates)
	markers := make(depMarkers, count)
	for _, idx := range indexes {
		if err = satisfyDependencies(idx, nil, graph, markers, names, states, list, actions); err != nil {
			break
		}
	}

	return
}

func satisfyDependencies(
	index int,
	parentIndex *int,
	graph depGraph,
	markers depMarkers,
	names []string,
	source jsonStates,
	dest depStates,
	actions ActionMap,
) *FsmError {

	if markers[index].visiting {
		return newFsmErrorLoading("State hierarchy is cycled")
	}
	if markers[index].visited {
		return nil
	}

	markers[index].visiting = true

	for on, depends := range graph[index] {
		if depends {
			err := satisfyDependencies(on, &index, graph, markers, names, source, dest, actions)
			if err != nil {
				return err
			}
		}
	}

	name := names[index]
	var parent *StateInfo
	if parentIndex != nil {
		parent = dest[names[*parentIndex]]
		if parent == nil {
			return newFsmErrorLoading("Parent is expected to be added before the child")
		}
	}

	si, err := source[name].StateInfo(name, parent, actions)
	if err != nil {
		return err
	}

	dest[name] = si

	markers[index].visited = true
	markers[index].visiting = true

	return nil
}
