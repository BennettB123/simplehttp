package simplehttp

import (
	"fmt"
	"strconv"
	"strings"
)

const LineEnd string = "\r\n"
const DoubleLineEnd string = LineEnd + LineEnd

type packet struct {
	rawMessage string
	startLine  string
	headers    map[string]string
	body       string
}

func (p packet) String() string {
	return p.rawMessage
}

// build string from the parsed packet components
func (p packet) buildString() string {
	return p.startLine +
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
		return packet{}, nil
	}

	endOfFirstLine := strings.Index(rawPacket, LineEnd)
	startLine := rawPacket[:endOfFirstLine]

	rawHeaders := strings.TrimSpace(rawPacket[endOfFirstLine:headerEnd])
	headers, err := parseHeaders(rawHeaders)
	if err != nil {
		return packet{}, err
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
		return packet{}, fmt.Errorf("invalid value in Content-Length header: `%s`", contentLengthStr)
	}

	content := rawPacket[headerEnd+len(DoubleLineEnd):]
	if len(content) < contentLength {
		return packet{}, fmt.Errorf("expecting %d bytes in body, only received %d", contentLength, len(content))
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
