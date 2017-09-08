package simple_fsm

import (
	"bytes"
	"strings"
)

type HistoryItem struct {
	from       string
	to         string
	transition string
}
type History []HistoryItem

// Dump
// Print out an object in a user-friendly way
func (h *History) Dump() string {
	buf := bytes.NewBufferString("")
	h.dump(buf, 0)
	return buf.String()
}

// dumpImpl
// Print out an object in a user-friendly way, composable
func (h *History) dump(buf *bytes.Buffer, indent int) {
	indentStr := strings.Repeat("\t", indent)
	items := *h

	if len(items) == 0 {
		buf.WriteString(indentStr)
		buf.WriteString("(empty)")
	} else {
		for _, it := range items {
			buf.WriteString(indentStr)
			buf.WriteString("from: ")
			buf.WriteString(it.from)
			buf.WriteString(", to: ")
			buf.WriteString(it.to)
			buf.WriteString(", transition: ")
			buf.WriteString(it.transition)
			buf.WriteString("\n")
		}
	}
	buf.WriteString("\n")
}
