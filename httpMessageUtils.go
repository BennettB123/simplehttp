package simplehttp

import (
	"fmt"
	"strings"
	"time"
)

const lineEnd string = "\r\n"
const doubleLineEnd string = lineEnd + lineEnd

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

type Headers map[string]string

func (h Headers) String() string {
	// TODO: HTTP/1.0 RFC says "it is 'good practice' to send General-Header fields first,
	//   followed by Request-Header or Response-Header fields prior to the Entity-Header fields"
	ret := ""
	headerCount := len(h)
	i := 1
	for k, v := range h {
		if i == headerCount {
			ret += fmt.Sprintf("%s: %s", k, v)
		} else {
			ret += fmt.Sprintf("%s: %s%s", k, v, lineEnd)
		}
		i++
	}

	return ret
}

func parseHeaders(message string) (map[string]string, error) {
	headers := make(map[string]string)
	lines := strings.Split(message, lineEnd)

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

func formatHttpDate(time time.Time) string {
	// HTTP/1.0 recommends IMF-fixdate format
	//   ex: Sun, 06 Nov 1994 08:49:37 GMT
	time = time.UTC()
	dayName := time.Weekday().String()[:3]
	dateAndTime := time.Format("02 Jan 2006 15:04:05")

	return fmt.Sprintf("%s, %s GMT", dayName, dateAndTime)
}

func getReasonPhrase(status uint) string {
	switch status {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 202:
		return "Accepted"
	case 204:
		return "No Content"
	case 300:
		return "Multiple Choices"
	case 301:
		return "Moved Permanently"
	case 302:
		return "Moved Temporarily"
	case 304:
		return "Not Modified"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	case 501:
		return "Not Implemented"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	default:
		return "unknown"
	}
}
