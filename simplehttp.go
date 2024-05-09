package simplehttp

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const defaultMaxRequestBytes uint = 1 * 1024 * 1024 // 1 MB
const defaultReadTimeoutSeconds int = 60

type Server struct {
	Port               uint16
	callbackMap        callbackMap
	MaxRequestBytes    uint
	ReadTimeoutSeconds int
}

func NewServer(port uint16) Server {
	return Server{
		Port:               port,
		callbackMap:        newCallbackMap(),
		MaxRequestBytes:    defaultMaxRequestBytes,
		ReadTimeoutSeconds: defaultReadTimeoutSeconds,
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

		if s.ReadTimeoutSeconds > 0 {
			conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(s.ReadTimeoutSeconds)))
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	fmt.Println("============= Talking to", conn.RemoteAddr(), "=============")
	defer conn.Close()

	request, err := readRequest(conn, s.MaxRequestBytes)
	if err != nil {
		fmt.Print("Unable to read message from the connection: ", err)
		fmt.Print("\n=====================================================\n\n")
		return
	}

	fmt.Print(request)

	response := newResponse()

	// call end-user's callback
	method := request.method
	path := request.Path()
	err = s.callbackMap.invokeCallback(method, path, request, &response)
	if err != nil {
		fmt.Println(err)
		errorResponse := new500StatusResponse()
		conn.Write([]byte(errorResponse.String()))
		return
	}

	// send a response
	conn.Write([]byte(response.String()))
}

func readRequest(conn net.Conn, maxBytes uint) (Request, error) {
	// read data in chunks of min(1kB, maxBytes)
	var chunkSize uint = 1024
	if maxBytes < 1024 {
		chunkSize = maxBytes
	}

	tmp := make([]byte, chunkSize)
	data := make([]byte, 0)
	message := Request{}
	var length uint = 0

	for {
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
		length += uint(n)
		clear(tmp)

		// check if we have a full http message yet
		message, err = parseRequest(string(data))
		if err != nil {
			if length > maxBytes {
				return Request{}, fmt.Errorf("incoming request exceeded the maximum request size")
			}

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
