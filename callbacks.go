package simplehttp

import "fmt"

// Potential errors when dealing with callbacks
type CallbackRuntimeError struct {
	innerErr error
	HttpVerb uint
	Path     string
}

func (err CallbackRuntimeError) Error() string {
	return fmt.Sprintf("an error occurred while invoking a callback (%s with path '%s'): %s",
		getVerbString(err.HttpVerb), err.Path, err.innerErr.Error())
}

func newCallbackRuntimeError(err error) error {
	return CallbackRuntimeError{
		innerErr: err,
	}
}

type CallbackAlreadyRegisteredError struct {
	HttpVerb uint
	Path     string
}

func (err CallbackAlreadyRegisteredError) Error() string {
	return fmt.Sprintf("%s callback with path '%s' has already been registered",
		getVerbString(err.HttpVerb), err.Path)
}

func newCallbackAlreadyRegisteredError(httpVerb uint, path string) error {
	return CallbackAlreadyRegisteredError{
		httpVerb,
		path,
	}
}

type CallbackNotRegisteredError struct {
	HttpVerb uint
	Path     string
}

func (err CallbackNotRegisteredError) Error() string {
	return fmt.Sprintf("%s callback with path '%s' has not been registered",
		getVerbString(err.HttpVerb), err.Path)
}

func newCallbackNotRegisteredError(httpVerb uint, path string) error {
	return CallbackNotRegisteredError{
		httpVerb,
		path,
	}
}

type CallbackFunc = func() error

type callbackMap struct {
	callbacks map[uint]map[string]CallbackFunc
	// ex. [GET]["/"] = func(...)
	// ex. [GET]["/login"] = func(...)
	// ex. [POST]["/login"] = func(...)
}

func createCallbackMap() callbackMap {
	callbacks := make(map[uint]map[string]CallbackFunc)
	callbacks[GET] = make(map[string]CallbackFunc)
	callbacks[POST] = make(map[string]CallbackFunc)
	callbacks[PUT] = make(map[string]CallbackFunc)
	callbacks[DELETE] = make(map[string]CallbackFunc)
	return callbackMap{
		callbacks,
	}
}

func (cbm *callbackMap) registerCallback(method uint, path string, callback CallbackFunc) error {
	_, exists := cbm.callbacks[method][path]
	if exists {
		return newCallbackAlreadyRegisteredError(method, path)
	}

	cbm.callbacks[method][path] = callback
	return nil
}

func (cbm *callbackMap) invokeCallback(method uint, path string) error {
	callback, exists := cbm.callbacks[method][path]
	if !exists {
		return newCallbackNotRegisteredError(method, path)
	}

	err := callback()
	if err != nil {
		return newCallbackRuntimeError(err)
	}

	return nil
}
