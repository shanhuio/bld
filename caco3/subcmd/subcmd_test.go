package subcmd

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestRunNoCommand(t *testing.T) {
	lst := New()
	lst.Add("foo", "", func([]string) error { return nil })
	if got := lst.Run([]string{"prog"}); got != -1 {
		t.Errorf("Run with no command: got %d, want -1", got)
	}
}

func TestRunHelp(t *testing.T) {
	lst := New()
	lst.Add("foo", "", func([]string) error {
		t.Fatal("foo should not be invoked for help")
		return nil
	})
	for _, name := range []string{"-h", "help"} {
		if got := lst.Run([]string{"prog", name}); got != 0 {
			t.Errorf("Run %q: got %d, want 0", name, got)
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	lst := New()
	if got := lst.Run([]string{"prog", "nope"}); got != -1 {
		t.Errorf("unknown command: got %d, want -1", got)
	}
}

func TestRunNilFunc(t *testing.T) {
	lst := New()
	lst.Add("noop", "", nil)
	if got := lst.Run([]string{"prog", "noop"}); got != 0 {
		t.Errorf("nil-func command: got %d, want 0", got)
	}
}

func TestRunHandlerSuccess(t *testing.T) {
	lst := New()
	var gotArgs []string
	lst.Add("foo", "", func(args []string) error {
		gotArgs = args
		return nil
	})
	if got := lst.Run([]string{"prog", "foo", "a", "b"}); got != 0 {
		t.Errorf("Run: got %d, want 0", got)
	}
	wantArgs := []string{"a", "b"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Errorf("handler args: got %v, want %v", gotArgs, wantArgs)
	}
}

func TestRunHandlerError(t *testing.T) {
	lst := New()
	lst.Add("foo", "", func([]string) error { return errors.New("boom") })
	if got := lst.Run([]string{"prog", "foo"}); got != -1 {
		t.Errorf("erroring handler: got %d, want -1", got)
	}
}

func TestAddDuplicatePanics(t *testing.T) {
	lst := New()
	lst.Add("foo", "", nil)
	defer func() {
		if recover() == nil {
			t.Error("adding duplicate command did not panic")
		}
	}()
	lst.Add("foo", "", nil)
}

func TestHelpSortedOutput(t *testing.T) {
	lst := New()
	lst.Add("zoo", "z desc", nil)
	lst.Add("apple", "a desc", nil)
	lst.Add("mango", "m desc", nil)

	var buf bytes.Buffer
	lst.Help(&buf)

	want := "apple - a desc\nmango - m desc\nzoo - z desc\n"
	if got := buf.String(); got != want {
		t.Errorf("Help output:\ngot:\n%s\nwant:\n%s", got, want)
	}
}
