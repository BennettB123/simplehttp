package simplehttp

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	httpVersion  string
	statusCode   uint
	reasonPhrase string
	headers      headers
	body         string
}

func (r Response) String() string {
	statusLine := fmt.Sprintf("%s %d %s", r.httpVersion, r.statusCode, r.reasonPhrase)
	return statusLine +
		lineEnd +
		r.headers.String() +
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
		statusCode:   200,
		reasonPhrase: getReasonPhrase(200),
		headers:      headers,
		body:         "",
	}
}

func new500StatusResponse() Response {
	res := newResponse()
	res.SetStatus(500)
	return res
}

func (r Response) Headers() headers {
	return r.headers
}

func (r Response) StatusCode() uint {
	return r.statusCode
}

func (r Response) ReasonPhrase() string {
	return r.reasonPhrase
}

func (r Response) Body() string {
	return r.body
}

func (r *Response) SetHeader(key string, value string) error {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	colonInKey := strings.Index(key, ":")

	if colonInKey > -1 {
		return fmt.Errorf("header key cannot contain a colon")
	}

	key = url.QueryEscape(key)
	value = url.QueryEscape(value)

	r.headers[key] = value
	return nil
}

func (r *Response) SetHtml(html string) {
	r.body = html
	r.headers["Content-Length"] = strconv.Itoa(len(html))
	r.headers["Content-Type"] = "text/html"
}

func (r *Response) SetJson(obj any) error {
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
	r.headers["Content-Length"] = strconv.Itoa(len(body))
	r.headers["Content-Type"] = "application/json"
	return nil
}

func (r *Response) SetFile(path string) error {
	fileContents, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	extension := filepath.Ext(path)
	if extension == "" {
		return fmt.Errorf("unable to determine a Content-Type because the file does not have an extension")
	}

	contentType := mime.TypeByExtension(extension)
	if contentType == "" {
		return fmt.Errorf("unable to determine a Content-Type based on the file's extension")
	}

	body := string(fileContents)
	r.body = body
	r.headers["Content-Length"] = strconv.Itoa(len(body))
	r.headers["Content-Type"] = contentType
	return nil
}

func (r *Response) SetFileWithContentType(path string, contentType string) error {
	fileContents, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	body := string(fileContents)
	r.body = body
	r.headers["Content-Length"] = strconv.Itoa(len(body))
	r.headers["Content-Type"] = contentType
	return nil
}

func (r *Response) SetStatus(status uint) {
	r.statusCode = status
	r.reasonPhrase = getReasonPhrase(status)
}
