package simplehttp

import (
	"strconv"
	"testing"
)

func TestResponse_SetHeader(t *testing.T) {
	res := newResponse()

	key := "MyNewHeader-Key"
	val := "MyNewHeader-Value"
	err := res.SetHeader(key, val)

	if err != nil {
		t.Fatalf("did not expect an error but received: %v", err)
	}

	if res.headers[key] != val {
		t.Fatalf("the added header value was was correct. Expected '%s' | Actual '%s'",
			val, res.headers[key])
	}
}

func TestResponse_SetHeader_ErrorIfColonInKey(t *testing.T) {
	res := newResponse()

	key := "MyNewHeader:Key"
	val := "MyNewHeader-Value"
	err := res.SetHeader(key, val)

	if err == nil {
		t.Fatalf("expected an error, but it was nil")
	}
}

func TestResponse_SetHtml(t *testing.T) {
	body := "<h1>Hello World!</h1>"
	expectedLength := strconv.Itoa(len(body))

	res := newResponse()
	res.SetHtml(body)

	if res.body != body {
		t.Fatalf("request had an incorrect body. Expected '%s' | Actual '%s'", body, res.body)
	}

	if res.headers["Content-Length"] != expectedLength {
		t.Fatalf("request had an Content-Length. Expected '%s' | Actual '%s'",
			expectedLength, res.headers["Content-Length"])
	}

	if res.headers["Content-Type"] != "text/html" {
		t.Fatalf("request had an incorrect body. Expected 'text/html' | Actual '%s'",
			res.headers["Content-Type"])
	}
}

func TestResponse_SetJson(t *testing.T) {
	body := `{
		foo: foo,
		bar: bar
	}`
	expectedLength := strconv.Itoa(len(body))

	res := newResponse()
	res.SetJson(body)

	if res.body != body {
		t.Fatalf("request had an incorrect body. Expected '%s' | Actual '%s'", body, res.body)
	}

	if res.headers["Content-Length"] != expectedLength {
		t.Fatalf("request had an Content-Length. Expected '%s' | Actual '%s'",
			expectedLength, res.headers["Content-Length"])
	}

	if res.headers["Content-Type"] != "application/json" {
		t.Fatalf("request had an incorrect body. Expected 'text/html' | Actual '%s'",
			res.headers["Content-Type"])
	}
}
