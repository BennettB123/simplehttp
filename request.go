package simplehttp

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Request struct {
	rawMessage  string
	Method      uint
	Uri         url.URL
	HttpVersion string
	Headers     Headers
	Body        string
}

// build string from the parsed message components
func (r Request) String() string {
	requestLine := fmt.Sprintf("%s %s %s",
		getHttpMethodString(r.Method), r.Uri.String(), r.HttpVersion)

	return requestLine +
		lineEnd +
		r.Headers.String() +
		doubleLineEnd +
		r.Body
}

func (r Request) getPath() string {
	return r.Uri.EscapedPath()
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
