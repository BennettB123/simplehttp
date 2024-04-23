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
		callbackMap: newCallbackMap(),
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

	fmt.Print(request)

	response := newResponse()

	// call end-user's callback
	method := request.Method
	path := request.getPath()
	err = s.callbackMap.invokeCallback(method, path, request, &response)
	if err != nil {
		fmt.Println(err)
	}

	// send a response
	conn.Write([]byte(response.String()))

	fmt.Print("\n=====================================================\n\n")
}

func readRequest(conn net.Conn) (Request, error) {
	// read data in chunks of 1kB
	tmp := make([]byte, 1024)
	data := make([]byte, 0)
	message := Request{}
	length := 0

	for {
		// TODO: add timeout here and also for the whole loop
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				return Request{}, err
			}

			// if EOF, check if we have a full message before returning
			message, err = parseRequest(string(data))
			if err != nil {
				return Request{}, fmt.Errorf(
					"got an EOF from the client before a full message was received")
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

			return Request{}, err
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
