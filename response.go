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

// Response represents an HTTP reponse to be returned to a client.
// It provides several getter methods to view its properties as well as
// several methods to edit the response before it is sent.
type Response struct {
	httpVersion  string
	statusCode   uint
	reasonPhrase string
	headers      headers
	body         string
}

// Builds a string that represents the entire HTTP response.
func (r Response) String() string {
	statusLine := fmt.Sprintf("%s %d %s", r.httpVersion, r.statusCode, r.reasonPhrase)

	return statusLine + lineEnd +
		r.headers.String() + doubleLineEnd +
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

// Returns the response's headers.
func (r Response) Headers() map[string]string {
	return r.headers
}

// Returns the response's Status-Code.
func (r Response) StatusCode() uint {
	return r.statusCode
}

// Returns the response's Reason-Phrase.
func (r Response) ReasonPhrase() string {
	return r.reasonPhrase
}

// Returns the response's body.
func (r Response) Body() string {
	return r.body
}

// Adds a single header to the Response.
// key is the header-field to be added, and value is the
// value of the new header. An error will be returned if the key
// contains a colon.
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

// Sets the Response's body to the provided html string.
// This method will also set the Content-Length header to the length
// of the provided input. The Content-Type header will be set to "text/html".
func (r *Response) SetHtml(html string) {
	r.body = html
	r.headers["Content-Length"] = strconv.Itoa(len(html))
	r.headers["Content-Type"] = "text/html"
}

// Sets the Response's body to a JSON string. If obj is a string,
// the body will be set to the provided string. If obj is any other type,
// it will be marshalled to JSON using [json.Marshal].
// This method will set the Content-Length header appropriately as well
// as setting the Content-Type header to "application/json".
// An error will be returned if there was an issue marshalling the obj.
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

// Sets the Response's body to the content in the file provided
// by the path parameter. The path can be either absolute or relative
// to the current working directory. The Content-Length header will be set
// appropriately. The Content-Type header will be determined by the file's
// extension using the [mime.TypeByExtension] function.
// Returns an error if the file could not be read, or if a Content-Type
// was unable to be determined.
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

// Sets the Response's body to the content in the file provided
// by the path parameter. The path can be either absolute or relative
// to the current working directory. Sets the Content-Type header to the value
// provided by the contentType parameter. The Content-Length header will be set
// appropriately. Returns an error if the file could not be read.
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

// Sets the Response's Status-Code to the value provided in the
// status parameter. A Reason-Phrase will also be set based on the
// status code. See [RFC 1945] Section 9 for a complete
// list of Status-Codes and Reason-Phrases. If a Status-Code
// is provided that does not appear in this list, the provided code will
// be set, however the Reason-Phrase will be 'unknown'.
//
// [RFC 1945]: https://www.rfc-editor.org/rfc/rfc1945.html#section-9
func (r *Response) SetStatus(status uint) {
	r.statusCode = status
	r.reasonPhrase = getReasonPhrase(status)
}
