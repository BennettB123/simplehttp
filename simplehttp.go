package simplehttp

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

func StartServer(port int) {
	listener, err := net.Listen("tcp4", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("Failed to open tcp listener: ", err)
	}

	fmt.Println("Listening on port", port)

	for {
		fmt.Print("Waiting for connection...\n")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Failed to accept incoming connection: ", err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	fmt.Println("============= Talking to", conn.RemoteAddr(), "=============")

	// read data in chunks of 1kB
	tmp := make([]byte, 1024)
	data := make([]byte, 0)
	packet := packet{}
	length := 0

	for {
		// TODO: add timeout here and also for the whole loop
		n, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Read error -", err)
			}
			fmt.Println("here in EOF")
			break
		}

		data = append(data, tmp[:n]...)
		length += n
		clear(tmp)

		// check if we have a full packet yet
		packet, err = parsePacket(string(data))
		if err != nil {
			// TODO: don't continue forever. Set a maximum packet size?
			continue
		}

		break
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
