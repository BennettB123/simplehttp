package simplehttp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	httpVersion  string
	StatusCode   uint
	ReasonPhrase string
	Headers      Headers
	body         string // private so we can ensure Content-Length is set correctly
}

func (r Response) String() string {
	statusLine := fmt.Sprintf("%s %d %s", r.httpVersion, r.StatusCode, r.ReasonPhrase)
	return statusLine +
		lineEnd +
		r.Headers.String() +
		doubleLineEnd +
		r.body
}

func newResponse() Response {
	headers := make(map[string]string)
	headers["Date"] = formatHttpDate(time.Now())
	headers["Server"] = "simplehttp"
	headers["Content-Length"] = "0"
	headers["Connection"] = "close"

	return Response{
		httpVersion:  "HTTP/1.0",
		StatusCode:   200,
		ReasonPhrase: getReasonPhrase(200),
		Headers:      headers,
		body:         "",
	}
}

func (r *Response) SetHeader(key string, value string) error {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	colonInKey := strings.Index(key, ":")
	colonInValue := strings.Index(value, ":")

	if colonInKey > -1 || colonInValue > -1 {
		return fmt.Errorf("header field or value cannot contain a colon")
	}

	key = url.QueryEscape(key)
	value = url.QueryEscape(value)

	r.Headers[key] = value
	return nil
}

func (r *Response) SendHtml(html string) {
	r.body = html
	r.Headers["Content-Length"] = strconv.Itoa(len(html))
	r.Headers["Content-Type"] = "text/html"
}

func (r *Response) SendJson(obj any) error {
	body := ""

	// TODO: find a better way to check if it's already a string
	objType := fmt.Sprintf("%T", obj)
	if objType == "string" {
		body = fmt.Sprintf("%s", obj)
	} else {
		marshalled, err := json.Marshal(obj)
		if err != nil {
			return fmt.Errorf("error while marshalling object: %s", err)
		}
		body = string(marshalled)
	}

	r.body = body
	r.Headers["Content-Length"] = strconv.Itoa(len(body))
	r.Headers["Content-Type"] = "application/json"
	return nil
}

func (r *Response) SetStatus(status uint) {
	r.StatusCode = status
	r.ReasonPhrase = getReasonPhrase(status)
}
