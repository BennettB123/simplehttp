package simplehttp

import (
	"fmt"
	"testing"
)

func dummyCallback(_ Request, _ *Response) error {
	return nil
}

func TestRegisterCallback(t *testing.T) {
	cbm := newCallbackMap()
	cbm.registerCallback(get, "/get1", dummyCallback)
	cbm.registerCallback(post, "/post1", dummyCallback)
	cbm.registerCallback(put, "/put1", dummyCallback)
	cbm.registerCallback(delete, "/delete1", dummyCallback)

	cbm.registerCallback(get, "/get2", dummyCallback)
	cbm.registerCallback(post, "/post2", dummyCallback)
	cbm.registerCallback(put, "/put2", dummyCallback)
	cbm.registerCallback(delete, "/delete2", dummyCallback)

	if len(cbm.callbacks[get]) != 2 {
		t.Fatalf("expected 2 entries in the 'get' map but found %d", len(cbm.callbacks[get]))
	}
	if len(cbm.callbacks[post]) != 2 {
		t.Fatalf("expected 2 entries in the 'post' map but found %d", len(cbm.callbacks[post]))
	}
	if len(cbm.callbacks[put]) != 2 {
		t.Fatalf("expected 2 entries in the 'put' map but found %d", len(cbm.callbacks[put]))
	}
	if len(cbm.callbacks[delete]) != 2 {
		t.Fatalf("expected 2 entries in the 'delete' map but found %d", len(cbm.callbacks[delete]))
	}
}

func TestDuplicateRegistrationError(t *testing.T) {
	cbm := newCallbackMap()
	cbm.registerCallback(get, "/get1", dummyCallback)
	err := cbm.registerCallback(get, "/get1", dummyCallback)

	if len(cbm.callbacks[get]) != 1 {
		t.Fatalf("expected 1 entries in the 'get' map but found %d", len(cbm.callbacks[get]))
	}

	if err == nil {
		t.Fatalf("did not receive an error from registering a duplicate callback")
	}
}

func TestInvokeCallback(t *testing.T) {
	invoked := false
	cbm := newCallbackMap()

	cbm.registerCallback(get, "/invokeMe", func(_ Request, _ *Response) error {
		invoked = true
		return nil
	})

	cbm.invokeCallback(get, "/invokeMe", Request{}, &Response{})

	if invoked != true {
		t.Fatalf("the callback was not invoked")
	}
}

func TestCallbackNotRegisteredError(t *testing.T) {
	cbm := newCallbackMap()
	err := cbm.invokeCallback(get, "/404", Request{}, &Response{})

	if err == nil {
		t.Fatalf("did not receive an error from invoking an unregistered callback")
	}
}

func TestCallbackRuntimeError(t *testing.T) {
	cbm := newCallbackMap()
	cbm.registerCallback(get, "/invokeMe", func(_ Request, _ *Response) error {
		return fmt.Errorf("I'm a runtime error!")
	})

	err := cbm.invokeCallback(get, "/invokeMe", Request{}, &Response{})

	if err == nil {
		t.Fatalf("did not receive an error from invoking a callback that returned an error")
	}
}
