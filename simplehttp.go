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
	Logger             Logger
}

func NewServer(port uint16) Server {
	return Server{
		Port:               port,
		callbackMap:        newCallbackMap(),
		MaxRequestBytes:    defaultMaxRequestBytes,
		ReadTimeoutSeconds: defaultReadTimeoutSeconds,
		Logger:             nilLogger{},
	}
}

func (s *Server) Get(path string, callback callbackFunc) error {
	return s.callbackMap.registerCallback(get, path, callback)
}

func (s *Server) Post(path string, callback callbackFunc) error {
	return s.callbackMap.registerCallback(post, path, callback)
}

func (s *Server) Put(path string, callback callbackFunc) error {
	return s.callbackMap.registerCallback(put, path, callback)
}

func (s *Server) Delete(path string, callback callbackFunc) error {
	return s.callbackMap.registerCallback(delete, path, callback)
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return fmt.Errorf("failed to open tcp listener: %v", err)
	}

	s.Logger.LogMessage(fmt.Sprintf("Listening on port %d", s.Port))

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.Logger.LogMessage(fmt.Sprintf("Failed to accept incoming connection: %v", err))
		}

		if s.ReadTimeoutSeconds > 0 {
			conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(s.ReadTimeoutSeconds)))
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	s.Logger.LogMessage(fmt.Sprintf("Connected to remote address %s", conn.RemoteAddr()))
	defer conn.Close()

	request, err := readRequest(conn, s.MaxRequestBytes)
	if err != nil {
		s.Logger.LogMessage(fmt.Sprintf("Unable to read message from the connection: %v", err))
		s.Logger.LogMessage(fmt.Sprintf("Disconnecting from remote address %s", conn.RemoteAddr()))
		return
	}

	s.Logger.LogMessage(fmt.Sprintf("Request from %s:", conn.RemoteAddr()))
	s.Logger.LogMessage("<<<<<<<<")
	s.Logger.LogMessage(request.rawMessage)
	s.Logger.LogMessage("<<<<<<<<")

	response := newResponse()

	// call end-user's callback
	method := request.method
	path := request.Path()
	err = s.callbackMap.invokeCallback(method, path, request, &response)
	if err != nil {
		s.Logger.LogMessage(err.Error())
		errorResponse := new500StatusResponse()
		conn.Write([]byte(errorResponse.String()))
		s.Logger.LogMessage(fmt.Sprintf("Disconnecting from remote address %s", conn.RemoteAddr()))
		return
	}

	// send a response
	s.Logger.LogMessage(fmt.Sprintf("Sending request to %s:", conn.RemoteAddr()))
	s.Logger.LogMessage(">>>>>>>>")
	s.Logger.LogMessage(response.String())
	s.Logger.LogMessage(">>>>>>>>")
	conn.Write([]byte(response.String()))

	s.Logger.LogMessage(fmt.Sprintf("Disconnecting from remote address %s", conn.RemoteAddr()))
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
