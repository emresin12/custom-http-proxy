package main

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
)

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	buff := make([]byte, 1024)

	// Read the initial request from the client (browser)
	n, err := clientConn.Read(buff)
	if err != nil {
		fmt.Println("Error reading from client:", err)
		return
	}

	// Convert the buffer to a string
	requestLine := string(buff[:n])
	// fmt.Println("Request Line:", requestLine)

	// Split the request line
	splitStr := strings.Split(requestLine, " ")

	// Check if the request is a CONNECT request
	if splitStr[0] == "CONNECT" {
		// Extract the host:port from the CONNECT request
		hostPort := splitStr[1]
		// fmt.Println("HostPort:", hostPort)

		// Establish a connection to the target server (e.g., alive.github.com:443)
		serverConn, err := net.Dial("tcp", hostPort)
		if err != nil {
			fmt.Println("Error connecting to target server:", err)
			return
		}
		defer serverConn.Close()

		// Respond to the client with a successful connection status
		_, err = clientConn.Write([]byte("HTTP/1.1 200\r\n\r\n"))
		if err != nil {
			fmt.Println("Error writing 200 OK to client:", err)
			return
		}
		go io.Copy(serverConn, clientConn)
		io.Copy(clientConn, serverConn)
	}

	rawURL := splitStr[1]

	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}

	conn, dialErr := net.Dial("tcp", parsedURL.Host+":http")
	if dialErr != nil {
		fmt.Println("Error dialing:", dialErr)
		return
	}
	defer conn.Close()
	conn.Write(buff[:n])

	io.Copy(clientConn, conn)
}

func main() {
	// Listen on port 8080 for incoming connections
	listen, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listen.Close()

	fmt.Println("Listening on :8080...")

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

// func manualCopy(dst io.Writer, src io.Reader) error {
// 	buffer := make([]byte, 32*1024) // 32 KB buffer
// 	for {
// 		n, err := src.Read(buffer) // Read into buffer
// 		if n > 0 {
// 			fmt.Println("Read", n, "bytes")
// 			fmt.Println("Buffer:", string(buffer[:n]))
// 			if _, writeErr := dst.Write(buffer[:n]); writeErr != nil { // Write to destination
// 				return writeErr
// 			}
// 		}
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			return err // Handle read error
// 		}
// 	}
// 	return nil
// }
