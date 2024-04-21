package simplehttp

import (
	"fmt"
)

const (
	get    = iota
	post   = iota
	put    = iota
	delete = iota
)

func parseHttpMethod(method string) (uint, error) {
	switch method {
	case "GET":
		return get, nil
	case "POST":
		return post, nil
	case "PUT":
		return put, nil
	case "DELETE":
		return delete, nil
	default:
		return 0, fmt.Errorf("unsupported HTTP method")
	}
}

func getHttpMethodString(method uint) string {
	switch method {
	case get:
		return "GET"
	case post:
		return "POST"
	case put:
		return "PUT"
	case delete:
		return "DELETE"
	default:
		return ""
	}
}
