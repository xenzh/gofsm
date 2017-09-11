package simple_fsm

import (
	"bytes"
	"reflect"
)

// Dumper
// Interface for composable text debug output for package's entities
type Dumper interface {
	dump(*bytes.Buffer, int)
}

// Dump
// Print out an object in a user-friendly way to a string
// Works with objects that implement Dumper interface
func Dump(obj Dumper) string {
	buf := bytes.NewBufferString("")
	obj.dump(buf, 0)
	return buf.String()
}

// castToFloat64
// Tries to cast anything to float64 (aka universal number type)
// using reflection
func castToFloat64(what interface{}) (fl float64, err *FsmError) {
	v := reflect.ValueOf(what)
	v = reflect.Indirect(v)

	floatType := reflect.TypeOf(fl)
	if !v.Type().ConvertibleTo(floatType) {
		err = newFsmErrorRuntime("Cannot convert to float64", what)
		return
	}

	fv := v.Convert(floatType)
	fl = fv.Float()
	return
}
