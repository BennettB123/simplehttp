package simplehttp

import (
	"fmt"
	"io"
	"net"
)

type Server struct {
	Port uint16
}

func NewServer(port uint16) Server {
	return Server{
		Port: port,
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

	packet, err := readPacket(conn)
	if err != nil {
		fmt.Println("Unable to read a packet from the connection: ", err)
	}

	fmt.Print(packet.buildString())

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

	conn.Close()

	fmt.Print("\n=====================================================\n\n")
}

func readPacket(conn net.Conn) (packet, error) {
	// read data in chunks of 1kB
	tmp := make([]byte, 1024)
	data := make([]byte, 0)
	pack := packet{}
	length := 0

	for {
		// TODO: add timeout here and also for the whole loop
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				return packet{}, err
			}

			// if we get an EOF, check if we have a full packet then return
			pack, err = parsePacket(string(data))
			if err != nil {
				return packet{}, fmt.Errorf("got an EOF from the client before a full packet was received")
			}
			return pack, nil
		}

		data = append(data, tmp[:n]...)
		length += n
		clear(tmp)

		// check if we have a full packet yet
		pack, err = parsePacket(string(data))
		if err != nil {
			// TODO: don't continue forever. Set a maximum packet size?
			continue
		}

		break
	}

	return pack, nil
}
