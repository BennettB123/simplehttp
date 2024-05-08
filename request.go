package simplehttp

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Request struct {
	rawMessage  string
	method      uint
	uri         url.URL
	httpVersion string
	headers     headers
	body        string
}

// build string from the parsed message components
func (r Request) String() string {
	requestLine := fmt.Sprintf("%s %s %s",
		getHttpMethodString(r.method), r.uri.String(), r.httpVersion)

	return requestLine +
		lineEnd +
		r.headers.String() +
		doubleLineEnd +
		r.body
}

func (r Request) RawMessage() string {
	return r.rawMessage
}

func (r Request) Uri() string {
	return r.uri.String()
}

func (r Request) Path() string {
	return r.uri.EscapedPath()
}

func (r Request) Method() string {
	return getHttpMethodString(r.method)
}

func (r Request) Headers() map[string]string {
	return r.headers
}

func (r Request) Body() string {
	return r.body
}

func (r Request) Parameters() map[string][]string {
	return r.uri.Query()
}

func (r Request) RawParameters() string {
	return r.uri.RawQuery
}

func parseRequest(rawMessage string) (Request, error) {
	headerEnd := strings.Index(rawMessage, doubleLineEnd)
	if headerEnd == -1 {
		return Request{}, &incompleteMessage{"a double line-end was not found"}
	}

	endOfFirstLine := strings.Index(rawMessage, lineEnd)
	rawRequestLine := rawMessage[:endOfFirstLine]

	method, uri, httpVersion, err := parseRequestLine(rawRequestLine)
	if err != nil {
		return Request{}, &invalidMessage{err.Error()}
	}

	rawHeaders := strings.TrimSpace(rawMessage[endOfFirstLine:headerEnd])
	headers, err := parseHeaders(rawHeaders)
	if err != nil {
		return Request{}, &invalidMessage{err.Error()}
	}

	contentLengthStr, exists := headers["Content-Length"]
	if !exists {
		return Request{
			rawMessage,
			method,
			uri,
			httpVersion,
			headers,
			"",
		}, nil
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return Request{}, &invalidMessage{fmt.Sprintf(
			"invalid value in Content-Length header: `%s`", contentLengthStr)}
	}

	content := rawMessage[headerEnd+len(doubleLineEnd):]
	if len(content) < contentLength {
		return Request{}, &incompleteMessage{fmt.Sprintf(
			"expecting %d bytes in body, only received %d", contentLength, len(content))}
	}

	body := content[:contentLength]

	return Request{
		rawMessage,
		method,
		uri,
		httpVersion,
		headers,
		body,
	}, nil
}

// returns method, URI, and HttpVersion
func parseRequestLine(content string) (uint, url.URL, string, error) {
	split := strings.Split(content, " ")
	if len(split) != 3 {
		return 0, url.URL{}, "", fmt.Errorf("unable to parse HTTP request-line")
	}

	method, err := parseHttpMethod(strings.TrimSpace(split[0]))
	if err != nil {
		return 0, url.URL{}, "", err
	}

	uri, err := url.ParseRequestURI(strings.TrimSpace(split[1]))
	if err != nil {
		return 0, url.URL{}, "", err
	}

	httpVersion := strings.TrimSpace(split[2])

	return method, *uri, httpVersion, nil
}
