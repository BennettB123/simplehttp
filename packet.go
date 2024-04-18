package simplehttp

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const LineEnd string = "\r\n"
const DoubleLineEnd string = LineEnd + LineEnd

// Potential errors returned when dealing with packets
type incompletePacket struct {
	err string
}

func (e *incompletePacket) Error() string {
	return e.err
}

type invalidPacket struct {
	err string
}

func (e *invalidPacket) Error() string {
	return e.err
}

type packet struct {
	rawMessage string
	startLine  startline
	headers    map[string]string
	body       string
}

func (p packet) String() string {
	return p.rawMessage
}

// build string from the parsed packet components
func (p packet) buildString() string {
	return p.startLine.String() +
		LineEnd +
		p.headersToString() +
		DoubleLineEnd +
		p.body
}

func (p packet) headersToString() string {
	ret := ""
	for k, v := range p.headers {
		ret += fmt.Sprintf("%s: %s%s", k, v, LineEnd)
	}

	return ret
}

func parsePacket(rawPacket string) (packet, error) {
	headerEnd := strings.Index(rawPacket, DoubleLineEnd)
	if headerEnd == -1 {
		return packet{}, &incompletePacket{"a double line-end was not found"}
	}

	endOfFirstLine := strings.Index(rawPacket, LineEnd)
	rawStartLine := rawPacket[:endOfFirstLine]

	startLine, err := parseStartLine(rawStartLine)
	if err != nil {
		return packet{}, &invalidPacket{err.Error()}
	}

	rawHeaders := strings.TrimSpace(rawPacket[endOfFirstLine:headerEnd])
	headers, err := parseHeaders(rawHeaders)
	if err != nil {
		return packet{}, &invalidPacket{err.Error()}
	}

	contentLengthStr, exists := headers["Content-Length"]
	if !exists {
		return packet{
			rawMessage: rawPacket,
			startLine:  startLine,
			headers:    headers,
			body:       "",
		}, nil
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return packet{}, &invalidPacket{fmt.Sprintf("invalid value in Content-Length header: `%s`", contentLengthStr)}
	}

	content := rawPacket[headerEnd+len(DoubleLineEnd):]
	if len(content) < contentLength {
		return packet{}, &incompletePacket{fmt.Sprintf("expecting %d bytes in body, only received %d", contentLength, len(content))}
	}

	body := content[:contentLength]

	return packet{
		rawPacket,
		startLine,
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

type startline struct {
	verb        uint
	uri         url.URL
	httpversion string
}

func (s startline) String() string {
	return fmt.Sprintf("%s %s %s", getVerbString(s.verb), s.uri.String(), s.httpversion)
}

func (s startline) getPath() string {
	return s.uri.EscapedPath()
}

func parseStartLine(content string) (startline, error) {
	split := strings.Split(content, " ")
	if len(split) != 3 {
		return startline{}, fmt.Errorf("unable to parse HTTP start-line")
	}

	verb, err := parseHttpVerb(strings.TrimSpace(split[0]))
	if err != nil {
		return startline{}, err
	}

	uri, err := url.ParseRequestURI(strings.TrimSpace(split[1]))
	if err != nil {
		return startline{}, err
	}

	return startline{
		verb:        verb,
		uri:         *uri,
		httpversion: strings.TrimSpace(split[2]),
	}, nil
}
