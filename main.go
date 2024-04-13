package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

const PORT int = 3030

func main() {
	listener, err := net.Listen("tcp4", ":"+strconv.Itoa(PORT))
	if err != nil {
		log.Fatal("Failed to open tcp listener: ", err)
	}

	fmt.Println("Listening on port", PORT)

	for {
		fmt.Print("Waiting for connection...\n\n")
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Failed to accept incoming connection: ", err)
		}

		fmt.Println("============= Talking to", conn.RemoteAddr(), "=============")

		// read data in chunks of 1kB
		tmp := make([]byte, 1024)
		data := make([]byte, 0)
		length := 0

		for {
			n, err := conn.Read(tmp)
			if err != nil {
				if err != io.EOF {
					fmt.Println("Read error -", err)
				}
				fmt.Println("here in EOF")
				break
			}

			fmt.Print(string(tmp))

			data = append(data, tmp[:n]...)
			length += n
			clear(tmp)

			// check if we have a full packet yet
			fullPacket, _ := isFullPacket(string(data))
			if fullPacket {
				break
			}
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

		conn.Close()

		fmt.Print("\n=====================================================\n\n")
	}

}

func isFullPacket(message string) (bool, error) {
	headerStart := strings.Index(message, "\r\n")
	headerEnd := strings.Index(message, "\r\n\r\n")
	if headerEnd > -1 {
		// if we got here, we have all the headers
		rawHeaders := strings.TrimSpace(message[headerStart:headerEnd])
		headers, err := parseHeaders(rawHeaders)

		if err != nil {
			return false, err
		}

		// if no 'Content-Length' header, packet is complete
		contentLengthStr, exists := headers["Content-Length"]
		if !exists {
			return true, nil
		}

		// do we have enough content yet?
		contentLength, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return false, fmt.Errorf("invalid value in Content-Length header: `%s`", contentLengthStr)
		}

		content := message[headerEnd+4:] // +4 strips off the \r\n\r\n
		if len(content) < contentLength {
			return false, nil
		}

		return true, nil
	}
	return false, nil
}

func parseHeaders(message string) (map[string]string, error) {
	headers := make(map[string]string)
	lines := strings.Split(message, "\r\n")

	for _, line := range lines {
		split := strings.SplitN(line, ":", 2)

		if len(split) != 2 {
			return nil, fmt.Errorf("could not parse the following header: `%s`", line)
		}

		field := strings.TrimSpace(split[0])
		value := strings.TrimSpace(split[1])
		headers[field] = value
	}

	return headers, nil
}
