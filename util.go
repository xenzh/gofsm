package simple_fsm

import (
	"bytes"
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
