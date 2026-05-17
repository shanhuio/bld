package caco3

import (
	"reflect"
	"testing"
)

func TestLoadTracerEmpty(t *testing.T) {
	tr := newLoadTracer()
	if got := tr.stack(); len(got) != 0 {
		t.Errorf("new tracer stack: got %v, want empty", got)
	}
	tr.pop() // pop on empty should be a no-op
	if got := tr.stack(); len(got) != 0 {
		t.Errorf("after pop on empty: got %v, want empty", got)
	}
}

func TestLoadTracerPushPop(t *testing.T) {
	tr := newLoadTracer()
	if !tr.push("a") {
		t.Error(`push("a"): got false, want true`)
	}
	if !tr.push("b") {
		t.Error(`push("b"): got false, want true`)
	}
	if want := []string{"a", "b"}; !reflect.DeepEqual(tr.stack(), want) {
		t.Errorf("stack: got %v, want %v", tr.stack(), want)
	}

	// Re-pushing a name on the stack must fail.
	if tr.push("a") {
		t.Error(`push("a") again: got true, want false`)
	}
	if tr.push("b") {
		t.Error(`push("b") again: got true, want false`)
	}

	tr.pop()
	if want := []string{"a"}; !reflect.DeepEqual(tr.stack(), want) {
		t.Errorf("after one pop: got %v, want %v", tr.stack(), want)
	}

	// "b" is off the stack now, so it can be pushed again.
	if !tr.push("b") {
		t.Error(`push("b") after pop: got false, want true`)
	}
	if want := []string{"a", "b"}; !reflect.DeepEqual(tr.stack(), want) {
		t.Errorf("after re-push: got %v, want %v", tr.stack(), want)
	}

	tr.pop()
	tr.pop()
	if got := tr.stack(); len(got) != 0 {
		t.Errorf("after popping everything: got %v, want empty", got)
	}

	// All names are off the stack; both should be pushable again.
	if !tr.push("a") {
		t.Error(`push("a") after full unwind: got false, want true`)
	}
}
