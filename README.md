# simplehttp
A bare-bones HTTP/1.0 web framework for go.

Note: this package should not be used in a production environment. It
was purely created as a learning opportunity to gain experience with go.

## Documentation
Full documentation can be found on go's official documentation site at
https://pkg.go.dev/github.com/BennettB123/simplehttp.

### Features
✅ custom routing in the form of registering callback methods to be invoked when specific HTTP methods/paths are requested. <br>
✅ handling multiple concurrent requests in parallel. <br>
✅ methods to view information about incoming HTTP requests. <br>
✅ convienent methods to modify HTTP responses. <br>
✅ custom logger interface to receive messages about connections, incoming requests, and outgoing responses. <br>

## Basic Example
The following program will create a Server listening on port 8080. It will respond to incoming GET requests
to the '/hello-world' path.

```go
func main() {
	server := simplehttp.NewServer(8080)

	server.Get("/hello-world", func(req simplehttp.Request, res *simplehttp.Response) error {
		res.SetHtml("<h1>Hello, world!</h1>")
		return nil
	})

	err := server.Start()
	if err != nil {
		fmt.Println("There was an error starting the server:", err)
	}
}
```

See cmd/basicserver/basicserver.go for more examples.
