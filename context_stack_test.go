package simple_fsm

import (
	"testing"
)

func TestPush(t *testing.T) {
	cs := newContextStack()
	if !cs.Empty() || cs.Depth() != 0 {
		t.Log("New context stack should be empty")
		t.FailNow()
	}
	cs.Push(&StateInfo{})
	if cs.Empty() || cs.Depth() != 1 {
		t.Log("Context stack should contain single entry")
		t.FailNow()
	}
}

func TestPop(t *testing.T) {
	cs := newContextStack()
	ctx := cs.Pop()
	if ctx != nil {
		t.Log("Empty context stack should Pop() nil contexts")
		t.FailNow()
	}
	cs.Push(&StateInfo{})
	ctx = cs.Pop()
	if ctx == nil {
		t.Log("Non-empty context stack should Pop() valid contexts")
		t.FailNow()
	}
	if !cs.Empty() || cs.Depth() != 0 {
		t.Log("Context stack should be empty")
		t.FailNow()
	}
}

func TestPeek(t *testing.T) {
	cs := newContextStack()
	ctx := cs.Peek()
	if ctx != nil {
		t.Log("Empty context stack should Peek() nil contexts")
		t.FailNow()
	}
	cs.Push(&StateInfo{})
	ctx = cs.Peek()
	if ctx == nil {
		t.Log("Non-empty context stack should Peek() valid contexts")
		t.FailNow()
	}
}

func TestParent(t *testing.T) {
	cs := newContextStack()
	ctx := cs.Parent()
	if ctx != nil {
		t.Log("Empty context stack should PeekParent() nil contexts")
		t.FailNow()
	}
	cs.Push(&StateInfo{})
	ctx = cs.Parent()
	if ctx != nil {
		t.Log("Context stack w/ 1 element should PeekParent() nil contexts")
		t.FailNow()
	}
	cs.Push(&StateInfo{})
	ctx = cs.Parent()
	if ctx == nil {
		t.Log("Context stack w/ 2+ elements should PeekParent() valid contexts")
		t.FailNow()
	}
}

func TestGlobal(t *testing.T) {
	cs := newContextStack()
	cs.Push(&StateInfo{}).Put("global", 42)
	cs.Push(&StateInfo{}).Put("global", 7)

	if val, err := cs.Parent().context.Int("global"); val != 42 || err != nil {
		t.Log("Should get global key from outermost context")
		t.FailNow()
	}
}

func TestByState(t *testing.T) {
	cs := newContextStack()
	cs.Push(&StateInfo{Name: "global"}).Put("key", 9000)
	cs.Push(&StateInfo{Name: "interm"}).Put("key", 1337)
	cs.Push(&StateInfo{Name: "local"}).Put("key", 8080)

	ctx := cs.ByState("interm").context
	if val, err := ctx.Int("key"); val != 1337 || err != nil {
		t.Log("Should get key/value from interm state context")
		t.FailNow()
	}
}

func TestStackRaw(t *testing.T) {
	cs := newContextStack()
	cs.Push(&StateInfo{}).Put("base", 42)
	cs.Push(&StateInfo{}).Put("child", "imma string")

	raw, err := cs.Raw("child")
	if raw == nil || err != nil {
		t.Log("Child key should be found in head context")
		t.FailNow()
	}
	if val, ok := raw.(string); !ok || val != "imma string" {
		t.Log("Context member changed its value somehow")
		t.FailNow()
	}

	raw, err = cs.Raw("base")
	if raw == nil || err != nil {
		t.Log("Base key should be found in tail context")
		t.FailNow()
	}
	if val, ok := raw.(int); !ok || val != 42 {
		t.Log("Context member changed its value somehow")
		t.FailNow()
	}

	raw, err = cs.Raw("void")
	if raw != nil || err == nil {
		t.Log("Non-existent key should not be found")
		t.FailNow()
	}
}

func TestStackHas(t *testing.T) {
	cs := newContextStack()
	if found := cs.Has("key"); found {
		t.Log("Non-existent key should not be found")
		t.FailNow()
	}
	cs.Push(&StateInfo{}).Put("key", 42)
	if found := cs.Has("key"); !found {
		t.Log("Existing key should be found")
		t.FailNow()
	}
}

func TestStackBool(t *testing.T) {
	cs := newContextStack()
	cs.Push(&StateInfo{}).Put("bool", true)
	if value, err := cs.Bool("bool"); !value || err != nil {
		t.Log("Key should be found, value should be the same/have the same type")
		t.FailNow()
	}
	cs.Push(&StateInfo{}).Put("str", "imma string")
	if _, err := cs.Bool("str"); err == nil {
		t.Log("Key should be found, mismatching type error should be reported")
		t.FailNow()
	}
}

func TestStackInt(t *testing.T) {
	cs := newContextStack()
	cs.Push(&StateInfo{}).Put("int", 7)
	if value, err := cs.Int("int"); value != 7 || err != nil {
		t.Log("Key should be found, value should be the same/have the same type")
		t.FailNow()
	}
	cs.Push(&StateInfo{}).Put("str", "imma string")
	if _, err := cs.Int("str"); err == nil {
		t.Log("Key should be found, mismatching type error should be reported")
		t.FailNow()
	}
}

func TestStackStr(t *testing.T) {
	cs := newContextStack()
	cs.Push(&StateInfo{}).Put("str", "imma string")
	if value, err := cs.Str("str"); value != "imma string" || err != nil {
		t.Log("Key should be found, value should be the same/have the same type")
		t.FailNow()
	}
	cs.Push(&StateInfo{}).Put("bool", true)
	if _, err := cs.Str("bool"); err == nil {
		t.Log("Key should be found, mismatching type error should be reported")
		t.FailNow()
	}
}
