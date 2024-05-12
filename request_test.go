package simplehttp

import (
	"net/url"
	"testing"
)

func TestParseRequestLine(t *testing.T) {
	method, uri, version, err := parseRequestLine("GET /index.html HTTP/1.0")

	if err != nil {
		t.Fatalf("did not expect an error, but got %v", err)
	}

	if method != get {
		t.Fatalf("invalid method. Expected '%d' | Actual '%d'", get, method)
	}

	expectedUri, _ := url.ParseRequestURI("/index.html")
	if uri != *expectedUri {
		t.Fatalf("invalid method. Expected '%s' | Actual '%s'", expectedUri.String(), uri.String())
	}

	if version != "HTTP/1.0" {
		t.Fatalf("invalid HTTP version. Expected 'HTTP/1.0' | Actual '%s'", version)
	}
}

func TestParseRequestLine_InvalidLength(t *testing.T) {
	_, _, _, err := parseRequestLine("GET")
	if err == nil {
		t.Fatalf("expected an error, but it was nil")
	}

	_, _, _, err = parseRequestLine("GET /index.html")
	if err == nil {
		t.Fatalf("expected an error, but it was nil")
	}

	_, _, _, err = parseRequestLine("GET /index.html HTTP/1.0 Extra")
	if err == nil {
		t.Fatalf("expected an error, but it was nil")
	}
}

func TestParseRequestLine_InvalidMethod(t *testing.T) {
	_, _, _, err := parseRequestLine("FAKE /index.html HTTP/1.0")
	if err == nil {
		t.Fatalf("expected an error, but it was nil")
	}
}

func TestParseRequestLine_InvalidURI(t *testing.T) {
	_, _, _, err := parseRequestLine("GET - HTTP/1.0")
	if err == nil {
		t.Fatalf("expected an error, but it was nil")
	}

	_, _, _, err = parseRequestLine("GET no/leading/slash HTTP/1.0")
	if err == nil {
		t.Fatalf("expected an error, but it was nil")
	}
}

func TestParseRequest_IncompleteMessage_NoDoubleNewLine(t *testing.T) {
	raw := "GET /index.html HTTP/1.0"
	_, err := parseRequest(raw)
	if err == nil {
		t.Fatalf("expected an error from a request with no newlines, but it was nil")
	}

	raw = "GET /index.html HTTP/1.0" + lineEnd
	_, err = parseRequest(raw)
	if err == nil {
		t.Fatalf("expected an error from a request with one newline, but it was nil")
	}
}

func TestParseRequest_IncompleteMessage_ContentLengthNotMet(t *testing.T) {
	raw := "POST / HTTP/1.0" + lineEnd +
		"Content-Length: 13" + doubleLineEnd + // actual  content length is 12
		"hello world!"

	_, err := parseRequest(raw)
	if err == nil {
		t.Fatalf("expected an error from a request with not enough content, but it was nil")
	}
}

func TestParseRequest_InvalidRequestLine(t *testing.T) {
	raw := "GET / HTTP/1.0 extra-stuff"

	_, err := parseRequest(raw)
	if err == nil {
		t.Fatalf("expected an error from a request with an invalid request-line, but it was nil")
	}
}

func TestParseRequest_ParseWithNoBody(t *testing.T) {
	raw := "GET /index.html HTTP/1.0" + lineEnd +
		"Host: client:8080" + lineEnd +
		"Accept: */*" + doubleLineEnd

	request, err := parseRequest(raw)
	if err != nil {
		t.Fatalf("did not expect an error but received: %v", err)
	}

	if request.method != get {
		t.Fatalf("request's method was incorrect. Expected '%d' | Actual '%d'", get, request.method)
	}

	if request.uri.String() != "/index.html" {
		t.Fatalf("request's uri was incorrect. Expected '/index.html' | Actual '%s'", request.uri.String())
	}

	if request.httpVersion != "HTTP/1.0" {
		t.Fatalf("request's httpVersion was incorrect. Expected 'HTTP/1.0' | Actual '%s'", request.httpVersion)
	}

	if request.headers["Host"] != "client:8080" || request.headers["Accept"] != "*/*" {
		t.Fatalf("request's headers were incorrect")
	}

	if request.body != "" {
		t.Fatalf("request's body was incorrect. Expected '' | Actual '%s'", request.body)
	}
}

func TestParseRequest_ParseWithBody(t *testing.T) {
	body := "hello world!" + lineEnd +
		"hello again!"

	raw := "POST /api HTTP/1.0" + lineEnd +
		"Host: client:8080" + lineEnd +
		"Accept: */*" + lineEnd +
		"Content-Length: 26" + doubleLineEnd +
		body

	request, err := parseRequest(raw)
	if err != nil {
		t.Fatalf("did not expect an error but received: %v", err)
	}

	if request.method != post {
		t.Fatalf("request's method was incorrect. Expected '%d' | Actual '%d'", post, request.method)
	}

	if request.uri.String() != "/api" {
		t.Fatalf("request's uri was incorrect. Expected '/api' | Actual '%s'", request.uri.String())
	}

	if request.httpVersion != "HTTP/1.0" {
		t.Fatalf("request's httpVersion was incorrect. Expected 'HTTP/1.0' | Actual '%s'", request.httpVersion)
	}

	if request.headers["Host"] != "client:8080" &&
		request.headers["Accept"] != "*/*" &&
		request.headers["Content-Length"] != "26" {
		t.Fatalf("request's headers were incorrect")
	}

	if request.body != body {
		t.Fatalf("request's body was incorrect. Expected '%s' | Actual '%s'", body, request.body)
	}
}
