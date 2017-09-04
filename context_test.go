package simple_fsm

import (
	"testing"
)

func TestHas(t *testing.T) {
	ctx := newContext()
	ctx.Put("key", 42)
	if !ctx.Has("key") {
		t.Log("key should be present")
		t.Fail()
	}
}

func TestRawExisting(t *testing.T) {
	ctx := newContext()
	ctx.Put("key", 42)

	raw, err := ctx.Raw("key")
	if err != nil {
		t.Log("Key should be present")
		t.FailNow()
	}

	val, ok := raw.(int)
	if !ok {
		t.Log("Context member changed its type somehow")
		t.FailNow()
	}
	if val != 42 {
		t.Log("Context member changed its value somehow")
	}
}

func TestRawMissing(t *testing.T) {
	ctx := newContext()
	if _, err := ctx.Raw("key"); err == nil {
		t.Log("Key should be missing")
		t.FailNow()
	}
}

func TestBoolRightType(t *testing.T) {
	ctx := newContext()
	ctx.Put("key", true)

	val, err := ctx.Bool("key")
	if err != nil {
		t.Log("Key should be present")
		t.FailNow()
	}
	if !val {
		t.Log("Context member changed its value somehow")
		t.FailNow()
	}
}

func TestBoolWrongType(t *testing.T) {
	ctx := newContext()
	ctx.Put("key", "a string")
	if _, err := ctx.Bool("key"); err == nil {
		t.Log("Key type should be different")
		t.FailNow()
	}
}

func TestBoolMissing(t *testing.T) {
	ctx := newContext()
	if _, err := ctx.Bool("key"); err == nil {
		t.Log("Key should be missing")
		t.FailNow()
	}
}

func TestIntRightType(t *testing.T) {
	ctx := newContext()
	ctx.Put("key", 42)

	val, err := ctx.Int("key")
	if err != nil {
		t.Log("Key should be present")
		t.FailNow()
	}
	if val != 42 {
		t.Log("Context member changed its value somehow")
		t.FailNow()
	}
}

func TestIntWrongType(t *testing.T) {
	ctx := newContext()
	ctx.Put("key", "a string")
	if _, err := ctx.Int("key"); err == nil {
		t.Log("Key type should be different")
		t.FailNow()
	}
}

func TestIntMissing(t *testing.T) {
	ctx := newContext()
	if _, err := ctx.Int("key"); err == nil {
		t.Log("Key should be missing")
		t.FailNow()
	}
}

func TestStrRightType(t *testing.T) {
	ctx := newContext()
	ctx.Put("key", "a string")

	val, err := ctx.Str("key")
	if err != nil {
		t.Log("Key should be present")
		t.FailNow()
	}
	if val != "a string" {
		t.Log("Context member changed its value somehow")
		t.FailNow()
	}
}

func TestStrWrongType(t *testing.T) {
	ctx := newContext()
	ctx.Put("key", 42)
	if _, err := ctx.Str("key"); err == nil {
		t.Log("Key type should be different")
		t.FailNow()
	}
}

func TestStrMissing(t *testing.T) {
	ctx := newContext()
	if _, err := ctx.Str("key"); err == nil {
		t.Log("Key should be missing")
		t.FailNow()
	}
}
