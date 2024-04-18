package simplehttp

import (
	"fmt"
)

const (
	GET    = iota
	POST   = iota
	PUT    = iota
	DELETE = iota
)

func parseHttpVerb(verb string) (uint, error) {
	switch verb {
	case "GET":
		return GET, nil
	case "POST":
		return POST, nil
	case "PUT":
		return PUT, nil
	case "DELETE":
		return DELETE, nil
	default:
		return 0, fmt.Errorf("unsupported HTTP verb")
	}
}

func getVerbString(verb uint) string {
	switch verb {
	case GET:
		return "GET"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case DELETE:
		return "DELETE"
	default:
		return ""
	}
}
