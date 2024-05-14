// The simplehttp package is a bare-bones HTTP/1.0 web framework for go.
// It supports registering callbacks that are invoked whenever
// specific HTTP methods and URLs are requested by a client.
// It takes heavy inspiration from the [Express] web framework for Node.
//
// Note: this package should not be used in a production environment. It
// was purely created as a learning opportunity to gain experience with go.
//
// [Express]: https://expressjs.com/
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

// A Server represents an HTTP server that listens on a specific port.
// A Server should only be created using the [NewServer] method to ensure
// it is properly initialized. The server does not support persistent connections.
// After a request has been received and a response has been returned, the server will
// close the connection.
type Server struct {
	// The Port that the server listens on
	Port uint16
	// MaxRequestBytes is the maximum number of bytes an incoming request can
	// be before the server rejects it.
	MaxRequestBytes uint
	// ReadTimeoutSeconds is the time in seconds that the server waits for a
	// request before closing the connection.
	ReadTimeoutSeconds int
	// Logger is a user-implementation of the [Logger] interface that the
	// server will send diagnostic messages about incoming requests and
	// outgoing responses. If a Logger is not provided, the server will
	// discard all log messages.
	Logger      Logger
	callbackMap callbackMap
}

// Creates and initializes a new [Server] object and
// assigns port to [Server.Port]
func NewServer(port uint16) Server {
	return Server{
		Port:               port,
		callbackMap:        newCallbackMap(),
		MaxRequestBytes:    defaultMaxRequestBytes,
		ReadTimeoutSeconds: defaultReadTimeoutSeconds,
		Logger:             nilLogger{},
	}
}

// Registers a callback that will be invoked whenever a GET request is
// made to the provided path. The callback is a function that takes in a [Request] and
// [*Response] and returns an error. See [CallbackFunc] for details on this function
func (s *Server) Get(path string, callback CallbackFunc) error {
	return s.callbackMap.registerCallback(get, path, callback)
}

// Registers a callback that will be invoked whenever a POST request is
// made to the provided path. The callback is a function that takes in a [Request] and
// [*Response] and returns an error. See [CallbackFunc] for details on this function
func (s *Server) Post(path string, callback CallbackFunc) error {
	return s.callbackMap.registerCallback(post, path, callback)
}

// Registers a callback that will be invoked whenever a PUT request is
// made to the provided path. The callback is a function that takes in a [Request] and
// [*Response] and returns an error. See [CallbackFunc] for details on this function
func (s *Server) Put(path string, callback CallbackFunc) error {
	return s.callbackMap.registerCallback(put, path, callback)
}

// Registers a callback that will be invoked whenever a DELETE request is
// made to the provided path. The callback is a function that takes in a [Request] and
// [*Response] and returns an error. See [CallbackFunc] for details on this function
func (s *Server) Delete(path string, callback CallbackFunc) error {
	return s.callbackMap.registerCallback(delete, path, callback)
}

// Starts the Server and begins listening for requests.
// Multiple requests can be handled in parallel with no hard limit.
// Returns an error if the Server was unable to open a TCP listener.
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
