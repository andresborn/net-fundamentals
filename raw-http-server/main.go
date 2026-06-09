package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:8080")

	if err != nil {
		log.Fatal("Error listening: ", err)
	}

	defer listener.Close()

	log.Println("Listening...")

	for {

		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error accepting connection: ", err)
		}

		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		request, err := readRequest(conn)

		if err == io.EOF {
			return // Client done,
		}

		if err != nil {
			log.Println("Error: ", err)
		}

		method, uri, _ := parseRequest(request)

		// Bonus: could add checking method and uri for malformed requests.

		if method != "GET" {
			body := "<h1>405 Method Not Allowed</h1>"
			response := fmt.Sprintf("HTTP/1.1 405 Method Not Allowed\r\nContent-Length: %d\r\nContent-Type: text/html\r\n\r\n%s", len(body), body)
			conn.Write([]byte(response))
			continue
		}

		if uri == "/" {
			body := "<h1>Welcome</h1>"
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Length: %d\r\nContent-Type: text/html\r\n\r\n%s", len(body), body)
			conn.Write([]byte(response))
			continue
		}

		if uri == "/about" {
			body := "<h1>This is my raw HTTP Server, written completely by hand.</h1>"
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Length: %d\r\nContent-Type: text/html\r\n\r\n%s", len(body), body)
			conn.Write([]byte(response))
			continue
		}

		body := "<h1>404 Not Found</h1>"
		response := fmt.Sprintf("HTTP/1.1 404 Not Found\r\nContent-Length: %d\r\nContent-Type: text/html\r\n\r\n%s", len(body), body)
		conn.Write([]byte(response))
	}

}

func readRequest(conn net.Conn) (string, error) {
	var accumulated []byte
	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)

		if err == io.EOF {
			return "", err
		}

		if err != nil {
			log.Println("Error reading connection: ", err)
			return "", err
		}

		accumulated = append(accumulated, buffer[:n]...)

		if strings.Contains(string(accumulated), "\r\n\r\n") {
			return string(accumulated), nil
		}
	}

}

func parseRequest(request string) (string, string, string) {
	// Split by new lines
	splitLines := strings.Split(request, "\r\n")
	if len(splitLines) == 0 {
		return "", "", ""
	}

	// Split start line by empty spaces
	startLine := strings.Split(splitLines[0], " ")
	if len(startLine) < 3 {
		return "", "", ""
	}

	method := startLine[0]
	uri := startLine[1]
	protocolVersion := startLine[2]
	return method, uri, protocolVersion
}
