package simplehttp

import (
	"fmt"
	"strings"
)

const LineEnd string = "\r\n"
const DoubleLineEnd string = LineEnd + LineEnd

// Potential errors returned when dealing with HTTP messages
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
	rawMessage string
	headers    map[string]string
	body       string
}

func (p httpMessage) String() string {
	return p.rawMessage
}

func (p httpMessage) headersToString() string {
	ret := ""
	for k, v := range p.headers {
		ret += fmt.Sprintf("%s: %s%s", k, v, LineEnd)
	}

	return ret
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
