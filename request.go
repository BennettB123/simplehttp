package simplehttp

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type request struct {
	requestLine requestLine
	httpMessage
}

// build string from the parsed message components
func (r request) buildString() string {
	return r.requestLine.String() +
		LineEnd +
		r.headersToString() +
		DoubleLineEnd +
		r.body
}

func parseRequest(rawMessage string) (request, error) {
	headerEnd := strings.Index(rawMessage, DoubleLineEnd)
	if headerEnd == -1 {
		return request{}, &incompleteMessage{"a double line-end was not found"}
	}

	endOfFirstLine := strings.Index(rawMessage, LineEnd)
	rawRequestLine := rawMessage[:endOfFirstLine]

	requestLine, err := parseRequestLine(rawRequestLine)
	if err != nil {
		return request{}, &invalidMessage{err.Error()}
	}

	rawHeaders := strings.TrimSpace(rawMessage[endOfFirstLine:headerEnd])
	headers, err := parseHeaders(rawHeaders)
	if err != nil {
		return request{}, &invalidMessage{err.Error()}
	}

	contentLengthStr, exists := headers["Content-Length"]
	if !exists {
		return request{
			requestLine,
			httpMessage{
				rawMessage,
				headers,
				"",
			},
		}, nil
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return request{}, &invalidMessage{fmt.Sprintf("invalid value in Content-Length header: `%s`", contentLengthStr)}
	}

	content := rawMessage[headerEnd+len(DoubleLineEnd):]
	if len(content) < contentLength {
		return request{}, &incompleteMessage{fmt.Sprintf("expecting %d bytes in body, only received %d", contentLength, len(content))}
	}

	body := content[:contentLength]

	return request{
		requestLine,
		httpMessage{
			rawMessage,
			headers,
			body,
		},
	}, nil
}

////// requestLine ///////

type requestLine struct {
	verb        uint
	uri         url.URL
	httpversion string
}

func (s requestLine) String() string {
	return fmt.Sprintf("%s %s %s", getVerbString(s.verb), s.uri.String(), s.httpversion)
}

func (s requestLine) getPath() string {
	return s.uri.EscapedPath()
}

func parseRequestLine(content string) (requestLine, error) {
	split := strings.Split(content, " ")
	if len(split) != 3 {
		return requestLine{}, fmt.Errorf("unable to parse HTTP request-line")
	}

	verb, err := parseHttpVerb(strings.TrimSpace(split[0]))
	if err != nil {
		return requestLine{}, err
	}

	uri, err := url.ParseRequestURI(strings.TrimSpace(split[1]))
	if err != nil {
		return requestLine{}, err
	}

	return requestLine{
		verb:        verb,
		uri:         *uri,
		httpversion: strings.TrimSpace(split[2]),
	}, nil
}
