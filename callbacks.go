package simplehttp

import "fmt"

// Potential errors when dealing with callbacks
type callbackRuntimeError struct {
	innerErr   error
	HttpMethod uint
	Path       string
}

func (err callbackRuntimeError) Error() string {
	return fmt.Sprintf("an error occurred while invoking a callback (%s with path '%s'): %s",
		getHttpMethodString(err.HttpMethod), err.Path, err.innerErr.Error())
}

func newCallbackRuntimeError(err error) error {
	return callbackRuntimeError{
		innerErr: err,
	}
}

type callbackAlreadyRegisteredError struct {
	HttpMethod uint
	Path       string
}

func (err callbackAlreadyRegisteredError) Error() string {
	return fmt.Sprintf("%s callback with path '%s' has already been registered",
		getHttpMethodString(err.HttpMethod), err.Path)
}

func newCallbackAlreadyRegisteredError(httpMethod uint, path string) error {
	return callbackAlreadyRegisteredError{
		httpMethod,
		path,
	}
}

type callbackNotRegisteredError struct {
	httpMethod uint
	Path       string
}

func (err callbackNotRegisteredError) Error() string {
	return fmt.Sprintf("%s callback with path '%s' has not been registered",
		getHttpMethodString(err.httpMethod), err.Path)
}

func newCallbackNotRegisteredError(httpMethod uint, path string) error {
	return callbackNotRegisteredError{
		httpMethod,
		path,
	}
}

type callbackFunc = func(Request, *Response) error

type callbackMap struct {
	callbacks map[uint]map[string]callbackFunc
	// ex. [GET]["/"] = func(...)
	// ex. [GET]["/login"] = func(...)
	// ex. [POST]["/login"] = func(...)
}

func newCallbackMap() callbackMap {
	callbacks := make(map[uint]map[string]callbackFunc)
	callbacks[get] = make(map[string]callbackFunc)
	callbacks[post] = make(map[string]callbackFunc)
	callbacks[put] = make(map[string]callbackFunc)
	callbacks[delete] = make(map[string]callbackFunc)
	return callbackMap{
		callbacks,
	}
}

func (cbm *callbackMap) registerCallback(method uint, path string, callback callbackFunc) error {
	_, exists := cbm.callbacks[method][path]
	if exists {
		return newCallbackAlreadyRegisteredError(method, path)
	}

	cbm.callbacks[method][path] = callback
	return nil
}

func (cbm *callbackMap) invokeCallback(method uint, path string, req Request, res *Response) error {
	callback, exists := cbm.callbacks[method][path]
	if !exists {
		return newCallbackNotRegisteredError(method, path)
	}

	err := callback(req, res)
	if err != nil {
		return newCallbackRuntimeError(err)
	}

	return nil
}
