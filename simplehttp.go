package simplehttp

import (
	"errors"
	"fmt"
	"io"
	"net"
)

type Server struct {
	Port        uint16
	callbackMap callbackMap
}

func NewServer(port uint16) Server {
	return Server{
		Port:        port,
		callbackMap: createCallbackMap(),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return fmt.Errorf("failed to open tcp listener: %s", err)
	}

	fmt.Println("Listening on port", s.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept incoming connection: ", err)
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	fmt.Println("============= Talking to", conn.RemoteAddr(), "=============")
	defer conn.Close()

	request, err := readRequest(conn)
	if err != nil {
		fmt.Print("Unable to read message from the connection: ", err)
		fmt.Print("\n=====================================================\n\n")
		return
	}

	fmt.Print(request.buildString())

	// call end-user's callback
	method := request.requestLine.method
	path := request.requestLine.getPath()
	err = s.callbackMap.invokeCallback(method, path)
	if err != nil {
		fmt.Println(err)
	}

	// send a response
	conn.Write([]byte("HTTP/1.0 200 OK\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Length: 237\r\n" +
		"\r\n" +
		"<!DOCTYPE html>\r\n" +
		"<html lang=\"en\">\r\n" +
		"<head>\r\n" +
		"  <meta charset=\"UTF-8\" />\r\n" +
		"  <title>Hello, world!</title>\r\n" +
		"  <meta name=\"viewport\" content=\"width=device-width,initial-scale=1\" />\r\n" +
		"</head>\r\n" +
		"<body>\r\n" +
		"  <h1>Hello, world!</h1>\r\n" +
		"</body>\r\n" +
		"</html>\r\n"))

	fmt.Print("\n=====================================================\n\n")
}

func readRequest(conn net.Conn) (request, error) {
	// read data in chunks of 1kB
	tmp := make([]byte, 1024)
	data := make([]byte, 0)
	message := request{}
	length := 0

	for {
		// TODO: add timeout here and also for the whole loop
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				return request{}, err
			}

			// if EOF, check if we have a full message before returning
			message, err = parseRequest(string(data))
			if err != nil {
				return request{}, fmt.Errorf("got an EOF from the client before a full message was received")
			}
			return message, nil
		}

		data = append(data, tmp[:n]...)
		length += n
		clear(tmp)

		// check if we have a full http message yet
		message, err = parseRequest(string(data))
		if err != nil {
			// TODO: don't continue forever. Set a maximum message size?
			incompleteErr := &incompleteMessage{}
			if errors.As(err, &incompleteErr) {
				continue
			}

			return request{}, err
		}

		break
	}

	return message, nil
}

// Public methods to register callbacks
func (s *Server) Get(path string, callback CallbackFunc) error {
	return s.callbackMap.registerCallback(get, path, callback)
}

func (s *Server) Post(path string, callback CallbackFunc) error {
	return s.callbackMap.registerCallback(post, path, callback)
}

func (s *Server) Put(path string, callback CallbackFunc) error {
	return s.callbackMap.registerCallback(put, path, callback)
}

func (s *Server) Delete(path string, callback CallbackFunc) error {
	return s.callbackMap.registerCallback(delete, path, callback)
}
