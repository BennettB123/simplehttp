package simplehttp

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const LineEnd string = "\r\n"
const DoubleLineEnd string = LineEnd + LineEnd

// Potential errors returned when dealing with http messages
type incompleteMessage struct {
	err string
}

func (e *incompleteMessage) Error() string {
	return e.err
}

type invalidMessage struct {
	err string
}

func (e *invalidMessage) Error() string {
	return e.err
}

type httpMessage struct {
	rawMessage  string
	requestLine requestLine
	headers     map[string]string
	body        string
}

func (p httpMessage) String() string {
	return p.rawMessage
}

// build string from the parsed message components
func (p httpMessage) buildString() string {
	return p.requestLine.String() +
		LineEnd +
		p.headersToString() +
		DoubleLineEnd +
		p.body
}

func (p httpMessage) headersToString() string {
	ret := ""
	for k, v := range p.headers {
		ret += fmt.Sprintf("%s: %s%s", k, v, LineEnd)
	}

	return ret
}

func parseHttpMessage(rawMessage string) (httpMessage, error) {
	headerEnd := strings.Index(rawMessage, DoubleLineEnd)
	if headerEnd == -1 {
		return httpMessage{}, &incompleteMessage{"a double line-end was not found"}
	}

	endOfFirstLine := strings.Index(rawMessage, LineEnd)
	rawRequestLine := rawMessage[:endOfFirstLine]

	requestLine, err := parseRequestLine(rawRequestLine)
	if err != nil {
		return httpMessage{}, &invalidMessage{err.Error()}
	}

	rawHeaders := strings.TrimSpace(rawMessage[endOfFirstLine:headerEnd])
	headers, err := parseHeaders(rawHeaders)
	if err != nil {
		return httpMessage{}, &invalidMessage{err.Error()}
	}

	contentLengthStr, exists := headers["Content-Length"]
	if !exists {
		return httpMessage{
			rawMessage:  rawMessage,
			requestLine: requestLine,
			headers:     headers,
			body:        "",
		}, nil
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return httpMessage{}, &invalidMessage{fmt.Sprintf("invalid value in Content-Length header: `%s`", contentLengthStr)}
	}

	content := rawMessage[headerEnd+len(DoubleLineEnd):]
	if len(content) < contentLength {
		return httpMessage{}, &incompleteMessage{fmt.Sprintf("expecting %d bytes in body, only received %d", contentLength, len(content))}
	}

	body := content[:contentLength]

	return httpMessage{
		rawMessage,
		requestLine,
		headers,
		body,
	}, nil
}

func parseHeaders(message string) (map[string]string, error) {
	headers := make(map[string]string)
	lines := strings.Split(message, LineEnd)

	for _, line := range lines {
		split := strings.SplitN(line, ":", 2)

		if len(split) != 2 {
			return nil, fmt.Errorf("could not parse the following header: `%s`", line)
		}

		field := strings.TrimSpace(split[0])
		value := strings.TrimSpace(split[1])
		headers[field] = value
	}

	return headers, nil
}

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
